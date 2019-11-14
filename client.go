// Package hg659 provides the Go client library for the undocumented API
// of Huawei HG659 devices.
package hg659

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/ericyan/hg659/internal/quirks"
)

// A Client is a scoped http.Client for the HTTP API of HG659 devices.
type Client struct {
	*http.Client

	base *url.URL
	csrf *quirks.CSRF
}

// NewClient returns a new Client for the server.
func NewClient(server string) (*Client, error) {
	c := &Client{
		new(http.Client),
		&url.URL{Scheme: "http", Host: server, Path: "/"},
		nil,
	}
	c.Jar, _ = cookiejar.New(nil)

	resp, err := c.Get(c.base.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if len(c.Jar.Cookies(c.base)) == 0 {
		c.Jar.SetCookies(c.base, resp.Cookies())
	}

	csrf, err := quirks.ExtractCSRF(resp.Body)
	if err != nil {
		return nil, err
	}
	c.csrf = csrf

	return c, nil
}

func (c *Client) get(path string) ([]byte, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	resp, err := c.Get(c.base.ResolveReference(u).String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return quirks.ReadAll(resp.Body)
}

func (c *Client) post(path string, data interface{}) ([]byte, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(struct {
		CSRF *quirks.CSRF `json:"csrf"`
		Data interface{}  `json:"data"`
	}{c.csrf, data})
	if err != nil {
		return nil, err
	}

	resp, err := c.Post(c.base.ResolveReference(u).String(), "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return quirks.ReadAll(resp.Body)
}

// Login authenticates the session with given user credentials.
func (c *Client) Login(username, password string) error {
	req := quirks.NewLoginRequest(username, password, c.csrf)
	data, err := c.post("/api/system/user_login", req)
	if err != nil {
		return err
	}

	var resp struct {
		*quirks.CSRF
		Error     string `json:"errorCategory"`
		ErrorCode int    `json:"errcode"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	if resp.CSRF != nil {
		c.csrf = resp.CSRF
	}

	if resp.Error != "ok" {
		return errors.New(resp.Error)
	}

	if resp.ErrorCode != 0 {
		return errors.New("unknown error")
	}

	return nil
}

// Heartbeat sends a heartbeat request to keep the session alive.
func (c *Client) Heartbeat() error {
	if _, err := c.get("/api/system/heartbeat"); err != nil {
		return err
	}

	return nil
}

// DeviceInfo represents device metadata.
type DeviceInfo struct {
	DeviceID string
	Model    string
	Version  string
	Uptime   int
}

// GetDeviceInfo returns information about the device itself.
func (c *Client) GetDeviceInfo() (*DeviceInfo, error) {
	data, err := c.get("/api/system/deviceinfo")
	if err != nil {
		return nil, err
	}

	var info struct {
		DeviceName      string
		ManufacturerOUI string
		SerialNumber    string
		SoftwareVersion string
		HardwareVersion string
		UpTime          int
	}
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &DeviceInfo{
		DeviceID: info.ManufacturerOUI + "-" + info.SerialNumber,
		Model:    info.DeviceName + " " + info.HardwareVersion,
		Version:  info.SoftwareVersion,
		Uptime:   info.UpTime,
	}, nil
}

// Host represents a connected device.
type Host struct {
	Name          string
	IsConnected   bool
	MACAddr       net.HardwareAddr
	IPAddrs       []net.IP
	InterfaceName string
	LeaseTime     int
}

// GetHosts returns a list of known hosts.
func (c *Client) GetHosts() ([]Host, error) {
	data, err := c.get("/api/system/HostInfo")
	if err != nil {
		return nil, err
	}

	var hostinfo []struct {
		HostName        string
		Active46        bool
		Layer2Interface string
		MACAddress      string
		IPAddress       string
		Ipv6Addrs       []map[string]string
		LeaseTime       int
	}
	if err := json.Unmarshal(data, &hostinfo); err != nil {
		return nil, err
	}

	hosts := make([]Host, len(hostinfo))
	for i, h := range hostinfo {
		hostname := strings.TrimSuffix(h.HostName, "_Wireless")
		if len(hostname) == len(h.HostName) {
			hostname = strings.TrimSuffix(h.HostName, "_Ethernet")
		}

		mac, err := net.ParseMAC(h.MACAddress)
		if err != nil {
			return nil, err
		}

		addrs := make([]net.IP, 0)
		if h.IPAddress != "" {
			addrs = append(addrs, net.ParseIP(h.IPAddress).To4())
		}
		for _, addr := range h.Ipv6Addrs {
			addrs = append(addrs, net.ParseIP(addr["Ipv6Addr"]).To16())
		}

		hosts[i] = Host{
			Name:          hostname,
			IsConnected:   h.Active46,
			InterfaceName: h.Layer2Interface,
			MACAddr:       mac,
			IPAddrs:       addrs,
			LeaseTime:     h.LeaseTime,
		}
	}

	return hosts, nil
}

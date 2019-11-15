package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/ericyan/hg659"
)

var (
	bind = flag.String("bind", "127.0.0.1", "interface to bind")
	port = flag.Int("port", 8659, "port to run on")
	host = flag.String("host", "192.168.1.1", "hostname of IP of the HG659 device")
	user = flag.String("user", "user", "username for login")
	pass = flag.String("pass", "", "password for login")
)

func writeJSON(w http.ResponseWriter, data interface{}) {
	resp, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func main() {
	flag.Parse()

	c, err := hg659.NewClient(*host)
	if err != nil {
		log.Fatalln(err)
	}

	err = c.Login(*user, *pass)
	if err != nil {
		log.Fatalln(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		info, err := c.GetDeviceInfo()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		writeJSON(w, info)
	})
	mux.HandleFunc("/hosts", func(w http.ResponseWriter, req *http.Request) {
		hosts, err := c.GetHosts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		writeJSON(w, hosts)
	})

	srv := &http.Server{
		Addr:    *bind + ":" + strconv.Itoa(*port),
		Handler: mux,
	}

	log.Printf("Listening on %s:%d...\n", *bind, *port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start HTTP server: %s\n", err.Error())
	}
}

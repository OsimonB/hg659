package quirks

import (
	"io"

	"golang.org/x/net/html"
)

// CSRF holds the Cross-Site Request Forgery parameter and token.
type CSRF struct {
	Param string `json:"csrf_param"`
	Token string `json:"csrf_token"`
}

// ExtractCSRF extracts CSRF token from HTML meta tags.
func ExtractCSRF(r io.Reader) (*CSRF, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	c := new(CSRF)

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			if n.Attr[0].Val == "csrf_param" {
				c.Param = n.Attr[1].Val
			}

			if n.Attr[0].Val == "csrf_token" {
				c.Token = n.Attr[1].Val
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return c, nil
}

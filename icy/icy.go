package icy

import (
	"bytes"
	"net"
	"net/http"
	"time"
)

type dialer struct {
	*net.Dialer
}

type conn struct {
	net.Conn
	read bool
}

var DefaultClient = http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&dialer{
			Dialer: &net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			},
		}).Dial,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

// Read modifies the first line of an ICY stream response
// to conform to Go's HTTP version requirements:
// http://golang.org/pkg/net/http/#ParseHTTPVersion.
func (c *conn) Read(b []byte) (int, error) {
	if c.read {
		return c.Conn.Read(b)
	}

	const headerICY = "ICY"
	const headerHTTP = "HTTP/1.1"

	n, err := c.Conn.Read(b[:len(headerICY)])
	if err != nil {
		return n, err
	}

	if bytes.HasPrefix(b, []byte(headerICY)) {
		copy(b, []byte(headerHTTP))
		return len(headerHTTP), nil
	}

	c.read = true
	return n, err
}

func (d *dialer) Dial(network, address string) (net.Conn, error) {
	c, err := d.Dialer.Dial(network, address)
	cn := conn{
		Conn: c,
	}
	return &cn, err
}

func Get(url string) (resp *http.Response, err error) {
	return DefaultClient.Get(url)
}

package tor

import (
	"fmt"
	"golang.org/x/net/proxy"
	"net/http"
	"net/url"
)

// Return new Tor proxy client.
func NewClient(host string, port int) (*http.Client, error) {
	addr := fmt.Sprintf("socks5://%s:%d", host, port)
	proxyAddr, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	dialer, err := proxy.FromURL(proxyAddr, proxy.Direct)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{Dial: dialer.Dial}
	client := &http.Client{Transport: transport}
	return client, nil
}

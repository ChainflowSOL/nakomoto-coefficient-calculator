package chains

import (
	"crypto/tls"
	"net/http"
	"time"
)

// http11Client forces HTTP/1.1 via ALPN. CloudFront / Akamai in front of
// several chain APIs (api.bnbchain.org, mainnet.base.org, api.hyperliquid.xyz)
// silently reset Go's HTTP/2 handshakes from datacenter IP ranges while
// accepting HTTP/1.1. Using this client avoids those resets.
func http11Client(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSNextProto:    map[string]func(string, *tls.Conn) http.RoundTripper{},
			TLSClientConfig: &tls.Config{NextProtos: []string{"http/1.1"}},
		},
	}
}

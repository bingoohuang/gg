package netx

import (
	"bufio"
	"github.com/juju/errors"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
)

// NewProxyDialer based on proxy url.
func NewProxyDialer(proxyUrl string) (proxy.Dialer, error) {
	u, err := url.Parse(proxyUrl)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "socks5":
		return &socks5ProxyClient{proxyUrl: u}, nil
	case "http":
		return &httpProxyClient{proxyUrl: u}, nil
	default:
		return &defaultProxyClient{}, nil
	}
}

// defaultProxyClient is used to implement a proxy in default.
type defaultProxyClient struct {
	rAddr *net.TCPAddr
}

// Dial implementation of ProxyConn.
// Set KeepAlive=-1 to reduce the call of syscall.
func (dc *defaultProxyClient) Dial(network string, address string) (conn net.Conn, err error) {
	if dc.rAddr == nil {
		if dc.rAddr, err = net.ResolveTCPAddr("tcp", address); err != nil {
			return nil, err
		}
	}
	return net.DialTCP(network, nil, dc.rAddr)
}

// socks5ProxyClient is used to implement a proxy in socks5
type socks5ProxyClient struct {
	proxyUrl *url.URL
}

// Dial implementation of ProxyConn.
func (s5 *socks5ProxyClient) Dial(network string, address string) (net.Conn, error) {
	d, err := proxy.FromURL(s5.proxyUrl, nil)
	if err != nil {
		return nil, err
	}

	return d.Dial(network, address)
}

// httpProxyClient is used to implement a proxy in http.
type httpProxyClient struct {
	proxyUrl *url.URL
}

// Dial implementation of ProxyConn
func (hc *httpProxyClient) Dial(_ string, address string) (net.Conn, error) {
	req, err := http.NewRequest("CONNECT", "http://"+address, nil)
	if err != nil {
		return nil, err
	}
	password, _ := hc.proxyUrl.User.Password()
	req.SetBasicAuth(hc.proxyUrl.User.Username(), password)
	proxyConn, err := net.Dial("tcp", hc.proxyUrl.Host)
	if err != nil {
		return nil, err
	}
	if err := req.Write(proxyConn); err != nil {
		return nil, err
	}
	res, err := http.ReadResponse(bufio.NewReader(proxyConn), req)
	if err != nil {
		return nil, err
	}
	_ = res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New("Proxy error " + res.Status)
	}
	return proxyConn, nil
}

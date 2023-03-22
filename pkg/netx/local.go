package netx

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bingoohuang/gg/pkg/goip"
	"github.com/bingoohuang/gg/pkg/netx/freeport"
	"github.com/bingoohuang/gg/pkg/ss"
	"github.com/bingoohuang/golog/pkg/randx"
	"github.com/go-resty/resty/v2"
)

// IsLocalAddr 判断addr（ip，域名等）是否指向本机
// 由于IP可能经由 iptable 指向，或者可能是域名，或者其它，不能直接与本机IP做对比
// 本方法构建一个临时的HTTP服务，然后使用指定的addr去连接改HTTP服务，如果能连接上，说明addr是指向本机的地址
func IsLocalAddr(addr string) (bool, error) {
	if addr == "localhost" || addr == "127.0.0.1" || addr == "::1" {
		return true, nil
	}

	if ss.AnyOf(addr, goip.ListIfaceNames()...) {
		return true, nil
	}

	port, err := freeport.PortE()
	if err != nil {
		return false, err
	}

	radStr := randx.String(512)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, radStr)
	})

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}

	go func() { _ = server.ListenAndServe() }()

	time.Sleep(100 * time.Millisecond) // nolint gomnd

	addr = `http://` + JoinHostPort(addr, port)
	resp, err := resty.New().SetTimeout(3 * time.Second).R().Get(addr)
	_ = server.Close()

	if err != nil {
		return false, err
	}

	return resp.String() == radStr, nil
}

// JoinHostPort make IP:Port for ipv4/domain or [IPv6]:Port for ipv6.
func JoinHostPort(host string, port int) string {
	if goip.IsIPv6(host) {
		return fmt.Sprintf("[%s]:%d", host, port)
	}

	return fmt.Sprintf("%s:%d", host, port)
}

package http_transport

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/dysnix/predictkube-libs/external/configs"
)

// Transport implements the http.RoundTripper interface with
// the net/http client.
type netHttpTransport struct {
	rtp       http.RoundTripper
	dialer    *net.Dialer
	connStart time.Time
	connEnd   time.Time
	reqStart  time.Time
	reqEnd    time.Time
	conf      configs.TransportGetter
	tlsConf   *tls.Config
}

type HttpTransport interface {
	http.RoundTripper
	configs.SignalCloser
}

func NewNetHttpTransport(options ...configs.TransportOption) (out *netHttpTransport, err error) {
	out = &netHttpTransport{}

	for _, op := range options {
		err := op(out)
		if err != nil {
			return nil, err
		}
	}

	out.dialer = &net.Dialer{
		Timeout:   out.conf.GetTransportConfigs().NetTransport.DialTimeout,
		KeepAlive: out.conf.GetTransportConfigs().NetTransport.KeepAlive,
	}

	tmpTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	if out.conf != nil {
		tmpTransport.DialContext = out.dial
		tmpTransport.DisableKeepAlives = out.conf.GetTransportConfigs().NetTransport.DisableKeepAlives
		tmpTransport.DisableCompression = out.conf.GetTransportConfigs().NetTransport.DisableCompression
		tmpTransport.TLSHandshakeTimeout = out.conf.GetTransportConfigs().NetTransport.TLSHandshakeTimeout
		tmpTransport.MaxIdleConns = out.conf.GetTransportConfigs().NetTransport.MaxIdleConns
		tmpTransport.MaxIdleConnsPerHost = out.conf.GetTransportConfigs().NetTransport.MaxIdleConnsPerHost
		tmpTransport.MaxConnsPerHost = out.conf.GetTransportConfigs().NetTransport.MaxConnsPerHost
		tmpTransport.IdleConnTimeout = out.conf.GetTransportConfigs().MaxIdleConnDuration
		tmpTransport.ResponseHeaderTimeout = out.conf.GetTransportConfigs().NetTransport.ResponseHeaderTimeout
		tmpTransport.ExpectContinueTimeout = out.conf.GetTransportConfigs().NetTransport.ExpectContinueTimeout
		tmpTransport.MaxResponseHeaderBytes = out.conf.GetTransportConfigs().NetTransport.MaxResponseHeaderBytes

		if out.conf.GetTransportConfigs().NetTransport.Buffer != nil {
			tmpTransport.WriteBufferSize = int(out.conf.GetTransportConfigs().NetTransport.Buffer.WriteBufferSize)
			tmpTransport.ReadBufferSize = int(out.conf.GetTransportConfigs().NetTransport.Buffer.ReadBufferSize)
		}
	}

	if out.tlsConf == nil {
		tmpTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	} else {
		tmpTransport.TLSClientConfig = out.tlsConf
	}

	out.rtp = tmpTransport

	return out, nil
}

func (t *netHttpTransport) SetConfigs(configs configs.TransportGetter) {
	t.conf = configs
}

func (t *netHttpTransport) SetTLS(conf *tls.Config) {
	t.tlsConf = conf
}

func (t *netHttpTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	t.reqStart = time.Now()
	resp, err := t.rtp.RoundTrip(r)
	t.reqEnd = time.Now()
	return resp, err
}

func (t *netHttpTransport) dial(_ context.Context, network, addr string) (net.Conn, error) {
	t.connStart = time.Now()
	cn, err := t.dialer.Dial(network, addr)
	t.connEnd = time.Now()
	return cn, err
}

type HttpTransportWithRequestStats interface {
	ReqDuration() time.Duration
	ConnDuration() time.Duration
	Duration() time.Duration
}

func (t *netHttpTransport) ReqDuration() time.Duration {
	return t.Duration() - t.ConnDuration()
}

func (t *netHttpTransport) ConnDuration() time.Duration {
	return t.connEnd.Sub(t.connStart)
}

func (t *netHttpTransport) Duration() time.Duration {
	return t.reqEnd.Sub(t.reqStart)
}

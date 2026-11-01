//go:build linux || darwin

package preview3

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/net/idna"
)

type httpTimeouts [3]time.Duration

func (p *HTTP) defaultTimeouts() httpTimeouts {
	return httpTimeouts{p.config.Timeout, p.config.Timeout, p.config.Timeout}
}
func (p *HTTP) makeTransport(timeouts httpTimeouts) *http.Transport {
	// Trailer names arrive with the final guest future. HTTP/1.1 supports
	// streaming these without predeclaring names or buffering the whole body.
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	return &http.Transport{Proxy: nil, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		return p.dial(ctx, network, address, timeouts[0])
	}, TLSClientConfig: p.config.TLSConfig, TLSHandshakeTimeout: timeouts[0], ResponseHeaderTimeout: timeouts[1], IdleConnTimeout: 30 * time.Second, MaxIdleConns: int(p.config.MaxRequests), MaxIdleConnsPerHost: int(p.config.MaxRequests), MaxConnsPerHost: int(p.config.MaxRequests), MaxResponseHeaderBytes: int64(p.config.MaxHeaderBytes), Protocols: protocols, DisableCompression: true}
}
func (p *HTTP) transportFor(timeouts httpTimeouts) (*http.Transport, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, wasm.ErrClosed
	}
	if transport := p.transports[timeouts]; transport != nil {
		p.mu.Unlock()
		return transport, nil
	}
	transport := p.makeTransport(timeouts)
	var evicted *http.Transport
	if len(p.transportOrder) == 8 {
		key := p.transportOrder[0]
		p.transportOrder = p.transportOrder[1:]
		evicted = p.transports[key]
		delete(p.transports, key)
	}
	p.transports[timeouts] = transport
	p.transportOrder = append(p.transportOrder, timeouts)
	p.mu.Unlock()
	if evicted != nil {
		evicted.CloseIdleConnections()
	}
	return transport, nil
}
func (p *HTTP) dial(ctx context.Context, network, address string, timeout time.Duration) (net.Conn, error) {
	if p.config.AllowConnect == nil {
		return nil, httpErr("destination-IP-prohibited")
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	host, portText, err := net.SplitHostPort(address)
	if err != nil {
		return nil, httpErr("HTTP-request-URI-invalid")
	}
	port, err := strconv.ParseUint(portText, 10, 16)
	if err != nil {
		return nil, httpErr("HTTP-request-URI-invalid")
	}
	var addresses []netip.Addr
	if ip, err := netip.ParseAddr(host); err == nil {
		addresses = []netip.Addr{ip}
	} else {
		if !p.network.config.AllowNameLookup {
			return nil, httpErr("destination-IP-prohibited")
		}
		if uint64(len(host)) > uint64(p.network.config.MaxNameBytes) {
			return nil, httpErr("DNS-error")
		}
		name, failed := idna.Lookup.ToASCII(host)
		if failed != nil || name == "" || len(name) > 253 {
			return nil, httpErr("DNS-error")
		}
		func() {
			defer func() {
				if recover() != nil {
					err = httpErr("DNS-error")
				}
			}()
			addresses, err = p.network.config.Resolve(ctx, name)
		}()
		if err != nil {
			return nil, httpTransportError(err, "DNS-error")
		}
		if len(addresses) == 0 || uint64(len(addresses)) > uint64(p.network.config.MaxAddresses) {
			return nil, httpErr("DNS-error")
		}
	}
	var last error
	for _, ip := range addresses {
		if !ip.IsValid() {
			return nil, httpErr("DNS-error")
		}
		scope := uint32(0)
		zone := ip.Zone()
		if zone != "" {
			n, err := strconv.ParseUint(zone, 10, 32)
			if err != nil {
				iface, failed := net.InterfaceByName(zone)
				if failed != nil {
					return nil, httpErr("destination-IP-unroutable")
				}
				n = uint64(iface.Index)
			}
			scope = uint32(n)
		}
		destination := SocketAddress{Address: netip.AddrPortFrom(ip.WithZone("").Unmap(), uint16(port)), ScopeID: scope}
		if err := p.network.authorize(ctx, destination, p.config.AllowConnect); err != nil {
			last = httpErr("destination-IP-prohibited")
			continue
		}
		conn, failed := (&net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second}).DialContext(ctx, network, net.JoinHostPort(ip.String(), portText))
		if failed == nil {
			return conn, nil
		}
		last = httpTransportError(failed, "connection-refused")
	}
	if last == nil {
		last = httpErr("destination-not-found")
	}
	return nil, last
}
func httpTransportError(err error, fallback string) error {
	var typed *httpError
	if errors.As(err, &typed) {
		return typed
	}
	name := fallback
	var dns *net.DNSError
	var cert *tls.CertificateVerificationError
	var unknown x509.UnknownAuthorityError
	var record tls.RecordHeaderError
	var neterr net.Error
	switch {
	case errors.As(err, &dns):
		name = "DNS-error"
		if dns.IsTimeout {
			name = "DNS-timeout"
		}
	case errors.As(err, &cert), errors.As(err, &unknown):
		name = "TLS-certificate-error"
	case errors.As(err, &record):
		name = "TLS-protocol-error"
	case errors.As(err, &neterr) && neterr.Timeout(), errors.Is(err, context.DeadlineExceeded):
		name = "connection-timeout"
	case errors.Is(err, io.ErrUnexpectedEOF):
		name = "HTTP-response-incomplete"
	}
	e := httpErr(name)
	e.cause = err
	return e
}
func (p *HTTP) fieldsFromHeader(header http.Header) (*host.HTTPFields, error) {
	if uint64(len(header)) > uint64(p.config.MaxHeaderFields) {
		return nil, wasm.ErrLimit
	}
	keys := make([]string, 0, len(header))
	for key := range header {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var fields []host.HTTPField
	for _, key := range keys {
		if uint64(len(fields))+uint64(len(header[key])) > uint64(p.config.MaxHeaderFields) {
			return nil, wasm.ErrLimit
		}
		for _, value := range header[key] {
			fields = append(fields, host.HTTPField{Name: key, Value: value})
		}
	}
	return host.NewHTTPFields(p.fields, fields, true)
}
func httpHeader(fields *host.HTTPFields) http.Header {
	out := http.Header{}
	if fields != nil {
		for _, field := range fields.Entries() {
			out.Add(field.Name, field.Value)
		}
	}
	return out
}
func httpContentLength(fields *host.HTTPFields) (*uint64, error) {
	values := fields.Values("content-length")
	var result *uint64
	for _, value := range values {
		n, err := strconv.ParseUint(value, 10, 64)
		if err != nil || strings.TrimSpace(value) != value {
			return nil, httpErr("HTTP-protocol-error")
		}
		if result != nil && *result != n {
			return nil, httpErr("HTTP-protocol-error")
		}
		copy := n
		result = &copy
	}
	return result, nil
}
func (p *HTTP) requestURL(message *httpMessage) (*url.URL, error) {
	scheme := "http"
	if message.scheme != nil {
		scheme = strings.ToLower(*message.scheme)
	}
	if scheme != "http" && scheme != "https" {
		return nil, httpErr("HTTP-protocol-error")
	}
	if message.authority == nil || *message.authority == "" {
		return nil, httpErr("HTTP-request-URI-invalid")
	}
	path := "/"
	if message.path != nil && *message.path != "" {
		path = *message.path
	}
	if strings.HasPrefix(path, "?") {
		path = "/" + path
	}
	parsed, err := url.ParseRequestURI(path)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || parsed.Fragment != "" {
		return nil, httpErr("HTTP-request-URI-invalid")
	}
	authority, err := url.Parse(scheme + "://" + *message.authority)
	if err != nil || authority.Hostname() == "" || authority.User != nil || authority.Path != "" || authority.RawQuery != "" || authority.Fragment != "" {
		return nil, httpErr("HTTP-request-URI-invalid")
	}
	parsed.Scheme = scheme
	parsed.Host = authority.Host
	return parsed, nil
}
func (p *HTTP) authorizeRequest(ctx context.Context, target *url.URL) (err error) {
	if p.config.AllowRequest == nil {
		return httpErr("HTTP-request-denied")
	}
	copy := *target
	complete := false
	defer func() {
		if !complete {
			_ = recover()
			err = httpErr("HTTP-request-denied")
		}
	}()
	err = p.config.AllowRequest(ctx, &copy)
	complete = true
	if err != nil {
		return httpErr("HTTP-request-denied")
	}
	return ctx.Err()
}

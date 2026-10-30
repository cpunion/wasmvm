//go:build linux || darwin

package preview2

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

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/net/idna"
)

type httpTimeouts [3]time.Duration

func (p *HTTP) defaultTimeouts() httpTimeouts {
	return httpTimeouts{p.config.Timeout, p.config.Timeout, p.config.Timeout}
}
func (p *HTTP) newTransport() *http.Transport { return p.makeTransport(p.defaultTimeouts()) }
func (p *HTTP) makeTransport(timeouts httpTimeouts) *http.Transport {
	return &http.Transport{Proxy: nil, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		return p.dial(ctx, network, address, timeouts[0])
	}, TLSClientConfig: p.config.TLSConfig, TLSHandshakeTimeout: timeouts[0], ResponseHeaderTimeout: timeouts[1], IdleConnTimeout: 30 * time.Second, MaxIdleConns: int(p.config.MaxRequests), MaxIdleConnsPerHost: int(p.config.MaxRequests), MaxConnsPerHost: int(p.config.MaxRequests), MaxResponseHeaderBytes: int64(p.config.MaxHeaderBytes), ForceAttemptHTTP2: true, DisableCompression: true}
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
		if err := socketAuthorize(ctx, p.ctx, destination, p.config.AllowConnect); err != nil {
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
func (p *HTTP) fieldsFromHeader(header http.Header) (*httpFields, error) {
	if uint64(len(header)) > uint64(p.config.MaxHeaderFields) {
		return nil, wasm.ErrLimit
	}
	keys := make([]string, 0, len(header))
	for key := range header {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var fields []httpField
	for _, key := range keys {
		if uint64(len(fields))+uint64(len(header[key])) > uint64(p.config.MaxHeaderFields) {
			return nil, wasm.ErrLimit
		}
		for _, value := range header[key] {
			fields = append(fields, httpField{key, value})
		}
	}
	return newHTTPFields(p.fields, fields, true)
}
func httpHeader(fields *httpFields) http.Header {
	out := http.Header{}
	if fields != nil {
		for _, field := range fields.entries() {
			out.Add(field.name, field.value)
		}
	}
	return out
}
func httpContentLength(fields *httpFields) (*uint64, error) {
	values := fields.values("content-length")
	var result *uint64
	for _, value := range values {
		n, err := strconv.ParseUint(value, 10, 63)
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
	if message.scheme.set {
		scheme = strings.ToLower(message.scheme.value)
	}
	if scheme != "http" && scheme != "https" {
		return nil, httpErr("HTTP-protocol-error")
	}
	if !message.authority.set || message.authority.value == "" {
		return nil, httpErr("HTTP-request-URI-invalid")
	}
	path := "/"
	if message.path.set && message.path.value != "" {
		path = message.path.value
	}
	if strings.HasPrefix(path, "?") {
		path = "/" + path
	}
	parsed, err := url.ParseRequestURI(path)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || parsed.Fragment != "" {
		return nil, httpErr("HTTP-request-URI-invalid")
	}
	authority, err := url.Parse(scheme + "://" + message.authority.value)
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
func (p *HTTP) sendRequest(c *component.Caller, args []component.Value) (component.Value, error) {
	entry, err := p.take(c, args[0], httpOutgoingRequest)
	if err != nil {
		return component.Value{}, err
	}
	entry.mu.Lock()
	message := *entry.message
	entry.mu.Unlock()
	handedOff := false
	defer func() {
		if !handedOff {
			message.body.close()
		}
	}()
	timeouts := p.defaultTimeouts()
	if args[1].Case == 1 {
		options, failed := p.take(c, args[1].Elements[0], httpRequestOptions)
		if failed != nil {
			return component.Value{}, failed
		}
		for i, value := range options.options {
			if value != nil {
				timeouts[i] = time.Duration(*value)
			}
		}
	}
	target, err := p.requestURL(&message)
	if err != nil {
		return httpErrorResult(err), nil
	}
	length, err := httpContentLength(message.headers)
	if err != nil {
		return httpErrorResult(err), nil
	}
	ctx, cancel := context.WithTimeout(p.ctx, p.config.Timeout)
	joint := context.AfterFunc(c.Context(), cancel)
	if err = p.authorizeRequest(ctx, target); err != nil {
		joint()
		cancel()
		return httpErrorResult(err), nil
	}
	// Ownership outlives this host call. Caller cancellation is relevant to the
	// policy phase only; subsequent guest polling has its own call context.
	joint()
	transport, err := p.transportFor(timeouts)
	if err != nil {
		cancel()
		return component.Value{}, err
	}
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		cancel()
		return component.Value{}, wasm.ErrClosed
	}
	if p.requests >= p.config.MaxRequests {
		p.mu.Unlock()
		cancel()
		return httpErrorResult(httpErr("connection-limit-reached")), nil
	}
	p.requests++
	p.mu.Unlock()
	rollback := func() { cancel(); p.mu.Lock(); p.requests--; p.mu.Unlock() }
	future := &httpFuture{changed: make(chan struct{}), cancel: cancel}
	value, err := p.publish(c, &httpEntry{kind: httpFutureResponse, future: future})
	if err != nil {
		message.body.close()
		rollback()
		return component.Value{}, err
	}
	if err = p.startWorker(); err != nil {
		c.DropResource(value.Resource)
		rollback()
		return component.Value{}, err
	}
	handedOff = true
	go func() {
		defer p.workers.Done()
		defer func() { p.mu.Lock(); p.requests--; p.mu.Unlock() }()
		if timeouts[1] == 0 {
			p.completeResponse(future, nil, httpErr("connection-read-timeout"))
			message.body.close()
			return
		}
		request := &http.Request{Method: message.method, URL: target, Header: httpHeader(message.headers), Host: target.Host}
		for key := range request.Header {
			if strings.EqualFold(key, "content-length") {
				delete(request.Header, key)
			}
		}
		if message.bodyTaken {
			trailerMap := http.Header{}
			request.Trailer = trailerMap
			request.Body = &httpBodyReader{body: message.body, ctx: ctx, trailers: trailerMap}
			request.ContentLength = -1
			if length != nil {
				request.ContentLength = int64(*length)
				if *length == 0 {
					request.Body = http.NoBody
				}
			}
		} else if length != nil && *length != 0 {
			p.completeResponse(future, nil, httpErr("HTTP-request-body-size"))
			message.body.close()
			return
		} else {
			message.body.finish(nil, nil)
			message.body.close()
		}
		response, failed := transport.RoundTrip(request.WithContext(ctx))
		if failed != nil {
			p.completeResponse(future, nil, httpTransportError(failed, "HTTP-protocol-error"))
			message.body.close()
			return
		}
		fields, failed := p.fieldsFromHeader(response.Header)
		if failed != nil {
			response.Body.Close()
			p.completeResponse(future, nil, httpErr("HTTP-response-header-section-size"))
			return
		}
		body, failed := p.newBody(false)
		if failed != nil {
			response.Body.Close()
			p.completeResponse(future, nil, failed)
			return
		}
		body.errorName = "HTTP-response-body-size"
		body.mu.Lock()
		body.externalCancel = cancel
		body.mu.Unlock()
		if failed = body.pump(response.Body, response.Trailer, timeouts[1], timeouts[2]); failed != nil {
			response.Body.Close()
			body.close()
			p.completeResponse(future, nil, failed)
			return
		}
		p.completeResponse(future, &httpMessage{headers: fields, status: uint16(response.StatusCode), body: body}, nil)
	}()
	return httpOK(value), nil
}
func (p *HTTP) completeResponse(future *httpFuture, message *httpMessage, err error) {
	future.mu.Lock()
	if future.ready {
		future.mu.Unlock()
		if message != nil {
			message.body.close()
		}
		return
	}
	future.ready = true
	future.response = message
	future.failure = err
	close(future.changed)
	future.mu.Unlock()
}

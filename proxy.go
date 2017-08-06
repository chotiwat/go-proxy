package proxy

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	logger "github.com/blendlabs/go-logger"
)

const (
	// EventProxyRequest is a logger flag.
	EventProxyRequest logger.Event = "proxy.request"
)

// New returns a new proxy.
func New() *Proxy {
	return &Proxy{}
}

// Proxy is a factory for a simple reverse proxy.
type Proxy struct {
	logger    *logger.Agent
	upstreams []*Upstream
	resolver  Resolver
}

// WithLogger sets a property and returns the proxy reference.
func (p *Proxy) WithLogger(log *logger.Agent) *Proxy {
	if !log.HasListener(EventProxyRequest) {
		log.AddEventListener(EventProxyRequest, NewRequestListener(p.proxyRequestWriter))
	}
	p.logger = log
	return p
}

// WithUpstream adds an upstream by URL.
func (p *Proxy) WithUpstream(upstream *Upstream) *Proxy {
	p.upstreams = append(p.upstreams, upstream)
	return p
}

// WithResolver sets a property and returns the proxy reference.
func (p *Proxy) WithResolver(resolver Resolver) *Proxy {
	p.resolver = resolver
	return p
}

// ServeHTTP is the http entrypoint.
func (p *Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// set the default resolver if unset.
	if p.resolver == nil {
		p.resolver = RoundRobinResolver(p.upstreams)
	}

	upstream, err := p.resolver(req, p.upstreams)

	if err != nil {
		p.logger.Error(err)
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	if upstream == nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	upstream.ServeHTTP(rw, req)
}

func (p *Proxy) proxyRequestWriter(writer *logger.Writer, ts logger.TimeSource, target *url.URL, req *http.Request, statusCode, contentLength int, elapsed time.Duration) {
	buffer := writer.GetBuffer()
	defer writer.PutBuffer(buffer)

	buffer.WriteString(writer.FormatEvent(EventProxyRequest, logger.ColorBlue))
	buffer.WriteRune(logger.RuneSpace)
	buffer.WriteString(target.String())
	buffer.WriteRune(logger.RuneSpace)
	buffer.WriteString(logger.GetIP(req))
	buffer.WriteRune(logger.RuneSpace)
	buffer.WriteString(writer.Colorize(req.Method, logger.ColorBlue))
	buffer.WriteRune(logger.RuneSpace)
	buffer.WriteString(req.URL.Path)
	buffer.WriteRune(logger.RuneSpace)
	buffer.WriteString(writer.ColorizeByStatusCode(statusCode, strconv.Itoa(statusCode)))
	buffer.WriteRune(logger.RuneSpace)
	buffer.WriteString(elapsed.String())
	buffer.WriteRune(logger.RuneSpace)
	buffer.WriteString(logger.File.FormatSize(contentLength))

	writer.WriteWithTimeSource(ts, buffer.Bytes())
}

package proxy

import "net/http"

// Resolver is a function that takes a request and produces a destination `url.URL`.
type Resolver func(*http.Request, []*Upstream) (*Upstream, error)

// RoundRobinResolver returns a closure based resolver that rotates through upstreams uniformly.
func RoundRobinResolver(upstreams []*Upstream) Resolver {
	current, total := 0, len(upstreams)
	return func(req *http.Request, upstreams []*Upstream) (*Upstream, error) {
		upstream := upstreams[current]
		current = (current + 1) % total
		return upstream, nil
	}
}

package main

import (
	"flag"
	"net/http"
	"net/url"
	"os"
	"strings"

	logger "github.com/blendlabs/go-logger"
	proxy "github.com/blendlabs/go-proxy"
)

const (
	// DefaultPort is the default port the proxy listens on.
	DefaultPort = "8888"
)

// Upstreams is a flag variable for upstreams.
type Upstreams []string

// String returns a string representation of the upstreams.
func (u *Upstreams) String() string {
	if u == nil {
		return "<nil>"
	}
	return strings.Join(*u, ", ")
}

// Set adds a flag value.
func (u *Upstreams) Set(value string) error {
	*u = append(*u, value)
	return nil
}

func main() {
	log := logger.NewFromEnv()

	var upstreams Upstreams
	flag.Var(&upstreams, "upstream", "An upstream server to proxy traffic to")

	var tlsCert string
	flag.StringVar(&tlsCert, "tls-cert", "", "The path to the tls certificate file (--tls-key must also be set)")

	var tlsKey string
	flag.StringVar(&tlsKey, "tls-key", "", "The path to the tls key file (--tls-cert must also be set)")

	var bindAddr string
	flag.StringVar(&bindAddr, "listen", ":8080", "The address to listen on.")

	var logEvents string
	flag.StringVar(&logEvents, "log-events", "", "Logger events to enable or disable. Coalesced with `LOG_EVENTS`")

	var close bool
	flag.BoolVar(&close, "close", false, "Should close connection after proxying them.")

	flag.Parse()

	if len(upstreams) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if len(logEvents) > 0 {
		eventSet := logger.NewEventSetFromCSV(logEvents)
		log.Events().CoalesceWith(eventSet)
	}

	reverseProxy := proxy.New().WithLogger(log)

	for _, upstream := range upstreams {
		log.Sync().Infof("upstream: %s", upstream)
		target, err := url.Parse(upstream)
		if err != nil {
			log.Sync().FatalExit(err)
		}

		proxyUpstream := proxy.NewUpstream(target).
			WithLogger(log).
			WithClose(close).
			WithBufferPool(proxy.NewBufferPool())

		reverseProxy.WithUpstream(proxyUpstream)
	}

	server := &http.Server{}
	server.Handler = reverseProxy
	server.Addr = bindAddr

	if len(tlsCert) > 0 && len(tlsKey) == 0 {
		log.Fatalf("`--tls-key` is unset, cannot continue")
	}
	if len(tlsCert) == 0 && len(tlsKey) > 0 {
		log.Fatalf("`--tls-cert` is unset, cannot continue")
	}

	if len(tlsCert) > 0 && len(tlsKey) > 0 {
		log.Sync().Infof("proxy using tls cert: %s", tlsCert)
		log.Sync().Infof("proxy using tls key: %s", tlsKey)
		log.Sync().Infof("proxy listening: %s", bindAddr)
		log.Fatal(server.ListenAndServeTLS(tlsCert, tlsKey))
	}

	log.Sync().Infof("proxy listening: %s", bindAddr)
	log.Fatal(server.ListenAndServe())
}

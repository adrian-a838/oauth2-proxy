package options

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/oauth2-proxy/oauth2-proxy/pkg/logger"
	"github.com/spf13/pflag"
)

type LegacyOptions struct {
	// Legacy options related to upstream servers
	LegacyUpstreams LegacyUpstreams `cfg:",squash"`

	Options Options `cfg:",squash"`
}

func NewLegacyOptions() *LegacyOptions {
	return &LegacyOptions{
		LegacyUpstreams: LegacyUpstreams{
			PassHostHeader:  true,
			ProxyWebSockets: true,
			FlushInterval:   DefaultFlushInterval,
			Timeout:         DefaultTimeout,
			KeepAlive:       DefaultKeepAlive,
		},

		Options: *NewOptions(),
	}
}

func (l *LegacyOptions) ToOptions() (*Options, error) {
	upstreams, err := l.LegacyUpstreams.convert()
	if err != nil {
		return nil, fmt.Errorf("error converting upstreams: %v", err)
	}
	l.Options.UpstreamServers = upstreams

	return &l.Options, nil
}

type LegacyUpstreams struct {
	FlushInterval                 time.Duration `flag:"flush-interval" cfg:"flush_interval"`
	Timeout                       time.Duration `flag:"timeout" cfg:"timeout"`
	KeepAlive                     time.Duration `flag:"keep-alive" cfg:"keep_alive"`
	PassHostHeader                bool          `flag:"pass-host-header" cfg:"pass_host_header"`
	ProxyWebSockets               bool          `flag:"proxy-websockets" cfg:"proxy_websockets"`
	SSLUpstreamInsecureSkipVerify bool          `flag:"ssl-upstream-insecure-skip-verify" cfg:"ssl_upstream_insecure_skip_verify"`
	Upstreams                     []string      `flag:"upstream" cfg:"upstreams"`
}

func legacyUpstreamsFlagSet() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("upstreams", pflag.ExitOnError)

	flagSet.Duration("flush-interval", DefaultFlushInterval, "period between response flushing when streaming responses")
	flagSet.Duration("timeout", DefaultTimeout, "the maximum amount of time a network dial will wait for a connect to complete")
	flagSet.Duration("keep-alive", DefaultKeepAlive, "the interval between keep-alive probes for an active network connection")
	flagSet.Bool("pass-host-header", true, "pass the request Host Header to upstream")
	flagSet.Bool("proxy-websockets", true, "enables WebSocket proxying")
	flagSet.Bool("ssl-upstream-insecure-skip-verify", false, "skip validation of certificates presented when using HTTPS upstreams")
	flagSet.StringSlice("upstream", []string{}, "the http url(s) of the upstream endpoint, file:// paths for static files or static://<status_code> for static response. Routing is based on the path")

	return flagSet
}

func (l *LegacyUpstreams) convert() (Upstreams, error) {
	upstreams := Upstreams{}

	for _, upstreamString := range l.Upstreams {
		u, err := url.Parse(upstreamString)
		if err != nil {
			return nil, fmt.Errorf("could not parse upstream %q: %v", upstreamString, err)
		}

		if u.Path == "" {
			u.Path = "/"
		}

		upstream := Upstream{
			ID:                    u.Path,
			Path:                  u.Path,
			URI:                   upstreamString,
			InsecureSkipTLSVerify: l.SSLUpstreamInsecureSkipVerify,
			PassHostHeader:        &l.PassHostHeader,
			ProxyWebSockets:       &l.ProxyWebSockets,
			FlushInterval:         &l.FlushInterval,
			Timeout:               &l.Timeout,
			KeepAlive:             &l.KeepAlive,
		}

		switch u.Scheme {
		case "file":
			if u.Fragment != "" {
				upstream.ID = u.Fragment
				upstream.Path = u.Fragment
			}
		case "static":
			responseCode, err := strconv.Atoi(u.Host)
			if err != nil {
				logger.Errorf("unable to convert %q to int, use default \"200\"", u.Host)
				responseCode = 200
			}
			upstream.Static = true
			upstream.StaticCode = &responseCode

			// These are not allowed to be empty and must be unique
			upstream.ID = upstreamString
			upstream.Path = upstreamString

			// Force defaults compatible with static responses
			upstream.URI = ""
			upstream.InsecureSkipTLSVerify = false
			upstream.PassHostHeader = nil
			upstream.ProxyWebSockets = nil
			upstream.FlushInterval = &DefaultFlushInterval
			upstream.Timeout = &DefaultTimeout
			upstream.KeepAlive = &DefaultKeepAlive
		}

		upstreams = append(upstreams, upstream)
	}

	return upstreams, nil
}

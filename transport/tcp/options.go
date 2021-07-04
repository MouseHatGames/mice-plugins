package tcp

import "time"

type Option func(opts *tcpOptions)

type tcpOptions struct {
	UseConnectionPooling  bool
	ConnectionIdleTimeout time.Duration
}

func ConnectionPooling(enabled bool) Option {
	return func(opts *tcpOptions) {
		opts.UseConnectionPooling = enabled
	}
}
func IdleTimeout(dur time.Duration) Option {
	return func(opts *tcpOptions) {
		opts.ConnectionIdleTimeout = dur
	}
}

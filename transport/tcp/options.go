package tcp

type Option func(opts *tcpOptions)

type tcpOptions struct {
	UseConnectionPooling bool
}

func ConnectionPooling(enabled bool) Option {
	return func(opts *tcpOptions) {
		opts.UseConnectionPooling = enabled
	}
}

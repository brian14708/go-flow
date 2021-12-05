package flow

type connectOption interface {
	apply(*connectOptions)
}

type connectOptions struct {
	allowInterfaceCast bool
}

type interfaceCastOpt struct{ EmptyConnectOption }

func (interfaceCastOpt) apply(o *connectOptions) {
	o.allowInterfaceCast = true
}

func WithInterfaceCast() ConnectOption {
	return interfaceCastOpt{}
}

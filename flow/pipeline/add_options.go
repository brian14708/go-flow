package pipeline

import (
	"github.com/brian14708/go-flow/flow"
)

var emptyAddOptions = new(addOptions)

type addOptions struct {
	sideInputs map[*Pipeline]sideInputDesc
}

type sideInputDesc struct {
	ports   []string
	options []flow.ConnectOption
}

func extractAddOptions(opts []flow.ConnectOption) (o *addOptions, rest []flow.ConnectOption) {
	for _, opt := range opts {
		if addOpt, ok := opt.(interface{ apply(*addOptions) }); ok {
			if o == nil {
				o = new(addOptions)
			}
			addOpt.apply(o)
		} else {
			rest = append(rest, opt)
		}
	}
	if o == nil {
		o = emptyAddOptions
	}
	return
}

func parsePortList(portArg interface{}) (ports []string) {
	switch arg := portArg.(type) {
	case string:
		ports = []string{arg}
	case []string:
		ports = arg
	default:
		panic("invalid port list name")
	}
	return
}

type addOptionFunc struct {
	flow.EmptyConnectOption
	fn func(o *addOptions)
}

func (fn addOptionFunc) apply(o *addOptions) {
	fn.fn(o)
}

func SideInput(src *Pipeline, portArg interface{}, extraOpts ...flow.ConnectOption) flow.ConnectOption {
	ports := parsePortList(portArg)
	return addOptionFunc{fn: func(o *addOptions) {
		if o.sideInputs == nil {
			o.sideInputs = make(map[*Pipeline]sideInputDesc)
		}
		if src == nil {
			if len(extraOpts) != 0 {
				panic("no options allowed for discarding SideInput")
			}
			prev := o.sideInputs[src]
			prev.ports = append(prev.ports, ports...)
			o.sideInputs[src] = prev
			return
		}
		o.sideInputs[src] = sideInputDesc{
			ports:   ports,
			options: extraOpts,
		}
	}}
}

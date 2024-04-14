package flow

import (
	"github.com/brian14708/go-flow/flow/channel"
	"github.com/brian14708/go-flow/flow/port"
	"github.com/brian14708/go-flow/flowtype"
)

// type aliases.
type (
	Chan       = channel.Channel
	Port       = port.Port
	PortMap    = port.Map
	AnyMessage = flowtype.AnyMessage

	// connect.
	ConnectOption      = channel.Option
	EmptyConnectOption = channel.EmptyOption
	WithChanSize       = channel.WithSize
)

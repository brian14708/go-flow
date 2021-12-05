package channel

import (
	"golang.org/x/time/rate"
)

var defaultOptions = options{
	size:      32,
	drainRate: rate.Limit(1),
}

type Option interface {
	apply(*options)
}

type Storage struct {
	options
}

type options struct {
	size         int
	interceptor  []func(interface{}) interface{}
	dropHandler  func(interface{})
	adaptiveGain float64
	drainRate    rate.Limit
}

type EmptyOption struct{}

func (EmptyOption) apply(*options) {}

type WithSize int

func (s WithSize) apply(o *options) {
	if s < 0 {
		panic("invalid channel size")
	}
	o.size = int(s)
}

type WithObserver func(interface{})

func (ob WithObserver) apply(o *options) {
	o.interceptor = append(o.interceptor, func(v interface{}) interface{} {
		ob(v)
		return v
	})
}

type WithInterceptor func(interface{}) interface{}

func (i WithInterceptor) apply(o *options) {
	o.interceptor = append(o.interceptor, i)
}

type WithAdaptiveGain float64

func (a WithAdaptiveGain) apply(o *options) {
	if a <= 0 {
		panic("invalid adaptive gain")
	}
	o.adaptiveGain = float64(a)
}

type WithDropHandler func(interface{})

func (d WithDropHandler) apply(o *options) {
	o.dropHandler = d
}

type WithDrainRate rate.Limit

func (d WithDrainRate) apply(o *options) {
	o.drainRate = rate.Limit(d)
}

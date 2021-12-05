package rrworker

type options struct {
	chanSize int
}

type Option interface {
	apply(*options)
}

type WithChanSize int

func (c WithChanSize) apply(opt *options) {
	opt.chanSize = int(c)
}

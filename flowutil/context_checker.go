package flowutil

import (
	"context"

	"github.com/kpango/fastime"
)

type ContextChecker struct {
	ctx     context.Context
	chkTime int64
}

func NewContextChecker(ctx context.Context) *ContextChecker {
	return &ContextChecker{ctx: ctx}
}

func (c *ContextChecker) Valid() bool {
	if c.chkTime < 0 {
		return false
	}

	now := fastime.UnixNanoNow()
	// check every 100 millisecond
	if now-c.chkTime > 100*1000 {
		c.chkTime = now
	} else {
		return true
	}

	ret := (c.ctx.Err() == nil)
	if !ret {
		c.chkTime = -1
	}
	return ret
}

func (c *ContextChecker) Err() error {
	return c.ctx.Err()
}

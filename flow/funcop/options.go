package funcop

import (
	"bytes"
	"fmt"
)

type options struct {
	parallel int
	ordered  bool
}

type FuncOption func(*options)

func WithParallel(p int, ordered bool) FuncOption {
	if p < 1 {
		panic("invalid parallel value")
	}
	return func(o *options) {
		o.parallel = p
		o.ordered = ordered
	}
}

func (o *options) String() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "Parallel: %d\n", o.parallel)
	fmt.Fprintf(buf, "Ordered: %v", o.ordered)
	return buf.String()
}

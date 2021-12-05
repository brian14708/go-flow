package channel

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAssignable(t *testing.T) {
	tests := []struct {
		success bool
		typ     interface{}
		dir     reflect.ChanDir
		elem    reflect.Type
	}{
		{false, (*chan int)(nil), reflect.BothDir, nil},
		{false, (*chan<- int)(nil), reflect.RecvDir, nil},
		{true, (*chan int)(nil), reflect.RecvDir, nil},
		{true, (*<-chan int)(nil), reflect.RecvDir, nil},
		{false, (*<-chan int)(nil), reflect.RecvDir, reflect.TypeOf("")},
		{true, (*chan<- int)(nil), reflect.SendDir, nil},
	}

	for _, test := range tests {
		_, err := IsAssignable(test.dir, test.elem, reflect.TypeOf(test.typ))
		if test.success {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func BenchmarkIsAssignable(b *testing.B) {
	var c chan int
	for i := 0; i < b.N; i++ {
		_, _ = IsAssignable(reflect.RecvDir, nil, reflect.PtrTo(reflect.TypeOf(c)))
	}
}

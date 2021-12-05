package funcx

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatch(t *testing.T) {
	testcase := [...]struct {
		f         interface{}
		templates *Template
		idx       int
		matches   map[TypeWildcard]reflect.Type
		success   bool
	}{
		{
			123,
			MustNewTemplate(
				(func(int) error)(nil),
			),
			-1, nil, false,
		},
		{
			(func() error)(nil),
			MustNewTemplate(
				(func(int) error)(nil),
			),
			-1, nil, false,
		},

		{
			(func(int) error)(nil),
			MustNewTemplate(
				(func(int))(nil),
				(func() error)(nil),
				(func(error) int)(nil),
				(func(int) error)(nil),
			),
			3, nil, true,
		},
		{
			(func(int) error)(nil),
			MustNewTemplate(
				(func(T0))(nil),
				(func(T0) T0)(nil),
				(func(T0) T1)(nil),
			),
			2,
			map[TypeWildcard]reflect.Type{
				T0{}: reflect.TypeOf(0),
				T1{}: reflect.TypeOf((*error)(nil)).Elem(),
			},
			true,
		},
		{
			(func(int, int, int, int) (int, int, int, int))(nil),
			MustNewTemplate(
				(func(T0, T1, T2, T3) (T4, T5, T6, T7))(nil),
			),
			0,
			map[TypeWildcard]reflect.Type{
				T0{}: reflect.TypeOf(0),
				T1{}: reflect.TypeOf(0),
				T2{}: reflect.TypeOf(0),
				T3{}: reflect.TypeOf(0),
				T4{}: reflect.TypeOf(0),
				T5{}: reflect.TypeOf(0),
				T6{}: reflect.TypeOf(0),
				T7{}: reflect.TypeOf(0),
			},
			true,
		},
		// map
		{
			(func(map[int]float64) int)(nil),
			MustNewTemplate(
				(func(map[float64]T1) T2)(nil),
				(func(map[T0]int) T2)(nil),
				(func(map[T0]T1) T2)(nil),
			),
			2,
			map[TypeWildcard]reflect.Type{
				T0{}: reflect.TypeOf(0),
				T1{}: reflect.TypeOf(0.0),
				T2{}: reflect.TypeOf(0),
			},
			true,
		},
		// chan
		{
			(func(chan<- int) int)(nil),
			MustNewTemplate(
				(func(chan T0) T0)(nil),
				(func(<-chan T0) T0)(nil),
				(func(chan<- T0) T0)(nil),
			),
			2,
			map[TypeWildcard]reflect.Type{
				T0{}: reflect.TypeOf(0),
			},
			true,
		},
		// array
		{
			(func([4]int) int)(nil),
			MustNewTemplate(
				(func([]int) T0)(nil),
				(func(*T0) int)(nil),
				(func([3]T0) int)(nil),
				(func([4]T0) int)(nil),
			),
			3,
			map[TypeWildcard]reflect.Type{
				T0{}: reflect.TypeOf(0),
			},
			true,
		},
		// slice
		{
			(func([]int) int)(nil),
			MustNewTemplate(
				(func([]float32) T0)(nil),
				(func(*T0) int)(nil),
				(func([]T0) int)(nil),
			),
			2,
			map[TypeWildcard]reflect.Type{
				T0{}: reflect.TypeOf(0),
			},
			true,
		},
		// variadic
		{
			(func(float32, ...interface{}) int)(nil),
			MustNewTemplate(
				(func(float32, T0) T1)(nil),
				(func(float32, []T0) T1)(nil),
				(func(float32, ...T0) T1)(nil),
			),
			2,
			map[TypeWildcard]reflect.Type{
				T0{}: reflect.TypeOf((*interface{})(nil)).Elem(),
				T1{}: reflect.TypeOf(1),
			},
			true,
		},
		// raw types
		{
			make(chan int),
			MustNewTemplate(
				(chan T0)(nil),
			),
			0,
			map[TypeWildcard]reflect.Type{
				T0{}: reflect.TypeOf(0),
			},
			true,
		},
		{
			map[string]int{},
			MustNewTemplate(
				([]T1)(nil),
				(map[T0]T1)(nil),
			),
			1,
			map[TypeWildcard]reflect.Type{
				T0{}: reflect.TypeOf(""),
				T1{}: reflect.TypeOf(0),
			},
			true,
		},
	}
	for _, testcase := range testcase {
		idx, m, err := testcase.templates.MatchValue(testcase.f)
		if testcase.success {
			assert.NoError(t, err)
			assert.Equal(t, testcase.idx, idx)
			if testcase.matches == nil {
				testcase.matches = make(map[TypeWildcard]reflect.Type)
			}
			assert.Equal(t, testcase.matches[T0{}], m.Get(T0{}))
			assert.Equal(t, testcase.matches[T1{}], m.Get(T1{}))
			assert.Equal(t, testcase.matches[T2{}], m.Get(T2{}))
			assert.Equal(t, testcase.matches[T3{}], m.Get(T3{}))
			assert.Equal(t, testcase.matches[T4{}], m.Get(T4{}))
			assert.Equal(t, testcase.matches[T5{}], m.Get(T5{}))
			assert.Equal(t, testcase.matches[T6{}], m.Get(T6{}))
			assert.Equal(t, testcase.matches[T7{}], m.Get(T7{}))
		} else {
			assert.Error(t, err)
			assert.Equal(t, -1, idx)
		}
	}
}

func TestMatchError(t *testing.T) {
	_, _, err := MustNewTemplate(
		(func([]float32) T0)(nil),
		(func(*T0) int)(nil),
		(func([]T0) T0)(nil),
	).MatchValue(
		(func([]int) float32)(nil),
	)
	assert.Contains(t, err.Error(), "funcx.T0")

	_, _, err = MustNewTemplate(
		(func([]float32) T0)(nil),
	).MatchValue(
		(func([]int) float32)(nil),
	)
	assert.Contains(t, err.Error(), "funcx.T0")
}

func TestWildcardName(t *testing.T) {
	s := fmt.Sprintf("%v", map[TypeWildcard]reflect.Type{
		T0{}: reflect.TypeOf(0),
		T1{}: reflect.TypeOf(0),
		T2{}: reflect.TypeOf(0),
		T3{}: reflect.TypeOf(0),
		T4{}: reflect.TypeOf(0),
		T5{}: reflect.TypeOf(0),
		T6{}: reflect.TypeOf(0),
		T7{}: reflect.TypeOf(0),
	})
	assert.Contains(t, s, "T7")
}

func BenchmarkMatch(b *testing.B) {
	t := MustNewTemplate(
		(func([]float32) T0)(nil),
		(func(*T0) int)(nil),
		(func([]T0) T0)(nil),
		(chan int)(nil),
	)

	for i := 0; i < b.N; i++ {
		_, x, _ := t.MatchValue(
			(func([]float32) float32)(nil),
		)
		x.Free()
	}
}

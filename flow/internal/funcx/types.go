package funcx

import "reflect"

type TypeWildcard interface {
	wildcard() int
}

type (
	T0 struct{}
	T1 struct{}
	T2 struct{}
	T3 struct{}
	T4 struct{}
	T5 struct{}
	T6 struct{}
	T7 struct{}
)

func (T0) wildcard() int { return 0 }
func (T1) wildcard() int { return 1 }
func (T2) wildcard() int { return 2 }
func (T3) wildcard() int { return 3 }
func (T4) wildcard() int { return 4 }
func (T5) wildcard() int { return 5 }
func (T6) wildcard() int { return 6 }
func (T7) wildcard() int { return 7 }

func (T0) String() string { return "T0" }
func (T1) String() string { return "T1" }
func (T2) String() string { return "T2" }
func (T3) String() string { return "T3" }
func (T4) String() string { return "T4" }
func (T5) String() string { return "T5" }
func (T6) String() string { return "T6" }
func (T7) String() string { return "T7" }

var wildcardCache = map[reflect.Type]int{
	reflect.TypeOf(T0{}): T0{}.wildcard(),
	reflect.TypeOf(T1{}): T1{}.wildcard(),
	reflect.TypeOf(T2{}): T2{}.wildcard(),
	reflect.TypeOf(T3{}): T3{}.wildcard(),
	reflect.TypeOf(T4{}): T4{}.wildcard(),
	reflect.TypeOf(T5{}): T5{}.wildcard(),
	reflect.TypeOf(T6{}): T6{}.wildcard(),
	reflect.TypeOf(T7{}): T7{}.wildcard(),
}

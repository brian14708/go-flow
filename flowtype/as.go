package flowtype

import "reflect"

func As(base, target interface{}) bool {
	if target == nil {
		panic("convert target must not be nil")
	}

	val := reflect.ValueOf(target)
	t := val.Type()
	if t.Kind() != reflect.Ptr || val.IsNil() {
		panic("convrt target must be a non-nil pointer")
	}
	t = t.Elem()
	if t.Kind() != reflect.Interface {
		panic("convert *target must be interface")
	}

	if base != nil {
		if reflect.TypeOf(base).AssignableTo(t) {
			val.Elem().Set(reflect.ValueOf(base))
			return true
		}
		if w, ok := base.(interface{ As(interface{}) bool }); ok && w.As(target) {
			return true
		}
	}
	return false
}

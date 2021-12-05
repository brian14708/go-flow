package channel

import (
	"fmt"
	"reflect"
)

func IsAssignable(d reflect.ChanDir, elemType reflect.Type, t reflect.Type) (storageType reflect.Type, err error) {
	if t.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("type `%s' is not pointer", t)
	}
	t = t.Elem()

	switch d {
	case reflect.RecvDir:
		if t.Kind() != reflect.Chan || (t.ChanDir()&reflect.RecvDir) == 0 {
			return nil, fmt.Errorf("expect channel type `<-chan T' got `%s'", t)
		}
		storageType = t.Elem()
	case reflect.SendDir:
		if t.Kind() != reflect.Chan || (t.ChanDir()&reflect.SendDir) == 0 {
			return nil, fmt.Errorf("expect channel type `chan<- T' got `%s'", t)
		}
		storageType = t.Elem()
	default:
		return nil, fmt.Errorf("invalid chan direction")
	}

	if elemType != nil {
		if !elemType.AssignableTo(storageType) {
			return nil, fmt.Errorf(
				"runtime type `%s' is not assignable to compiled type `%s'",
				elemType,
				t.Elem(),
			)
		}
	}
	return storageType, nil
}

func AssignableDir(elemType reflect.Type, t reflect.Type) (dir reflect.ChanDir, storageType reflect.Type) {
	if t.Kind() != reflect.Ptr {
		return
	}
	t = t.Elem()

	if t.Kind() != reflect.Chan {
		return
	}

	s := t.Elem()
	if elemType != nil && !elemType.AssignableTo(s) {
		return
	}

	d := t.ChanDir()
	if (d & reflect.RecvDir) != 0 {
		dir |= reflect.RecvDir
	}
	if (d & reflect.SendDir) != 0 {
		dir |= reflect.SendDir
	}
	if dir != 0 {
		storageType = s
	}

	return
}

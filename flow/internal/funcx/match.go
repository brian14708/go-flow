package funcx

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sync"
)

var pool = sync.Pool{
	New: func() interface{} {
		tmp := new(struct {
			obj MatchResult
			buf [2]resultEntry
		})
		tmp.obj.t = tmp.buf[:0]
		return &tmp.obj
	},
}

type (
	resultEntry struct {
		wildcard int
		replace  reflect.Type
	}
	MatchResult struct {
		t []resultEntry
	}
)

func (m *MatchResult) Get(t TypeWildcard) reflect.Type {
	id := t.wildcard()
	for _, m := range m.t {
		if m.wildcard == id {
			return m.replace
		}
	}
	return nil
}

func (m *MatchResult) Free() {
	pool.Put(m)
}

func insertMatch(r []resultEntry, template, value reflect.Type) ([]resultEntry, bool) {
	if w, ok := wildcardCache[template]; ok {
		for _, prev := range r {
			if prev.wildcard == w {
				return r, prev.replace == value
			}
		}
		return append(r, resultEntry{w, value}), true
	}

	if template.Kind() != value.Kind() {
		return r, false
	}
	switch template.Kind() {
	case reflect.Chan:
		if template.ChanDir() != value.ChanDir() {
			return r, false
		}
		return insertMatch(r, template.Elem(), value.Elem())
	case reflect.Map:
		r, ok := insertMatch(r, template.Key(), value.Key())
		if !ok {
			return r, false
		}
		return insertMatch(r, template.Elem(), value.Elem())
	case reflect.Array:
		if template.Len() != value.Len() {
			return r, false
		}
		return insertMatch(r, template.Elem(), value.Elem())
	case reflect.Slice, reflect.Ptr:
		return insertMatch(r, template.Elem(), value.Elem())
	case reflect.Func:
		if template.NumIn() != value.NumIn() {
			return r, false
		}
		if template.NumOut() != value.NumOut() {
			return r, false
		}
		if template.IsVariadic() != value.IsVariadic() {
			return r, false
		}

		var ok bool
		for i := 0; i < template.NumIn(); i++ {
			r, ok = insertMatch(r, template.In(i), value.In(i))
			if !ok {
				return r, false
			}
		}
		for i := 0; i < template.NumOut(); i++ {
			r, ok = insertMatch(r, template.Out(i), value.Out(i))
			if !ok {
				return r, false
			}
		}
		return r, true
	}
	return r, template == value
}

func (tpl *Template) TryMatch(t reflect.Type) (int, *MatchResult) {
	h := typeHash(t)
	if _, ok := tpl.hashes[h]; !ok {
		return -1, nil
	}

	result := pool.Get().(*MatchResult)
	for idx, typ := range tpl.types {
		if h != typeHash(typ) {
			continue
		}
		var match bool
		result.t, match = insertMatch(result.t[:0], typ, t)
		if match {
			return idx, result
		}
	}
	result.Free()
	return -1, nil
}

func (tpl *Template) Match(t reflect.Type) (int, *MatchResult, error) {
	idx, m := tpl.TryMatch(t)
	if idx < 0 {
		return -1, nil, matchErr{t, tpl}
	}
	return idx, m, nil
}

func (tpl *Template) MatchValue(fn interface{}) (int, *MatchResult, error) {
	return tpl.Match(reflect.TypeOf(fn))
}

type hash int32

func typeHash(t reflect.Type) hash {
	k := t.Kind()

	var h uint32
	if k == reflect.Func {
		h = uint32(t.NumIn()<<8 | t.NumOut())
	} else {
		h = uint32(t.Size())
	}
	return hash(h | (uint32(k) << 24))
}

type Template struct {
	types  []reflect.Type
	hashes map[hash]struct{}
}

func NewTemplate(templates ...interface{}) (*Template, error) {
	types := make([]reflect.Type, len(templates))
	hashes := make(map[hash]struct{})
	for i, t := range templates {
		t := reflect.TypeOf(t)
		types[i] = t
		hashes[typeHash(t)] = struct{}{}
	}
	return &Template{
		types:  types,
		hashes: hashes,
	}, nil
}

func MustNewTemplate(args ...interface{}) *Template {
	t, err := NewTemplate(args...)
	if err != nil {
		panic("new function template failed: " + err.Error())
	}
	return t
}

type matchErr struct {
	have reflect.Type
	want *Template
}

func (e matchErr) Error() string {
	b := new(bytes.Buffer)
	fmt.Fprintf(b, "no type match found, `%s', expected", e.have)
	if len(e.want.types) == 1 {
		fmt.Fprintf(b, " `%s'", e.want.types[0])
	} else {
		_, _ = io.WriteString(b, ":")
		for _, w := range e.want.types {
			fmt.Fprintf(b, "\n\t* %s", w)
		}
	}
	return b.String()
}

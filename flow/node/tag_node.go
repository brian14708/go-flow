package node

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	lru "github.com/hashicorp/golang-lru"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/channel"
	"github.com/brian14708/go-flow/flow/port"
	"github.com/brian14708/go-flow/flowtype"
)

const (
	maxDepth      = 4
	typeCacheSize = 1024
)

var (
	typeCache     *lru.ARCCache
	typeCacheInit sync.Once
)

type tagNode interface {
	Run(context.Context) error
}

type tagNodeImpl struct {
	tagNode
	tag string
}

func (t *tagNodeImpl) Ports() (in, out flow.PortMap) {
	in, out = port.MakeMap(), port.MakeMap()
	_ = getTagPorts(t.tag, reflect.ValueOf(t.tagNode), in, out)
	return
}

func (t *tagNodeImpl) NodeHash() interface{} {
	return t.tagNode
}

func (t *tagNodeImpl) As(i interface{}) bool {
	return flowtype.As(t.tagNode, i)
}

func (t *tagNodeImpl) NodeType() string {
	var n interface{ NodeType() string }
	if t.As(&n) {
		return n.NodeType()
	}
	return reflect.TypeOf(t.tagNode).String()
}

func NewTagNode(n interface{}, tag string) (flow.NodeWrapper, error) {
	node, ok := n.(tagNode)
	if !ok {
		return nil, errors.New("TagNode should implement method `Run(context.Context) error'")
	}

	in, out := port.MakeMap(), port.MakeMap()
	defer func() {
		port.RecycleMap(in)
		port.RecycleMap(out)
	}()
	if err := getTagPorts(tag, reflect.ValueOf(n), in, out); err != nil {
		return nil, fmt.Errorf("fail to make TagNode: %w", err)
	}

	if len(in) == 0 && len(out) == 0 {
		return nil, fmt.Errorf("TagNode should have at least one port")
	}
	return &tagNodeImpl{
		tagNode: node,
		tag:     tag,
	}, nil
}

func getTagPorts(tagName string, val reflect.Value, in, out flow.PortMap) error {
	typeCacheInit.Do(func() {
		var err error
		typeCache, err = lru.NewARC(typeCacheSize)
		if err != nil {
			panic("failed to init type typeCache: " + err.Error())
		}
	})

	val, err := resolvePointer(val)
	if err != nil {
		return err
	}

	if cache, ok := typeCache.Get(val.Type()); ok {
		for tag, idx := range cache.(map[string][]int) {
			if idx[len(idx)-1] == -1 {
				v, err := fieldByIndex(val, idx[:len(idx)-1])
				if err != nil {
					return err
				}
				in[tag] = v.Addr().Interface()
			} else {
				v, err := fieldByIndex(val, idx)
				if err != nil {
					return err
				}
				out[tag] = v.Addr().Interface()
			}
		}
		return nil
	}

	var path [maxDepth]int
	cache := make(map[string][]int)
	err = getTagPortsImpl(tagName, val, in, out, path[:0], cache)
	if err == nil {
		typeCache.Add(val.Type(), cache)
	}
	return err
}

func getTagPortsImpl(
	tagName string, val reflect.Value, in, out flow.PortMap,
	path []int, cache map[string][]int,
) error {
	if len(path) >= maxDepth {
		return nil
	}

	st := val.Type()
	if st.Kind() != reflect.Struct {
		return nil
	}

	var buffer []int
	for i := 0; i < st.NumField(); i++ {
		ft := st.Field(i)
		if ft.Anonymous {
			subVal, err := resolvePointer(val.Field(i))
			if err != nil {
				return err
			}
			err = getTagPortsImpl(tagName, subVal, in, out, append(path, i), cache)
			if err != nil {
				return err
			}
			continue
		}

		tag := ft.Tag.Get(tagName)
		if tag == "" {
			continue
		}
		if _, ok := cache[tag]; ok {
			return fmt.Errorf("channel `%s' already exists", tag)
		}

		dir, _ := channel.AssignableDir(nil, reflect.PtrTo(ft.Type))
		switch dir {
		case reflect.BothDir:
			return fmt.Errorf("bidirectional port unsupported")
		case reflect.SendDir:
			out[tag] = val.Field(i).Addr().Interface()
		case reflect.RecvDir:
			in[tag] = val.Field(i).Addr().Interface()
		default:
			return fmt.Errorf("invalid port type `%s'", ft.Type)
		}

		if len(buffer) < len(path)+1 {
			buffer = make([]int, 2*maxDepth)
		}
		var n int
		if dir == reflect.SendDir {
			n = copy(buffer, append(path, i))
		} else {
			n = copy(buffer, append(path, i, -1))
		}
		cache[tag], buffer = buffer[0:n], buffer[n:]
	}
	return nil
}

func resolvePointer(val reflect.Value) (reflect.Value, error) {
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return reflect.Value{}, errors.New("nil pointer to embedded struct")
		}
		val = val.Elem()
	}
	return val, nil
}

func fieldByIndex(v reflect.Value, index []int) (reflect.Value, error) {
	for _, x := range index {
		var err error
		v, err = resolvePointer(v)
		if err != nil {
			return reflect.Value{}, err
		}
		v = v.Field(x)
	}
	return v, nil
}

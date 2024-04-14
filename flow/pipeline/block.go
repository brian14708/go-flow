package pipeline

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/internal/funcx"
	"github.com/brian14708/go-flow/flow/node"
)

type block struct {
	in, out []string
}

var blockMakerTemplate = funcx.MustNewTemplate(
	(func() funcx.T0)(nil),
	(func() (funcx.T0, error))(nil),
)

func (p *Pipeline) addBlock(name string, n interface{}) (*block, error) {
	if fn, ok := n.(func(*Pipeline)); ok {
		baseName := strings.TrimPrefix(name, p.prefix)
		ppl := p.SubPipeline(baseName)
		fn(ppl)
		n = ppl
	}

	switch val := n.(type) {
	case func(*Pipeline):
		baseName := strings.TrimPrefix(name, p.prefix)
		ppl := p.SubPipeline(baseName)
		val(ppl)
		return p.addBlock(name, ppl)
	case func(*Pipeline) error:
		baseName := strings.TrimPrefix(name, p.prefix)
		ppl := p.SubPipeline(baseName)
		if err := val(ppl); err != nil {
			return nil, err
		}
		return p.addBlock(name, ppl)
	case *Pipeline:
		return newPipelineBlock(val, p.g)
	case flow.Node:
		return newNodeBlock(val, p.g, name)
	case *block:
		return val, nil
	case interface{ Run(context.Context) error }:
		if tn, err := node.NewTagNode(n, "pipeline"); err == nil {
			return newNodeBlock(tn, p.g, name)
		}
	}

	t := reflect.TypeOf(n)
	switch t.Kind() {
	case reflect.Chan:
		if ch, err := node.NewChanNode(n); err == nil {
			return newNodeBlock(ch, p.g, name)
		}
	case reflect.Func:
		if idx, types := blockMakerTemplate.TryMatch(t); idx >= 0 {
			types.Free()
			ret := reflect.ValueOf(n).Call(nil)
			if idx == 1 {
				if err := ret[1]; !err.IsNil() {
					return nil, err.Interface().(error)
				}
			}
			return p.addBlock(name, ret[0].Interface())
		}

		if fn, err := node.NewFuncNode(n); err != nil {
			return nil, err
		} else {
			return newNodeBlock(fn, p.g, name)
		}
	case reflect.Slice, reflect.Array:
		val := reflect.ValueOf(n)
		blks := make([]*block, val.Len())
		for i := range blks {
			blk, err := p.addBlock(
				name+"_"+strconv.Itoa(i),
				val.Index(i).Interface(),
			)
			if err != nil {
				return nil, err
			}
			blks[i] = blk
		}
		if len(blks) > 0 {
			return newParallelBlock(blks), nil
		}
	}
	return nil, fmt.Errorf("invalid block type `%s'", t)
}

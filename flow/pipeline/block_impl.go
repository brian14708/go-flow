package pipeline

import (
	"errors"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/port"
)

var discardBlock = new(block)

func newBlock(in, out []string) *block {
	if len(in) == 0 && len(out) == 0 {
		return discardBlock
	}
	return &block{in, out}
}

func newParallelBlock(blks []*block) *block {
	inCnt, cnt := 0, 0
	for _, buf := range blks {
		cnt += len(buf.in) + len(buf.out)
		inCnt += len(buf.in)
	}

	tmp := make([]string, 0, cnt)
	for _, buf := range blks {
		tmp = append(tmp, buf.in...)
	}
	for _, buf := range blks {
		tmp = append(tmp, buf.out...)
	}
	return newBlock(tmp[:inCnt], tmp[inCnt:])
}

func newNodeBlock(node flow.Node, g *flow.Graph, name string) (*block, error) {
	err := g.AddNode(name, node)
	if err != nil {
		return nil, err
	}

	in, out := node.Ports()
	obj := new(struct {
		block
		buf [2]string
	})
	buf := obj.buf[:0]
	for i := range in {
		buf = append(buf, name+":"+i)
	}
	for o := range out {
		buf = append(buf, name+":"+o)
	}

	blk := &obj.block
	blk.in = buf[:len(in)]
	blk.out = buf[len(in):]

	port.RecycleMap(in)
	port.RecycleMap(out)
	return blk, nil
}

func newPipelineBlock(ppl *Pipeline, g *flow.Graph) (*block, error) {
	if ppl.first == nil {
		return nil, errors.New("cannot make block from empty pipeline")
	}
	if ppl.g != g {
		return nil, errors.New("can only add sub-pipeline to pipeline")
	}
	blk := newBlock(ppl.first.in, ppl.last.out)
	ppl.IgnoreInput().IgnoreOutput()
	return blk, nil
}

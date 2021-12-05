package codegen

import (
	"bytes"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	testcase := [...]struct {
		success bool
		opt     GenerateOptions
	}{
		{true, GenerateOptions{}},
		{false, GenerateOptions{
			PkgName:   "testgen",
			TypeNames: []string{"/a"},
		}},
		{true, GenerateOptions{
			PkgName: "testgen",
			Builtin: true,
		}},
		{true, GenerateOptions{
			PkgName: "testgen",
			TypeNames: []string{
				"*Decoder",
				"encoding/json.Decoder",
				"encoding/json.*Decoder",
			},
		}},
	}
	for _, test := range testcase {
		_, err := Generate(test.opt)
		if test.success {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
		_, err = GenerateTest(test.opt)
		if test.success {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestMakeFuncCaller(t *testing.T) {
	s, err := makeFuncCaller([]caller{
		{nil, nil},
		{[]jen.Code{jen.Int()}, nil},
		{nil, []jen.Code{jen.Int()}},
		{[]jen.Code{jen.Int(), jen.Int()}, nil},
		{nil, []jen.Code{jen.Int(), jen.Int()}},
		{[]jen.Code{jen.Int(), jen.Int()}, []jen.Code{jen.Int(), jen.Int()}},
	}, jen.Id("fn"))
	assert.NoError(t, err)

	assert.NoError(t, s.(*jen.Statement).Render(new(bytes.Buffer)))
}

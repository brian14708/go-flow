package codegen

import (
	"bytes"
	"fmt"
	"strings"

	. "github.com/dave/jennifer/jen"

	"github.com/brian14708/go-flow/flowtype"
)

const pkg = "github.com/brian14708/go-flow/flowtype"

type GenerateOptions struct {
	PkgName   string
	Builtin   bool
	TypeNames []string
}

type typeSpec struct {
	name     string
	typename Code
}

func makeTypeSpecs(opt GenerateOptions) []typeSpec {
	var types []typeSpec
	if opt.Builtin {
		types = append(types, []typeSpec{
			{"builtinBool", Bool()},
			{"builtinComplex64", Complex64()},
			{"builtinComplex128", Complex128()},
			{"builtinError", Error()},
			{"builtinFloat32", Float32()},
			{"builtinFloat64", Float64()},
			{"builtinInt", Int()},
			{"builtinInt8", Int8()},
			{"builtinInt16", Int16()},
			{"builtinInt32", Int32()},
			{"builtinInt64", Int64()},
			{"builtinString", String()},
			{"builtinUint", Uint()},
			{"builtinUint8", Uint8()},
			{"builtinUint16", Uint16()},
			{"builtinUint32", Uint32()},
			{"builtinUint64", Uint64()},
			{"builtinUintptr", Uintptr()},
			{"builtinInterface", Interface()},
			// rune and byte are aliases for int32 and uint8
		}...)
	}

	for _, n := range opt.TypeNames {
		parts := strings.Split(n, ".")
		pkg := strings.Join(parts[:len(parts)-1], ".")
		name := strings.TrimLeft(parts[len(parts)-1], "*")
		t := typeSpec{
			name: "dispatch" + name,
		}

		if pkg == "" {
			t.typename = Id(name)
		} else {
			t.typename = Qual(pkg, name)
		}
		for _, ch := range parts[len(parts)-1] {
			if ch == '*' {
				t.typename = Op("*").Add(t.typename)
			} else {
				break
			}
		}
		types = append(types, t)
	}
	return types
}

func Generate(opt GenerateOptions) ([]byte, error) {
	types := makeTypeSpecs(opt)

	f := NewFilePath(opt.PkgName)
	f.HeaderComment("//go:build !skipflowtype")
	f.HeaderComment("// +build !skipflowtype")
	f.PackageComment(fmt.Sprintf("Code generated by %s/codegen. DO NOT EDIT.", pkg))
	for _, entry := range types {
		fields, err := makeFields(entry.typename, entry.name)
		if err != nil {
			return nil, err
		}
		f.Var().Id(entry.name).Op("=").Op("&").Qual(pkg, "DispatchTable").Values(
			fields,
		)
		f.Line()
	}
	f.Func().Id("init").Params().BlockFunc(func(g *Group) {
		for _, entry := range types {
			g.Qual(pkg, "MustRegisterDispatch").Call(
				Qual("reflect", "TypeOf").Call(
					Parens(Op("*").Add(entry.typename)).Parens(Nil()),
				).Op(".").Id("Elem").Call(),
				Id(entry.name),
			)
		}
	})

	buf := new(bytes.Buffer)
	if err := f.Render(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func GenerateTest(opt GenerateOptions) ([]byte, error) {
	types := makeTypeSpecs(opt)

	f := NewFilePath(opt.PkgName)
	f.HeaderComment("//go:build !skipflowtype")
	f.HeaderComment("// +build !skipflowtype")
	f.PackageComment(fmt.Sprintf("Code generated by %s/codegen. DO NOT EDIT.", pkg))

	for _, entry := range types {
		f.Func().Id("Test" + strings.Title(entry.name)).Params(
			Id("t").Op("*").Qual("testing", "T"),
		).Block(
			Qual(pkg+"/testutil", "TestWithNilRegistry").Call(
				Id("t"),
				Func().Params(
					Id("t").Op("*").Qual("testing", "T"),
				).Block(
					Var().Id("arg").Add(entry.typename),
					Id("cnt").Op(":=").Lit(0),
					Id("rtype").Op(":=").Qual("reflect", "TypeOf").Call(
						Op("&").Id("arg"),
					).Op(".").Id("Elem").Call(),
					Qual(pkg+"/testutil", "RunChanTest").Call(
						Id("t"),
						Id("rtype"),
					),
					BlockFunc(func(g *Group) {
						for _, f := range listFunc(entry.typename) {
							g.Qual(pkg+"/testutil", "RunFuncCallerTest").Call(
								Id("t"),
								Id("rtype"),
								Op("&").Id("cnt"),
								Func().Params(f.In...).ParamsFunc(
									func(g *Group) {
										for _, o := range f.Out {
											g.Id("_").Add(o)
										}
									},
								).Block(
									Id("cnt").Op("++"),
									Return(),
								),
								Id("arg"),
							)
						}
					}),
				),
			),
		)
		f.Line()
	}

	buf := new(bytes.Buffer)
	if err := f.Render(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type caller struct {
	In, Out []Code
}

func makeFuncCaller(funcs []caller, v Code) (Code, error) {
	var callBlks []Code

	dedup := map[string]struct{}{}
	for _, f := range funcs {
		b := bytes.Buffer{}
		err := Func().Id("_").Params(f.In...).Params(f.Out...).Render(&b)
		if err != nil {
			return nil, err
		}
		s := b.String()
		if _, ok := dedup[s]; ok {
			continue
		}
		dedup[s] = struct{}{}

		block := []Code{
			If(Lit(len(f.In)).Op("!=").Len(Id("args"))).Block(
				Panic(Lit("wrong number of arguments")),
			),
		}
		var args []Code
		for i, in := range f.In {
			n := fmt.Sprintf("i%d", i)
			args = append(args, Id(n))
			block = append(block,
				Var().Id(n).Add(in),
				If(Id("i").Op(":=").Id("args").Index(Lit(i)).Op(";").Id("i").Op("!=").Nil()).Block(
					Id(n).Op("=").Id("i").Op(".").Parens(in),
				),
			)
		}

		var outs *Statement
		for i := range f.Out {
			if outs == nil {
				outs = Id(fmt.Sprintf("o%d", i))
			} else {
				outs.Add(Op(",").Id(fmt.Sprintf("o%d", i)))
			}
		}
		if outs != nil {
			block = append(block,
				Add(outs).Op(":=").Add(v).Call(args...),
				Return(Id("append").Call(Id("ret").Op(",").Add(outs))),
			)
		} else {
			block = append(block,
				Add(v).Call(args...),
				Return(Nil()),
			)
		}

		fn := Func().Params(f.In...).Params(f.Out...)
		callBlks = append(callBlks,
			Case(fn),
			Return(
				Func().Params(Id("args").Op(",").Id("ret").Index().Interface()).Index().Interface().Block(
					block...,
				),
			),
		)
	}

	return Switch(Id("fn").Op(":=").Id("fn").Op(".").Parens(Type())).Block(
		callBlks...,
	), nil
}

func listFunc(typename Code) []caller {
	return []caller{
		{[]Code{typename}, []Code{typename}},
		{[]Code{typename}, []Code{Bool()}},
	}
}

func makeFields(typename Code, name string) (Dict, error) {
	caller, err := makeFuncCaller(listFunc(typename), Id("fn"))
	if err != nil {
		return nil, err
	}
	return Dict{
		Id("Version"): Lit(flowtype.DispatchVersion),
		Id("FuncCaller"): Func().Params(Id("fn").Interface()).
			Qual(pkg, "CallFunc").
			Block(
				caller,
				Return(Qual(pkg, "GenericDispatch").Op(".").Id("FuncCaller").Call(Id("fn"))),
			),

		Id("ChanSender"): Func().Params(Id("c").Interface()).
			Qual(pkg, "SendFunc").
			Block(
				Var().Id("ch").Chan().Op("<-").Add(typename),
				Switch(Id("c").Op(":=").Id("c").Op(".").Parens(Type())).Block(
					Case(Chan().Op("<-").Add(typename)),
					Id("ch").Op("=").Id("c"),
					Case(Chan().Add(typename)),
					Id("ch").Op("=").Id("c"),
					Default(),
					Return(Qual(pkg, "GenericDispatch").Op(".").Id("ChanSender").Call(Id("c"))),
				),
				Return(
					Func().Params(Id("v").Interface(), Id("cancel").Op("<-").Chan().Struct(), Id("block").Bool()).Bool().Block(
						Var().Id("el").Add(typename),
						If(Id("v").Op("!=").Nil()).Block(
							Id("el").Op("=").Id("v").Op(".").Parens(typename),
						),
						If(Op("!").Id("block")).Block(
							Select().Block(
								Case(Id("ch").Op("<-").Id("el")),
								Default(),
								Return(False()),
							),
						).Else().If(Id("cancel").Op("==").Nil()).Block(
							Id("ch").Op("<-").Id("el"),
						).Else().Block(
							Select().Block(
								Case(Id("ch").Op("<-").Id("el")),
								Case(Op("<-").Id("cancel")),
								Return(False()),
							),
						),
						Return(True()),
					),
				),
			),
		Id("ChanRecver"): Func().Params(Id("c").Interface()).
			Qual(pkg, "RecvFunc").
			Block(
				Var().Id("ch").Op("<-").Chan().Add(typename),
				Switch(Id("c").Op(":=").Id("c").Op(".").Parens(Type())).Block(
					Case(Op("<-").Chan().Add(typename)),
					Id("ch").Op("=").Id("c"),
					Case(Chan().Add(typename)),
					Id("ch").Op("=").Id("c"),
					Default(),
					Return(Qual(pkg, "GenericDispatch").Op(".").Id("ChanRecver").Call(Id("c"))),
				),
				Return(
					Func().Params(Id("cancel").Op("<-").Chan().Struct(), Id("block").Bool()).Params(Id("v").Interface(), Id("ok").Bool()).Block(
						If(Op("!").Id("block")).Block(
							Select().Block(
								Case(Id("v").Op(",").Id("ok").Op("=").Op("<-").Id("ch")),
								Default(),
							),
						).Else().If(Id("cancel").Op("==").Nil()).Block(
							Id("v").Op(",").Id("ok").Op("=").Op("<-").Id("ch"),
						).Else().Block(
							Select().Block(
								Case(Id("v").Op(",").Id("ok").Op("=").Op("<-").Id("ch")),
								Case(Op("<-").Id("cancel")),
							),
						),
						Return(),
					),
				),
			),
	}, nil
}
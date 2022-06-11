// Copyright 2020 CUE Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"encoding/json"
	"fmt"
	"log"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/parser"
	"cuelang.org/go/internal"
	"cuelang.org/go/internal/core/adt"
	"cuelang.org/go/internal/core/compile"
	"cuelang.org/go/internal/core/convert"
	"cuelang.org/go/internal/core/runtime"
	"github.com/kr/pretty"
)

// A Builtin is a Builtin function or constant.
//
// A function may return and a constant may be any of the following types:
//
//	error (translates to bottom)
//	nil   (translates to null)
//	bool
//	int*
//	uint*
//	float64
//	string
//	*big.Float
//	*big.Int
//
//	For any of the above, including interface{} and these types recursively:
//	[]T
//	map[string]T
type Builtin struct {
	Name   string
	Pkg    adt.Feature
	Params []Param
	Result adt.Kind
	Func   func(c *CallCtxt)
	Const  string
}

type Param struct {
	Kind  adt.Kind
	Value adt.Value // input constraint (may be nil)
}

type Package struct {
	Funcs map[string]func(c *CallCtxt)
	CUE   string
}

func newContext() *cue.Context {
	return (*cue.Context)(runtime.New())
}

var (
	funcPath  = cue.MakePath(cue.Str("func"))
	valuePath = cue.MakePath(cue.Str("value"))
)

func (p *Package) MustCompile(opCtx *adt.OpContext, importPath string) *adt.Vertex {
	log.Printf("opCtx.Runtime: %T", opCtx.Runtime)

	//ctx := (*cue.Context)(opCtx.Runtime.(*runtime.Runtime))
	//pkgLabel := ctx.StringLabel(importPath)
	//st := &adt.StructLit{}
	//if len(p.Native) > 0 {
	//	obj.AddConjunct(adt.MakeRootConjunct(nil, st))
	//}

	f, err := parser.ParseFile(importPath, p.CUE)
	if err != nil {
		panic(fmt.Errorf("could not parse %v: %v", p.CUE, err))
	}

	v, err := compile.Files(nil, opCtx, importPath, f)
	if err != nil {
		panic(fmt.Errorf("could compile parse %v: %v", p.CUE, err))
	}
	ensureStruct(opCtx, v)
	//pkgLabel := opCtx.StringLabel(importPath)
	//st := &adt.StructLit{}
	funcsLabel := opCtx.StringLabel("funcs")
	log.Printf("%d arcs; %d conjuncts; %d structs", len(v.Arcs), len(v.Conjuncts), len(v.Structs))
	log.Printf("conjunct decls %#v", v.Conjuncts[0].Field().(*adt.StructLit).Decls)
	var funcsArc *adt.Vertex
	for _, a := range v.Arcs {
		a.Finalize(opCtx)
		log.Printf("feature: (%v): %v; %d conjuncts; %d arcs; base %T", a.Label == funcsLabel, a.Label.IdentString(opCtx), len(a.Conjuncts), len(a.Arcs), a.BaseValue)
		if a.Label == funcsLabel {
			funcsArc = a
		}
	}
	if funcsArc != nil {
		ensureStruct(opCtx, funcsArc)
		log.Printf("printing funcsArc (%d members)", len(funcsArc.Arcs))
		for _, a := range funcsArc.Arcs {
			log.Printf("func: %v", a.Label.IdentString(opCtx))
		}
	}

	//obj := &adt.Vertex{}
	//if len(p.Native) > 0 {
	//	obj.AddConjunct(adt.MakeRootConjunct(nil, st))
	//}
	// TODO
	//export.VertexFeatures(ctx, v)

	//	v := ctx.BuildFile(f)
	//	if err := v.Err(); err != nil {
	//		panic(fmt.Errorf("cannot build %s: %v", importPath, err))
	//	}
	//	iter, err := v.Fields()
	//	if err != nil {
	//		panic(err)
	//	}
	//	for iter.Next() {
	//		fieldVal := iter.Value()
	//		if v := fieldVal.LookupPath(funcPath); v.Err() == nil {
	//			log.Printf("make builtin %v", iter.Selector())
	//		} else if v := fieldVal.LookupPath(valuePath); v.Err() == nil {
	//			log.Printf("make value %v", iter.Selector())
	//		} else {
	//			panic(fmt.Errorf("unrecognized builtin kind %v", iter.Selector()))
	//		}
	//	}

	//	obj.AddConjunct(c)
	//	for _, b := range p.Native {
	//		b.Pkg = pkgLabel
	//
	//		f := ctx.StringLabel(b.Name) // never starts with _
	//		// n := &node{baseValue: newBase(imp.Path)}
	//		var v adt.Expr = toBuiltin(ctx, b)
	//		if b.Const != "" {
	//			v = mustParseConstBuiltin(ctx, b.Name, b.Const)
	//		}
	//		st.Decls = append(st.Decls, &adt.Field{
	//			Label: f,
	//			Value: v,
	//		})
	//	}
	//
	//	// Parse builtin CUE
	//	if p.CUE != "" {
	//	}
	//
	//	// We could compile lazily, but this is easier for debugging.
	//	obj.Finalize(ctx)
	//	if err := obj.Err(ctx, adt.Finalized); err != nil {
	//		panic(err.Err)
	//	}

	return nil
}

func ensureStruct(opCtx *adt.OpContext, v *adt.Vertex) {
	v.Finalize(opCtx)
	v = v.Default()
	if v, ok := v.BaseValue.(*adt.Bottom); ok {
		panic(fmt.Errorf("got bottom: %v", v.Err))
	}
	if k := v.Kind(); k != adt.StructKind {
		pretty.Println(v)
		panic(fmt.Errorf("unexpected kind for export data; got %v want %v", k, adt.StructKind))
	}
}

func toBuiltin(ctx *adt.OpContext, b *Builtin) *adt.Builtin {
	params := make([]adt.Param, len(b.Params))
	for i, p := range b.Params {
		params[i].Value = p.Value
		if params[i].Value == nil {
			params[i].Value = &adt.BasicType{K: p.Kind}
		}
	}

	x := &adt.Builtin{
		Params:  params,
		Result:  b.Result,
		Package: b.Pkg,
		Name:    b.Name,
	}
	x.Func = func(ctx *adt.OpContext, args []adt.Value) (ret adt.Expr) {
		// call, _ := ctx.Source().(*ast.CallExpr)
		c := &CallCtxt{
			ctx:     ctx,
			args:    args,
			builtin: b,
		}
		defer func() {
			var errVal interface{} = c.Err
			if err := recover(); err != nil {
				errVal = err
			}
			ret = processErr(c, errVal, ret)
		}()
		b.Func(c)
		switch v := c.Ret.(type) {
		case nil:
			// Validators may return a nil in case validation passes.
			return nil
		case *adt.Bottom:
			// deal with API limitation: catch nil interface issue.
			if v != nil {
				return v
			}
			return nil
		case adt.Value:
			return v
		case bottomer:
			// deal with API limitation: catch nil interface issue.
			if b := v.Bottom(); b != nil {
				return b
			}
			return nil
		}
		if c.Err != nil {
			return nil
		}
		return convert.GoValueToValue(ctx, c.Ret, true)
	}
	return x
}

// newConstBuiltin parses and creates any CUE expression that does not have
// fields.
func mustParseConstBuiltin(ctx adt.Runtime, name, val string) adt.Expr {
	expr, err := parser.ParseExpr("<builtin:"+name+">", val)
	if err != nil {
		panic(err)
	}
	c, err := compile.Expr(nil, ctx, "_", expr)
	if err != nil {
		panic(err)
	}
	return c.Expr()

}

func (x *Builtin) name(ctx *adt.OpContext) string {
	if x.Pkg == 0 {
		return x.Name
	}
	return fmt.Sprintf("%s.%s", x.Pkg.StringValue(ctx), x.Name)
}

func (x *Builtin) isValidator() bool {
	return len(x.Params) == 1 && x.Result == adt.BoolKind
}

func processErr(call *CallCtxt, errVal interface{}, ret adt.Expr) adt.Expr {
	ctx := call.ctx
	switch err := errVal.(type) {
	case nil:
	case *adt.Bottom:
		ret = err
	case *callError:
		ret = err.b
	case *json.MarshalerError:
		if err, ok := err.Err.(bottomer); ok {
			if b := err.Bottom(); b != nil {
				ret = b
			}
		}
	case bottomer:
		ret = wrapCallErr(call, err.Bottom())

	case errors.Error:
		// Convert lists of errors to a combined Bottom error.
		if list := errors.Errors(err); len(list) != 0 && list[0] != errVal {
			var errs *adt.Bottom
			for _, err := range list {
				if b, ok := processErr(call, err, nil).(*adt.Bottom); ok {
					errs = adt.CombineErrors(nil, errs, b)
				}
			}
			if errs != nil {
				return errs
			}
		}

		ret = wrapCallErr(call, &adt.Bottom{Err: err})
	case error:
		if call.Err == internal.ErrIncomplete {
			err := ctx.NewErrf("incomplete value")
			err.Code = adt.IncompleteError
			ret = err
		} else {
			// TODO: store the underlying error explicitly
			ret = wrapCallErr(call, &adt.Bottom{Err: errors.Promote(err, "")})
		}
	default:
		// Likely a string passed to panic.
		ret = wrapCallErr(call, &adt.Bottom{
			Err: errors.Newf(call.Pos(), "%s", err),
		})
	}
	return ret
}

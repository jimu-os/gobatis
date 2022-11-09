package sgo

import (
	"fmt"
	"github.com/antonmedv/expr"
	"reflect"
	"testing"
)

type User struct {
	name string
	Map  []map[string][]*Aaa
}

type Aaa struct {
	Name string
}

func TestMap(t *testing.T) {
	v := User{
		name: "awen",
		Map: []map[string][]*Aaa{
			{
				"a": []*Aaa{
					&Aaa{Name: "1"},
					&Aaa{Name: "2"},
					&Aaa{Name: "3"},
				},
			},
		},
	}
	code := `name`
	compile, _ := expr.Compile(code, expr.Env(v))
	run, err := expr.Run(compile, v)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(run)
}

func TestExpr(t *testing.T) {
	env := map[string]any{
		"name": "1",
		"ctx": map[string]any{
			"number": 1,
			"arr":    []int{1, 2, 3, 4},
		},
	}
	code := "1==1 and 1==1"
	compile, err := expr.Compile(code, expr.Env(env))
	if err != nil {
		t.Error(err.Error())
		return
	}
	run, err := expr.Run(compile, env)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(run)
}

type Mappers struct {
	UserFind func(ctx any) map[string]any
}

func TestMaps(t *testing.T) {
	mapper := Mappers{}
	of := reflect.ValueOf(mapper)
	fmt.Println(of.NumField())
	field := of.Field(0)
	fmt.Println(field.Kind())
	fnt := field.Type()
	fmt.Println(fnt.String())
	fmt.Println(fnt.In(0).String())
	fmt.Println(fnt.In(0).Kind())
	fmt.Println(fnt.Out(0).String())
	fmt.Println(fnt.Out(0).Kind())

	mapperFunc := func(ctx any) MapperFunc {
		return func(values []reflect.Value) []reflect.Value {
			Ctx := ctx
			fmt.Println(Ctx)
			return []reflect.Value{reflect.ValueOf(map[string]any{})}
		}
	}

	makeMapper := func(v, ctx any) {
		fn := reflect.ValueOf(v).Elem()
		f := reflect.MakeFunc(fn.Type(), mapperFunc(ctx))
		fn.Set(f)
	}
	makeMapper(&mapper.UserFind, map[string]any{"1": 1})
	find := mapper.UserFind(nil)
	fmt.Println(find)
}

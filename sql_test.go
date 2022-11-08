package sgo

import (
	"github.com/antonmedv/expr"
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
	code := `ctx.arr`
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

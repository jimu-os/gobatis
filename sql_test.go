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

func BenchmarkExpr(b *testing.B) {
	env := map[string]any{
		"name": "1",
		"ctx": map[string]any{
			"number": map[string]any{
				"q": 1,
			},
			"arr": []int{1, 2, 3, 4},
		},
	}
	code := `ctx.number.q`
	compile, err := expr.Compile(code, expr.Env(env))
	if err != nil {
		b.Error(err.Error())
		return
	}

	for i := 0; i < b.N; i++ {
		_, err := expr.Run(compile, env)
		if err != nil {
			b.Error(err.Error())
			return
		}
	}
}

func BenchmarkMap(b *testing.B) {
	env := map[string]any{
		"name": "1",
		"ctx": map[string]any{
			"number": map[string]any{
				"1": 1,
			},
			"arr": []int{1, 2, 3, 4},
		},
	}
	m := toMap(env)
	for i := 0; i < b.N; i++ {
		_, err := ctxValue(m, []string{"ctx", "number", "1"})
		if err != nil {
			b.Error(err.Error())
			return
		}
	}
}

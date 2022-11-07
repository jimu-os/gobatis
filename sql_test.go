package sgo

import (
	"testing"
)

type User struct {
	Name string
	Map  []map[string][]*Aaa
}

type Aaa struct {
	Name string
}

func TestMap(t *testing.T) {
	v := User{
		Name: "awen",
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
	ctx := toMap(v)
	t.Log(ctx)
}

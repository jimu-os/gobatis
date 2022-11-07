package sgo

import "testing"

type User struct {
	Name string
	Bbb  *Aaa
	Ccc  map[string]*Aaa
}

type Aaa struct {
	Name string
}

func TestMap(t *testing.T) {
	v := User{
		Name: "awen",
		Bbb:  &Aaa{Name: "saber"},
		Ccc:  map[string]*Aaa{"test": &Aaa{Name: "testCcc"}},
	}
	ctx := toMap(v)
	t.Log(ctx)
}

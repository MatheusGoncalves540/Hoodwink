package models

import (
	"reflect"
	"testing"
)

func TestMatchReplayPlayerIDsScan(t *testing.T) {
	cases := []struct {
		name string
		src  any
		want StringArray
	}{
		{"lista", "{a,b,c}", StringArray{"a", "b", "c"}},
		{"vazia", "{}", StringArray{}},
		{"bytes", []byte("{a}"), StringArray{"a"}},
		{"aspas", `{"a","b"}`, StringArray{"a", "b"}},
		{"nula", nil, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got StringArray
			if err := got.Scan(c.src); err != nil {
				t.Fatalf("Scan(%v): %v", c.src, err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("Scan(%v) = %#v, quer %#v", c.src, got, c.want)
			}
		})
	}
}

func TestMatchReplayPlayerIDsValue(t *testing.T) {
	got, err := StringArray{"a", "b"}.Value()
	if err != nil {
		t.Fatal(err)
	}
	if got != "{a,b}" {
		t.Fatalf("Value() = %v, quer {a,b}", got)
	}
}

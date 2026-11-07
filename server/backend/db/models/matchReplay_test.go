package models

import (
	"reflect"
	"testing"
)

func TestUUIDArrayScan(t *testing.T) {
	cases := []struct {
		name string
		src  any
		want UUIDArray
	}{
		{"lista", "{a,b,c}", UUIDArray{"a", "b", "c"}},
		{"vazia", "{}", UUIDArray{}},
		{"bytes", []byte("{a}"), UUIDArray{"a"}},
		{"aspas", `{"a","b"}`, UUIDArray{"a", "b"}},
		{"nula", nil, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got UUIDArray
			if err := got.Scan(c.src); err != nil {
				t.Fatalf("Scan(%v): %v", c.src, err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("Scan(%v) = %#v, quer %#v", c.src, got, c.want)
			}
		})
	}
}

func TestUUIDArrayValue(t *testing.T) {
	got, err := UUIDArray{"a", "b"}.Value()
	if err != nil {
		t.Fatal(err)
	}
	if got != "{a,b}" {
		t.Fatalf("Value() = %v, quer {a,b}", got)
	}
}

package models

import (
	"errors"
	"reflect"
	"testing"
)

func TestStringArrayScan(t *testing.T) {
	tests := []struct {
		name string
		src  any
		want StringArray
	}{
		{"nil deixa nil", nil, nil},
		{"vazio", "{}", StringArray{}},
		{"string", "{a,b}", StringArray{"a", "b"}},
		{"aspas", `{"a","b"}`, StringArray{"a", "b"}},
		{"virgula e aspas escapadas", `{"a,b","c\"d"}`, StringArray{"a,b", `c"d`}},
		{"bytes", []byte("{a,b}"), StringArray{"a", "b"}},
		{"slice", []string{"a"}, StringArray{"a"}},
		{"espaços", "  {a}  ", StringArray{"a"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got StringArray
			if err := got.Scan(tc.src); err != nil {
				t.Fatalf("Scan(%v) falhou: %v", tc.src, err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Scan(%v) = %#v, esperado %#v", tc.src, got, tc.want)
			}
		})
	}
}

func TestStringArrayScanErros(t *testing.T) {
	for _, src := range []any{"lixo", "{a", "a}", "{a}b", 42} {
		var got StringArray
		if err := got.Scan(src); err == nil {
			t.Fatalf("Scan(%v) deveria falhar", src)
		}
	}
}

func TestStringArrayValue(t *testing.T) {
	tests := []struct {
		name string
		in   StringArray
		want any
	}{
		{"nil vira NULL", nil, nil},
		{"vazio", StringArray{}, "{}"},
		{"simples", StringArray{"a", "b"}, "{a,b}"},
		{"aspas e virgula", StringArray{"a,b", `c"d`}, `{"a,b","c\"d"}`},
		{"null literal", StringArray{"NULL"}, `{"NULL"}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.in.Value()
			if err != nil {
				t.Fatalf("Value() falhou: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Value() = %#v, esperado %#v", got, tc.want)
			}
		})
	}
}

func TestStringArrayRoundTrip(t *testing.T) {
	in := StringArray{"", "a b", `x"y`, "z,w", `{r}`, `back\slash`}
	v, err := in.Value()
	if err != nil {
		t.Fatalf("Value() falhou: %v", err)
	}
	var out StringArray
	if err := out.Scan(v); err != nil {
		t.Fatalf("Scan(%v) falhou: %v", v, err)
	}
	if !reflect.DeepEqual(out, in) {
		t.Fatalf("ida e volta = %#v, esperado %#v (literal %v)", out, in, v)
	}
}

func TestInt32ArrayScan(t *testing.T) {
	tests := []struct {
		name string
		src  any
		want Int32Array
	}{
		{"nil deixa nil", nil, nil},
		{"vazio", "{}", Int32Array{}},
		{"string", "{1,2}", Int32Array{1, 2}},
		{"aspas", `{"1"}`, Int32Array{1}},
		{"bytes", []byte("{3}"), Int32Array{3}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got Int32Array
			if err := got.Scan(tc.src); err != nil {
				t.Fatalf("Scan(%v) falhou: %v", tc.src, err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Scan(%v) = %#v, esperado %#v", tc.src, got, tc.want)
			}
		})
	}
}

func TestInt32ArrayScanErros(t *testing.T) {
	for _, src := range []any{"{1,x}", "{1", "1}", "{99999999999999999999}", true} {
		var got Int32Array
		err := got.Scan(src)
		if err == nil {
			t.Fatalf("Scan(%v) deveria falhar", src)
		}
		if errors.Is(err, nil) {
			t.Fatal("erro inesperadamente nulo")
		}
	}
}

func TestInt32ArrayValue(t *testing.T) {
	if v, err := Int32Array(nil).Value(); v != nil || err != nil {
		t.Fatalf("Value(nil) = %v, %v", v, err)
	}
	empty := Int32Array{}
	if v, _ := empty.Value(); v != "{}" {
		t.Fatalf("Value(vazio) = %v", v)
	}
	full := Int32Array{1, 2, 3}
	if v, _ := full.Value(); v != "{1,2,3}" {
		t.Fatalf("Value = %v", v)
	}
}

func TestInt32ArrayRoundTrip(t *testing.T) {
	in := Int32Array{0, 4, 10}
	v, err := in.Value()
	if err != nil {
		t.Fatalf("Value() falhou: %v", err)
	}
	var out Int32Array
	if err := out.Scan(v); err != nil {
		t.Fatalf("Scan(%v) falhou: %v", v, err)
	}
	if !reflect.DeepEqual(out, in) {
		t.Fatalf("ida e volta = %#v, esperado %#v", out, in)
	}
}

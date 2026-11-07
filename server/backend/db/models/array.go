package models

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

// StringArray lê e escreve arrays do Postgres (`uuid[]`, `text[]`).
//
// O driver devolve o array como o literal `{a,b}` e o GORM não sabe converter
// isso para `[]string` ("unsupported Scan, storing driver.Value type string into
// type *[]string"). Este tipo implementa sql.Scanner e driver.Valuer.
type StringArray []string

// Int32Array lê e escreve arrays `int[]` do Postgres (coluna turns).
type Int32Array []int32

func (a *StringArray) Scan(src any) error {
	elems, ok, err := scanArray(src)
	if err != nil {
		return err
	}
	if !ok {
		*a = nil
		return nil
	}
	out := make(StringArray, len(elems))
	copy(out, elems)
	*a = out
	return nil
}

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return arrayLiteral(a), nil
}

func (a *Int32Array) Scan(src any) error {
	elems, ok, err := scanArray(src)
	if err != nil {
		return err
	}
	if !ok {
		*a = nil
		return nil
	}
	out := make(Int32Array, len(elems))
	for i, e := range elems {
		if e == "" {
			continue
		}
		n, err := strconv.ParseInt(e, 10, 32)
		if err != nil {
			return fmt.Errorf("int[] inválido: elemento %q", e)
		}
		out[i] = int32(n)
	}
	*a = out
	return nil
}

func (a Int32Array) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	parts := make([]string, len(a))
	for i, n := range a {
		parts[i] = strconv.FormatInt(int64(n), 10)
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}

// scanArray normaliza o que o driver entrega em uma lista de elementos. ok é
// false para NULL (o slice de destino fica nil).
func scanArray(src any) (elems []string, ok bool, err error) {
	switch v := src.(type) {
	case nil:
		return nil, false, nil
	case string:
		elems, err := parseArray(v)
		return elems, true, err
	case []byte:
		elems, err := parseArray(string(v))
		return elems, true, err
	case []string:
		return v, true, nil
	}
	return nil, false, fmt.Errorf("tipo inesperado para array do Postgres: %T", src)
}

// parseArray lê o literal `{a,"b,c"}` do Postgres: elementos entre aspas aceitam
// vírgula e escape com barra invertida. NULL (sem aspas) vira elemento vazio.
func parseArray(src string) ([]string, error) {
	s := strings.TrimSpace(src)
	if len(s) < 2 || s[0] != '{' || s[len(s)-1] != '}' {
		return nil, fmt.Errorf("array do Postgres inválido: %q", src)
	}
	body := s[1 : len(s)-1]
	if body == "" {
		return []string{}, nil
	}
	elems := make([]string, 0, 4)
	i := 0
	for {
		var b strings.Builder
		quoted := false
		if i < len(body) && body[i] == '"' {
			quoted = true
			i++
			for i < len(body) && body[i] != '"' {
				if body[i] == '\\' && i+1 < len(body) {
					b.WriteByte(body[i+1])
					i += 2
					continue
				}
				b.WriteByte(body[i])
				i++
			}
			i++ // fecha a aspa
		} else {
			for i < len(body) && body[i] != ',' {
				if body[i] == '\\' && i+1 < len(body) {
					b.WriteByte(body[i+1])
					i += 2
					continue
				}
				b.WriteByte(body[i])
				i++
			}
		}
		el := b.String()
		if !quoted && el == "NULL" {
			el = ""
		}
		elems = append(elems, el)

		if i >= len(body) {
			break
		}
		if body[i] != ',' {
			return nil, fmt.Errorf("array do Postgres inválido: %q", src)
		}
		i++
	}
	return elems, nil
}

// arrayLiteral monta `{a,b}`; elemento vazio, nulo, com vírgula, aspas, barra
// ou espaço precisa de aspas.
func arrayLiteral[S ~[]string](a S) string {
	if len(a) == 0 {
		return "{}"
	}
	parts := make([]string, len(a))
	for i, s := range a {
		parts[i] = quoteArrayElement(s)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func quoteArrayElement(s string) string {
	if s != "" && !strings.EqualFold(s, "NULL") && !strings.ContainsAny(s, ",\"\\{} \t\n\r") {
		return s
	}
	escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s)
	return `"` + escaped + `"`
}

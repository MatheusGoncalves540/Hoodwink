// Code generated pelo enginebump (make engine-bump). NÃO edite à mão.
//
// Package versions lista as versões congeladas do motor sem importá-las
// (por isso fica num pacote separado de engines).
package versions

// Version identifica uma cópia congelada do motor.
type Version struct {
	Number int
	Hash   string // sha256 (hex) dos arquivos de replay/engines/vN
}

// All lista todas as versões congeladas, em ordem crescente.
var All = []Version{
	{Number: 1, Hash: "1059f86c00fe1f0bdeeef8e2c966d6c0ea970ecebecdda8543b3df20038285b6"},
	{Number: 2, Hash: "87ccce0ebe030d6cb0da7231577c658be4bb990faefcc7a1c89d476b6783e8ee"},
	{Number: 3, Hash: "f2b889c29902770896bf87cf40c85f9eaf2b1910b19c8878b152311e9335f5cc"},
	{Number: 4, Hash: "36b4d88f382d1fb44a0159845159fb0b66e78bed908a52f9fe8d736d283907a1"},
	{Number: 5, Hash: "e0bd2a4107a62aaca0e9f0dce9ec6c387a88519dc11239d258cfd662852c80ff"},
}

// Latest devolve o número da versão mais recente.
func Latest() int {
	if len(All) == 0 {
		return 0
	}
	return All[len(All)-1].Number
}

// ByNumber localiza uma versão pelo número.
func ByNumber(n int) (Version, bool) {
	for _, v := range All {
		if v.Number == n {
			return v, true
		}
	}
	return Version{}, false
}

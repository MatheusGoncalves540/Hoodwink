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

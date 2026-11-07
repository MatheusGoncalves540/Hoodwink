package handlers

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/utils"
	"github.com/MatheusGoncalves540/Hoodwink/version"
)

// versionDTO é a resposta pública de GET /version.
type versionDTO struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// VersionHandler devolve a versão do jogo embutida no binário. É pública: serve
// para conferir o que está no ar depois do deploy.
func (h *Handler) VersionHandler(w http.ResponseWriter, r *http.Request) {
	utils.SendSuccess(w, versionDTO{Version: version.Version, Commit: version.Commit}, "")
}

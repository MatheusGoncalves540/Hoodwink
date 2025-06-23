package external

import (
	"net/http"

	"github.com/markbates/goth/gothic"
)

func SetupExternalAuths() {
	SetupGoogleOAuth()
}

func BeginAuthHandler(w http.ResponseWriter, r *http.Request) {
	gothic.BeginAuthHandler(w, r)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	userInfo, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Error(w, "Erro na autenticação: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Salva o email e provider na sessão temporária
	session, _ := gothic.Store.Get(r, "temp-auth-session")
	session.Values["email"] = userInfo.Email
	session.Values["provider"] = userInfo.Provider
	session.Save(r, w)

	// Retorne um status ou redirecione para a página de formulário (ex: /additional-user-data)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"need_additional_data"}`))
}

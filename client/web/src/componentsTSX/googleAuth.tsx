import { useEffect, useRef, useState } from "react";
import handleCredentialResponse from "../Auth/handleCredentialResponse";
import { useAuthStore } from "../store/authStore";
import { ENV } from "../config/envs";
import { copy } from "../game/copy";
import isTokenExpired from "../Auth/isTokenExpired";

const gisReady = () => !!window.google?.accounts?.id;

function GoogleAuth() {
  const jwtToken = useAuthStore((s) => s.jwtToken);
  const closeLogin = useAuthStore((s) => s.closeLogin);
  const setAuth = useAuthStore((s) => s.setAuth);
  const buttonRef = useRef<HTMLDivElement>(null);
  // o script do Google carrega com async: se o modal abrir antes, espera o load em vez de ficar sem botão
  const [ready, setReady] = useState(gisReady);

  useEffect(() => {
    if (ready) return;
    const script = document.querySelector<HTMLScriptElement>('script[src*="accounts.google.com/gsi/client"]');
    const onLoad = () => setReady(gisReady());
    script?.addEventListener("load", onLoad);
    return () => script?.removeEventListener("load", onLoad);
  }, [ready]);

  useEffect(() => {
    if (!ready || !buttonRef.current) return;

    window.google.accounts.id.initialize({
      client_id: ENV.VITE_GOOGLE_CLIENT_ID,
      callback: (response) => handleCredentialResponse(response, setAuth),
    });

    window.google.accounts.id.renderButton(buttonRef.current, {
      type: "standard",
      theme: document.documentElement.dataset.theme === "dark" ? "filled_black" : "outline",
      size: "large",
      shape: "pill",
      text: "continue_with",
      locale: "pt-BR",
    });
  }, [ready, setAuth]);

  // Verifica a validade do token JWT sempre que ele for atualizado
  useEffect(() => {
    if (!jwtToken) return;

    if (isTokenExpired(jwtToken)) {
      alert("Token JWT expirado. Faça login novamente.");
      closeLogin();
    }
  }, [jwtToken, closeLogin]);

  return (
    <div className="hw-login-google">
      {ready ? <div ref={buttonRef} /> : <p className="hw-login-espera">{copy.loginModal.loading}</p>}
    </div>
  );
}

export default GoogleAuth;

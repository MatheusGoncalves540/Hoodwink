import { useCallback, useEffect, useRef, useState } from "react";
import handleCredentialResponse from "../Auth/handleCredentialResponse";
import { useAuthStore } from "../store/authStore";
import { ENV } from "../config/envs";
import { copy } from "../game/copy";
import isTokenExpired from "../Auth/isTokenExpired";
import { Button } from "../ui/Button/Button";

const gisReady = () => !!window.google?.accounts?.id;

function GoogleAuth() {
  const jwtToken = useAuthStore((s) => s.jwtToken);
  const setAuth = useAuthStore((s) => s.setAuth);
  const buttonRef = useRef<HTMLDivElement>(null);
  const nameRef = useRef<HTMLInputElement>(null);
  // o script do Google carrega com async: se o modal abrir antes, espera o load em vez de ficar sem botão
  const [ready, setReady] = useState(gisReady);
  const [error, setError] = useState("");
  // resposta need_additional_data: o modal pede o nome em vez de abrir um prompt nativo
  const [pending, setPending] = useState("");
  const [name, setName] = useState("");
  const [busy, setBusy] = useState(false);

  const handle = useCallback(
    async (credential: string, username?: string) => {
      setError("");
      setBusy(true);
      const outcome = await handleCredentialResponse({ credential }, setAuth, username);
      setBusy(false);
      if (outcome.status === "need_name") {
        setPending(credential);
        setName("");
        return;
      }
      if (outcome.status === "error") {
        setError(outcome.message);
        return;
      }
      setPending("");
    },
    [setAuth],
  );

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
      callback: (response) => void handle(response.credential),
    });

    window.google.accounts.id.renderButton(buttonRef.current, {
      type: "standard",
      theme: document.documentElement.dataset.theme === "dark" ? "filled_black" : "outline",
      size: "large",
      shape: "pill",
      text: "continue_with",
      locale: "pt-BR",
    });
  }, [ready, setAuth, handle]);

  // Verifica a validade do token JWT sempre que ele for atualizado
  useEffect(() => {
    if (jwtToken && isTokenExpired(jwtToken)) setError(copy.loginModal.expired);
  }, [jwtToken]);

  useEffect(() => {
    if (pending) nameRef.current?.focus();
  }, [pending]);

  return (
    <div className="hw-login-google">
      {ready ? <div ref={buttonRef} /> : <p className="hw-login-espera">{copy.loginModal.loading}</p>}
      {pending && (
        <form
          className="hw-login-local"
          onSubmit={(e) => {
            e.preventDefault();
            const clean = name.trim();
            if (clean && !busy) void handle(pending, clean);
          }}
        >
          <p className="hw-login-texto">{copy.loginModal.needName}</p>
          <input
            ref={nameRef}
            className="hw-login-local-nome"
            aria-label={copy.loginModal.nameLabel}
            placeholder={copy.loginModal.nameLabel}
            value={name}
            onChange={(e) => {
              setName(e.target.value);
              setError("");
            }}
            maxLength={20}
            autoComplete="off"
            spellCheck={false}
          />
          <Button type="submit" variant="outline" disabled={!name.trim() || busy}>
            {copy.loginModal.nameSubmit}
          </Button>
        </form>
      )}
      {error && (
        <p role="alert" className="hw-login-erro">
          {error}
        </p>
      )}
    </div>
  );
}

export default GoogleAuth;

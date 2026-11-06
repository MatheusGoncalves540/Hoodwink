import { useState } from "react";
import { Dialog } from "radix-ui";
import { copy } from "../game/copy";
import { useAuthStore } from "../store/authStore";
import { Button } from "../ui/Button/Button";
import { Label } from "../ui/Label/Label";
import { Moldura } from "../ui/Moldura/Moldura";
import "./styles/loginModal.css";
import GoogleAuth from "./googleAuth";
import loginLocal from "../Auth/loginLocal";

/** Entrar sem Google, para quem está na mesma rede (Radmin). O backend só aceita em ambiente local. */
function LocalLogin() {
  const setAuth = useAuthStore((s) => s.setAuth);
  const [name, setName] = useState("");
  const clean = name.trim();

  return (
    <form
      className="hw-login-local"
      onSubmit={(e) => {
        e.preventDefault();
        if (clean) loginLocal(clean, setAuth);
      }}
    >
      <input
        className="hw-login-local-nome"
        aria-label={copy.loginModal.localName}
        placeholder={copy.loginModal.localName}
        value={name}
        onChange={(e) => setName(e.target.value)}
        maxLength={20}
        autoComplete="off"
        spellCheck={false}
      />
      <Button type="submit" variant="outline" disabled={!clean}>
        {copy.loginModal.localEnter}
      </Button>
    </form>
  );
}

/** Bilheteria: o login aparece como um cartaz pequeno, com o mesmo letreiro de lâmpadas da entrada. */
function LoginModal() {
  const isLoginOpen = useAuthStore((s) => s.isLoginOpen);
  const closeLogin = useAuthStore((s) => s.closeLogin);

  return (
    <Dialog.Root open={isLoginOpen} onOpenChange={(open) => !open && closeLogin()}>
      <Dialog.Portal>
        <Dialog.Overlay className="m-scrim hw-login-veu" />
        <Dialog.Content className="hw-login" aria-describedby="hw-login-desc">
          <Moldura lugar="lobby" className="m-hang hw-login-cartaz">
            <Label tone="ouro">{copy.loginModal.kicker}</Label>
            <Dialog.Title className="hw-login-titulo">{copy.loginModal.title}</Dialog.Title>
            <Dialog.Description id="hw-login-desc" className="hw-login-texto">
              {copy.loginModal.text}
            </Dialog.Description>
            <GoogleAuth />
            <p className="hw-login-ou">{copy.loginModal.localOr}</p>
            <LocalLogin />
            <Dialog.Close asChild>
              <Button variant="link" className="hw-login-voltar">
                {copy.loginModal.back}
              </Button>
            </Dialog.Close>
          </Moldura>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

export default LoginModal;

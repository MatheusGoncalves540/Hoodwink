/**
 * Entrada (tela 1): cartaz do Proso, "Criar sala" e entrar pelo código.
 * Lobby.dc.html / LobbyDesk.dc.html; o perfil é o do login existente (Auth/).
 */
import { useMemo, useState } from "react";
import { useLayout } from "../../app/useLayout";
import { useMotionPref } from "../../app/useMotionPref";
import LoginModal from "../../componentsTSX/loginModal";
import { copy } from "../../game/copy";
import { useAuthStore } from "../../store/authStore";
import { MotionToggles } from "../sheets/RoomMenu";
import { createInput, roomCodeFromSearch, useSessionStore, usernameFromToken } from "../../store/sessionStore";
import { Button } from "../../ui/Button/Button";
import { Label } from "../../ui/Label/Label";
import { Moldura } from "../../ui/Moldura/Moldura";
import { TextField } from "../../ui/TextField/TextField";
import { Ticket } from "../../ui/Ticket/Ticket";
import { cx } from "../../ui/cx";
import s from "./Lobby.module.css";

/** O design desenha 4 letras (K7Q2); o servidor gera ids em hexa maiores. Entrar libera em 4. */
const MIN_CODE = 4;
const MAX_CODE = 12;

/** Canto superior direito do cartaz: inicial do ator logado ou "Entrar com Google". */
export function Profile({ signedIn, name, onLogin }: { signedIn: boolean; name: string; onLogin(): void }) {
  if (!signedIn) {
    return (
      <Button variant="outline" className={s.login} onClick={onLogin}>
        {copy.login}
      </Button>
    );
  }
  const who = name || copy.you;
  return (
    <span role="img" aria-label={copy.profile(who)} className={s.avatar}>
      {who.charAt(0).toUpperCase()}
    </span>
  );
}

export function Lobby() {
  const layout = useLayout();
  const [code, setCode] = useState(() => (typeof window === "undefined" ? "" : roomCodeFromSearch(window.location.search) ?? ""));
  const jwtToken = useAuthStore((st) => st.jwtToken);
  const openLogin = useAuthStore((st) => st.openLogin);
  const busy = useSessionStore((st) => st.busy);
  const error = useSessionStore((st) => st.error);
  const createRoom = useSessionStore((st) => st.create);
  const joinRoom = useSessionStore((st) => st.join);
  const username = useMemo(() => usernameFromToken(jwtToken), [jwtToken]);
  const [mask, truth] = copy.epigraph.split(" e ele ");
  const motion = useMotionPref();

  const trimmed = code.trim();
  const valid = trimmed.length >= MIN_CODE;
  const message =
    error === "Sala não encontrada"
      ? copy.roomNotFound(trimmed.toUpperCase())
      : error === "Sala está cheia"
        ? copy.roomFull
        : error;

  /** Convite: o código vai para a URL (04-arquitetura.md → sem roteador). */
  const goToRoom = (roomCode: string) => {
    const url = new URL(window.location.href);
    url.searchParams.delete("cena");
    url.searchParams.delete("fixture");
    url.searchParams.set("sala", roomCode);
    window.history.replaceState(null, "", url);
  };

  const onCreate = async () => {
    if (!jwtToken) return openLogin();
    const roomCode = await createRoom(jwtToken, createInput(username));
    if (roomCode) goToRoom(roomCode);
  };

  const onEnter = async () => {
    if (!jwtToken) return openLogin();
    if (!valid) return;
    const roomCode = await joinRoom(jwtToken, trimmed.toLowerCase());
    if (roomCode) goToRoom(roomCode);
  };

  return (
    <main className={cx(s.screen, layout === "desktop" && s.desk)}>
      <Moldura lugar="lobby" as="section" aria-label={copy.cartazRegion} className={cx("m-hang", s.cartaz)}>
        <div className={s.profileRow}>
          <Profile signedIn={!!jwtToken} name={username} onLogin={openLogin} />
        </div>

        {/* Proso em placeholder até a arte nova: moldura tracejada em contorno + "?" em tira */}
        <div role="img" aria-label={copy.prosoAlt} className={cx("m-contorno", s.proso)}>
          {[0, 1, 2].map((i) => (
            <svg key={i} aria-hidden="true" className={cx("m-f", s.prosoFrame)} preserveAspectRatio="none" viewBox="0 0 100 100">
              <rect x="1" y="1" width="98" height="98" rx={4 + i} vectorEffect="non-scaling-stroke" />
            </svg>
          ))}
          <span aria-hidden="true" className={cx("m-tira", s.prosoAsk)}>
            <span>
              <i>?</i>
              <i>?</i>
              <i>?</i>
            </span>
          </span>
          <span aria-hidden="true" className={s.prosoNote}>{copy.prosoPlaceholder}</span>
        </div>

        <Label tone="ouro" className={s.nowShowing}>
          {copy.nowShowing}
        </Label>
        <h1 aria-label={copy.brand} className={s.logo}>
          <span aria-hidden="true">{copy.brand}</span>
          <svg aria-hidden="true" viewBox="0 0 100 160">
            <polygon points="52,104 18,152 86,152" fill="color-mix(in srgb, var(--hw-ouro) 22%, transparent)" />
            <path
              d="M20 50 C14 22 34 8 52 8 C72 8 82 24 80 42 C78 62 52 66 52 90 L52 104"
              fill="none"
              stroke="var(--hw-tinta)"
              strokeWidth="14"
              strokeLinecap="round"
            />
            <ellipse cx="52" cy="150" rx="34" ry="8" fill="var(--hw-ouro)" />
            <ellipse cx="52" cy="150" rx="18" ry="4" fill="var(--hw-lampada)" />
          </svg>
        </h1>
        <p className={s.epigraph}>
          <i>{truth ? <>“{mask}<br />e ele {truth}”</> : <>“{copy.epigraph}”</>}</i>&nbsp;
        </p>
      </Moldura>

      <section aria-label={copy.playRegion} className={cx("m-rise", s.actions)}>
        <Ticket size="l" disabled={busy} onClick={onCreate}>
          {copy.createRoom}
        </Ticket>

        <TextField
          label={copy.roomCodePlaceholder}
          placeholder={copy.roomCodePlaceholder}
          value={code}
          onChange={(v) => setCode(v.toLowerCase())}
          action={copy.enter}
          actionDisabled={!valid || busy}
          onSubmit={onEnter}
          upper
          maxLength={MAX_CODE}
          autoCapitalize="none"
          spellCheck={false}
        />

        {message && (
          <p role="alert" className={s.error}>
            {message}
          </p>
        )}

        <a className={s.how} href="#como-se-joga">
          ☞ {copy.howToPlay}
        </a>

        <details className={s.prefs}>
          <summary>{copy.prefsTitle}</summary>
          <MotionToggles motion={motion} onMotion={motion.setAll} onIdle={motion.setIdle} />
        </details>
      </section>

      <LoginModal />
    </main>
  );
}

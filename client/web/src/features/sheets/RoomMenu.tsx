import { useState } from "react";
import { setPref, usePref } from "../../app/prefs";
import { useMotionPref } from "../../app/useMotionPref";
import { copy } from "../../game/copy";
import type { GameView } from "../../game/selectors";
import { useGameStore } from "../../store/gameStore";
import { useSessionStore } from "../../store/sessionStore";
import { Button } from "../../ui/Button/Button";
import { Sheet } from "../../ui/Sheet/Sheet";
import { Toggle } from "../../ui/Toggle/Toggle";
import { FlagButton } from "../bugReport/FlagButton";
import s from "./RoomMenu.module.css";

const SOUND_KEY = "hw-som";
const VIBRATION_KEY = "hw-vibracao";

type ContentProps = {
  code: string;
  turn: number;
  sound: boolean;
  vibration: boolean;
  motion: { all: boolean; idle: boolean };
  onSound(value: boolean): void;
  onVibration(value: boolean): void;
  onMotion(value: boolean): void;
  onIdle(value: boolean): void;
  onLeave(): void;
  matchId: string;
  mode: "mobile" | "desktop";
};

export function RoomMenuContent({ code, turn, sound, vibration, motion, onSound, onVibration, onMotion, onIdle, onLeave, matchId, mode }: ContentProps) {
  const [confirm, setConfirm] = useState(false);
  return (
    <>
      <div className={s.head}>
        <h2 className={s.title}>
          {copy.roomNumber} <span className={s.code}>{code}</span>
        </h2>
        <span className={s.round}>{copy.roundN(turn)}</span>
      </div>

      <a className={s.how} href="#como-se-joga">
        <span>☞ {copy.howToPlay}</span>
        <span aria-hidden="true" className={s.leader} />
        <span className={s.rolesLink}>{copy.rolesLink}</span>
      </a>

      <div className={s.prefs}>
        <Toggle label={copy.sound} value={sound} onChange={onSound} />
        <Toggle label={copy.vibration} hint={copy.vibrationHint} value={vibration} onChange={onVibration} />
        <MotionToggles motion={motion} onMotion={onMotion} onIdle={onIdle} />
      </div>

      <a className={s.how} href="#camarim">
        <span>☞ {copy.camarim.title}</span>
      </a>

      <p className={s.warning}>{copy.menuWarning(code)}</p>

      <FlagButton mode={mode} matchId={matchId} turn={turn} />

      {confirm ? (
        <div className={s.confirm}>
          <p className={s.confirmText}>{copy.leaveConfirm}</p>
          <div className={s.confirmRow}>
            <Button variant="danger" onClick={onLeave}>
              {copy.leaveYes}
            </Button>
            <Button variant="outline" onClick={() => setConfirm(false)}>
              {copy.leaveNo}
            </Button>
          </div>
        </div>
      ) : (
        <Button variant="danger" className={s.leave} onClick={() => setConfirm(true)}>
          {copy.leaveScene}
        </Button>
      )}
    </>
  );
}

/** As duas chaves de movimento (fase 3.1): sem Animações, a espera aparece desligada e travada. */
export function MotionToggles({ motion, onMotion, onIdle }: { motion: { all: boolean; idle: boolean }; onMotion(v: boolean): void; onIdle(v: boolean): void }) {
  return (
    <>
      <Toggle label={copy.motionAll} hint={copy.motionAllHint} value={motion.all} onChange={onMotion} />
      <Toggle
        label={copy.motionIdle}
        hint={motion.all ? copy.motionIdleHint : copy.motionIdleOff}
        value={motion.idle}
        disabled={!motion.all}
        onChange={onIdle}
      />
    </>
  );
}

type Props = { view: GameView; mode: "mobile" | "desktop"; onClose(): void };

/** Menu da sala (telas 10): preferências e saída, na mesma gaveta. */
export function RoomMenu({ view, mode, onClose }: Props) {
  const code = (view.room?.id ?? "").toUpperCase();
  const turn = view.room?.turn ?? 0;
  const sound = usePref(SOUND_KEY) !== "nao";
  const vibration = usePref(VIBRATION_KEY) !== "nao";
  const motion = useMotionPref();

  const leave = () => {
    useGameStore.getState().reset();
    useSessionStore.getState().reset();
    if (typeof window !== "undefined") {
      const url = new URL(window.location.href);
      url.searchParams.delete("sala");
      url.searchParams.delete("cena");
      url.searchParams.delete("fixture");
      window.history.replaceState(null, "", url);
    }
    onClose();
  };

  return (
    <Sheet open onClose={onClose} label={copy.menuTitle} mode={mode}>
      <RoomMenuContent
        code={code}
        turn={turn}
        sound={sound}
        vibration={vibration}
        motion={motion}
        onSound={(v) => setPref(SOUND_KEY, v ? "sim" : "nao")}
        onVibration={(v) => setPref(VIBRATION_KEY, v ? "sim" : "nao")}
        onMotion={motion.setAll}
        onIdle={motion.setIdle}
        onLeave={leave}
        matchId={view.room?.matchId ?? ""}
        mode={mode}
      />
    </Sheet>
  );
}

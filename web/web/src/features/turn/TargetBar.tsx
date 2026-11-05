import { copy } from "../../game/copy";
import type { RoleId } from "../../game/protocol";
import { roleName } from "../../game/roles";
import { Button } from "../../ui/Button/Button";
import s from "./TargetBar.module.css";

/** Fluxo de alvo inferido (05-protocolo.md): escolhido o papel, a mesa espera o camarote. */
export function TargetBar({ role, onBack }: { role: RoleId; onBack(): void }) {
  return (
    <section aria-label={copy.pickTarget} className={s.bar}>
      <p className={s.title}>{copy.pickTarget}</p>
      <p className={s.hint}>{copy.pickTargetHint(roleName(role))}</p>
      <Button variant="link" className={s.back} onClick={onBack}>
        {copy.backToActions}
      </Button>
    </section>
  );
}

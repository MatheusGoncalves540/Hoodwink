import type { Ref } from "react";
import { copy } from "../../game/copy";
import type { ChatItem, GameView } from "../../game/selectors";
import { selectChat } from "../../game/selectors";
import { chat as chatMessage } from "../../net/actions";
import { send } from "../../net/GameSocket";
import { useUiStore } from "../../store/uiStore";
import { Label } from "../../ui/Label/Label";
import { TextField } from "../../ui/TextField/TextField";
import { SceneLog } from "./SceneLog";
import s from "./DesktopRail.module.css";

type Line = Extract<ChatItem, { kind: "line" }>;

/** Trilho do desktop sem gaveta: diário em cima, conversa encostada embaixo e o campo no pé. */
export function DesktopRail({ view, chatInputRef }: { view: GameView; chatInputRef?: Ref<HTMLInputElement> }) {
  const draft = useUiStore((st) => st.chatDraft);
  const setDraft = useUiStore((st) => st.setDraft);
  const lines = selectChat(view).filter((it): it is Line => it.kind === "line").slice(-3);

  const submit = () => {
    const text = draft.trim();
    if (!text) return;
    send(chatMessage(text));
    setDraft("");
  };

  return (
    <div className={s.rail}>
      <SceneLog view={view} />
      <div className={s.chat}>
        <Label>{copy.chatShort}</Label>
        <div className={s.lines}>
          {lines.map((it) => (
            <p key={it.key} className={s.line}>
              <span className={s.name}>{it.name}.</span> {it.text}
            </p>
          ))}
        </div>
        <TextField
          label={copy.sayToTable}
          placeholder={copy.sayToTable}
          value={draft}
          onChange={setDraft}
          action={copy.send}
          tone="dark"
          inputRef={chatInputRef}
          onSubmit={submit}
          maxLength={140}
        />
      </div>
    </div>
  );
}

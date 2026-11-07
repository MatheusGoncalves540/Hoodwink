import { useEffect, useRef, useState, type Ref } from "react";
import { copy } from "../../game/copy";
import type { ChatItem, GameView } from "../../game/selectors";
import { selectChat } from "../../game/selectors";
import { chat as chatMessage } from "../../net/actions";
import { send } from "../../net/GameSocket";
import { useUiStore } from "../../store/uiStore";
import { Sheet } from "../../ui/Sheet/Sheet";
import { TextField } from "../../ui/TextField/TextField";
import { cx } from "../../ui/cx";
import s from "./ChatSheet.module.css";

export function ChatSheetBody({ items, endRef }: { items: ChatItem[]; endRef?: Ref<HTMLDivElement> }) {
  return (
    <div className={s.chat}>
      <div className={s.head}>
        <h2 className={s.title}>{copy.chatTitle}</h2>
        <span className={s.keeps}>{copy.chatKeeps}</span>
      </div>
      <div role="log" aria-label={copy.chatList} className={s.log}>
        {items.map((it) =>
          it.kind === "event" ? (
            <p key={it.key} className={cx(s.item, s.event)}>
              {it.text}
            </p>
          ) : (
            <p key={it.key} className={cx(s.item, it.mine && s.mine)}>
              <span className={s.time}>{it.time}</span>
              <span className={s.name}>{it.name}.</span> <span className={s.text}>{it.text}</span>
            </p>
          ),
        )}
        <div ref={endRef} />
      </div>
    </div>
  );
}

type Props = { view: GameView; mode: "mobile" | "desktop"; onClose(): void };

/** Conversa da mesa (telas 11 e 15): falas, acontecimentos e o campo no pé. */
export function ChatSheet({ view, mode, onClose }: Props) {
  const items = selectChat(view);
  const [draft, setDraft] = useState("");
  const markSeen = useUiStore((st) => st.markChatSeen);
  const endRef = useRef<HTMLDivElement>(null);
  const count = view.room?.chat.length ?? 0;

  useEffect(() => {
    markSeen(count);
  }, [markSeen, count]);

  useEffect(() => {
    endRef.current?.scrollIntoView({ block: "end" });
  }, [items.length]);

  const submit = () => {
    const text = draft.trim();
    if (!text) return;
    send(chatMessage(text));
    setDraft("");
  };

  return (
    <Sheet
      open
      onClose={onClose}
      label={copy.chatTitle}
      mode={mode}
      fill
      footer={
        <TextField
          label={copy.yourLine}
          placeholder={copy.yourLinePlaceholder}
          value={draft}
          onChange={setDraft}
          action={copy.speak}
          actionDisabled={!draft.trim()}
          onSubmit={submit}
          maxLength={140}
        />
      }
    >
      <ChatSheetBody items={items} endRef={endRef} />
    </Sheet>
  );
}

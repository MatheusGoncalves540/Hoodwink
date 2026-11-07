import { useMemo, type CSSProperties, type ReactNode } from "react";
import type { Layout } from "../../app/useLayout";
import { copy } from "../../game/copy";
import { basicPrice, canSwap, priceOf, type Price } from "../../game/pricing";
import { BASIC_GROUP, ROLE_GROUPS, type ActionId, type RoleGroup, type RoleId } from "../../game/roles";
import { CLASSIC_RULES } from "../../game/rules.classic";
import { kamikazeAlly, selectFootlights, selectYou, type GameView } from "../../game/selectors";
import { Coin } from "../../ui/Coin/Coin";
import { Timer } from "../../ui/Timer/Timer";
import { cx } from "../../ui/cx";
import { ActionButton } from "./ActionButton";
import s from "./ActionBoard.module.css";

/** Ordem na cascata de entrada da gaveta (ActionBoard.module.css). */
const cascade = (i: number) => ({ "--i": i }) as CSSProperties;

type Props = {
  view: GameView;
  layout: Layout;
  onPick(action: ActionId): void;
};

/**
 * Cabeçalho da gaveta de ações (vai no handle da CenterCurtain): fica à vista com a gaveta fechada,
 * para olhar o palco sem perder suas moedas e o tempo da jogada.
 */
export function ActionPeek({ view, open }: { view: GameView; open: boolean }) {
  const you = useMemo(() => selectYou(view), [view]);
  const footlights = useMemo(() => selectFootlights(view), [view]);
  const ev = view.room?.gameEvent;
  const total = footlights.suddenDeath ? CLASSIC_RULES.waitingAction.suddenDeath : CLASSIC_RULES.waitingAction.default;
  return (
    <>
      <span className={s.peekRow}>
        <span className={s.peekTitle}>
          <span className={cx("hw-rotulo", s.kicker)}>{copy.yourTurnKicker}</span>
          <span className={s.h1}>{copy.declareRole}</span>
        </span>
        <span aria-hidden="true" className={cx("hw-rotulo", s.peekHint)}>
          <span className={s.hintText}>{open ? copy.seeStage : copy.openShort}</span>
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={cx(s.chevron, open && s.chevronDown)}>
            <path d="M3 10l5-5 5 5" />
          </svg>
        </span>
        <Coin size={46} value={you.coins} label={copy.youHave(you.coins)} />
      </span>
      {ev && <Timer expiresAt={ev.expiresAt} total={total} label={copy.moveTime} decision />}
    </>
  );
}

/**
 * Seu turno (tela 6), corpo da gaveta: "Na sua mão" no topo (papéis que você tem, sem blefe, com a carta
 * ao lado) e depois as ações do turno e os personagens por grupo, com o preço de pricing.ts.
 */
export function ActionBoard({ view, layout, onPick }: Props) {
  const room = view.room;
  const you = useMemo(() => selectYou(view), [view]);
  const held = useMemo(() => new Set(you.cards.filter((c) => c.role && !c.dead).map((c) => c.role)), [you]);
  // preço de cada ação só muda com a sala/taxa ou com suas moedas
  const prices = useMemo(() => {
    const basic = new Map<ActionId, Price>();
    const roles = new Map<string, Price>();
    if (room) {
      for (const action of BASIC_GROUP.actions) basic.set(action, basicPrice(action, room, you.coins));
      for (const group of ROLE_GROUPS) for (const role of group.roles) roles.set(role, priceOf(role, room));
    }
    return { basic, roles };
  }, [room, you.coins]);
  const short = (p: Price) => (p.kind !== "gain" && p.coins > you.coins ? p.coins - you.coins : null);

  const roleButton = (role: RoleId, group: RoleGroup, inHand: boolean): ReactNode => {
    const price: Price = prices.roles.get(role) ?? { kind: "cost", coins: 0 };
    const offTurn = !!group.offTurn;
    const missing = offTurn ? null : short(price);
    // no seu turno o Kamikaze só serve se você matar um Kamikaze vivo seu (kamikazeValidation.go)
    const blocked = role === "KAMIKAZE" && kamikazeAlly(view) == null ? copy.needKamikaze : undefined;
    const card = inHand ? you.cards.find((c) => c.role === role && !c.dead) : undefined;
    return (
      <ActionButton
        key={role}
        role={role}
        price={price}
        held={inHand}
        heldProtected={card?.protected}
        offTurn={offTurn}
        missing={blocked ? null : missing}
        blocked={blocked}
        onClick={onPick}
      />
    );
  };
  const heldRoles = ROLE_GROUPS.flatMap((group) => group.roles.filter((role) => held.has(role)).map((role) => ({ role, group })));

  return (
    <div className={cx(s.screen, layout === "desktop" && s.desk)}>
      <section aria-label={copy.chooseAction} className={s.board}>
        {heldRoles.length > 0 && (
          <section aria-label={copy.panelCards} className={s.held} data-slot="hand">
            <div className={s.heldHead}>
              <h2 className={cx("hw-rotulo", s.heldTitle)}>{copy.heldTitle}</h2>
              <span className={s.heldHint}>{copy.heldHint}</span>
            </div>
            <div className={s.heldList}>{heldRoles.map(({ role, group }) => roleButton(role, group, true))}</div>
          </section>
        )}

        <div className={s.list}>
          <div className={s.group} style={cascade(1)}>
            <h2 className={s.groupTitle}>{BASIC_GROUP.title}</h2>
            {BASIC_GROUP.actions.map((action) => {
              const price = prices.basic.get(action) ?? ({ kind: "cost", coins: 0 } as const);
              const blocked = action === "SWAP_CARDS" && !canSwap(you.coins) ? copy.needSwapCoins(CLASSIC_RULES.basic.swapMinCoins) : undefined;
              return (
                <ActionButton
                  key={action}
                  role={action}
                  price={price}
                  held={false}
                  offTurn={false}
                  missing={blocked ? null : short(price)}
                  blocked={blocked}
                  onClick={onPick}
                />
              );
            })}
          </div>
          {ROLE_GROUPS.map((group, gi) => {
            // o que está na mão já aparece no topo
            const rest = group.roles.filter((role) => !held.has(role));
            if (rest.length === 0) return null;
            return (
              <div key={group.title} className={s.group} style={cascade(gi + 2)}>
                <h2 className={s.groupTitle}>{group.title}</h2>
                {rest.map((role) => roleButton(role, group, false))}
              </div>
            );
          })}
        </div>
      </section>
    </div>
  );
}

import { useMemo } from "react";
import type { Layout } from "../../app/useLayout";
import { copy } from "../../game/copy";
import { basicPrice, canSwap, priceOf, type Price } from "../../game/pricing";
import { BASIC_GROUP, ROLE_GROUPS, type ActionId } from "../../game/roles";
import { CLASSIC_RULES } from "../../game/rules.classic";
import { kamikazeAlly, selectFootlights, selectYou, type GameView } from "../../game/selectors";
import { Coin } from "../../ui/Coin/Coin";
import { CardBack, CardDead, CardFace } from "../../ui/Card/Card";
import { Timer } from "../../ui/Timer/Timer";
import { cx } from "../../ui/cx";
import { ActionButton } from "./ActionButton";
import s from "./ActionBoard.module.css";

type Props = {
  view: GameView;
  layout: Layout;
  onPick(action: ActionId): void;
};

/** Seu turno (tela 6): ações do turno no topo, personagens por grupo, preço de pricing.ts e o tempo da jogada no pé. */
export function ActionBoard({ view, layout, onPick }: Props) {
  const room = view.room;
  const you = useMemo(() => selectYou(view), [view]);
  const footlights = useMemo(() => selectFootlights(view), [view]);
  const ev = room?.gameEvent;
  const total = footlights.suddenDeath ? CLASSIC_RULES.waitingAction.suddenDeath : CLASSIC_RULES.waitingAction.default;
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

  return (
    <div className={cx(s.screen, layout === "desktop" && s.desk)}>
      <div className={s.top}>
        <span className={s.round}>{copy.roundTax(room?.turn ?? 0, footlights.tax)}</span>
        <span className={s.title}>{copy.yourTurnInScene}</span>
      </div>

      <section aria-label={copy.chooseAction} className={s.board}>
        <div className={s.head}>
          <h1 className={s.h1}>{copy.declareRole}</h1>
          <Coin size={50} value={you.coins} label={copy.youHave(you.coins)} />
        </div>

        {/* fileira só de leitura: escolher a carta própria continua na gaveta Você */}
        <section aria-label={copy.panelCards} className={s.hand} data-slot="hand">
          {you.cards.map((c) =>
            c.dead ? (
              <CardDead key={c.index} size="s" role={c.role} className={s.handCard} />
            ) : c.role ? (
              <CardFace
                key={c.index}
                size="s"
                role={c.role}
                protected={c.protected}
                bottom={c.protected ? copy.cardState.protected : undefined}
                className={s.handCard}
              />
            ) : (
              <CardBack key={c.index} size="s" className={s.handCard} />
            ),
          )}
        </section>

        <div className={s.list}>
          <div className={s.group}>
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
          {ROLE_GROUPS.map((group) => (
            <div key={group.title} className={s.group}>
              <h2 className={s.groupTitle}>{group.title}</h2>
              {group.roles.map((role) => {
                const price: Price = prices.roles.get(role) ?? { kind: "cost", coins: 0 };
                const offTurn = !!group.offTurn;
                const missing = offTurn ? null : short(price);
                // no seu turno o Kamikaze só serve se você matar um Kamikaze vivo seu (kamikazeValidation.go)
                const blocked = role === "KAMIKAZE" && kamikazeAlly(view) == null ? copy.needKamikaze : undefined;
                return (
                  <ActionButton
                    key={role}
                    role={role}
                    price={price}
                    held={!offTurn && held.has(role)}
                    offTurn={offTurn}
                    missing={blocked ? null : missing}
                    blocked={blocked}
                    onClick={onPick}
                  />
                );
              })}
            </div>
          ))}
        </div>

        {ev && (
          <div className={s.foot}>
            <Timer
              expiresAt={ev.expiresAt}
              total={total}
              label={copy.moveTime}
              decision
              caption={<span className={s.footLabel}>{copy.moveTime}</span>}
            />
          </div>
        )}
      </section>
    </div>
  );
}

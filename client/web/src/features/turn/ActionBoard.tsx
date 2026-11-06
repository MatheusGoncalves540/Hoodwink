import type { Layout } from "../../app/useLayout";
import { copy } from "../../game/copy";
import { basicPrice, canSwap, priceOf, type Price } from "../../game/pricing";
import type { RoleId } from "../../game/protocol";
import { BASIC_GROUP, ROLE_GROUPS, type ActionId } from "../../game/roles";
import { CLASSIC_RULES } from "../../game/rules.classic";
import { selectFootlights, selectYou, type GameView } from "../../game/selectors";
import { Coin } from "../../ui/Coin/Coin";
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
  const you = selectYou(view);
  const footlights = selectFootlights(view);
  const ev = room?.gameEvent;
  const total = footlights.suddenDeath ? CLASSIC_RULES.waitingAction.suddenDeath : CLASSIC_RULES.waitingAction.default;
  const held = (role: RoleId) => you.cards.some((c) => c.role === role && !c.dead);
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

        <div className={s.list}>
          <div className={s.group}>
            <h2 className={s.groupTitle}>{BASIC_GROUP.title}</h2>
            {BASIC_GROUP.actions.map((action) => {
              const price = room ? basicPrice(action, room, you.coins) : ({ kind: "cost", coins: 0 } as const);
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
                const price: Price = room ? priceOf(role, room) : { kind: "cost", coins: 0 };
                const offTurn = !!group.offTurn;
                const missing = offTurn ? null : short(price);
                return (
                  <ActionButton
                    key={role}
                    role={role}
                    price={price}
                    held={!offTurn && held(role)}
                    offTurn={offTurn}
                    missing={missing}
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

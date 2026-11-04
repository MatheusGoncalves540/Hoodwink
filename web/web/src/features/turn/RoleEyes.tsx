import type { Look } from "../../game/roles";

const almond = (x: number, h: number) =>
  `M${x - 30} 40Q${x} ${40 - h * 1.6} ${x + 30} 40Q${x} ${40 + h * 1.6} ${x - 30} 40Z`;

function pupil(x: number, look: Look, h: number) {
  const px = x + (look.side ?? 0);
  if (h < 3 || look.dollar) return "";
  if (look.star) return `M${px} 31l2 5 5 .5-4 3.5 1.3 5-4.3-2.7-4.3 2.7 1.3-5-4-3.5 5-.5z`;
  const r = look.small ? 3 : Math.min(7, h - 1);
  return `M${px - r} 40a${r} ${r} 0 1 0 ${2 * r} 0a${r} ${r} 0 1 0 ${-2 * r} 0`;
}

/**
 * Faixa de olhos provisória da prancha Acoes (arte final pendente).
 * Cores fixas da arte (iguais nos dois temas), como no desenho.
 */
export function RoleEyes({ look, band }: { look: Look; band: string }) {
  const left = look.wink ? 1 : look.h;
  const extra =
    (look.monocle ? "M126 40a26 26 0 1 0 52 0a26 26 0 1 0 -52 0M176 50Q196 76 222 58" : "") +
    (look.dollar
      ? [78, 152]
          .map((x) => `M${x + 5} 33Q${x} 29 ${x - 4} 33Q${x - 7} 37 ${x} 40Q${x + 7} 43 ${x + 4} 47Q${x} 51 ${x - 5} 47M${x} 28L${x} 52`)
          .join("")
      : "");
  const brows =
    (look.raise ? "M46 8Q78 -2 106 12" : `M48 ${20 - look.tilt}L104 ${20 + look.tilt}`) +
    `M122 ${20 + look.tilt}L178 ${20 - look.tilt}`;

  return (
    <svg aria-hidden="true" viewBox="0 0 230 76">
      <rect x="18" y="14" width="194" height="50" rx="25" fill={band} opacity="0.55" />
      <path d={almond(78, left)} fill="none" stroke="#F4E6C4" strokeWidth="2.5" />
      <path d={almond(152, look.h)} fill="none" stroke="#F4E6C4" strokeWidth="2.5" />
      <path d={pupil(78, look, left) + pupil(152, look, look.h)} fill="#F4E6C4" />
      <path d={brows} stroke="#F4E6C4" strokeWidth="4" strokeLinecap="round" />
      <path d={extra} fill="none" stroke="#B98A2E" strokeWidth="3" strokeLinecap="round" />
    </svg>
  );
}

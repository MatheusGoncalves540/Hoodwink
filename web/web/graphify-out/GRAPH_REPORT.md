# Graph Report - web  (2026-11-05)

## Corpus Check
- 99 files · ~105,881 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 470 nodes · 1455 edges · 12 communities (11 shown, 1 thin omitted)
- Extraction: 100% EXTRACTED · 0% INFERRED · 0% AMBIGUOUS · INFERRED: 5 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `818a4802`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- cx
- App.tsx
- selectors.ts
- fixtures.ts
- protocol.ts
- Table.tsx
- dependencies
- RoomMenu.tsx
- compilerOptions
- Frontend Básico com Google OAuth
- vite-env.d.ts

## God Nodes (most connected - your core abstractions)
1. `cx()` - 76 edges
2. `copy` - 33 edges
3. `fixture` - 26 edges
4. `roleName()` - 23 edges
5. `GameView` - 21 edges
6. `RoleId` - 18 edges
7. `Layout` - 17 edges
8. `useGameStore` - 17 edges
9. `selectYou()` - 15 edges
10. `useSessionStore` - 15 edges

## Surprising Connections (you probably didn't know these)
- `view()` --calls--> `fixture`  [EXTRACTED]
  src/features/sheets/Sheets.test.tsx → src/game/fixtures.ts
- `view()` --calls--> `fixture`  [EXTRACTED]
  src/features/turn/Turn.test.tsx → src/game/fixtures.ts
- `App()` --calls--> `useMotionPref()`  [EXTRACTED]
  src/app/App.tsx → src/app/useMotionPref.ts
- `App()` --calls--> `useTheme()`  [EXTRACTED]
  src/app/App.tsx → src/app/useTheme.ts
- `useLayout()` --calls--> `useMedia()`  [EXTRACTED]
  src/app/useLayout.ts → src/app/prefs.ts

## Import Cycles
- None detected.

## Communities (12 total, 1 thin omitted)

### Community 0 - "cx"
Cohesion: 0.08
Nodes (49): CORES, g(), MOVIMENTOS, Primitivos(), Layout, Props, Cast(), Props (+41 more)

### Community 1 - "App.tsx"
Cohesion: 0.07
Nodes (49): App(), isFixture(), isScene(), param(), Scene, SCENES, usePresentation(), useLayout() (+41 more)

### Community 2 - "selectors.ts"
Cohesion: 0.06
Nodes (57): WaitingRoom(), ChatSheet(), ChatSheetBody(), Props, DesktopRail(), Line, RoomMenuContent(), SceneLog() (+49 more)

### Community 3 - "fixtures.ts"
Cohesion: 0.05
Nodes (53): NOW, view(), Finale(), OutOfScene(), NOW, view(), NOW, view() (+45 more)

### Community 4 - "protocol.ts"
Cohesion: 0.06
Nodes (49): Props, SheetCard(), ActionBoard(), Props, f, NOW, out, ActionButton() (+41 more)

### Community 5 - "Table.tsx"
Cohesion: 0.08
Nodes (36): CURTAIN_DESK_HTML, CURTAIN_HTML, STAGE_DESK_HTML, STAGE_HTML, NOW, Props, SuddenDeathAlert(), SuddenDeathStage() (+28 more)

### Community 6 - "dependencies"
Cohesion: 0.05
Nodes (38): autoprefixer, gh-pages, dependencies, gh-pages, radix-ui, react, react-dom, @types/google.accounts (+30 more)

### Community 7 - "RoomMenu.tsx"
Cohesion: 0.26
Nodes (13): Amostras(), listeners, read(), setPref(), subscribe(), useMedia(), usePref(), useMotionPref() (+5 more)

### Community 8 - "compilerOptions"
Cohesion: 0.40
Nodes (4): google.accounts, compilerOptions, jsx, types

### Community 9 - "Frontend Básico com Google OAuth"
Cohesion: 0.50
Nodes (3): Frontend Básico com Google OAuth, Instalação, Rodar localmente

## Knowledge Gaps
- **125 isolated node(s):** `name`, `version`, `type`, `dev`, `build` (+120 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **1 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `cx()` connect `cx` to `App.tsx`, `selectors.ts`, `fixtures.ts`, `protocol.ts`, `Table.tsx`?**
  _High betweenness centrality (0.078) - this node is a cross-community bridge._
- **Why does `fixture` connect `fixtures.ts` to `App.tsx`, `selectors.ts`, `protocol.ts`, `Table.tsx`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **Why does `useSessionStore` connect `App.tsx` to `Table.tsx`, `RoomMenu.tsx`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **What connects `name`, `version`, `type` to the rest of the system?**
  _125 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `cx` be split into smaller, more focused modules?**
  _Cohesion score 0.07824561403508772 - nodes in this community are weakly interconnected._
- **Should `App.tsx` be split into smaller, more focused modules?**
  _Cohesion score 0.07003129890453834 - nodes in this community are weakly interconnected._
- **Should `selectors.ts` be split into smaller, more focused modules?**
  _Cohesion score 0.059076682316118935 - nodes in this community are weakly interconnected._
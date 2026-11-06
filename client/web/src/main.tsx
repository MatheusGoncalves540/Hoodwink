import ReactDOM from "react-dom/client";
// estilos globais primeiro: os CSS Modules dos componentes vêm depois e vencem empates de especificidade
import "./styles/index.css";
import { App } from "./app/App";
import { readPref, SKIN_KEY } from "./app/prefs";
import { applySkin } from "./skins/applySkin";
import { resolveEquip } from "./skins/equip";

// skin antes do primeiro render (chunks pequenos); se falhar, segue com o padrão
applySkin(document.documentElement, resolveEquip(readPref(SKIN_KEY)))
  .catch(() => undefined)
  .finally(() => ReactDOM.createRoot(document.getElementById("root")!).render(<App />));

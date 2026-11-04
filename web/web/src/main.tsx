import ReactDOM from "react-dom/client";
// estilos globais primeiro: os CSS Modules dos componentes vêm depois e vencem empates de especificidade
import "./styles/index.css";
import { App } from "./app/App";

ReactDOM.createRoot(document.getElementById("root")!).render(<App />);

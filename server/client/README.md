# Card Game Frontend

Este projeto é um front-end leve e modular para um jogo de cartas multiplayer, feito com React e TypeScript, utilizando tecnologias modernas com foco em desempenho e manutenibilidade.

## 🚀 Tecnologias Utilizadas

- **[React](https://reactjs.org/)** — biblioteca principal para interface
- **[TypeScript](https://www.typescriptlang.org/)** — tipagem estática
- **[Vite](https://vitejs.dev/)** — bundler moderno e rápido
- **[TailwindCSS](https://tailwindcss.com/)** — utilitário de CSS leve e produtivo
- **[Zustand](https://github.com/pmndrs/zustand)** — gerenciamento de estado global leve
- **WebSocket nativo** — comunicação em tempo real com o servidor

## 📁 Estrutura de Pastas

```
card-game-frontend/
├── public/                 # Arquivos estáticos (index.html)
├── src/
│   ├── assets/             # Imagens e recursos estáticos
│   ├── components/         # Componentes reutilizáveis (ex: Card, Button)
│   ├── features/           # Funcionalidades específicas (lobby, jogo, etc.)
│   ├── hooks/              # Custom hooks (ex: useSocket)
│   ├── pages/              # Páginas da aplicação (Home, Lobby, Game)
│   ├── state/              # Zustand stores
│   ├── utils/              # Funções auxiliares
│   ├── App.tsx             # Componente raiz
│   └── main.tsx            # Ponto de entrada da aplicação
├── index.html
├── package.json
├── vite.config.ts
├── tailwind.config.ts
├── postcss.config.js
├── tsconfig.json
├── Makefile
└── README.md
```

## 🛠️ Scripts via Makefile

- `make install`: instala dependências via `pnpm`
- `make dev`: roda o projeto em modo desenvolvimento
- `make build`: gera o build de produção
- `make preview`: roda o preview local do build
- `make deploy`: faz deploy com `gh-pages` (requer configuração)
- `make clean`: remove `node_modules` e `dist`

## 🧱 Estrutura Modular

- **Zustand** cuida do estado global de forma leve (ex: nickname, lista de jogadores)
- **Hooks** isolam lógica reusável (como WebSocket)
- **Tailwind** evita CSS excessivo e facilita prototipagem rápida
- **Componentes** são pensados para serem simples e reusáveis

## 🌐 Deploy

Este projeto pode ser facilmente hospedado de forma gratuita usando:

- [GitHub Pages](https://pages.github.com/)
- [Vercel](https://vercel.com/)
- [Netlify](https://netlify.com/)

Configure o `vite.config.ts` corretamente para o caminho do seu repositório caso use GitHub Pages.

---

Desenvolvido como base leve para projetos de front-end com foco em desempenho e modularidade.

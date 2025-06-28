# Hoodwink

Este projeto é o front-end leve e modular do jogo de cartas **Hoodwink**, feito com React e TypeScript, utilizando tecnologias modernas com foco em desempenho, modularidade e suporte a hot reload via Vite.

## 🚀 Tecnologias Utilizadas

- **React** — biblioteca principal para interface
- **TypeScript** — tipagem estática
- **Vite** — bundler moderno e rápido, com hot reload automático
- **TailwindCSS** — utilitário de CSS leve e produtivo
- **Zustand** — gerenciamento de estado global leve
- **WebSocket nativo** — comunicação em tempo real com o servidor

## 📁 Estrutura de Pastas

```
hoodwink/
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

> O Vite já oferece **hot reload automático** em modo desenvolvimento com `make dev`.

- `make install`: instala dependências via `pnpm`
- `make dev`: roda o projeto em modo desenvolvimento com hot reload
- `make build`: gera o build de produção
- `make preview`: roda o preview local do build
- `make deploy`: faz deploy com `gh-pages` (requer configuração)
- `make clean`: remove `node_modules` e `dist`

## 🌐 Deploy

Este projeto pode ser facilmente hospedado de forma gratuita usando:

- [GitHub Pages](https://pages.github.com/)
- [Vercel](https://vercel.com/)
- [Netlify](https://netlify.com/)

Configure o `vite.config.ts` corretamente para o caminho do seu repositório caso use GitHub Pages.

---

Desenvolvido como base leve para o projeto Hoodwink, com foco em desempenho, modularidade e suporte a hot reload via Vite.

# Regras para o Claude neste repositório

## Git

- **Nome de branch:** sempre `tipo/descricao-curta-em-kebab-case`, com o tipo igual ao dos commits:
  `feat/`, `fix/`, `bug/`, `chore/`, `docs/`, `refactor/`, `test/`, `style/`.
  Ex.: `feat/tela-camarim`, `fix/login-nao-fecha-modal`.
  **Nunca** use nomes gerados como `claude/<adjetivo>-<nome>-<hash>`. Se a sessão vier com uma branch assim,
  crie a branch no padrão acima antes do primeiro commit e trabalhe nela.
- **Commits:** Conventional Commits em português (`feat(web): ...`, `fix(engine): ...`).
- **Sem coautoria:** não adicione `Co-Authored-By: Claude ...` (nem outro trailer de coautor) nas mensagens de commit.

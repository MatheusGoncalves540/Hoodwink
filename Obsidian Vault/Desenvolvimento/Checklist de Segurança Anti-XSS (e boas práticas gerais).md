### ✅ 1. **Validação e sanitização de inputs**

Use bibliotecas como `go-playground/validator` para validar entradas.

Se você renderiza algum conteúdo do usuário (como HTML markdown), sanitize usando uma lib como [`bluemonday`](https://github.com/microcosm-cc/bluemonday).

---

### ✅ 2. **Escape de saída**

Nunca envie HTML montado diretamente com dados do usuário. No caso de enviar algo HTML no JSON, avise o front para sanitizar.

---

## 🧼 **Frontend (React SPA ou WebView)**

### ✅ 1. **Evite `dangerouslySetInnerHTML`**

Use somente se for absolutamente necessário, e sempre com `DOMPurify`:

```ts
import DOMPurify from 'dompurify';

const sanitized = DOMPurify.sanitize(htmlFromBackend);
<div dangerouslySetInnerHTML={{ __html: sanitized }} />
```

---

### ✅ 2. **Armazenamento de JWT**

| Opção | Segurança | Persistência | Uso universal |
| --- | --- | --- | --- |
| `localStorage` | ❌ médio | ✅ | ✅ |
| `memory` (state) | ✅ alto | ❌ | ✅ |
| `cookie httpOnly` | ✅ alto | ✅ | ❌ (só web) |

**Recomendado**:

-   Use `localStorage` com CSP + `DOMPurify`
    
-   Use JWT de vida curta + Refresh token (opcional)
    

---

### ✅ 3. **CSP (Content Security Policy)**

Configure via `<meta>` no React:

```html
<meta http-equiv="Content-Security-Policy" content="default-src 'self'; script-src 'self'">
```

Ou configure pelo seu reverse proxy (Caddy, NGINX) ou headers do Go.

---

### ✅ 4. **Evite bibliotecas de terceiros não confiáveis**

Scripts externos maliciosos podem ser vetor de XSS. Prefira:

-   CDN confiáveis (Google, jsDelivr, UNPKG)
    
-   Dependências auditadas
    

---

## 📱 **Mobile (WebView / React Native)**

### ✅ 1. **Desative JavaScript onde não for necessário**

### ✅ 2. **Não use `addJavascriptInterface` (Android nativo)**

### ✅ 3. **Valide todo conteúdo renderizado em WebViews**

Use `DOMPurify` mesmo em WebViews (se renderizar HTML).

---

## 🖥️ **Electron**

### ✅ 1. **Nunca use `nodeIntegration` com UI**

```js
webPreferences: {
  nodeIntegration: false,
  contextIsolation: true,
  enableRemoteModule: false,
}
```

### ✅ 2. **Use CSP**

Mesmo em janelas locais, CSP ajuda a bloquear execuções indesejadas.

### ✅ 3. **Use `keytar` ou `electron-store` para guardar tokens**

Evite salvar JWT em `localStorage` dentro do Electron.

---

## 📋 **Checklist rápido para testar XSS**

| Teste | O que fazer? | Ferramenta |
| --- | --- | --- |
| HTML Injection | Tente injetar `<script>` ou `<img onerror>` em campos do usuário | Manual |
| Policy Headers | Verifique CSP, X-Frame-Options etc. | [securityheaders.com](https://securityheaders.com) |
| Lint front | Verifique uso inseguro de `innerHTML`, `eval`, etc. | ESLint |
| Scan de vulnerabilidade | Faça scan da API e do front | [ZAP](https://www.zaproxy.org/), OWASP Dependency-Check |

---

## 🧰 Sugestão de ferramentas para automatizar

-   ✅ `DOMPurify` → sanitização HTML no front
    
-   ✅ `bluemonday` → sanitização no Go
    
-   ✅ `helmet` (se usar Node para front)
    
-   ✅ `ZAP CLI` para testes de XSS automatizados
    

---

## 🔄 Bônus: fluxo com Refresh Token (opcional)

Se quiser segurança adicional:

1.  Armazene o access token no `memory` ou `localStorage`
    
2.  Guarde o `refresh_token` no `cookie httpOnly`
    
3.  Crie endpoint `/auth/refresh` que gera novo access token
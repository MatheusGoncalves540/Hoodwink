## ✅ Visão geral: onde guardar o JWT?

| Plataforma | Armazenamento Recomendado | Segurança | Observações |
| --- | --- | --- | --- |
| **Web (SPA/React)** | `Memory` (em runtime) ou `localStorage` ou `HTTP-only cookie` | 🟡 / 🔒 | Depende do modelo |
| **Mobile (Android/iOS)** | `SecureStorage`, `Keystore`, `Keychain` | 🔒 | APIs nativas |
| **Desktop (Electron, CLI)** | Disco criptografado / memória / config interna | 🔒 / 🟡 | Evite plaintext |

---

## 🔹 1. **Web (React, Vite, SPA)**

### 🔸 Opção A: **Armazenar em memória (recomendado para apps sensíveis)**

```ts
let jwtToken = null;

// após login:
jwtToken = resposta.token;
```

-   ✅ Alta segurança (some ao fechar aba)
    
-   ❌ Perde ao recarregar
    

### 🔸 Opção B: **LocalStorage**

```ts
localStorage.setItem("jwt", token);
```

-   ✅ Persistente
    
-   ❌ Vulnerável a XSS → use só com atenção
    

### 🔸 Opção C: **HTTP-only cookies** (com `Secure`, `SameSite=Strict`)

```go
http.SetCookie(w, &http.Cookie{
	Name:     "jwt",
	Value:    token,
	HttpOnly: true,
	Secure:   true,
	SameSite: http.SameSiteStrictMode,
})
```

-   ✅ Alta segurança contra XSS
    
-   ✅ Token vai automático com cada request
    
-   ❌ Não visível pelo JavaScript → difícil de manipular diretamente
    
-   ✅ Pode ser útil se o back-end quiser controlar a sessão com mais transparência
    

> 🧠 **Conclusão Web:**  
> Para apps simples, `localStorage` com boas práticas pode servir.  
> Para apps sensíveis, combine **cookie httpOnly + fallback na memória**.

---

## 🔹 2. **Mobile (Android/iOS)**

### ✅ Use o storage seguro nativo:

-   **Android**: `EncryptedSharedPreferences` ou `Keystore`
    
-   **iOS**: `Keychain`
    

Exemplo em Flutter:

```dart
final storage = new FlutterSecureStorage();
await storage.write(key: "jwt", value: token);
```

> Isso protege contra **root/jailbreak** e leitura indevida.

---

## 🔹 3. **Desktop App (Electron, CLI, etc)**

### Opções:

-   `Electron`: `secure-store` ou arquivo criptografado com `node-keytar`
    
-   CLI Go: Salvar em um arquivo `.hoodwink_config` com criptografia local (por exemplo, usando `AES + salt`)
    

> ✅ Nunca deixe em plaintext em disco, se puder evitar

---

## ✅ Como usar o JWT após armazenado?

Em qualquer cliente:

```http
GET /minha-rota-protegida
Authorization: Bearer <TOKEN_JWT>
```

No seu Go backend:

```go
authHeader := r.Header.Get("Authorization")
// extrai o token e valida com ValidateJWT(...)
```

---

## 🔐 Dica de segurança final

-   Tokens devem **expirar rapidamente** (ex: `15min - 24h`)
    
-   Use um **JWT de refresh** se quiser sessões longas (ex: `refresh_token` criptografado)
    
-   Use `Secure`, `SameSite=Strict` e `HttpOnly` nos cookies, se usá-los
    
-   Sempre valide assinatura e expiração no backend
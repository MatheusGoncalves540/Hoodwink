Você pode salvar o JWT de 3 formas:

| Método                       | Persistência   | Segurança                | Ideal para                     |
| ---------------------------- | -------------- | ------------------------ | ------------------------------ |
| `localStorage`               | Alta           | Média (acessível via JS) | SPAs simples / controle manual |
| `sessionStorage`             | Até fechar aba | Média                    | Sessões curtas                 |
| **Memória (em React state)** | Volátil        | Alta (mais segura)       | Apps sensíveis + token curto   |
| **Cookies (HTTP-only)**      | Alta           | Alta (backend envia)     | Só se backend configurar       |

> 🔐 **Recomendado**: use `localStorage` para facilidade e `Authorization` para envio seguro no header.

---

### 💾 **2\. Salvando o JWT após login**

```ts
// Exemplo após o login
fetch("http://localhost:8080/auth/external/google", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ id_token }),
})
  .then(res => res.json())
  .then(data => {
    if (data.token) {
      // Salva o JWT no localStorage
      localStorage.setItem("auth_token", data.token);
    }
  });
```

---

### 📤 **3\. Usando o token nas próximas requisições**

Você precisa usar o header `Authorization: Bearer <token>`, exemplo:

```ts
const token = localStorage.getItem("auth_token");

fetch("http://localhost:8080/create-room", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "Authorization": `Bearer ${token}`,
  },
  body: JSON.stringify({ nome: "Sala 1" }),
});
```

---

### ✅ Exemplo completo com Hook React

```ts
// utils/auth.ts
export const saveToken = (token: string) => {
  localStorage.setItem("auth_token", token);
};

export const getToken = () => {
  return localStorage.getItem("auth_token");
};

export const logout = () => {
  localStorage.removeItem("auth_token");
};
```

```ts
// services/api.ts
export const authFetch = (url: string, options = {}) => {
  const token = localStorage.getItem("auth_token");

  return fetch(url, {
    ...options,
    headers: {
      ...(options as any).headers,
      "Authorization": `Bearer ${token}`,
      "Content-Type": "application/json",
    },
  });
};
```

---

## 🧪 Teste no Insomnia / Bruno

Use sempre:

```makefile
Authorization: Bearer <seu_jwt>
```

---

## ⚠️ Segurança extra (importante no futuro)

Se quiser segurança ainda maior:

-   Use `accessToken` + `refreshToken`
    
-   Salve o `accessToken` na memória
    
-   Use cookies **HTTP-only** apenas com backends configurados para isso
    
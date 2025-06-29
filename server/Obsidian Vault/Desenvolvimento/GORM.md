## 📦 1. Migrations no GORM

### ✅ O que tem:

-   **AutoMigrate**: cria e atualiza tabelas automaticamente com base nas structs.
    
-   Suporte a **constraints**, `indexes`, `foreign keys`, etc via tags.
    

### ⚠️ O que não tem:

-   Não versiona nem reverte migrations (diferente de Sequelize, TypeORM).
    
-   Não gera arquivos `.sql` automaticamente.
    

### 🔧 Exemplo de uso:

```go
import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Post{},
		// outros modelos
	)
}
```

📌 Você pode rodar isso no `main.go` ou em um `cmd/migrate` separado.

---

## 🌱 2. Seeds no GORM

Não existe um sistema de `seeders` pronto como no Sequelize, mas você pode criar uma função como esta:

```go
func SeedUsers(db *gorm.DB) error {
	users := []User{
		{Name: "Matheus", Email: "matheus@dev.com"},
		{Name: "Joana", Email: "joana@dev.com"},
	}

	for _, user := range users {
		err := db.FirstOrCreate(&user, User{Email: user.Email}).Error
		if err != nil {
			return err
		}
	}
	return nil
}
```

🔁 Pode ser executado em `main.go`, em um comando CLI, ou `make seed`.

---

## 🔒 3. Como deixar o GORM mais seguro

Embora o GORM já escape parâmetros automaticamente e evite SQL Injection, você pode e **deve fortalecer a segurança** com as boas práticas abaixo:

### ✅ 1. **Evite SQL bruto com interpolação**

Evite:

```go
db.Exec("DELETE FROM users WHERE name = " + name) // ❌ perigo
```

Use:

```go
db.Exec("DELETE FROM users WHERE name = ?", name) // ✅ seguro
```

### ✅ 2. **Use `.Where()` com bindings**

```go
db.Where("email = ?", email).First(&user)
```

### ✅ 3. **Valide structs com `go-playground/validator`**

```go
type UserInput struct {
  Name  string `validate:"required,min=3"`
  Email string `validate:"required,email"`
}

validate := validator.New()
err := validate.Struct(input)
```

### ✅ 4. **Use Soft Deletes se apropriado**

```go
type User struct {
  gorm.Model // já inclui ID, CreatedAt, UpdatedAt, DeletedAt
}
```

Com isso, `Delete` não remove do banco fisicamente, apenas marca como excluído.

### ✅ 5. **Restrições de unicidade no GORM**

```go
Email string `gorm:"uniqueIndex"`
```

### ✅ 6. **Gerencie transações corretamente**

```go
tx := db.Begin()
err := tx.Create(&User{Name: "Test"}).Error
if err != nil {
	tx.Rollback()
	return err
}
tx.Commit()
```

---

## 🎯 Ferramentas auxiliares que você pode usar junto

| Ferramenta | Finalidade |
| --- | --- |
| [`golang-migrate`](https://github.com/golang-migrate/migrate) | Gerenciador de migrations `.sql`, versão, rollback |
| [`go-playground/validator`](https://github.com/go-playground/validator) | Validação de structs antes de salvar |
| \[`uber-go/zap` ou `rs/zerolog`\] | Logging seguro de queries |
| [`sqlc`](https://sqlc.dev) | Caso queira misturar GORM com queries SQL type-safe |

---

## ✅ Resumo

| Recurso | Suporte nativo em GORM? | Comentário |
| --- | --- | --- |
| Migrations | 🔶 Sim (AutoMigrate) | Mas sem versionamento ou rollback |
| Seeds | 🔸 Manual | Fácil de implementar |
| Segurança | ✅ Sim | Seguro por padrão se usar direito |
| Validação | ❌ Não nativo | Use com `validator` externo |
| SQL Injection | ✅ Protegido por `?` | Evite interpolação manual |

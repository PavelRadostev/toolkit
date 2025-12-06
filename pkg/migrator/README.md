# Migrator

Инструмент для управления миграциями базы данных PostgreSQL.

## Использование

### Интеграция в приложение

Чтобы использовать migrator в вашем приложении, добавьте вызов `migrator.Execute()` в `main.go`:

```go
package main

import (
	"github.com/PavelRadostev/toolkit/pkg/migrator"
)

func main() {
	// ... ваш код ...
	
	// Запуск CLI миграций
	migrator.Execute()
}
```

### Структура миграций

Миграции должны находиться в директории, указанной в `config/settings.yaml` (по умолчанию `migrations`).

Формат имен файлов миграций:
- `{version}_{name}.up.sql` - для применения миграции
- `{version}_{name}.down.sql` - для отката миграции

Пример:
```
migrations/
  000001_create_users_table.up.sql
  000001_create_users_table.down.sql
  000002_add_email_to_users.up.sql
  000002_add_email_to_users.down.sql
```

### Команды

#### Применить все миграции
```bash
go run cmd/main.go migrate up
```

#### Откатить последнюю миграцию
```bash
go run cmd/main.go migrate down
```

#### Показать текущую версию
```bash
go run cmd/main.go migrate version
```

## Конфигурация

Путь к директории с миграциями настраивается в `config/settings.yaml`:

```yaml
migration:
  dir: "migrations"
```

Путь указывается относительно корня проекта.


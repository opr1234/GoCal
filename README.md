# 🚀 Distributed Calculator Service

Микросервис для вычисления арифметических выражений с поддержкой многопользовательского режима и персистентностью данных.

## 🛠 Установка

### Требования

- Go 1.21+
- PostgreSQL 14+
- protoc (для компиляции proto-файлов)

# Клонирование репозитория

```bash
git clone https://github.com/opr1234/calculator.git
cd calculator
```

# Установка зависимостей

```bash
go mod tidy
```

## 🔧 Настройка PostgreSQL

1. Создайте БД и пользователя:

```sql
CREATE DATABASE calculator;
CREATE USER calc_user WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE calculator TO calc_user;
```

2. Настройте переменные окружения (`.env`):

```ini
DB_HOST=localhost
DB_PORT=5432
DB_USER=calc_user
DB_PASSWORD=secure_password
DB_NAME=calculator
DB_SSLMODE=disable
JWT_SECRET=your_super_secret_key
```

## 🗄 Схема базы данных

### Таблица `users`
| Поле            | Тип        | Описание                |
|------------------|------------|-------------------------|
| id              | SERIAL     | Первичный ключ         |
| login           | VARCHAR(50)| Уникальный логин       |
| password_hash   | TEXT       | Хеш пароля            |
| created_at      | TIMESTAMP  | Время создания         |

### Таблица `expressions`
| Поле            | Тип        | Описание                |
|------------------|------------|-------------------------|
| id              | SERIAL     | Первичный ключ         |
| user_id         | INTEGER    | Ссылка на пользователя|
| expression      | TEXT       | Выражение для вычисления|
| status          | VARCHAR(20)| Статус (pending/completed/error)|
| result          | FLOAT      | Результат вычисления   |
| created_at      | TIMESTAMP  | Время создания         |
| updated_at      | TIMESTAMP  | Время обновления       |

```sql
-- Создание таблиц
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(50) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE expressions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    expression TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    result FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 📋 Примеры SQL-запросов

### 1. Создание пользователя

```sql
INSERT INTO users (login, password_hash) 
VALUES ('alice', 'hashed_password')
RETURNING *;
```

### 2. Добавление выражения

```sql
INSERT INTO expressions (user_id, expression)
VALUES (1, '2+2*2')
RETURNING id, status, created_at;
```

### 3. Получение выражений пользователя

```sql
SELECT e.id, e.expression, e.status, e.result, e.created_at
FROM expressions e
JOIN users u ON e.user_id = u.id
WHERE u.login = 'alice'
ORDER BY e.created_at DESC;
```

### 4. Обновление статуса выражения

```sql
UPDATE expressions 
SET 
    status = 'completed',
    result = 6,
    updated_at = CURRENT_TIMESTAMP
WHERE id = 1;
```

## 🚦 Использование API

### Регистрация пользователя

```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"login": "alice", "password": "qwerty123"}'
```

### Получение токена

```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"login": "alice", "password": "qwerty123"}'
```

### Отправка выражения

```bash
curl -X POST http://localhost:8080/api/v1/calculate \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"expression": "(2+3)*4/2"}'
```

## 🔄 Миграции

Миграции выполняются автоматически при старте приложения с использованием встроенной системы миграций.


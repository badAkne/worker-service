# MoM Boilerplate V2

Шаблон для создания Go микросервисов.

## Технологии

| Компонент | Технология |
|-----------|------------|
| HTTP Server | `net/http` + `gorilla/mux` |
| ORM | `uptrace/bun` (PostgreSQL) |
| Логирование | `zerolog` |
| Конфигурация | `envconfig` + `godotenv` |
| CLI | `urfave/cli/v2` |
| Валидация | `go-playground/validator` |
| Proto | `buf` |
| Линтер | `golangci-lint` v2 |

## Требования

| Зависимость | Версия |
|-------------|--------|
| [Go](https://go.dev) | 1.24+ |
| [Docker](https://docker.com) | 20+ |
| make | - |

## Быстрый старт

### 1. Клонирование шаблона

```bash
git clone --depth=1 git@github.com:MoM-Repo/worker-service.git YOUR_PROJECT_NAME
cd YOUR_PROJECT_NAME
```

### 2. Настройка проекта

```bash
chmod +x setup.sh
./setup.sh your-project-name your-github-username
```

Скрипт автоматически:
- Обновит все импорты в Go файлах
- Изменит `go.mod` на новый модуль
- Обновит конфигурационные файлы

### 3. Инициализация репозитория

```bash
rm -rf .git
git init
git add .
git commit -m "Initial commit"
git remote add origin git@github.com:your-username/your-project-name.git
git push -u origin main
```

### 4. Запуск

```bash
# Поднять PostgreSQL
make up

# Применить миграции
make migrate

# Запустить сервер
make run
```

Сервер доступен на `http://localhost:8080`

## Доступные команды

```bash
make help  # Показать все команды
```

### Разработка
| Команда | Описание |
|---------|----------|
| `make run` | Запустить приложение |
| `make build` | Собрать бинарник |
| `make test` | Запустить тесты |

### Качество кода
| Команда | Описание |
|---------|----------|
| `make lint` | Запустить линтер |
| `make lint-fix` | Линтер + автофикс |

### Proto
| Команда | Описание |
|---------|----------|
| `make proto-lint` | Линтинг proto файлов |
| `make proto-format` | Форматирование proto |
| `make proto-gen` | Генерация Go кода |

### Окружение
| Команда | Описание |
|---------|----------|
| `make up` | Поднять PostgreSQL |
| `make down` | Остановить контейнеры |
| `make logs` | Показать логи |
| `make migrate` | Применить миграции |

### CI/CD
| Команда | Описание |
|---------|----------|
| `make ci` | Все CI проверки |
| `make ci-full` | CI + Docker build |
| `make docker-build` | Собрать Docker образ |

## Структура проекта

```
.
├── .github/workflows/      # GitHub Actions CI
├── api/
│   └── proto/              # Proto файлы
├── cmd/                    # CLI команды
├── internal/
│   ├── app/
│   │   ├── builder/        # Dependency injection
│   │   ├── config/         # Конфигурация
│   │   ├── entity/         # Модели/DTO
│   │   ├── handler/        # HTTP обработчики
│   │   ├── processor/      # Процессоры (HTTP server, migrations)
│   │   ├── repository/     # Слой данных
│   │   └── util/           # Утилиты
│   └── pkg/
│       ├── constant/       # Константы
│       └── http/           # HTTP утилиты
│           ├── binding/    # Парсинг запросов
│           ├── httph/      # HTTP хелперы
│           ├── mzerolog/   # Логирование middleware
│           └── respondent/ # Форматирование ответов
├── migration/
│   └── postgres/           # SQL миграции
├── buf.yaml                # Конфигурация buf
├── buf.gen.yaml            # Генерация proto
├── docker-compose.local.yml
├── Dockerfile
├── Makefile
└── setup.sh                # Скрипт настройки
```

## Конфигурация

Переменные окружения (см. `.env.dist`):

```env
# Database
APP_REPOSITORY_POSTGRES_ADDRESS=127.0.0.1:5432
APP_REPOSITORY_POSTGRES_USERNAME=postgres
APP_REPOSITORY_POSTGRES_PASSWORD=postgres
APP_REPOSITORY_POSTGRES_NAME=app_db
APP_REPOSITORY_POSTGRES_READ_TIMEOUT=30s
APP_REPOSITORY_POSTGRES_WRITE_TIMEOUT=30s

# Web Server
APP_PROCESSOR_WEB_SERVER_LISTEN_PORT=8080

# Monitor
APP_MONITOR_LOG_LEVEL=debug
APP_MONITOR_ENVIRONMENT=development
```

## API Endpoints

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/health` | Health check |
| POST | `/v1/example` | Пример (JSON body) |
| GET | `/v1/example` | Пример (query params) |

### Примеры запросов

```bash
# Health check
curl http://localhost:8080/health

# POST /v1/example (JSON body)
curl -X POST http://localhost:8080/v1/example \
  -H "Content-Type: application/json" \
  -d '{"name": "test", "count": 5}'

# GET /v1/example (query params)
curl "http://localhost:8080/v1/example?name=test&count=5"
```

## CI/CD

GitHub Actions выполняет:
- Build
- Test (с coverage)
- Lint (golangci-lint)
- Mod check (go.mod/go.sum)
- Proto lint (buf)
- Docker build

# Архитектура интерактивной онлайн-доски для совместной работы

## 1. Введение

Проект представляет собой **интерактивную онлайн-доску для совместной работы** (аналог Miro), которая позволяет нескольким пользователям одновременно работать на одной доске в реальном времени.  
Основные функции включают:
- создание и редактирование досок;
- добавление графических и текстовых элементов;
- совместное редактирование через WebSocket;
- комментарии, приглашения, экспорт/импорт данных;
- систему ролей (Guest, Edit, Owner).

Система построена на **микросервисной архитектуре** с использованием **Go (Golang)** на бэкенде и **React** на фронтенде.  
Для хранения данных используется **PostgreSQL** и **S3-совместимое хранилище**, а обмен событиями осуществляется через **Kafka**.

---

## 2. Архитектура системы (Frontend–Backend–DB–API)

Архитектура основана на микросервисном подходе с разделением по зонам ответственности.  
Клиент взаимодействует с системой через **API Gateway**, который маршрутизирует запросы к микросервисам.

### Схема архитектуры

```mermaid
flowchart LR
    %% Колонка 1: клиенты
    subgraph Clients[Клиенты]
        A1[Web App]
    end

    %% Колонка 2: API
    subgraph Gateway[API Gateway / BFF]
        GW[REST/gRPC API]
    end

    %% Колонка 3: микросервисы
    subgraph Services[Микросервисы на Go]
        direction TB
        Auth[AuthService]
        Board[BoardService]
        Collab[CollabService]
        File[FileService]
        Payment[PaymentService]
        Admin[AdminService]
        Notify[NotificationService]
    end

    %% Колонка 4: данные
    subgraph Data[Хранилища и события]
        direction TB
        PG[(PostgreSQL)]
        Redis[(Redis)]
        S3[(S3 Object Storage)]
        Kafka[(Kafka EventBus)]
    end

    %% Колонка 5: фоновые задачи
    subgraph Jobs[Фоновые задачи]
        Cron[CronJobs]
    end

    %% Колонка 6: внешние консьюмеры
    subgraph Consumers[Внешние консьюмеры]
        Push[Push Service]
        Carrot[Carrot Quest]
        Mix[Mixpanel]
        Dash[Analytics/Dashboards]
        Tpay[T-pay]
    end

    %% Связи верхнеуровневые
    A1 --> GW

    GW --> Services
    Services --> Data
    Services --> Kafka
    Kafka --> Services

    Jobs --> Data
    Jobs --> Kafka

    Kafka --> Consumers
    Payment --> Tpay
```

### Описание уровней

- **Frontend** — React-приложение на TypeScript, использует REST и WebSocket для синхронизации.
- **Backend** — микросервисы на Go, взаимодействие по gRPC и через Kafka.
- **API Gateway** — единая точка входа с проверкой токенов и маршрутизацией.
- **Хранилища** — PostgreSQL для метаданных, Redis для сессий, S3 для версий досок.
- **Event Bus (Kafka)** — публикация и потребление доменных событий.
- **Внешние интеграции** — Keycloak, Tpay, CDN, аналитика, уведомления.

---

## Основные сущности и таблицы (PostgreSQL)

```sql
CREATE TYPE role_type AS ENUM ('owner', 'admin', 'editor', 'viewer');

CREATE TABLE board (
    id              UUID PRIMARY KEY,
    team_id         UUID NOT NULL,
    owner_id        UUID NOT NULL,
    title           TEXT NOT NULL,
    description     TEXT,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now(),
    is_archived     BOOLEAN DEFAULT false,
    version_ref     TEXT -- ссылка на ключ в S3 с JSON-схемой текущей версии
);

CREATE TABLE account (
    id              UUID PRIMARY KEY,
    email           CITEXT UNIQUE NOT NULL,
    password_hash   TEXT,
    display_name    TEXT,
    created_at      TIMESTAMPTZ DEFAULT now(),
    last_login_at   TIMESTAMPTZ,
    is_active       BOOLEAN DEFAULT true
);

CREATE TABLE team (
    id              UUID PRIMARY KEY,
    name            TEXT NOT NULL,
    owner_id        UUID NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE team_mate (
    id              UUID PRIMARY KEY,
    team_id         UUID NOT NULL,
    account_id      UUID NOT NULL,
    role            role_type NOT NULL,    
    created_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE(team_id, account_id)
);

CREATE TABLE event (
    id              BIGSERIAL PRIMARY KEY,
    board_id        UUID NOT NULL,
    account_id      UUID,
    event_type      TEXT NOT NULL, -- create_object, update_object, comment и т.п.
    payload         JSONB NOT NULL, -- параметры события
    created_at      TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_event_board ON event(board_id, created_at);

CREATE TABLE account_settings (
    account_id      UUID PRIMARY KEY,
    locale          TEXT DEFAULT 'en',
    theme           TEXT DEFAULT 'light',
    notifications   JSONB DEFAULT '{}' -- гибкие настройки уведомлений
);

CREATE TABLE invite (
    id              UUID PRIMARY KEY,
    account_id      UUID NOT NULL,
    team_id         UUID NOT NULL,
    email           CITEXT NOT NULL,
    role            role_type NOT NULL,
    token           TEXT NOT NULL UNIQUE,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT now(),
    accepted        BOOLEAN DEFAULT false
);
```

## ER-диаграмма основных сущностей

```mermaid
erDiagram
    ACCOUNT ||--o{ TEAM : "owns"
    ACCOUNT ||--o{ TEAM_MATE : "member of"
    TEAM ||--o{ TEAM_MATE : "has"
    TEAM ||--o{ BOARD : "contains"
    ACCOUNT ||--o{ BOARD : "owns"
    BOARD ||--o{ EVENT : "generates"
    ACCOUNT ||--o{ EVENT : "performs"
    ACCOUNT ||--|| ACCOUNT_SETTINGS : "has"
    TEAM ||--o{ INVITE : "issues"
```

## JSON-схема доски

```json
{
  "board_id": "string",
  "last_event": "int64", 
  "entities": [
    {
      "id": "string",
      "type": "string", 
      "data": {}
    }
  ]
}
```
## JSON-схема эвента ws

```json
{
  "board_id": "string",
  "entity": {
    "id": "string",
    "type": "string"
  },
  "action": "create | update | delete",
  "payload": {}
}
```
---

## 4. Диаграмма компонентов

```mermaid
flowchart TB
    subgraph Core[Система]
        BoardService
        CollabService
        FileService
        PaymentService
        NotificationService
        AdminService
    end
    subgraph External[Внешние системы]
        Keycloak
        Tpay
        CarrotQuest
        Mixpanel
        FCM[Push/Email/Chats]
        CDN
        Prometheus
        ELK
    end
    subgraph Storage[Хранилища]
        PostgreSQL
        Redis
        S3
        Kafka
    end
    BoardService --> PostgreSQL
    CollabService --> Redis
    FileService --> S3
    PaymentService --> PostgreSQL
    PaymentService --> Tpay
    NotificationService --> Kafka
    Kafka --> CarrotQuest
    Kafka --> Mixpanel
    Kafka --> FCM
    BoardService --> Kafka
    CollabService --> Kafka
    FileService --> Kafka
    Core --> Prometheus
    Core --> ELK
    FileService --> CDN
    Core --> Keycloak
```

---

## 5. Техническое обоснование выбранных решений

| Компонент                | Технология          | Обоснование                                                      |
|--------------------------|---------------------|------------------------------------------------------------------|
| **Frontend**             | React + TypeScript  | Высокая скорость разработки, SSR/CSR, JSX, зрелая экосистема.    |
| **Backend**              | Go (Golang)         | Простота, высокая производительность, поддержка многопоточности. |
| **API**                  | gRPC + REST         | gRPC — для микросервисов, REST — для фронтенда.                  |
| **PostgreSQL**           | Реляционная БД      | ACID, JSONB, масштабируемость.                                   |
| **Redis**                | In-memory cache     | Быстрый доступ к состоянию пользователей и presence.             |
| **S3**                   | Объектное хранилище | Хранение JSON-снимков и медиа с версионированием.                |
| **Kafka**                | Шина событий        | Асинхронная обработка событий и интеграция с аналитикой.         |
| **Keycloak**             | OpenID Connect IdP  | Централизованная авторизация и ролевая модель (RBAC).            |
| **Tpay**                 | Платёжный шлюз      | Рекуррентные платежи, вебхуки, токенизация карт.                 |
| **Prometheus / Grafana** | Мониторинг          | Метрики и визуализация состояния системы.                        |
| **ELK / Loki**           | Логи и аудит        | Централизованный сбор логов и DLP.                               |

---

## 6. Заключение

Архитектура системы обеспечивает:

- **Масштабируемость** — за счёт микросервисной структуры и контейнеризации (Kubernetes).
- **Надёжность** — PostgreSQL, Redis и S3 обеспечивают отказоустойчивость.
- **Гибкость** — новые микросервисы добавляются без влияния на существующие.
- **Реальное время** — через WebSocket и CollabService с Redis.
- **Устойчивость к сбоям** — Kafka гарантирует доставку событий и интеграцию с аналитикой.

Система поддерживает до **1000 активных пользователей** при задержках API менее 150 мс.  
Её можно масштабировать горизонтально, добавляя ноды и консьюмеры Kafka.  
Такая архитектура сочетает производительность, гибкость и простоту сопровождения, что делает её устойчивой к росту нагрузки и расширению функционала.

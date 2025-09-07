# Описание программного обеспечения системы

## Обзор
Система интерактивной онлайн-доски построена на микросервисной архитектуре (Go) с фронтендом на React.  
Коммуникации: REST/gRPC через API Gateway, realtime — WebSocket.  
Состояние досок хранится как JSON-снимки в S3, доступ к медиа — по presigned URL (раздача через CDN).  
События и интеграции — через Kafka.

## Frontend

- **Язык/Фреймворк:** TypeScript, React
- **Рендеринг доски:** HTML5 Canvas API (для отрисовки объектов, кривых, drag&drop)
- **Сборка:** Vite или Next.js (возможность SSG/SSR)
- **Realtime:** WebSocket (нативный клиент для синхронизации событий)
- **State management:** Zustand или Redux Toolkit (управление состоянием), React Query для работы с данными API
- **Роутинг:** React Router / Next Router
- **UI библиотека:** TailwindCSS + Radix/Headless UI
- **i18n:** i18next
- **Ошибки и перфоманс:** Sentry (frontend only)
- **Тестирование:** Vitest или Jest + Testing Library, Playwright (e2e)  

## Backend (Go)

- **Язык:** Go (Golang)
- **Фреймворк HTTP:** grpc gateway
- **RPC:** gRPC (protoc), REST (OpenAPI)
- **Аутентификация:** Gatekeeper + Keycloak (OIDC/JWT), RBAC
- **Данные:** PostgreSQL (через pgx), Redis (sessions/presence)
- **Хранилище:** S3-совместимое (JSON-снимки досок и медиа; версионирование включено)
- **События:** Kafka (Sarama)
- **Миграции БД:** golang-migrate
- **Конфигурации:** Viper (env + YAML), поддержка feature-flags
- **Валидаторы:** go-playground/validator
- **Тестирование:** testify, httptest, gomock
- **Observability:** Prometheus (метрики), OpenTelemetry/Jaeger (трейсинг), ELK (логи)  

## Интеграции

- **Идентификация**
    - **Keycloak (OIDC/SAML) + Gatekeeper** — регистрация/вход, валидация JWT, выдача ролей (RBAC).
    - Потоки: *Frontend → API Gateway → Keycloak* (login/refresh).

- **Платежи**
    - **Tpay** — прямые и рекуррентные платежи.
    - Исходящие: *PaymentService → Tpay API* (инициация/статусы).
    - Входящие: *Tpay → API Gateway (webhook)* → маршрутизация в PaymentService.
    - Идемпотентность: `event_id` + transactional outbox.

- **Уведомления**
    - **FCM/APNs/Email/Chats (Max/Telegram)** — Push/email/чат-нотификации.
    - Потоки: *сервисы публикуют события → Kafka → NotificationConsumer → провайдеры*.
    - Шаблоны и локализация сообщений на стороне NotificationService.

- **Аналитика**
    - **Carrot Quest, Mixpanel** (+ возможность подключать другие).
    - Потоки: *Kafka → аналитические консьюмеры → провайдеры*.
    - PII-минимизация, анонимизация user_id при необходимости.

- **Хранилище и CDN**
    - **S3-совместимое** — JSON-снимки досок и медиа, включено версионирование бакета.
    - **CDN** — раздача напрямую клиенту по **presigned URL**, которые генерирует FileService.

- **Мониторинг и логи**
    - **Prometheus/Grafana** — метрики и алерты.
    - **ELK + DLP** — централизованные логи и маскирование чувствительных полей.
    - **Sentry (Frontend only)** — ошибки и перфоманс UI/JS.

- **Шина событий**
    - **Kafka** — доменные события и интеграционные потоки.
    - Схемы: **Schema Registry** (JSON-Schema/Proto).

## Сервисы и назначение

| Сервис / Компонент       | Назначение                                                                                                      | Ключевые технологии                                 |
|--------------------------|-----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| **API Gateway / BFF**    | Единая точка входа (REST/gRPC), маршрутизация, аутентификация, rate-limit, валидация вебхуков (Tpay).           | Go, chi/gin, OpenAPI, gRPC                          |
| **AuthService**          | Интеграция с Keycloak/Gatekeeper, проверка JWT, маппинг ролей (RBAC).                                           | Go, OIDC/JWT, Keycloak                              |
| **BoardService**         | CRUD досок/метаданных, политики доступа, сохранение JSON-снимков в S3, генерация presigned ссылок.              | Go, PostgreSQL, S3                                  |
| **CollabService**        | Реалтайм-редактирование, синхронизация операций (OT/CRDT), presence, курсоры.                                   | Go, WebSocket, Redis                                |
| **FileService**          | Загрузка/верификация файлов, presigned URL, миниатюры/превью, политики хранения/жизни объектов.                 | Go, S3, Image libs                                  |
| **PaymentService**       | Инициация платежей/рекуррентных списаний (исходящие в Tpay), обработка статусов через webhooks (через Gateway). | Go, Tpay API/Webhooks, PostgreSQL                   |
| **NotificationService**  | Консьюмер Kafka; шаблоны и отправка уведомлений (Push/Email/Chats), ретраи/идемпотентность.                     | Go, Kafka, FCM/APNs/SMTP                            |
| **AdminService**         | Тарифы/лимиты, управление пользователями/воркспейсами, справочники, фиче-флаги.                                 | Go, PostgreSQL                                      |
| **CronJobs**             | Фоновые задания: валидация тарифов, очистка старых снапшотов, переиндексации, ретраи интеграций.                | K8s CronJob, Go                                     |
| **Kafka / EventBus**     | Доменные события и интеграционные потоки; фан-аут на аналитику и нотификации (Carrot Quest, Mixpanel, Push).    | Kafka, Schema Registry                              |
| **Monitoring/Logging**   | Метрики/алерты, централизованные логи, трассировка.                                                             | Prometheus, Grafana, ELK/Loki, OpenTelemetry/Jaeger |


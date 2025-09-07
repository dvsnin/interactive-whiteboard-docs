# Структура системы

## Описание проекта

Интерактивная онлайн-доска для совместной работы в реальном времени (аналог Miro).  
Система построена на микросервисной архитектуре с поддержкой горизонтального масштабирования.  
Хранение версий и состояние досок реализовано через S3-хранилище в формате JSON-схем.
Бэкенд реализован на языке *Go (Golang)*

## Основные компоненты системы

| Компонент              | Описание                                                                                         |
|------------------------|--------------------------------------------------------------------------------------------------|
| Frontend               | Веб-клиент (React, WebSocket, REST)                                                              |
| AuthService            | Аутентификация и авторизация (Gatekeeper, Keycloak, JWT, OAuth2/OIDC, RBAC)                      |
| BoardService           | CRUD-досок, хранение метаданных в PostgreSQL                                                     |
| CollabService          | Совместное редактирование в реальном времени (WebSocket)                                         |
| FileService            | Загрузка файлов, хранение медиа в S3, генерация presigned URL                                    |
| PaymentService         | Интеграция с платёжными провайдерами (T-pay) через webhooks                                      |
| NotificationService    | Email/Push/Chats уведомления                                                                     |
| AdminService           | Панель администратора, управление пользователями и тарифами                                      |
| Monitoring/Logging     | Метрики (Prometheus, Grafana), логи и аудит (ELK, Jaeger, DLP)                                   |
| Kafka / EventBus       | Очереди сообщений и событий, аналитика, интеграция с внешними сервисами (carrot quest, mixpanel) |
| CronJobs               | Фоновые задания: валидация тарифов, очистка устаревших данных, пересчёт аналитики, рассылки      |

> **Kafka используется как единая шина событий**: все микросервисы могут публиковать туда доменные события (например, действия пользователей, изменения досок, платежи), а отдельные консьюмеры разбирают их для аналитики, уведомлений и интеграций с внешними сервисами.

## Архитектура (логическая схема)

```mermaid
flowchart LR
    %% ===== Колонка 1: клиенты =====
    subgraph Col1[ ]
    direction TB
        ios[iOS App]
        android[Android App]
        web[Web App]
    end

    %% ===== Колонка 2: шлюз/авторизация =====
    subgraph Col2[ ]
    direction TB
        gw[API Gateway / BFF]
        auth[AuthService]
    end

    %% ===== Колонка 3: микросервисы (Go) =====
    subgraph Col3[ ]
    direction TB
        board[BoardService]
        collab[CollabService]
        file[FileService]
        pay[PaymentService]
        admin[AdminService]
        notify[NotificationService]
    end

    %% ===== Колонка 4: данные/события/фон =====
    subgraph Col4[ ]
    direction TB
        redis[(Redis)]
        kafka[(Kafka EventBus)]
        s3[(S3: JSON схемы досок + медиа)]
        pg[(PostgreSQL)]
        cron[CronJobs]
    end

    %% ===== Колонка 5: внешние потребители =====
    subgraph Col5[Внешние консьюмеры]
    direction TB
        push[Push Service]
        carrot[Carrot Quest]
        mix[Mixpanel]
        dash[Analytics / Dashboards]
        tpay[T-pay]
    end

    %% --- Потоки слева направо ---
    ios --> gw
    android --> gw
    web --> gw

    gw --> auth
    gw --> board
    gw --> collab
    gw --> file
    gw --> pay
    gw --> admin
    gw --> notify

    %% Микросервисы -> хранилища
    board --> pg
    pay --> pg
    admin --> pg
    notify --> pg
    collab --> redis
    file --> s3
    board --> s3

    %% Kafka как общая шина
    auth --> kafka
    board --> kafka
    collab --> kafka
    file --> kafka
    pay --> kafka
    admin --> kafka
    notify --> kafka

    %% NotificationService читает события из Kafka (outbox)
    kafka --> notify

    %% CronJobs работают с БД/S3/Kafka
    cron --> pg
    cron --> s3
    cron --> kafka

    %% Внешние интеграции
    kafka --> push
    kafka --> carrot
    kafka --> mix
    kafka --> dash
    pay --> tpay

    %% Немного стилей
    classDef wide fill:#fff,stroke:#bbb;
    class gw,auth,board,collab,file,pay,admin,notify,redis,kafka,s3,pg,cron,push,carrot,mix,dash,tpay wide;
```
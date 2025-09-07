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

## Архитектура (логическая схема, упрощённая)

```mermaid
flowchart LR
    %% Колонка 1: клиенты
    subgraph Clients[Клиенты]
        A1[Web App]
        A2[Android App]
        A3[iOS App]
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
    A2 --> GW
    A3 --> GW

    GW --> Services
    Services --> Data
    Services --> Kafka
    Kafka --> Services

    Jobs --> Data
    Jobs --> Kafka

    Kafka --> Consumers
    Payment --> Tpay
```
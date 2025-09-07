# Взаимосвязи системы с внешними смежными

## Обзор
Система интегрируется с внешними провайдерами аутентификации, платежей, уведомлений, аналитики, мониторинга и хранения. Все внешние обмены проходят через контролируемые интерфейсы (REST/webhook, Kafka-консьюмеры/продюсеры, SMTP/FCM) с аутентификацией и подписанием.

## Карта интеграций

| Группа            | Система/Провайдер                   | Назначение                                      |
|-------------------|-------------------------------------|-------------------------------------------------|
| Идентификация     | Keycloak (OIDC/SAML)                | Логин/SSO, проверка токенов, RBAC               |
| Платежи           | T-pay (webhooks)                    | Привязка карт, рекуррентные списания, статусы   |
| Хранилище         | S3-совместимое хранилище            | JSON-версии досок и медиа                       |
| Уведомления       | FCM/APNs/Email/Chats (Max/Telegram) | Push, email, чат-нотификации                    |
| Аналитика         | Carrot Quest, Mixpanel              | Поведение пользователей, продуктовые метрики    |
| Доставка статики  | CDN                                 | Раздача статики и пользовательских медиа        |
| Мониторинг        | Prometheus/Grafana, Sentry          | Метрики, алерты, ошибки                         |
| Логи/Аудит        | ELK/Loki, DLP                       | Централизованные логи, контроль утечек          |
| Event Bus         | Kafka                               | Доменные события и интеграционные потоки        |

> Примечание: схемы событий фиксируются в Schema Registry (JSON-Schema/Proto). Вебхуки валидируются по подписи/secret.

## Диаграмма взаимодействий (высокоуровнево)

```mermaid
flowchart LR
    %% Внутренняя система
    subgraph System[Онлайн-доска (внутренние компоненты)]
        direction TB
        Gateway[API Gateway/BFF]
        Services[Микросервисы на Go]
        Data[(PostgreSQL/Redis/S3)]
        Bus[(Kafka EventBus)]
    end

    %% Внешние смежные
    subgraph IdP[Идентификация]
        KC[Keycloak]
    end

    subgraph Pay[Платежи]
        Tpay[T-pay (webhooks)]
    end

    subgraph Notify[Уведомления]
        Push[FCM/APNs]
        Mail[SMTP/Email]
        Chats[Slack/Telegram]
    end

    subgraph Analytics[Аналитика]
        Carrot[Carrot Quest]
        Mix[Mixpanel]
    end

    subgraph Ops[Мониторинг/Логи/CDN]
        CDN[CDN]
        Mon[Prometheus/Grafana]
        Sentry[Sentry]
        Logs[ELK/Loki/DLP]
    end

    %% Потоки
    Gateway <--> KC
    Services --> Tpay
    Tpay --> Services
    Services --> Data
    Services <--> Bus

    %% Уведомления и аналитика из EventBus
    Bus --> Push
    Bus --> Mail
    Bus --> Chats
    Bus --> Carrot
    Bus --> Mix

    %% Экспорт статики/медиа
    Data --> CDN

    %% Мониторинг и ошибки
    Services --> Mon
    Services --> Sentry
    Services --> Logs

# Структура системы

## Описание проекта

Интерактивная онлайн-доска для совместной работы в реальном времени (аналог Miro).  
Система построена на микросервисной архитектуре с поддержкой горизонтального масштабирования.  
Хранение версий и состояние досок реализовано через S3-хранилище в формате JSON-схем.

## Основные компоненты системы

| Компонент              | Описание                                                                    |
|------------------------|-----------------------------------------------------------------------------|
| Frontend               | Веб-клиент (React, WebSocket, REST)                                         |
| AuthService            | Аутентификация и авторизация (Gatekeeper, Keycloak, JWT, OAuth2/OIDC, RBAC) |
| BoardService           | CRUD-досок, хранение метаданных в PostgreSQL                                |
| CollabService          | Совместное редактирование в реальном времени (WebSocket)                    |
| FileService            | Загрузка файлов, хранение медиа в S3, генерация presigned URL               |
| PaymentService         | Интеграция с платёжными провайдерами (T-pay) через webhooks                 |
| NotificationService    | Email/Push/Chats уведомления                                                |
| AdminService           | Панель администратора, управление пользователями и тарифами                 |
| Monitoring/Logging     | Метрики (Prometheus, Grafana), логи и аудит (ELK, Jaeger, DLP)              |

---

## Архитектура (логическая схема)

```mermaid
flowchart TB
    subgraph Client[Клиенты]
        A[Web/Mobile App]
    end

    subgraph Gateway[API Gateway / BFF]
        G1[REST/gRPC API]
    end

    subgraph Services[Микросервисы]
        S1[AuthService]
        S2[BoardService]
        S3[CollabService]
        S4[FileService]
        S5[PaymentService]
        S6[NotificationService]
        S7[AdminService]
    end

    subgraph Data[Хранилища]
        D1[(PostgreSQL)]
        D2[(Redis)]
        D3[(S3 Object Storage: JSON схемы досок + медиа)]
    end

    A --> G1
    G1 --> S1
    G1 --> S2
    G1 --> S3
    G1 --> S4
    G1 --> S5
    G1 --> S6
    G1 --> S7

    S2 --> D1
    S3 --> D2
    S2 --> D3
    S4 --> D3
    S5 --> D1
    S6 --> D1
    S7 --> D1
```

# Валидатор контрактов AsyncAPI

Инструмент валидации контраков для AsyncAPI 3.0 спецификаций потребителя и поставщика.
Инструмент отлично вписывается в экосистему, где все системы описываются в AsyncAPI спецификациях.
Валидация контрактов позволяет обеспечить совместимость между системами, что является важным свойством для интеграции и совместной работы программных систем.

## Архитектура

### Модули

#### 1. Парсер (`parser/`)
Парсер AsyncAPI 3.0 спецификаций с полной поддержкой всех элементов:
- Парсинг из файлов, URL, строк
- Резолвинг ссылок (`$ref`)
- Валидация версий спецификаций

#### 2. Валидатор (`validator/`)
Основной пакет валидации, состоящий из:

**types.go** - Структуры данных:
- Конфигурация (`Config`, `ContractTests`)
- Результаты валидации (`ValidationResult`, `ChannelValidationResult`)
- Информация о каналах и сообщениях (`ChannelInfo`, `MessageInfo`)

**contract_validator.go** - Главный модуль валидации (ROP паттерн):
- Загрузка конфигурации из `contract-tests.yaml`
- Парсинг спецификаций потребителя/поставщика
- Координация процесса валидации

**channels_validator.go** - Валидация совместимости каналов:
- Сопоставление каналов по протоколам
- Проверка совместимости структур сообщений
- Обработка request-reply паттернов

**parser.go** - Обёртки для работы с AsyncAPI:
- Извлечение каналов и сообщений
- Экранирование имён каналов
- Получение протоколов через серверы

### Принцип работы

```
contract-tests.yaml
        ↓
[Валидатор контрактов]
        ↓
    Загрузка конфигурации
        ↓
    Парсинг спецификаций (Парсер)
        ↓
    Формирование ContractValidate
        ↓
[Валидатор каналов]
        ↓
    Извлечение канала потребителя
        ↓
    Поиск совместимого канала поставщика
        ↓
    Валидация структур сообщений
        ↓
[Результат валидации]
```

### Алгоритм валидации каналов

1. **Извлечение информации о канале потребителя:**
   - Получение протокола через связанные серверы
   - Извлечение исходящих (send) и входящих (reply) сообщений

2. **Поиск совместимого канала поставщика:**
   - Фильтрация по протоколу
   - Проверка совместимости структур сообщений
   - Валидация по обязательным полям

3. **Результат:**
   - Детальная информация о совместимых каналах
   - Структуры сообщений с типами данных

## Использование

### Конфигурация

Создайте файл `contract-tests.yaml`:

```yaml
consumer:
  name: "сервис-потребитель"
  channel: "user/signedup"
  specPath: "./specs/consumer.yaml"

provider:
  name: "сервис-поставщик"
  specUrl: "https://example.com/provider-spec.yaml"
```

### Запуск валидации

```go
validator := NewContractValidator()
result, err := validator.Validate("./contract-tests.yaml")
if err != nil {
    log.Fatal(err)
}
```

## Особенности

- Поддержка AsyncAPI 3.0
- Автоматическое экранирование путей каналов
- Валидация по протоколам и структурам сообщений
- ROP (Railway Oriented Programming) для обработки ошибок
- Поддержка загрузки спецификаций из файлов и URL

## Основные паттерны коммуникации между сервисами

В современной микросервисной архитектуре и Event-Driven Architecture (EDA) используются различные паттерны для организации взаимодействия между сервисами. AsyncAPI 3.0 поддерживает описание всех этих паттернов через свою спецификацию.

### Базовые паттерны обмена сообщениями

#### 1. Request-Reply (Запрос-Ответ)
Синхронный паттерн, где отправитель ждёт ответа от получателя.
- **Применение**: REST API вызовы, RPC, синхронные операции
- **AsyncAPI**: Используется `operation` с `reply` для описания
- **Пример**: Запрос данных пользователя, валидация транзакции
- **Документация**: [AsyncAPI Request-Reply](https://www.asyncapi.com/docs/tutorials/getting-started/request-reply)
- **Спецификация**: [Operation Reply Object](https://www.asyncapi.com/docs/reference/specification/v3.0.0#operationReplyObject)

#### 2. Fire-and-Forget (Отправить-и-Забыть)
Асинхронный паттерн без ожидания подтверждения.
- **Применение**: Логирование, аудит, уведомления
- **AsyncAPI**: `operation` типа `send` без `reply`
- **Пример**: Отправка событий аналитики, email-уведомления
- **Документация**: [AsyncAPI Operations](https://www.asyncapi.com/docs/concepts/asyncapi-document/operations)

####  
Паттерн для развязки производителей и потребителей через брокер сообщений.
- **Применение**: Event-driven системы, распределение событий
- **AsyncAPI**: Каналы с множественными подписчиками
- **Протоколы**: Kafka, RabbitMQ, MQTT, AMQP
- **Пример**: Событие "Заказ создан" обрабатывается складом, биллингом и доставкой
- **Документация**: [Event-Driven Architectures](https://www.asyncapi.com/docs/tutorials/getting-started/event-driven-architectures)
- **Примеры**: [Kafka Example](https://github.com/asyncapi/asyncapi/tree/master/examples/simple-asyncapi.yml)

### Паттерны управления состоянием

#### 4. Event Sourcing (Событийное хранилище)
Сохранение всех изменений состояния как последовательности событий.
- **Применение**: Аудит, восстановление состояния, временные запросы
- **AsyncAPI**: Описание событий изменения состояния
- **Преимущества**: Полная история изменений, возможность replay
- **Пример**: История изменений банковского счёта
- **Паттерн**: [Event Sourcing Pattern](https://microservices.io/patterns/data/event-sourcing.html)
- **С AsyncAPI**: [Event-Driven AsyncAPI](https://www.asyncapi.com/blog/event-driven-api-asyncapi)

#### 5. CQRS (Command Query Responsibility Segregation)
Разделение операций чтения и записи.
- **Применение**: Высоконагруженные системы, сложные домены
- **AsyncAPI**: Отдельные каналы для команд и событий
- **Связь с Event Sourcing**: Часто используются вместе
- **Пример**: Отдельные модели для записи заказов и их отображения
- **Паттерн**: [CQRS Pattern](https://microservices.io/patterns/data/cqrs.html)
- **Руководство**: [CQRS with Event Sourcing](https://learn.microsoft.com/en-us/azure/architecture/patterns/cqrs)

### Паттерны координации

#### 6. Saga Pattern
Управление распределёнными транзакциями через последовательность локальных транзакций.
- **Реализации**:
  - **Choreography**: Сервисы координируются через события
  - **Orchestration**: Центральный координатор управляет процессом
- **AsyncAPI**: Описание цепочек событий и компенсирующих действий
- **Пример**: Процесс оформления заказа с резервированием товара, оплатой и доставкой
- **Паттерн**: [Saga Pattern](https://microservices.io/patterns/data/saga.html)
- **Пример AsyncAPI**: [Saga Example](https://github.com/asyncapi/asyncapi/blob/master/examples/mercure-asyncapi.yml)

#### 7. Event Streaming
Непрерывная передача потока событий.
- **Применение**: Real-time аналитика, IoT, мониторинг
- **AsyncAPI**: Каналы с потоковой передачей данных
- **Протоколы**: Kafka Streams, Apache Pulsar
- **Пример**: Поток биржевых котировок, метрики IoT устройств
- **Документация**: [Streaming with AsyncAPI](https://www.asyncapi.com/blog/websocket-streaming-asyncapi)
- **Kafka Streams**: [AsyncAPI Kafka Bindings](https://github.com/asyncapi/bindings/tree/master/kafka)

### Дополнительные паттерны

#### 8. Message Routing
Маршрутизация сообщений на основе содержимого или метаданных.
- **Типы**: Content-based routing, Message filtering
- **AsyncAPI**: Использование `bindings` и `headers`
- **Документация**: [AsyncAPI Bindings](https://www.asyncapi.com/docs/concepts/asyncapi-document/bindings)
- **Примеры**: [AMQP Routing](https://github.com/asyncapi/bindings/tree/master/amqp)

#### 9. Dead Letter Queue (DLQ)
Обработка сообщений, которые не могут быть доставлены.
- **AsyncAPI**: Отдельные каналы для недоставленных сообщений
- **Пример**: [RabbitMQ DLQ Example](https://github.com/asyncapi/asyncapi/blob/master/examples/rpc-server-asyncapi.yml)
- **Документация**: [Error Handling Patterns](https://www.asyncapi.com/blog/error-handling-patterns)

#### 10. Circuit Breaker
Защита от каскадных сбоев в распределённых системах.
- **Применение**: Устойчивость к отказам
- **Паттерн**: [Circuit Breaker Pattern](https://microservices.io/patterns/reliability/circuit-breaker.html)
- **С AsyncAPI**: Описание через расширения и метаданные

### Поддержка в AsyncAPI 3.0

AsyncAPI 3.0 предоставляет богатые возможности для описания:
- **Каналы** (`channels`): Адресуемые компоненты для организации сообщений
- **Операции** (`operations`): Действия отправки/получения с поддержкой `reply`
- **Сообщения** (`messages`): Структура данных, заголовки, корреляция
- **Серверы** (`servers`): Поддержка различных протоколов (AMQP, MQTT, Kafka, WebSocket)
- **Привязки** (`bindings`): Специфичные для протокола расширения

### Выбор паттерна

При выборе паттерна учитывайте:
- **Консистентность данных**: Eventual vs Strong consistency
- **Производительность**: Синхронные vs асинхронные операции
- **Масштабируемость**: Возможность горизонтального масштабирования
- **Сложность**: Простота реализации и поддержки
- **Отказоустойчивость**: Обработка сбоев и восстановление

Эти паттерны формируют основу для построения надёжных, масштабируемых микросервисных систем и служат картой для дальнейшего развития валидатора контрактов. 
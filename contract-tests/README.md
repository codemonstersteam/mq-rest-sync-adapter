# Утилита проверки контрактов AsyncAPI 3.0

Утилита для проверки совместимости контрактов между потребителем и провайдером на основе спецификаций AsyncAPI 3.0.

## Функциональность

- ✅ Проверка совместимости каналов по адресам
- ✅ Нормализация адресов с параметрами
- ✅ Поддержка локальных и удалённых спецификаций
- ✅ JSON отчёты о совместимости
- ✅ Настраиваемые уровни логирования
- ✅ Валидация конфигурации
- ✅ **Словарь протоколов AsyncAPI 3.0**
- ✅ **Валидация неизвестных протоколов**
- ✅ **Предупреждения о несовместимости протоколов**

## Быстрый старт

### Установка

```bash
# Клонируйте репозиторий и перейдите в папку
cd contract-tests

# Установите зависимости
go mod tidy
```

### Использование

```bash
# Запуск с конфигурацией по умолчанию
go run . 

# Запуск с указанием конфигурации
go run . custom-config.yaml

# Сборка утилиты
go build -o contract-checker .

# Запуск скомпилированной утилиты
./contract-checker contract-tests.yaml
```

## Конфигурация

Пример файла `contract-tests.yaml`:

```yaml
contract_tests:
  # Спецификация потребителя
  consumer:
    spec_path: "../api-specification/asyncapi.yml"
    name: "mq-rest-sync-adapter"
    channels:
      - restGetBalanceRequest
    
  # Спецификация провайдера
  provider:
    spec_url: "https://example.com/provider-spec.yml"
    name: "wallet-balance-service"
    
  # Настройки
  settings:
    log_level: "info"              # debug, info, warn, error
    save_json_report: true
    json_report_file: "compatibility_report.json"
    timeout: 30                    # секунды
    ignore_warnings: false
```

## Тестирование

```bash
# Запуск всех тестов
go test -v

# Запуск с покрытием
go test -cover

# Запуск конкретного теста
go test -v -run TestCheckChannels
```

### Тестирование валидации протоколов

```bash
# Все тесты связанные с протоколами
go test -v -run="Protocol"

# Тесты для неизвестных протоколов
go test -v -run="TestIsProtocolKnown"               # Базовая валидация
go test -v -run="TestConsumerUnknownProtocol"       # Неизвестный протокол у потребителя
go test -v -run="TestProviderUnknownProtocol"       # Неизвестный протокол у провайдера
go test -v -run="TestProtocolValidationIntegration" # Интеграционный тест всех случаев

# Все тесты валидации неизвестных протоколов
go test -v -run="UnknownProtocol"
```

## Примеры результатов

### Совместимые контракты
```bash
✅ Контракты совместимы
📋 Потребитель: mq-rest-sync-adapter v1.0.0
🏢 Провайдер: wallet-balance-service v1.0.0
📄 JSON отчет сохранен: compatibility_report.json
```

### Несовместимые контракты
```bash
❌ Контракты НЕ совместимы
📋 Потребитель: test-consumer v1.0.0
🏢 Провайдер: test-provider v1.0.0
🚨 Критических изменений: 1
  • channel_address_not_supported: Channel 'restGetBalanceRequest' with address '/api/client/{Api-clientId}/wallet/balance' required by consumer but not supported by provider
```

### Неизвестные протоколы
```bash
❌ Контракты НЕ совместимы
📋 Потребитель: test-consumer-https v1.0.0
🏢 Провайдер: test-provider-http v1.0.0
🚨 Критических изменений: 2

🔍 Детали несовместимости:
  • unknown_protocol: Канал 'restGetBalanceRequest' использует неизвестный протокол 'https'. Поддерживаются только протоколы из официального словаря AsyncAPI 3.0
  • protocol_mismatch: Канал 'restGetBalanceRequest' использует протокол 'https', но провайдер поддерживает протокол 'http'. Протоколы несовместимы
```

## Структура проекта

```
contract-tests/
├── config.go              # Конфигурация
├── config_test.go         # Тесты конфигурации
├── checker.go             # Основная логика проверки
├── checker_test.go        # Тесты проверки
├── main.go                # Точка входа
├── protocols.yml          # Словарь протоколов AsyncAPI 3.0
├── contract-tests.yaml    # Основная конфигурация
├── testdata/              # Тестовые данные
│   ├── testdata-*.yaml    # Тестовые конфигурации
│   └── testdata_*.yml     # Тестовые спецификации
├── README.md              # Документация
├── TODO.md                # План развития
└── CLAUDE.md              # Описание состояния проекта
```

## Поддерживаемые форматы

- **AsyncAPI**: 3.0
- **Форматы файлов**: YAML
- **Транспорт**: HTTP/HTTPS для загрузки спецификаций
- **Типы каналов**: REST endpoints, MQ каналы

## Словарь протоколов AsyncAPI 3.0

Утилита поддерживает валидацию протоколов согласно официальному стандарту AsyncAPI 3.0. Словарь протоколов находится в файле `protocols.yml`.

### Поддерживаемые протоколы:

| **Категория** | **Протоколы** |
|---------------|---------------|
| **Web** | `http`, `ws`, `wss`, `mercure` |
| **Messaging** | `amqp`, `amqp1`, `jms`, `stomp`, `nats`, `solace`, `ibmmq` |
| **Streaming** | `kafka`, `kafka-secure`, `pulsar` |
| **IoT** | `mqtt`, `mqtt5` |
| **Cloud** | `sns`, `sqs`, `anypointmq`, `googlepubsub` |
| **Caching** | `redis` |

### Валидация протоколов:

- ✅ **Известные протоколы** - проходят валидацию
- ❌ **Неизвестные протоколы** - генерируют критические ошибки (breaking changes)
- 🔒 **Строгое соответствие** - протоколы должны точно совпадать
- 🛡️ **Предохранители** - проверка наличия серверов в каналах

### Тесты для проверки валидации протоколов:

| **Тест** | **Строка в коде** | **Описание** |
|----------|------------------|--------------|
| `TestIsProtocolKnown` | 605 | Проверяет функцию валидации известных/неизвестных протоколов |
| `TestConsumerUnknownProtocol` | 737 | Тестирует случай с неизвестным протоколом у потребителя (https) |
| `TestProviderUnknownProtocol` | 800 | Тестирует случай с неизвестным протоколом у провайдера (grpc) |
| `TestProtocolValidationIntegration` | 863 | Интеграционный тест всех случаев валидации |
| `TestUnknownProtocolBreakingChanges` | 673 | Проверяет генерацию критических ошибок для неизвестных протоколов |

### Тестовые данные для валидации протоколов:

```bash
# Потребитель с неизвестным протоколом
testdata/testdata_consumer_https.yml
testdata/testdata_config_consumer_https.yaml

# Провайдер с неизвестным протоколом  
testdata/testdata_provider_unknown_protocol_only.yml
testdata/testdata_config_provider_unknown.yaml

# Оба с неизвестными протоколами
testdata/testdata_consumer_unknown_protocol.yml
testdata/testdata_provider_unknown_protocol.yml
testdata/testdata_config_unknown_protocol.yaml
```

### JSON отчет с ошибками о неизвестных протоколах:

```json
{
  "isCompatible": false,
  "breakingChanges": [
    {
      "type": "unknown_protocol",
      "field": "channels.restGetBalanceRequest.protocol",
      "oldValue": "https",
      "newValue": "supported_protocol",
      "description": "Канал 'restGetBalanceRequest' использует неизвестный протокол 'https'. Поддерживаются только протоколы из официального словаря AsyncAPI 3.0",
      "severity": "high"
    },
    {
      "type": "protocol_mismatch",
      "field": "channels.restGetBalanceRequest.protocol",
      "oldValue": "https",
      "newValue": "http",
      "description": "Канал 'restGetBalanceRequest' использует протокол 'https', но провайдер поддерживает протокол 'http'. Протоколы несовместимы",
      "severity": "high"
    }
  ],
  "warnings": []
}
```

## Разработка

### Архитектура

Утилита состоит из следующих компонентов:

1. **Config** - парсинг и валидация конфигурации
2. **ContractChecker** - основная логика проверки
3. **AsyncAPISpec** - структуры для работы со спецификациями
4. **Reports** - генерация отчётов
5. **ProtocolRegistry** - словарь и валидация протоколов AsyncAPI 3.0
6. **LoadProtocols()** - загрузка протоколов из protocols.yml
7. **isProtocolKnown()** - проверка поддерживаемых протоколов

### Добавление новых проверок

Для добавления новой проверки:

1. Добавьте новый тип в `BreakingChange.Type`
2. Реализуйте логику в `ContractChecker`
3. Добавьте тесты в `checker_test.go`

## Лицензия

[Укажите лицензию]

## Вклад в проект

1. Форкните репозиторий
2. Создайте ветку для новой функциональности
3. Добавьте тесты
4. Создайте Pull Request

## Поддержка

- Создайте Issue для сообщения об ошибках
- Обратитесь к TODO.md для планируемых улучшений
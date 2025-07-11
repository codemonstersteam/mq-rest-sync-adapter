# Contract Compatibility Checker Tests

Этот проект содержит автомат для проверки совместимости контрактов между потребителем и поставщиком API, а также комплексные тесты.

## Структура файлов

### Основные файлы
- `contract_compatibility_checker.go` - Основной автомат для проверки совместимости
- `models.go` - Модели данных для результатов проверки
- `contract-tests.yaml` - Конфигурация для запуска проверок

### Тестовые файлы
- `contract_compatibility_checker_test.go` - Юнит тесты для основных функций
- `contract_compatibility_integration_test.go` - Интеграционные тесты с реальными спецификациями  
- `contract_test_checker.go` - Тестовая версия чекера для работы с локальными файлами
- `contract_test_utils.go` - Утилиты и хелперы для тестирования
- `contract_simple_test.go` - Простые тесты базовой функциональности

### Тестовые данные
- `testdata_consumer_asyncapi.yml` - Зафиксированная копия спецификации потребителя
- `testdata_provider_main.yml` - Спецификация провайдера из main ветки
- `testdata_provider_fails.yml` - Спецификация провайдера с ошибками
- `testdata/` - Директория с тестовыми спецификациями (создается автоматически)

## Запуск автомата

```bash
# Основная проверка с конфигурацией по умолчанию
go run contract_compatibility_checker.go models.go

# Проверка с кастомной конфигурацией  
go run contract_compatibility_checker.go models.go custom-config.yaml

# Прямая проверка двух спецификаций
go run contract_compatibility_checker.go models.go ../api-specification/asyncapi.yml https://example.com/provider-spec.yml
```

## Запуск тестов

```bash
# Все юнит тесты
go test -v contract_compatibility_checker_test.go contract_compatibility_checker.go contract_test_checker.go models.go

# Простые тесты функциональности
go test -v contract_simple_test.go contract_compatibility_checker.go contract_test_checker.go models.go

# Интеграционные тесты (требуют реальные спецификации)
go test -v contract_compatibility_integration_test.go contract_compatibility_checker.go contract_test_checker.go models.go

# Бенчмарки производительности
go test -bench=. contract_compatibility_checker_test.go contract_compatibility_checker.go contract_test_checker.go models.go
```

## Функциональность автомата

### 🔍 Анализ спецификаций
- Автоматический анализ потребителя и провайдера
- Разделение REST endpoints и MQ каналов
- Детальный план проверки с объяснениями

### 🌐 Проверка REST Endpoints
- Умное сопоставление путей с нормализацией параметров
- Сравнение структур запросов и ответов
- Валидация типов полей и форматов
- Проверка enum значений

### 📨 Проверка MQ Каналов  
- Проверка адресов очередей
- Сравнение структур сообщений
- Предупреждения для consumer-specific каналов

### 📋 Проверка схем данных
- Совместимость типов полей (integer ↔ number)
- Валидация обязательных полей
- Проверка критических изменений форматов
- Анализ enum значений

## Типы ошибок

### Breaking Changes (высокая критичность)
- `rest_endpoint_not_found` - REST endpoint не найден в провайдере
- `rest_message_type_missing` - Тип сообщения отсутствует
- `required_response_field_missing` - Обязательное поле ответа отсутствует
- `required_request_field_added` - Добавлено новое обязательное поле запроса
- `field_type_incompatible` - Несовместимые типы полей
- `request_enum_value_unsupported` - Неподдерживаемое enum значение в запросе
- `schema_removed` - Схема удалена из провайдера

### Warnings (предупреждения)
- `mq_channel_missing` - MQ канал не найден (может быть ожидаемо)
- `message_structure` - Требуется ручная проверка структуры сообщений
- `response_enum_value_added` - Новое enum значение в ответе

## Конфигурация

```yaml
contract_tests:
  consumer:
    spec_path: "../api-specification/asyncapi.yml"
    name: "mq-rest-sync-adapter"
    
  provider:
    spec_url: "https://git.codemonsters.team/guides/wallet-balance/-/raw/main/api-specification/asyncapi.yml"
    name: "wallet-balance-service"
    # Ограничения проверки каналов
    channel_restrictions:
      check_mode: "rest_only"  # all, rest_only, mq_only
    
  settings:
    log_level: "info"
    save_json_report: true
    json_report_file: "compatibility_report.json"
    timeout: 30
    ignore_warnings: false
```

## Ограничения проверки каналов

Для конкретных провайдеров можно ограничить типы каналов для проверки:

### Режимы проверки (`check_mode`)

- **`all`** (по умолчанию) - Проверяет все каналы (REST endpoints + MQ каналы)
- **`rest_only`** - Проверяет только HTTP взаимодействие (REST endpoints)
- **`mq_only`** - Проверяет только MQ каналы

### Пример для wallet-balance-service

```yaml
provider:
  name: "wallet-balance-service"
  channel_restrictions:
    check_mode: "rest_only"
```

**Результат:**
- ✅ REST endpoints проверяются полностью
- ⚠️ MQ каналы пропускаются с информационным сообщением
- 📊 Отсутствуют MQ-related предупреждения

### Вывод в логах

```
📝 ПЛАН ПРОВЕРКИ СОВМЕСТИМОСТИ:
  2. 📨 MQ Каналы:
     ⚠️  ПРОПУЩЕНО - провайдер 'wallet-balance-service' настроен только для HTTP (режим: rest_only)

📨 ПРОВЕРКА MQ КАНАЛОВ...
  ℹ️  Проверка MQ каналов пропущена согласно конфигурации (режим: rest_only)
```

## Структура тестов

### Юнит тесты
- Тестирование отдельных функций (нормализация путей, типы, enum)
- Проверка парсинга YAML
- Валидация экстракции имен схем и сообщений

### Интеграционные тесты  
- Тестирование с реальными спецификациями
- Проверка полного цикла совместимости
- Валидация структуры JSON отчета
- Тестирование производительности

### Тестовые утилиты
- `TestHelper` - хелперы для проверки результатов
- `SpecBuilder` - построитель тестовых спецификаций
- `MockChecker` - мок для изоляции тестов
- Готовые схемы для тестирования совместимости

## Лучшие практики

### При написании тестов
1. Используйте `TestHelper` для проверки результатов
2. Создавайте isolated тесты с `SpecBuilder`
3. Тестируйте как положительные, так и отрицательные сценарии
4. Проверяйте структуру JSON результатов

### При разработке чекера
1. Добавляйте детальные описания ошибок
2. Используйте правильные уровни severity
3. Предоставляйте полезные suggestions
4. Логируйте процесс проверки для отладки

## Примеры использования

### Простая проверка
```go
checker := NewContractCompatibilityChecker(
    "../api-specification/asyncapi.yml",
    "https://provider.com/spec.yml",
)
result, err := checker.CheckCompatibility()
```

### С тестовыми данными
```go
helper := NewTestHelper(t)
checker := NewTestContractCompatibilityChecker(
    "testdata_consumer_asyncapi.yml", 
    "testdata_provider_main.yml",
)
result, err := checker.CheckCompatibility()
helper.AssertContainsBreakingChange(result, "field_type_incompatible")
```

### Построение тестовых спецификаций
```go
spec := NewSpecBuilder().
    WithTitle("Test API").
    WithRESTChannel("test", "/api/test").
    WithMessage("testRequest", "#/components/schemas/testSchema").
    WithSchema("testSchema", GetTestRequestSchema()).
    Build()
```

## Exit коды

- `0` - Совместимость подтверждена или только warnings
- `1` - Найдены breaking changes или критические ошибки

## JSON отчет

Автомат создает детальный JSON отчет `compatibility_report.json` с:
- Полным списком breaking changes
- Предупреждениями и рекомендациями  
- Метаданными проверки (версии, время)
- Структурированными данными для CI/CD интеграции
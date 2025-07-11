# Contract Compatibility Checker

Утилита для проверки совместимости контрактов потребителя с поставщиком на основе спецификаций AsyncAPI.

## 🎯 Назначение

Утилита автоматически проверяет, совместимы ли изменения в спецификации поставщика с текущей спецификацией потребителя. Это предотвращает неожиданные поломки сервиса при изменении контрактов.

## 📁 Структура файлов

- `contract_compatibility_checker.go` - основная утилита проверки совместимости
- `contract_compatibility_test.go` - тесты для утилиты
- `models.go` - модели данных AsyncAPI
- `check_compatibility.sh` - скрипт для запуска проверки
- `Makefile` - команды автоматизации

## 🔍 Что проверяется

### Breaking Changes (критические изменения):
- ❌ Удаление обязательных полей
- ❌ Изменение типов полей  
- ❌ Удаление значений enum
- ❌ Удаление каналов/операций
- ❌ Изменение адресов каналов
- ❌ Удаление сообщений

### Warnings (предупреждения):
- ⚠️ Добавление новых значений enum
- ⚠️ Удаление необязательных полей

## 🚀 Запуск

### 🚀 Простой способ (рекомендуемый):
```bash
cd contract-tests
make check-compatibility
```

### 📝 Альтернативные способы:

#### 1. Через скрипт:
```bash
./check_compatibility.sh
```

#### 2. Напрямую через Go с конфигурацией:
```bash
# Использует contract-tests.yaml по умолчанию
go run contract_compatibility_checker.go models.go

# Использует кастомный конфиг
go run contract_compatibility_checker.go models.go my-config.yaml
```

#### 3. Напрямую через Go с параметрами:
```bash
go run contract_compatibility_checker.go models.go ../api-specification/asyncapi.yml https://git.codemonsters.team/guides/wallet-balance/-/raw/main/api-specification/asyncapi.yml
```

### Тестирование утилиты:
```bash
make compatibility-tests
```

## 📊 Результат

После запуска вы получите:
- **Консольный отчет** с результатами проверки
- **JSON файл** `compatibility_report.json` с детальными данными
- **Exit code 0** если совместимо, **exit code 1** если есть breaking changes

## 🛠️ Доступные команды

```bash
make help  # Показать все доступные команды
```

- `setup` - Установка зависимостей
- `check-compatibility` - Проверка совместимости контрактов
- `compatibility-tests` - Запуск тестов утилиты
- `clean` - Очистка сгенерированных файлов
- `fmt` - Форматирование кода
- `lint` - Проверка кода
- `all` - Все проверки

## 🏗️ Использование в CI/CD

Утилита возвращает exit code 1 при обнаружении breaking changes, что позволяет интегрировать её в CI/CD пайплайны:

```bash
if make check-compatibility; then
    echo "✅ Контракты совместимы"
else
    echo "❌ Найдены breaking changes!"
    exit 1
fi
```

## 📋 Примеры отчетов

### Совместимые контракты:
```
✅ Result
The consumer contract is compatible with the provider specification.
```

### Несовместимые контракты:
```
🚨 Breaking Changes
- field_type_changed (high): Field 'id' type changed from 'string' to 'number'
- required_field_added (high): Field 'email' became required

❌ Result  
The consumer contract is NOT COMPATIBLE with the provider specification.
```
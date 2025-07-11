#!/bin/bash

# Скрипт для проверки совместимости контрактов
# Использование: ./check_compatibility.sh

set -e

CONFIG_FILE="contract-tests.yaml"
COMPATIBILITY_CHECKER="contract_compatibility_checker.go"

echo "🔍 Проверка совместимости контрактов..."
echo "⚙️  Конфигурация: $CONFIG_FILE"
echo

# Проверяем, что конфигурационный файл существует
if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ Конфигурационный файл не найден: $CONFIG_FILE"
    exit 1
fi

# Проверяем, что утилита существует
if [ ! -f "$COMPATIBILITY_CHECKER" ]; then
    echo "❌ Утилита проверки совместимости не найдена: $COMPATIBILITY_CHECKER"
    exit 1
fi

# Запускаем проверку совместимости
echo "🚀 Запуск проверки совместимости..."
if go run "$COMPATIBILITY_CHECKER" models.go; then
    echo
    echo "✅ Контракты совместимы!"
    echo "📊 Подробный отчет сохранен в: compatibility_report.json"
    exit 0
else
    echo
    echo "❌ Обнаружены несовместимые изменения!"
    echo "📊 Подробный отчет сохранен в: compatibility_report.json"
    echo
    echo "🔧 Возможные действия:"
    echo "  1. Обновить спецификацию потребителя"
    echo "  2. Согласовать изменения с поставщиком"
    echo "  3. Использовать версионирование API"
    exit 1
fi
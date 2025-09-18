#!/bin/bash

# Скрипт для простого запуска утилиты в Docker

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода сообщений
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Проверяем наличие Docker
if ! command -v docker &> /dev/null; then
    error "Docker не установлен. Установите Docker для запуска утилиты."
    exit 1
fi

# Создаем директорию для отчетов если её нет
if [ ! -d "reports" ]; then
    log "Создаем директорию для отчетов..."
    mkdir -p reports
fi

# Устанавливаем права доступа для reports
chmod 777 reports/

# Проверяем наличие файла спецификации
SPEC_FILE="../api-specification/asyncapi.yml"
if [ ! -f "$SPEC_FILE" ]; then
    error "Файл спецификации не найден: $SPEC_FILE"
    error "Убедитесь, что файл существует или измените конфигурацию"
    exit 1
fi

# Определяем конфигурационный файл
CONFIG_FILE=${1:-"docker-contract-tests.yaml"}

if [ ! -f "$CONFIG_FILE" ]; then
    error "Конфигурационный файл не найден: $CONFIG_FILE"
    exit 1
fi

log "Используется конфигурация: $CONFIG_FILE"

# Проверяем наличие образа, если нет - собираем
if ! docker image inspect contract-checker &> /dev/null; then
    log "Образ contract-checker не найден, собираем..."
    docker build -t contract-checker .
    success "Образ собран успешно"
fi

log "Запускаем проверку контрактов..."

# Запускаем контейнер
docker run --rm \
  -v "$(pwd)/$CONFIG_FILE:/app/contract-tests.yaml:ro" \
  -v "$(pwd)/../api-specification/asyncapi.yml:/app/specs/asyncapi.yml:ro" \
  -v "$(pwd)/reports:/app/reports" \
  contract-checker

# Проверяем результат
if [ $? -eq 0 ]; then
    success "Проверка завершена успешно"
    if [ -f "reports/compatibility_report.json" ]; then
        log "Отчет сохранен: reports/compatibility_report.json"
    fi
else
    error "Проверка завершилась с ошибками"
    exit 1
fi
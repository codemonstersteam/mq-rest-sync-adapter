.PHONY: clean setup compatibility-tests check-compatibility all

# Настройка зависимостей
setup:
	go mod tidy
	go mod download

# Очистка
clean:
	rm -f compatibility_report.json
	go clean -testcache

# Проверка форматирования
fmt:
	go fmt ./...

# Проверка кода
lint:
	go vet ./...

# Проверка совместимости контрактов
check-compatibility:
	@echo "🔍 Проверка совместимости контрактов..."
	./check_compatibility.sh

# Тесты утилиты совместимости
compatibility-tests: setup
	go test -v -run TestContractCompatibilityChecker ./contract_compatibility_test.go ./contract_compatibility_checker.go ./models.go

# Все проверки
all: fmt lint compatibility-tests

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  setup               - Установка зависимостей"
	@echo "  compatibility-tests - Тесты утилиты совместимости"
	@echo "  check-compatibility - Проверка совместимости с поставщиком"
	@echo "  clean               - Очистка сгенерированных файлов"
	@echo "  fmt                 - Форматирование кода"
	@echo "  lint                - Проверка кода"
	@echo "  all                 - Все проверки"
	@echo "  help                - Показать эту справку"
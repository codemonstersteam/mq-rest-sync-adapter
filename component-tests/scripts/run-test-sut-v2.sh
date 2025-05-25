#!/bin/bash
echo "🚀 Starting tests for sut with stubs in containers..."
docker compose --file docker-compose-sut.yml up -d
echo "⏳ Waiting for test-runner container to start..."
docker logs -f sut & docker wait sut
EXIT_CODE=$?

# Проверяем результат тестов
if [ -f "./build/reports/cucumber.json" ]; then
    if grep -q '"status":"failed"' ./build/reports/cucumber.json; then
        echo "❌ Tests failed! Found failed scenarios in cucumber.json"
        docker compose --file docker-compose-sut.yml down
        exit 1
    fi
else
    docker compose --file docker-compose-sut.yml down
    echo "⚠️ Warning: cucumber.json not found in ./build/reports/"
    exit 1
fi

docker compose --file docker-compose-sut.yml down
echo "✅ Test-runner container finished"

# Возвращаем код ошибки от контейнера
exit $EXIT_CODE

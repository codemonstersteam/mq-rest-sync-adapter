#!/bin/bash

# Запоминаем время начала выполнения в миллисекундах
START_TIME=$(date +%s%N | cut -b1-13)

echo "🚀 Starting tests for sut with stubs in containers..."
docker compose --file docker-compose-test-sut.yml up -d
echo "⏳ Waiting for test-runner container to start..."
docker logs -f test-runner & docker wait test-runner
EXIT_CODE=$?

# Проверяем результат тестов
if [ -f "./build/reports/cucumber.json" ]; then
    if grep -q '"status":"failed"' ./build/reports/cucumber.json; then
        echo "❌ Tests failed! Found failed scenarios in cucumber.json"
        docker logs sut 
        docker compose --file docker-compose-test-sut.yml down
        docker rmi component-tests-sut:latest
        exit 1
    fi
else
    docker compose --file docker-compose-test-sut.yml down
    docker rmi component-tests-sut:latest
    echo "⚠️ Warning: cucumber.json not found in ./build/reports/"
    exit 1
fi
# docker logs sut
# TODO: check logs for feature toggle was enabled/disabled
echo "Checking feature toggle was enabled in logs: for wallet-balance workflow..."
docker logs sut | grep -i "wallet-balance workflow is enabled"
echo "Checking feature toggle was disabled in logs: for Incoming transmission into PipelineServicePS legacy flow..."
docker logs sut | grep -i "Incoming transmission into PipelineServicePS legacy flow"

docker compose --file docker-compose-test-sut.yml down

echo "✅ Test-runner container finished"
docker rmi component-tests-sut:latest

# Вычисляем время выполнения в миллисекундах
END_TIME=$(date +%s%N | cut -b1-13)
ELAPSED_MS=$((END_TIME - START_TIME))

# Конвертируем в минуты, секунды и миллисекунды
MINUTES=$((ELAPSED_MS / 60000))
SECONDS=$(((ELAPSED_MS % 60000) / 1000))
MILLISECONDS=$((ELAPSED_MS % 1000))

echo "⏱️  Script execution time: ${MINUTES}m ${SECONDS}s ${MILLISECONDS}ms"

# Возвращаем код ошибки от контейнера
exit $EXIT_CODE

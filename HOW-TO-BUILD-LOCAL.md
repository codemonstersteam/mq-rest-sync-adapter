# Локальная сборка

Для локальное сборки необходимо установить и настроить docker и docker compose.

## Build in docker
В файле [docker-compose.yml](docker-compose.yml) настроены два сервиса:  
- builder - для сборки  
- app - для запуска приложения (на понимаю зачем тебе это, когда можно и нужно тестироровать и запускать приложение автоматизированно), 

При запуске приложения локально, приложение крашнется без найденных зависимостей,
зависимости необходимо запустить командой:  

```bash
docker compose --file component-tests/docker-compose-stubs-local.yml up -d
docker compose --profile run up
```

(как это работает в window и на mac я не проверял, пока еще)  

Для сборки необходимо установить переменные окружения USER_ID и GROUP_ID. Для безопасности рекомендуется использовать UID и GID пользователя, который будет запускать сборку. [Dockerfile.build](Dockerfile.build)описан с опорой на [лучшие практики](https://docs.docker.com/build/building/best-practices/#user) docker.  


```bash
export USER_ID=$(id -u)
export GROUP_ID=$(id -g)
docker compose --profile build up
```

## Build in local JDK 21
Необходимо установить JDK 21 и настроить переменные окружения JAVA_HOME и PATH.
поднять зависимости 
настроить переменные для подключения к ресту и брокеру в конфиг файле [application.yml](src/main/resources/application.yml)  

```bash
docker compose --file component-tests/docker-compose-stubs-local.yml up -d
```

Вуаля!

```bash
./gradlew build test
```

Можно запускать локально тесты, но лучше использовать docker compose для запуска тестов.

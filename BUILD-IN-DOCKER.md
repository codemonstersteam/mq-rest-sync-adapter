# Build in docker

```bash
export USER_ID=$(id -u)
export GROUP_ID=$(id -g)
docker compose --profile build up
docker rmi mq-rest-sync-adapter-builder:latest
```

## Build in local JDK 21

docker for test-container

```bash
./gradlew build test
```

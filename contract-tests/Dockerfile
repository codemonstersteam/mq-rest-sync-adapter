# Используем официальный Go образ для сборки
FROM golang:1.21-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем статический бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o contract-checker .

# Минимальный образ для запуска
FROM alpine:latest

# Добавляем CA сертификаты для HTTPS запросов
RUN apk --no-cache add ca-certificates

# Создаем рабочую директорию
WORKDIR /app

# Копируем бинарник из builder образа
COPY --from=builder /app/contract-checker .

# Копируем словарь протоколов
COPY --from=builder /app/protocols.yml .

# Создаем директорию для конфигураций
RUN mkdir -p /app/config

# Создаем пользователя для безопасности
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Меняем владельца файлов
RUN chown -R appuser:appgroup /app

# Переключаемся на непривилегированного пользователя
USER appuser

# Точка входа
ENTRYPOINT ["./contract-checker"]

# По умолчанию используем стандартную конфигурацию
CMD ["contract-tests.yaml"]
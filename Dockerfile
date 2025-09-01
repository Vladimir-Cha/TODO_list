# Используем официальный образ Go на основе Alpine
FROM golang:1.24.5-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем зависимости и скачиваем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем все файлы проекта
COPY . .

# Компилируем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /TODO_app ./cmd/app/

# Создаем финальный образ
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /

# Копируем скомпилированное приложение
COPY --from=builder /TODO_app /TODO_app

# Копируем файлы миграции
COPY migrations /migrations
COPY .env .

# Открываем порт
EXPOSE 8080

# Переменные окружения по умолчанию
ENV PORT=8080

# Запускаем приложение
ENTRYPOINT ["/TODO_app"]
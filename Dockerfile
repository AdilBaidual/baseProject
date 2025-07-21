FROM golang:1.22

# Установка зависимостей
RUN go version

# Установка GOPATH
ENV GOPATH=/

# Копирование всех файлов в рабочую директорию контейнера
COPY ./ ./

# Загрузка зависимостей
RUN go mod download

# Сборка Go-приложения
RUN go build -o app ./cmd/main.go

# Команда для запуска приложения
CMD ["./app"]

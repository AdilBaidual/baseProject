FROM golang:1.22

# Проверка версии Go
RUN go version

# Установка GOPATH
ENV GOPATH=/

# Копирование всех файлов в рабочую директорию контейнера
COPY ./ ./

# Загрузка зависимостей
RUN go mod download

# Сборка Go-приложения
RUN go build -o /app/api ./cmd/main.go

# Проверка прав доступа к исполняемому файлу
RUN chmod +x /app/api

# Команда для запуска приложения
CMD ["/app/api"]

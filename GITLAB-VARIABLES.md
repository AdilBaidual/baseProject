# 🔐 GitLab CI/CD Variables Configuration

Полный список переменных для настройки в **GitLab → Settings → CI/CD → Variables**

## 📋 Как добавить переменные в GitLab

1. Заходите в ваш проект GitLab
2. **Settings** → **CI/CD** → **Variables** 
3. Нажимаете **Add Variable**
4. Указываете **Key** и **Value** из списков ниже
5. ✅ **Protect variable** (для защищенных веток)
6. ✅ **Mask variable** (для секретных данных)

---

## 🧪 Development Environment Variables

### 🖥️ **Сервер Development**

| Variable | Value | Description | Example |
|----------|-------|-------------|---------|
| `DEV_SSH_HOST` | IP или домен dev сервера | Адрес сервера для деплоя | `dev.yourproject.com` |
| `DEV_SSH_USER` | Пользователь SSH | Пользователь для подключения | `deploy` |
| `DEV_SSH_PRIVATE_KEY` | SSH приватный ключ | Private key для доступа | `-----BEGIN OPENSSH PRIVATE KEY-----...` |
| `DEV_SSH_KNOWN_HOSTS` | SSH fingerprint | Fingerprint сервера | `dev.yourproject.com ssh-rsa AAAAB3NzaC1...` |
| `DEV_PROJECT_PATH` | Путь к проекту dev | Папка с проектом develop ветки | `/home/deploy/project-dev` |
| `DEV_API_PORT` | Порт API в dev | HTTP порт для health check | `11066` |

---

## 🚀 Production Environment Variables

### 🖥️ **Сервер Production**

| Variable | Value | Description | Example |
|----------|-------|-------------|---------|
| `PROD_SSH_HOST` | IP или домен prod сервера | Адрес продакшн сервера | `prod.yourproject.com` |
| `PROD_SSH_USER` | Пользователь SSH | Пользователь для подключения | `deploy` |
| `PROD_SSH_PRIVATE_KEY` | SSH приватный ключ | Private key для доступа | `-----BEGIN OPENSSH PRIVATE KEY-----...` |
| `PROD_SSH_KNOWN_HOSTS` | SSH fingerprint | Fingerprint сервера | `prod.yourproject.com ssh-rsa AAAAB3NzaC1...` |
| `PROD_PROJECT_PATH` | Путь к проекту prod | Папка с проектом main ветки | `/home/deploy/project-prod` |
| `PROD_API_PORT` | Порт API в prod | HTTP порт для health check | `8066` |

---

## 🗂️ Структура проекта на сервере

### **Development Сервер:**
```
/home/deploy/
├── project-dev/          # ← DEV_PROJECT_PATH
│   ├── .git/
│   ├── .env              # ← Ваш dev .env файл
│   ├── docker-compose.dev.yml
│   ├── Dockerfile
│   └── ... (весь проект из develop ветки)
```

### **Production Сервер:**
```
/home/deploy/
├── project-prod/         # ← PROD_PROJECT_PATH  
│   ├── .git/
│   ├── .env              # ← Ваш prod .env файл
│   ├── docker-compose.prod.yml
│   ├── Dockerfile
│   └── ... (весь проект из main ветки)
```

---

## 📄 Примеры .env файлов для серверов

### **Development (.env для dev сервера):**
```bash
# API сервис порты
API_HTTP_PORT=11066
API_GRPC_PORT=50051

# API сервис настройки
GRPC_SERVER_HOST=0.0.0.0
GRPC_SERVER_PORT=50051
HTTP_SERVER_HOST=0.0.0.0
HTTP_SERVER_PORT=11066

# PostgreSQL настройки
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=myproject_dev
POSTGRES_USER=myproject_user
POSTGRES_PASSWORD=secure_dev_password_123
POSTGRES_EXTERNAL_PORT=5433
POSTGRES_SSLMODE=disable

# Jaeger настройки
JAEGER_AGENT_HOST=jaeger
JAEGER_UI_PORT=16686
JAEGER_AGENT_PORT=14268

# Общие настройки
ENV=development
```

### **Production (.env для prod сервера):**
```bash
# API сервис порты
API_HTTP_PORT=8066
API_GRPC_PORT=8051

# API сервис настройки
GRPC_SERVER_HOST=0.0.0.0
GRPC_SERVER_PORT=50051
HTTP_SERVER_HOST=0.0.0.0
HTTP_SERVER_PORT=11066

# PostgreSQL настройки
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=myproject_prod
POSTGRES_USER=myproject_user
POSTGRES_PASSWORD=very_secure_prod_password_456
POSTGRES_EXTERNAL_PORT=8432
POSTGRES_SSLMODE=disable

# Jaeger настройки
JAEGER_AGENT_HOST=jaeger
JAEGER_UI_PORT=8686
JAEGER_AGENT_PORT=8268

# Общие настройки
ENV=production
```

---

## 🔐 Как получить SSH ключи

### 1. **Генерация SSH ключей**

```bash
# Генерируем ключ для CI/CD
ssh-keygen -t rsa -b 4096 -C "gitlab-ci@yourproject.com" -f ~/.ssh/gitlab_ci

# Результат:
# ~/.ssh/gitlab_ci     (private key - для GitLab переменных)
# ~/.ssh/gitlab_ci.pub (public key - для серверов)
```

### 2. **Копирование public key на серверы**

```bash
# Development сервер
ssh-copy-id -i ~/.ssh/gitlab_ci.pub deploy@dev.yourproject.com

# Production сервер  
ssh-copy-id -i ~/.ssh/gitlab_ci.pub deploy@prod.yourproject.com
```

### 3. **Получение SSH fingerprints**

```bash
# Development
ssh-keyscan dev.yourproject.com

# Production
ssh-keyscan prod.yourproject.com
```

---

## 🖥️ Подготовка серверов

### **На каждом сервере выполните:**

```bash
# Создание пользователя deploy
sudo adduser deploy
sudo usermod -aG docker deploy

# Установка Docker и Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker deploy

sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Установка curl для health checks
sudo apt update && sudo apt install -y curl git

# Создание рабочих директорий и клонирование проекта
sudo -u deploy bash << 'EOF'
cd /home/deploy

# Development папка (клонируем develop ветку)
git clone -b develop YOUR_REPOSITORY_URL project-dev
cd project-dev
# Создайте здесь .env файл по примеру выше

# Production папка (клонируем main ветку)  
cd /home/deploy
git clone -b main YOUR_REPOSITORY_URL project-prod
cd project-prod
# Создайте здесь .env файл по примеру выше
EOF
```

**⚠️ Важно:** Замените `YOUR_REPOSITORY_URL` на реальный URL вашего GitLab репозитория

---

## ✅ Быстрый чеклист

### **GitLab переменные:**
- [ ] `DEV_SSH_HOST`
- [ ] `DEV_SSH_USER` 
- [ ] `DEV_SSH_PRIVATE_KEY`
- [ ] `DEV_SSH_KNOWN_HOSTS`
- [ ] `DEV_PROJECT_PATH`
- [ ] `DEV_API_PORT`
- [ ] `PROD_SSH_HOST`
- [ ] `PROD_SSH_USER`
- [ ] `PROD_SSH_PRIVATE_KEY` 
- [ ] `PROD_SSH_KNOWN_HOSTS`
- [ ] `PROD_PROJECT_PATH`
- [ ] `PROD_API_PORT`

### **Сервер настройка:**
- [ ] Docker установлен
- [ ] Docker Compose установлен
- [ ] Пользователь deploy создан
- [ ] SSH ключи настроены
- [ ] Проект клонирован в dev и prod папки
- [ ] .env файлы созданы в каждой папке

---

## 🚀 Workflow процесс

### **1. Разработка:**
```bash
# Создание feature ветки
git checkout -b feature/new-feature
git push origin feature/new-feature
# → Создает MR → Запускается только LINT
```

### **2. Development деплой:**
```bash
git checkout develop
git merge feature/new-feature
git push origin develop
# → Автоматически: ssh → cd dev → git pull → docker-compose up --build
```

### **3. Production деплой:**
```bash
git checkout main
git merge develop  
git push origin main
# → Автоматически: ssh → cd prod → git pull → docker-compose up --build
```

---

## 🧪 Локальная разработка

Для локального запуска используйте:

```bash
# Используйте env.local файл
cp env.local .env

# Запуск проекта
docker-compose -f docker-compose.local.yml --env-file .env up --build
```

---

## 🆘 Troubleshooting

### **SSH ошибки:**
- Проверьте что public key добавлен на сервер
- Проверьте права на приватный ключ: `chmod 600 ~/.ssh/gitlab_ci`
- Проверьте SSH fingerprint: `ssh-keyscan yourserver.com`

### **Git ошибки на сервере:**
```bash
# Проверьте что репозиторий клонирован
ssh deploy@yourserver.com "ls -la /home/deploy/"

# Проверьте что ветки настроены правильно
ssh deploy@yourserver.com "cd /home/deploy/project-dev && git branch -a"
```

### **Docker ошибки:**
- Проверьте что пользователь добавлен в группу docker
- Проверьте .env файлы на серверах
- Проверьте логи: `docker-compose logs`

### **Pipeline ошибки:**
- Проверьте логи в GitLab CI/CD → Pipelines
- Проверьте что все переменные добавлены и masked
- Проверьте SSH доступ с локальной машины

---

🎉 **Готово! Теперь CI/CD будет работать с отдельными dev и prod папками** 
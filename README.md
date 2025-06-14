# AuthProxy - Аутентификационный прокси с поддержкой OAuth2

## 1. Краткое описание функций

AuthProxy - это высокоэффективный аутентификационный прокси-сервер, обеспечивающий:

- **Защиту веб-приложений** через OAuth2-аутентификацию
- **Поддержку нескольких профилей** одновременно на разных портах
- **Два режима работы**:
  - Прокси-режим: перенаправление запросов к бэкенду с добавлением JWT
  - Статический режим: обслуживание HTML/CSS/JS файлов с контролем доступа
- **Управление сессиями** с использованием защищенных HTTP-кук
- **Гибкую конфигурацию** через YAML-файл и JSON-профили
- **Поддержку провайдеров OAuth**: Google, Yandex, любые совместимые с OAuth2

## 2. Сборка приложения

### Требования
- Go 1.16 или новее
- Git (для управления зависимостями)

### Инструкция по сборке
```bash
# Клонирование репозитория (если необходимо)
git clone https://github.com/your-repo/authproxy.git
cd authproxy

# Установка зависимостей
go get golang.org/x/oauth2
go get gopkg.in/yaml.v3

# Сборка бинарного файла
CGO_ENABLED=0 go build -o authproxy -ldflags="-s -w" *.go

# Проверка сборки
./authproxy --help
```

## 3. Конфигурационный файл

Основной конфигурационный файл в формате YAML (authproxy.yaml):

```yaml
profiles:
  - name: "backend_proxy"          # Идентификатор профиля
    port: "8080"                   # Порт для этого профиля
    public_url: "http://localhost:8080"  # Публичный URL
    oauth_config: "google.json"    # Файл конфигурации OAuth
    destination: "http://backend-app:3000"  # Целевой сервер (для прокси-режима)
    welcome_page: "/path/to/welcome.html"  # Страница приветствия (опционально)

  - name: "static_site"
    port: "8081"
    public_url: "http://localhost:8081"
    oauth_config: "yandex.json"
    static_dir: "/srv/static-files"  # Директория со статикой
    welcome_page: "/srv/welcome.html"
```

| Настройка    | Прокси-режим  | Статический режим | Описание                                         |
|--------------|---------------|-------------------|--------------------------------------------------|
| destination  | Обязательно   | -                 | URL целевого приложения                          |
| static_dir   | -             | Обязательно       | Путь к статическим файлам                        |
| welcome_page | Рекомендуется | Рекомендуется     | Страница для неаутентифицированных пользователей |
| oauth_config | Обязательно   | Обязательно       | JSON-файл с OAuth-конфигурацией                  |
| public_url   | Обязательно   | Обязательно       | Базовый URL для callback-адресов                 |

## 4. Настройка OAuth-клиента через Google Cloud Console

Пошаговая инструкция:

  1. Перейдите в Google Cloud Console
  2. Создайте новый проект или выберите существующий
  3. Перейдите в раздел "APIs & Services" > "Credentials"
  4. Нажмите "Create Credentials" > "OAuth client ID"
  5. Выберите тип приложения "Web application"
  6. Заполните поля:
  7. Name: Произвольное имя клиента
  8. Authorized JavaScript origins: https://ваш-домен
  9. Authorized redirect URIs: https://ваш-домен/callback
  10. Нажмите "Create"
  11. Скачайте JSON-конфигурацию (кнопка "Download JSON")

Пример JSON-файла профиля:
```json
{
  "web": {
    "client_id": "930311019656-xxxxxx.apps.googleusercontent.com",
    "project_id": "your-project-id",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_secret": "GOCSPX-xxxxxxxxxxxx",
    "redirect_uris": [
      "https://your-domain.com/callback"
    ],
    "javascript_origins": [
      "https://your-domain.com"
    ]
  }
}
```

## 5. Локальный запуск приложения

```bash
./authproxy --config custom_profile.yaml
```

Параметры командной строки:

| Параметр | По умолчанию   | Описание                 |
|----------|----------------|--------------------------|
| --config | profiles.yaml  | Путь к YAML-конфигурации |
| --help   | -              | Показать справку         |

Тестирование профилей:

```bash
# Должен произойти редирект на Welcome page
curl http://localhost:8080

# Должен произойти редирект на провайдера OAuth
curl http://localhost:8080/login
```

## 6. Развертывание как системного сервиса

### Подготовка окружения

```bash
sudo mkdir -p /opt/authproxy
sudo mkdir -p /etc/authproxy
sudo cp authproxy /opt/authproxy/
sudo cp authproxy.yaml /etc/authproxy/
sudo cp *.json /etc/authproxy/  # JSON-конфиги OAuth
```

### Создание systemd-сервиса

Файл: /etc/systemd/system/authproxy.service

```ini
[Unit]
Description=AuthProxy Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/authproxy
ExecStart=/opt/authproxy/authproxy --config /etc/authproxy/authproxy.yaml
Restart=on-failure
RestartSec=5
Environment="PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

# Настройки безопасности
CapabilityBoundingSet=
PrivateTmp=true
PrivateDevices=true
ProtectSystem=full
ProtectHome=true
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
```

### Запуск и управление сервисом

```bash
# Перезагрузка конфигурации systemd
sudo systemctl daemon-reload

# Включение автозапуска
sudo systemctl enable authproxy

# Запуск сервиса
sudo systemctl start authproxy

# Проверка статуса
sudo systemctl status authproxy

# Просмотр логов
journalctl -u authproxy -f
```

### Обновление сервиса

```bash
# Остановка сервиса
sudo systemctl stop authproxy

# Копирование новой версии
sudo cp authproxy /opt/authproxy/

# Запуск сервиса
sudo systemctl start authproxy
```

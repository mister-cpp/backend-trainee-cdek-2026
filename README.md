# Wishlist API

REST API сервис для создания вишлистов к праздникам.

## Стек
- Go 1.25
- PostgreSQL
- Docker Compose
- Роутер `chi`

## Полное задание в файле `Go.md`

## Запуск проекта

1. Необходмо скопировать конфиг: `cp .env.example .env` (или будут использованы дефолтные данные)
2. Запустите через Docker Compose или Makefile: 
   ```bash
   docker-compose up --build
   ```
   Или
3. ```bash
   make up
   ```
   
## Вебсайт доступен по ссылке

http://localhost:8080

## Примеры запросов

### 1. Регистрация
```bash
curl -X POST http://localhost:8080/api/register \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com", "password":"password123"}'
```

### 2. Вход (получение JWT)
```bash
curl -X POST http://localhost:8080/api/login \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com", "password":"password123"}'
```

### 3. Создание вишлиста
```bash
curl -X POST http://localhost:8080/api/wishlists \
     -H "Authorization: Bearer <JWT_TOKEN>" \
     -H "Content-Type: application/json" \
     -d '{"title":"День Рождения", "description":"Мой список желаний", "event_date":"2026-12-31T00:00:00Z"}'
```

### 4. Добавление подарка
```bash
curl -X POST http://localhost:8080/api/wishlists/1/items \
     -H "Authorization: Bearer <JWT_TOKEN>" \
     -H "Content-Type: application/json" \
     -d '{"title":"Гитара", "priority":10, "url":"https://example.com/guitar"}'
```

### 5. Публичный просмотр
```bash
curl -X GET http://localhost:8080/api/wishlists/public/<TOKEN>
```

### 6. Бронирование подарка (без авторизации)
```bash
curl -X POST http://localhost:8080/api/wishlists/public/<TOKEN>/items/1/reserve
```

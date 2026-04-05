[![Review Assignment Due Date](https://classroom.github.com/assets/deadline-readme-button-22041afd0340ce965d47ae6ef1cefeee28c7c493a6346c4f15d667ab976d596c.svg)](https://classroom.github.com/a/uvnTmvcw)



```bash
docker compose up --build
```

Сервис доступен на `http://localhost:8080`.

## Стек

Язык  Go 
HTTP-фреймворк Gin 
База данных PostgreSQL 17
Драйвер БД pgx v5
Авторизация JWT (HS256)
Деплой  Docker Compose

## Архитектура

Проект разделён на три слоя:

```
transport/       хендлеры, middleware, роутер
iternal/
  repository/
    service/     бизнес-логика
    postgres/    реализация репозиториев (SQL)
  models/        модели и ошибки
```


## Генерация слотов  «по запросу»


слоты создаются при первом запросе** `GET /rooms/{roomId}/slots/list?date=...`.

99.9% запросов приходятся на ближайшие 7 дней. Хранить слоты на месяцы вперёд нет смысла, они создаются тогда, когда нужны.
`INSERT ... ON CONFLICT (room_id, start_time) DO NOTHING` делает операцию идемпотентной

слоты генерируются только для дат, день недели входит в поле `daysOfWeek` расписания. 

## Схема БД

- `users` — пользователи
- `rooms` — переговорные комнаты
- `schedules` — расписание (дни недели, время начала/конца)
- `slots` — 30-минутные слоты (UUID стабильный по room_id + start_time)
- `bookings` — брони (partial unique index исключает двойное бронирование)

# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Сборка с метаданными (build info)

При старте **агент** (`cmd/agent`) и **сервер** (`cmd/server`) выводят в stdout:

```text
Build version: ...
Build date: ...
Build commit: ...
```

Значения задаются в пакете `internal/buildinfo`. По умолчанию для всех полей используется `N/A`. При компиляции их можно переопределить через `-ldflags` и флаг линковщика `-X`:

```bash
MODULE=github.com/MaxDrattcev/metrics_alerting_service
LDFLAGS="-X ${MODULE}/internal/buildinfo.BuildVersion=v1.2.3 \
  -X ${MODULE}/internal/buildinfo.BuildDate=2025-05-31 \
  -X ${MODULE}/internal/buildinfo.BuildCommit=abc1234"

go build -ldflags "${LDFLAGS}" -o agent ./cmd/agent
go build -ldflags "${LDFLAGS}" -o server ./cmd/server
```

Пример с подстановкой даты и коммита из git:

```bash
go build -ldflags "\
  -X ${MODULE}/internal/buildinfo.BuildVersion=v1.0.0 \
  -X ${MODULE}/internal/buildinfo.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%S) \
  -X ${MODULE}/internal/buildinfo.BuildCommit=$(git rev-parse --short HEAD 2>/dev/null || echo unknown)" \
  -o server ./cmd/server
```

Проверка:

```bash
./server
./agent
```

Без `-ldflags` в выводе останется `N/A` для всех полей.

## Benchmarks

Бенчмарки измеряют скорость важнейших компонентов сервиса сбора метрик: in-memory хранилище (`MemStorage`), слой `service` и обработку JSON в `handler` (путь `POST /updates`).

### Запуск

```bash
go test -bench=. -benchmem ./internal/repository/... ./internal/service/... ./internal/handler/...
```

### Окружение

| Параметр | Значение |
|----------|----------|
| ОС | darwin (macOS) |
| Архитектура | arm64 |
| CPU | Apple M1 Pro |
| Ветка | iter19 |
| Go | go1.24.3 darwin/arm64 |

Бенчмарки запускаются с `testing.B.Loop` (Go 1.24+).

### Сравнение до и после оптимизации (ключевые метрики)

| Бенчмарк | До (iter17) | После (iter19) |
|----------|-------------|----------------|
| `MemStorage_UpdateGauge` | 102.3 ns/op, 3 allocs/op | 46.2 ns/op, 1 alloc/op |
| `MemStorage_UpdateMetrics` | 3199 ns/op, 88 allocs/op | 1467 ns/op, 30 allocs/op |
| `MetricsService_UpdateMetrics` | 3608 ns/op, 89 allocs/op | 1617 ns/op, 31 allocs/op |

Основной выигрыш по alloc/op в batch update — за счёт `key()` и снижения давления на GC; gzip-оптимизация сильнее видна в pprof под HTTP-нагрузкой.

### Результаты (после оптимизации)

```
goos: darwin
goarch: arm64
pkg: github.com/MaxDrattcev/metrics_alerting_service/internal/repository
cpu: Apple M1 Pro
BenchmarkMemStorage_UpdateGauge-10              26854023                46.24 ns/op           16 B/op          1 allocs/op
BenchmarkMemStorage_UpdateCounter-10            20889394                57.22 ns/op           24 B/op          1 allocs/op
BenchmarkMemStorage_UpdateMetrics-10              823875              1467 ns/op            2584 B/op         30 allocs/op
BenchmarkMemStorage_GetAllMetrics-10             2352592               501.5 ns/op          2048 B/op          1 allocs/op
PASS
ok      github.com/MaxDrattcev/metrics_alerting_service/internal/repository     4.838s

goos: darwin
goarch: arm64
pkg: github.com/MaxDrattcev/metrics_alerting_service/internal/service
cpu: Apple M1 Pro
BenchmarkMetricsService_UpdateMetrics-10          732765              1617 ns/op            3064 B/op         31 allocs/op
BenchmarkMetricsService_UpdateGauge-10          20533221                57.65 ns/op           32 B/op          2 allocs/op
PASS
ok      github.com/MaxDrattcev/metrics_alerting_service/internal/service        2.380s

goos: darwin
goarch: arm64
pkg: github.com/MaxDrattcev/metrics_alerting_service/internal/handler
cpu: Apple M1 Pro
BenchmarkUnmarshalUpdatesBody-10          181932              6549 ns/op            2648 B/op         44 allocs/op
BenchmarkMarshalUpdatesResponse-10      13969887                96.45 ns/op           56 B/op          2 allocs/op
PASS
ok      github.com/MaxDrattcev/metrics_alerting_service/internal/handler        2.559s
```

### Что измеряется

| Бенчмарк | Пакет | Описание |
|----------|-------|----------|
| `BenchmarkMemStorage_UpdateMetrics` | `internal/repository` | Пакетное обновление ~29 метрик |
| `BenchmarkMemStorage_UpdateGauge` | `internal/repository` | Одна gauge-метрика |
| `BenchmarkMemStorage_UpdateCounter` | `internal/repository` | Одна counter-метрика |
| `BenchmarkMemStorage_GetAllMetrics` | `internal/repository` | Все метрики из памяти |
| `BenchmarkMetricsService_UpdateMetrics` | `internal/service` | Пакетное обновление через сервис |
| `BenchmarkMetricsService_UpdateGauge` | `internal/service` | Одна метрика через сервис |
| `BenchmarkUnmarshalUpdatesBody` | `internal/handler` | JSON `POST /updates` |
| `BenchmarkMarshalUpdatesResponse` | `internal/handler` | Пустой JSON-ответ |

### Расшифровка

- **ns/op** — время на операцию
- **B/op** — байты памяти на операцию
- **allocs/op** — число аллокаций на операцию

Наибольшая нагрузка: batch update (~30–31 allocs/op) и JSON unmarshal (44 allocs/op).

## Memory optimization (pprof)

Анализ и оптимизация потребления памяти под нагрузкой на `POST /updates`.

### Подготовка

```bash
mkdir -p profiles
```

Сервер запускается с pprof на порту `:6060` (см. `cmd/server/main.go`).

**1. Профиль до оптимизации (`base.pprof`):** запустить сервер **без** правок → нагрузка → снять heap:

```bash
hey -n 10000 -c 30 -m POST \
  -H "Content-Type: application/json" \
  -D payload.json \
  http://localhost:8080/updates

curl -o profiles/base.pprof http://localhost:6060/debug/pprof/heap
```

**2. Профиль после оптимизации (`result.pprof`):** внести изменения в код → **перезапустить** сервер → повторить `hey` → снять heap:

```bash
hey -n 10000 -c 30 -m POST \
  -H "Content-Type: application/json" \
  -D payload.json \
  http://localhost:8080/updates

curl -o profiles/result.pprof http://localhost:6060/debug/pprof/heap
```

### Анализ (до оптимизации)

```bash
go tool pprof profiles/base.pprof
go tool pprof -alloc_space profiles/base.pprof
```

Основные аллокации (`-alloc_space`):

- ~98% cumulative: `compress/flate.NewWriter` в middleware `Compress` — новый `gzip.NewWriter` на каждый запрос с заголовком `Accept-Encoding: gzip`;
- ~0.14%: `fmt.Sprintf` в `MemStorage.key`.

Команды в интерактивном режиме pprof: `top`, `top10 -cum`, `list key`, `list UpdateMetrics`, `web`.

### Изменения в коде

1. **`internal/middleware/compress.go`** — `sync.Pool` для `gzip.Writer` (переиспользование вместо создания writer на каждый запрос).
2. **`internal/repository/mem_storage.go`** — в `key()` конкатенация строк вместо `fmt.Sprintf`.

### Сравнение профилей

```bash
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
go tool pprof -top -diff_base=profiles/base.pprof -alloc_space profiles/result.pprof
```

### Результат diff (alloc_space)

Отрицательные значения означают уменьшение объёма аллокаций после оптимизации.

```
Showing nodes accounting for -7578.48MB, 96.95% of 7817.07MB total
Dropped 124 nodes (cum <= 39.09MB)
      flat  flat%   sum%        cum   cum%
-6271.39MB 80.23% 80.23% -7610.04MB 97.35%  compress/flate.NewWriter (inline)
-1308.46MB 16.74% 96.97% -1308.46MB 16.74%  compress/flate.(*compressor).initDeflate (inline)
    2.50MB 0.032% 96.93% -7652.53MB 97.90%  github.com/MaxDrattcev/metrics_alerting_service/internal.SetupRouter.Logger.func1
   -0.64MB 0.0081% 96.94% -1338.65MB 17.12%  compress/flate.(*compressor).init
   -0.50MB 0.0064% 96.95% -7660.59MB 98.00%  net/http.(*conn).serve
         0     0% 96.95% -7607.49MB 97.32%  compress/gzip.(*Writer).Close
         0     0% 96.95% -7609.54MB 97.35%  compress/gzip.(*Writer).Write
         0     0% 96.95% -7644.52MB 97.79%  github.com/MaxDrattcev/metrics_alerting_service/internal.SetupRouter.Compress.func2
```

### Анализ результатов оптимизации

#### Что показал pprof (до оптимизации)

Под нагрузкой на `POST /updates` (пакет из ~29 метрик, JSON, gzip в ответе) основной объём аллокаций (`alloc_space`) приходился не на бизнес-логику хранилища, а на инфраструктуру HTTP:

1. **Middleware `Compress`** — на каждый запрос с `Accept-Encoding: gzip` создавался новый `gzip.Writer` (`compress/flate.NewWriter`). В профиле это давало **~98% cumulative alloc_space**. При высокой частоте запросов это главный источник давления на GC.
2. **`MemStorage.key()`** — ключ строился через `fmt.Sprintf`, что давало лишние аллокации на каждое обращение к map (доля меньше, но заметна в `top`).

Бенчмарки до оптимизации подтверждали, что самые «тяжёлые» сценарии — пакетное обновление метрик и JSON, а не одиночный `UpdateGauge`.

#### Что сделали и как

| Место | Проблема | Решение |
|-------|----------|---------|
| `internal/middleware/compress.go` | Новый `gzip.Writer` на каждый ответ | **`sync.Pool`** на уровне пакета: writer берётся из пула, после `Close` возвращается в пул. `Reset(w)` перед использованием |
| `internal/repository/mem_storage.go` | `fmt.Sprintf` в `key()` | Конкатенация `mType + "/" + mName` без форматирования |

Идея: убрать повторяющиеся дорогие аллокации на горячем пути HTTP, не меняя внешнее API сервиса.

#### Что получилось

**По pprof (diff `base` → `result`, `-alloc_space`):**

- Суммарное снижение alloc_space на порядок **~7.6 GB** за тот же сценарий нагрузки (в diff: **~97%** относительно baseline).
- Почти полностью «обнулились» вкладки `compress/flate.NewWriter`, `initDeflate`, `gzip.Writer.Write/Close` — пул переиспользует deflate-состояние вместо создания с нуля.
- Цепочка `SetupRouter` → `Compress` перестала доминировать в профиле памяти.

**По бенчмаркам (см. таблицу в разделе Benchmarks):**

- `UpdateGauge`: ~102 ns/op → ~46 ns/op, 3 → 1 alloc/op.
- `UpdateMetrics` (repository): ~3200 ns/op → ~1470 ns/op, 88 → 30 allocs/op.
- `UpdateMetrics` (service): ~3600 ns/op → ~1620 ns/op, 89 → 31 allocs/op.

**По смыслу для сервиса:**

- Меньше работы GC под постоянной нагрузкой агента на `/updates`.
- Стабильнее latency ответов при gzip (меньше всплесков из-за аллокаций в middleware).
- Оптимизация `key()` — небольшой, но бесплатный выигрыш на каждом update/get в `MemStorage`.

**Что сознательно не трогали:**

- Логику JSON (`encoding/json`) — 44 allocs/op на unmarshal остаются ожидаемыми для `POST /updates`; это отдельный кандидат на оптимизацию (пулы буферов, другой encoder), но не входило в этот инкремент.
- Batch update в storage/service — основная стоимость там в копировании слайса метрик и работе с map, а не в gzip.

#### Вывод

Оптимизация была направлена на **главный узкий участок по памяти**, найденный через pprof: создание gzip-writer'ов в middleware. Переход на **`sync.Pool`** дал наибольший эффект; замена `fmt.Sprintf` в `key()` — дополнительное точечное улучшение. Подход: снять профиль под реальной нагрузкой → найти top по `alloc_space` → изменить минимальный участок кода → переснять профиль и сравнить через `diff_base`.

Файлы профилей (размер на диске): `profiles/base.pprof` — до оптимизации, `profiles/result.pprof` — после. В репозиторий не коммитить.

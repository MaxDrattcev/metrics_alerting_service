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
| Ветка | iter17 |
| Go | go1.24.3 darwin/arm64 |

### Результаты

```
goos: darwin
goarch: arm64
pkg: github.com/MaxDrattcev/metrics_alerting_service/internal/repository
cpu: Apple M1 Pro
BenchmarkMemStorage_UpdateGauge-10              10815037               102.3 ns/op            48 B/op          3 allocs/op
BenchmarkMemStorage_UpdateCounter-10             9927234               118.2 ns/op            56 B/op          3 allocs/op
BenchmarkMemStorage_UpdateMetrics-10              364395              3199 ns/op            3513 B/op         88 allocs/op
BenchmarkMemStorage_GetAllMetrics-10             2339740               495.6 ns/op          2048 B/op          1 allocs/op
PASS
ok      github.com/MaxDrattcev/metrics_alerting_service/internal/repository     6.277s

goos: darwin
goarch: arm64
pkg: github.com/MaxDrattcev/metrics_alerting_service/internal/service
cpu: Apple M1 Pro
BenchmarkMetricsService_UpdateMetrics-10          351152              3608 ns/op            3994 B/op         89 allocs/op
BenchmarkMetricsService_UpdateGauge-10          10244538               116.6 ns/op            64 B/op          4 allocs/op
PASS
ok      github.com/MaxDrattcev/metrics_alerting_service/internal/service        2.633s

goos: darwin
goarch: arm64
pkg: github.com/MaxDrattcev/metrics_alerting_service/internal/handler
cpu: Apple M1 Pro
BenchmarkUnmarshalUpdatesBody-10          180859              6478 ns/op            2648 B/op         44 allocs/op
BenchmarkMarshalUpdatesResponse-10      13988832                85.03 ns/op           56 B/op          2 allocs/op
PASS
ok      github.com/MaxDrattcev/metrics_alerting_service/internal/handler        2.535s
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

Наибольшая нагрузка: batch update (~88–89 allocs/op) и JSON unmarshal (44 allocs/op).

## Memory optimization (pprof)

Анализ и оптимизация потребления памяти под нагрузкой на `POST /updates`.

### Подготовка

```bash
mkdir -p profiles
```

Сервер запускается с pprof на порту `:6060` (см. `cmd/server/main.go`). Нагрузка:

```bash
hey -n 10000 -c 30 -m POST \
  -H "Content-Type: application/json" \
  -D payload.json \
  http://localhost:8080/updates
```

Профили heap снимаются сразу после нагрузки:

```bash
curl -o profiles/base.pprof http://localhost:6060/debug/pprof/heap
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

Файлы профилей: `base.pprof` (7.8K, до), `result.pprof` (4.6K, после).

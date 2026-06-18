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

## Бенчмарки

Бенчмарки покрывают ключевые компоненты системы:

| Пакет | Бенчмарки |
|-------|-----------|
| `internal/repository` | `WriteMetricsJSON`, `SaveToFile`, `LoadFromFile`, `SetGauge`, `ApplyBatch`, `GetAll` |
| `internal/agent` | `ReportMetrics_BuildBatch`, `SendBatchMetrics_Encode` |
| `internal/audit` | `Publisher_Notify`, `FileObserver_Update` |
| `internal/handler` | `CreateMetricsFromJSONBatch` |

Запуск:

```bash
go test -bench=. -benchmem ./internal/repository/... ./internal/agent/... ./internal/audit/... ./internal/handler/...
```

## Профилирование памяти (pprof)
### Анализ и оптимизация

**Проблема.** Профиль `base.pprof` показал, что основной потребитель памяти при нагрузке - `MemRepository.SaveToFile`. Метод вызывается при каждом обновлении метрик (если автосохранение отключено) и из `ApplyBatch` после пакетного POST `/updates/`. Старый подход собирал все gauge/counter в один слайс `[]storedMetric`, после чего вызывал `json.Marshal`. Это создавало промежуточный буфер размером с весь файл, а также аллокировал указатели для каждого значения (`*float64`, `*int64`).

**Что изменил.** Вместо `json.Marshal` всего слайса добавлена потоковая функция `writeMetricsJSON`, которая пишет JSON напрямую во временный файл через `json.Encoder`:

- массив собирается по одному элементу - без промежуточного слайса всех метрик
- для gauge и counter используются отдельные структуры `storedGaugeMetric` / `storedCounterMetric` с полями по значению, без лишних указателей
- сериализация идёт сразу в `io.Writer` (временный файл), а не в `[]byte` в памяти

**Результат.** Сравнение профилей (`-diff_base`) показывает отрицательные значения - память уменьшилась:

| Метрика | До | После | Изменение |
|---------|-----|-------|-----------|
| `alloc_space` (суммарно) | ~30.5 MB | ~19.0 MB | **−37.7%** |
| `SaveToFile` (cum, alloc_space) | ~20.7 MB | ~9 MB | **−~56%** |
| `SaveToFile` (flat, inuse_space) | - | -525 kB | меньше удерживаемой памяти |

Из топа diff исчезли `encoding/json.Marshal`, `bytes.growSlice` и `encoding/json.newEncodeState` как основные источники аллокаций в цепочке сохранения. Снизилось потребление и у HTTP-обработчика `CreateMetricsFromJSONBatch`, потому что он вызывает `ApplyBatch` -> `SaveToFile`.

Бенчмарки `BenchmarkWriteMetricsJSON` и `BenchmarkMemRepository_SaveToFile` позволяют отслеживать регрессии по времени и числу аллокаций (`-benchmem`).

### Сравнение профилей

```bash
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
```

```
File: metrics-server
Build ID: b79ef11502fabc6b84708b6a19040fdbfd052688
Type: inuse_space
Time: 2026-06-09 17:06:20 MSK
Showing nodes accounting for -2587.09kB, 50.15% of 5159.13kB total
Dropped 2 nodes (cum <= 25.79kB)
      flat  flat%   sum%        cum   cum%
 -525.43kB 10.18% 10.18%  -525.43kB 10.18%  github.com/Viva-Fidel/metrics-and-alerting/internal/repository.(*MemRepository).SaveToFile
 -525.43kB 10.18% 20.37%  -525.43kB 10.18%  github.com/go-playground/validator/v10.map.init.7
 -512.09kB  9.93% 30.29%  -512.09kB  9.93%  github.com/gabriel-vasile/mimetype/internal/magic.init
 -512.09kB  9.93% 40.22%  -512.09kB  9.93%  reflect.growslice
 -512.05kB  9.93% 50.15%  -512.05kB  9.93%  bufio.NewReaderSize (inline)
```

По `alloc_space` суммарные аллокации снизились с ~30.5 MB до ~19.0 MB (−37.7%), `SaveToFile` — с ~20.7 MB до ~9 MB cum.

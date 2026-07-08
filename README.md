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


## Результат оптимизации

```text
File: server
Build ID: 05c283beff63a04df19ed625233cf06207c4adf2
Type: alloc_space
Time: 2026-07-08 14:29:53 MSK
Showing nodes accounting for -898.15MB, 94.74% of 948.05MB total
Dropped 142 nodes (cum <= 4.74MB)
      flat  flat%   sum%        cum   cum%
 -470.68MB 49.65% 49.65%  -866.45MB 91.39%  compress/flate.NewWriter (inline)
 -253.01MB 26.69% 76.33%  -395.76MB 41.74%  compress/flate.(*compressor).init
 -137.25MB 14.48% 90.81%  -137.25MB 14.48%  compress/flate.newDeflateFast (inline)
  -22.18MB  2.34% 93.15%   -22.18MB  2.34%  compress/flate.(*dictDecoder).init (inline)
   -8.51MB   0.9% 94.05%    -8.51MB   0.9%  sync.(*Pool).pinSlow
   -4.02MB  0.42% 94.47%   -26.19MB  2.76%  compress/flate.NewReader
   -2.50MB  0.26% 94.74%    -5.50MB  0.58%  compress/flate.newHuffmanBitWriter (inline)
   -0.50MB 0.053% 94.79%   -27.71MB  2.92%  compress/gzip.NewReader (inline)
    0.50MB 0.053% 94.74%    -4.92MB  0.52%  github.com/scarypuppp/metrics-service/internal/middlewares.LogRequest.func1
         0     0% 94.74%    -8.55MB   0.9%  bufio.(*Writer).Flush
         0     0% 94.74%   -23.68MB  2.50%  compress/gzip.(*Reader).Reset
         0     0% 94.74%   -26.19MB  2.76%  compress/gzip.(*Reader).readHeader
         0     0% 94.74%  -862.04MB 90.93%  compress/gzip.(*Writer).Close
         0     0% 94.74%  -866.45MB 91.39%  compress/gzip.(*Writer).Write
         0     0% 94.74%    -6.92MB  0.73%  github.com/go-chi/chi/v5.(*Mux).Mount.func1
         0     0% 94.74%  -894.64MB 94.37%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 94.74%    -5.92MB  0.62%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 94.74%    -4.91MB  0.52%  github.com/go-chi/chi/v5/middleware.NoCache.func1
         0     0% 94.74%  -891.64MB 94.05%  github.com/scarypuppp/metrics-service/internal/middlewares.CompressResponse.func1
         0     0% 94.74%    16.75MB  1.77%  github.com/scarypuppp/metrics-service/internal/middlewares.CompressResponse.func1.1
         0     0% 94.74%   -29.10MB  3.07%  github.com/scarypuppp/metrics-service/internal/middlewares.DecompressRequest.func1
         0     0% 94.74%  -892.14MB 94.10%  github.com/scarypuppp/metrics-service/internal/middlewares.LogResponse.func1
         0     0% 94.74%    -4.98MB  0.53%  github.com/scarypuppp/metrics-service/internal/middlewares.gzipWriter.Write
         0     0% 94.74%    -5.53MB  0.58%  io.Copy (inline)
         0     0% 94.74%    -5.53MB  0.58%  io.CopyN
         0     0% 94.74%    -5.53MB  0.58%  io.copyBuffer
         0     0% 94.74%    -5.53MB  0.58%  io.discard.ReadFrom
         0     0% 94.74%    -7.03MB  0.74%  net/http.(*chunkWriter).Write
         0     0% 94.74%    -7.03MB  0.74%  net/http.(*chunkWriter).writeHeader
         0     0% 94.74%  -905.18MB 95.48%  net/http.(*conn).serve
         0     0% 94.74%    -8.03MB  0.85%  net/http.(*response).finishRequest
         0     0% 94.74%  -892.14MB 94.10%  net/http.HandlerFunc.ServeHTTP
         0     0% 94.74%  -894.64MB 94.37%  net/http.serverHandler.ServeHTTP
         0     0% 94.74%   -10.03MB  1.06%  sync.(*Pool).Get
         0     0% 94.74%    -8.51MB   0.9%  sync.(*Pool).pin
```

Были оптимизированы middleware'ы: compress и decompress, а также handler, отдающий список метрик в html.
Это дало результат:
Суммарные аллокации (alloc_space): 948Мб -> 39Мб
Количество аллокаци (alloc_objects): 233 815 -> 175 182 объектов (−25%)
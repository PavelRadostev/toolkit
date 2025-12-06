# Регистрация Query в Bus

Этот документ описывает, как зарегистрировать Query handler в системе Bus с использованием HandlerFactory.

## Обзор

Для регистрации Query необходимо:
1. Создать структуру handler'а с полями запроса и репозиторием
2. Реализовать функцию-конструктор для десериализации CBOR данных и инициализации репозитория
3. Реализовать метод `Handle` для обработки запроса
4. Зарегистрировать handler и репозиторий в `HandlerFactory`
5. Зарегистрировать stream в `Bus`

## Архитектура

Система использует паттерн Factory для управления созданием handlers:
- **HandlerFactory** — управляет регистрацией handlers и repositories
- **Bus** — регистрирует только stream names и использует фабрику для создания handlers

## Пример 1: AllGGISImportTemplatesQuery

### Шаг 1: Определение структуры Handler

```go
type AllGGISImportTemplatesQueryHandler struct {
    EnterpriseID EnterpriseId `cbor:"enterprise_id"`
    Repository   bus.Repository
}
```

Структура содержит:
- Поля запроса с CBOR тегами для корректной десериализации
- Поле `Repository` для доступа к данным

### Шаг 2: Функция-конструктор

```go
func NewAllGGISImportTemplatesQueryFromCBOR(data []byte, repo bus.Repository) (bus.Subscriber, error) {
    var handler AllGGISImportTemplatesQueryHandler
    if err := cbor.Unmarshal(data, &handler); err != nil {
        return nil, err
    }
    handler.Repository = repo
    return &handler, nil
}
```

Функция принимает:
- CBOR-закодированные данные (`[]byte`)
- Репозиторий (`bus.Repository`) — может быть `nil`, если не зарегистрирован

Десериализует данные в структуру handler'а, устанавливает репозиторий и возвращает `bus.Subscriber`.

### Шаг 3: Реализация метода Handle

```go
func (a *AllGGISImportTemplatesQueryHandler) Handle(ctx context.Context) (any, error) {
    fmt.Printf("HandlingAllGGISImportTemplatesQuery")
    
    // Использование репозитория для получения данных
    // if a.Repository != nil {
    //     // работа с репозиторием
    // }
    
    // Бизнес-логика обработки запроса
    return map[string]any{
        "id":         EnterpriseId(rand.Intn(1000) + 1),
        "name":       fmt.Sprintf("ProcessedTemplate_%d", rand.Intn(100)),
        "enterprise": EnterpriseId(rand.Intn(100) + 1),
        "filetype":   Filetype("CSV"),
        "delimiter":  ",",
        "status":     "success",
        "processed":  true,
        "records":    rand.Intn(10000),
    }, nil
}
```

Метод `Handle` реализует интерфейс `bus.Subscriber` и содержит логику обработки запроса. Возвращает результат и ошибку (если есть).

### Шаг 4: Регистрация в HandlerFactory и Bus

```go
factory := bus.NewHandlerFactory()

// Регистрация handler конструктора
factory.RegisterHandler("vist_domain.query.ggis_import.AllGGISImportTemplatesQuery", NewAllGGISImportTemplatesQueryFromCBOR)

// Регистрация репозитория (опционально)
// factory.RegisterRepository("vist_domain.query.ggis_import.AllGGISImportTemplatesQuery", someRepository)

busInstance := bus.NewBus(redisClient, ctx)
busInstance.SetFactory(factory)

// Регистрация stream в Bus
busInstance.Register("vist_domain.query.ggis_import.AllGGISImportTemplatesQuery")
```

## Полный пример использования

```go
func main() {
    ctx := context.Background()
    cfg := config.Load()
    
    redisClient := redis.NewClient(&redis.Options{
        Addr:     cfg.Redis.Addr,
        Password: cfg.Redis.Password,
        DB:       cfg.Redis.DB,
    })
    
    busInstance := bus.NewBus(redisClient, ctx)
    factory := bus.NewHandlerFactory()
    
    // Регистрация handlers в factory
    factory.RegisterHandler("vist_domain.query.ggis_import.AllGGISImportTemplatesQuery", NewAllGGISImportTemplatesQueryFromCBOR)
    factory.RegisterHandler("vist_domain.query.pit.plan.IsPlanApprovedQuery", NewIsPlanApprovedQueryFromCBOR)
    
    // Регистрация repositories в factory (опционально)
    // factory.RegisterRepository("vist_domain.query.ggis_import.AllGGISImportTemplatesQuery", someRepository)
    // factory.RegisterRepository("vist_domain.query.pit.plan.IsPlanApprovedQuery", anotherRepository)
    
    // Установка factory в Bus
    busInstance.SetFactory(factory)
    
    // Регистрация streams в Bus
    busInstance.Register("vist_domain.query.ggis_import.AllGGISImportTemplatesQuery")
    busInstance.Register("vist_domain.query.pit.plan.IsPlanApprovedQuery")
    
    // Запуск обработки сообщений
    busInstance.Run()
}
```

## Создание Repository

Для создания собственного репозитория необходимо реализовать интерфейс `bus.Repository`:

```go
type MyRepository struct {
    // поля репозитория
}

// MyRepository реализует bus.Repository
// (интерфейс пустой, можно добавить методы по необходимости)
```

Пример использования:

```go
myRepo := &MyRepository{/* инициализация */}
factory.RegisterRepository("vist_domain.query.ggis_import.AllGGISImportTemplatesQuery", myRepo)
```

## Важные моменты

1. **Интерфейс Subscriber**: Handler должен реализовывать интерфейс `bus.Subscriber` с методом `Handle(ctx context.Context) (any, error)`

2. **Интерфейс Repository**: Репозитории должны реализовывать интерфейс `bus.Repository` (базовый интерфейс, можно расширять)

3. **HandlerConstructor**: Конструктор handler'а должен иметь сигнатуру `func(data []byte, repo bus.Repository) (bus.Subscriber, error)`

4. **CBOR десериализация**: Данные приходят в формате CBOR, поэтому используйте `cbor.Unmarshal` для десериализации

5. **Именование stream'ов**: Используйте согласованное именование в формате `domain.query.module.QueryName`

6. **Обработка ошибок**: Функция-конструктор должна возвращать ошибку при неудачной десериализации

7. **Возвращаемые значения**: Метод `Handle` может возвращать любой тип данных (`any`), который будет сериализован в ответ

8. **Репозитории опциональны**: Если репозиторий не зарегистрирован для stream'а, в конструктор передается `nil`

9. **Порядок регистрации**: Сначала регистрируйте handlers и repositories в factory, затем устанавливайте factory в Bus, и только потом регистрируйте streams

## API HandlerFactory

- `NewHandlerFactory()` — создает новый экземпляр фабрики
- `RegisterHandler(streamName, constructor)` — регистрирует конструктор handler'а для stream'а
- `RegisterRepository(streamName, repo)` — регистрирует репозиторий для stream'а
- `CreateHandler(streamName, data)` — создает handler для указанного stream'а (используется Bus'ом)
- `HasHandler(streamName)` — проверяет, зарегистрирован ли handler для stream'а
- `GetStreams()` — возвращает список всех зарегистрированных stream'ов

## API Bus

- `NewBus(redis, ctx)` — создает новый экземпляр Bus
- `SetFactory(factory)` — устанавливает HandlerFactory для Bus
- `Register(streamName)` — регистрирует stream name в Bus (handler должен быть зарегистрирован в factory)
- `Run()` — запускает обработку сообщений для всех зарегистрированных streams

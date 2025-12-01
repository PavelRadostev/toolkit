# Регистрация Query в Bus

Этот документ описывает, как зарегистрировать Query handler в системе Bus.

## Обзор

Для регистрации Query необходимо:
1. Создать структуру handler'а с полями запроса
2. Реализовать функцию-конструктор для десериализации CBOR данных
3. Реализовать метод `Handle` для обработки запроса
4. Зарегистрировать handler через `bus.Register()`

## Пример 1: AllGGISImportTemplatesQuery

### Шаг 1: Определение структуры Handler

```go
type AllGGISImportTemplatesQueryHandler struct {
    EnterpriseID EnterpriseId `cbor:"enterprise_id"`
}
```

Структура содержит поля запроса с CBOR тегами для корректной десериализации.

### Шаг 2: Функция-конструктор

```go
func NewAllGGISImportTemplatesQueryFromCBOR(data []byte) (bus.Subscriber, error) {
    var handler AllGGISImportTemplatesQueryHandler
    if err := cbor.Unmarshal(data, &handler); err != nil {
        return nil, err
    }
    return &handler, nil
}
```

Функция принимает CBOR-закодированные данные (`[]byte`), десериализует их в структуру handler'а и возвращает `bus.Subscriber`.

### Шаг 3: Реализация метода Handle

```go
func (a *AllGGISImportTemplatesQueryHandler) Handle(ctx context.Context) (any, error) {
    fmt.Printf("HandlingAllGGISImportTemplatesQuery")
    
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

### Шаг 4: Регистрация в Bus

```go
bus.Register("vist_domain.query.ggis_import.AllGGISImportTemplatesQuery", NewAllGGISImportTemplatesQueryFromCBOR)
```

Регистрация выполняется через метод `Register`, который принимает:
- Имя stream'а (обычно в формате `domain.query.module.QueryName`)
- Функцию-конструктор для создания handler'а


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
    
    bus := bus.NewBus(redisClient, ctx)
    
    // Регистрация Query handlers
    bus.Register("vist_domain.query.ggis_import.AllGGISImportTemplatesQuery", NewAllGGISImportTemplatesQueryFromCBOR)
    
    // Запуск обработки сообщений
    bus.Run()
}
```

## Важные моменты

1. **Интерфейс Subscriber**: Handler должен реализовывать интерфейс `bus.Subscriber` с методом `Handle(ctx context.Context) (any, error)`

2. **CBOR десериализация**: Данные приходят в формате CBOR, поэтому используйте `cbor.Unmarshal` для десериализации

3. **Именование stream'ов**: Используйте согласованное именование в формате `domain.query.module.QueryName`

4. **Обработка ошибок**: Функция-конструктор должна возвращать ошибку при неудачной десериализации

5. **Возвращаемые значения**: Метод `Handle` может возвращать любой тип данных (`any`), который будет сериализован в ответ


package redis_bus

// import (
// 	"context"
// 	"log"
// 	"sync"
// 	"time"

// 	"github.com/redis/go-redis/v9"
// )

// type Broker struct {
// 	redis     *redis.Client
// 	listeners []listener
// 	mu        sync.Mutex
// 	buffer    []interface{} // временное хранилище
// }

// func New(redis *redis.Client, listeners []listener) *Broker {
// 	return &Broker{
// 		redis:     redis,
// 		listeners: listeners,
// 		buffer:    make([]interface{}, 0),
// 	}
// }

// func (b *Broker) Run(ctx context.Context) {
// 	const fn = "internal/broker/broker.Run"
// 	streams := make([]string, 0, len(b.listeners.Registered))
// 	ids := make([]string, 0, len(b.listeners.Registered))
// 	for _, stream := range b.listeners.Registered {
// 		streams = append(streams, stream)
// 		ids = append(ids, "0") // "0" Прочитать ВСЁ (в т.ч. старое); "$" только новые сообщения, которые появятся после запуска XREAD (блокирующий режим)
// 	}

// 	for {
// 		xres, err := b.redis.XRead(ctx, &redis.XReadArgs{
// 			Streams: append(streams, ids...),
// 			Block:   2 * time.Second,
// 		}).Result()
// 		if err != nil && err != redis.Nil {
// 			log.Printf("%s: XRead error: %v", fn, err)
// 			continue
// 		}

// 		for _, stream := range xres {
// 			for _, msg := range stream.Messages {
// 				entity, err := b.listeners.Create(stream.Stream) // создаём новый объект // создаём новый объект
// 				if err != nil {
// 					log.Printf("%s: Unable to clone template for stream %s", fn, stream.Stream)
// 					continue
// 				}

// 				transportReq, err := serializer.RadisMsgToTransportReq(msg, entity)
// 				if err != nil {
// 					log.Printf("%s: Decode error for stream %s: %v", fn, stream.Stream, err)
// 					continue
// 				}

// 				b.processMessage(transportReq)
// 			}
// 		}
// 		time.Sleep(100 * time.Microsecond)
// 	}
// }

// // processMessage — заглушка: складывает обработанные данные в буфер
// func (b *Broker) processMessage(data *serializer.TransportRequest) {
// 	b.mu.Lock()
// 	defer b.mu.Unlock()
// 	b.buffer = append(b.buffer, data)
// 	log.Printf("Buffered message: %#v", data)
// }

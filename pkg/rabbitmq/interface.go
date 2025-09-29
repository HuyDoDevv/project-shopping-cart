package rabbitmq

import "context"

type RabbitMQSerivce interface {
	Publish(ctx context.Context, queue string, message any) error
	Consume(ctx context.Context, queue string, handler func([]byte) error) error
	Close() error
}

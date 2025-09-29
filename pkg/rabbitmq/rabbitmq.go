package rabbitmq

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

type rabbitMQSerivce struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	logger  *zerolog.Logger
}

func NewRabbitMQService(amqpURL string, logger *zerolog.Logger) (RabbitMQSerivce, error) {
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		logger.Error().Err(err).Msg("failed to connect to RabbitMQ")
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		logger.Error().Err(err).Msg("failed to open channel")
		return nil, err
	}
	return &rabbitMQSerivce{
		conn:    conn,
		channel: ch,
		logger:  logger,
	}, nil
}

func (rb *rabbitMQSerivce) Publish(ctx context.Context, queue string, message any) error {
	_, err := rb.channel.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		rb.logger.Error().Err(err).Msg("failed to declare queue")
		return err
	}

	body, err := json.Marshal(message)
	if err != nil {
		rb.logger.Error().Err(err).Msg("failed to parse message")
		return err
	}

	err = rb.channel.PublishWithContext(ctx, "", queue, false, false, amqp091.Publishing{
		ContentType: "text/plain",
		Body:        []byte(body),
	})

	if err != nil {
		rb.logger.Error().Err(err).Msg("failed to publish message")
		return err
	}

	return nil
}

func (rb *rabbitMQSerivce) Consume(ctx context.Context, queue string, handler func([]byte) error) error {
	_, err := rb.channel.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		rb.logger.Error().Err(err).Msg("failed to declare queue")
		return err
	}

	msgs, err := rb.channel.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		rb.logger.Error().Err(err).Msg("failed to declare consume")
		return err
	}

	go func() {
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				if err := handler(msg.Body); err != nil {
					msg.Nack(false, false)
				} else {
					msg.Ack(false)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (rb *rabbitMQSerivce) Close() error {
	if rb.channel != nil {
		if err := rb.channel.Close(); err != nil {
			rb.logger.Error().Err(err).Msg("failed to close channel")
			return err
		}
	}

	if rb.conn != nil {
		if err := rb.conn.Close(); err != nil {
			rb.logger.Error().Err(err).Msg("failed to close connection")
			return err
		}
	}
	return nil
}

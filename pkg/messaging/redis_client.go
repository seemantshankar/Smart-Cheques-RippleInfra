package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

type Message struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
	Retries   int                    `json:"retries"`
}

func NewRedisClient(addr, password string, db int) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	// Test connection
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Successfully connected to Redis at %s", addr)

	return &RedisClient{
		client: rdb,
		ctx:    ctx,
	}, nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) HealthCheck() error {
	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
	defer cancel()

	_, err := r.client.Ping(ctx).Result()
	return err
}

func (r *RedisClient) Publish(channel string, message *Message) error {
	message.Timestamp = time.Now()

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.client.Publish(r.ctx, channel, data).Err()
	if err != nil {
		return fmt.Errorf("failed to publish message to channel %s: %w", channel, err)
	}

	log.Printf("Published message %s to channel %s", message.ID, channel)
	return nil
}

func (r *RedisClient) Subscribe(channel string, handler func(*Message) error) error {
	pubsub := r.client.Subscribe(r.ctx, channel)
	defer pubsub.Close()

	log.Printf("Subscribed to channel: %s", channel)

	ch := pubsub.Channel()
	for msg := range ch {
		var message Message
		if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		if err := handler(&message); err != nil {
			log.Printf("Handler error for message %s: %v", message.ID, err)
			// TODO: Implement retry logic
		}
	}

	return nil
}

func (r *RedisClient) EnqueueWithRetry(queue string, message *Message, maxRetries int) error {
	message.Timestamp = time.Now()

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Add to main queue
	err = r.client.LPush(r.ctx, queue, data).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue message: %w", err)
	}

	log.Printf("Enqueued message %s to queue %s", message.ID, queue)
	return nil
}

func (r *RedisClient) DequeueWithRetry(queue string, handler func(*Message) error, maxRetries int) error {
	for {
		// Blocking pop from queue
		result, err := r.client.BRPop(r.ctx, 0, queue).Result()
		if err != nil {
			return fmt.Errorf("failed to dequeue from %s: %w", queue, err)
		}

		var message Message
		if err := json.Unmarshal([]byte(result[1]), &message); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		// Try to process the message
		if err := handler(&message); err != nil {
			message.Retries++
			log.Printf("Handler error for message %s (retry %d/%d): %v",
				message.ID, message.Retries, maxRetries, err)

			if message.Retries < maxRetries {
				// Re-queue for retry
				if retryErr := r.EnqueueWithRetry(queue, &message, maxRetries); retryErr != nil {
					log.Printf("Failed to re-queue message for retry: %v", retryErr)
				}
			} else {
				// Move to dead letter queue
				deadLetterQueue := queue + "_dlq"
				if dlqErr := r.EnqueueWithRetry(deadLetterQueue, &message, 0); dlqErr != nil {
					log.Printf("Failed to move message to dead letter queue: %v", dlqErr)
				}
				log.Printf("Message %s moved to dead letter queue after %d retries",
					message.ID, message.Retries)
			}
		} else {
			log.Printf("Successfully processed message %s", message.ID)
		}
	}
}

func (r *RedisClient) GetQueueLength(queue string) (int64, error) {
	return r.client.LLen(r.ctx, queue).Result()
}

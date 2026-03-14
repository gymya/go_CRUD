package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"gin-quickstart/internal/cache"
	"gin-quickstart/internal/domain"
	"gin-quickstart/internal/events"

	"github.com/segmentio/kafka-go"
)

const (
	consumerCacheAllProductsKey   = "products:all"
	consumerCacheProductKeyPrefix = "products:id:"
)

func consumerProductByIDCacheKey(id int) string {
	return consumerCacheProductKeyPrefix + strconv.Itoa(id)
}

// StockConsumer consumes stock update events and applies them to the database.
type StockConsumer struct {
	reader *kafka.Reader
	repo   domain.ProductRepository
	cache  cache.Cache
}

// NewStockConsumer initializes a Kafka consumer for stock update events.
func NewStockConsumer(brokers []string, topic string, groupID string, repo domain.ProductRepository, cacheStore cache.Cache) (*StockConsumer, error) {
	cleanBrokers := make([]string, 0, len(brokers))
	for _, broker := range brokers {
		b := strings.TrimSpace(broker)
		if b != "" {
			cleanBrokers = append(cleanBrokers, b)
		}
	}
	if len(cleanBrokers) == 0 || strings.TrimSpace(topic) == "" {
		return nil, errors.New("kafka brokers or topic not configured")
	}
	if strings.TrimSpace(groupID) == "" {
		groupID = "product-stock-consumer"
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cleanBrokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	return &StockConsumer{reader: r, repo: repo, cache: cacheStore}, nil
}

// Start begins consuming messages until the context is cancelled.
func (c *StockConsumer) Start(ctx context.Context) error {
	if c == nil || c.reader == nil {
		return errors.New("kafka reader not initialized")
	}
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			log.Printf("Kafka consumer fetch message failed: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var event events.StockUpdateEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Kafka consumer invalid message: %v", err)
			_ = c.reader.CommitMessages(ctx, msg)
			continue
		}
		if event.ProductID == 0 || event.Stock < 0 {
			log.Printf("Kafka consumer invalid payload: product_id=%d stock=%d", event.ProductID, event.Stock)
			_ = c.reader.CommitMessages(ctx, msg)
			continue
		}

		existing, err := c.repo.GetByID(event.ProductID)
		if err != nil {
			if err == domain.ErrNotFound {
				_ = c.reader.CommitMessages(ctx, msg)
				continue
			}
			log.Printf("Kafka consumer get product failed: %v", err)
			continue
		}

		updated := *existing
		updated.Stock = event.Stock
		if _, err := c.repo.Update(event.ProductID, updated); err != nil {
			log.Printf("Kafka consumer update stock failed: %v", err)
			continue
		}

		if c.cache != nil {
			_ = c.cache.Del(ctx, consumerCacheAllProductsKey)
			_ = c.cache.Del(ctx, consumerProductByIDCacheKey(event.ProductID))
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("Kafka consumer commit failed: %v", err)
		}
	}
}

// Close releases the underlying Kafka reader.
func (c *StockConsumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}
	return c.reader.Close()
}

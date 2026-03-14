package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"gin-quickstart/internal/events"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// StockProducer publishes stock update events to Kafka.
type StockProducer struct {
	writer *kafka.Writer
	source string
}

// NewStockProducer initializes a Kafka producer for stock update events.
func NewStockProducer(brokers []string, topic string, source string) (*StockProducer, error) {
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
	if err := ensureTopicExists(cleanBrokers[0], topic); err != nil {
		return nil, err
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(cleanBrokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
		BatchTimeout: 100 * time.Millisecond,
	}

	return &StockProducer{writer: w, source: source}, nil
}

func ensureTopicExists(broker string, topic string) error {
	if strings.TrimSpace(broker) == "" || strings.TrimSpace(topic) == "" {
		return errors.New("kafka broker or topic not configured")
	}

	dialer := &kafka.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	_ = conn.Close()
	controllerAddr := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	controllerConn, err := dialer.Dial("tcp", controllerAddr)
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	err = controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	if err != nil {
		errText := strings.ToLower(err.Error())
		if strings.Contains(errText, "topic with this name already exists") || strings.Contains(errText, "topic_already_exists") {
			return nil
		}
		return err
	}

	return nil
}

func (p *StockProducer) PublishStockUpdate(ctx context.Context, event events.StockUpdateEvent) error {
	if p == nil || p.writer == nil {
		return errors.New("kafka writer not initialized")
	}
	if event.EventID == "" {
		event.EventID = uuid.NewString()
	}
	if event.EventType == "" {
		event.EventType = events.EventTypeProductStockUpdated
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}
	if event.Source == "" {
		event.Source = p.source
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(strconv.Itoa(event.ProductID)),
		Value: payload,
		Time:  time.Now(),
	}

	return p.writer.WriteMessages(ctx, msg)
}

// Close releases the underlying Kafka writer.
func (p *StockProducer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

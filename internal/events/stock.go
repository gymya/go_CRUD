package events

import (
	"context"
	"time"
)

const EventTypeProductStockUpdated = "product.stock.updated"

// StockUpdateEvent represents a stock change request that should be handled asynchronously.
type StockUpdateEvent struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	OccurredAt time.Time `json:"occurred_at"`
	ProductID  int       `json:"product_id"`
	Stock      int       `json:"stock"`
	Source     string    `json:"source,omitempty"`
}

// StockEventPublisher publishes stock update events to an external system (e.g., Kafka).
type StockEventPublisher interface {
	PublishStockUpdate(ctx context.Context, event StockUpdateEvent) error
}
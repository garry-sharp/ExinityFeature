package services

import (
	"fmt"
	"math"
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreaker configuration for kafka publisher doing only 1 request at a time each five seconds
var cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
	Name:        "KafkaPublisher",
	MaxRequests: 1,
	Interval:    5 * time.Second,
	Timeout:     3 * time.Second,
})

// publishes a transaction to Kafka using a circuit breaker
func PublishWithCircuitBreaker(operation func() error) error {
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, operation()
	})
	return err
}

// Retry operation with exponential backoff
func RetryOperation(operation func() error, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		if err := PublishWithCircuitBreaker(operation); err == nil {
			return nil
		}
		backoff := time.Duration(math.Pow(2, float64(i))) * time.Second //the 2^i was a bitwise XOR operation not an exponential backoff
		time.Sleep(backoff)
	}
	return fmt.Errorf("operation failed after %d attempts", maxRetries)
}

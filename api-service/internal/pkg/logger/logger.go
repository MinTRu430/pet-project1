package logger

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

type kafkaLogWriter struct {
	writer      *kafka.Writer
	serviceName string
}

func (w *kafkaLogWriter) Write(p []byte) (int, error) {
	msg := kafka.Message{
		Key:   []byte(w.serviceName),
		Value: p,
	}
	err := w.writer.WriteMessages(context.Background(), msg)

	return len(p), err
}

type KafkaLogger struct {
	logger zerolog.Logger
	writer *kafka.Writer
}

func (l *KafkaLogger) Logger() *zerolog.Logger {
	return &l.logger
}

func (l *KafkaLogger) Close() error {
	return l.writer.Close()
}

func NewKafkaLogger(serviceName string) *KafkaLogger {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		Topic:    "api-logs",
		Balancer: &kafka.RoundRobin{},
		Async:    true,
	})

	kafkaWriter := &kafkaLogWriter{
		writer:      writer,
		serviceName: serviceName,
	}

	logger := zerolog.New(kafkaWriter).With().Timestamp().Str("service", serviceName).Logger()

	return &KafkaLogger{
		logger: logger,
		writer: writer,
	}
}

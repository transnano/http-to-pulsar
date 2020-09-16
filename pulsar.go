package main

import (
	"context"
	"fmt"

	"github.com/apache/pulsar-client-go/pulsar"
)

// Client allows sending batches of Prometheus samples to InfluxDB.
type Client struct {
	client   pulsar.Client
	producer pulsar.Producer
}

// ClientOptions client options
type ClientOptions pulsar.ClientOptions

// Authentication authN
type Authentication pulsar.Authentication

// NewAuthenticationTLS create auth TLS
func NewAuthenticationTLS(certificatePath string, privateKeyPath string) Authentication {
	return pulsar.NewAuthenticationTLS(certificatePath, privateKeyPath)
}

// Config pulsar config
type Config struct {
	ClientOptions
	Topic string
}

// NewClient create client
func NewClient(config Config) (*Client, error) {

	c, err := pulsar.NewClient(pulsar.ClientOptions(config.ClientOptions))
	if err != nil {
		return nil, err
	}
	producer, err := c.CreateProducer(pulsar.ProducerOptions{
		Topic:           config.Topic,
		CompressionType: pulsar.LZ4,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating producer: %w", err)
	}
	return &Client{
		client:   c,
		producer: producer,
	}, nil
}

// Close close client and producer
func (c *Client) Close() error {
	if c.producer != nil {
		err := c.producer.Flush()
		if err != nil {
			return err
		}
		c.producer.Close()
		c.producer = nil
	}
	c.client.Close()
	return nil
}

// Send send to pulsar broker
func (c *Client) Send(s string) error {
	ctx := context.Background()
	msg := &pulsar.ProducerMessage{
		Payload: []byte(s),
		// Key:     s.partitionKey(c.replicaLabels),
	}
	if _, err := c.producer.Send(ctx, msg); err != nil {
		return err
	}
	return nil
}

// SendAsync send to pulsar broker async
func (c *Client) SendAsync(s string) error {
	ctx := context.Background()
	msg := &pulsar.ProducerMessage{
		Payload: []byte(s),
		// Key:     s.partitionKey(c.replicaLabels),
	}

	c.producer.SendAsync(
		ctx, msg,
		func(id pulsar.MessageID, message *pulsar.ProducerMessage, err error) {
			if err != nil {

			}
		},
	)
	return c.producer.Flush()
}

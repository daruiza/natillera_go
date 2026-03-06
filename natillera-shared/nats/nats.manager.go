package nats

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type NatsModuleManager struct {
	natsClient  *nats.Conn
	jetStream   jetstream.JetStream
	orderStream jetstream.Stream
	ctx         context.Context
}

func NewManager(
	natsClient *nats.Conn,
	jetStream jetstream.JetStream,
	orderStream jetstream.Stream,
	ctx context.Context,
) *NatsModuleManager {
	return &NatsModuleManager{
		natsClient:  natsClient,
		jetStream:   jetStream,
		orderStream: orderStream,
		ctx:         ctx,
	}
}

func (manager NatsModuleManager) GetClient() *nats.Conn {
	return manager.natsClient
}

func (manager NatsModuleManager) GetJetStream() jetstream.JetStream {
	return manager.jetStream
}

func (manager NatsModuleManager) GetOrderStream() jetstream.Stream {
	return manager.orderStream
}

func (manager NatsModuleManager) GetContext() context.Context {
	return manager.ctx
}

func (manager *NatsModuleManager) Close() {
	if manager.natsClient != nil {
		manager.natsClient.Drain()
		manager.natsClient.Close()
	}
}

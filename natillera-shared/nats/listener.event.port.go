package nats

import (
	"os"

	"github.com/nats-io/nats.go/jetstream"
)

type ListenerEventPort interface {
	Execute(
		stream string,
		exit <-chan os.Signal,
		batch int,
		eventName string,
		durable string,
		execute func(msg jetstream.Msg),
	) error
}

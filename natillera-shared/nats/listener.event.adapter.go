package nats

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"natillera-shared/utils"

	"github.com/nats-io/nats.go/jetstream"
)

type ListenerEventAdapter struct {
	NatsModuleManager *NatsModuleManager
}

func NewListenerEventAdapter(manager *NatsModuleManager) ListenerEventPort {
	return &ListenerEventAdapter{
		NatsModuleManager: manager,
	}
}

func (listener *ListenerEventAdapter) getContext() context.Context {
	return listener.NatsModuleManager.GetContext()
}

func (listener *ListenerEventAdapter) Execute(
	streamName string,
	exit <-chan os.Signal,
	batch int,
	eventName string,
	durable string,
	execute func(msg jetstream.Msg),
) error {
	eventName = strings.ToLower(eventName)
	js := listener.NatsModuleManager.GetJetStream()
	ctx := listener.getContext()

	stream, err := js.Stream(ctx, streamName)
	if err != nil {
		return fmt.Errorf("error al acceder al stream '%s': %w", streamName, err)
	}

	consume, err := stream.CreateOrUpdateConsumer(
		ctx,
		jetstream.ConsumerConfig{
			Durable:        durable,
			AckPolicy:      jetstream.AckExplicitPolicy,
			DeliverPolicy:  jetstream.DeliverAllPolicy,
			FilterSubjects: []string{eventName},
			AckWait:        10 * time.Second,
		})
	if err != nil {
		return err
	}

	utils.Info.Printf("listening event [%s]\n", eventName)

	go func() {
		defer utils.Info.Println("Se ha dejado de recibir datos")
		for {
			// Usamos select para escuchar tanto los mensajes como la señal de salida.
			select {
			case <-exit:
				return // Termina la goroutine de manera limpia
			default:
				// El default es importante para que la goroutine no se bloquee.
				// Luego, intenta buscar mensajes.
			}

			// Intentamos buscar mensajes en batches
			msgBatch, err := consume.Fetch(batch)
			if err != nil {
				utils.Error.Println("Error en el consumidor de nats: ", err.Error())
				// Puedes añadir una pausa para evitar un bucle de errores muy rápido
				time.Sleep(1 * time.Second)
				continue
			}

			if msgBatch.Error() != nil {
				utils.Error.Println(msgBatch.Error().Error())
				continue
			}

			// Iteramos sobre los mensajes recibidos
			for msg := range msgBatch.Messages() {
				// No necesitamos otro select aquí, el de arriba ya maneja el exit.
				utils.Info.Println("[debug-nats] receiving new message (pre-execute)")
				execute(msg)
			}
		}
	}()

	// El error devuelto al final de la función Execute debe ser el del consumidor.
	return err
}

package nats

import (
	"context"
	"time"

	"natillera-shared/utils"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type NatsStarter struct {
	ManagerDataNats *NatsModuleManager
	EventSender     SenderEventPort
	EventListener   ListenerEventPort
}

type OptsNats struct {
	NameStream string
	Subjects   []string
	MaxAge     time.Duration
}

func NewStartNats(urlNats string, data *OptsNats) *NatsStarter {
	ctx := context.Background()
	for {
		nc, err := nats.Connect(
			urlNats,
			nats.DontRandomize(),
			nats.ReconnectWait(time.Second*3),
			nats.MaxReconnects(-1),
			nats.MaxPingsOutstanding(5),
			nats.PingInterval(10*time.Second),
		)
		if err != nil {
			utils.Error.Println("error al conectar a nats", err)
			time.Sleep(1 * time.Second)
			continue
		}
		js, err := jetstream.New(nc)
		if err != nil {
			utils.Error.Println("error al inicializar jetstream en nats", err)
			time.Sleep(1 * time.Second)
			continue
		}
		utils.Info.Println("connected to JetStream " + urlNats)
		var stream jetstream.Stream
		if data != nil {
			utils.Info.Println("created or updated Stream to JetStream " + urlNats)
			stream, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
				Name:         data.NameStream,
				Subjects:     data.Subjects,
				MaxConsumers: -1,
				Retention:    jetstream.WorkQueuePolicy,
				MaxAge:       data.MaxAge, // 15 días
			})
			if err != nil {
				utils.Error.Println("error al crear el STREAM "+data.NameStream+" en nats", err)
				time.Sleep(1 * time.Second)
				continue
			}
		}

		DataNats := NewManager(nc, js, stream, ctx)
		Sender := NewSenderEventAdapter(DataNats)
		Listener := NewListenerEventAdapter(DataNats)
		return &NatsStarter{
			ManagerDataNats: DataNats,
			EventSender:     Sender,
			EventListener:   Listener,
		}
	}
}
func (st NatsStarter) GetStreamOrder() jetstream.Stream {
	return st.ManagerDataNats.GetOrderStream()
}

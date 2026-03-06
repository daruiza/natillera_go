package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"
)

type SenderEventAdapter struct {
	*NatsModuleManager
}

func NewSenderEventAdapter(manager *NatsModuleManager) SenderEventPort {
	return &SenderEventAdapter{
		NatsModuleManager: manager,
	}
}
func (s SenderEventAdapter) Execute(event *Event) error {
	dataByte, err := json.Marshal(event.Data)
	if err != nil {
		fmt.Println("Cannot serialize object", event.EventType, err.Error())
		return nil
	}
	log.Println("sending " + event.EventType)
	_, err = s.GetJetStream().PublishAsync(strings.ToLower(event.EventType), dataByte)
	if err != nil {
		return err
	}
	return nil
}

func (s SenderEventAdapter) ExecuteMsg(msg *nats.Msg) (string, error) {
	inbox, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	inboxAsString := inbox.String()
	_, err = s.GetJetStream().PublishMsg(context.Background(), msg, jetstream.WithMsgID(inboxAsString))
	if err != nil {
		panic(err)
		return "", err
	}
	return inboxAsString, nil
}

func (s SenderEventAdapter) SendMsgString(event string, msg string) {
	dataByte := []byte(msg)
	log.Println("sending " + event)
	_, err := s.GetJetStream().Publish(context.Background(), strings.ToLower(event), dataByte)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
func (s SenderEventAdapter) SendMsgPB(event string, message proto.Message) {
	dataByte, _ := proto.Marshal(message)

	_, err := s.GetJetStream().Publish(context.Background(), strings.ToLower(event), dataByte)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	log.Println("sending " + event)
}

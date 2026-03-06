package nats

import (
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type SenderEventPort interface {
	Execute(event *Event) error
	ExecuteMsg(msg *nats.Msg) (string, error)
	SendMsgString(event string, msg string)
	SendMsgPB(event string, message proto.Message)
}

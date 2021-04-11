package zeromq

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	zmq "github.com/pebbe/zmq4"
)

// 如果需要Publish新的信息，在下面加
const (
	FilterFrame        = iota // Frame
	NetworkResponse           // Response
	TrackerTrackResult        // TrackResult
	EncodeTime                // int64(nanosecond)
	TrackingTime              // int64(nanosecond)
	ProcessTime               // int64(nanosecond)
	ClientToServerTime        // int64(nanosecond)
	ServerToClientTime        // int64(nanosecond)
)

// Topic represents message type
type Topic int

// Message 消息的定义
type Message struct {
	Topic   Topic
	Content interface{}
}

type MQ struct {
	publisher *zmq.Socket
	port      int
}

func NewMQ(port int) MQ {
	MQ := MQ{}

	p, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		log.Errorln(err)
	}
	p.Bind(fmt.Sprintf("tcp://*:%v", port))

	MQ.publisher = p
	MQ.port = port

	return MQ
}

func (mq *MQ) Subscribe(topic Topic) chan Message{
	s, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		log.Errorln(err)
	}
	ch := make(chan Message, 100)
	s.Connect(fmt.Sprintf("tcp://localhost:%v", mq.port))
	s.SetSubscribe("")
	// go func(){
	// 	s.RecvMessage()
	// }
	return ch
}

func (mq *MQ) Unsubscribe() {

}

func (mq *MQ) Publish(message Message) {

}

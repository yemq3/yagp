package main

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// 如果需要Publish新的信息，在下面加
const (
	FilterFrame        = iota // Frame
	DetectResult              // DetectResult
	TrackResult               // TrackResult
	EncodeTime                // int64(nanosecond)
	TrackingTime              // int64(nanosecond)
	ProcessTime               // int64(nanosecond)
	ClientToServerTime        // int64(nanosecond)
	ServerToClientTime        // int64(nanosecond)
)

// Topic represents message type
type Topic int

// MessageCenter 用来发布Frame, Detection Result, Track Result，Qos等信息
type MessageCenter struct {
	publishChannel     chan Message
	subscriberChannels map[Topic]map[chan Message]struct{}
	channelToTopic     map[chan Message]Topic
	ensureOrder        bool

	mu *sync.Mutex
}

// Message 消息的定义
type Message struct {
	Topic   Topic
	Content interface{}
}

// NewMessageCenter create a new MessageCenter
func NewMessageCenter(ensureOrder bool) MessageCenter {
	messageCenter := MessageCenter{}

	messageCenter.publishChannel = make(chan Message, 100)
	messageCenter.subscriberChannels = make(map[Topic]map[chan Message]struct{})
	messageCenter.channelToTopic = make(map[chan Message]Topic)
	messageCenter.ensureOrder = ensureOrder
	messageCenter.mu = &sync.Mutex{}

	return messageCenter
}

func (m *MessageCenter) run() {
	log.Infof("Message Center running...")
	for {
		select {
		case msg := <-m.publishChannel:
			m.mu.Lock()
			chs := m.subscriberChannels[msg.Topic]
			for ch := range chs {
				if m.ensureOrder {
					m.sendMessage(ch, msg)
				} else {
					go m.sendMessage(ch, msg)
				}
			}
			m.mu.Unlock()
		}
	}
}

func (m *MessageCenter) sendMessage(ch chan Message, msg Message) {
	select {
	case ch <- msg:
		log.Debugf("send a msg of %v", msg.Topic)
	// 当开启顺序保证时，下面这串代码有可能阻塞0.5秒的信息
	case <-time.After(1000 * time.Millisecond):
		m.Unsubscribe(ch)
		log.Warnf("send message timeout, msg:%#v", msg)
	}
}

// Subscribe 订阅某个Topic的信息
func (m *MessageCenter) Subscribe(topic Topic) chan Message {
	m.mu.Lock()
	defer m.mu.Unlock()

	subscriberChannel := make(chan Message, 100)

	subs, ok := m.subscriberChannels[topic]
	if !ok {
		subs = make(map[chan Message]struct{})
	}
	subs[subscriberChannel] = struct{}{}
	m.subscriberChannels[topic] = subs
	m.channelToTopic[subscriberChannel] = topic

	return subscriberChannel
}

// Unsubscribe 取消订阅信息
func (m *MessageCenter) Unsubscribe(ch chan Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	topic := m.channelToTopic[ch]
	subs := m.subscriberChannels[topic]
	delete(subs, ch)
	delete(m.channelToTopic, ch)
	m.subscriberChannels[topic] = subs
}

// Publish 发布某个Topic的信息
func (m *MessageCenter) Publish(message Message) {
	m.publishChannel <- message
}

package main

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

// 如果需要Publish新的信息，在下面加
const (
	FilterFrame        = iota
	NetworkResponse
	TrackerTrackResult
)

type Topic int

type MessageCenter struct {
	publishChannel     chan Message
	subscriberChannels map[Topic]map[chan Message]struct{}
	channelToTopic     map[chan Message]Topic

	mu sync.Mutex
}

type Message struct {
	Topic   Topic
	Content interface{}
}

func (m *MessageCenter) init() {
	m.publishChannel = make(chan Message, 100)
	m.subscriberChannels = make(map[Topic]map[chan Message]struct{})
	m.channelToTopic = make(map[chan Message]Topic)
}

func (m *MessageCenter) run() {
	for {
		select {
		case msg := <-m.publishChannel:
			m.mu.Lock()
			chs := m.subscriberChannels[msg.Topic]
			m.mu.Unlock()
			for ch, _ := range chs {
				m.sendMessage(ch, msg)
			}
		}
	}
}

func (m *MessageCenter) sendMessage(ch chan Message, msg Message) {
	select {
	case ch <- msg:
		log.Debugf("send a msg of %v", msg.Topic)

	// 用goroutine发送信息的次序不能保证正确，不用goroutine下面这串代码有可能阻塞1秒的信息，先注释了
	case <-time.After(1 * time.Second):
		m.Unsubscribe(ch)
		log.Debugf("send message timeout")
	}
}

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

func (m *MessageCenter) Unsubscribe(ch chan Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	topic := m.channelToTopic[ch]
	subs := m.subscriberChannels[topic]
	delete(subs, ch)
	delete(m.channelToTopic, ch)
	m.subscriberChannels[topic] = subs
}

func (m *MessageCenter) Publish(message Message) {
	m.publishChannel <- message
}

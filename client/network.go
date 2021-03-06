package main

import (
	"bytes"
	"compress/gzip"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/yemq3/yagp/box"
)

// Network ...
type Network struct {
	serverAddr     url.URL
	conn           *websocket.Conn
	NetworkChannel chan EncodedFrame
	messageCenter  MessageCenter
	// enableCompression bool
}

func gzipCompress(b []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(b); err != nil {
		log.Errorln(err)
		return []byte{}, err
	}
	zw.Flush()
	return buf.Bytes(), nil
}

func (network *Network) replyHandler() {
	for {
		select {
		default:
			response := Response{}
			err := network.conn.ReadJSON(&response)
			if err != nil {
				log.Errorln(err)
				return
			}
			log.Debugf("recv: %v", response)
			serverToClientTime := time.Now().UnixNano() - response.SendTime

			boxes := make([]box.AbsoluteBox, 0)
			for _, box := range response.Boxes {
				abbox := box.ToAbsoluteBox(WIDTH, HEIGHT)
				abbox.UUID = uuid.NewV4()
				boxes = append(boxes, abbox)
			}
			detectResult := ResultWithAbsoluteBox{
				FrameID:  response.FrameID,
				Boxes:    boxes,
				DoneTime: time.Now().UnixNano(),
				Method:   DETECT,
			}

			msg := Message{
				Topic:   DetectResult,
				Content: detectResult,
			}
			network.messageCenter.Publish(msg)

			msg = Message{
				Topic:   ClientToServerTime,
				Content: response.ClientToServerTime,
			}
			network.messageCenter.Publish(msg)

			msg = Message{
				Topic:   ServerToClientTime,
				Content: serverToClientTime,
			}
			network.messageCenter.Publish(msg)

			msg = Message{
				Topic:   ProcessTime,
				Content: response.ProcessTime,
			}
			network.messageCenter.Publish(msg)
		}
	}
}

// NewNetwork create a new network
func NewNetwork(addr url.URL, messageCenter MessageCenter) (Network, error) {
	network := Network{}

	dialer := websocket.DefaultDialer

	conn, _, err := dialer.Dial(addr.String(), nil)
	if err != nil {
		log.Errorln(err)
		return network, err
	}

	network.serverAddr = addr
	network.NetworkChannel = make(chan EncodedFrame)
	network.conn = conn
	network.messageCenter = messageCenter

	return network, nil

}

func (network *Network) run() {
	log.Infof("Network running...")
	go network.replyHandler()
	go network.send()
}

func (network *Network) send() {
	defer network.conn.Close()
	for {
		select {
		case frame := <-network.NetworkChannel:
			log.Debugf("get a new image to send")
			request := Request{
				FrameID:  frame.FrameID,
				Frame:    frame.Frame,
				SendTime: time.Now().UnixNano(),
			}

			err := network.conn.WriteJSON(request)

			if err != nil {
				log.Errorln(err)
				return
			}
		}
	}
}

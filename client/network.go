package main

import (
	"bytes"
	"compress/gzip"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// Network ...
type Network struct {
	serverAddr     url.URL
	conn           *websocket.Conn
	NetworkChannel chan EncodedFrame
	messageCenter  MessageCenter
	// enableCompression bool
}

// Request ...
type Request struct {
	FrameID  int
	Frame    []byte
	SendTime int64
}

// Response ...
type Response struct {
	FrameID            int
	Boxes              []Box
	ClientToServerTime int64
	SendTime           int64
	ProcessTime        int64
}

// Box ...
type Box struct {
	X1      float64
	Y1      float64
	X2      float64
	Y2      float64
	Unknown float64 // 我真不知道这是个啥
	Conf    float64
	Name    string
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

			msg := Message{
				Topic:   NetworkResponse,
				Content: response,
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

func (network *Network) init(addr url.URL, messageCenter MessageCenter) error {
	log.Infoln("Network Init")
	dialer := websocket.DefaultDialer
	// dialer.EnableCompression = true

	conn, _, err := dialer.Dial(addr.String(), nil)
	// conn.EnableWriteCompression(true)
	if err != nil {
		log.Errorln(err)
		return err
	}

	network.serverAddr = addr
	network.NetworkChannel = make(chan EncodedFrame)
	network.conn = conn
	network.messageCenter = messageCenter

	return nil
}

func (network *Network) run() {
	go network.replyHandler()
	go network.send()
}

func (network *Network) send() {
	defer network.conn.Close()
	for {
		select {
		case frame := <-network.NetworkChannel:
			log.Debugf("get a new image to send")
			request := Request{}
			request.FrameID = frame.FrameID
			request.Frame = frame.Frame
			request.SendTime = time.Now().UnixNano()

			err := network.conn.WriteJSON(request)
			// b, err := json.Marshal(frame)
			// if err != nil {
			// 	log.Errorln(err)
			// 	return
			// }
			// err = network.conn.WriteMessage(websocket.BinaryMessage, b)
			if err != nil {
				log.Errorln(err)
				return
			}
		}
	}
}

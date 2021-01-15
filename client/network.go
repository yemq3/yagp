package main

import (
	"net/url"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Network struct {
	serverAddr        url.URL
	SendChannel       chan []byte
	enableCompression bool
}

func (network *Network) init() (*websocket.Conn, error) {
	dialer := websocket.DefaultDialer
	// dialer.EnableCompression = true

	conn, _, err := dialer.Dial(network.serverAddr.String(), nil)
	// ws.EnableWriteCompression(true)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return conn, nil
}

func (network *Network) run(){
	conn, err := network.init()
	defer conn.close()
	
}

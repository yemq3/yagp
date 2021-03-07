package main

import (
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
	"time"
)

// Encoder 用于进行图像编码
type Encoder struct {
	EncoderChannel chan Frame
	networkChannel chan EncodedFrame
	encodeQuality  int
	messageCenter MessageCenter
}

// EncodedFrame ...
type EncodedFrame struct {
	Frame         []byte
	FrameID       int
	EncodeQuality int
	// EncodeMethod  string // 暂时只支持jpg，在非嵌入式设备上用webp可能会更好
}

// func (encoder *Processer) SetEncodeQuality(encodeQuality int){
// 	encoder.encodeQuality = encodeQuality
// }

// NewEncoder creates a new Encoder
func NewEncoder (encodeQuality int, networkChannel chan EncodedFrame, messageCenter MessageCenter) Encoder { 
	encoder := Encoder{}

	encoder.EncoderChannel = make(chan Frame)
	encoder.encodeQuality = encodeQuality
	encoder.networkChannel = networkChannel
	encoder.messageCenter = messageCenter

	return encoder
}

func (encoder *Encoder) run() {
	log.Infof("Encoder running...")
	for {
		select {
		case frame := <-encoder.EncoderChannel:
			log.Debugf("get a new image to encode")
			img := frame.Frame
			start := time.Now().UnixNano()
			buffer, err := gocv.IMEncodeWithParams(gocv.JPEGFileExt, img, []int{gocv.IMWriteJpegQuality, encoder.encodeQuality})
			encodeTime := time.Now().UnixNano() - start
			if err != nil {
				log.Errorln(err)
				return
			}
			encodedFrame := EncodedFrame{}
			encodedFrame.Frame = buffer
			encodedFrame.FrameID = frame.FrameID
			encodedFrame.EncodeQuality = encoder.encodeQuality
			go func(encodedFrame EncodedFrame) {
				encoder.networkChannel <- encodedFrame
			}(encodedFrame)

			msg := Message{
				Topic:   EncodeTime,
				Content: encodeTime,
			}
			encoder.messageCenter.Publish(msg)
		}
	}
}

package main

import (
	"time"

	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

// Encoder 用于进行图像编码
type Encoder struct {
	EncoderChannel chan EncodeTask
	networkChannel chan EncodedFrame
	encodeQuality  int
	messageCenter  MessageCenter
}

// EncodeTask 编码任务的信息
type EncodeTask struct {
	frame         Frame
	encodeQuality int
	resizeFactor  float64
	encodeMethod  gocv.FileExt
}

// EncodedFrame ...
type EncodedFrame struct {
	Frame         []byte
	FrameID       int
	EncodeQuality int
}

// func (encoder *Processer) SetEncodeQuality(encodeQuality int){
// 	encoder.encodeQuality = encodeQuality
// }

// NewEncoder creates a new Encoder
func NewEncoder(encodeQuality int, networkChannel chan EncodedFrame, messageCenter MessageCenter) Encoder {
	encoder := Encoder{}

	encoder.EncoderChannel = make(chan EncodeTask)
	encoder.encodeQuality = encodeQuality
	encoder.networkChannel = networkChannel
	encoder.messageCenter = messageCenter

	return encoder
}

func (encoder *Encoder) run() {
	log.Infof("Encoder running...")
	for {
		select {
		case task := <-encoder.EncoderChannel:
			log.Debugf("get a new image to encode")
			img := task.frame.Frame
			start := time.Now().UnixNano()
			buffer, err := gocv.IMEncodeWithParams(task.encodeMethod, img, []int{gocv.IMWriteJpegQuality, task.encodeQuality})
			encodeTime := time.Now().UnixNano() - start
			if err != nil {
				log.Errorln(err)
				return
			}
			encodedFrame := EncodedFrame{}
			encodedFrame.Frame = buffer
			encodedFrame.FrameID = task.frame.FrameID
			encodedFrame.EncodeQuality = task.encodeQuality
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

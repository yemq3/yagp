package main

import (
	"time"
	
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

// EncodedFrame ...
type EncodedFrame struct {
	Frame         []byte
	FrameID       int
	EncodeQuality int
	// EncodeMethod  string // 暂时只支持jpg，在非嵌入式设备上用webp可能会更好
}

// Encoder ...
type Encoder struct {
	EncodeChannel chan Frame
	sendChannel   chan EncodedFrame
	encodeQuality int
}

// func (encoder *Encoder) SetEncodeQuality(encodeQuality int){
// 	encoder.encodeQuality = encodeQuality
// }

func (encoder *Encoder) init(encodeQuality int, sendChannel chan EncodedFrame) error {
	log.Infoln("Encoder Init")
	encoder.EncodeChannel = make(chan Frame)
	encoder.encodeQuality = encodeQuality
	encoder.sendChannel = sendChannel

	return nil
}

func (encoder *Encoder) run() {
	for {
		select {
		case frame := <-encoder.EncodeChannel:
			log.Debugf("get a new image to encode")
			img := frame.Frame
			start := time.Now().UnixNano()
			buffer, err := gocv.IMEncodeWithParams(gocv.JPEGFileExt, img, []int{gocv.IMWriteJpegQuality, encoder.encodeQuality})
			log.Infof("encode use time %v", (time.Now().UnixNano()-start))
			if err != nil {
				log.Errorln(err)
				return
			}
			encodedFrame := EncodedFrame{}
			encodedFrame.Frame = buffer
			encodedFrame.FrameID = frame.FrameID
			encodedFrame.EncodeQuality = encoder.encodeQuality
			go func(encodedFrame EncodedFrame) {
				encoder.sendChannel <- encodedFrame
			}(encodedFrame)
		}
	}
}

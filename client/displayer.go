package main

import (
	"image"
	"image/color"

	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

// Displayer 用来显示结果，可以参照下面的display()写
type Displayer struct {
	messageCenter MessageCenter
}

func (displayer *Displayer) init(messageCenter MessageCenter) error {
	log.Infoln("Displayer Init")
	displayer.messageCenter = messageCenter

	return nil
}

func (displayer *Displayer) display() {
	window := gocv.NewWindow("Oringin")
	defer window.Close()

	ch := displayer.messageCenter.Subscribe("Frame")
	defer displayer.messageCenter.Unsubscribe(ch)

	for {
		select {
		case msg := <-ch:
			frame, ok := msg.Content.(Frame)
			if !ok {
				log.Errorf("wrong msg")
				return
			}
			window.IMShow(frame.Frame)
			window.WaitKey(10)
		}
	}
}

func (displayer *Displayer) displayDetectionRes() {
	window := gocv.NewWindow("Detection Result")
	defer window.Close()
	red := color.RGBA{255, 0, 0, 0}

	frameChannel := displayer.messageCenter.Subscribe("Frame")
	defer displayer.messageCenter.Unsubscribe(frameChannel)

	responseChannel := displayer.messageCenter.Subscribe("Response")
	defer displayer.messageCenter.Unsubscribe(responseChannel)

	frames := make(map[int]gocv.Mat)

	for {
		select {
		case msg := <- frameChannel:
			frame, ok := msg.Content.(Frame)
			if !ok{
				log.Errorf("wrong msg")
				return
			}
			frames[frame.FrameID] = frame.Frame
		case msg := <- responseChannel:
			response, ok := msg.Content.(Response)
			if !ok{
				log.Errorf("wrong msg")
				return
			}
			img := frames[response.FrameID]
			height := img.Size()[0]
			weight := img.Size()[1]
			for _, box := range response.Boxes {
				rect := image.Rectangle{}
				rect.Min = image.Point{int(box.X1 * float64(weight)), int(box.Y1 * float64(height))}
				rect.Max = image.Point{int(box.X2 * float64(weight)), int(box.Y2 * float64(height))}
				gocv.Rectangle(&img, rect, red, 3)
			}
			// 这个window的size不对，不知道怎么搞的= =
			window.IMShow(img)
			window.WaitKey(10)
			delete(frames, response.FrameID)
		}
	}
}

package main

import (
	"image"
	"image/color"

	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

type Result struct {
	frame  gocv.Mat
	rects  []image.Rectangle
	Method string
}

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

	ch := displayer.messageCenter.Subscribe(FilterFrame)
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
			window.WaitKey(1)
		}
	}
}

// 包含Tracking和Detection的结果
func (displayer *Displayer) displayResult() {
	window := gocv.NewWindow("Both")
	defer window.Close()
	red := color.RGBA{255, 0, 0, 0}
	blue := color.RGBA{0, 0, 255, 0}

	frameChannel := displayer.messageCenter.Subscribe(FilterFrame)
	defer displayer.messageCenter.Unsubscribe(frameChannel)

	responseChannel := displayer.messageCenter.Subscribe(NetworkResponse)
	defer displayer.messageCenter.Unsubscribe(responseChannel)

	trackingChannel := displayer.messageCenter.Subscribe(TrackerTrackResult)
	defer displayer.messageCenter.Unsubscribe(trackingChannel)

	frames := make(map[int]gocv.Mat)

	for {
		select {
		case msg := <-frameChannel:
			frame, ok := msg.Content.(Frame)
			if !ok {
				log.Errorf("wrong msg")
				return
			}
			frames[frame.FrameID] = frame.Frame
		case msg := <-responseChannel:
		priority:
			// 如果此时frameChannel还有frame没处理，先处理，要不然之后可能会取出nil
			for {
				select {
				case msg := <-frameChannel:
					frame, ok := msg.Content.(Frame)
					if !ok {
						log.Errorf("get wrong msg")
						return
					}
					frames[frame.FrameID] = frame.Frame
				default:
					break priority
				}
			}
			response, ok := msg.Content.(Response)
			if !ok {
				log.Errorf("wrong msg")
				return
			}
			img := frames[response.FrameID]
			copyImg := gocv.NewMat()
			img.CopyTo(&copyImg)
			height := copyImg.Size()[0]
			weight := copyImg.Size()[1]
			for _, box := range response.Boxes {
				x0 := max(int(box.X1 * float64(weight)), 0)
				y0 := max(int(box.Y1 * float64(height)), 0)
				x1 := min(int(box.X2 * float64(weight)), weight)
				y1 := min(int(box.Y2 * float64(height)), height)
				rect := image.Rect(x0, y0, x1, y1)
				gocv.Rectangle(&copyImg, rect, red, 3)
			}

			window.IMShow(copyImg)
			window.WaitKey(10)
			delete(frames, response.FrameID)
		case msg := <-trackingChannel:
		priority2:
			// 如果此时frameChannel还有frame没处理，先处理，要不然之后可能会取出nil
			for {
				select {
				case msg := <-frameChannel:
					frame, ok := msg.Content.(Frame)
					if !ok {
						log.Errorf("get wrong msg")
						return
					}
					frames[frame.FrameID] = frame.Frame
				default:
					break priority2
				}
			}
			trackResult, ok := msg.Content.(TrackResult)
			if !ok {
				log.Errorf("wrong msg")
				return
			}
			img := frames[trackResult.FrameID]
			copyImg := gocv.NewMat()
			img.CopyTo(&copyImg)

			for _, rect := range trackResult.Boxes {
				gocv.Rectangle(&copyImg, rect, blue, 3)
			}
			window.IMShow(copyImg)
			window.WaitKey(1)
			delete(frames, trackResult.FrameID)
		}
	}
}

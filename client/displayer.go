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

func (displayer *Displayer) displayDetectionRes() {
	window := gocv.NewWindow("Detection Result")
	defer window.Close()
	red := color.RGBA{255, 0, 0, 0}

	frameChannel := displayer.messageCenter.Subscribe(FilterFrame)
	defer displayer.messageCenter.Unsubscribe(frameChannel)

	responseChannel := displayer.messageCenter.Subscribe(NetworkResponse)
	defer displayer.messageCenter.Unsubscribe(responseChannel)

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
			response, ok := msg.Content.(Response)
			if !ok {
				log.Errorf("wrong msg")
				return
			}
			i := frames[response.FrameID]
			img := i
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
			window.WaitKey(1)
			delete(frames, response.FrameID)
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
			log.Infof("%v", response)
			img := frames[response.FrameID]
			copyImg := gocv.NewMat()
			img.CopyTo(&copyImg)
			height := copyImg.Size()[0]
			weight := copyImg.Size()[1]
			for _, box := range response.Boxes {
				rect := image.Rectangle{}
				rect.Min = image.Point{int(box.X1 * float64(weight)), int(box.Y1 * float64(height))}
				rect.Max = image.Point{int(box.X2 * float64(weight)), int(box.Y2 * float64(height))}
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

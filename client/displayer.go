package main

import (
	"image"
	"image/color"

	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

// Displayer 用来显示结果，可以参照下面的display()写
type Displayer struct {
	NewFrameNotify    chan Frame
	NewResponseNotify chan DetectionResult
}

func (displayer *Displayer) init() error {
	log.Infoln("Displayer Init")
	displayer.NewFrameNotify = make(chan Frame)
	displayer.NewResponseNotify = make(chan DetectionResult)

	return nil
}

func (displayer *Displayer) display() {
	window := gocv.NewWindow("Oringin")
	defer window.Close()
	for {
		select {
		case frame := <-displayer.NewFrameNotify:
			window.IMShow(frame.Frame)
			window.WaitKey(10)
		}
	}
}

func (displayer *Displayer) displayDetectionRes() {
	window := gocv.NewWindow("Detection Result")
	defer window.Close()
	red := color.RGBA{255, 0, 0, 0}
	for {
		select {
		case detectionResult := <-displayer.NewResponseNotify:
			img := detectionResult.Frame.Frame
			height := img.Size()[0]
			weight := img.Size()[1]
			response := detectionResult.Response
			for _, box := range response.Boxes {
				rect := image.Rectangle{}
				rect.Min = image.Point{int(box.X1 * float64(weight)), int(box.Y1 * float64(height))}
				rect.Max = image.Point{int(box.X2 * float64(weight)), int(box.Y2 * float64(height))}
				gocv.Rectangle(&img, rect, red, 3)
			}
			// 这个window的size不对，不知道怎么搞的= =
			window.IMShow(img)
			window.WaitKey(10)
		}
	}
}

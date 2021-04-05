package main

import (
	"fmt"
	"image"
	"image/color"

	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func display(frame currentFrame, window *gocv.Window) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered in f", r)
		}
	}()
	red := color.RGBA{255, 0, 0, 0}
	// blue := color.RGBA{0, 0, 255, 0}

	if frame.method == NONE {
		window.IMShow(frame.frame)
		window.WaitKey(1)
		return
	}

	copyFrame := gocv.NewMat()
	frame.frame.CopyTo(&copyFrame)

	for _, box := range frame.Boxes {
		gocv.Rectangle(&copyFrame, box.Rect, red, 2)
	}

	if frame.resultFrameID <= frame.frameID {
		text := fmt.Sprintf("delay: %v", frame.frameID-frame.resultFrameID)
		position := image.Point{X: 50, Y: 80}
		gocv.PutText(&copyFrame, text, position, gocv.FontHersheyPlain, 8, red, 8)
	}

	window.IMShow(copyFrame)
	window.WaitKey(1)
}

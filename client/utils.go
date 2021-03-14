package main

import (
	"fmt"
	"image"
	"image/color"

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
	red := color.RGBA{255, 0, 0, 0}
	// blue := color.RGBA{0, 0, 255, 0}
	// log.Infof("%v", frame.frameID)

	if frame.method == NONE {
		window.IMShow(frame.frame)
		window.WaitKey(1)
		return
	}

	copyFrame := gocv.NewMat()
	frame.frame.CopyTo(&copyFrame)
	height := copyFrame.Size()[0]
	weight := copyFrame.Size()[1]

	for _, box := range frame.Boxes {
		x0 := max(int(box.X1*float64(weight)), 0)
		y0 := max(int(box.Y1*float64(height)), 0)
		x1 := min(int(box.X2*float64(weight)), weight)
		y1 := min(int(box.Y2*float64(height)), height)
		rect := image.Rect(x0, y0, x1, y1)
		gocv.Rectangle(&copyFrame, rect, red, 2)
	}

	if frame.resultFrameID <= frame.frameID {
		text := fmt.Sprintf("delay: %v", frame.frameID-frame.resultFrameID)
		position := image.Point{X: 50, Y: 80}
		gocv.PutText(&copyFrame, text, position, gocv.FontHersheyPlain, 8, red, 8)
	}

	window.IMShow(copyFrame)
	window.WaitKey(1)
}

package main

import (
	"gocv.io/x/gocv"
	"image"
)

func main() {
	webcam, err := gocv.OpenVideoCapture(0)
	if err != nil {
		return
	}
	defer webcam.Close()

	// open display window
	window := gocv.NewWindow("Tracking")
	defer window.Close()

	// create a tracker instance
	// (one of MIL, KCF, TLD, MedianFlow, Boosting, MOSSE or CSRT)
	tracker := gocv.NewTrackerMIL()
	//tracker := contrib.NewTrackerKCF()
	defer tracker.Close()

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	// read an initial image
	if ok := webcam.Read(&img); !ok {
		return
	}

	// this work
	//rect := gocv.SelectROI("Tracking", img)

	// but this error
	rect := image.Rect(105, 41, 686, 654)

	tracker.Init(img, rect)

}

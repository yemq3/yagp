package main

import (
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"
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
	tracker := contrib.NewTrackerTLD()
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

	log.Infof("%v", img.Size())
	// but this error
	rect := image.Rect(-1, 0, 300, 400)

	tracker.Init(img, rect)

}

package main

import (
	"fmt"
	"log"
	"image/color"
	"os"
	"time"

	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"
)

func main() {
	if len(os.Args) < 2 {
		log.Println("How to run:\n\ttracking [camera ID]")
		return
	}

	// parse args
	deviceID := os.Args[1]

	// open webcam
	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		log.Printf("Error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// open display window
	window := gocv.NewWindow("Tracking")
	defer window.Close()

	// create a tracker instance
	// (one of MIL, KCF, TLD, MedianFlow, Boosting, MOSSE or CSRT)
	tracker := contrib.NewTrackerMOSSE()
	defer tracker.Close()

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	// read an initial image
	if ok := webcam.Read(&img); !ok {
		log.Printf("cannot read device %v\n", deviceID)
		return
	}

	// let the user mark a ROI to track
	rect := gocv.SelectROI("Tracking", img)
	if rect.Max.X == 0 {
		log.Printf("user cancelled roi selection\n")
		return
	}

	// initialize the tracker with the image & the selected roi
	init := tracker.Init(img, rect)
	if !init {
		log.Printf("Could not initialize the Tracker")
		return
	}

	// color for the rect to draw
	blue := color.RGBA{0, 0, 255, 0}
	fmt.Printf("Start reading device: %v\n", deviceID)
	his := make([]int64, 0)
	for {
		if ok := webcam.Read(&img); !ok {
			log.Printf("Device closed: %v\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		// update the roi
		start := time.Now().UnixNano()
		rect, _ := tracker.Update(img)
		usedTime := time.Now().UnixNano()-start
		log.Printf("tracking use time: %v\n", usedTime)
		his = append(his, usedTime)


		// draw it.
		gocv.Rectangle(&img, rect, blue, 3)

		// show the image in the window, and wait 10 millisecond
		window.IMShow(img)
		if window.WaitKey(10) >= 0 {
			break
		}
	}
	var total int64 = 0
	for _, v := range his{
		total += v
	}
	log.Printf("average time: %v\n", float64(total)/float64(len(his))/1000000000.0)
}
package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

// Frame ...
type Frame struct {
	FrameID int
	Frame   gocv.Mat
}

// Camera ...
type Camera struct {
	camera         *gocv.VideoCapture
	encodeChannel  chan gocv.Mat
	sendChannel    chan []byte
	history        map[int]gocv.Mat // 存之前的图像
	encodeQuality  int
	latestFrameID  int
	newFrameNotify chan int

	weight    int
	height    int
	frameRate int
}

func (camera *Camera) init(weight float64, height float64, frameRate float64) (*gocv.VideoCapture, error){
	c, err := gocv.VideoCaptureDevice(0)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	c.Set(gocv.VideoCaptureFrameWidth, weight)
	c.Set(gocv.VideoCaptureFrameHeight, height)
	c.Set(gocv.VideoCaptureFPS, frameRate)
	camera.camera = c
	return c, nil
}


func (camera *Camera) run(weight int, height int, frameRate int){
	c, err := camera.init()
	
	defer c.Close()
}

func defaultFilterFunc(img gocv.Mat) bool{
	return true
}

func (camera *Camera) getFrame(ctx context.Context, filterFunc func(gocv.Mat) bool, encodeChannel chan Frame) {
	img := gocv.NewMat()
	defer img.Close()
	for {
		select {
		case <-ctx.Done():
			log.Infoln("getFrame interrupt")
			return
		default:
			camera.camera.Read(&img)
			camera.latestFrameID++
			camera.history[camera.latestFrameID] = img

			if filterFunc(img){
				frame := Frame{}
				frame.FrameID = camera.latestFrameID
				frame.Frame = img
				go func(frame Frame) {
					encodeChannel <- frame
				}(frame)
				go func(frameID int) {
					camera.newFrameNotify <- frameID
				}(camera.latestFrameID)
			}
		}
	}
}

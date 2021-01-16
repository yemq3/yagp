package main

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

// FilterFunc ...
type FilterFunc func(gocv.Mat) bool

// Frame ...
type Frame struct {
	FrameID   int
	Frame     gocv.Mat
	Timestamp int64
}

// Camera ...
type Camera struct {
	camera        *gocv.VideoCapture
	encodeChannel chan Frame
	history       map[int]Frame // 存之前的图像
	latestFrameID int
	wg            sync.WaitGroup
	persister     Persister

	// weight    int
	// height    int
	// frameRate int
}

func (camera *Camera) init(weight float64, height float64, frameRate float64, encodeChannel chan Frame, persister Persister) error {
	log.Infoln("Camera Init")
	c, err := gocv.VideoCaptureDevice(0)
	if err != nil {
		log.Errorln(err)
		return err
	}
	// 以下设置不一定生效，得看摄像头支不支持，且Set()不会返回信息，校验的话得另想办法
	c.Set(gocv.VideoCaptureFrameWidth, weight)
	c.Set(gocv.VideoCaptureFrameHeight, height)
	c.Set(gocv.VideoCaptureFPS, frameRate)
	camera.camera = c

	camera.encodeChannel = encodeChannel
	camera.latestFrameID = -1
	camera.persister = persister
	camera.wg = sync.WaitGroup{}

	return nil
}

func (camera *Camera) run(filterFunc FilterFunc) {
	defer camera.camera.Close()

	camera.wg.Add(1)

	go camera.getFrame(filterFunc)

	camera.wg.Wait()
}

func defaultFilterFunc(img gocv.Mat) bool {
	return true
}

func (camera *Camera) getFrame(filterFunc func(gocv.Mat) bool) {
	img := gocv.NewMat()
	defer img.Close()
	defer camera.wg.Done()
	for {
		select {
		default:
			log.Debugf("Get a new frame")
			camera.camera.Read(&img)
			camera.latestFrameID++
			frame := Frame{}
			frame.FrameID = camera.latestFrameID
			frame.Frame = img
			frame.Timestamp = time.Now().Unix()
			go func(frame Frame) {
				camera.persister.persistFrame(frame)
			}(frame)

			if filterFunc(img) {
				go func(frame Frame) {
					camera.encodeChannel <- frame
				}(frame)
			}
		}
	}
}

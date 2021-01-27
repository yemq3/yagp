package main

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

// Frame ...
type Frame struct {
	FrameID   int
	Frame     gocv.Mat
	Timestamp int64
	isDetect  bool
}

// Camera ...
type Camera struct {
	camera        *gocv.VideoCapture
	filterChannel chan Frame
	history       map[int]Frame // 存之前的图像
	latestFrameID int
	wg            sync.WaitGroup

	// weight    int
	// height    int
	// frameRate int
}

func (camera *Camera) init(weight float64, height float64, frameRate float64, filterChannel chan Frame) error {
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

	camera.filterChannel = filterChannel
	camera.latestFrameID = -1
	camera.wg = sync.WaitGroup{}

	return nil
}

func (camera *Camera) run() {
	defer camera.camera.Close()

	camera.wg.Add(1)

	go camera.getFrame()

	camera.wg.Wait()
}

func (camera *Camera) getFrame() {
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
			frame.Timestamp = time.Now().UnixNano()

			go func(frame Frame) {
				camera.filterChannel <- frame
			}(frame)
		}
	}
}

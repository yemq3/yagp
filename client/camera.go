package main

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)


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
	//c, err := gocv.VideoCaptureDevice(0)
	c, err := gocv.VideoCaptureFile("../benchmark/tracker/video.mp4")
	if err != nil {
		log.Errorln(err)
		return err
	}
	// 以下设置不一定生效，得看摄像头支不支持，且Set()不会返回信息，校验的话得另想办法
	//c.Set(gocv.VideoCaptureFrameWidth, weight)
	//c.Set(gocv.VideoCaptureFrameHeight, height)
	//c.Set(gocv.VideoCaptureFPS, frameRate)
	camera.camera = c

	camera.filterChannel = filterChannel
	camera.latestFrameID = 0
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
		time.Sleep(50 * time.Millisecond)
		select {
		default:
			log.Debugf("Camera get a new frame, frameid:%v", camera.latestFrameID+1)
			ok := camera.camera.Read(&img)
			if !ok{
				log.Errorf("can't read image")
				return
			}
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

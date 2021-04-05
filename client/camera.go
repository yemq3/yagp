package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

// Camera ...
type Camera struct {
	camera        *gocv.VideoCapture
	frameRate     int
	filterChannel chan Frame
	history       map[int]Frame // 存之前的图像
	latestFrameID int

	// width    int
	// height    int
	// frameRate int
}

// NewCamera creates a new Camera
// @param video 设为true时捕获videoFile的视频
func NewCamera(frameRate int, filterChannel chan Frame, video bool, videoFile string) (Camera, error) {
	camera := Camera{}
	var c *gocv.VideoCapture
	var err error

	if video == false {
		c, err = gocv.VideoCaptureDevice(0)
	} else {
		c, err = gocv.VideoCaptureFile(videoFile)
	}
	if err != nil {
		log.Errorln(err)
		return camera, err
	}
	// 以下设置不一定生效，得看摄像头支不支持，且Set()不会返回信息，校验的话得另想办法
	// c.Set(gocv.VideoCaptureFrameWidth, float64(width))
	//c.Set(gocv.VideoCaptureFrameHeight, float64(height))
	//c.Set(gocv.VideoCaptureFPS, float64(frameRate))
	camera.camera = c
	camera.frameRate = frameRate
	camera.filterChannel = filterChannel
	camera.latestFrameID = 0

	return camera, nil
}

func (camera *Camera) run() {
	log.Infof("Camera running...")
	img := gocv.NewMat()
	defer img.Close()
	defer camera.camera.Close()
	for {
		// Set不一定管用，手动控制帧率
		time.Sleep(time.Duration(1000/camera.frameRate) * time.Millisecond)
		select {
		default:
			log.Debugf("Camera get a new frame, frameid:%v", camera.latestFrameID+1)
			ok := camera.camera.Read(&img)
			if !ok {
				log.Errorf("can't read image")
				time.Sleep(2 * time.Second)
				os.Exit(0)
				return
			}
			camera.latestFrameID++
			frame := Frame{}
			frame.FrameID = camera.latestFrameID
			frame.Frame = img
			frame.Timestamp = time.Now().UnixNano()

			WIDTH = img.Size()[1]
			HEIGHT = img.Size()[0]

			go func(frame Frame) {
				camera.filterChannel <- frame
			}(frame)
		}
	}
}

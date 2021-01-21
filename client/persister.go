package main

import (
	"time"

	log "github.com/sirupsen/logrus"
)

// DetectionResult ...
type DetectionResult struct {
	Frame    Frame
	Response Response
}

// Persister 不想写，先直接存内存
type Persister struct {
	history           map[int]DetectionResult
	NewFrameNotify    chan Frame
	NewResponseNotify chan DetectionResult
}

func (persister *Persister) persistFrame(frame Frame) {
	detectionResult := DetectionResult{}
	detectionResult.Frame = frame
	persister.history[frame.FrameID] = detectionResult
	go func(frame Frame) {
		persister.NewFrameNotify <- frame
	}(frame)
}

func (persister *Persister) persistResponse(response Response) {
	detectionResult := persister.readPersist(response.FrameID)
	detectionResult.Response = response
	persister.history[response.FrameID] = detectionResult
	log.Infof("use time %v", (time.Now().UnixNano()-detectionResult.Frame.Timestamp))
	go func(detectionResult DetectionResult) {
		persister.NewResponseNotify <- detectionResult
	}(detectionResult)
}

func (persister *Persister) readPersist(FrameID int) DetectionResult {
	return persister.history[FrameID]
}

func (persister *Persister) init(newFrameNotify chan Frame, newResponseNotify chan DetectionResult) error {
	log.Infoln("Persister Init")
	persister.NewFrameNotify = newFrameNotify
	persister.NewResponseNotify = newResponseNotify
	persister.history = make(map[int]DetectionResult)

	return nil
}

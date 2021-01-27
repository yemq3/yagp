package main

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	None int = 0
	Detect int = 1
	Track  int = 2
	Both   int = 3
)

// DetectionResult ...
type DetectionResult struct {
	Frame       Frame
	Method      int
	Response    Response
	TrackResult TrackResult
}

// Persister 不想写，先直接存内存
type Persister struct {
	history           map[int]DetectionResult
	NewFrameNotify    chan Frame
	NewResponseNotify chan DetectionResult

	mu sync.Mutex
}

func (persister *Persister) persistFrame(frame Frame) {
	persister.mu.Lock()
	defer persister.mu.Unlock()
	detectionResult := DetectionResult{}
	detectionResult.Frame = frame
	persister.history[frame.FrameID] = detectionResult

	// 给displayer发的
	go func(frame Frame) {
		persister.NewFrameNotify <- frame
	}(frame)
}

func (persister *Persister) persistResponse(response Response) {
	persister.mu.Lock()
	defer persister.mu.Unlock()
	detectionResult := persister.history[response.FrameID]
	detectionResult.Response = response
	if detectionResult.Method == Track{
		detectionResult.Method = Both
	} else {
		detectionResult.Method = Detect
	}
	persister.history[response.FrameID] = detectionResult
	log.Infof("use time %v", time.Now().UnixNano() - detectionResult.Frame.Timestamp)

	// 给displayer发的
	go func(detectionResult DetectionResult) {
		persister.NewResponseNotify <- detectionResult
	}(detectionResult)
}

func (persister *Persister) persistTrackResult(result TrackResult)  {
	persister.mu.Lock()
	defer persister.mu.Unlock()
	detectionResult := persister.history[result.FrameID]
	detectionResult.TrackResult = result
	if detectionResult.Method == Detect{
		detectionResult.Method = Both
	} else {
		detectionResult.Method = Track
	}
	persister.history[result.FrameID] = detectionResult

}

func (persister *Persister) readPersist(FrameID int) DetectionResult {
	persister.mu.Lock()
	defer persister.mu.Unlock()
	return persister.history[FrameID]
}

func (persister *Persister) init(newFrameNotify chan Frame, newResponseNotify chan DetectionResult) error {
	log.Infoln("Persister Init")
	persister.NewFrameNotify = newFrameNotify
	persister.NewResponseNotify = newResponseNotify
	persister.history = make(map[int]DetectionResult)

	return nil
}

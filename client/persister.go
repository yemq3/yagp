package main

import (
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
	"sync"
)

const (
	None   = iota
	Detect
	Track
	Both
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
	history       map[int]DetectionResult
	latestFrameID int
	messageCenter MessageCenter
	originVideoWriter *gocv.VideoWriter
	resultVideoWriter *gocv.VideoWriter

	mu sync.Mutex
}

func (persister *Persister) persistFrame(frame Frame) {
	persister.mu.Lock()
	defer persister.mu.Unlock()
	detectionResult := DetectionResult{}
	detectionResult.Frame = frame
	persister.history[frame.FrameID] = detectionResult
}

func (persister *Persister) persistResponse(response Response) {
	persister.mu.Lock()
	defer persister.mu.Unlock()
	detectionResult := persister.history[response.FrameID]
	detectionResult.Response = response
	if detectionResult.Method == Track {
		detectionResult.Method = Both
	} else {
		detectionResult.Method = Detect
	}
	persister.history[response.FrameID] = detectionResult

}

func (persister *Persister) persistTrackResult(result TrackResult) {
	persister.mu.Lock()
	defer persister.mu.Unlock()
	detectionResult := persister.history[result.FrameID]
	detectionResult.TrackResult = result
	if detectionResult.Method == Detect {
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

func (persister *Persister) init(messageCenter MessageCenter) error {
	log.Infoln("Persister Init")
	persister.history = make(map[int]DetectionResult)
	persister.messageCenter = messageCenter

	return nil
}

func (persister *Persister) run() {
	frameChannel := persister.messageCenter.Subscribe(FilterFrame)
	defer persister.messageCenter.Unsubscribe(frameChannel)

	responseChannel := persister.messageCenter.Subscribe(NetworkResponse)
	defer persister.messageCenter.Unsubscribe(responseChannel)

	trackChannel := persister.messageCenter.Subscribe(TrackerTrackResult)
	defer persister.messageCenter.Unsubscribe(trackChannel)



	for {
		select {
		case msg := <-frameChannel:
			frame, ok := msg.Content.(Frame)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			persister.persistFrame(frame)
		case msg := <-responseChannel:
			response, ok := msg.Content.(Response)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			persister.persistResponse(response)
		case msg := <-trackChannel:
			trackResult, ok := msg.Content.(TrackResult)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			persister.persistTrackResult(trackResult)
		}
	}
}

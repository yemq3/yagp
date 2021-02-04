package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"
	"image"
	"time"
)

var supportTrackingMethod = map[string]struct{}{
	"KCF":        {},
	"CSRT":       {},
	"MIL":        {},
	"MOSSE":      {},
	"Boosting":   {},
	"MedianFlow": {},
	"TLD":        {},
}

type TrackResult struct {
	FrameID  int
	Boxes    []image.Rectangle
	DoneTime int64
}

type Tracker struct {
	TrackerChannel    chan Frame
	messageCenter     MessageCenter
	trackingAlgorithm string
	trackers          []contrib.Tracker
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func (tracker *Tracker) init(messageCenter MessageCenter, trackingAlgorithm string) error {
	tracker.TrackerChannel = make(chan Frame)
	tracker.messageCenter = messageCenter
	tracker.trackingAlgorithm = trackingAlgorithm
	if _, ok := supportTrackingMethod[trackingAlgorithm]; !ok {
		return fmt.Errorf("unsupport tracking method")
	}
	return nil
}

func (tracker *Tracker) makeTracker(img gocv.Mat, rectangle image.Rectangle) (contrib.Tracker, error) {
	var t contrib.Tracker
	if tracker.trackingAlgorithm == "KCF" {
		t = contrib.NewTrackerKCF()
	} else if tracker.trackingAlgorithm == "CSRT" {
		t = contrib.NewTrackerCSRT()
	} else if tracker.trackingAlgorithm == "MIL" {
		t = contrib.NewTrackerMIL()
	} else if tracker.trackingAlgorithm == "MOSSE" {
		t = contrib.NewTrackerMOSSE()
	} else if tracker.trackingAlgorithm == "Boosting" {
		t = contrib.NewTrackerBoosting()
	} else if tracker.trackingAlgorithm == "MedianFlow" {
		t = contrib.NewTrackerMedianFlow()
	} else {
		t = contrib.NewTrackerTLD()
	}
	init := t.Init(img, rectangle)
	if !init {
		err := fmt.Errorf("could not initialize the tracker")
		return nil, err
	}
	return t, nil
}

func (tracker *Tracker) closeAllTracker() {
	for _, t := range tracker.trackers {
		t.Close()
	}
	tracker.trackers = make([]contrib.Tracker, 0)
}

func (tracker *Tracker) run() {
	frameChannel := tracker.messageCenter.Subscribe(FilterFrame)
	defer tracker.messageCenter.Unsubscribe(frameChannel)

	responseChannel := tracker.messageCenter.Subscribe(NetworkResponse)
	defer tracker.messageCenter.Unsubscribe(responseChannel)

	frames := make(map[int]gocv.Mat)

	for {
		select {
		case frame := <-tracker.TrackerChannel:
			log.Debugf("Tracker get Frame %v", frame.FrameID)
			Boxes := make([]image.Rectangle, 0)

			start := time.Now().UnixNano()
			for _, tracker := range tracker.trackers {
				rect, ok := tracker.Update(frame.Frame)
				if ok {
					Boxes = append(Boxes, rect)
				} else {
					log.Debugf("tracker lost object")
				}
			}
			trackingTime := time.Now().UnixNano() - start

			trackResult := TrackResult{
				FrameID: frame.FrameID,
				Boxes:   Boxes,
				DoneTime: time.Now().UnixNano(),
			}
			msg := Message{
				Topic:   TrackerTrackResult,
				Content: trackResult,
			}
			tracker.messageCenter.Publish(msg)

			msg = Message{
				Topic:   TrackingTime,
				Content: trackingTime,
			}
			tracker.messageCenter.Publish(msg)

		case msg := <-responseChannel:
		priority:
			// 如果此时frameChannel还有frame没处理，先处理，要不然之后可能会取出nil
			for {
				select {
				case msg := <-frameChannel:
					frame, ok := msg.Content.(Frame)
					if !ok {
						log.Errorf("get wrong msg")
						return
					}
					frames[frame.FrameID] = frame.Frame
				default:
					break priority
				}
			}
			response, ok := msg.Content.(Response)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}

			// 清空之前所有的tracker
			tracker.closeAllTracker()

			img := frames[response.FrameID]
			height := img.Size()[0]
			weight := img.Size()[1]
			for _, box := range response.Boxes {
				x0 := max(int(box.X1*float64(weight)), 0)
				y0 := max(int(box.Y1*float64(height)), 0)
				x1 := min(int(box.X2*float64(weight)), weight)
				y1 := min(int(box.Y2*float64(height)), height)
				rect := image.Rect(x0, y0, x1, y1)
				t, err := tracker.makeTracker(img, rect)
				if err != nil {
					return
				}
				tracker.trackers = append(tracker.trackers, t)
			}
			delete(frames, response.FrameID)
		case msg := <-frameChannel:
			frame, ok := msg.Content.(Frame)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			frames[frame.FrameID] = frame.Frame
		}
	}
}

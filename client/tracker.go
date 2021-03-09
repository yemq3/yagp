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

// Tracker 用于实现跟踪算法
type Tracker struct {
	TrackerChannel    chan Frame
	messageCenter     MessageCenter
	trackingAlgorithm string
	trackers          []trackerWithName
}

type trackerWithName struct {
	t                 contrib.Tracker
	name              string // 物体标签
	trackingAlgorithm string
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

// NewTracker creates a new tracker
func NewTracker(messageCenter MessageCenter, trackingAlgorithm string) (Tracker, error) {
	tracker := Tracker{}

	if _, ok := supportTrackingMethod[trackingAlgorithm]; !ok {
		return tracker, fmt.Errorf("unsupport tracking method")
	}

	tracker.TrackerChannel = make(chan Frame)
	tracker.messageCenter = messageCenter
	tracker.trackingAlgorithm = trackingAlgorithm

	return tracker, nil
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
	for _, tracker := range tracker.trackers {
		tracker.t.Close()
	}
	tracker.trackers = make([]trackerWithName, 0)
}

func (tracker *Tracker) run() {
	log.Infof("Tracker running...")
	frameChannel := tracker.messageCenter.Subscribe(FilterFrame)
	defer tracker.messageCenter.Unsubscribe(frameChannel)

	responseChannel := tracker.messageCenter.Subscribe(NetworkResponse)
	defer tracker.messageCenter.Unsubscribe(responseChannel)

	frames := make(map[int]gocv.Mat)
	totalObject := 0
	var totalTime int64 = 0
	for {
		select {
		case frame := <-tracker.TrackerChannel:
			log.Debugf("Tracker get Frame %v", frame.FrameID)
			Boxes := make([]image.Rectangle, 0)

			start := time.Now().UnixNano()
			for _, tracker := range tracker.trackers {
				rect, ok := tracker.t.Update(frame.Frame)
				if ok {
					Boxes = append(Boxes, rect)
				} else {
					log.Debugf("tracker lost object")
				}
			}
			trackingTime := time.Now().UnixNano() - start
			totalObject += len(tracker.trackers)
			totalTime += trackingTime
			//log.Infof("%v objects, tracking time: %v, average tracking time: %v",
			//	len(tracker.trackers),
			//	float64(trackingTime)/1000000000.0,
			//	float64(trackingTime)/float64(len(tracker.trackers))/1000000000.0,
			//)
			//log.Infof("total object: %v, total tracking time: %v, average tracking time: %v",
			//	totalObject,
			//	float64(totalTime)/1000000000.0,
			//	float64(totalTime)/float64(totalObject)/1000000000.0,
			//)

			trackResult := TrackResult{
				FrameID:  frame.FrameID,
				Boxes:    Boxes,
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

			// 对每个box初始化一个tracker
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
				tracker.trackers = append(tracker.trackers, trackerWithName{
					t:                 t,
					name:              box.Name,
					trackingAlgorithm: tracker.trackingAlgorithm,
				})
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

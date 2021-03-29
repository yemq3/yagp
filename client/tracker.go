package main

import (
	"fmt"
	"image"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/yemq3/yagp/box"
	"github.com/yemq3/yagp/lk"
	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"
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
	TrackerChannel    chan TrackTask
	messageCenter     MessageCenter
	trackingAlgorithm string
	trackers          []trackerWithInfo
	perviousBoxes     []box.AbsoluteBox
	lkTracker         lk.LKTracker
}

type TrackTask struct {
	frame  Frame //
	isDrop bool  // true 代表这帧不做检测，直接复制上一帧的结果
}

type trackerWithInfo struct {
	t                 contrib.Tracker
	name              string  // 物体标签
	conf              float64 //
	trackingAlgorithm string  //
}

// NewTracker creates a new tracker
func NewTracker(messageCenter MessageCenter, trackingAlgorithm string) (Tracker, error) {
	tracker := Tracker{}

	if _, ok := supportTrackingMethod[trackingAlgorithm]; !ok {
		return tracker, fmt.Errorf("unsupport tracking method")
	}

	tracker.TrackerChannel = make(chan TrackTask)
	tracker.messageCenter = messageCenter
	tracker.trackingAlgorithm = trackingAlgorithm

	return tracker, nil
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
	tracker.trackers = make([]trackerWithInfo, 0)
}

func (tracker *Tracker) publishTrackResult(trackResult ResultWithAbsoluteBox, trackingTime int64) {
	msg := Message{
		Topic:   TrackResult,
		Content: trackResult,
	}
	tracker.messageCenter.Publish(msg)

	msg = Message{
		Topic:   TrackingTime,
		Content: trackingTime,
	}
	tracker.messageCenter.Publish(msg)
}

func (tracker *Tracker) run() {
	log.Infof("Tracker running...")
	frameChannel := tracker.messageCenter.Subscribe(FilterFrame)
	defer tracker.messageCenter.Unsubscribe(frameChannel)

	detectChannel := tracker.messageCenter.Subscribe(DetectResult)
	defer tracker.messageCenter.Unsubscribe(detectChannel)

	frames := make(map[int]gocv.Mat)
	for {
		select {
		case task := <-tracker.TrackerChannel:
			log.Debugf("Tracker get Frame %v", task.frame.FrameID)
			if task.isDrop {
				// 该帧不跟踪，直接复制上一帧的结果发出
				trackResult := ResultWithAbsoluteBox{
					FrameID:  task.frame.FrameID,
					Boxes:    tracker.perviousBoxes,
					DoneTime: time.Now().UnixNano(),
					Method:   DROP,
				}
				tracker.publishTrackResult(trackResult, 0)
				break
			}
			Boxes := make([]box.AbsoluteBox, 0)

			start := time.Now().UnixNano()
			for _, tracker := range tracker.trackers {
				rect, ok := tracker.t.Update(task.frame.Frame)
				if ok {
					box := box.AbsoluteBox{
						Rect: rect,
						Conf: tracker.conf,
						Name: tracker.name,
					}
					Boxes = append(Boxes, box)
				} else {
					log.Debugf("tracker lost object")
				}
			}
			tracker.perviousBoxes = Boxes

			trackingTime := time.Now().UnixNano() - start

			trackResult := ResultWithAbsoluteBox{
				FrameID:  task.frame.FrameID,
				Boxes:    Boxes,
				DoneTime: time.Now().UnixNano(),
				Method:   TRACK,
			}
			tracker.publishTrackResult(trackResult, trackingTime)

		case msg := <-detectChannel:
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
			response, ok := msg.Content.(ResultWithAbsoluteBox)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}

			// 清空之前所有的tracker
			tracker.closeAllTracker()

			// 对每个box初始化一个tracker
			img := frames[response.FrameID]
			for _, box := range response.Boxes {
				t, err := tracker.makeTracker(img, box.Rect)
				if err != nil {
					return
				}
				tracker.trackers = append(tracker.trackers, trackerWithInfo{
					t:                 t,
					name:              box.Name,
					conf:              box.Conf,
					trackingAlgorithm: tracker.trackingAlgorithm,
				})
			}

			tracker.perviousBoxes = response.Boxes

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

func (tracker *Tracker) runUseLK() {
	log.Infof("Tracker running...")
	frameChannel := tracker.messageCenter.Subscribe(FilterFrame)
	defer tracker.messageCenter.Unsubscribe(frameChannel)

	detectChannel := tracker.messageCenter.Subscribe(DetectResult)
	defer tracker.messageCenter.Unsubscribe(detectChannel)

	frames := make(map[int]gocv.Mat)
	for {
		select {
		case task := <-tracker.TrackerChannel:
			log.Debugf("Tracker get Frame %v", task.frame.FrameID)
			if task.isDrop {
				// 该帧不跟踪，直接复制上一帧的结果发出
				trackResult := ResultWithAbsoluteBox{
					FrameID:  task.frame.FrameID,
					Boxes:    tracker.perviousBoxes,
					DoneTime: time.Now().UnixNano(),
					Method:   DROP,
				}
				tracker.publishTrackResult(trackResult, 0)
				break
			}

			start := time.Now().UnixNano()
			Boxes := tracker.lkTracker.Update(task.frame.Frame)
			
			tracker.perviousBoxes = Boxes

			trackingTime := time.Now().UnixNano() - start

			trackResult := ResultWithAbsoluteBox{
				FrameID:  task.frame.FrameID,
				Boxes:    Boxes,
				DoneTime: time.Now().UnixNano(),
				Method:   TRACK,
			}
			tracker.publishTrackResult(trackResult, trackingTime)

		case msg := <-detectChannel:
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
			response, ok := msg.Content.(ResultWithAbsoluteBox)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}


			img := frames[response.FrameID]
			tracker.perviousBoxes = response.Boxes

			t := lk.NewLKTracker(1000, 0.1, 10)
			t.Init(img, response.Boxes)

			tracker.lkTracker = t

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

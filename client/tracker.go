package main

import (
	"fmt"
	"image"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
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
	"lk":         {},
}

// Tracker 用于实现跟踪算法
type Tracker struct {
	TrackerChannel      chan TrackTask
	UpdateChannel       chan UpdateTask
	messageCenter       MessageCenter
	algorithm           string            // 当前用的算法
	realTracker         MultiTracker      //
	trackerRWMutex      *sync.RWMutex     // realTracker的读写锁
	lkTracker           *lk.LKTracker     //
	perviousBoxes       []box.AbsoluteBox // 记录前一次的box，drop的时候直接再传一次
	frameHistory        map[int]gocv.Mat  // 记录过去的帧
	latestUpdateFrameID int               //
	hisRWMutex          *sync.RWMutex     // frameHistory的读写锁
}

type MultiTracker interface {
	Init(gocv.Mat, []box.AbsoluteBox)
	Update(gocv.Mat) []box.AbsoluteBox
}

type TrackTask struct {
	Frame  Frame
	IsDrop bool // true 代表这帧不做检测，直接复制上一帧的结果
}

type UpdateTask struct {
	FrameID  int
	Boxes    []box.AbsoluteBox
	Interval int
}

type trackerWithInfo struct {
	t    contrib.Tracker
	uuid uuid.UUID
	name string // 物体标签
	conf float64
}

type OpencvMultiTracker struct {
	t         []trackerWithInfo
	algorithm string //
}

func NewOpencvMultiTracker(algorithm string) *OpencvMultiTracker {
	mt := &OpencvMultiTracker{
		algorithm: algorithm,
	}
	return mt
}

func (tracker *OpencvMultiTracker) Init(image gocv.Mat, boxes []box.AbsoluteBox) {
	// 清空之前所有的tracker
	tracker.closeAllTracker()

	// 对每个box初始化一个tracker
	for _, box := range boxes {
		t, err := tracker.makeTracker(image, box.Rect)
		if err != nil {
			return
		}
		tracker.t = append(tracker.t, trackerWithInfo{
			t:    t,
			uuid: box.UUID,
			name: box.Name,
			conf: box.Conf,
		})
	}
}

func (tracker *OpencvMultiTracker) Update(image gocv.Mat) []box.AbsoluteBox {
	Boxes := make([]box.AbsoluteBox, 0)
	for _, tracker := range tracker.t {
		rect, ok := tracker.t.Update(image)
		if ok {
			box := box.AbsoluteBox{
				Rect: rect,
				UUID: tracker.uuid,
				Conf: tracker.conf,
				Name: tracker.name,
			}
			Boxes = append(Boxes, box)
		} else {
			log.Debugf("tracker lost object")
		}
	}
	return Boxes
}

func (tracker *OpencvMultiTracker) makeTracker(img gocv.Mat, rectangle image.Rectangle) (contrib.Tracker, error) {
	var t contrib.Tracker
	switch tracker.algorithm {
	case "KCF":
		t = contrib.NewTrackerKCF()
	case "CSRT":
		t = contrib.NewTrackerCSRT()
	case "MIL":
		t = contrib.NewTrackerMIL()
	case "MOSSE":
		t = contrib.NewTrackerMOSSE()
	case "Boosting":
		t = contrib.NewTrackerBoosting()
	case "MedianFlow":
		t = contrib.NewTrackerMedianFlow()
	case "TLD":
		t = contrib.NewTrackerTLD()
	}
	init := t.Init(img, rectangle)
	if !init {
		err := fmt.Errorf("could not initialize the tracker")
		return nil, err
	}
	return t, nil
}

func (tracker *OpencvMultiTracker) closeAllTracker() {
	for _, tracker := range tracker.t {
		tracker.t.Close()
	}
	tracker.t = make([]trackerWithInfo, 0)
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

// NewTracker creates a new tracker
func NewTracker(messageCenter MessageCenter, trackingAlgorithm string) (Tracker, error) {
	tracker := Tracker{}

	if _, ok := supportTrackingMethod[trackingAlgorithm]; !ok {
		return tracker, fmt.Errorf("unsupport tracking method")
	}

	tracker.TrackerChannel = make(chan TrackTask)
	tracker.UpdateChannel = make(chan UpdateTask)
	tracker.messageCenter = messageCenter
	tracker.algorithm = trackingAlgorithm
	tracker.trackerRWMutex = new(sync.RWMutex)
	tracker.frameHistory = make(map[int]gocv.Mat)
	tracker.hisRWMutex = new(sync.RWMutex)

	return tracker, nil
}

// 监听frame channel，保存之前的帧用于跟踪
func (tracker *Tracker) frameWorker() {
	frameChannel := tracker.messageCenter.Subscribe(FilterFrame)
	defer tracker.messageCenter.Unsubscribe(frameChannel)

	for {
		select {
		case msg := <-frameChannel:
			frame, ok := msg.Content.(Frame)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			tracker.hisRWMutex.Lock()
			tracker.frameHistory[frame.FrameID] = frame.Frame
			tracker.hisRWMutex.Unlock()
		}
	}
}

// 监听trackTask，跟踪并发布结果
func (tracker *Tracker) trackWorker() {
	for {
		select {
		case task := <-tracker.TrackerChannel:
			log.Debugf("Tracker get Frame %v", task.Frame.FrameID)

			// 该帧不跟踪，直接复制上一帧的结果发出
			if task.IsDrop {
				trackResult := ResultWithAbsoluteBox{
					FrameID:  task.Frame.FrameID,
					Boxes:    tracker.perviousBoxes,
					DoneTime: time.Now().UnixNano(),
					Method:   DROP,
				}
				tracker.publishTrackResult(trackResult, 0)
				break
			}

			// 跟踪并发布
			start := time.Now().UnixNano()
			tracker.trackerRWMutex.RLock()
			if tracker.realTracker == nil {
				tracker.trackerRWMutex.RUnlock()
				break
			}
			Boxes := tracker.realTracker.Update(task.Frame.Frame)
			tracker.trackerRWMutex.RUnlock()
			tracker.perviousBoxes = Boxes

			trackingTime := time.Now().UnixNano() - start

			trackResult := ResultWithAbsoluteBox{
				FrameID:  task.Frame.FrameID,
				Boxes:    Boxes,
				DoneTime: time.Now().UnixNano(),
				Method:   TRACK,
			}
			tracker.publishTrackResult(trackResult, trackingTime)
		}
	}
}

// 监听updateTask，会在适当的时候更新tracker，并删除之前的帧
func (tracker *Tracker) updateWorker() {
	for {
		select {
		case task := <-tracker.UpdateChannel:
			var newTracker MultiTracker
			if tracker.algorithm == "lk" {
				newTracker = lk.NewLKTracker(500, 0.01, 10)
			} else {
				newTracker = NewOpencvMultiTracker(tracker.algorithm)
			}

			tracker.hisRWMutex.RLock()
			image := tracker.frameHistory[task.FrameID]
			newTracker.Init(image, task.Boxes)
			tracker.hisRWMutex.RUnlock()

			tracker.hisRWMutex.RLock()
			for i := 1; true; i++ {
				image, ok := tracker.frameHistory[task.FrameID+i*task.Interval]
				if !ok {
					break
				}
				_ = newTracker.Update(image)
			}
			tracker.hisRWMutex.RUnlock()

			tracker.trackerRWMutex.Lock()
			tracker.realTracker = newTracker
			tracker.trackerRWMutex.Unlock()

			tracker.hisRWMutex.Lock()
			// 把之前的frameid全部删掉
			for i := tracker.latestUpdateFrameID + 1; i <= task.FrameID; i++ {
				delete(tracker.frameHistory, i)
			}
			tracker.latestUpdateFrameID = task.FrameID
			tracker.hisRWMutex.Unlock()
		}
	}
}

func (tracker *Tracker) run() {
	log.Infof("Tracker running...")
	go tracker.frameWorker()
	go tracker.trackWorker()
	go tracker.updateWorker()
}

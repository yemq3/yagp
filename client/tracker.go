package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"
	"image"
	"time"
)

type TrackResult struct {
	FrameID int
	Boxes []image.Rectangle
}

type Tracker struct {
	TrackerChannel    chan Frame
	messageCenter     MessageCenter
	trackingAlgorithm string
	trackers          []gocv.Tracker
}

func (tracker *Tracker) init(messageCenter MessageCenter, trackingAlgorithm string) error {
	tracker.TrackerChannel = make(chan Frame)
	tracker.messageCenter = messageCenter
	if trackingAlgorithm != "KCF" && trackingAlgorithm != "CSRT" && trackingAlgorithm != "MIL" {
		return fmt.Errorf("unsupport tracking algorithm")
	} else {
		tracker.trackingAlgorithm = trackingAlgorithm
	}

	return nil
}

func (tracker *Tracker) makeTracker(img gocv.Mat, rectangle image.Rectangle) (gocv.Tracker, error){
	var t gocv.Tracker
	if tracker.trackingAlgorithm == "KCF" {
		t = contrib.NewTrackerKCF()
	} else if tracker.trackingAlgorithm == "CSRT" {
		t = contrib.NewTrackerCSRT()
	} else {
		t = gocv.NewTrackerMIL()
	}
	init := t.Init(img, rectangle)
	if !init {
		err := fmt.Errorf("could not initialize the tracker")
		return nil, err
	}
	return t, nil
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

			for _, tracker := range tracker.trackers{
				start := time.Now().UnixNano()
				rect, ok := tracker.Update(frame.Frame)
				log.Infof("tracking use time: %v, ok: %v", float64(time.Now().UnixNano()-start) / 1000000000.0, ok)
				if ok{
					Boxes = append(Boxes, rect)
				} else{
					log.Debugf("tracker lost object")
				}
			}

			trackResult := TrackResult{
				FrameID: frame.FrameID,
				Boxes:   Boxes,
			}

			//trackResult := TrackResult{}


			msg := Message{
				Topic:   TrackerTrackResult,
				Content: trackResult,
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
			for _, t := range tracker.trackers{
				t.Close()
			}
			tracker.trackers = make([]gocv.Tracker, 0) // 清空之前的tracker
			img := frames[response.FrameID]
			height := img.Size()[0]
			weight := img.Size()[1]
			for _, box := range response.Boxes {
				//rect := image.Rectangle{}
				//rect.Min = image.Point{int(box.X1 * float64(weight)), int(box.Y1 * float64(height))}
				//rect.Max = image.Point{int(box.X2 * float64(weight)), int(box.Y2 * float64(height))}
				rect := image.Rect(int(box.X1 * float64(weight)), int(box.Y1 * float64(height)), int(box.X2 * float64(weight)), int(box.Y2 * float64(height)))
				t, err := tracker.makeTracker(img, rect)
				if err != nil{
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

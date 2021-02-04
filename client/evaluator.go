package main

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type Evaluator struct {
	lastProcessFrameID int
	Delays             []int64
	EncodeTime         []int64
	TrackingTime       []int64
	ProcessTime        []int64
	C2STime            []int64
	S2CTime            []int64
	messageCenter      MessageCenter
}

func (evaluator *Evaluator) init(messageCenter MessageCenter) {
	evaluator.messageCenter = messageCenter
}

func (evaluator *Evaluator) run() {
	// 这里应该搞个参数来控制，先mark了，还没做参数的分离
	ticker := time.NewTicker(5 * time.Second)

	frameChannel := evaluator.messageCenter.Subscribe(FilterFrame)
	defer evaluator.messageCenter.Unsubscribe(frameChannel)
	frameTime := make(map[int]int64)

	responseChannel := evaluator.messageCenter.Subscribe(NetworkResponse)
	defer evaluator.messageCenter.Unsubscribe(responseChannel)

	trackingChannel := evaluator.messageCenter.Subscribe(TrackerTrackResult)
	defer evaluator.messageCenter.Unsubscribe(trackingChannel)

	encodeTimeChannel := evaluator.messageCenter.Subscribe(EncodeTime)
	defer evaluator.messageCenter.Unsubscribe(encodeTimeChannel)

	trackingTimeChannel := evaluator.messageCenter.Subscribe(TrackingTime)
	defer evaluator.messageCenter.Unsubscribe(trackingTimeChannel)

	processTimeChannel := evaluator.messageCenter.Subscribe(ProcessTime)
	defer evaluator.messageCenter.Unsubscribe(processTimeChannel)

	csTimeChannel := evaluator.messageCenter.Subscribe(ClientToServerTime)
	defer evaluator.messageCenter.Unsubscribe(csTimeChannel)

	scTimeChannel := evaluator.messageCenter.Subscribe(ServerToClientTime)
	defer evaluator.messageCenter.Unsubscribe(scTimeChannel)

	for {
		select {
		case msg := <-frameChannel:
			frame, ok := msg.Content.(Frame)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			frameTime[frame.FrameID] = frame.Timestamp
		case msg := <-responseChannel:
			response, ok := msg.Content.(Response)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			delay := response.SendTime - frameTime[response.FrameID]
			evaluator.Delays = append(evaluator.Delays, delay)
			delete(frameTime, response.FrameID)
		case msg := <-trackingChannel:
			trackResult, ok := msg.Content.(TrackResult)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			delay := trackResult.DoneTime - frameTime[trackResult.FrameID]
			evaluator.Delays = append(evaluator.Delays, delay)
			delete(frameTime, trackResult.FrameID)
		case msg := <-encodeTimeChannel:
			time, ok := msg.Content.(int64)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			evaluator.EncodeTime = append(evaluator.EncodeTime, time)
		case msg := <-trackingTimeChannel:
			time, ok := msg.Content.(int64)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			evaluator.TrackingTime = append(evaluator.TrackingTime, time)
		case msg := <-processTimeChannel:
			time, ok := msg.Content.(int64)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			evaluator.ProcessTime = append(evaluator.ProcessTime, time)
		case msg := <- csTimeChannel:
			time, ok := msg.Content.(int64)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			evaluator.C2STime = append(evaluator.C2STime, time)
		case msg := <-scTimeChannel:
			time, ok := msg.Content.(int64)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			evaluator.S2CTime = append(evaluator.S2CTime, time)
		case <-ticker.C:
			// 定时显示目前的运行状况
			//
			showAverageTime(evaluator.Delays, "delay")
			showAverageTime(evaluator.EncodeTime, "encode time")
			showAverageTime(evaluator.TrackingTime, "tracking time")
			showAverageTime(evaluator.ProcessTime, "process time")
			showAverageTime(evaluator.C2STime, "client to server time")
			showAverageTime(evaluator.S2CTime, "server to client time")
			evaluator.Delays = make([]int64, 0)
			evaluator.EncodeTime = make([]int64, 0)
			evaluator.TrackingTime = make([]int64, 0)
			evaluator.ProcessTime = make([]int64, 0)
			evaluator.C2STime = make([]int64, 0)
			evaluator.S2CTime = make([]int64, 0)
		}
	}
}

func showAverageTime(his []int64, name string) {
	var total int64 = 0
	for _, time := range his {
		total += time
	}
	log.Infof("average %v: %v", name, float64(total)/float64(len(his)*1000000000))
}

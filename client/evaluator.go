package main

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type Evaluator struct {
	lastProcessFrameID int
	Delays []int64
	EncodeTime []int64
	messageCenter MessageCenter
}

func (evaluator *Evaluator) init(messageCenter MessageCenter)  {
	evaluator.messageCenter = messageCenter
}

func (evaluator *Evaluator) run()  {
	// 这里应该搞个参数来控制，先mark了，还没做参数的分离
	ticker := time.NewTicker(2 * time.Second)

	frameChannel := evaluator.messageCenter.Subscribe("Frame")
	defer evaluator.messageCenter.Unsubscribe(frameChannel)
	frameTime := make(map[int]int64)
	
	responseChannel := evaluator.messageCenter.Subscribe("Response")
	defer evaluator.messageCenter.Unsubscribe(responseChannel)
	for {
		select{
		case msg := <- frameChannel:
			frame, ok := msg.Content.(Frame)
			if !ok{
				log.Errorf("get wrong msg")
				return
			}
			frameTime[frame.FrameID] = frame.Timestamp
		case msg := <- responseChannel:
			response, ok := msg.Content.(Response)
			if !ok{
				log.Errorf("get wrong msg")
				return
			}
			delay := time.Now().UnixNano()-frameTime[response.FrameID]
			evaluator.Delays = append(evaluator.Delays, delay)
			delete(frameTime, response.FrameID)
		case <-ticker.C:
			var total int64 = 0
			for _, delay := range evaluator.Delays{
				total += delay
			}
			log.Infof("average delay: %v", float64(total) / float64(len(evaluator.Delays)*1000000000))
			evaluator.Delays = make([]int64, 0)
		}
	}
}
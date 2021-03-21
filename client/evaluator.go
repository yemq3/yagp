package main

import (
	"math"
	"time"

	log "github.com/sirupsen/logrus"
)

// Jacobson / Karels 算法超参，根据RFC6289的推荐值定义
const (
	ALPHA = 0.125
	BETA  = 0.25
	K     = 4
)

// Evaluator 计算性能数据
type Evaluator struct {
	messageCenter MessageCenter
	// 存储过去n秒的数据
	DelaysHistory       []int64 // 发送检测到收到结果的时间
	EncodeTimeHistory   []int64 // 编码时间
	TrackingTimeHistory []int64 // 跟踪时间
	ProcessTimeHistory  []int64 // 服务器端处理一帧的时间
	C2STimeHistory      []int64 // 客户端传输数据到服务器端用时
	S2CTimeHistory      []int64 // 服务器端传输数据到客户端用时
	// 以下数据均根据根据Jacobson / Karels 算法计算
	// 平滑时间，根据Jacobson / Karels 算法计算
	SmoothedDelays       float64
	SmoothedEncodeTime   float64
	SmoothedTrackingTime float64
	SmoothedProcessTime  float64
	SmoothedC2STime      float64
	SmoothedS2CTime      float64
	// 偏离时间
	DeviationDelays       float64
	DeviationEncodeTime   float64
	DeviationTrackingTime float64
	DeviationProcessTime  float64
	DeviationC2STime      float64
	DeviationS2CTime      float64
	// 触发警告的时间（和RTO的计算方式一样）
	WarningDelays       float64
	WarningEncodeTime   float64
	WarningTrackingTime float64
	WarningProcessTime  float64
	WarningC2STime      float64
	WarningS2CTime      float64
}

// NewEvaluator creates a new Evaluator
func NewEvaluator(messageCenter MessageCenter) Evaluator {
	evaluator := Evaluator{
		messageCenter: messageCenter,
	}
	return evaluator
}

func (evaluator *Evaluator) run() {
	log.Infof("Evaluator running...")
	// 这里应该搞个参数来控制，先mark了，还没做参数的分离
	ticker := time.NewTicker(1 * time.Second)

	isStart := true

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
			delay := response.GetTime - frameTime[response.FrameID]
			evaluator.DelaysHistory = append(evaluator.DelaysHistory, delay)
			if !isStart {
				evaluator.SmoothedDelays = (1-ALPHA)*evaluator.SmoothedDelays + ALPHA*float64(delay)
				evaluator.DeviationDelays = (1-BETA)*evaluator.DeviationDelays + BETA*float64(math.Abs(evaluator.SmoothedDelays-float64(delay)))
				evaluator.WarningDelays = evaluator.SmoothedDelays + K*evaluator.DeviationDelays
			}
			delete(frameTime, response.FrameID)
		case msg := <-trackingChannel:
			trackResult, ok := msg.Content.(TrackResult)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			delay := trackResult.DoneTime - frameTime[trackResult.FrameID]
			evaluator.DelaysHistory = append(evaluator.DelaysHistory, delay)
			if !isStart {
				evaluator.SmoothedDelays = (1-ALPHA)*evaluator.SmoothedDelays + ALPHA*float64(delay)
				evaluator.DeviationDelays = (1-BETA)*evaluator.DeviationDelays + BETA*float64(math.Abs(evaluator.SmoothedDelays-float64(delay)))
				evaluator.WarningDelays = evaluator.SmoothedDelays + K*evaluator.DeviationDelays
			}
			delete(frameTime, trackResult.FrameID)
		case msg := <-encodeTimeChannel:
			time, ok := msg.Content.(int64)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			evaluator.EncodeTimeHistory = append(evaluator.EncodeTimeHistory, time)
			if !isStart {
				evaluator.SmoothedEncodeTime = (1-ALPHA)*evaluator.SmoothedEncodeTime + ALPHA*float64(time)
				evaluator.DeviationEncodeTime = (1-BETA)*evaluator.DeviationEncodeTime + BETA*float64(math.Abs(evaluator.SmoothedEncodeTime-float64(time)))
				evaluator.WarningEncodeTime = evaluator.SmoothedEncodeTime + K*evaluator.DeviationEncodeTime
			}
		case msg := <-trackingTimeChannel:
			time, ok := msg.Content.(int64)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			evaluator.TrackingTimeHistory = append(evaluator.TrackingTimeHistory, time)
			if !isStart {
				evaluator.SmoothedTrackingTime = (1-ALPHA)*evaluator.SmoothedTrackingTime + ALPHA*float64(time)
				evaluator.DeviationTrackingTime = (1-BETA)*evaluator.DeviationTrackingTime + BETA*float64(math.Abs(evaluator.SmoothedTrackingTime-float64(time)))
				evaluator.WarningTrackingTime = evaluator.SmoothedTrackingTime + K*evaluator.DeviationTrackingTime
			}
		case msg := <-processTimeChannel:
			time, ok := msg.Content.(int64)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			evaluator.ProcessTimeHistory = append(evaluator.ProcessTimeHistory, time)
			if !isStart {
				evaluator.SmoothedProcessTime = (1-ALPHA)*evaluator.SmoothedProcessTime + ALPHA*float64(time)
				evaluator.DeviationProcessTime = (1-BETA)*evaluator.DeviationProcessTime + BETA*float64(math.Abs(evaluator.SmoothedProcessTime-float64(time)))
				evaluator.WarningProcessTime = evaluator.SmoothedProcessTime + K*evaluator.DeviationProcessTime
			}
		case msg := <-csTimeChannel:
			time, ok := msg.Content.(int64)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			evaluator.C2STimeHistory = append(evaluator.C2STimeHistory, time)
			if !isStart {
				evaluator.SmoothedC2STime = (1-ALPHA)*evaluator.SmoothedC2STime + ALPHA*float64(time)
				evaluator.DeviationC2STime = (1-BETA)*evaluator.DeviationC2STime + BETA*float64(math.Abs(evaluator.SmoothedC2STime-float64(time)))
				evaluator.WarningC2STime = evaluator.SmoothedC2STime + K*evaluator.DeviationC2STime
			}
		case msg := <-scTimeChannel:
			time, ok := msg.Content.(int64)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			evaluator.S2CTimeHistory = append(evaluator.S2CTimeHistory, time)
			if !isStart {
				evaluator.SmoothedS2CTime = (1-ALPHA)*evaluator.SmoothedS2CTime + ALPHA*float64(time)
				evaluator.DeviationS2CTime = (1-BETA)*evaluator.DeviationS2CTime + BETA*float64(math.Abs(evaluator.SmoothedS2CTime-float64(time)))
				evaluator.WarningS2CTime = evaluator.SmoothedS2CTime + K*evaluator.DeviationS2CTime
			}
		case <-ticker.C:
			// 定时显示目前的运行状况
			//
			log.Infof("--------------Running status of last 1 seconds--------------")
			delays := showAverageTime(evaluator.DelaysHistory, "delay")
			encodeTime := showAverageTime(evaluator.EncodeTimeHistory, "encode time")
			trackingTime := showAverageTime(evaluator.TrackingTimeHistory, "tracking time")
			processTime := showAverageTime(evaluator.ProcessTimeHistory, "process time")
			c2sTime := showAverageTime(evaluator.C2STimeHistory, "client to server time")
			s2cTime := showAverageTime(evaluator.S2CTimeHistory, "server to client time")
			fps := getFPS(encodeTime, trackingTime, processTime, c2sTime, s2cTime)
			log.Infof("FPS: %v", fps)
			log.Infof("--------------Running status of last 1 seconds--------------")
			log.Infof("--------------Current status of Time --------------")
			log.Infof("Smoothed Delays: %v, Deviation Delays: %v, Warning Delays: %v", evaluator.SmoothedDelays/1000000000, evaluator.DeviationDelays/1000000000, evaluator.WarningDelays/1000000000)
			log.Infof("Smoothed EncodeTime: %v, Deviation EncodeTime: %v, Warning EncodeTime: %v", evaluator.SmoothedEncodeTime/1000000000, evaluator.DeviationEncodeTime/1000000000, evaluator.WarningEncodeTime/1000000000)
			log.Infof("Smoothed TrackingTime: %v, Deviation TrackingTime: %v, Warning TrackingTime: %v", evaluator.SmoothedTrackingTime/1000000000, evaluator.DeviationTrackingTime/1000000000, evaluator.WarningTrackingTime/1000000000)
			log.Infof("Smoothed ProcessTime: %v, Deviation ProcessTime: %v, Warning ProcessTime: %v", evaluator.SmoothedProcessTime/1000000000, evaluator.DeviationProcessTime/1000000000, evaluator.WarningProcessTime/1000000000)
			log.Infof("Smoothed C2STime: %v, Deviation C2STime: %v, Warning C2STime: %v", evaluator.SmoothedC2STime/1000000000, evaluator.DeviationC2STime/1000000000, evaluator.WarningC2STime/1000000000)
			log.Infof("Smoothed S2CTime: %v, Deviation S2CTime: %v, Warning S2CTime: %v", evaluator.SmoothedS2CTime/1000000000, evaluator.DeviationS2CTime/1000000000, evaluator.WarningS2CTime/1000000000)
			log.Infof("--------------Current status of Time --------------")
			evaluator.DelaysHistory = make([]int64, 0)
			evaluator.EncodeTimeHistory = make([]int64, 0)
			evaluator.TrackingTimeHistory = make([]int64, 0)
			evaluator.ProcessTimeHistory = make([]int64, 0)
			evaluator.C2STimeHistory = make([]int64, 0)
			evaluator.S2CTimeHistory = make([]int64, 0)
			if isStart {
				evaluator.SmoothedDelays = delays
				evaluator.DeviationDelays = delays / 2
				evaluator.WarningDelays = evaluator.SmoothedDelays + K*evaluator.DeviationDelays

				evaluator.SmoothedEncodeTime = encodeTime
				evaluator.DeviationEncodeTime = encodeTime / 2
				evaluator.WarningEncodeTime = evaluator.SmoothedEncodeTime + K*evaluator.DeviationEncodeTime

				evaluator.SmoothedTrackingTime = trackingTime
				evaluator.DeviationTrackingTime = trackingTime / 2
				evaluator.WarningTrackingTime = evaluator.SmoothedTrackingTime + K*evaluator.DeviationTrackingTime

				evaluator.SmoothedProcessTime = processTime
				evaluator.DeviationProcessTime = processTime / 2
				evaluator.WarningProcessTime = evaluator.SmoothedProcessTime + K*evaluator.DeviationProcessTime

				evaluator.SmoothedC2STime = c2sTime
				evaluator.DeviationC2STime = c2sTime / 2
				evaluator.WarningC2STime = evaluator.SmoothedC2STime + K*evaluator.DeviationC2STime

				evaluator.SmoothedS2CTime = s2cTime
				evaluator.DeviationS2CTime = s2cTime / 2
				evaluator.WarningS2CTime = evaluator.SmoothedS2CTime + K*evaluator.DeviationS2CTime
				isStart = false
			}
		}
	}
}

func showAverageTime(his []int64, name string) float64 {
	var total int64 = 0
	for _, time := range his {
		total += time
	}
	log.Infof("average %v: %v", name, float64(total)/float64(len(his)*1000000000))
	return float64(total) / float64(len(his))
}

func getFPS(encodeTime, trackingTime, processTime, c2sTime, s2cTime float64) float64{
	trackFPS := 1 / trackingTime * 1000000000
	detectFPS := 1 / (encodeTime+processTime+c2sTime+s2cTime) * 1000000000

	return trackFPS + detectFPS
}

package main

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type Controller struct {
	messageCenter      MessageCenter
	evaluator          Evaluator   // 性能数据
	executor           Executor    // 其他模块的调用单元
	ControllerChannel  chan Frame  //
	interval           int         // 每隔interval帧做一次检测
	counter            int         //
	isNextFrameControl bool        //
	nextFrameMethod    int         //
	mu                 *sync.Mutex //
}

func NewController(messageCenter MessageCenter, encoderChannel chan EncodeTask, trackerChannel chan TrackTask, updateChannel chan UpdateTask, interval int) Controller {
	controller := Controller{
		messageCenter:     messageCenter,
		evaluator:         NewEvaluator(messageCenter),
		executor:          NewExecutor(encoderChannel, trackerChannel, updateChannel),
		ControllerChannel: make(chan Frame),
		interval:          interval,
	}

	return controller
}

func (controller *Controller) SetInterval(interval int) {
	controller.interval = interval
}

// 设置下一帧用的方法，如果设为DETECT，下一帧会丢去检测，然后继续按interval的值进行跟踪，设为TRACK会让检测延迟一帧
func (controller *Controller) SetNextFrameMethod(method int) {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if method == DETECT || method == TRACK {
		controller.isNextFrameControl = true
		controller.nextFrameMethod = method
	}
}

func (controller *Controller) run() {
	log.Infof("controller running...")
	// TODO: remove magic number

	detectChannel := controller.messageCenter.Subscribe(DetectResult)
	defer controller.messageCenter.Unsubscribe(detectChannel)

	go controller.evaluator.run()
	for {
		select {
		case frame := <-controller.ControllerChannel:
			if controller.counter == 0 {
				log.Debugf("Frameid: %v, go to encoder", frame.FrameID)
				controller.executor.sendEncodeTask(frame, 75, 0.5)
				controller.executor.sendTrackTask(frame)
			} else {
				log.Debugf("Frameid: %v, go to tracker", frame.FrameID)
				controller.executor.sendTrackTask(frame)
				// controller.executor.sendDropTask(frame)
			}
			if controller.counter == controller.interval {
				controller.counter = 0
			} else {
				controller.counter++
			}
		case msg := <-detectChannel:
			response, ok := msg.Content.(ResultWithAbsoluteBox)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			controller.executor.sendUpdateTask(response)
			// if controller.evaluator.LatestFrameID - response.FrameID > 4{

			// }
		}
	}
}

// 新的检测结果到达时，跟踪可能已经过了几帧，直接reinit跟踪器的话跟踪效果不一定好
// 可以再发布相应的跟踪任务，让其从新的检测结果跟踪到较新的

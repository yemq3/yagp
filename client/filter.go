package main

import (
	"gocv.io/x/gocv"
)

// 用于过滤帧的函数定义，返回true代表这帧会被过滤
type FilterFunc func(gocv.Mat) bool

type Filter struct {
	FilterFunc        FilterFunc
	FilterChannel     chan Frame
	controllerChannel chan Frame
	messageCenter         MessageCenter
}

func defaultFilterFunc(img gocv.Mat) bool {
	return false
}

func (filter *Filter) init(controllerChannel chan Frame, messageCenter MessageCenter) error {
	filter.FilterChannel = make(chan Frame)
	filter.FilterFunc = defaultFilterFunc
	filter.controllerChannel = controllerChannel
	filter.messageCenter = messageCenter

	return nil
}

func (filter *Filter) SetFilterFunc(filterFunc FilterFunc){
	filter.FilterFunc = filterFunc
}

func (filter *Filter) run() {
	for {
		select {
		case frame := <-filter.FilterChannel:
			// 要不要传给持久层我没想好，或许这里应该加个if判断是不是要传过去
			go func(frame Frame) {
				msg := Message{
					Topic:   "Frame",
					Content: frame,
				}
				filter.messageCenter.Publish(msg)
			}(frame)
			if filter.FilterFunc(frame.Frame) {
				// 过滤掉，不传给Controller
				continue
			}
			go func(frame Frame) {
				filter.controllerChannel <- frame
			}(frame)
		}
	}
}

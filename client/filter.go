package main

import (
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

// FilterFunc 过滤帧的函数定义，返回true代表这帧会被过滤
type FilterFunc func(gocv.Mat) bool

// Filter 过滤帧用
type Filter struct {
	FilterFunc        FilterFunc
	FilterChannel     chan Frame
	controllerChannel chan Frame
	messageCenter     MessageCenter
}

func defaultFilterFunc(img gocv.Mat) bool {
	return false
}

// NewFilter creates a new Filter
func NewFilter(controllerChannel chan Frame, messageCenter MessageCenter) Filter {
	filter := Filter{}

	filter.FilterChannel = make(chan Frame)
	filter.FilterFunc = defaultFilterFunc
	filter.controllerChannel = controllerChannel
	filter.messageCenter = messageCenter

	return filter
}

// SetFilterFunc 设置过滤函数
func (filter *Filter) SetFilterFunc(filterFunc FilterFunc) {
	filter.FilterFunc = filterFunc
}

func (filter *Filter) run() {
	log.Infof("Filter running...")
	for {
		select {
		case frame := <-filter.FilterChannel:
			log.Debugf("filter get frame, id: %v", frame.FrameID)
			// 要不要传给持久层我没想好，或许这里应该加个if判断是不是要传过去

			msg := Message{
				Topic:   FilterFrame,
				Content: frame,
			}
			filter.messageCenter.Publish(msg)

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

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
	persister         Persister
}

func defaultFilterFunc(img gocv.Mat) bool {
	return false
}

func (filter *Filter) init(filterFunc FilterFunc, controllerChannel chan Frame, persister Persister) error {
	filter.FilterChannel = make(chan Frame)
	filter.FilterFunc = filterFunc
	filter.controllerChannel = controllerChannel
	filter.persister = persister

	return nil
}

func (filter *Filter) run() {
	for {
		select {
		case frame := <-filter.FilterChannel:
			// 要不要传给持久层我没想好，或许这里应该加个if判断是不是要传过去
			go func(frame Frame) {
				filter.persister.persistFrame(frame)
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

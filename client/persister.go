package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

// Persister 保存结果
type Persister struct {
	resultDir     string
	messageCenter MessageCenter
}

// NewPersister creates a new Persister
func NewPersister(messageCenter MessageCenter, dir string) Persister {
	persister := Persister{}

	persister.messageCenter = messageCenter
	persister.resultDir = dir

	return persister
}

func (persister *Persister) run() {
	log.Infof("Persister running...")

	frameChannel := persister.messageCenter.Subscribe(FilterFrame)
	defer persister.messageCenter.Unsubscribe(frameChannel)

	detectChannel := persister.messageCenter.Subscribe(DetectResult)
	defer persister.messageCenter.Unsubscribe(detectChannel)

	trackChannel := persister.messageCenter.Subscribe(TrackResult)
	defer persister.messageCenter.Unsubscribe(trackChannel)

	err := os.Mkdir(persister.resultDir, 0755)
	if err != nil {
		log.Errorf("can't create dir")
	}

	fpsFile, err := os.Create(fmt.Sprintf("%v/fps.txt", persister.resultDir))
	if err != nil {
		log.Errorf("can't create dir")
	}
	defer fpsFile.Close()

	ticker := time.NewTicker(1 * time.Second)
	detectCounter := 0
	trackCounter := 0

	frameTime := make(map[int]int64)

	for {
		select {
		case msg := <-frameChannel:
			frame, ok := msg.Content.(Frame)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			frameTime[frame.FrameID] = frame.Timestamp
		case msg := <-detectChannel:
			response, ok := msg.Content.(ResultWithAbsoluteBox)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			detectCounter += 1

			f, err := os.Create(fmt.Sprintf("%v/%v.txt", persister.resultDir, response.FrameID))
			if err != nil {
				log.Errorf("can't create dir")
			}
			// Method & delay
			f.WriteString(fmt.Sprintf("detect %v\n", response.DoneTime-frameTime[response.FrameID]))
			for _, box := range response.Boxes {
				relativeBox := box.ToRelativeBox(WIDTH, HEIGHT)
				f.WriteString(fmt.Sprintf("%v %v %v %v %v %v\n", box.Name, box.Conf, relativeBox.X1, relativeBox.Y1, relativeBox.X2, relativeBox.Y2))
			}

			f.Close()
		case msg := <-trackChannel:
			trackResult, ok := msg.Content.(ResultWithAbsoluteBox)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			trackCounter += 1

			f, err := os.Create(fmt.Sprintf("%v/%v.txt", persister.resultDir, trackResult.FrameID))
			if err != nil {
				log.Errorf("can't create dir")
			}
			// Method & delay
			f.WriteString(fmt.Sprintf("track %v\n", trackResult.DoneTime-frameTime[trackResult.FrameID]))
			for _, box := range trackResult.Boxes {
				relativeBox := box.ToRelativeBox(WIDTH, HEIGHT)
				f.WriteString(fmt.Sprintf("%v %v %v %v %v %v\n", box.Name, box.Conf, relativeBox.X1, relativeBox.Y1, relativeBox.X2, relativeBox.Y2))
			}

			f.Close()
		case <-ticker.C:
			fpsFile.WriteString(fmt.Sprintf("%v %v %v\n", detectCounter, trackCounter, detectCounter+trackCounter))
			detectCounter = 0
			trackCounter = 0
		}

	}
}

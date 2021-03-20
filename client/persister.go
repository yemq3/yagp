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

	// frameChannel := persister.messageCenter.Subscribe(FilterFrame)
	// defer persister.messageCenter.Unsubscribe(frameChannel)

	responseChannel := persister.messageCenter.Subscribe(NetworkResponse)
	defer persister.messageCenter.Unsubscribe(responseChannel)

	trackChannel := persister.messageCenter.Subscribe(TrackerTrackResult)
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

	for {
		select {
		case msg := <-responseChannel:
			response, ok := msg.Content.(Response)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			detectCounter += 1

			f, err := os.Create(fmt.Sprintf("%v/%v.txt", persister.resultDir, response.FrameID))
			if err != nil {
				log.Errorf("can't create dir")
			}
			f.WriteString("detect\n")
			for _, box := range response.Boxes {
				f.WriteString(fmt.Sprintf("%v %v %v %v %v %v\n", box.Name, box.Conf, box.X1, box.Y1, box.X2, box.Y2))
			}

			f.Close()
		case msg := <-trackChannel:
			trackResult, ok := msg.Content.(TrackResult)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			trackCounter += 1

			f, err := os.Create(fmt.Sprintf("%v/%v.txt", persister.resultDir, trackResult.FrameID))
			if err != nil {
				log.Errorf("can't create dir")
			}
			f.WriteString("track\n")
			for _, box := range trackResult.Boxes {
				f.WriteString(fmt.Sprintf("%v %v %v %v %v %v\n", box.Name, box.Conf, box.X1, box.Y1, box.X2, box.Y2))
			}

			f.Close()
		case <-ticker.C:
			fpsFile.WriteString(fmt.Sprintf("%v %v %v\n", detectCounter, trackCounter, detectCounter+trackCounter))
			detectCounter = 0
			trackCounter = 0
		}

	}
}

package main

import log "github.com/sirupsen/logrus"

type TrackResult struct {
	FrameID int
}

type Tracker struct {
	TrackerChannel chan Frame
	persister Persister
}

func (tracker *Tracker) init(persister Persister) error {
	tracker.TrackerChannel = make(chan Frame)
	tracker.persister = persister

	return nil
}

func (tracker Tracker) run()  {
	for{
		select {
		case frame := <- tracker.TrackerChannel:
			log.Infof("Tracker get Frame %v", frame.FrameID)
			trackResult := TrackResult{}
			trackResult.FrameID = frame.FrameID
			tracker.persister.persistTrackResult(trackResult)
			continue
		}
	}
}

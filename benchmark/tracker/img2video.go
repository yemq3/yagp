package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

func main() {
	log.SetLevel(log.InfoLevel)
	writer, err := gocv.VideoWriterFile("video.mp4", "MJPG", 10, 1242, 375, true)
	if err != nil {
		return
	}
	defer writer.Close()

	window := gocv.NewWindow("test")
	defer window.Close()

	for i := 0; i <= 153; i++ {
		file := fmt.Sprintf("./0000/0000/%06d.png", i)
		img := gocv.IMRead(file, gocv.IMReadColor)
		log.Infof("%v", img.Size())

		window.IMShow(img)
		window.WaitKey(100)

		writer.Write(img)
	}
}

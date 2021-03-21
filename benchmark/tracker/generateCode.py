size = ["10x10", "20x20", "50x50", "100x100", "200x200"]
method = ["KCF", "CSRT", "MIL", "MOSSE", "Boosting", "MedianFlow", "TLD"]

common = """
package tracker

import (
	"image"
	"testing"

	log "github.com/sirupsen/logrus"

	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"
)

func Init(rectangle image.Rectangle, trackingAlgorithm string) (contrib.Tracker, gocv.Mat){
	one := gocv.NewMat()
	two := gocv.NewMat()

	c, err := gocv.VideoCaptureFile("../video.mp4")
	if err != nil {
		log.Errorln(err)
	}
	defer c.Close()
	c.Read(&one)

	var t contrib.Tracker
	if trackingAlgorithm == "KCF" {
		t = contrib.NewTrackerKCF()
	} else if trackingAlgorithm == "CSRT" {
		t = contrib.NewTrackerCSRT()
	} else if trackingAlgorithm == "MIL" {
		t = contrib.NewTrackerMIL()
	} else if trackingAlgorithm == "MOSSE" {
		t = contrib.NewTrackerMOSSE()
	} else if trackingAlgorithm == "Boosting" {
		t = contrib.NewTrackerBoosting()
	} else if trackingAlgorithm == "MedianFlow" {
		t = contrib.NewTrackerMedianFlow()
	} else {
		t = contrib.NewTrackerTLD()
	}
	t.Init(one, rectangle)

	c.Read(&two)
	return t, two
}
"""

code_template = """
func Benchmark{}_{}(b *testing.B) {{
	rectangle := image.Rect(10, 10, {}, {})
	t, two := Init(rectangle, "{}")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {{
		t.Update(two)
	}}
}}
"""

with open("tracker_test.go", "w") as f:
    f.write(common)
    for m in method:
        for s in size:
            temp = code_template.format(m, s, 10 + int(s.split("x")[0]), 10 + int(s.split("x")[0]), m)
            f.write(temp)
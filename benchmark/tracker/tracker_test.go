package tracker

import (
	"image"
	"testing"

	log "github.com/sirupsen/logrus"

	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"
)

func Init(rectangle image.Rectangle, trackingAlgorithm string) (contrib.Tracker, gocv.Mat) {
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

func BenchmarkKCF_10x10(b *testing.B) {
	rectangle := image.Rect(10, 10, 20, 20)
	t, two := Init(rectangle, "KCF")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkKCF_20x20(b *testing.B) {
	rectangle := image.Rect(10, 10, 30, 30)
	t, two := Init(rectangle, "KCF")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkKCF_50x50(b *testing.B) {
	rectangle := image.Rect(10, 10, 60, 60)
	t, two := Init(rectangle, "KCF")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkKCF_100x100(b *testing.B) {
	rectangle := image.Rect(10, 10, 110, 110)
	t, two := Init(rectangle, "KCF")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkCSRT_10x10(b *testing.B) {
	rectangle := image.Rect(10, 10, 20, 20)
	t, two := Init(rectangle, "CSRT")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkCSRT_20x20(b *testing.B) {
	rectangle := image.Rect(10, 10, 30, 30)
	t, two := Init(rectangle, "CSRT")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkCSRT_50x50(b *testing.B) {
	rectangle := image.Rect(10, 10, 60, 60)
	t, two := Init(rectangle, "CSRT")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkCSRT_100x100(b *testing.B) {
	rectangle := image.Rect(10, 10, 110, 110)
	t, two := Init(rectangle, "CSRT")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkMIL_10x10(b *testing.B) {
	rectangle := image.Rect(10, 10, 20, 20)
	t, two := Init(rectangle, "MIL")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkMIL_20x20(b *testing.B) {
	rectangle := image.Rect(10, 10, 30, 30)
	t, two := Init(rectangle, "MIL")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkMIL_50x50(b *testing.B) {
	rectangle := image.Rect(10, 10, 60, 60)
	t, two := Init(rectangle, "MIL")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkMIL_100x100(b *testing.B) {
	rectangle := image.Rect(10, 10, 110, 110)
	t, two := Init(rectangle, "MIL")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkMOSSE_10x10(b *testing.B) {
	rectangle := image.Rect(10, 10, 20, 20)
	t, two := Init(rectangle, "MOSSE")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkMOSSE_20x20(b *testing.B) {
	rectangle := image.Rect(10, 10, 30, 30)
	t, two := Init(rectangle, "MOSSE")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkMOSSE_50x50(b *testing.B) {
	rectangle := image.Rect(10, 10, 60, 60)
	t, two := Init(rectangle, "MOSSE")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkMOSSE_100x100(b *testing.B) {
	rectangle := image.Rect(10, 10, 110, 110)
	t, two := Init(rectangle, "MOSSE")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkBoosting_10x10(b *testing.B) {
	rectangle := image.Rect(10, 10, 20, 20)
	t, two := Init(rectangle, "Boosting")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkBoosting_20x20(b *testing.B) {
	rectangle := image.Rect(10, 10, 30, 30)
	t, two := Init(rectangle, "Boosting")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkBoosting_50x50(b *testing.B) {
	rectangle := image.Rect(10, 10, 60, 60)
	t, two := Init(rectangle, "Boosting")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkBoosting_100x100(b *testing.B) {
	rectangle := image.Rect(10, 10, 110, 110)
	t, two := Init(rectangle, "Boosting")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkMedianFlow_10x10(b *testing.B) {
	rectangle := image.Rect(10, 10, 20, 20)
	t, two := Init(rectangle, "MedianFlow")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkMedianFlow_20x20(b *testing.B) {
	rectangle := image.Rect(10, 10, 30, 30)
	t, two := Init(rectangle, "MedianFlow")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkMedianFlow_50x50(b *testing.B) {
	rectangle := image.Rect(10, 10, 60, 60)
	t, two := Init(rectangle, "MedianFlow")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkMedianFlow_100x100(b *testing.B) {
	rectangle := image.Rect(10, 10, 110, 110)
	t, two := Init(rectangle, "MedianFlow")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkTLD_10x10(b *testing.B) {
	rectangle := image.Rect(10, 10, 20, 20)
	t, two := Init(rectangle, "TLD")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkTLD_20x20(b *testing.B) {
	rectangle := image.Rect(10, 10, 30, 30)
	t, two := Init(rectangle, "TLD")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkTLD_50x50(b *testing.B) {
	rectangle := image.Rect(10, 10, 60, 60)
	t, two := Init(rectangle, "TLD")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

func BenchmarkTLD_100x100(b *testing.B) {
	rectangle := image.Rect(10, 10, 110, 110)
	t, two := Init(rectangle, "TLD")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(two)
	}
}

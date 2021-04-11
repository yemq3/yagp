package lk

import (
	"image"

	"github.com/yemq3/yagp/box"
	"gocv.io/x/gocv"
)

// LKTracker LK光流法跟踪
type LKTracker struct {
	prevImg    gocv.Mat
	prevPoints gocv.Mat
	boxes      []box.AbsoluteBox
	maxCorners int
	quality    float64
	minDist    float64
}

type pointPairs struct {
	Point1 []image.Point
	Point2 []image.Point
}

func NewLKTracker(maxCorners int, quality float64, minDist float64) *LKTracker {
	// 这里的参数都是给GoodFeaturesToTrack的
	// maxCorners: 最多生成多少个点，0代表不限制
	// quality: 用来筛选点，比如全部点的最高得分为1500，quality为0.01，则低于15分的会被淘汰
	// minDist: 如果minDist内有更好的点，那么直接淘汰
	t := &LKTracker{
		maxCorners: maxCorners,
		quality:    quality,
		minDist:    minDist,
	}
	return t
}

func (t *LKTracker) Init(image gocv.Mat, boundingBoxs []box.AbsoluteBox) {
	grayImg := gocv.NewMat()
	gocv.CvtColor(image, &grayImg, gocv.ColorBGRToGray)
	corners := gocv.NewMat()
	gocv.GoodFeaturesToTrack(grayImg, &corners, t.maxCorners, t.quality, t.minDist)
	t.prevImg = grayImg
	t.boxes = boundingBoxs
	t.prevPoints = corners
}

func getMeanShift(point1 []image.Point, point2 []image.Point) (x, y int) {
	if len(point1) == 0 {
		return 0, 0
	}
	totalX := 0
	totalY := 0
	for i := 0; i < len(point1); i++ {
		totalX += point2[i].X - point1[i].X
		totalY += point2[i].Y - point1[i].Y
	}
	meanX := float64(totalX) / float64(len(point1))
	meanY := float64(totalY) / float64(len(point1))

	return int(meanX), int(meanY)
}

func isInRectangle(x int, y int, rect image.Rectangle) bool {
	if rect.Min.X <= x && x <= rect.Max.X && rect.Min.Y <= y && y <= rect.Max.Y {
		return true
	}
	return false
}

func (t *LKTracker) Update(nextImg gocv.Mat) []box.AbsoluteBox {
	nextGrayImg := gocv.NewMat()
	gocv.CvtColor(nextImg, &nextGrayImg, gocv.ColorBGRToGray)
	nextPoints := gocv.NewMat()
	gocv.GoodFeaturesToTrack(nextGrayImg, &nextPoints, t.maxCorners, t.quality, t.minDist)

	status := gocv.NewMat() // 记录点是否被跟踪到
	flowErr := gocv.NewMat()
	defer status.Close()
	defer flowErr.Close()

	gocv.CalcOpticalFlowPyrLK(t.prevImg, nextGrayImg, t.prevPoints, nextPoints, &status, &flowErr)

	bytesStatus := status.ToBytes()
	pointInBox := make(map[box.AbsoluteBox]*pointPairs)
	for _, box := range t.boxes {
		pointInBox[box] = &pointPairs{
			Point1: make([]image.Point, 0),
			Point2: make([]image.Point, 0),
		}
	}
	for i := 0; i < len(bytesStatus); i++ {
		if bytesStatus[i] == 1 {
			point := t.prevPoints.GetVecfAt(i, 0)
			point1 := image.Point{int(point[0]), int(point[1])}
			point = nextPoints.GetVecfAt(i, 0)
			point2 := image.Point{int(point[0]), int(point[1])}
			for _, box := range t.boxes {
				if point1.In(box.Rect) {
					pointInBox[box].Point1 = append(pointInBox[box].Point1, point1)
					pointInBox[box].Point2 = append(pointInBox[box].Point2, point2)
				}
			}
		}
	}

	newBoxes := make([]box.AbsoluteBox, 0)
	for b, pointPair := range pointInBox {
		shiftX, shiftY := getMeanShift(pointPair.Point1, pointPair.Point2)
		shiftVector := image.Point{shiftX, shiftY}
		newRect := b.Rect.Add(shiftVector)
		newBoxes = append(newBoxes, box.AbsoluteBox{
			Rect: newRect,
			UUID: b.UUID,
			Conf: b.Conf,
			Name: b.Name,
		})
	}

	return newBoxes
}

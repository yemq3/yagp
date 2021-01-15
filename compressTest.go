package main

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"gocv.io/x/gocv"
	// "math/rand"
)

func main() {
	c, _ := gocv.VideoCaptureDevice(0)
	c.Set(gocv.VideoCaptureFrameWidth, 800)
	c.Set(gocv.VideoCaptureFrameHeight, 600)
	c.Set(gocv.VideoCaptureFPS, 24)
	defer c.Close()
	img := gocv.NewMat()
	defer img.Close()
	c.Read(&img)
	str, _ := gocv.IMEncodeWithParams(gocv.JPEGFileExt, img, []int{gocv.IMWriteJpegQuality, 60})
	var b bytes.Buffer
	// str := make([]byte, 100000)
	// rand.Read(str)
	// str := []byte("hello worldaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n")
	fmt.Println(len(str))

	w := zlib.NewWriter(&b)
	w.Write(str)
	w.Close()
	fmt.Println(len(b.Bytes()))

	b.Reset()
	zw := gzip.NewWriter(&b)
	zw.Write(str)
	zw.Close()
	fmt.Println(len(b.Bytes()))

}
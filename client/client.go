package main

import (
	"context"
	// "encoding/json"
	"flag"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

// Reply ...
type Reply struct {
	Boxes    []Box
	SendTime float64
}

type Box struct {
	X1      float64
	Y1      float64
	X2      float64
	Y2      float64
	Unknown float64
	Conf    float64
	name    string
}

type Frame struct {
	FrameID int
	Frame   []byte
}

type Camera struct {
	camera        *gocv.VideoCapture
	encodeChannel chan gocv.Mat
	sendChannel   chan []byte
	history       map[int]gocv.Mat // 存之前的图像
	encodeQuality int
}

var addr = flag.String("addr", "127.0.0.1:8000", "http service address")
var currFrameID = 0

// 中断不了这程序= =，只能重开terminal
func interruptHandler(cancel context.CancelFunc, ws *websocket.Conn) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	for {
		select {
		case <-interrupt:
			log.Infoln("interrupt")

			cancel()

			err := ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Errorln(err)
				return
			}
			os.Exit(1)
			return
		}
	}
}

func replyHandler(ctx context.Context, ws *websocket.Conn) {
	window2 := gocv.NewWindow("reply")
	defer window2.Close()
	for {
		reply := Reply{}
		err := ws.ReadJSON(&reply)
		if err != nil {
			log.Errorln(err)
			return
		}
		log.Infof("recv: %v", reply)
		// gocv.Rectangle()
	}
}

func (camera *Camera) send(ctx context.Context, ws *websocket.Conn) {
	for {
		select {
		case img := <-camera.sendChannel:
			log.Debugf("get a new image to send")
			frame := Frame{}
			frame.FrameID = currFrameID
			currFrameID++
			frame.Frame = img

			// b, err := json.Marshal(frame)
			// if err != nil {
			// 	log.Errorln(err)
			// 	return
			// }
			// err := ws.WriteJSON(frame)
			err := ws.WriteMessage(websocket.BinaryMessage, img)
			log.Infoln(img)
			if err != nil {
				log.Errorln(err)
				return
			}
			return
		case <-ctx.Done():
			log.Infoln("send interrupt")
			return
		}
	}
}

func (camera *Camera) encode(ctx context.Context) {
	for {
		select {
		case img := <-camera.encodeChannel:
			log.Debugf("get a new image to encode")
			buffer, err := gocv.IMEncodeWithParams(gocv.JPEGFileExt, img, []int{gocv.IMWriteJpegQuality, camera.encodeQuality})
			if err != nil {
				log.Errorln(err)
				return
			}
			go func(buffer []byte) {
				camera.sendChannel <- buffer
			}(buffer)
		case <-ctx.Done():
			log.Infoln("encode interrupt")
			return
		}
	}
}

func main() {
	log.SetLevel(log.InfoLevel)
	flag.Parse()

	// websocket连接
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Infof("connecting to %s", u.String())

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Errorln(err)
		return
	}
	defer ws.Close()
	// websocket连接

	// 初始化摄像头以及显示所需的东西
	c, err := gocv.VideoCaptureDevice(0)
	if err != nil {
		log.Errorln(err)
		return
	}
	c.Set(gocv.VideoCaptureFrameWidth, 800)
	c.Set(gocv.VideoCaptureFrameHeight, 600)
	c.Set(gocv.VideoCaptureFPS, 24)
	defer c.Close()

	window := gocv.NewWindow("camera")
	defer window.Close()

	img := gocv.NewMat()
	defer img.Close()
	// 初始化摄像头以及显示所需的东西

	camera := Camera{}
	camera.camera = c
	camera.encodeChannel = make(chan gocv.Mat)
	camera.sendChannel = make(chan []byte)
	camera.encodeQuality = 75

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go camera.encode(ctx)
	go camera.send(ctx, ws)
	go replyHandler(ctx, ws)
	go interruptHandler(cancel, ws)

	for {
		camera.camera.Read(&img)
		go func(img gocv.Mat) {
			camera.encodeChannel <- img
		}(img)
		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}

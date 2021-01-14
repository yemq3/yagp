package main

import (
	"bytes"
	"compress/gzip"
	"context"
	// "encoding/json"
	"image"
	"image/color"

	// "compress/gzip"
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
	FrameID  int
	Boxes    []Box
	SendTime float64
}

// Box ...
type Box struct {
	X1      float64
	Y1      float64
	X2      float64
	Y2      float64
	Unknown float64
	Conf    float64
	name    string
}

// Frame ...
type Frame struct {
	FrameID int
	Frame   []byte
}

// Camera ...
type Camera struct {
	camera         *gocv.VideoCapture
	encodeChannel  chan gocv.Mat
	sendChannel    chan []byte
	history        map[int]gocv.Mat // 存之前的图像
	encodeQuality  int
	latestFrameID  int
	newFrameNotify chan int

	weight    int
	height    int
	frameRate int
}

var addr = flag.String("addr", "127.0.0.1:8000", "http service address")
var currFrameID = 0

// 有时候会中断不了这程序= =，只能重开terminal
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

func (camera *Camera) replyHandler(ctx context.Context, ws *websocket.Conn) {
	window := gocv.NewWindow("reply")
	defer window.Close()
	red := color.RGBA{255, 0, 0, 0}
	for {
		select {
		case <-ctx.Done():
			return
		default:
			reply := Reply{}
			err := ws.ReadJSON(&reply)
			if err != nil {
				log.Errorln(err)
				return
			}
			log.Infof("recv: %v", reply)
			img := camera.history[reply.FrameID]
			height := img.Size()[0]
			weight := img.Size()[1]
			for _, box := range reply.Boxes {
				rect := image.Rectangle{}
				rect.Min = image.Point{int(box.X1 * float64(weight)), int(box.Y1 * float64(height))}
				rect.Max = image.Point{int(box.X2 * float64(weight)), int(box.Y2 * float64(height))}
				gocv.Rectangle(&img, rect, red, 3)
			}
			// 这个window的size不对，不知道怎么搞的= =
			window.IMShow(img)
			window.WaitKey(10)
		}
	}
}

func (camera *Camera) display(ctx context.Context) {
	window := gocv.NewWindow("camera")
	defer window.Close()

	for {
		select {
		case frameID := <-camera.newFrameNotify:
			img := camera.history[frameID]
			window.IMShow(img)
			window.WaitKey(10)
		case <-ctx.Done():
			log.Infoln("display interrupt")
			return
		}
	}
}

func gzipCompress(b []byte) ([]byte, error){
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(b); err != nil {
		log.Errorln(err)
		return []byte{}, err
	}
	zw.Flush()
	return buf.Bytes(), nil
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

			err := ws.WriteJSON(frame)
			// err = ws.WriteMessage(websocket.BinaryMessage, img)
			if err != nil {
				log.Errorln(err)
				return
			}
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

func (camera *Camera) getFrame(ctx context.Context) {
	img := gocv.NewMat()
	defer img.Close()
	for {
		select {
		case <-ctx.Done():
			log.Infoln("getFrame interrupt")
			return
		default:
			camera.camera.Read(&img)
			camera.latestFrameID++
			camera.history[camera.latestFrameID] = img
			go func(img gocv.Mat) {
				camera.encodeChannel <- img
			}(img)
			go func(frameID int) {
				camera.newFrameNotify <- frameID
			}(camera.latestFrameID)
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
	// 初始化摄像头以及显示所需的东西

	camera := Camera{}
	camera.camera = c
	camera.encodeChannel = make(chan gocv.Mat)
	camera.sendChannel = make(chan []byte)
	camera.newFrameNotify = make(chan int)
	camera.history = make(map[int]gocv.Mat)
	camera.encodeQuality = 75
	camera.latestFrameID = -1

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go camera.getFrame(ctx)
	go camera.encode(ctx)
	go camera.send(ctx, ws)
	go camera.display(ctx)
	go camera.replyHandler(ctx, ws)
	go interruptHandler(cancel, ws)

	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}

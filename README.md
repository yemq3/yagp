# Yet Another Graduation Project

不知道怎么起名了                  

```
.
├── README.md
├── client                                  client端代码
│   ├── camera.go
│   ├── client.go
│   ├── common.go
│   ├── controller.go
│   ├── encoder.go
│   ├── evaluator.go
│   ├── executor.go
│   ├── filter.go
│   ├── go.mod
│   ├── go.sum
│   ├── message_center.go
│   ├── network.go
│   ├── persister.go
│   ├── scheduler.go
│   └── tracker.go
│   ├── utils.go
└── sanic                                   server端代码
    ├── app
    │   └── yolov4                          https://github.com/Tianxiaomo/pytorch-YOLOv4修改得来
    │       ├── __init__.py                 可以在这修改是否使用GPU，使用完整模型还是tiny
    │       ├── cfg
    │       ├── data
    │       ├── tool
    │       └── weight
    │           ├── download.sh             这里下载权重之后应该就能跑了
    └── main.py
└── simulation server                       模拟服务器
    ├── app
    │   └── yolov4                          https://github.com/Tianxiaomo/pytorch-YOLOv4修改得来
    │       ├── __init__.py                 可以在这修改是否使用GPU，使用完整模型还是tiny
    │       ├── cfg
    │       ├── data
    │       ├── tool
    │       └── weight
    │           ├── download.sh             这里下载权重之后应该就能跑了
    ├── result                              存放缓存的结果
    └── main.py
    └── video2result.py                     将视频的结果缓存下来
```

![framework](pic/framework.jpg)

## Prerequisites

- [Go](https://golang.org/)：建议使用1.15+版本
- [GoCV](https://gocv.io/)：0.25（0.26版本删除了部分tracker）
- [python](https://www.python.org/)：3.7+
- [pytorch](https://pytorch.org/)：应该1.0版本以上就行

## Getting Started

```
git clone https://github.com/yemq3/yagp.git
cd yagp
```

### Server

#### Download weights

如果你在linux系统你可以直接

```
cd sanic/app/yolov4/weight/
sudo chmod +x download.sh
./download.sh
```

如果你在windows系统，你可以去下面两个链接下完之后丢进weight里面

[yolov4.weights](https://github.com/AlexeyAB/darknet/releases/download/darknet_yolo_v3_optimal/yolov4.weights)

[yolov4-tiny.weights](https://github.com/AlexeyAB/darknet/releases/download/darknet_yolo_v4_pre/yolov4-tiny.weights)

#### Run

```
cd sanic
pip install sanic
python main.py
```

### Client

先按照[GoCV](https://gocv.io/getting-started/)的指引安装好gocv

#### Run

```
cd client
go run .
```

## 一些开发指引

- 如果你需要Frame，Detection Result这些数据，可以给自己的模块加个messageCenter对象，然后调用messageCenter的Subscribe方法拿到通知所用的channel（可以参考其他模块比如Persister），不知道Topic是什么的话，去对应的模块找一下调用Publish方法的地方就知道了
- 遇到问题问下我就懂了（逃

## TODO

- [ ] 更细致的日志信息
- [ ] 对网络情况的自适应
- [ ] 对检测结果是否可信的判断
- [ ] benchmark
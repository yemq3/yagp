# Yet Another Graduation Project

不知道怎么起名了

```
.
├── README.md
├── client                              client端代码
│   ├── client.go
│   ├── go.mod
│   └── go.sum
└── sanic                               server端代码
    ├── app                    	        
    │   └── yolov4                      https://github.com/Tianxiaomo/pytorch-YOLOv4修改得来
    │       ├── __init__.py             可以在这修改是否使用GPU，使用完整模型还是tiny
    │       ├── cfg
    │       ├── data
    │       ├── tool
    │       └── weight
    │           ├── download.sh         这里下载权重之后应该就能跑了
    ├── main.py                         
    └── test                            以前试图拿python做进程池+协程的失败尝试，别理  
        ├── 1.py
        └── ws.py
```

## Prerequisites

- [Go](https://golang.org/)
- [GoCV](https://gocv.io/)
- [python](https://www.python.org/)：3.6+
- [pytorch](https://pytorch.org/)：应该1.0版本以上就行

## Getting Started

```
git clone https://github.com/yemq3/yagp.git
cd yagp
```

### Server

```
cd sanic
pip install sanic
python main.py
```

### Client

```
cd client
go run client.go
```

## TODO

- [ ] 历史帧到达一定数量后自动删除
- [ ] gzip压缩
- [ ] 更细致的日志信息
- [ ] 对网络情况的自适应
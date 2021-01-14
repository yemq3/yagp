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
    ├── 1.py                            之前的代码，别理
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
pip install -r requirement.txt
python main.py
```

### Client

```
cd client
go run client.go
```


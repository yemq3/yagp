from app.yolov4.tool.utils import *
from app.yolov4.tool.torch_utils import *
from app.yolov4.tool.darknet2pytorch import Darknet

def NewDarknet(use_tiny=False, use_cuda=True):
    if not use_tiny:
        # yolov4
        darknet = Darknet("./app/yolov4/cfg/yolov4.cfg")
        darknet.load_weights("./app/yolov4/weight/yolov4.weights")
    else:
        # yolov4-tiny
        darknet = Darknet("./app/yolov4/cfg/yolov4-tiny.cfg")
        darknet.load_weights("./app/yolov4/weight/yolov4-tiny.weights")

    if use_cuda:
        darknet.cuda()

    darknet.print_network()

    num_classes = darknet.num_classes
    if num_classes == 20:
        namesfile = './app/yolov4/data/voc.names'
    elif num_classes == 80:
        namesfile = './app/yolov4/data/coco.names'
    else:
        namesfile = './app/yolov4/data/x.names'
    class_names = load_class_names(namesfile)

    return darknet, class_names

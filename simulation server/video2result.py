from app.yolov4 import *
import cv2

USE_TINY = False
USE_CUDA = True

darknet, class_names = NewDarknet(USE_TINY, USE_CUDA)
print(class_names)
for i in range(len(class_names)):
    class_names[i] = class_names[i].replace(" ", "_")

cap = cv2.VideoCapture("./video.mp4")

frameid = 1

while True:
    ret, img = cap.read()
    if not ret:
        break
    height, width = img.shape[:2]
    sized = cv2.resize(img, (darknet.width, darknet.height))
    sized = cv2.cvtColor(sized, cv2.COLOR_BGR2RGB)

    boxes = do_detect(darknet, sized, 0.4, 0.6, USE_CUDA)
    print(boxes)

    with open("./result/result_xywh/{}.txt".format(frameid), "w") as f:
        for box in boxes[0]:
            x = int(box[0]*width)
            y = int(box[1]*height)
            w = int(box[2]*width)-x
            h = int(box[3]*height)-y
            #  class confidence x y w h
            f.write("{} {} {} {} {} {}\n".format(class_names[int(box[6])], box[5], x, y, w, h))

    with open("./result/result_xyxy/{}.txt".format(frameid), "w") as f:
        for box in boxes[0]:
            #  class confidence x y x y
            f.write("{} {} {} {} {} {}\n".format(class_names[int(box[6])], box[5], box[0], box[1], box[2], box[3]))
    frameid += 1

    result_img = plot_boxes_cv2(img, boxes[0], savename=None, class_names=class_names)

    cv2.imshow('Yolo demo', result_img)
    cv2.waitKey(1)

cap.release()

import websockets
import time
import cv2
import numpy as np
import asyncio
import pickle
import queue
from concurrent.futures import ThreadPoolExecutor, ProcessPoolExecutor

FRAME_RATE = 13
url = "ws://127.0.0.1/sanic"

cap = cv2.VideoCapture(0)
assert cap.isOpened()

cap.set(cv2.CAP_PROP_FRAME_WIDTH, 640)
cap.set(cv2.CAP_PROP_FRAME_HEIGHT, 480)
cap.set(cv2.CAP_PROP_FPS, FRAME_RATE)

def plot_boxes_cv2(img, boxes):
    img = np.copy(img)

    width = img.shape[1]
    height = img.shape[0]
    for i in range(len(boxes)):
        box = boxes[i]
        x1 = int(box[0] * width)
        y1 = int(box[1] * height)
        x2 = int(box[2] * width)
        y2 = int(box[3] * height)


        rgb = (255, 0, 0)

        if len(box) == 7:
            cls_conf = box[5]
            cls_name = box[6]
            # print('%s: %f' % (cls_name, cls_conf))
            img = cv2.putText(img, cls_name, (x1, y1), cv2.FONT_HERSHEY_SIMPLEX, 1.2, rgb, 1)
        img = cv2.rectangle(img, (x1, y1), (x2, y2), rgb, 1)

    return img

async def encodeWorker(sendQueue, ProcessQueue):
    frameid = 0
    while True:
        rval, image = cap.read()
        if not rval:
            break

        _, img_encoded = cv2.imencode('.jpg', image, [cv2.IMWRITE_JPEG_QUALITY, 50])

        data = {
            "image": img_encoded.tobytes(),
            "frameid": frameid,
        }

        frameid += 1

        sendQueue.put_nowait(data)
        ProcessQueue.put_nowait((image, time.time()))

        print("encode")

        await asyncio.sleep(1./FRAME_RATE)

async def sendWorker(ws, sendQueue):
    while True:
        data = await sendQueue.get()

        data["send_time"] = time.time()

        send_data = pickle.dumps(data)

        await ws.send(send_data)


async def processWorker(ws, ProcessQueue):
    while True:
        response = await ws.recv()

        response_data = pickle.loads(response)

        boxes = response_data["boxes"]

        image, encode_time = ProcessQueue.get_nowait()

        delay = time.time()-encode_time
        print("delay:", delay)

        result_img = plot_boxes_cv2(image, boxes)

        cv2.imshow('res', result_img)
        cv2.waitKey(1)


# async def main(ws):
#     SendQueue = asyncio.Queue()
#     ProcessQueue = asyncio.Queue()

#     task1 = asyncio.create_task(encodeWorker(SendQueue, ProcessQueue))
#     task2 = asyncio.create_task(sendWorker(ws, SendQueue, ProcessQueue))
#     task3 = asyncio.create_task(processWorker(ws, ProcessQueue))

#     await task1
#     await task2
#     await task3


async def connect():
    ws = await websockets.connect(url)
    return ws


if __name__ == "__main__":
    loop = asyncio.get_event_loop()
    ws = loop.run_until_complete(connect())

    DataQueue = asyncio.Queue()
    ProcessQueue = asyncio.Queue()

    task1 = loop.create_task(encodeWorker(DataQueue, ProcessQueue))
    task2 = loop.create_task(sendWorker(ws, DataQueue))
    task3 = loop.create_task(processWorker(ws, ProcessQueue))

    # asyncio.run(main(ws))
    loop.run_forever()
    


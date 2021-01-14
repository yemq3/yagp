from sanic import Sanic
from sanic.response import json
from sanic.websocket import WebSocketProtocol
from app.yolov4 import *
import cv2
import numpy as np
import time
import json
import base64

app = Sanic("server")

def readb64(data):
   nparr = np.fromstring(base64.b64decode(data), np.uint8)
   img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
   return img

@app.websocket("/")
async def test(request, ws):
    while True:
        data = await ws.recv()
        print(len(data))

        frame = json.loads(data)
        # print(frame)

        # print(frame['Frame'])

        # nparr = np.frombuffer(frame['Frame'].encode(), np.uint8)
        # print(nparr[-1])
        # nparr = np.frombuffer(data, np.uint8)
        # print(nparr[-1])
        # print(nparr.shape)
        # img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
        img = readb64(frame['Frame'])
        # print(img)

        sized = cv2.resize(img, (darknet.width, darknet.height))
        sized = cv2.cvtColor(sized, cv2.COLOR_BGR2RGB)
        
        boxes = do_detect(darknet, sized, 0.4, 0.6, USE_CUDA)
        boxes = np.array(boxes[0]).tolist()

        for i in range(len(boxes)):
            boxes[i][6] = class_names[int(boxes[i][6])]

        formatBoxes = []
        for i in range(len(boxes)):
            formatBoxes.append({
                'X1': boxes[i][0],
                'Y1': boxes[i][1],
                'X2': boxes[i][2],
                'Y2': boxes[i][3],
                'Unknown': boxes[i][4],
                'Conf': boxes[i][5],
                'Name': boxes[i][6],
            })
        
        response = {
            'FrameID': frame["FrameID"],
            'Boxes': formatBoxes,
            'SendTime': time.time()
        }

        await ws.send(json.dumps(response))

if __name__ == "__main__":
    app.run(protocol=WebSocketProtocol)
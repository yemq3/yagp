from sanic import Sanic
from sanic.response import json
from sanic.websocket import WebSocketProtocol
# from sanic_gzip import Compress
from app.yolov4 import *
import cv2
import numpy as np
import time
import json
import base64
import logging

app = Sanic("server")
# compress = Compress()

USE_CUDA = False

darknet, class_names = NewDarknet(False, USE_CUDA)

logger = logging.getLogger(__name__)
logger.setLevel(level=logging.DEBUG)
handler = logging.StreamHandler()
formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s')
handler.setFormatter(formatter)
logger.addHandler(handler)

def readb64(data):
   nparr = np.fromstring(base64.b64decode(data), np.uint8)
   img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
   return img

@app.websocket("/")
# @compress.compress()
async def test(request, ws):
    while True:
        data = await ws.recv()

        start = time.time()
        frame = json.loads(data)
        logger.info("Received a new frame, Size: %d, FrameID: %d", len(data), frame['FrameID'])

        # nparr = np.frombuffer(data, np.uint8)
        # img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
        img = readb64(frame['Frame'])

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
                'Conf': boxes[i][5],
                'Name': boxes[i][6],
            })

        response = {
            'FrameID': frame["FrameID"],
            'Boxes': formatBoxes,
            'ClientToServerTime': int(time.time()*1000000000 - frame["SendTime"]),
            'SendTime': int(time.time()*1000000000),
            'ProcessTime': int((time.time() - start) *1000000000)
        }
        logger.info("FrameID: %d, Process Time: %f", frame['FrameID'], time.time()-start)

        await ws.send(json.dumps(response))

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=12345, protocol=WebSocketProtocol)

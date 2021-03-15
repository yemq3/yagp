from sanic import Sanic
from sanic.response import json
from sanic.websocket import WebSocketProtocol
import cv2
import numpy as np
import time
import json
import base64
import logging

logger = logging.getLogger(__name__)
logger.setLevel(level=logging.DEBUG)
handler = logging.StreamHandler()
formatter = logging.Formatter("%(asctime)s - %(levelname)s - %(message)s")
handler.setFormatter(formatter)
logger.addHandler(handler)


def readb64(data):
    nparr = np.fromstring(base64.b64decode(data), np.uint8)
    img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
    return img


def createApp(bandwidth, delay, processTime):
    app = Sanic("server")
    handler = createHandler(bandwidth, delay, processTime)
    app.add_websocket_route(handler, "/")
    return app


def createHandler(bandwidth, delay, processTime):
    async def detect(request, ws):
        while True:
            data = await ws.recv()

            start = time.time()
            frame = json.loads(data)
            logger.info(
                "Received a new frame, Size: %d, FrameID: %d",
                len(data),
                frame["FrameID"],
            )
            time.sleep(delay)
            time.sleep(len(data) / bandwidth)

            # img = readb64(frame['Frame'])
            # sized = cv2.resize(img, (darknet.width, darknet.height))
            # sized = cv2.cvtColor(sized, cv2.COLOR_BGR2RGB)

            boxes = []

            with open("./result/result_xyxy/{}.txt".format(frame["FrameID"]), "r") as f:
                lines = f.readlines()
                for line in lines:
                    box = line.split(" ")
                    boxes.append([box[2], box[3], box[4], box[5], box[1], box[0]])

            formatBoxes = []
            for box in boxes:
                formatBoxes.append(
                    {
                        "X1": float(box[0]),
                        "Y1": float(box[1]),
                        "X2": float(box[2]),
                        "Y2": float(box[3]),
                        "Conf": float(box[4]),
                        "Name": box[5],
                    }
                )
            time.sleep(processTime)

            response = {
                "FrameID": frame["FrameID"],
                "Boxes": formatBoxes,
                "ClientToServerTime": int(time.time() * 1000000000 - frame["SendTime"]),
                "SendTime": int(time.time() * 1000000000),
                "ProcessTime": int(processTime * 1000000000),
            }
            logger.info("FrameID: %d, Process Time: %f", frame["FrameID"], processTime)

            await ws.send(json.dumps(response))

    return detect


if __name__ == "__main__":
    # lte 50mbps 50ms
    # 5g 200mbps 20ms
    # wifi 40mbps 1ms
    # wifi 802.11ac 250mbps 1ms
    app = createApp(1e10, 0, 0.01)
    app.run(host="0.0.0.0", port=12345, protocol=WebSocketProtocol)

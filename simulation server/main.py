from sanic import Sanic
from sanic.response import json
from sanic.websocket import WebSocketProtocol
import cv2
import numpy as np
import time
import json
import base64
import logging
import math
import random

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


class SimulationServer:
    def __init__(self, bandwidth, delay, processTime, simulationMethod, fluctuationRange, frequency, theta):
        self.app = Sanic("server")

        self.baseBandwidth = bandwidth
        self.baseDelay = delay
        self.baseProcessTime = processTime
        self.simulationMethod = simulationMethod
        self.fluctuationRange = fluctuationRange
        self.frequency = frequency
        self.theta = theta / 360 * 2 * math.pi

        self.bandwidth = (1 + fluctuationRange[0] * math.sin(theta)) * bandwidth
        self.delay = (1 + fluctuationRange[1] * math.sin(theta)) * delay
        self.processTime = (1 + fluctuationRange[2] * math.sin(theta)) * processTime

        assert simulationMethod == "sin" or simulationMethod == "random"
        assert type(fluctuationRange) == list and len(fluctuationRange) == 3
        for r in fluctuationRange:
            assert 0 <= r <= 1

    def update(self):
        if self.simulationMethod == "sin":
            self.theta = self.theta + 2 * math.pi / self.frequency
            self.bandwidth = (1 + self.fluctuationRange[0] * math.sin(self.theta)) * self.baseBandwidth
            self.delay = (1 + self.fluctuationRange[1] * math.sin(self.theta)) * self.baseDelay
            self.processTime = (1 + self.fluctuationRange[2] * math.sin(self.theta)) * self.baseProcessTime
        elif self.simulationMethod == "random":
            self.bandwidth = self.baseBandwidth * random.uniform(1 - self.fluctuationRange[0], 1 + self.fluctuationRange[0])
            self.delay = self.baseDelay * random.uniform(1 - self.fluctuationRange[1], 1 + self.fluctuationRange[1])
            self.processTime = self.baseProcessTime * random.uniform(1 - self.fluctuationRange[2], 1 + self.fluctuationRange[2])

    async def handler(self, request, ws):
        while True:
            data = await ws.recv()

            start = time.time()
            frame = json.loads(data)
            logger.info(
                "Received a new frame, Size: %d, FrameID: %d",
                len(data),
                frame["FrameID"],
            )
            logger.info(f"banwidth: {self.bandwidth}, delay: {self.delay}, process time: {self.processTime}")

            client_to_server_time = time.time_ns() - frame["SendTime"]
            # 来回的delay
            time.sleep(self.delay * 2)
            time.sleep(len(data) / self.bandwidth)

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
            time.sleep(self.processTime)

            response = {
                "FrameID": frame["FrameID"],
                "Boxes": formatBoxes,
                "ClientToServerTime": client_to_server_time,
                "SendTime": time.time_ns(),
                "ProcessTime": int(self.processTime * 1000000000),
            }
            logger.info("FrameID: %d, Process Time: %f", frame["FrameID"], self.processTime)

            await ws.send(json.dumps(response))
            
            self.update()

    def run(self):
        self.app.add_websocket_route(self.handler, "/")
        self.app.run(host="0.0.0.0", port=12345, protocol=WebSocketProtocol)


if __name__ == "__main__":
    # lte 50mbps 50ms
    # 5g 200mbps 20ms
    # wifi 40mbps 1ms
    # wifi 802.11ac 250mbps 1ms
    simulationServer = SimulationServer(5e7, 0.0, 0.0, "sin", [0.2, 0, 0], 10, 0)
    simulationServer.run()

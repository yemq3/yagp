from sanic import Sanic
from sanic.response import json
from sanic.websocket import WebSocketProtocol
from app.yolov4 import *
import json
import cv2
import pickle
import asyncio
import random # just for test

app = Sanic("test")

@app.route("/")
async def test(request):
    return json({"hello": "world"})

@app.websocket('/sanic')
async def getImage(request, ws):
    while True:
        data = await ws.recv()
        # print(time.time())
        recv_data = pickle.loads(data)

        nparr = np.frombuffer(recv_data["image"], np.uint8)
        img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)

        sized = cv2.resize(img, (darknet.width, darknet.height))
        sized = cv2.cvtColor(sized, cv2.COLOR_BGR2RGB)
        
        boxes = do_detect(darknet, sized, 0.4, 0.6, USE_CUDA)
        boxes = np.array(boxes[0]).tolist()

        for i in range(len(boxes)):
            boxes[i][6] = class_names[int(boxes[i][6])]
        
        response = {
            'frameid': recv_data["frameid"],
            'boxes': boxes,
            'send_time': time.time()
        }

        # time.sleep(0.05 * (recv_data["frameid"]//100))

        response_data = pickle.dumps(response)

        await ws.send(response_data)



if __name__ == "__main__":
    app.run(protocol=WebSocketProtocol)
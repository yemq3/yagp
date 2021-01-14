import websocket
import time
import cv2
import numpy as np
import pickle
import queue
from multiprocessing import Process, Queue

FRAME_RATE = 5
url = "ws://127.0.0.1/sanic"

# cap = cv2.VideoCapture(0)
# assert cap.isOpened()

# cap.set(cv2.CAP_PROP_FRAME_WIDTH, 640)
# cap.set(cv2.CAP_PROP_FRAME_HEIGHT, 480)
# cap.set(cv2.CAP_PROP_FPS, FRAME_RATE)

def on_open(ws):
    def encode(queue):
        cap = cv2.VideoCapture(0)
        assert cap.isOpened()

        cap.set(cv2.CAP_PROP_FRAME_WIDTH, 640)
        cap.set(cv2.CAP_PROP_FRAME_HEIGHT, 480)
        cap.set(cv2.CAP_PROP_FPS, FRAME_RATE)
        frameid = 0
        prev = time.time()
        while True:
            time_elapsed = time.time() - prev
            if time_elapsed > 1./FRAME_RATE:
                prev = time.time()

                rval, image = cap.read()

                _, img_encoded = cv2.imencode('.jpg', image, [cv2.IMWRITE_JPEG_QUALITY, 50])

                data = {
                    "image": img_encoded.tobytes(),
                    "frameid": frameid,
                }

                frameid += 1
    p = Process(target=encode, args=(encodeData))

        send_data = pickle.dumps(data)

        ws.send(send_data)


def on_message(ws, message):
    print(message)

def on_error(ws, error):
    print(error)

def on_close(ws):
    print("### closed ###")

if __name__ == "__main__":
    encodeData = Queue()
    ws = websocket.WebSocket()
    ws = websocket.WebSocketApp(url, 
                                on_open=on_open, 
                                on_message = on_message, 
                                on_error = on_error, 
                                on_close = on_close)

    ws.run_forever()
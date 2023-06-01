# coding: utf-8

import functools
import threading
import socket
import json
import requests
import tornado


class LoginData:
    login_data = {
        "method": "login",
        "token": "NTkxMDAwMC01OTIwMDAwLHNscw=="
    }


class Config:
    # 下行代理TCP服务器监听地址
    download_proxy_addr = ('127.0.0.1', 19001)
    # 本地udp监听地址，负责接收上行代理转发来的请求
    udp_listen_addr = ('127.0.0.1', 60002)
    # url_get = 'http://127.0.0.1:8080/room/:1'
    # url_post = 'http://127.0.0.1:8080/msgstream/:1'
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36 SLBrowser/8.0.0.5261 SLBChan/10'
    }

# TCP Connection 与下行代理建立连接，可以完成对遥控指令的回应
# 以host:port为key，保证conn不重复建立
class TCPConnection:
    _lock = threading.Lock()
    _connections = {}

    def __new__(cls, host, port):
        key = f"{host}:{port}"
        # 额外的存在判断，提高锁利用的效率
        if key not in cls._connections:
            with cls._lock:
                if key not in cls._connections:
                    cls._connections[key] = super(
                        TCPConnection, cls).__new__(cls)
                    cls._connections[key].host = host
                    cls._connections[key].port = port
                    cls._connections[key].connection = socket.socket(
                        socket.AF_INET, socket.SOCK_STREAM)
        return cls._connections[key]

    def connect(self):
        try:
            with self._lock:
                self.connection = socket.socket(
                    socket.AF_INET, socket.SOCK_STREAM)
                self.connection.connect((self.host, self.port))
                self.connection.settimeout(0.1)  # 设置超时时间为0.1秒
                # 发送hello消息
                data = json.dumps(LoginData.login_data).encode()
                self.send(data)
                response = self.receive(1024*20)
                # to do 漏洞：需要对回复解析，才能确定connection能否正常建立
                print("连接建立请求收到回复:", response.decode())
        except socket.error as e:
            print("Error receiving data. Trying to reconnect...", e)
        except Exception as e:
            # 捕获所有异常的处理代码
            print("An error occurred", e)

    def send(self, data):
        if self.connection:
            self.connection.sendall(data)

    def send_and_rec(self, data):
        for _ in range(3):
            try:
                self.connection.sendall(data)
                response = self.connection.recv(1024)
                print("收到回复:", response.decode())
                return
            except socket.timeout as e:
                print("超时，重新发送数据", e)
                self.connect()
            except Exception as e:
                # 捕获所有异常的处理代码
                print("An error occurred", e)
                self.connect()
        print("发送失败")

    def receive(self, buffer_size):
        if self.connection:
            return self.connection.recv(buffer_size)

    def close(self):
        if self.connection:
            with self._lock:
                if self.connection:
                    self.connection.close()
                    self.connection = None

# data like this
# request_data = {
#    'user': 'tom',
#    'message': 'Hello, Jerry.',
# }
# data = {
#     "request_type": "get",
#     "url": "http://127.0.0.1:8080/room/:1",
#     "request_data": request_data,
# }

# post respone like this
# respone = {
#     "status": "success",
#     "message": message,
# }

# get respone like this
# respone = {
    # "status": "success",
    # "roomid": roomid,
    # "messagecache": messagelist,
# }


class WebClient:
    def __init__(self) -> None:
        self.tcpconn = TCPConnection(
            Config.download_proxy_addr[0], Config.download_proxy_addr[1])
        # register socket
        self.udpconn = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        # 设置为非阻塞，还是阻塞呢？
        self.udpconn.setblocking(0)
        self.udpconn.bind(Config.udp_listen_addr)

    def start(self):
        # start listening
        io_loop = tornado.ioloop.IOLoop.current()
        callback = functools.partial(self.handle_datagram, self.udpconn)
        io_loop.add_handler(self.udpconn.fileno(), callback, io_loop.READ)
        io_loop.start()

    def handle_datagram(self, udp_sock, fd, events):
        datagram, client_address = udp_sock.recvfrom(4096)
        print('receive datagram from %s', client_address)
        # ts = time.time()
        print(datagram)
        # print(datagram.decode('utf-8'))
        data = json.loads(datagram)
        # print(data)
        response = {}

        try:
            if data['request_type'] == 'get':
                response = self.get(data['url'])
            elif data['request_type'] == 'post':
                # print(type(data['request_data']))
                # print(data['url'])
                response = self.post(data['url'], data['request_data'])
            else:
                print('无效的data request_type: {0}'.format(data['request_type']))
        except Exception as e:
            print(e)
            print('send response to {0}\n response: {1}'.format(Config.download_proxy_addr, response))

        response = json.dumps(response).encode('utf-8')
        self.tcpconn.send_and_rec(response)

    def post(self, url_post: str, post_data: dict) -> dict:
        # to do
        data = {}
        response_post = requests.post(
            url=url_post, data=post_data, headers=Config.headers)
        content = response_post.text
        # 反序列化
        obj = json.loads(content)
        print(obj)
        return obj

    def get(self, url_get: str) -> dict:
        # to do
        response_get = requests.get(
            url=url_get, params=None, headers=Config.headers)
        content = response_get.text
        # 反序列化
        obj = json.loads(content)
        # cache = obj["messagecache"]
        print(obj)
        return obj


def main():
    camera_client = WebClient()
    camera_client.start()


if __name__ == '__main__':
    main()

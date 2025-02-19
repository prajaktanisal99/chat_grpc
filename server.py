# python3 -m grpc_tools.protoc -I./. --python_out=./. --grpc_python_out=./. chat.proto
import grpc
import asyncio
from concurrent import futures
from google.protobuf.timestamp_pb2 import Timestamp
import chat_pb2
import chat_pb2_grpc
import datetime

class Connection:
    def __init__(self, stream, user_id):
        self.stream = stream
        self.user_id = user_id
        self.active = True
        self.message_queue = asyncio.Queue()

class ChatService(chat_pb2_grpc.ChatServiceServicer):
    def __init__(self):
        self.connections = []
        self.lock = asyncio.Lock()

    async def CreateStream(self, request, context):
        conn = Connection(context, request.user.id)
        async with self.lock:
            self.connections.append(conn)
        print(f"User {request.user.name} connected.")
        
        try:
            while True:
                message = await conn.message_queue.get()
                await conn.stream.write(message)
        except Exception as e:
            pass
        finally:
            async with self.lock:
                self.connections.remove(conn)
            print(f"User {request.user.name} disconnected.")

    async def SendMessage(self, request, context):
        timestamp = Timestamp()
        timestamp.GetCurrentTime()  # This gets the current time in UTC
        
        async with self.lock:
            for conn in self.connections:
                if conn.user_id == request.receiverId:
                    await conn.message_queue.put(chat_pb2.Message(
                        senderId=request.senderId, 
                        receiverId=request.receiverId, 
                        content=request.content, 
                        timestamp=timestamp
                    ))
                    return chat_pb2.Close()
        return chat_pb2.Close()

    async def BroadcastMessage(self, request, context):
        timestamp = Timestamp()
        timestamp.GetCurrentTime()

        async with self.lock:
            for conn in self.connections:
                if conn.user_id != request.senderId:
                    await conn.message_queue.put(chat_pb2.Message(
                        senderId=request.senderId, 
                        content=request.content, 
                        timestamp=timestamp
                    ))
        return chat_pb2.Close()

async def serve():
    server = grpc.aio.server()
    chat_pb2_grpc.add_ChatServiceServicer_to_server(ChatService(), server)
    server.add_insecure_port("[::]:9090")
    print("Server started at port 9090")
    await server.start()
    await server.wait_for_termination()

if __name__ == "__main__":
    asyncio.run(serve())
import grpc
import asyncio
import chat_pb2
import chat_pb2_grpc
from google.protobuf.timestamp_pb2 import Timestamp
import datetime

def get_user_input():
    user_id = input("Enter User ID: ").strip()
    name = input("Enter Name: ").strip()
    return user_id, name

async def chat_client():
    channel = grpc.aio.insecure_channel("localhost:9090")
    stub = chat_pb2_grpc.ChatServiceStub(channel)
    user_id, name = get_user_input()
    
    async def receive_messages():
        async for msg in stub.CreateStream(chat_pb2.Connect(user=chat_pb2.User(id=user_id, name=name), active=True)):
            print(f"\n[New Message from {msg.senderId}]: {msg.content}")

    asyncio.create_task(receive_messages())
    
    while True:
        print("\nMenu")
        print("1. Send Message")
        print("2. Broadcast Message")
        print("3. Disconnect")
        choice = input("Enter your choice: ").strip()
        
        if choice == "1":
            receiverId = input("Enter Receiver ID: ").strip()
            message = input("Enter Message: ").strip()
            timestamp = Timestamp()
            timestamp.GetCurrentTime()  # Add current timestamp
            await stub.SendMessage(chat_pb2.Message(senderId=user_id, receiverId=receiverId, content=message, timestamp=timestamp))
        elif choice == "2":
            message = input("Enter Message to Broadcast: ").strip()
            timestamp = Timestamp()
            timestamp.GetCurrentTime()  # Add current timestamp
            await stub.BroadcastMessage(chat_pb2.Message(senderId=user_id, content=message, timestamp=timestamp))
        elif choice == "3":
            print("Disconnecting...")
            break
        else:
            print("Invalid choice. Try again.")

if __name__ == "__main__":
    asyncio.run(chat_client())
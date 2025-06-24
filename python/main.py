from fastapi import FastAPI, BackgroundTasks
from pydantic import BaseModel
from asyncio import Lock
from telethon import TelegramClient, events
from telethon.errors import SessionPasswordNeededError
from typing import Dict, List
from contextlib import asynccontextmanager
import os
import io

import httpx
from telethon.tl.functions.contacts import ImportContactsRequest
from telethon.tl.types import InputPhoneContact
from pyrogram import Client
from pyrogram.types import InputMediaPhoto
import httpx
from typing import Optional

# --- Telegram API credentials ---
api_id = 22832353
api_hash = 'cdbd9c38232206e4da015ee165a80966'

# --- Global states ---
active_clients: Dict[str, TelegramClient] = {}
received_messages: Dict[str, List[Dict]] = {}
session_locks: Dict[str, Lock] = {}

SESSION_FOLDER = "sessions"
os.makedirs(SESSION_FOLDER, exist_ok=True)

# TARGET_HOST = "http://ai-seller-bot:8081/telegram/chat/getpythonmessage"
TARGET_HOST = "http://localhost:8081/telegram/chat/getpythonmessage"

# --- Models ---
class PhoneNumber(BaseModel):
    phone: str

class CodeInput(BaseModel):
    phone: str
    code: str
    password: str = ""

class MessageRequest(BaseModel):
    phone: str
    user_id: int  # Changed from 'user' (username) to 'user_id' (ID)
    text: str

class IntegrationResponse(BaseModel):
    code: int
    message: str
    message_id: int

class ImageMessageRequest(BaseModel):
    phone: str
    user_id: int
    message: str
    image_urls: List[str]

# --- Helper: get session path ---
def get_session_path(phone: str) -> str:
    return os.path.join(SESSION_FOLDER, phone.replace("+", ""))


async def register_message_handler(client: TelegramClient, phone: str):
    @client.on(events.NewMessage(incoming=True))
    async def handler(event):
        try:
            sender = await event.get_sender()
            print("2222")
            print(f"Sender: {sender.id}, phone: {getattr(sender, 'phone', None)}")

            if hasattr(sender, 'phone') and sender.phone:
                try:
                    print("11111")
                    contact = InputPhoneContact(
                        client_id=0,
                        phone=sender.phone,
                        first_name=sender.first_name or "Ismsiz",
                        last_name=sender.last_name or ""
                    )
                    await client(ImportContactsRequest([contact]))
                    print(f"✅ {sender.phone} kontaktga qo‘shildi.")
                except Exception as import_error:
                    print(f"⚠️ Kontaktga qo‘shishda xatolik: {import_error}")
           
            response = {
                "code": 0,
                "message": event.message.message
            }

            reply_to_id = event.message.reply_to_msg_id  # None yoki int bo‘ladi
            reply_text = ""

            if reply_to_id:
                try:
                    replied_msg = await event.message.get_reply_message()
                    if replied_msg and replied_msg.message:
                        reply_text = replied_msg.message
                except Exception as e:
                    print(f"⚠️ Reply xabarni olishda xatolik: {e}")

            msg = {
                "phone": phone,
                "fromid": str(event.sender_id),
                "text": event.message.message,
                "message_id": event.message.id,
                "reply_to_message_id": reply_to_id,
                "reply_text": reply_text,
                "code": response["code"],
                "message": response["message"]
            }

            received_messages.setdefault(phone, []).append(msg)
            print("message: ", msg)
            async with httpx.AsyncClient() as http_client:
                await http_client.post(TARGET_HOST, json=msg)

        except Exception as e:
            print(f"❌ Xatolik forwarding: {str(e)}")
            msg = {
                "phone": phone,
                "fromid": str(event.sender_id),
                "text": "",
                "code": 1,
                "message": f"❌ Xabarni qayta yuborishda xatolik: {str(e)}"
            }
            async with httpx.AsyncClient() as http_client:
                await http_client.post(TARGET_HOST, json=msg)


# --- FastAPI app with lifespan ---
@asynccontextmanager
async def lifespan(app: FastAPI):
    session_files = [f for f in os.listdir(SESSION_FOLDER) if f.endswith(".session")]
    for file in session_files:
        phone = file.replace(".session", "")
        session_path = os.path.join(SESSION_FOLDER, phone)
        client = TelegramClient(session_path, api_id, api_hash)
        await client.connect()
        await register_message_handler(client, phone)
        active_clients[phone] = client
        print(f"✅ {phone} ulandi va handler tayyor")
    yield
    for client in active_clients.values():
        await client.disconnect()
    print("❌ Barcha sessiyalar to‘xtatildi")

app = FastAPI(lifespan=lifespan)

# --- 1. Send login code ---
@app.post("/login/send-code", response_model=IntegrationResponse)
async def send_code(data: PhoneNumber):
    print(data.phone)
    session_path = get_session_path(data.phone)
    client = TelegramClient(session_path, api_id, api_hash)
    await client.connect()
    try:
        await client.send_code_request(data.phone)
        active_clients[data.phone] = client
        return IntegrationResponse(code=0, message="📩 Kod yuborildi. Endi /login/verify bilan tasdiqlang.")
    except Exception as e:
        print(f"ERROR: {e}")
        await client.disconnect()
        return IntegrationResponse(code=1, message=f"❌ Kod yuborishda xatolik: {str(e)}")

# --- Kodni tasdiqlash ---
@app.post("/login/verify", response_model=IntegrationResponse)
async def verify_code(data: CodeInput):
    client = active_clients.get(data.phone)
    if not client:
        return IntegrationResponse(code=2, message="❗ Avval /login/send-code'dan foydalaning.")

    try:
        await client.connect()
        await client.sign_in(data.phone, data.code)
    except SessionPasswordNeededError:
        if not data.password:
            return IntegrationResponse(code=3, message="🔒 2 bosqichli parol kerak. Parol bilan qayta yuboring.")
        try:
            await client.sign_in(password=data.password)
        except Exception as e:
            return IntegrationResponse(code=4, message=f"❌ Parol bilan login xatosi: {str(e)}")
    except Exception as e:
        return IntegrationResponse(code=5, message=f"❌ Kod bilan login xatosi: {str(e)}")

    try:
        await register_message_handler(client, data.phone)
        return IntegrationResponse(code=0, message="✅ Telegramga muvaffaqiyatli ulandik!")
    except Exception as e:
        return IntegrationResponse(code=6, message=f"❌ Handler yaratishda xato: {str(e)}")


# --- Response struct (Go tarafdagi IntegrationResponse ga mos) ---
@app.post("/send-message/", response_model=IntegrationResponse)
async def send_message(req: MessageRequest):
    session_path = get_session_path(req.phone)
    if not os.path.exists(session_path + ".session"):
        return IntegrationResponse(code=1, message="❗ Bu telefon raqam bilan bog‘langan sessiya topilmadi.")

    lock = session_locks.setdefault(req.phone, Lock())
    async with lock:
        client = active_clients.get(req.phone)
        if not client:
            client = TelegramClient(session_path, api_id, api_hash)
            await client.connect()
            await register_message_handler(client, req.phone)
            active_clients[req.phone] = client

        try:
            user = await client.get_entity(req.user_id)
            msg = await client.send_message(user, req.text)
            return IntegrationResponse(
                code=0,
                message=f"✅ Xabar {req.user_id} ga yuborildi. Xabar ID: {msg.id}",
                message_id=msg.id
            )
        except Exception as e:
            return IntegrationResponse(code=2, message=f"❌ Xabar yuborishda xatolik: {str(e)}")

# --- 5. View received messages ---
@app.get("/get-messages/{phone}")
async def get_messages(phone: str):
    return {"messages": received_messages.get(phone, [])}

# --- 6. Receive forwarded messages (Frontend/Other API taraf) ---
@app.post("/chat/getpythonmessage")
async def receive_msg(msg: dict):
    print(f"📥 Telegramdan kelgan xabar: {msg}")
    return {"status": "received"}

@app.post("/session/stop", response_model=IntegrationResponse)
async def stop_session(data: PhoneNumber):
    client = active_clients.get(data.phone)
    if not client:
        return IntegrationResponse(code=1, message="❗ Sessiya mavjud emas.")
    
    await client.disconnect()
    del active_clients[data.phone]
    return IntegrationResponse(code=0, message="⏸ Sessiya vaqtincha to‘xtatildi.")


@app.post("/session/start", response_model=IntegrationResponse)
async def start_session(data: PhoneNumber):
    session_path = get_session_path(data.phone)
    if not os.path.exists(session_path + ".session"):
        return IntegrationResponse(code=1, message="❗ Bu telefon raqamga bog‘langan sessiya mavjud emas.")

    lock = session_locks.setdefault(data.phone, Lock())
    async with lock:
        if data.phone in active_clients:
            return IntegrationResponse(code=2, message="⚠️ Sessiya allaqachon faol.")

        client = TelegramClient(session_path, api_id, api_hash)
        await client.connect()
        try:
            await register_message_handler(client, data.phone)
            active_clients[data.phone] = client
            return IntegrationResponse(code=0, message="✅ Sessiya qayta ishga tushirildi.")
        except Exception as e:
            return IntegrationResponse(code=3, message=f"❌ Handlerda xatolik: {str(e)}")

@app.get("/sessions/")
async def list_sessions():
    all_sessions = [f.replace(".session", "") for f in os.listdir(SESSION_FOLDER) if f.endswith(".session")]
    active = list(active_clients.keys())
    return {
        "active_sessions": active,
        "stopped_sessions": list(set(all_sessions) - set(active))
    }


@app.post("/send-message-with-images", response_model=IntegrationResponse)
async def send_message_with_images(req: ImageMessageRequest):
    session_path = get_session_path(req.phone)
    if not os.path.exists(session_path + ".session"):
        return IntegrationResponse(code=1, message="❗ Sessiya topilmadi.")

    # Mavjud clientni ishlatish yoki yangi yaratish (send_message funksiyasidagi kabi)
    lock = session_locks.setdefault(req.phone, Lock())
    async with lock:
        client = active_clients.get(req.phone)
        if not client:
            client = TelegramClient(session_path, api_id, api_hash)
            await client.connect()
            await register_message_handler(client, req.phone)
            active_clients[req.phone] = client

        try:
            # User ID orqali foydalanuvchini tekshirish
            user = await client.get_entity(req.user_id)
            
            # Rasmlarni yuklab olish
            files = []
            if req.image_urls:  # Agar image_urls mavjud bo'lsa
                async with httpx.AsyncClient() as http_client:
                    file_index = 0
                    for url in req.image_urls:
                        # URL ni tekshirish
                        if not url or not url.strip():
                            print(f"⚠️ Bo'sh URL o'tkazib yuborildi")
                            continue
                        
                        # HTTP/HTTPS protokolini tekshirish
                        if not (url.startswith('http://') or url.startswith('https://')):
                            print(f"⚠️ Noto'g'ri URL protokoli: {url}")
                            continue
                        
                        try:
                            response = await http_client.get(url)
                            if response.status_code != 200:
                                print(f"⚠️ HTTP xatolik {response.status_code}: {url}")
                                continue
                            
                            file_bytes = response.content
                            if not file_bytes:
                                print(f"⚠️ Bo'sh fayl: {url}")
                                continue
                            
                            # Content-Type dan kengaytma aniqlash
                            content_type = response.headers.get('content-type', '').lower()
                            if 'jpeg' in content_type or 'jpg' in content_type:
                                ext = '.jpg'
                            elif 'png' in content_type:
                                ext = '.png'
                            elif 'gif' in content_type:
                                ext = '.gif'
                            elif 'webp' in content_type:
                                ext = '.webp'
                            else:
                                # URL dan kengaytma olishga harakat qilish
                                if url.lower().endswith(('.jpg', '.jpeg', '.png', '.gif', '.webp')):
                                    ext = url[url.rfind('.'):].lower()
                                else:
                                    ext = '.jpg'  # Default
                            
                            # BytesIO obyekti yaratish va seek(0) qilish
                            file_obj = io.BytesIO(file_bytes)
                            file_obj.seek(0)  # Muhim: cursor'ni boshiga qaytarish
                            file_obj.name = f"image_{file_index}{ext}"
                            files.append(file_obj)
                            file_index += 1
                            
                        except Exception as e:
                            print(f"⚠️ Rasmni yuklashda xatolik: {url} - {e}")
                            continue

            # Agar rasmlar mavjud bo'lsa - rasmlar bilan xabar yuborish
            if files:
                # Agar bitta rasm bo'lsa
                if len(files) == 1:
                    await client.send_file(
                        user, 
                        files[0], 
                        caption=req.message or "",
                        force_document=False
                    )
                # Agar bir nechta rasm bo'lsa, har birini alohida yuborish
                else:
                    for i, file_obj in enumerate(files):
                        caption = req.message if i == 0 else ""
                        await client.send_file(
                            user, 
                            file_obj, 
                            caption=caption,
                            force_document=False
                        )
                return IntegrationResponse(code=0, message="✅ Rasm(lar) yuborildi.")
            # Agar rasmlar yo'q bo'lsa - faqat xabar yuborish (message bo'sh bo'lsa ham)
            else:
                await client.send_message(user, req.message or "📭")
                return IntegrationResponse(code=0, message="✅ Faqat xabar yuborildi.")
                
        except Exception as e:
            return IntegrationResponse(code=2, message=f"❌ Rasm yuborishda xatolik: {str(e)}")
                
        except Exception as e:
            return IntegrationResponse(code=2, message=f"❌ Rasm yuborishda xatolik: {str(e)}")
        # finally blokini olib tashladik - client disconnectni qilmaymiz
from fastapi import FastAPI
from pydantic import BaseModel
from asyncio import Lock
from telethon import TelegramClient, events
from telethon.errors import SessionPasswordNeededError
from contextlib import asynccontextmanager
from typing import Dict
import io
import os

MEDIA_FOLDER = "downloaded_media"
os.makedirs(MEDIA_FOLDER, exist_ok=True)
# --- Telegram API credentials ---
api_id = 22832353
api_hash = 'cdbd9c38232206e4da015ee165a80966'

# --- Global states ---
active_clients: Dict[str, TelegramClient] = {}
session_locks: Dict[str, Lock] = {}
SESSION_FOLDER = "sessions"
os.makedirs(SESSION_FOLDER, exist_ok=True)
from telethon.tl.types import DocumentAttributeFilename
# --- Models ---
class PhoneNumber(BaseModel):
    phone: str

class CodeInput(BaseModel):
    phone: str
    code: str
    password: str = ""

# --- Helper: get session path ---
def get_session_path(phone: str) -> str:
    return os.path.join(SESSION_FOLDER, phone.replace("+", ""))


async def register_message_handler(client: TelegramClient, phone: str):
    @client.on(events.NewMessage(incoming=True))
    async def handler(event):
        try:
            sender = await event.get_sender()
            chat = await event.get_chat()
            message = event.message
            text = message.message or ""

            print("\n--- Yangi xabar ---")
            print(f"📩 From: {getattr(sender, 'username', sender.id)}")
            print(f"🧑‍💬 Chat: {getattr(chat, 'title', 'Private')} ({chat.id})")

            if text.strip():
                print(f"📝 Text: {text}")

            if message.media:
                media_type = type(message.media).__name__
                print(f"📷 Media turi: {media_type}")

                # Agar hujjat bo‘lsa — mime turi, file_name va size ni log qilamiz
                if hasattr(message, "document") and message.document:
                    mime_type = message.document.mime_type
                    size = message.document.size
                    file_name = "unknown"

                    # Faqat file_name bo'lgan attribute ni izlaymiz
                    for attr in message.document.attributes:
                        if isinstance(attr, DocumentAttributeFilename):
                            file_name = attr.file_name
                            break

                    print(f"🧾 MIME turi: {mime_type}")
                    print(f"📦 Fayl nomi: {file_name}")
                    print(f"📐 Hajmi: {size / 1024 / 1024:.2f} MB")

                # Fayl nomini yasash (agar kengaytma yo‘q bo‘lsa, qo‘shilmaydi)
                filename = f"{MEDIA_FOLDER}/{chat.id}_{message.id}"
                try:
                    print(f"⏬ Yuklash boshlandi: {filename}")
                    path = await client.download_media(message.media, file=filename)
                    if path:
                        size_bytes = os.path.getsize(path)
                        print(f"📥 Saqlandi: {path}")
                        print(f"📏 Yakuniy hajmi: {size_bytes / 1024 / 1024:.2f} MB")
                    else:
                        print("⚠️ Media faylni yuklab bo‘lmadi (path=None)")
                except Exception as e:
                    print(f"❌ Media faylni yuklashda xatolik: {e}")
            else:
                print("ℹ️ Media mavjud emas.")
            print("-------------------\n")

        except Exception as e:
            print(f"❌ Xatolik xabarni ko'rsatishda: {e}")
# --- FastAPI app with lifespan ---
@asynccontextmanager
async def lifespan(app: FastAPI):
    session_files = [f for f in os.listdir(SESSION_FOLDER) if f.endswith(".session")]
    for file in session_files:
        phone = file.replace(".session", "")
        session_path = get_session_path(phone)
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
@app.post("/login/send-code")
async def send_code(data: PhoneNumber):
    session_path = get_session_path(data.phone)
    client = TelegramClient(session_path, api_id, api_hash)
    await client.connect()
    try:
        await client.send_code_request(data.phone)
        active_clients[data.phone] = client
        return {"code": 0, "message": "📩 Kod yuborildi."}
    except Exception as e:
        await client.disconnect()
        return {"code": 1, "message": f"❌ Kod yuborishda xatolik: {str(e)}"}

# --- 2. Verify code ---
@app.post("/login/verify")
async def verify_code(data: CodeInput):
    client = active_clients.get(data.phone)
    if not client:
        return {"code": 2, "message": "❗ Avval /login/send-code'dan foydalaning."}
    try:
        await client.connect()
        await client.sign_in(data.phone, data.code)
    except SessionPasswordNeededError:
        if not data.password:
            return {"code": 3, "message": "🔒 2 bosqichli parol kerak."}
        try:
            await client.sign_in(password=data.password)
        except Exception as e:
            return {"code": 4, "message": f"❌ Parol bilan login xatosi: {str(e)}"}
    except Exception as e:
        return {"code": 5, "message": f"❌ Kod bilan login xatosi: {str(e)}"}

    try:
        await register_message_handler(client, data.phone)
        return {"code": 0, "message": "✅ Telegramga muvaffaqiyatli ulandik!"}
    except Exception as e:
        return {"code": 6, "message": f"❌ Handler yaratishda xato: {str(e)}"}

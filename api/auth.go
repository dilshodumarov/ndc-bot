package api

import (
	"fmt"
	"ndc/ai_bot/config"
	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/internal/infrastructure/telegramuser"
	"net/http"

	instagram "ndc/ai_bot/internal/infrastructure/instagram"

	"github.com/gin-gonic/gin"
)

// TelegramRoutes struct to handle routes
type TelegramRoutes struct {
	cfg             *config.Config
	telegramUscase  telegramuser.Handler
	instagramUscase instagram.Handler
}

// NewTelegramRoutes creates a new Telegram routes controller
func NewTelegramRoutes(telegramUscase telegramuser.Handler, instagramUscase instagram.Handler, cfg *config.Config, r *gin.Engine) *TelegramRoutes {
	handler := &TelegramRoutes{
		cfg:            cfg,
		telegramUscase: telegramUscase,
		instagramUscase: instagramUscase,
	}

	telegramGroup := r.Group("/telegram")
	{
		telegramGroup.POST("/send-code", handler.SendTelegramCode)
		telegramGroup.POST("/verify", handler.SendTelegramVerify)
		telegramGroup.POST("/chat/getpythonmessage", handler.ReceivePythonMessage)
	}

	instagramGroup := r.Group("/instagram")
	{

		instagramGroup.POST("/chat/getpythonmessage", handler.SendMessageToinstagram)
	}


	return handler
}

// SendTelegramCode sends a phone number to the Python backend and returns the response
func (h *TelegramRoutes) SendTelegramCode(c *gin.Context) {
	var phone entity.PhoneNumber
	if err := c.ShouldBindJSON(&phone); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.telegramUscase.SendTelegramCode(phone)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, resp)
}

// SendTelegramVerify verifies the code and optional password with the Python backend
func (h *TelegramRoutes) SendTelegramVerify(c *gin.Context) {
	var input entity.CodeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.telegramUscase.SendTelegramVerify(input)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, resp)
}

// --- Receive message from Telegram (called from Python) ---
func (h *TelegramRoutes) ReceivePythonMessage(c *gin.Context) {
	var message entity.MessageResponse
	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message format"})
		return
	}
	if message.Code != 0 {
		fmt.Println("error: ", message.Message)
		return
	}
	fmt.Printf("📥 Yangi xabar: %+v\n", message)

	h.telegramUscase.AiResponse(message)

	c.JSON(http.StatusOK, gin.H{"message": "✅ Xabar qabul qilindi"})
}

func (h *TelegramRoutes) SendMessageToinstagram(c *gin.Context) {
	var payload entity.InstagramWebhookPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Instagram webhook format"})
		return
	}

	// Foydali ma'lumotni ajratish
	for _, entry := range payload.Entry {
		for _, msg := range entry.Messaging {
			// Echo bo‘lmagan (ya'ni foydalanuvchidan kelgan) va xabar mavjud bo‘lgan holat
			if msg.Message != nil && !msg.Message.IsEcho {
				var incomingText string
				if msg.Message.Text != "" {
					incomingText = msg.Message.Text
				} else if len(msg.Message.Attachments) > 0 {
					// Agar attachment bo'lsa, URL ni o'qish
					incomingText = fmt.Sprintf("Attachment type: %s, URL: %s",
						msg.Message.Attachments[0].Type,
						msg.Message.Attachments[0].Payload.URL,
					)
				} else {
					continue
				}

				// Log qilish
				fmt.Printf("📥 Instagramdan xabar: sender=%s, text=%s\n", msg.Sender.ID, incomingText)

				// AiResponse chaqirilsin (kerakli struct tuzib)
				h.instagramUscase.SendInstagramMessage(msg,msg.Message.Text)
			}
		}
	}

	// Javob qaytarish
	c.JSON(http.StatusOK, gin.H{"message": "✅ Instagram webhook qabul qilindi"})
}


// // SendTelegramMessage sends a message request to the Python backend
// func (h *TelegramRoutes) SendTelegramMessage(c *gin.Context) {
// 	var msg entity.MessageRequest
// 	if err := c.ShouldBindJSON(&msg); err != nil {
// 		c.JSON(400, gin.H{"error": err.Error()})
// 		return
// 	}

// 	resp, err := h.telegramUscase.SendTelegramMessage(msg)
// 	if err != nil {
// 		c.JSON(500, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(200, resp)
// }

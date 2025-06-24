package api

import (
	"context"
	"fmt"
	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/internal/infrastructure/telegram"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(r *gin.Engine) {
	r.POST("/start", StartNewBot)
	r.POST("/stop", StopBotAPI)
	r.POST("/notification", SendNotification)
	r.POST("/send-message", SendMEssageToBot)
	
}

func StartNewBot(c *gin.Context) {
	var req entity.BotIntegration

	// JSON orqali kelgan ma'lumotni binding qilish
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entity.IntegrationResponse{
			Code:    1,
			Message: "struct is not correct",
		})
		return
	}

	// Bot qo‘shish
	if err := telegram.AddNewBot(req); err != nil {
		c.JSON(http.StatusBadRequest, entity.IntegrationResponse{
			Code:    2,
			Message: "Error while start bot" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusBadRequest, entity.IntegrationResponse{
		Code:    0,
		Message: "Bot started successfully",
	})
}

func StopBotAPI(c *gin.Context) {
	var req entity.BotIntegration

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entity.IntegrationResponse{
			Code:    1,
			Message: "struct is not correct",
		})
		return
	}
	fmt.Println(1111, req)
	err := telegram.StopBot(req.BusinessID)
	if err != nil {
		c.JSON(http.StatusBadRequest, entity.IntegrationResponse{
			Code:    2,
			Message: "Error while stop bot" + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"message": "Bot stopped successfully"})
}

func SendNotification(c *gin.Context) {
	var req entity.BotNotification
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entity.IntegrationResponse{
			Code:    1,
			Message: "struct is not correct",
		})
		return
	}
	fmt.Println(1111, req)
	err := telegram.SendTelegramNotification(req.Guid, req.ProductId)
	if err != nil {
		c.JSON(http.StatusBadRequest, entity.IntegrationResponse{
			Code:    2,
			Message: "Error while SendTelegramNotification" + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"message": "send notification successfully"})
}

func SendMEssageToBot(c *gin.Context) {
	var req entity.SendMessage

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println("error", err)
		c.JSON(http.StatusBadRequest, entity.IntegrationResponse{
			Code:    1,
			Message: "struct is not correct",
		})
		return
	}

	fmt.Println("Request:", req)

	err := telegram.SendMessageBot(context.Background(),req)
	if err != nil {
		fmt.Println("error", err)
		c.JSON(http.StatusBadRequest, entity.IntegrationResponse{
			Code:    2,
			Message: "Error while sending message to bot: " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"message": "Bot message sent successfully"})
}


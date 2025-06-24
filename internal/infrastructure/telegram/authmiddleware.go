package telegram

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/pkg/jwt"

	"github.com/casbin/casbin"
	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/cast"
)

func (h *Handler) AuthMiddleware(e *casbin.Enforcer, message *tgbotapi.Message) gin.HandlerFunc {
	fmt.Println(1111111111)
	return func(c *gin.Context) {
		var (
			userRole string
			act      = true
			obj      = true
		)
		fmt.Println(222222222)
		token := c.GetHeader("Authorization")
		if token == "" {
			userRole = "unauthorized"
		}

		if userRole == "" {
			token = strings.TrimPrefix(token, "Bearer ")

			claims, err := jwt.ParseJWT(token, h.cfg.JWT.Secret)
			if err != nil {
				userRole = "unauthorized"
			}

			v, ok := claims["user_role"].(string)
			if !ok {
				userRole = "unauthorized"
			} else {
				userRole = v
			}

			for key, value := range claims {
				c.Request.Header.Set(key, fmt.Sprintf("%v", value))
			}
		}

		ok := e.Enforce(userRole, obj, act)
		fmt.Println("role: ", userRole)
		fmt.Println("path: ", obj)
		fmt.Println("method: ", act)

		state:=entity.ClientState{State: "first_name"}

		if !ok {
			log.Println("ACCESS DENIED 1")
			h.handleRegistrationInput(message, message.Chat.ID, &state)
			log.Println("ACCESS DENIED 2")
			return
		}

		c.Next()
	}
}

func (h *Handler) GetIdFromToken(c *gin.Context) (string, int) {
	var softToken string
	token := c.GetHeader("Authorization")
	if token == "" {
		return "unauthorized", http.StatusUnauthorized
	} else if strings.Contains(token, "Bearer") {
		softToken = strings.TrimPrefix(token, "Bearer ")
	} else {
		softToken = token
	}

	claims, err := jwt.ParseJWT(softToken, h.cfg.JWT.Secret)
	if err != nil {
		return "unauthorized", http.StatusUnauthorized
	}

	return cast.ToString(claims["sub"]), 0
}

package telegramuser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"ndc/ai_bot/internal/entity"
	"net/http"
	"strings"
)

// const pythonBaseURL = "http://telegram-fastapi:8000"
const pythonBaseURL = "http://localhost:8000"

// SendTelegramCode sends a phone number to the Python backend and returns the response
func (t *Handler) SendTelegramCode(phone entity.PhoneNumber) (*entity.IntegrationResponse, error) {
	body, err := json.Marshal(phone)
	if err != nil {
		log.Printf("❌ JSON marshal xatolik: %v", err)
		return nil, err
	}

	resp, err := http.Post(pythonBaseURL+"/login/send-code", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("❌ HTTP POST xatolik (send-code): %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	var Response entity.IntegrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&Response); err != nil {
		log.Printf("❌ JSON decode xatolik: %v", err)
		return nil, err
	}
	if Response.Code != 0 {
		log.Printf("⚠️ Python botdan kelgan xatolik kodi: %d, message: %s", Response.Code, Response.Message)
		return &Response, nil
	}

	log.Printf("✅ Kod yuborildi: %+v", Response)
	return &Response, nil
}

// SendTelegramVerify verifies the code and optional password with the Python backend
func (t *Handler) SendTelegramVerify(input entity.CodeInput) (*entity.IntegrationResponse, error) {
	body, err := json.Marshal(input)
	if err != nil {
		log.Printf("❌ JSON marshal xatolik (verify): %v", err)
		return nil, err
	}

	resp, err := http.Post(pythonBaseURL+"/login/verify", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("❌ HTTP POST xatolik (verify): %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	var Resp entity.IntegrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&Resp); err != nil {
		log.Printf("❌ JSON decode xatolik (verify): %v", err)
		return nil, err
	}

	if Resp.Code != 0 {
		log.Printf("⚠️ Python botdan verify xatolik: %d, message: %s", Resp.Code, Resp.Message)
		return &Resp, nil
	}

	log.Printf("✅ Kod tasdiqlandi: %+v", Resp)
	return &Resp, nil
}

// SendTelegramMessage sends a message request to the Python backend
func (t *Handler) SendTelegramMessage(msg entity.MessageRequest) (*entity.IntegrationResponse, error) {
	body, err := json.Marshal(msg)
	if err != nil {
		log.Printf("❌ JSON marshal xatolik (send-message): %v", err)
		return nil, err
	}
	fmt.Println(444)
	resp, err := http.Post(pythonBaseURL+"/send-message/", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("❌ HTTP POST xatolik (send-message): %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	var Resp entity.IntegrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&Resp); err != nil {
		log.Printf("❌ JSON decode xatolik (send-message): %v", err)
		return nil, err
	}
	if Resp.Code != 0 {
		log.Printf("⚠️ Python botdan message xatolik: %d, message: %s", Resp.Code, Resp.Message)
		return &Resp, nil
	}

	log.Printf("✅ Xabar yuborildi: %+v", Resp)
	return &Resp, nil
}

func (t *Handler) SendMessageToAdmin(chat entity.SendMessageResponse) {
	body, err := json.Marshal(chat)
	if err != nil {
		fmt.Println("Error while marshaling chat message:", err)
		return
	}
	// http://ai-seller-admin:8080/v1/websocket/chat/send-message
	// http://localhost:8080/v1/websocket/chat/send-message
	resp, err := http.Post(pythonBaseURL+"http://localhost:8080/v1/websocket/chat/send-message", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error while sending POST request:", err)
		return
	}
	defer resp.Body.Close()

	var botResp entity.IntegrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&botResp); err != nil {
		fmt.Println("Error while decoding response:", err)
		return
	}

	if botResp.Code != 0 {
		fmt.Println("Bot returned error code:", botResp.Code)
		return
	}
}


func (t *Handler) SendImageMessage(req entity.ImageMessageRequest) {
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Println("❌ Error marshaling request:", err)
		return
	}
	fmt.Println(11111111111111)
	fmt.Println(pythonBaseURL+"/send-message-with-images")
	// URL FastAPI xizmatingizga mos bo'lishi kerak
	resp, err := http.Post(pythonBaseURL+"/send-message-with-images", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("❌ Error sending POST request:", err)
		return
	}
	defer resp.Body.Close()

	var res entity.IntegrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		fmt.Println("❌ Error decoding response:", err)
		return
	}

	if res.Code != 0 {
		fmt.Printf("⚠️ Error from server: %s\n", res.Message)
	} else {
		fmt.Printf("✅ Success: %s\n", res.Message)
	}
}

func EstimateTokenCount(prompt string) int {
	words := strings.Fields(prompt)
	estimatedTokens := float64(len(words)) * 1.5
	return int(estimatedTokens)
}

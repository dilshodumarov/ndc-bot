package chatgpt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ndc/ai_bot/config"
	"ndc/ai_bot/internal/entity"
	"regexp"
	"strconv"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type ChatGpt struct {
	Client *openai.Client
	Model  string
}

func NewChatGPTModel(cfg *config.Config) (*ChatGpt, error) {
	client := openai.NewClient(cfg.ChatGpt.APIKey)
	fmt.Println(2222, cfg.ChatGpt.APIKey)
	return &ChatGpt{
		Client: client,
		Model:  openai.GPT4oMini, // yoki openai.GPT3Dot5Turbo
	}, nil
}

func (g *ChatGpt) FirstStep(ctx context.Context, smartnessPercent int, userInput, ownerCustomPrompts, orderProcessingRules string, chatHistory []map[string]interface{}) (*entity.ActionResponse, error) {
	fmt.Println(1111, userInput)
	prompt := fmt.Sprintf(`
	🧠 Aqllilik darajasi: %d%%
	
	📚 Chat tarixi:
	%s
	
	🏪 Do'kon egasi instruksiyasi:
	%s
	
	⚙️ Buyurtma ishlov ketma-ketligi:
	%s
	
	🎯 Vazifangiz:
	- Agar foydalanuvchi mahsulot qidirayotgan bo‘lsa (mahsulot nomi, turi, kategoriyasi haqida so‘rasa): "is_product_search": true formatini qaytaring.
	- Foydalanuvchi hozirgi so‘rovini yuqoridagi kontekst asosida tahlil qiling.
	- Quyidagi JSON formatlardan birini *faqat* qaytaring.
	
	📦 JSON Formatlar:
	
	⚠️ MUHIM:
	- Barcha ID maydonlar (order_id) **string formatda** bo‘lishi kerak. Misol: '"order_id": "130"'
	- message_id va product_id int 
	- Foydalanuvchi mahsulot qidirayotgan bo'lsa, "is_product_search": true formatni qaytaring
	1. Oddiy javob:
	{
	  "AiResponse": "javob matni",
	  "IsAiResponse": true
	}
	
	2. Mahsulot qidirish:
	{
	  "is_product_search": true,
	  "product":"mahsulotlar",
	  "user_message": "..."
	}
	
	3. Buyurtma berish:
	{
	  "action": "confirm_order",
	  "products": [
		{
		  "product_id": "ID" int bolishi shart,
		  "count": MIQDOR
		}
	  ],
	  "user_message": "...",
	  "message_id": "ID" int
	}
	
	4. Buyurtmalar holatini tekshirish:
	{
	  "action": "get_order_status_all",
	  "user_message": "..."
	}
	
	5. To'lov usuli tanlash:
	{
	  "action": "set_payment_method",
	  "method": "...",
	  "order_id": "...",
	  "user_message": "..."
	}
	
	6. To'lov tasdiqlash:
	{
	  "action": "confirm_payment",
	  "order_id": "...",
	  "payment_screenshot_url": "...",
	  "user_message": "..."
	}
	
	7. Lokatsiya qabul qilish:
	{
	  "action": "set_order_location",
	  "order_id": "...",
	  "location_url": "URL",
	  "location": "manzil",
	  "user_note": "qo'shimcha",
	  "user_message": "..."
	}
	
	8. Buyurtmani bekor qilish:
	{
	  "action": "cancel_order",
	  "order_id": "...",
	  "reason": "...",
	  "user_message": "..."
	}
	
	9. Buyurtma holatini tekshirish:
	{
	  "action": "get_order_status",
	  "order_id": "...",
	  "user_message": "..."
	}
	
	10. Tizimdan tashqari holat:
	{
	  "action": "notification_to_admin",
	  "message": "...",
	  "title": "holat title",
	  "user_message": "muloyimlik bilan userga javop qaytaring"
	}
	
	📝 Qoidalar:
	- ID maydonlar har doim string bo‘lishi shart.
	- Har doim *faqat* sof JSON javob qaytaring. Qo‘shimcha matn yoki izohsiz.
	- JSON faqat '{ va }' orasida bo‘lsin.
	
	👤 Foydalanuvchi so‘rovi:
	"%s"
	`, smartnessPercent, chatHistory, ownerCustomPrompts, orderProcessingRules, userInput)

	resp, err := g.Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: g.Model, // openai.GPT4 yoki openai.GPT3Dot5Turbo
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("AI so'rovi xatosi: %v", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("ChatGPT AI javobi bo'sh")
	}

	answer := strings.TrimSpace(resp.Choices[0].Message.Content)
	fmt.Println("✅ ChatGPT javobi:", answer)

	cleanedJSON := cleanJSONResponse(answer)

	var query entity.ActionResponse
	if err := json.Unmarshal([]byte(cleanedJSON), &query); err != nil {
		log.Println("❌ JSON o‘girish xatosi:", err)
		fmt.Println("JSON tarkibi: ", cleanedJSON)

		return &entity.ActionResponse{
			AiResponse:   answer,
			IsAiResponse: true,
		}, nil
	}

	if (query.Action == "get_order_status" || query.Action == "confirm_order") && query.UserMessage == "" {
		switch query.Action {
		case "get_order_status":
			query.UserMessage = "Buyurtmangiz holatini tekshiryapman..."
		}
	}

	fmt.Println("🔍 Olingan ActionResponse:", query)
	return &query, nil
}

func cleanJSONResponse(input string) string {

	input = strings.TrimPrefix(input, "```json")
	input = strings.TrimPrefix(input, "```")
	input = strings.TrimSuffix(input, "```")
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "\n", "")
	input = strings.ReplaceAll(input, "\r", "")

	if !strings.HasPrefix(input, "{") && !strings.HasSuffix(input, "}") {

		return fmt.Sprintf(`{"AiResponse": %s, "IsAiResponse": true}`, strconv.Quote(input))
	}

	input = regexp.MustCompile(`\s+`).ReplaceAllString(input, " ")

	var jsonObj interface{}
	if err := json.Unmarshal([]byte(input), &jsonObj); err != nil {

		return fmt.Sprintf(`{"AiResponse": %s, "IsAiResponse": true}`, strconv.Quote(input))
	}

	cleanJSON, _ := json.Marshal(jsonObj)
	return string(cleanJSON)
}

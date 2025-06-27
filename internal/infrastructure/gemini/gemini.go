package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ndc/ai_bot/config"
	"ndc/ai_bot/internal/entity"

	repo "ndc/ai_bot/internal/usecase/postgres"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Gemini struct {
	GeminiModel *genai.GenerativeModel
	UseCase     *repo.UseCase
}

func NewGeminiModel(cfg *config.Config) (*Gemini, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.Gemini.APIKey))
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
		return nil, err
	}

	geminiModel := client.GenerativeModel("gemini-1.5-pro")

	return &Gemini{
		GeminiModel: geminiModel,
	}, nil
}

func (g *Gemini) FirstStep(ctx context.Context, smartnessPercent int, userInput, ownerCustomPrompts, orderProcessingRules string, chatHistory []map[string]interface{}) (*entity.ActionResponse, error) {

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
	- Barcha ID maydonlar (product_id, order_id, message_id) **string formatda** bo‘lishi kerak. Misol: '"order_id": "130"'
	- Foydalanuvchi mahsulot qidirayotgan bo'lsa, "is_product_search": true formatni qaytaring
	1. Oddiy javob:
	{
	  "AiResponse": "javob matni",
	  "IsAiResponse": true
	}
	
	2. Mahsulot qidirish:
	{
	  "is_product_search": true
	}
	
	3. Buyurtma berish:
	{
	  "action": "confirm_order",
	  "products": [
		{
		  "product_id": "ID",
		  "count": MIQDOR
		}
	  ],
	  "user_message": "...",
	  "message_id": "ID"
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

	resp, err := g.GeminiModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("AI so'rovi xatosi: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("Gemini AI javobi bo'sh")
	}

	textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return nil, fmt.Errorf("Kutilmagan Gemini kontent turi")
	}

	answer := strings.TrimSpace(string(textPart))
	fmt.Println("✅ Gemini javobi:", answer)
	cleanedJSON := cleanJSONResponse(answer)

	var query entity.ActionResponse
	if err := json.Unmarshal([]byte(cleanedJSON), &query); err != nil {
		log.Println("❌ JSON o'girish xatosi:", err)

		fmt.Println("JSON tarkibi: ", cleanedJSON)

		return &entity.ActionResponse{
			AiResponse:   answer,
			IsAiResponse: true,
		}, nil
	}

	if (query.Action == "get_order_status" ||
		query.Action == "confirm_order") && query.UserMessage == "" {
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

// func (g *Gemini) SecondStep(ctx context.Context, userQuery string) (bool, error) {
// 		prompt := fmt.Sprintf(`
// 		Siz onlayn do'kon uchun aqlli yordamchi bo'lasiz.

// 		Quyida foydalanuvchi so‘rovi berilgan:
// 		"%s"

// 		Sizning vazifangiz:
// 		- Faqat tekshiring: foydalanuvchi mahsulot qidiryaptimi yoki yo'qmi.
// 		- Agar mahsulot qidirayotgan bo'lsa → **faqat** 'true' ni qaytaring.
// 		- Aks holda → **faqat** 'false' ni qaytaring.

// 		Qoidalar:
// 		- Faqat 'true' yoki 'false' qiymat qaytariladi.
// 		- Hech qanday qo‘shimcha izoh, matn, belgilar yoki JSON kiritilmaydi.
// 		- Faqat foydalanuvchi mahsulot qidirayotganda 'true' qaytarasiz.
// 		- Sotip olish boyicha gap ketsa 'false' qaytaring
// 		`, userQuery)

// 	// AI so'rovini yuborish
// 	resp, err := g.GeminiModel.GenerateContent(ctx, genai.Text(prompt))
// 	if err != nil {
// 		return false, fmt.Errorf("AI so'rovi xatosi: %v", err)
// 	}

// 	if len(resp.Candidates) == 0 {
// 		return false, fmt.Errorf("Gemini AI javobi bo'sh")
// 	}

// 	textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
// 	if !ok {
// 		return false, fmt.Errorf("Kutilmagan Gemini kontent turi")
// 	}

// 	answer := strings.TrimSpace(string(textPart))
// 	fmt.Println("✅ Gemini javobi:", answer)

// 	// To'g'ridan-to'g'ri true/false ga tekshirish
// 	var isAction bool
// 	switch strings.ToLower(answer) {
// 	case "true":
// 		isAction = true
// 	case "false":
// 		isAction = false
// 	default:
// 		return false, fmt.Errorf("Kutilmagan javob: %s", answer)
// 	}

// 	return isAction, nil
// }

func (g *Gemini) ExtractProductName(ctx context.Context, userQuery, promt string, productsJSON []byte, chatHistory []map[string]interface{}) (*entity.ProductQuery, error) {

	fmt.Println(string(productsJSON))

	prompt := fmt.Sprintf(`
Siz onlayn do'kon uchun mahsulot tanlashga yordam beruvchi AI-assistent bo'lasiz.

🧑‍💻 Foydalanuvchi so'rovi: "%s"

📦 Mahsulotlar ro'yxati: %s

🏪 Do'kon egasining ko'rsatmalari: %s

🎯 Vazifa:
- Avval foydalanuvchi so‘rovini aniqlang:
  * Agar foydalanuvchi **mahsulotni buyurtma qilish niyatida** bo‘lsa (masalan: "menga iPhone kerak", "Samsung kerak", "Apple Watch sotib olmoqchiman"), u holda mahsulot IDlarini qaytaring.
  * Agar foydalanuvchi **mahsulot haqida maslahat, tavsif, narx yoki xususiyat haqida so‘rasa** (masalan: "qaysi telefon yaxshi?", "narxi qancha?", "kamerasi qanday?"), unga matnli javob bering.

✅ Mahsulot topish natijasi:

{
  "products": [
    {"id": MAHSULOT_ID, "user_message": "Do'kon egasi formatidagi matn"}
  ]
}

✅ Ma'lumot beruvchi javob:

{
  "AiResponse": "Foydalanuvchiga javob",
  "IsAiResponse": true
}

📌 Qoidalar:
- Mahsulot faqat berilgan ro‘yxatdan olinadi.
- Do'kon egasi formatidagi matn — do'kon egasi ko'rsatmasida berilgan formatga mos bo'lishi shart.
- Faqat JSON qaytaring. Hech qanday qo'shimcha matn, tushuntirish yoki izoh bo‘lmasin.
`, userQuery, string(productsJSON), promt)

	// Generate AI response
	resp, err := g.GeminiModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("AI request error: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("Gemini AI response is empty")
	}

	// Extract plain text from Gemini
	textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return nil, fmt.Errorf("Unexpected Gemini content type")
	}

	fmt.Println("✅ Gemini raw JSON response:", textPart)

	answer := string(textPart)

	answer = strings.TrimPrefix(answer, "```json")
	answer = strings.Replace(answer, "```", "", -1)

	answer = strings.TrimSpace(answer)

	fmt.Println("answer: ", answer)

	var query entity.ProductQuery
	if err := json.Unmarshal([]byte(answer), &query); err != nil {
		log.Println("❌ JSON unmarshal error:", err)
		return nil, err
	}

	fmt.Println("🔍 Extracted ProductQuery:", query)

	return &query, nil
}

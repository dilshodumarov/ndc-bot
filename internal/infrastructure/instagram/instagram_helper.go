package telegramuser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

func (t *Handler) SendInstagramMessage(msg entity.Messaging, text string) (*entity.IntegrationResponseInsta, error) {
	payload := map[string]interface{}{
		"recipient": map[string]string{
			"id": msg.Sender.ID,
		},
		"message": map[string]string{
			"text": text,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("❌ JSON marshal xatolik (send-instagram-message): %v", err)
		return nil, err
	}

	url := "https://graph.instagram.com/v23.0/" + msg.Recipient.ID + "/messages"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("❌ HTTP request yaratishda xatolik: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+msg.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Instagramga POST xatolik: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Instagram API status %d: %s", resp.StatusCode, string(bodyBytes))
		return nil, fmt.Errorf("instagram API error: %s", string(bodyBytes))
	}

	var igResp entity.IntegrationResponseInsta
	if err := json.NewDecoder(resp.Body).Decode(&igResp); err != nil {
		log.Printf("❌ JSON decode xatolik: %v", err)
		return nil, err
	}

	log.Printf("✅ Instagram xabar yuborildi: %+v", igResp)
	return &igResp, nil
}

func (t *Handler) SendMessageToAdmin(chat entity.SendMessageResponse) {
	body, err := json.Marshal(chat)
	if err != nil {
		fmt.Println("Error while marshaling chat message:", err)
		return
	}
	// http://ai-seller-admin:8080/v1/websocket/chat/send-message
	// http://localhost:8080/v1/websocket/chat/send-message
	resp, err := http.Post("//http://ai-seller-admin:8080/v1/websocket/chat/send-message", "application/json", bytes.NewBuffer(body))
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
	fmt.Println(pythonBaseURL + "/send-message-with-images")
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


// ----------------------------------------- send request for test ---------------------------------------------------

// curl -X POST http://localhost:8081/instagram/chat/getpythonmessage \
//   -H "Content-Type: application/json" \
//   -d '{
//     "object": "instagram",
//     "entry": [
//       {
//         "id": "17841456246834130",
//         "time": 1718273500,
//         "messaging": [
//           {
//             "sender": {
//               "id": "985732513633547"
//             },
//             "recipient": {
//               "id": "17841456246834130"
//             },
//             "timestamp": 1718273500,
//             "access_token": "IGAAJ9eVGFRoNBZAE9EcjlnaTNJWmE4NEs1U1IxRldiTlpPY19ZAOHdwalBac0J5UXRfY0Q3aW5abmZASTXUzQTdEZAk9PSy1ZAcTZAXbW5xQWNsSXVQd2h0dVVVV3R1RjMzTGdEOWhfRzU3MTZACdDFUS0ZAGMmpB",
//             "message": {
//               "mid": "m_1",
//               "text": "Assalomu alaykum!",
//               "is_echo": false
//             }
//           }
//         ]
//       }
//     ]
//   }'

//   ------------------------------------------ send message -------------------------------------------------------------

//   curl -X POST "https://graph.instagram.com/v23.0/17841456246834130/messages" \
//   -H "Authorization: Bearer IGAAJ9eVGFRoNBZAE1mRnk0TlY2UjA0YV9YMWpvdllYMThGcjJIaFBHczRuSEhlUG9TRk5wckZAhb2VwVzhGYWRQempOMlpTU1JTSE1Pa2tuNFRjQ3o1VDF5X2ZAOdWtwa01ab3lwOWllR2ZAIMlRHb3NSd2NJakt4bFNqcEtibXBkYTZAYZAnRzYzMzYldB" \
//   -H "Content-Type: application/json" \
//   -d '{
//         "recipient": {
//           "id": "985732513633547"
//         },
//         "message": {
//           "text": "Assalomu alaykum!"
//         }
//       }'

//--------------------------------------------- Get a long-lived access token -------------------------------------------

// curl -i -X GET "https://graph.instagram.com/access_token?grant_type=ig_exchange_token\
// &client_secret=dff534402f4026921ee41af2f8a5c415\
// &access_token=IGAAJ9eVGFRoNBZAE1mRnk0TlY2UjA0YV9YMWpvdllYMThGcjJIaFBHczRuSEhlUG9TRk5wckZAhb2VwVzhGYWRQempOMlpTU1JTSE1Pa2tuNFRjQ3o1VDF5X2ZAOdWtwa01ab3lwOWllR2ZAIMlRHb3NSd2NJakt4bFNqcEtibXBkYTZAYZAnRzYzMzYldB"

			//Description:
			//  Converts a short-lived token (received after login) into a long-lived token (valid for 60 days).

			//  Required after user login.

			//  Use the access_token from login response and your app's client_secret.

//-------------------------------------------- Refresh a long-lived token -----------------------------------------------

// curl -X GET "https://graph.instagram.com/refresh_access_token?grant_type=ig_refresh_token&access_token=IGAAJ9eVGFRoNBZAE9EcjlnaTNJWmE4NEs1U1IxRldiTlpPY19ZAOHdwalBac0J5UXRfY0Q3aW5abmZASTXUzQTdEZAk9PSy1ZAcTZAXbW5xQWNsSXVQd2h0dVVVV3R1RjMzTGdEOWhfRzU3MTZACdDFUS0ZAGMmpB"

		// Description:
		//     Extends the expiration of an existing long-lived token by another 60 days.

		//     Token must be long-lived and not expired.

		//     Useful to keep the token active without re-login.








// 		curl -X POST "https://graph.instagram.com/v23.0/17841456246834130/media" \
//   -H "Content-Type: application/json" \
//   -H "Authorization: Bearer IGAAJ9eVGFRoNBZAE9ScDV1RVo2XzNzbmVwRXY3MmpzRnlDZA1ZAxaXNidHFoZAUt4ZAHBSTHViZAnhwaHFvc1JZAa05aaHpGbzUxMkp3YlNhVVNjUGRBT2ExRnhqX0VzZAnhTOFR2bTVtdnk0ZAmhCSTRSSTFSZAFl3" \
//   -d '{
//     "image_url": "https://image.dilshodforever.uz/photos/572609ec-1eb2-401d-a573-8969922534fb.png",
//     "caption": "Hello from API"
//   }'


  

// curl -i -X GET "https://graph.instagram.com/access_token?grant_type=ig_exchange_token&client_secret=dff534402f4026921ee41af2f8a5c415&access_token=IGAAJ9eVGFRoNBZAE9odGtZAWWI3QVlER2dHQ3VTNkFsbTRQanBZANW1oLXBYRDVoTk80NHlGcGpKaTRwYTlSdDhsZAGkxN01oRThJdFJMNHpIbTNrSDlLZAjFMODQyb1YxU3lVZA0lySUI0dXZApb3hIZA3FxRmZAxQW43b1JBcl96ZAUVKYi1pdHlkbFZApa293"



// curl -X POST "https://graph.instagram.com/v23.0/17841474027681924/media" \
//   -H "Content-Type: application/json" \
//   -H "Authorization: Bearer IGAAJ9eVGFRoNBZAE1JN2dlSkJ3RTQ5VGpuTF9CXzlqRlMwQTRnbzJxdTB1SUFWQ3hYaGZAfcVZALMzJSX2NmWmJYazRyaHFNUGxvcG1GS042bG5mbEduc3NvRzlmMDg2bGlaNDQ0SkxkeVFlc0U3UkhKaWxB" \
//   -d '{
//     "image_url": "https://image.dilshodforever.uz/photos/572609ec-1eb2-401d-a573-8969922534fb.png",
//     "is_carousel_item": true
//   }'




// curl -X POST "https://graph.instagram.com/v23.0/17841474027681924/media" \
//   -H "Content-Type: application/json" \
//   -H "Authorization: Bearer IGAAJ9eVGFRoNBZAE1JN2dlSkJ3RTQ5VGpuTF9CXzlqRlMwQTRnbzJxdTB1SUFWQ3hYaGZAfcVZALMzJSX2NmWmJYazRyaHFNUGxvcG1GS042bG5mbEduc3NvRzlmMDg2bGlaNDQ0SkxkeVFlc0U3UkhKaWxB" \
//   -d '{
//     "caption": "Bizning yangi mahsulotlar 📦",
//     "media_type": "CAROUSEL",
//     "children": "17923431765086580,18034460660676615,18008321843760698"
//   }'




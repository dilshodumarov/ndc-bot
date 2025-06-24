package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	baseURL = "https://api.assemblyai.com"
	apiKey  = "906a24e6a0374a94a6f2ab591e6e7738"
)

func main() {
	audioURL, err := uploadAudio("./audio_2025-06-23_13-45-06.ogg")
	if err != nil {
		log.Fatal("Error uploading audio:", err)
	}
	fmt.Println("Uploaded audio URL:", audioURL)

	transcriptID, err := createTranscription(audioURL)
	if err != nil {
		log.Fatal("Error creating transcription:", err)
	}
	fmt.Println("Transcription ID:", transcriptID)

	err = pollTranscription(transcriptID)
	if err != nil {
		log.Fatal("Error polling transcription:", err)
	}
}

// Upload audio file
func uploadAudio(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	req, err := http.NewRequest("POST", baseURL+"/v2/upload", file)
	if err != nil {
		return "", err
	}
	req.Header.Set("authorization", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var uploadResp struct {
		UploadURL string `json:"upload_url"`
	}

	err = json.NewDecoder(resp.Body).Decode(&uploadResp)
	if err != nil {
		return "", err
	}
	return uploadResp.UploadURL, nil
}

// Create transcription
func createTranscription(audioURL string) (string, error) {
	data := map[string]interface{}{
		"audio_url":    audioURL,
		"speech_model": "universal",
	}
	body, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", baseURL+"/v2/transcript", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("authorization", apiKey)
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var transcriptResp struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&transcriptResp)
	if err != nil {
		return "", err
	}
	return transcriptResp.ID, nil
}

// Struct to unmarshal polling response
type Word struct {
	Text       string  `json:"text"`
	Start      int     `json:"start"`
	End        int     `json:"end"`
	Confidence float64 `json:"confidence"`
}

type TranscriptionResponse struct {
	ID         string  `json:"id"`
	Status     string  `json:"status"`
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Words      []Word  `json:"words"`
}

func pollTranscription(transcriptID string) error {
	pollingURL := baseURL + "/v2/transcript/" + transcriptID
	client := &http.Client{}

	for {
		req, err := http.NewRequest("GET", pollingURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("authorization", apiKey)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var result TranscriptionResponse
		err = json.Unmarshal(bodyBytes, &result)
		if err != nil {
			return err
		}

		fmt.Println("Current Status:", result.Status)

		if result.Status == "completed" {
			fmt.Println("Transcript Text:", result.Text)
			fmt.Println("Overall Confidence:", result.Confidence)
			fmt.Println("Words:")
			for _, w := range result.Words {
				fmt.Printf("  - %s (start: %d, end: %d, confidence: %.2f)\n", w.Text, w.Start, w.End, w.Confidence)
			}
			break
		} else if result.Status == "error" {
			return fmt.Errorf("Transcription failed")
		}

		time.Sleep(3 * time.Second)
	}

	return nil
}

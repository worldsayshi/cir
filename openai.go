package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func streamOpenAI(messages []string) (chan string, chan error) {
	resultChan := make(chan string)
	errChan := make(chan error)

	go func() {
		defer close(resultChan)
		defer close(errChan)

		openAIMessages := []Message{}
		for _, msg := range messages {
			openAIMessages = append(openAIMessages, Message{Role: "user", Content: msg})
		}

		reqBody := OpenAIRequest{
			Model:    "gpt-4o-2024-08-06",
			Messages: openAIMessages,
			Stream:   true,
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			log.Println("Error marshalling request body:", err)
			errChan <- err
			return
		}

		req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Println("Error creating request:", err)
			errChan <- err
			return
		}

		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			log.Println("OPENAI_API_KEY environment variable not set")
			errChan <- fmt.Errorf("OPENAI_API_KEY environment variable not set")
			return
		}

		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Error sending request:", err)
			errChan <- err
			return
		}

		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Println("Error reading response:", err)
				errChan <- err
				return
			}

			// Why is this prefix here?
			line = bytes.TrimPrefix(bytes.TrimSpace(line), []byte("data: "))

			if len(line) == 0 {
				continue
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason *string `json:"finish_reason"`
				} `json:"choices"`
			}

			if string(line) == "[DONE]" {
				return
			}

			if err := json.Unmarshal(line, &chunk); err != nil {
				log.Println("Error decoding response:", err)
				log.Println("Response:", string(line))
				errChan <- err
				return
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				resultChan <- chunk.Choices[0].Delta.Content
			}

			if chunk.Choices[0].FinishReason != nil && *chunk.Choices[0].FinishReason == "stop" {
				return
			}
		}
	}()

	return resultChan, errChan
}

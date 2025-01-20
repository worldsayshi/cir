package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
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

	log.Println("Streaming to OpenAI")

	go func() {
		defer close(resultChan)
		defer close(errChan)

		log.Println("Sending messages to OpenAI:", messages)

		openAIMessages := []Message{}
		for _, msg := range messages {
			openAIMessages = append(openAIMessages, Message{Role: "user", Content: msg})
		}

		reqBody := OpenAIRequest{
			Model:    "gpt-4o-2024-08-06",
			Messages: openAIMessages,
			Stream:   true,
		}

		log.Println("Request body:", reqBody)

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			log.Println("Error marshalling request body:", err)
			errChan <- err
			return
		}

		log.Println("Request JSON:", string(jsonData))

		req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Println("Error creating request:", err)
			errChan <- err
			return
		}

		log.Println("Request:", req)

		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			log.Println("OPENAI_API_KEY environment variable not set")
			errChan <- fmt.Errorf("OPENAI_API_KEY environment variable not set")
			return
		}

		log.Println("API key retrieved")

		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Error sending request:", err)
			errChan <- err
			return
		}

		log.Println("Response:", resp)
		log.Printf("Response body: %+v\n", &resp.Body)
		log.Println("Response body type:", reflect.TypeOf(resp.Body))

		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}

			if err := decoder.Decode(&chunk); err != nil {
				if err == io.EOF {
					break
				}
				log.Println("Error decoding response:", err)
				log.Println("Response:", chunk)
				errChan <- err
				return
			}

			log.Println("Chunk:", chunk)

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				resultChan <- chunk.Choices[0].Delta.Content
			}

			log.Println("Result:", chunk.Choices[0].Delta.Content)
		}
	}()

	return resultChan, errChan
}

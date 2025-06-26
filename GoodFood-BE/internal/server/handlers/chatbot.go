package handlers

import (
	"GoodFood-BE/internal/service"
	"context"
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/genai"
)

type PromptRequest struct {
	Prompt string `json:"prompt"`
}

func CallVertexAI(c *fiber.Ctx) error {
	var body PromptRequest
	if err := c.BodyParser(&body); err != nil {
		return service.SendError(c, 400, "Invalid prompt!")
	}

	if body.Prompt == "" {
		return service.SendError(c, 400, "Prompt cannot be empty")
	}

	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  "322745191572",
		Location: "us-central1",
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return service.SendError(c, 500, "Failed to create client: "+err.Error())
	}

	// ✅ Khai báo generation config giới hạn token
	config := &genai.GenerateContentConfig{
		MaxOutputTokens: 1024, // giới hạn token output
		Temperature:     float32Ptr(0.7), // độ sáng tạo
		TopP:            float32Ptr(0.9), // nucleus sampling
	}

	// ✅ Nội dung prompt
	contents := []*genai.Content{
		{	
			Role: "user",
			Parts: []*genai.Part{
				{Text: "You are a concise assistant specialized in answering questions about GoodFood24h - an e-commerce website about ordering and shipping food online. Answer this question in 1 sentence: "+ body.Prompt},
			},
		},
	}

	// ✅ Gọi model fine-tuned
	res, err := client.Models.GenerateContent(ctx,
		"projects/322745191572/locations/us-central1/endpoints/5530821664155107328",
		contents,
		config,
	)
	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}

	debug, _ := json.MarshalIndent(res, "", "  ")
	log.Printf("Full response:\n%s\n", debug)

	// ✅ Parse kết quả
	if len(res.Candidates) == 0 || res.Candidates[0].Content == nil || len(res.Candidates[0].Content.Parts) == 0 {
		return service.SendError(c, 500, "Model responded with no content")
	}
	result := ""
	for _, candidate := range res.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				result += part.Text
			}
		}
	}

	// ✅ Trả kết quả về client
	return c.JSON(fiber.Map{
		"status":  "Success",
		"data":    result,
		"message": "Fine-tuned model response OK",
	})
}

func float32Ptr(f float32) *float32{
	return &f
}

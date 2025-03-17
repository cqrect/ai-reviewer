package ai

import (
	"context"
	"log"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var client *openai.Client

type Answer struct {
	Pass     bool   `json:"pass"`
	Desc     string `json:"desc"`
	Lang     string `json:"lang"`
	Problems []struct {
		File       string `json:"file"`
		Reason     string `json:"reason"`
		Code       string `json:"code"`
		Suggestion string `json:"suggestion"`
	} `json:"problems"`
}

type Summary struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func Init(token, baseUrl string) {
	if baseUrl == "" {
		client = openai.NewClient(option.WithAPIKey(token))
	} else {
		if !strings.HasSuffix(baseUrl, "/") {
			baseUrl = baseUrl + "/"
		}
		client = openai.NewClient(option.WithAPIKey(token), option.WithBaseURL(baseUrl))
	}
}

func Chat(ctx context.Context, model, prompt, v string) (string, error) {
	if client == nil {
		log.Fatal("ai client is nil, maybe you should init it first")
	}

	completion, err := client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(prompt),
				openai.UserMessage(v),
			}),
			Model: openai.F(model),
		},
	)
	if err != nil {
		return "", err
	}

	log.Printf("Input tokens: %d", completion.Usage.PromptTokens)
	log.Printf("Reasoning tokens: %d", completion.Usage.CompletionTokensDetails.ReasoningTokens)
	log.Printf("Output tokens: %d", completion.Usage.CompletionTokens)

	// remove json label
	result := completion.Choices[0].Message.Content
	result = strings.Trim(result, "\n")
	result = strings.TrimSpace(result)
	result = strings.TrimPrefix(result, "```json")
	result = strings.TrimSuffix(result, "```")

	return result, nil
}

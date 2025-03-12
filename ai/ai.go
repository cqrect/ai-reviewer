package ai

import (
	"context"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var client *openai.Client

type Answer struct {
	// 是否通过
	Pass bool `json:"pass"`
	// PR 标题
	Title string `json:"title"`
	// PR 描述
	Description string `json:"description"`
	// 问题
	Problems []struct {
		// 文件名
		File string `json:"file"`
		// 错误代码
		Code string `json:"code"`
		// 错误描述
		Description string `json:"description"`
		// 修改建议
		Suggestion string `json:"suggestion"`
	} `json:"problems"`
}

func Init(token, baseUrl string) {
	if baseUrl == "" {
		client = openai.NewClient(option.WithAPIKey(token))
	} else {
		client = openai.NewClient(option.WithBaseURL(baseUrl), option.WithAPIKey(token))
	}
}

// Chat AI问答
func Chat(ctx context.Context, model, prompt, message string) (string, error) {
	completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(prompt),
			openai.UserMessage(message),
		}),
		Model: openai.F(model),
	})
	if err != nil {
		return "", err
	}

	return completion.Choices[0].Message.Content, nil
}

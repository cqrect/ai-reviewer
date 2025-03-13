package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/cqrect/ai-reviewer/ai"
	"github.com/cqrect/ai-reviewer/gh"
)

var (
	REPO_OWNER      string
	REPO_NAME       string
	PR_NUMBER       int
	GITHUB_TOKEN    string
	OPENAI_API_KEY  string
	OPENAI_BASE_URL string
	MODEL_NAME      string
	LANG            string
)

func init() {
	var err error

	prNumberStr := os.Getenv("INPUT_PR_NUMBER")
	PR_NUMBER, err = strconv.Atoi(prNumberStr)
	if err != nil {
		log.Fatalf("无效的 PR 编号输入: %v", err)
	}

	repoFullName := os.Getenv("GITHUB_REPOSITORY")
	parts := strings.Split(repoFullName, "/")
	REPO_OWNER = parts[0]
	REPO_NAME = parts[1]

	GITHUB_TOKEN = os.Getenv("INPUT_GITHUB_TOKEN")
	OPENAI_API_KEY = os.Getenv("INPUT_OPENAI_API_KEY")
	OPENAI_BASE_URL = os.Getenv("INPUT_OPENAI_BASE_URL")
	MODEL_NAME = os.Getenv("INPUT_MODEL_NAME")
	LANG = os.Getenv("LANG")
}

func main() {
	ctx := context.Background()

	ai.Init(OPENAI_API_KEY, OPENAI_BASE_URL)

	gClient := gh.NewGHClient(GITHUB_TOKEN)

	// 获取 PR 信息
	pr, err := gClient.GetPRDetails(ctx, REPO_OWNER, REPO_NAME, PR_NUMBER)
	if err != nil {
		log.Println("get PR details error")
		log.Fatal(err)
	}

	// 获取修改文件
	files, err := gClient.ListChangeFiles(ctx, REPO_OWNER, REPO_NAME, PR_NUMBER)
	if err != nil {
		log.Println("list change files error")
		log.Fatal(err)
	}

	// 拼接代码
	codeTemplate := "<filename>%s</filename>\n<change>%s</change>\n\n"
	code := fmt.Sprintf("<title>%s</title>\n", *pr.Title)
	for _, file := range files {
		code += fmt.Sprintf(codeTemplate, *file.Filename, *file.Patch)
	}

	// 读取 prompt
	promptTemp, err := os.ReadFile("/app/prompt.txt")
	if err != nil {
		log.Println("read prompt.txt error")
		log.Fatal(err)
	}

	// 设置回答语言
	prompt := strings.ReplaceAll(string(promptTemp), "{{LANG}}", LANG)

	// AI 审查
	result, err := ai.Chat(ctx, MODEL_NAME, prompt, code)
	if err != nil {
		log.Println("AI chat error")
		log.Fatal(err)
	}

	// 解析回答
	var answer ai.Answer
	if err := json.Unmarshal([]byte(result), &answer); err != nil {
		log.Println("parse answer json error")
		log.Fatal(err)
	}

	// 审查通过
	if answer.Pass {
		// 修改 PR 详情
		if err := gClient.UpdatePRDetails(ctx, REPO_OWNER, REPO_NAME, PR_NUMBER, answer.Title, answer.Description); err != nil {
			log.Println("update PR details error")
			log.Fatal(err)
		}

		// 设置审查状态
		if err := gClient.SetReviewStatus(ctx, REPO_OWNER, REPO_NAME, PR_NUMBER, true, "LGTM"); err != nil {
			log.Println("set review status to true error")
			log.Fatal(err)
		}

		return
	}

	// 审核未通过
	comment := "❌ 您的 PR 审核**未通过**，请参考错误信息修改后重新提交哦～\n\n"
	prombleTemplate := "---\n**文件名**: `%s`\n%s\n⚠️ **问题描述**\n%s\n\n💡 **修改建议**\n%s\n\n"
	for _, promble := range answer.Problems {
		comment += fmt.Sprintf(prombleTemplate, promble.File, promble.Code, promble.Description, promble.Suggestion)
	}

	// 审查不通过
	if err := gClient.SetReviewStatus(ctx, REPO_OWNER, REPO_NAME, PR_NUMBER, false, comment); err != nil {
		log.Println("set review status to false error")
		log.Fatal(err)
	}
}

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	_ "embed"

	"github.com/cqrect/ai-reviewer/ai"
	"github.com/cqrect/ai-reviewer/conf"
	"github.com/cqrect/ai-reviewer/gh"
	"github.com/google/go-github/v69/github"
)

//go:embed  prompt.txt
var prompt string

//go:embed summary.txt
var summary string

var (
	GITHUB_TOKEN            string
	GITHUB_REPOSITORY_OWNER string
	GITHUB_REPOSITORY       string
	PR_NUMBER               int
	OPENAI_API_KEY          string
	OPENAI_BASE_URL         string
	MODEL_NAME              string
)

func init() {
	var err error
	PR_NUMBER, err = strconv.Atoi(os.Getenv("INPUT_PR_NUMBER"))
	if err != nil {
		log.Fatalf("invalid PR number: %v", err)
	}
	GITHUB_TOKEN = os.Getenv("INPUT_GITHUB_TOKEN")
	GITHUB_REPOSITORY = os.Getenv("GITHUB_REPOSITORY")
	GITHUB_REPOSITORY_OWNER = os.Getenv("GITHUB_REPOSITORY_OWNER")
	OPENAI_API_KEY = os.Getenv("INPUT_OPENAI_API_KEY")
	OPENAI_BASE_URL = os.Getenv("INPUT_OPENAI_BASE_URL")
	MODEL_NAME = os.Getenv("INPUT_MODEL_NAME")
}

func main() {
	ctx := context.Background()

	client := gh.NewGHClient(GITHUB_TOKEN, GITHUB_REPOSITORY_OWNER, GITHUB_REPOSITORY, PR_NUMBER)
	ai.Init(OPENAI_API_KEY, OPENAI_BASE_URL)

	// get PR details
	pr, err := client.GetPRDetails(ctx)
	if err != nil {
		log.Fatalf("get PR details error: %s", err.Error())
	}

	// try to read config file from repostory
	var config *conf.ReviewConf
	confFile, err := client.GetRawContent(ctx, pr, conf.ConfName)
	if err != nil {
		log.Println("using default config")
	} else {
		data, err := confFile.GetContent()
		if err != nil {
			log.Fatalf("read config file %s error: %s", conf.ConfName, err.Error())
		}
		config, err = conf.LoadConf(data)
		if err != nil {
			log.Fatalf("load config file error: %s", err.Error())
		}
	}

	// addtional prompt
	if config.GetPrompt() != "" {
		prompt += fmt.Sprintf("\nHere are some additional rules you need to follow:\n%s", config.GetPrompt())
	}

	// get PR change files
	files, err := client.ListPRFiles(ctx)
	if err != nil {
		log.Fatalf("list PR files error: %s", err.Error())
	}

	changes := make([]string, 0)
	flags := make([]bool, 0)

	// get PR changes
	for _, file := range files {
		// pass exclude file
		if config.MatchAnyPattern(file.GetFilename()) {
			continue
		}

		// pass deleted file
		if file.GetStatus() == "removed" {
			continue
		}

		// pass file has tag
		fileContent, err := client.GetRawContent(ctx, pr, file.GetFilename())
		if err != nil {
			log.Fatalf("%s get raw content error: %s", file.GetFilename(), err.Error())
		}

		content, err := fileContent.GetContent()
		if err != nil {
			log.Fatalf("get file content error: %s", err.Error())
		}

		if hasReviewPassHeader(content) {
			continue
		}

		code := fmt.Sprintf("Filename: %s\n\nChanges: %s", file.GetFilename(), file.GetPatch())
		result, err := ai.Chat(ctx, MODEL_NAME, prompt, code)
		if err != nil {
			log.Fatalf("ai chat error: %s", err.Error())
		}

		var answer ai.Answer
		if err := json.Unmarshal([]byte(result), &answer); err != nil {
			log.Fatalf("parse ai answer from json error: %s", err.Error())
		}

		changes = append(changes, answer.Desc)
		flags = append(flags, answer.Pass)
		if !answer.Pass {
			temp := make([]string, 0, len(answer.Problems))
			for _, problem := range answer.Problems {
				comment := fmt.Sprintf("### ⚠️ Problem Found\n\n%s\n\n\n**File**:\n%s\n\n\n**Error Code**:\n%s\n\n\n**Suggestion**:\n%s\n", problem.Reason, problem.File, problem.Code, problem.Suggestion)

				temp = append(temp, comment)
			}

			if len(temp) > 0 {
				comment := strings.Join(temp, "\n---\n")

				position := getPosition(file.GetPatch())

				if err := client.CreateComments(ctx, pr, []*github.DraftReviewComment{
					{
						Path:     file.Filename,
						Position: &position,
						Body:     &comment,
					},
				}); err != nil {
					log.Fatalf("create comments error: %s", err.Error())
				}
			}
		}
	}

	// summary
	for _, each := range flags {
		if !each {
			if err := client.UpdatePRReviewStatus(ctx, pr, false, "⚠️ Problem Found"); err != nil {
				log.Fatalf("update review status error: %s", err.Error())
			}
			return
		}
	}

	// all pass
	result, err := ai.Chat(ctx, MODEL_NAME, summary, fmt.Sprintf("Origin PR Title: %s\n\nCommits:\n %s", pr.GetTitle(), strings.Join(changes, "\n\n")))
	if err != nil {
		log.Fatalf("generate PR summary error: %s", err.Error())
	}

	var summary ai.Summary
	if err := json.Unmarshal([]byte(result), &summary); err != nil {
		log.Fatalf("parse ai summary json error: %s", err.Error())
	}

	if err := client.UpdatePRDetails(ctx, summary.Title, summary.Description); err != nil {
		log.Fatalf("update PR details error: %s", err.Error())
	}

	if err := client.UpdatePRReviewStatus(ctx, pr, true, "LGTM"); err != nil {
		log.Fatalf("update review status error: %s", err.Error())
	}
}

func hasReviewPassHeader(content string) bool {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineCount := 0
	for scanner.Scan() {
		if lineCount >= 5 {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "Review: PASS") {
			return true
		}
		lineCount++
	}
	return false
}

func getPosition(v string) int {
	count := 0
	scanner := bufio.NewScanner(strings.NewReader(v))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "@@") {
			continue
		}
		count += 1
	}

	return count
}

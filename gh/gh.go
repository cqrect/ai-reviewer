package gh

import (
	"context"
	"fmt"

	"github.com/google/go-github/v69/github"
)

type GHClient struct {
	client *github.Client
}

func NewGHClient(token string) *GHClient {
	return &GHClient{
		client: github.NewClient(nil).WithAuthToken(token),
	}
}

// GetPRDetails 获取 PR 详情
func (g *GHClient) GetPRDetails(ctx context.Context, repoOwner, repoName string, prNumber int) (*github.PullRequest, error) {
	pr, _, err := g.client.PullRequests.Get(ctx, repoOwner, repoName, prNumber)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

// Comment 给指定 PR 评论
func (g *GHClient) Comment(ctx context.Context, repoOwner, repoName string, prNumber int, comment string) error {
	c := &github.IssueComment{
		Body: &comment,
	}

	_, _, err := g.client.Issues.CreateComment(ctx, repoOwner, repoName, prNumber, c)

	return err
}

// AtComment @PR提交者并评论
func (g *GHClient) AtComment(ctx context.Context, repoOwner, repoName string, prNumber int, comment string) error {
	pr, err := g.GetPRDetails(ctx, repoOwner, repoName, prNumber)
	if err != nil {
		return err
	}

	author := pr.GetUser().GetLogin()
	comment = fmt.Sprintf("@%s %s", author, comment)

	return g.Comment(ctx, repoOwner, repoName, prNumber, comment)
}

// ListChangeFiles 获取修改的文件
func (g *GHClient) ListChangeFiles(ctx context.Context, repoOwner, repoName string, prNumber int) ([]*github.CommitFile, error) {
	files, _, err := g.client.PullRequests.ListFiles(ctx, repoOwner, repoName, prNumber, nil)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// SetReviewStatus 设置审查状态
func (g *GHClient) SetReviewStatus(ctx context.Context, repoOwner, repoName string, prNumber int, pass bool, comment string) error {
	pr, err := g.GetPRDetails(ctx, repoOwner, repoName, prNumber)
	if err != nil {
		return err
	}

	author := pr.GetUser().GetLogin()
	comment = fmt.Sprintf("@%s %s", author, comment)

	var event string
	if pass {
		event = "APPROVE"
	} else {
		event = "REQUEST_CHANGES"
	}

	review := &github.PullRequestReviewRequest{
		Body:  &comment,
		Event: &event,
	}

	_, _, err = g.client.PullRequests.CreateReview(ctx, repoOwner, repoName, prNumber, review)
	return err
}

// UpdatePRDetails 修改 PR 详情
func (g *GHClient) UpdatePRDetails(ctx context.Context, repoOwner, repoName string, prNumber int, title, body string) error {
	prUpdate := &github.PullRequest{
		Title: &title,
		Body:  &body,
	}

	_, _, err := g.client.PullRequests.Edit(ctx, repoOwner, repoName, prNumber, prUpdate)
	return err
}

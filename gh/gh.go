package gh

import (
	"context"
	"fmt"

	"github.com/google/go-github/v69/github"
)

// GHClient encapsulates GitHub API client operations for Pull Request management
type GHClient struct {
	// Authenicated GitHub API client instance
	client *github.Client

	// Repository owner (user or organization)
	owner string

	// Target repository name
	repo string

	// Pull Request number to operate on
	prNumber int
}

// NewGHClient constructs and initializes a GitHub client instance
func NewGHClient(token, owner, repo string, prNumber int) *GHClient {
	return &GHClient{
		client:   github.NewClient(nil).WithAuthToken(token),
		owner:    owner,
		repo:     repo,
		prNumber: prNumber,
	}
}

// GetPRDetails retrieves detailed information about a specific Pull Request
func (c *GHClient) GetPRDetails(ctx context.Context) (*github.PullRequest, error) {
	pr, _, err := c.client.PullRequests.Get(
		ctx,
		c.owner,
		c.repo,
		c.prNumber,
	)
	return pr, err
}

// ListPRFiles lists all modified files in a Pull Request
func (c *GHClient) ListPRFiles(ctx context.Context) ([]*github.CommitFile, error) {
	files, _, err := c.client.PullRequests.ListFiles(
		ctx,
		c.owner,
		c.repo,
		c.prNumber,
		nil,
	)
	return files, err
}

// GetRawContent retrieves raw file content (without diff formatting)
func (c *GHClient) GetRawContent(ctx context.Context, pr *github.PullRequest, name string) (*github.RepositoryContent, error) {
	fileContent, _, _, err := c.client.Repositories.GetContents(
		ctx,
		c.owner,
		c.repo,
		name,
		&github.RepositoryContentGetOptions{
			Ref: pr.Head.GetSHA(),
		},
	)
	return fileContent, err
}

// CreateComments creates multi review comments
func (c *GHClient) CreateComments(ctx context.Context, pr *github.PullRequest, comments []*github.DraftReviewComment) error {
	var event string = "COMMENT"

	_, _, err := c.client.PullRequests.CreateReview(ctx, c.owner, c.repo, c.prNumber, &github.PullRequestReviewRequest{
		CommitID: pr.Head.SHA,
		Event:    &event,
		Comments: comments,
	})

	return err
}

func (c *GHClient) UpdatePRReviewStatus(ctx context.Context, pr *github.PullRequest, pass bool, comment string) error {
	var event string
	if pass {
		event = "COMMENT"
	} else {
		event = "REQUEST_CHANGES"
	}

	author := pr.User.GetLogin()

	mention := fmt.Sprintf("@%s %s", author, comment)

	_, _, err := c.client.PullRequests.CreateReview(ctx, c.owner, c.repo, c.prNumber, &github.PullRequestReviewRequest{
		Event: &event,
		Body:  &mention,
	})
	return err
}

func (c *GHClient) UpdatePRDetails(ctx context.Context, title, body string) error {
	prUpdate := &github.PullRequest{
		Title: &title,
		Body:  &body,
	}

	_, _, err := c.client.PullRequests.Edit(ctx, c.owner, c.repo, c.prNumber, prUpdate)
	return err
}

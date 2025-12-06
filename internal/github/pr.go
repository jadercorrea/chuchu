package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type PullRequest struct {
	Number     int      `json:"number"`
	Title      string   `json:"title"`
	Body       string   `json:"body"`
	State      string   `json:"state"`
	HeadBranch string   `json:"headRefName"`
	BaseBranch string   `json:"baseRefName"`
	URL        string   `json:"url"`
	Author     string   `json:"author"`
	Labels     []string `json:"labels"`
	Assignees  []string `json:"assignees"`
	Reviewers  []string `json:"reviewers"`
	IsDraft    bool     `json:"isDraft"`
	Repository string   `json:"repository"`
}

type ReviewComment struct {
	ID        string `json:"id"`
	Author    string `json:"author"`
	Body      string `json:"body"`
	Path      string `json:"path"`
	Line      int    `json:"line"`
	State     string `json:"state"`
	CreatedAt string `json:"createdAt"`
}

type Review struct {
	ID        string `json:"id"`
	Author    string `json:"author"`
	State     string `json:"state"`
	Body      string `json:"body"`
	Comments  []ReviewComment
	CreatedAt string `json:"createdAt"`
}

type CommitOptions struct {
	Message     string
	IssueNumber int
	FilePaths   []string
	AllFiles    bool
}

func (c *Client) CreateBranch(branchName string, fromBranch string) error {
	if fromBranch == "" {
		fromBranch = "main"
	}

	checkoutCmd := exec.Command("git", "checkout", "-b", branchName, fromBranch)
	if c.workDir != "" {
		checkoutCmd.Dir = c.workDir
	}
	output, err := checkoutCmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "already exists") {
			checkoutCmd = exec.Command("git", "checkout", branchName)
			if c.workDir != "" {
				checkoutCmd.Dir = c.workDir
			}
			output, err = checkoutCmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to checkout existing branch %s: %w\nOutput: %s", branchName, err, string(output))
			}
			return nil
		}
		return fmt.Errorf("failed to create branch %s: %w\nOutput: %s", branchName, err, string(output))
	}

	return nil
}

func (c *Client) CommitChanges(opts CommitOptions) error {
	if opts.AllFiles {
		addCmd := exec.Command("git", "add", "-A")
		if c.workDir != "" {
			addCmd.Dir = c.workDir
		}
		if output, err := addCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to stage all files: %w\nOutput: %s", err, string(output))
		}
	} else if len(opts.FilePaths) > 0 {
		args := append([]string{"add"}, opts.FilePaths...)
		addCmd := exec.Command("git", args...)
		if c.workDir != "" {
			addCmd.Dir = c.workDir
		}
		if output, err := addCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to stage files: %w\nOutput: %s", err, string(output))
		}
	}

	commitMsg := opts.Message
	if opts.IssueNumber > 0 {
		commitMsg = fmt.Sprintf("%s\n\nCloses #%d", commitMsg, opts.IssueNumber)
	}

	commitCmd := exec.Command("git", "commit", "-m", commitMsg)
	if c.workDir != "" {
		commitCmd.Dir = c.workDir
	}
	output, err := commitCmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "nothing to commit") {
			return nil
		}
		return fmt.Errorf("failed to commit: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (c *Client) PushBranch(branchName string) error {
	pushCmd := exec.Command("git", "push", "-u", "origin", branchName)
	if c.workDir != "" {
		pushCmd.Dir = c.workDir
	}
	output, err := pushCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to push branch %s: %w\nOutput: %s", branchName, err, string(output))
	}

	return nil
}

func (c *Client) CreatePR(opts PRCreateOptions) (*PullRequest, error) {
	args := []string{"pr", "create"}

	if opts.Title != "" {
		args = append(args, "--title", opts.Title)
	}

	if opts.Body != "" {
		args = append(args, "--body", opts.Body)
	}

	if opts.BaseBranch != "" {
		args = append(args, "--base", opts.BaseBranch)
	}

	if opts.IsDraft {
		args = append(args, "--draft")
	}

	for _, label := range opts.Labels {
		args = append(args, "--label", label)
	}

	for _, assignee := range opts.Assignees {
		args = append(args, "--assignee", assignee)
	}

	for _, reviewer := range opts.Reviewers {
		args = append(args, "--reviewer", reviewer)
	}

	args = append(args, "--repo", c.repo)

	cmd := exec.Command("gh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w\nOutput: %s", err, string(output))
	}

	prURL := strings.TrimSpace(string(output))

	parts := strings.Split(prURL, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid PR URL format: %s", prURL)
	}

	prNumber := 0
	if numStr := parts[len(parts)-1]; numStr != "" {
		prNumber, _ = strconv.Atoi(numStr)
	}

	return &PullRequest{
		Number:     prNumber,
		Title:      opts.Title,
		Body:       opts.Body,
		URL:        prURL,
		HeadBranch: opts.HeadBranch,
		BaseBranch: opts.BaseBranch,
		IsDraft:    opts.IsDraft,
		Labels:     opts.Labels,
		Assignees:  opts.Assignees,
		Reviewers:  opts.Reviewers,
		Repository: c.repo,
	}, nil
}

type PRCreateOptions struct {
	Title      string
	Body       string
	HeadBranch string
	BaseBranch string
	IsDraft    bool
	Labels     []string
	Assignees  []string
	Reviewers  []string
}

func (c *Client) AddLabelsToPR(prNumber int, labels []string) error {
	for _, label := range labels {
		cmd := exec.Command("gh", "pr", "edit", strconv.Itoa(prNumber), "--add-label", label, "--repo", c.repo)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to add label %s to PR #%d: %w\nOutput: %s", label, prNumber, err, string(output))
		}
	}
	return nil
}

func (c *Client) AddReviewersToPR(prNumber int, reviewers []string) error {
	for _, reviewer := range reviewers {
		cmd := exec.Command("gh", "pr", "edit", strconv.Itoa(prNumber), "--add-reviewer", reviewer, "--repo", c.repo)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to add reviewer %s to PR #%d: %w\nOutput: %s", reviewer, prNumber, err, string(output))
		}
	}
	return nil
}

func (c *Client) FetchPRReviews(prNumber int) ([]Review, error) {
	cmd := exec.Command("gh", "pr", "view", strconv.Itoa(prNumber),
		"--json", "reviews",
		"--repo", c.repo)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PR reviews: %w\nOutput: %s", err, string(output))
	}

	var result struct {
		Reviews []struct {
			ID     string `json:"id"`
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
			State     string `json:"state"`
			Body      string `json:"body"`
			CreatedAt string `json:"createdAt"`
		} `json:"reviews"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse reviews: %w", err)
	}

	var reviews []Review
	for _, r := range result.Reviews {
		reviews = append(reviews, Review{
			ID:        r.ID,
			Author:    r.Author.Login,
			State:     r.State,
			Body:      r.Body,
			CreatedAt: r.CreatedAt,
		})
	}

	return reviews, nil
}

func (c *Client) FetchPRComments(prNumber int) ([]ReviewComment, error) {
	cmd := exec.Command("gh", "api",
		fmt.Sprintf("/repos/%s/pulls/%d/comments", c.repo, prNumber))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PR comments: %w\nOutput: %s", err, string(output))
	}

	var apiComments []struct {
		ID   int64 `json:"id"`
		User struct {
			Login string `json:"login"`
		} `json:"user"`
		Body      string `json:"body"`
		Path      string `json:"path"`
		Line      int    `json:"line"`
		State     string `json:"state"`
		CreatedAt string `json:"created_at"`
	}

	if err := json.Unmarshal(output, &apiComments); err != nil {
		return nil, fmt.Errorf("failed to parse comments: %w", err)
	}

	var comments []ReviewComment
	for _, c := range apiComments {
		comments = append(comments, ReviewComment{
			ID:        strconv.FormatInt(c.ID, 10),
			Author:    c.User.Login,
			Body:      c.Body,
			Path:      c.Path,
			Line:      c.Line,
			State:     c.State,
			CreatedAt: c.CreatedAt,
		})
	}

	return comments, nil
}

func (c *Client) GetUnresolvedComments(prNumber int) ([]ReviewComment, error) {
	comments, err := c.FetchPRComments(prNumber)
	if err != nil {
		return nil, err
	}

	var unresolved []ReviewComment
	for _, comment := range comments {
		if comment.State != "RESOLVED" && comment.Body != "" {
			unresolved = append(unresolved, comment)
		}
	}

	return unresolved, nil
}

func GeneratePRBody(issue *Issue, changes []string) string {
	body := fmt.Sprintf("Closes #%d\n\n", issue.Number)
	body += "## Changes\n\n"

	for _, change := range changes {
		body += fmt.Sprintf("- %s\n", change)
	}

	if len(issue.ExtractRequirements()) > 0 {
		body += "\n## Requirements Addressed\n\n"
		for _, req := range issue.ExtractRequirements() {
			body += fmt.Sprintf("- %s\n", req)
		}
	}

	return body
}

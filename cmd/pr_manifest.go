package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type PRInfo struct {
	Number int      `json:"number"`
	Title  string   `json:"title"`
	Files  []string `json:"files"`
}

type PullRequest struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
}

type File struct {
	Filename string `json:"filename"`
}

func main() {
	owner := os.Getenv("GITHUB_OWNER")
	repo := os.Getenv("GITHUB_REPO")

	if owner == "" || repo == "" {
		log.Fatal("GITHUB_OWNER and GITHUB_REPO environment variables must be set")
	}

	prs, err := fetchOpenPRs(owner, repo)
	if err != nil {
		log.Fatalf("Failed to fetch PRs: %v", err)
	}

	var results []PRInfo
	for _, pr := range prs {
		files, err := fetchPRFiles(owner, repo, pr.Number)
		if err != nil {
			log.Fatalf("Failed to fetch files for PR #%d: %v", pr.Number, err)
		}
		results = append(results, PRInfo{
			Number: pr.Number,
			Title:  pr.Title,
			Files:  files,
		})
	}

	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	fmt.Println(string(output))
}

func fetchOpenPRs(owner, repo string) ([]PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=open", owner, repo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var prs []PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, err
	}

	return prs, nil
}

func fetchPRFiles(owner, repo string, prNumber int) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/files", owner, repo, prNumber)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var files []File
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, err
	}

	var filenames []string
	for _, file := range files {
		filenames = append(filenames, file.Filename)
	}

	return filenames, nil
}

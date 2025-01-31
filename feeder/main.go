package main

import (
	"bytes"
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-github/v68/github"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func writePullRequests(prs []pullRequest) {
	db, err := sql.Open("mysql", "username:password@/database")
	failOnError(err)

	failOnError(db.Ping())

	_, err = db.Exec("DROP TABLE IF EXISTS pull_requests")
	failOnError(err)

	_, err = db.Exec(`CREATE TABLE pull_requests (
    id SERIAL PRIMARY KEY,
    repo_name VARCHAR(255),
    pr_number INT,
    status VARCHAR(50)
);`)
	failOnError(err)

	for _, pr := range prs {
		_, err = db.Exec("INSERT INTO pull_requests (repo_name, pr_number, status) VALUES (?, ?, ?)", pr.RepoName, pr.PrNumber, pr.Status)
		failOnError(err)
	}

	failOnError(db.Close())
}

func failOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func getGitHubToken() string {
	cmd := exec.Command("gh", "auth", "token")

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		failOnError(err)
	}

	return strings.TrimSpace(out.String())
}

type pullRequest struct {
	RepoName string
	PrNumber int
	Status   string
}

func getPullRequests(token string) []pullRequest {
	client := github.NewClient(http.DefaultClient).WithAuthToken(token)

	// get all repos for username
	repos, _, err := client.Repositories.ListByOrg(context.Background(), "go-task", nil)
	failOnError(err)

	var result []pullRequest

	for _, r := range repos {
		prs, _, err := client.PullRequests.List(context.Background(), "go-task", r.GetName(), nil)
		failOnError(err)

		for _, curr := range prs {
			result = append(result, pullRequest{RepoName: r.GetName(), PrNumber: curr.GetNumber(), Status: curr.GetState()})
		}

	}
	return result
}

func main() {
	prs := getPullRequests(getGitHubToken())
	writePullRequests(prs)
}

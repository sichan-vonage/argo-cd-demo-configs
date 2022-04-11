package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	githubUsername string
	githubToken    string
	repoURL        string
	pathToFile     string
	content        string
	commitMessage  string
)

type getFileResponse struct {
	Sha     string `json:"sha"`
	Content string `json:"content"`
}

type commitRequest struct {
	Sha     string `json:"sha"`
	Content string `json:"content"`
	Message string `json:"message"`
	Branch  string `json:"branch"`
}

type getBranchResponse struct {
	Ref    string `json:"ref"`
	NodeID string `json:"node_id"`
	URL    string `json:"url"`
	Object struct {
		Sha  string `json:"sha"`
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"object"`
}

type createRefRequest struct {
	Ref string `json:"ref"`
	Sha string `json:"sha"`
}

type createPrRequest struct {
	Title string `json:"title"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}

type createPrResponse struct {
	Number int `json:"number"`
}

type mergeRequest struct {
	CommitMessage string `json:""`
}

// get sha of main branch https://api.github.com/repos/sichan-vonage/argo-cd-demo-configs/branches/main

func main() {
	flag.StringVar(&githubUsername, "github-user", "", "provide the github username")
	flag.StringVar(&githubToken, "github-token", "", "provide the github token")
	flag.StringVar(&repoURL, "repo-url", "", "provide the github repo api url")
	flag.StringVar(&pathToFile, "file-path", "", "provide the the path to the file that needs updating")
	flag.StringVar(&content, "content", "", "content of the file")
	flag.StringVar(&commitMessage, "message", "ci commit", "commit message")
	flag.Parse()

	if githubUsername == "" {
		log.Fatal("missing github-user flag")
	}
	if githubToken == "" {
		log.Fatal("missing github-token flag")
	}
	if repoURL == "" {
		log.Fatal("missing repo-url flag")
	}
	if pathToFile == "" {
		log.Fatal("missing from-url flag")
	}
	if content == "" {
		log.Fatal("missing content flag")
	}

	client := resty.New()
	client = client.SetBasicAuth(githubUsername, githubToken)

	// 1. create new branch from main
	//    * get main branch sha
	//    * create new ref from main

	mainBranchURL := repoURL + "/git/refs/heads/main"
	var getBranchResp getBranchResponse
	resp, err := client.R().SetResult(&getBranchResp).Get(mainBranchURL)
	if err != nil {
		log.Fatalf("failed to make api call to %q: %s", mainBranchURL, err)
	}
	if resp.IsError() {
		log.Fatalf("bad response returned from api call to %q: status=%v body=%s", mainBranchURL, resp.StatusCode(), resp.Body())
	}

	newBranchName := fmt.Sprintf("ci-upgrade-dev-image-%v", time.Now().Unix())
	refsURL := repoURL + "/git/refs"
	resp, err = client.R().SetBody(createRefRequest{Sha: getBranchResp.Object.Sha, Ref: "refs/heads/" + newBranchName}).Post(refsURL)
	if err != nil {
		log.Fatalf("failed to make api call to %q: %s", refsURL, err)
	}
	if resp.IsError() {
		log.Fatalf("bad response returned from api call to %q: status=%v body=%s", refsURL, resp.StatusCode(), resp.Body())
	}

	// 2. get sha of oriignal file

	fileURL := repoURL + "/contents/" + pathToFile
	var devFileResp getFileResponse
	resp, err = client.R().SetResult(&devFileResp).Get(fileURL)
	if err != nil {
		log.Fatalf("failed to make api call to %q: %s", fileURL, err)
	}
	if resp.IsError() {
		log.Fatalf("bad response returned from api call to %q: status=%v body=%s", fileURL, resp.StatusCode(), resp.Body())
	}

	req := commitRequest{
		Sha:     devFileResp.Sha,
		Content: base64.StdEncoding.EncodeToString([]byte(content)),
		Message: commitMessage,
		Branch:  newBranchName,
	}
	resp, err = client.R().SetBody(req).Put(fileURL)
	if err != nil {
		log.Fatalf("failed to make api call to %q: %s", fileURL, err)
	}
	if resp.IsError() {
		log.Fatalf("bad response returned from api call to %q: status=%v body=%s", fileURL, resp.StatusCode(), resp.Body())
	}

	reqPR := createPrRequest{
		Title: fmt.Sprintf("update dev file with '%s'", content),
		Head:  newBranchName,
		Base:  "main",
	}
	prURL := repoURL + "/pulls"
	var prResp createPrResponse
	resp, err = client.R().SetBody(reqPR).SetResult(&prResp).Post(prURL)
	if err != nil {
		log.Fatalf("failed to make api call to %q: %s", prURL, err)
	}
	if resp.IsError() {
		log.Fatalf("bad response returned from api call to %q: status=%v body=%s", prURL, resp.StatusCode(), resp.Body())
	}

	mergeURL := fmt.Sprintf("%s/pulls/%v/merge", repoURL, prResp.Number)
	resp, err = client.R().SetBody(reqPR).Put(mergeURL)
	if err != nil {
		log.Fatalf("failed to make api call to %q: %s", mergeURL, err)
	}
	if resp.IsError() {
		log.Fatalf("bad response returned from api call to %q: status=%v body=%s", mergeURL, resp.StatusCode(), resp.Body())
	}

	log.Printf("response: %s\n", resp.Body())
	log.Print("updated dev image")
}

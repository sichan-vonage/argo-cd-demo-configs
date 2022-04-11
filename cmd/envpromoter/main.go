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
	promoteFromURL string
	promoteToURL   string
	repoURL        string
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

// get sha of main branch https://api.github.com/repos/sichan-vonage/argo-cd-demo-configs/branches/main

func main() {
	flag.StringVar(&githubUsername, "github-user", "", "provide the github username")
	flag.StringVar(&githubToken, "github-token", "", "provide the github token")
	flag.StringVar(&repoURL, "repo-url", "", "provide the github repo api url")
	flag.StringVar(&promoteFromURL, "from-url", "", "the url of the file we want to promote from")
	flag.StringVar(&promoteToURL, "to-url", "", "the url of the file we want to promote to")
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
	if promoteFromURL == "" {
		log.Fatal("missing from-url flag")
	}
	if promoteToURL == "" {
		log.Fatal("missing to-url flag")
	}

	client := resty.New()
	client = client.SetBasicAuth(githubUsername, githubToken)

	// 1. check whether there has been a change from dev to prod
	//    * get dev file
	//    * get prod file
	//    * compare

	var devFileResp getFileResponse
	resp, err := client.R().SetResult(&devFileResp).Get(promoteFromURL)
	if err != nil {
		log.Fatalf("failed to make api call to %q: %s", promoteFromURL, err)
	}
	if resp.IsError() {
		log.Fatalf("bad response returned from api call to %q: status=%v body=%s", promoteFromURL, resp.StatusCode(), resp.Body())
	}

	var prodFileResp getFileResponse
	resp, err = client.R().SetResult(&prodFileResp).Get(promoteToURL)
	if err != nil {
		log.Fatalf("failed to make api call to %q: %s", promoteToURL, err)
	}
	if resp.IsError() {
		log.Fatalf("bad response returned from api call to %q: status=%v body=%s", promoteToURL, resp.StatusCode(), resp.Body())
	}

	if devFileResp.Content == prodFileResp.Content {
		log.Print("no changes made to image")
		return
	}

	buf1, err := base64.StdEncoding.DecodeString(devFileResp.Content)
	if err != nil {
		log.Fatalf("failed to decode file content: %s", err)
	}
	buf2, err := base64.StdEncoding.DecodeString(prodFileResp.Content)
	if err != nil {
		log.Fatalf("failed to decode file content: %s", err)
	}
	log.Print("changes to images have been detected")
	log.Printf("dev file contents: '%s'", buf1)
	log.Printf("prod file contents: '%s'", buf2)
	log.Print("creating PR to promote image...")

	// 2. create new branch from main
	//    * get main branch sha
	//    * create new ref from main

	mainBranchURL := repoURL + "/git/refs/heads/main"
	var getBranchResp getBranchResponse
	resp, err = client.R().SetResult(&getBranchResp).Get(mainBranchURL)
	if err != nil {
		log.Fatalf("failed to make api call to %q: %s", mainBranchURL, err)
	}
	if resp.IsError() {
		log.Fatalf("bad response returned from api call to %q: status=%v body=%s", mainBranchURL, resp.StatusCode(), resp.Body())
	}

	newBranchName := fmt.Sprintf("ci-promotion-%v", time.Now().Unix())
	refsURL := repoURL + "/git/refs"
	resp, err = client.R().SetBody(createRefRequest{Sha: getBranchResp.Object.Sha, Ref: "refs/heads/" + newBranchName}).Post(refsURL)
	if err != nil {
		log.Fatalf("failed to make api call to %q: %s", refsURL, err)
	}
	if resp.IsError() {
		log.Fatalf("bad response returned from api call to %q: status=%v body=%s", refsURL, resp.StatusCode(), resp.Body())
	}

	message := fmt.Sprintf("promiting prod file with: '%s'", buf1)

	req := commitRequest{
		Sha:     prodFileResp.Sha,
		Content: devFileResp.Content,
		Message: message,
		Branch:  newBranchName,
	}
	resp, err = client.R().SetBody(req).Put(promoteToURL)
	if err != nil {
		log.Fatalf("failed to make api call to %q: %s", promoteToURL, err)
	}
	if resp.IsError() {
		log.Fatalf("bad response returned from api call to %q: status=%v body=%s", promoteToURL, resp.StatusCode(), resp.Body())
	}

	reqPR := createPrRequest{
		Title: message,
		Head:  newBranchName,
		Base:  "main",
	}
	prURL := repoURL + "/pulls"
	resp, err = client.R().SetBody(reqPR).Post(prURL)
	if err != nil {
		log.Fatalf("failed to make api call to %q: %s", prURL, err)
	}
	if resp.IsError() {
		log.Fatalf("bad response returned from api call to %q: status=%v body=%s", prURL, resp.StatusCode(), resp.Body())
	}

	log.Printf("response: %s\n", resp.Body())
	log.Print("PR created to promote image")
}

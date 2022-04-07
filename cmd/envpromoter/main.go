package main

import (
	"encoding/base64"
	"flag"
	"log"
	"strings"

	"github.com/go-resty/resty/v2"
)

var (
	githubUsername string
	githubToken    string
	promoteFromURL string
	promoteToURL   string
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
}

func main() {
	flag.StringVar(&githubUsername, "github-user", "", "provide the github username")
	flag.StringVar(&githubToken, "github-token", "", "provide the github token")
	flag.StringVar(&promoteFromURL, "from-url", "", "the url of the file we want to promote from")
	flag.StringVar(&promoteToURL, "to-url", "", "the url of the file we want to promote to")
	flag.StringVar(&commitMessage, "message", "promote image", "the commit message that will be used")
	flag.Parse()

	if githubUsername == "" {
		log.Fatal("missing github-user flag")
	}
	if githubToken == "" {
		log.Fatal("missing github-token flag")
	}
	if promoteFromURL == "" {
		log.Fatal("missing from-url flag")
	}
	if promoteToURL == "" {
		log.Fatal("missing to-url flag")
	}

	client := resty.New()
	client = client.SetBasicAuth(githubUsername, githubToken)

	var fromFileResp getFileResponse
	resp, err := client.R().SetResult(&fromFileResp).Get(promoteFromURL)
	if err != nil {
		log.Fatalf("failed to make api call to %q: %s", promoteFromURL, err)
	}
	if resp.IsError() {
		log.Fatalf("bad response returned from api call to %q: status=%v body=%s", promoteFromURL, resp.StatusCode(), resp.Body())
	}

	var toFileResp getFileResponse
	resp, err = client.R().SetResult(&toFileResp).Get(promoteToURL)
	if err != nil {
		log.Fatalf("failed to make api call to %q: %s", promoteToURL, err)
	}
	if resp.IsError() {
		log.Fatalf("bad response returned from api call to %q: status=%v body=%s", promoteToURL, resp.StatusCode(), resp.Body())
	}

	buf, err := base64.StdEncoding.DecodeString(fromFileResp.Content)
	if err != nil {
		log.Fatalf("failed to decode file contents from %q: %s", promoteToURL, err)
	}
	log.Printf("from file contents: %s", buf)
	log.Printf("from file contents encoded: %s", fromFileResp.Content)

	buf, err = base64.StdEncoding.DecodeString(toFileResp.Content)
	if err != nil {
		log.Fatalf("failed to decode file contents from %q: %s", promoteToURL, err)
	}
	log.Printf("to file contents: %s", buf)
	log.Printf("to file sha: %q", toFileResp.Sha)

	req := commitRequest{
		Sha:     toFileResp.Sha,
		Content: strings.Replace(fromFileResp.Content, "\n", "", 1),
		Message: commitMessage,
	}
	resp, err = client.R().SetBody(req).Put(promoteToURL)
	if err != nil {
		log.Fatalf("failed to make api call to %q: %s", promoteToURL, err)
	}
	if resp.IsError() {
		log.Fatalf("bad response returned from api call to %q: status=%v body=%s", promoteToURL, resp.StatusCode(), resp.Body())
	}

	log.Printf("response: %s\n", resp.Body())
	log.Print("succesfully promoted image")
}

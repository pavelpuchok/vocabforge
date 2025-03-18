package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/pavelpuchok/vocabforge/preply"
	"github.com/pavelpuchok/vocabforge/server"
)

func main() {
	cookiesFile := flag.String("cookiesFile", "", "Path to file with preply cookies")
	serverUrl := flag.String("serverUrl", "", "Vocabforge Server URL")
	tokenFile := flag.String("tokenFile", "", "Path to file with Vocabforge server token")
	userId := flag.Int64("userId", -1, "User ID")

	flag.Parse()

	if *cookiesFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *serverUrl == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *tokenFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *userId == -1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	client := Client{
		URL:    *serverUrl,
		Token:  readSecretsFile(*tokenFile),
		UserID: *userId,
	}

	err := fetchWords(client, strings.TrimSpace(readSecretsFile(*cookiesFile)))
	if err != nil {
		slog.Error("unable to fetch vocabulary", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func readSecretsFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("unable to read secrets file %s. Error: %s", path, err))
	}
	return string(data)
}

func fetchWords(client Client, cookies string) error {
	fetcher := preply.VocabFetcher{}
	limit := 100
	var offset int
	var count int

	for {
		v, err := fetcher.Fetch(cookies, limit, offset)
		if err != nil {
			return fmt.Errorf("failed to fetch vocab. %w", err)
		}

		if err := client.Enqueue(v); err != nil {
			return fmt.Errorf("failed to enqueue vocab. %w", err)
		}

		count += len(v.Words.Nodes)

		slog.Info(fmt.Sprintf("Fetched another %d words", len(v.Words.Nodes)))
		offset += limit
		if offset >= v.Words.TotalCount {
			break
		}
	}

	slog.Info(fmt.Sprintf("Total words fetched %d", count))

	return nil
}

type Client struct {
	URL    string
	Token  string
	UserID int64
}

func (c Client) Enqueue(v preply.Vocabulary) error {
	reqPayload := server.PreplyRequest{
		UserID:     c.UserID,
		Vocabulary: v,
	}
	d, err := json.Marshal(reqPayload)
	if err != nil {
		return fmt.Errorf("unable to marshal payload %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.URL, bytes.NewReader(d))
	if err != nil {
		return fmt.Errorf("unable to create request %w", err)
	}
	req.Header.Add("authorization", "Bearer "+c.Token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to make request %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status %d", res.StatusCode)
	}

	return nil
}

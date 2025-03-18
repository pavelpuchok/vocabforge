package preply

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const graphqlQuery = `query Vocab($subjectAlias: String!, $limit: Int!, $offset: Int!, $translationLangCode: String) {
  vocabulary(subjectAlias: $subjectAlias) {
    progress {
      unviewedWordsAmount
      untrainedWordsAmount
      trainedWordsAmount
      nextExerciseDatetime
      __typename
    }
    practice {
      totalCount
      __typename
    }
    words(limit: $limit, offset: $offset) {
      totalCount
      nodes {
        id
        imageUrl
        spelling
        definition
        translation(languageCode: $translationLangCode) {
          id
          spelling
          __typename
        }
        trainingProgress
        trainingStatus
        viewedAmount
        exampleSentences
        pronunciationUrl
        lexicalCategory
        pronunciationUrl
        __typename
      }
      __typename
    }
    __typename
  }
}`

type response struct {
	Data struct {
		Vocabulary Vocabulary `json:"vocabulary"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type Vocabulary struct {
	Progress struct {
		UnviewedWordsAmount  int    `json:"unviewedWordsAmount"`
		UntrainedWordsAmount int    `json:"untrainedWordsAmount"`
		TrainedWordsAmount   int    `json:"trainedWordsAmount"`
		NextExerciseDatetime string `json:"nextExerciseDatetime"`
	} `json:"progress"`
	Practice struct {
		TotalCount int `json:"totalCount"`
	} `json:"practice"`
	Words struct {
		TotalCount int    `json:"totalCount"`
		Nodes      []Node `json:"nodes"`
	} `json:"words"`
}

type Node struct {
	ID          string `json:"id"`
	ImageUrl    string `json:"imageUrl"`
	Spelling    string `json:"spelling"`
	Definition  string `json:"definition"`
	Translation struct {
		ID       string `json:"id"`
		Spelling string `json:"spelling"`
	} `json:"translation"`
	TrainingProgress int      `json:"trainingProgress"`
	TrainingStatus   string   `json:"trainingStatus"`
	ViewedAmount     int      `json:"viewedAmount"`
	ExampleSentences []string `json:"exampleSentences"`
	PronunciationUrl string   `json:"pronunciationUrl"`
	LexicalCategory  string   `json:"lexicalCategory"`
}

const preplyVocabUrl = "https://preply.com/graphql/v2/Vocab"

type VocabFetcher struct {
}

func (f *VocabFetcher) Fetch(cookie string, limit, offset int) (Vocabulary, error) {
	body, err := f.payload(limit, offset)
	if err != nil {
		return Vocabulary{}, fmt.Errorf("VocabFetcher.Fetch unable to build payload. %w", err)
	}

	res, err := f.fetch(cookie, body)
	if err != nil {
		return Vocabulary{}, fmt.Errorf("VocabFetcher.Fetch unable to fetch. %w", err)
	}

	return res.Data.Vocabulary, nil
}

func (f *VocabFetcher) payload(limit, offset int) (*bytes.Reader, error) {
	payload := map[string]any{
		"operationName": "Vocab",
		"query":         graphqlQuery,
		"variables": map[string]any{
			"subjectAlias":        "english",
			"translationLangCode": "en",
			"limit":               limit,
			"offset":              offset,
		},
	}

	raw, err := json.Marshal(&payload)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal payload. %w", err)
	}

	return bytes.NewReader(raw), nil
}

func (f *VocabFetcher) fetch(cookie string, payload *bytes.Reader) (response, error) {
	req, err := http.NewRequest("POST", preplyVocabUrl, payload)
	if err != nil {
		return response{}, err
	}

	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.9,be;q=0.8,ru;q=0.7")
	req.Header.Set("apollographql-client-name", "edu-frontend")
	req.Header.Set("apollographql-client-version", "edu-frontend-2025-03-18-20-40-arm64")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("sec-ch-ua", "\"Brave\";v=\"131\", \"Chromium\";v=\"131\", \"Not_A Brand\";v=\"24\"")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "\"Linux\"")
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("sec-gpc", "1")
	req.Header.Set("x-accept-language", "en")
	req.Header.Set("Referer", "https://preply.com/edu/english/learn/vocab")
	req.Header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	req.Header.Set("origin", "https://preply.com")
	req.Header.Set("cookie", cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return response{}, err
	}
	defer resp.Body.Close()

	d := json.NewDecoder(resp.Body)

	var res response
	err = d.Decode(&res)
	if err != nil {
		return response{}, fmt.Errorf("unable to decode response. %w", err)
	}

	if len(res.Errors) > 0 {
		return response{}, fmt.Errorf("unable perform requests. Preply Error: %s", res.Errors[0].Message)
	}

	return res, nil
}

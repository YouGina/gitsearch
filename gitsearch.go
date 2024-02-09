package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type GitHubResponse struct {
	Items []struct {
		Url string `json:"url"`
	} `json:"items"`
}

type FileContent struct {
	Content string `json:"content"`
}

func readTokensFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tokens []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tokens = append(tokens, scanner.Text())
	}
	return tokens, scanner.Err()
}


func makeRequestWithRateLimit(url string, tokens []string) (*http.Response, error) {
	client := &http.Client{}
	var resp *http.Response
	var err error

	for _, token := range tokens {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Authorization", "Bearer "+token)

		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}

		
		if resp.StatusCode == 403 || resp.StatusCode == 429 {
			fmt.Println("Rate limit hit, switching tokens...")
			time.Sleep(1 * time.Second) 
			continue
		}
		break 
	}

	return resp, err
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run script.go <tokens-file-path> <query>")
		os.Exit(1)
	}
	tokensFilePath := os.Args[1]
	query := os.Args[2]

	tokens, err := readTokensFromFile(tokensFilePath)
	if err != nil {
		fmt.Println("Error reading tokens from file:", err)
		return
	}
	if len(tokens) == 0 {
		fmt.Println("No tokens provided.")
		os.Exit(1)
	}

	page := 1
	for {
		searchUrl := fmt.Sprintf("https://api.github.com/search/code?per_page=100&type=Code&q=%s&page=%d", query, page)
		resp, err := makeRequestWithRateLimit(searchUrl, tokens)
		if err != nil {
			fmt.Println("Error making request:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Request failed with status code: %d\n", resp.StatusCode)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}

		var gitHubResponse GitHubResponse
		if err := json.Unmarshal(body, &gitHubResponse); err != nil {
			fmt.Println("Error unmarshaling response:", err)
			return
		}

		if len(gitHubResponse.Items) == 0 {
			break 
		}

		for _, item := range gitHubResponse.Items {
			fileResp, err := makeRequestWithRateLimit(item.Url, tokens)
			if err != nil {
				fmt.Println("Error fetching file:", err)
				continue
			}
			defer fileResp.Body.Close()

			fileBody, err := ioutil.ReadAll(fileResp.Body)
			if err != nil {
				fmt.Println("Error reading file response body:", err)
				continue
			}

			var fileContent FileContent
			if err := json.Unmarshal(fileBody, &fileContent); err != nil {
				fmt.Println("Error unmarshaling file content:", err)
				continue
			}

			decodedContent, err := base64.StdEncoding.DecodeString(strings.TrimSpace(fileContent.Content))
			if err != nil {
				fmt.Println("Error decoding base64 content:", err)
				continue
			}

			fmt.Println(string(decodedContent))
		}

		if len(gitHubResponse.Items) < 100 {
			break 
		}

		page++ 
		if page > 10 {
			break // GitHub's limit, stop at 1000 results
		}
	}
}

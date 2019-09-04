package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/r-cbb/cbbpoll/internal/errors"
)

type RedditClient interface {
	UsernameFromToken(token string) (name string, err error)
}

type redditClient struct {
	baseUrl string
}

func NewRedditClient(baseUrl string) RedditClient {
	rc := redditClient{
		baseUrl: baseUrl,
	}

	return rc
}

func (rc redditClient) UsernameFromToken(token string) (name string, err error) {
	var op errors.Op = "reddit.usernameFromRedditToken"
	url := rc.baseUrl + "me"

	req, err := http.NewRequest(http.MethodGet, "https://oauth.reddit.com/api/v1/me", nil)
	if err != nil {
		return "", errors.E(op, err, "error creating http request")
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("User-Agent", "cbbpoll_backend/0.1.0")

	client := &http.Client{}
	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		return "", errors.E(op, err, "error on http request to reddit API", errors.KindServiceUnavailable)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return "", errors.E(op, fmt.Errorf("reddit api returned status %d %s", resp.StatusCode, resp.Status), errors.KindAuthError)
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.E(op, fmt.Errorf("reddit api returned status %d %s", resp.StatusCode, resp.Status, errors.KindServiceUnavailable))
	}

	var content []byte
	content, err = ioutil.ReadAll(resp.Body)
	data := make(map[string]interface{})
	err = json.Unmarshal(content, &data)
	if err != nil {
		return "", errors.E(op, err, "error unmarshaling response from reddit API")
	}

	name, ok := data["name"].(string)
	if !ok {
		return "", errors.E(op, fmt.Errorf("response from reddit API doesn't include expected field 'name'"))
	}

	return name, nil
}
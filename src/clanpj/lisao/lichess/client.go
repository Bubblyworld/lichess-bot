package lichess

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var apiHost = "http://lichess.org/"

var rateLimitCooloff = time.Minute
var rateLimitRetries = 4
var ErrRateLimited = errors.New("api: request was rate limited on each attempt")

type LichessClient struct {
	apiKey string
	client *http.Client

	rateLimitMu   sync.Mutex
	rateLimitTime time.Time
}

func NewLichessClient(apiKey string) *LichessClient {
	return &LichessClient{
		apiKey: apiKey,
		client: &http.Client{
			CheckRedirect: redirectPolicyFunc(apiKey),
		},
	}
}

// Redirects remove the authorization header and by default redirect using a
// GET request. Lichess has privately moved its API so we need to handle these
// two cases directly.
func redirectPolicyFunc(apiKey string) func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		req.Header.Add("Authorization", "Bearer "+apiKey)
		req.Method = via[0].Method
		return nil
	}
}

func (lc *LichessClient) newRequest(method, apiUrl string, params url.Values) (*http.Request, error) {
	body := strings.NewReader(params.Encode())
	url := apiHost + strings.Trim(apiUrl, "/")
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+lc.apiKey)
	return req, nil
}

type requestError struct {
	Error string
}

func (lc *LichessClient) doRequest(req *http.Request) (*http.Response, error) {
	for attempts := 0; attempts < rateLimitRetries; attempts++ {
		cooloff := lc.getRateLimitCooloff()
		if cooloff != 0 {
			log.Printf("api: Rate limited, sleeping for %f seconds.", cooloff.Seconds())
			time.Sleep(cooloff)
		}

		res, err := lc.client.Do(req)
		if err != nil {
			return nil, err
		}

		// We were rate limited.
		if res.StatusCode == 429 {
			lc.setRateLimitTime(time.Now())
			continue
		}

		// An error occurred.
		if res.StatusCode != 200 {
			lichessError := requestError{}
			bytes, err := ioutil.ReadAll(res.Body)
			log.Printf("%s", bytes)

			if err != nil {
				return nil, err
			}

			json.Unmarshal(bytes, &lichessError)
			return nil, errors.New(lichessError.Error)
		}

		return res, nil
	}

	return nil, ErrRateLimited
}

func (lc *LichessClient) doJSONRequest(req *http.Request, buffer interface{}) error {
	res, err := lc.doRequest(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, buffer)
}

func (lc *LichessClient) getRateLimitTime() time.Time {
	lc.rateLimitMu.Lock()
	defer lc.rateLimitMu.Unlock()

	return lc.rateLimitTime
}

func (lc *LichessClient) setRateLimitTime(rateLimitTime time.Time) {
	lc.rateLimitMu.Lock()
	defer lc.rateLimitMu.Unlock()

	lc.rateLimitTime = rateLimitTime
}

func (lc *LichessClient) getRateLimitCooloff() time.Duration {
	rateLimitTime := lc.getRateLimitTime()
	diff := time.Now().Sub(rateLimitTime)

	if diff <= rateLimitCooloff {
		return diff
	}

	return 0
}

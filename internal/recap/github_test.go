package recap

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/cli/go-gh/v2/pkg/api"
)

type graphQLRequest struct {
	Query     string
	Variables map[string]any
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

// Use go-gh's real encoder and decoder with an in-memory transport. Tests never
// read GitHub credentials, open a socket, or invoke the gh executable.
func testAPI(t *testing.T, respond func(graphQLRequest) (string, error)) API {
	t.Helper()
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		defer request.Body.Close()
		var body graphQLRequest
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		response, err := respond(body)
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(response)),
			Request:    request,
		}, nil
	})
	client, err := api.NewGraphQLClient(api.ClientOptions{
		Host: "github.example", AuthToken: "test-token", Transport: transport,
		LogIgnoreEnv: true,
	})
	if err != nil {
		t.Fatalf("create test client: %v", err)
	}
	return client
}

func fixture(t *testing.T, name string) string {
	t.Helper()
	content, err := os.ReadFile("testdata/" + name + ".json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return string(content)
}

func compactQuery(query string) string {
	return strings.Join(strings.Fields(query), "")
}

package verboten

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/reiver/go-pathmux"
)

func TestServeHTTP(t *testing.T) {

	var mux pathmux.Mux
	{
		var handler pathmux.PatternHandler = pathmux.PatternHandlerFunc(serveHTTP)

		err := mux.HandlePattern(handler, pattern)
		if nil != err {
			t.Fatalf("failed to register pattern: %v", err)
		}
	}

	tests := []struct {
		Name             string
		Path             string
		Host             string
		ExpectStatusCode int
		ExpectActorName  string
	}{
		{
			Name:             "valid search- actor",
			Path:             "/gozaar/search-golang",
			Host:             "example.com",
			ExpectStatusCode: http.StatusOK,
			ExpectActorName:  "search-golang",
		},
		{
			Name:             "valid search: actor",
			Path:             "/gozaar/search:golang",
			Host:             "example.com",
			ExpectStatusCode: http.StatusOK,
			ExpectActorName:  "search:golang",
		},
		{
			Name:             "valid example- actor",
			Path:             "/gozaar/example-test",
			Host:             "example.com",
			ExpectStatusCode: http.StatusOK,
			ExpectActorName:  "example-test",
		},
		{
			Name:             "valid example: actor",
			Path:             "/gozaar/example:test",
			Host:             "example.com",
			ExpectStatusCode: http.StatusOK,
			ExpectActorName:  "example:test",
		},
		{
			Name:             "invalid actor - no valid prefix",
			Path:             "/gozaar/joeblow",
			Host:             "example.com",
			ExpectStatusCode: http.StatusNotFound,
		},
		{
			Name:             "invalid actor - empty name",
			Path:             "/gozaar/",
			Host:             "example.com",
			ExpectStatusCode: http.StatusNotFound,
		},
		{
			Name:             "valid actor with port in host",
			Path:             "/gozaar/search-golang",
			Host:             "localhost:8080",
			ExpectStatusCode: http.StatusOK,
			ExpectActorName:  "search-golang",
		},
	}

	for testNumber, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.Path, nil)
			request.Host = test.Host

			recorder := httptest.NewRecorder()

			mux.ServeHTTP(recorder, request)

			result := recorder.Result()

			if test.ExpectStatusCode != result.StatusCode {
				t.Errorf("For test #%d, the actual status-code is not what was expected.", testNumber)
				t.Logf("EXPECTED: %d", test.ExpectStatusCode)
				t.Logf("ACTUAL:   %d", result.StatusCode)
				return
			}

			if http.StatusOK != test.ExpectStatusCode {
				return
			}

			// Check content type
			{
				contentType := result.Header.Get("Content-Type")
				if "application/activity+json" != contentType {
					t.Errorf("For test #%d, the actual content-type is not what was expected.", testNumber)
					t.Logf("EXPECTED: %q", "application/activity+json")
					t.Logf("ACTUAL:   %q", contentType)
					return
				}
			}

			// Parse response body
			var body map[string]any
			{
				err := json.NewDecoder(result.Body).Decode(&body)
				if nil != err {
					t.Errorf("For test #%d, failed to decode response JSON.", testNumber)
					t.Logf("ERROR: %v", err)
					return
				}
			}

			expectedSelf := "https://" + test.Host + "/gozaar/" + test.ExpectActorName

			// Check id
			{
				id, exists := body["id"]
				if !exists {
					t.Errorf("For test #%d, response missing 'id' field.", testNumber)
					return
				}
				if expectedSelf != id {
					t.Errorf("For test #%d, the actual 'id' is not what was expected.", testNumber)
					t.Logf("EXPECTED: %q", expectedSelf)
					t.Logf("ACTUAL:   %q", id)
					return
				}
			}

			// Check name
			{
				name, exists := body["name"]
				if !exists {
					t.Errorf("For test #%d, response missing 'name' field.", testNumber)
					return
				}
				if test.ExpectActorName != name {
					t.Errorf("For test #%d, the actual 'name' is not what was expected.", testNumber)
					t.Logf("EXPECTED: %q", test.ExpectActorName)
					t.Logf("ACTUAL:   %q", name)
					return
				}
			}

			// Check summary
			{
				summary, exists := body["summary"]
				if !exists {
					t.Errorf("For test #%d, response missing 'summary' field.", testNumber)
					return
				}
				expectedSummary := "Search for: " + test.ExpectActorName
				if expectedSummary != summary {
					t.Errorf("For test #%d, the actual 'summary' is not what was expected.", testNumber)
					t.Logf("EXPECTED: %q", expectedSummary)
					t.Logf("ACTUAL:   %q", summary)
					return
				}
			}

			// Check inbox
			{
				inbox, exists := body["inbox"]
				if !exists {
					t.Errorf("For test #%d, response missing 'inbox' field.", testNumber)
					return
				}
				expectedInbox := expectedSelf + "/inbox"
				if expectedInbox != inbox {
					t.Errorf("For test #%d, the actual 'inbox' is not what was expected.", testNumber)
					t.Logf("EXPECTED: %q", expectedInbox)
					t.Logf("ACTUAL:   %q", inbox)
					return
				}
			}

			// Check outbox
			{
				outbox, exists := body["outbox"]
				if !exists {
					t.Errorf("For test #%d, response missing 'outbox' field.", testNumber)
					return
				}
				expectedOutbox := expectedSelf + "/outbox"
				if expectedOutbox != outbox {
					t.Errorf("For test #%d, the actual 'outbox' is not what was expected.", testNumber)
					t.Logf("EXPECTED: %q", expectedOutbox)
					t.Logf("ACTUAL:   %q", outbox)
					return
				}
			}

			// Check endpoints.sharedInbox
			{
				endpointsRaw, exists := body["endpoints"]
				if !exists {
					t.Errorf("For test #%d, response missing 'endpoints' field.", testNumber)
					return
				}
				endpoints, ok := endpointsRaw.(map[string]any)
				if !ok {
					t.Errorf("For test #%d, expected 'endpoints' to be object but got %T.", testNumber, endpointsRaw)
					return
				}
				sharedInbox, exists := endpoints["sharedInbox"]
				if !exists {
					t.Errorf("For test #%d, response missing 'endpoints.sharedInbox' field.", testNumber)
					return
				}
				expectedSharedInbox := "https://" + test.Host + "/inbox"
				if expectedSharedInbox != sharedInbox {
					t.Errorf("For test #%d, the actual 'endpoints.sharedInbox' is not what was expected.", testNumber)
					t.Logf("EXPECTED: %q", expectedSharedInbox)
					t.Logf("ACTUAL:   %q", sharedInbox)
					return
				}
			}
		})
	}
}

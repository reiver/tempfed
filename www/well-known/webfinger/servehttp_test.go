package verboten

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/reiver/go-errhttp"
)

func TestServeActorHost(t *testing.T) {

	tests := []struct {
		Name             string
		Resource         string
		Actor            string
		Host             string
		ExpectError      bool
		ExpectStatusCode int
	}{
		{
			Name:     "valid search- actor",
			Resource: "acct:search-golang@example.com",
			Actor:    "search-golang",
			Host:     "example.com",
		},
		{
			Name:     "valid search: actor",
			Resource: "acct:search:golang@example.com",
			Actor:    "search:golang",
			Host:     "example.com",
		},
		{
			Name:     "valid example- actor",
			Resource: "acct:example-test@example.com",
			Actor:    "example-test",
			Host:     "example.com",
		},
		{
			Name:     "valid example: actor",
			Resource: "acct:example:test@example.com",
			Actor:    "example:test",
			Host:     "example.com",
		},
		{
			Name:             "invalid actor - no valid prefix",
			Resource:         "acct:joeblow@example.com",
			Actor:            "joeblow",
			Host:             "example.com",
			ExpectError:      true,
			ExpectStatusCode: http.StatusNotFound,
		},
		{
			Name:             "invalid actor - empty string",
			Resource:         "acct:@example.com",
			Actor:            "",
			Host:             "example.com",
			ExpectError:      true,
			ExpectStatusCode: http.StatusNotFound,
		},
		{
			Name:             "invalid actor - prefix only no parameter",
			Resource:         "acct:search-@example.com",
			Actor:            "search-",
			Host:             "example.com",
		},
		{
			Name:     "valid actor with port in host",
			Resource: "acct:search-golang@localHost:8080",
			Actor:    "search-golang",
			Host:     "localHost:8080",
		},
	}

	for testNumber, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			bytes, err := serveActorHost(test.Resource, test.Actor, test.Host)

			if test.ExpectError {
				if nil == err {
					t.Errorf("For test #%d, expected an error but got nil.", testNumber)
					return
				}

				httpErr, ok := err.(errhttp.Error)
				if !ok {
					t.Errorf("For test #%d, expected errhttp.Error but got %T.", testNumber, err)
					t.Logf("ERROR: %v", err)
					return
				}

				if test.ExpectStatusCode != httpErr.ErrHTTP() {
					t.Errorf("For test #%d, the actual HTTP status-code is not what was expected.", testNumber)
					t.Logf("EXPECTED: %d", test.ExpectStatusCode)
					t.Logf("ACTUAL:   %d", httpErr.ErrHTTP())
					return
				}

				return
			}

			if nil != err {
				t.Errorf("For test #%d, received an unexpected error.", testNumber)
				t.Logf("ERROR: %v", err)
				return
			}

			if nil == bytes {
				t.Errorf("For test #%d, expected non-nil bytes but got nil.", testNumber)
				return
			}

			var jrd map[string]any
			err = json.Unmarshal(bytes, &jrd)
			if nil != err {
				t.Errorf("For test #%d, failed to unmarshal JRD JSON.", testNumber)
				t.Logf("ERROR: %v", err)
				return
			}

			// Check subject
			{
				subject, exists := jrd["subject"]
				if !exists {
					t.Errorf("For test #%d, JRD missing 'subject' field.", testNumber)
					return
				}
				if test.Resource != subject {
					t.Errorf("For test #%d, the actual 'subject' is not what was expected.", testNumber)
					t.Logf("EXPECTED: %q", test.Resource)
					t.Logf("ACTUAL:   %q", subject)
					return
				}
			}

			// Check aliases
			{
				aliasesRaw, exists := jrd["aliases"]
				if !exists {
					t.Errorf("For test #%d, JRD missing 'aliases' field.", testNumber)
					return
				}
				aliases, ok := aliasesRaw.([]any)
				if !ok {
					t.Errorf("For test #%d, expected 'aliases' to be an array but got %T.", testNumber, aliasesRaw)
					return
				}
				if 1 != len(aliases) {
					t.Errorf("For test #%d, the actual number of aliases is not what was expected.", testNumber)
					t.Logf("EXPECTED: %d", 1)
					t.Logf("ACTUAL:   %d", len(aliases))
					return
				}

				expectedSelf := "https://" + test.Host + "/gozaar/" + test.Actor
				if expectedSelf != aliases[0] {
					t.Errorf("For test #%d, the actual alias[0] is not what was expected.", testNumber)
					t.Logf("EXPECTED: %q", expectedSelf)
					t.Logf("ACTUAL:   %q", aliases[0])
					return
				}
			}

			// Check links
			{
				linksRaw, exists := jrd["links"]
				if !exists {
					t.Errorf("For test #%d, JRD missing 'links' field.", testNumber)
					return
				}
				links, ok := linksRaw.([]any)
				if !ok {
					t.Errorf("For test #%d, expected 'links' to be an array but got %T.", testNumber, linksRaw)
					return
				}
				if 3 != len(links) {
					t.Errorf("For test #%d, the actual number of links is not what was expected.", testNumber)
					t.Logf("EXPECTED: %d", 3)
					t.Logf("ACTUAL:   %d", len(links))
					return
				}

				expectedSelf := "https://" + test.Host + "/gozaar/" + test.Actor
				expectedOutbox := expectedSelf + "/outbox"

				// Link 0: self
				{
					link, ok := links[0].(map[string]any)
					if !ok {
						t.Errorf("For test #%d, expected link[0] to be object but got %T.", testNumber, links[0])
						return
					}
					if "self" != link["rel"] {
						t.Errorf("For test #%d, the actual link[0] 'rel' is not what was expected.", testNumber)
						t.Logf("EXPECTED: %q", "self")
						t.Logf("ACTUAL:   %q", link["rel"])
						return
					}
					if "application/activity+json" != link["type"] {
						t.Errorf("For test #%d, the actual link[0] 'type' is not what was expected.", testNumber)
						t.Logf("EXPECTED: %q", "application/activity+json")
						t.Logf("ACTUAL:   %q", link["type"])
						return
					}
					if expectedSelf != link["href"] {
						t.Errorf("For test #%d, the actual link[0] 'href' is not what was expected.", testNumber)
						t.Logf("EXPECTED: %q", expectedSelf)
						t.Logf("ACTUAL:   %q", link["href"])
						return
					}
				}

				// Link 1: outbox (activity+json)
				{
					link, ok := links[1].(map[string]any)
					if !ok {
						t.Errorf("For test #%d, expected link[1] to be object but got %T.", testNumber, links[1])
						return
					}
					if "outbox" != link["rel"] {
						t.Errorf("For test #%d, the actual link[1] 'rel' is not what was expected.", testNumber)
						t.Logf("EXPECTED: %q", "outbox")
						t.Logf("ACTUAL:   %q", link["rel"])
						return
					}
					if "application/activity+json" != link["type"] {
						t.Errorf("For test #%d, the actual link[1] 'type' is not what was expected.", testNumber)
						t.Logf("EXPECTED: %q", "application/activity+json")
						t.Logf("ACTUAL:   %q", link["type"])
						return
					}
					if expectedOutbox != link["href"] {
						t.Errorf("For test #%d, the actual link[1] 'href' is not what was expected.", testNumber)
						t.Logf("EXPECTED: %q", expectedOutbox)
						t.Logf("ACTUAL:   %q", link["href"])
						return
					}
				}

				// Link 2: outbox (text/event-stream)
				{
					link, ok := links[2].(map[string]any)
					if !ok {
						t.Errorf("For test #%d, expected link[2] to be object but got %T.", testNumber, links[2])
						return
					}
					if "outbox" != link["rel"] {
						t.Errorf("For test #%d, the actual link[2] 'rel' is not what was expected.", testNumber)
						t.Logf("EXPECTED: %q", "outbox")
						t.Logf("ACTUAL:   %q", link["rel"])
						return
					}
					if "text/event-stream" != link["type"] {
						t.Errorf("For test #%d, the actual link[2] 'type' is not what was expected.", testNumber)
						t.Logf("EXPECTED: %q", "text/event-stream")
						t.Logf("ACTUAL:   %q", link["type"])
						return
					}
					if expectedOutbox != link["href"] {
						t.Errorf("For test #%d, the actual link[2] 'href' is not what was expected.", testNumber)
						t.Logf("EXPECTED: %q", expectedOutbox)
						t.Logf("ACTUAL:   %q", link["href"])
						return
					}
				}
			}
		})
	}
}

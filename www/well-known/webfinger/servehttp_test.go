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

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			bytes, err := serveActorHost(test.Resource, test.Actor, test.Host)

			if test.ExpectError {
				if nil == err {
					t.Fatal("expected an error but got nil")
				}

				httpErr, ok := err.(errhttp.Error)
				if !ok {
					t.Fatalf("expected errhttp.Error but got %T: %v", err, err)
				}

				if test.ExpectStatusCode != httpErr.ErrHTTP() {
					t.Fatalf("expected HTTP status %d but got %d", test.ExpectStatusCode, httpErr.ErrHTTP())
				}

				return
			}

			if nil != err {
				t.Fatalf("unexpected error: %v", err)
			}

			if nil == bytes {
				t.Fatal("expected non-nil bytes but got nil")
			}

			var jrd map[string]any
			err = json.Unmarshal(bytes, &jrd)
			if nil != err {
				t.Fatalf("failed to unmarshal JRD JSON: %v", err)
			}

			// Check subject
			{
				subject, exists := jrd["subject"]
				if !exists {
					t.Fatal("JRD missing 'subject' field")
				}
				if test.Resource != subject {
					t.Fatalf("expected subject %q but got %q", test.Resource, subject)
				}
			}

			// Check aliases
			{
				aliasesRaw, exists := jrd["aliases"]
				if !exists {
					t.Fatal("JRD missing 'aliases' field")
				}
				aliases, ok := aliasesRaw.([]any)
				if !ok {
					t.Fatalf("expected aliases to be an array but got %T", aliasesRaw)
				}
				if 1 != len(aliases) {
					t.Fatalf("expected 1 alias but got %d", len(aliases))
				}

				expectedSelf := "https://" + test.Host + "/gozaar/" + test.Actor
				if expectedSelf != aliases[0] {
					t.Fatalf("expected alias %q but got %q", expectedSelf, aliases[0])
				}
			}

			// Check links
			{
				linksRaw, exists := jrd["links"]
				if !exists {
					t.Fatal("JRD missing 'links' field")
				}
				links, ok := linksRaw.([]any)
				if !ok {
					t.Fatalf("expected links to be an array but got %T", linksRaw)
				}
				if 3 != len(links) {
					t.Fatalf("expected 3 links but got %d", len(links))
				}

				expectedSelf := "https://" + test.Host + "/gozaar/" + test.Actor
				expectedOutbox := expectedSelf + "/outbox"

				// Link 0: self
				{
					link, ok := links[0].(map[string]any)
					if !ok {
						t.Fatalf("expected link[0] to be object but got %T", links[0])
					}
					if "self" != link["rel"] {
						t.Fatalf("expected link[0] rel 'self' but got %q", link["rel"])
					}
					if "application/activity+json" != link["type"] {
						t.Fatalf("expected link[0] type 'application/activity+json' but got %q", link["type"])
					}
					if expectedSelf != link["href"] {
						t.Fatalf("expected link[0] href %q but got %q", expectedSelf, link["href"])
					}
				}

				// Link 1: outbox (activity+json)
				{
					link, ok := links[1].(map[string]any)
					if !ok {
						t.Fatalf("expected link[1] to be object but got %T", links[1])
					}
					if "outbox" != link["rel"] {
						t.Fatalf("expected link[1] rel 'outbox' but got %q", link["rel"])
					}
					if "application/activity+json" != link["type"] {
						t.Fatalf("expected link[1] type 'application/activity+json' but got %q", link["type"])
					}
					if expectedOutbox != link["href"] {
						t.Fatalf("expected link[1] href %q but got %q", expectedOutbox, link["href"])
					}
				}

				// Link 2: outbox (text/event-stream)
				{
					link, ok := links[2].(map[string]any)
					if !ok {
						t.Fatalf("expected link[2] to be object but got %T", links[2])
					}
					if "outbox" != link["rel"] {
						t.Fatalf("expected link[2] rel 'outbox' but got %q", link["rel"])
					}
					if "text/event-stream" != link["type"] {
						t.Fatalf("expected link[2] type 'text/event-stream' but got %q", link["type"])
					}
					if expectedOutbox != link["href"] {
						t.Fatalf("expected link[2] href %q but got %q", expectedOutbox, link["href"])
					}
				}
			}
		})
	}
}

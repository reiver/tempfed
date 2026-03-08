package verboten

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"tempfed/srv/http"
)

const path string = "/inbox"

func init() {
	// Skip this if we are running inside of a Go test.
	if nil != flag.Lookup("test.v") || strings.HasSuffix(os.Args[0], ".test") {
		return
	}

	{
		var handler http.Handler = http.HandlerFunc(serveHTTP)

		err := httpsrv.Mux.HandlePath(handler, path)
		if nil != err {
			panic(err)
		}
	}
}

func serveHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if nil == responseWriter {
		return
	}

	http.Error(responseWriter, "403 Forbidden", http.StatusForbidden)
}

package verboten

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"io"
	"net/http"
	"os"
	"strings"

	"codeberg.org/reiver/go-field/stringly"
	"github.com/reiver/go-etag"
	"github.com/reiver/go-http500"

	"tempfed/srv/http"
	"tempfed/srv/log"
)

const path string = "/"

func init() {
	// Skip this if we are running inside of a Go test.
	if nil != flag.Lookup("test.v") || strings.HasSuffix(os.Args[0], ".test") {
		return
	}

	var handler http.Handler = http.HandlerFunc(serveHTTP)

	err := httpsrv.Mux.HandlePath(handler, path)
	if nil != err {
		panic(err)
	}
}

func serveHTTP(responsewriter http.ResponseWriter, request *http.Request) {
	log := logsrv.Begin()
	defer log.End()

	if nil == responsewriter {
		log.Error(stringly.S("nil response-writer"))
		return
	}
	if nil == request {
		http500.InternalServerError(responsewriter, request)
		log.Error(stringly.S("nil request"))
		return
	}

	var host string = request.Host
	if "" == host && nil != request.URL {
		host = request.URL.Host
	}
	if "" == host {
		http500.InternalServerError(responsewriter, request)
		log.Error(stringly.S("empty host"))
		return
	}

	var html string
	{
		const needle string = "{{HOST}}"
		html = strings.ReplaceAll(webpage, needle, host)
	}

	var digest string
	{
		digestBytes := sha256.Sum256([]byte(html))
		digest = hex.EncodeToString(digestBytes[:])
	}
	log.Debugf("digest: %s", digest)

	var eTag string = "sha256-" + digest
	log.Debugf("eTag: %s", eTag)

	if etag.Handle(responsewriter, request, eTag) {
		log.Debug(
			stringly.S("etag caching HIT"),
			stringly.String("host", host),
			stringly.String("path", path),
		)
		return
	} else {
		log.Debug(
			stringly.S("etag caching MISS"),
			stringly.String("host", host),
			stringly.String("path", path),
		)
	}

	_, err := io.WriteString(responsewriter, html)
	if nil != err {
		log.Errorf("problem writing HTML content to client: %s", err)
	}
}

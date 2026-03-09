package verboten

import (
	"flag"
	"io"
	"net/http"
	"os"
	"strings"

	gojson "encoding/json"

	"codeberg.org/reiver/go-asns"
	"codeberg.org/reiver/go-field/stringly"
	"github.com/reiver/go-http500"

	"tempfed/lib/db"
	"tempfed/srv/db"
	"tempfed/srv/http"
	"tempfed/srv/log"
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
	log := logsrv.Begin(stringly.String("www.path", path))
	defer log.End()

	if nil == responseWriter {
		log.Error(stringly.S("nil HTTP response-writer"))
		return
	}
	if nil == request {
		http500.InternalServerError(responseWriter, request)
		log.Error(stringly.S("nil HTTP request"))
		return
	}

	if http.MethodPost != request.Method {
		http.Error(responseWriter, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(request.Body)
	if nil != err {
		http500.InternalServerError(responseWriter, request)
		log.Error(
			stringly.S("failed to read request body"),
			stringly.E(err),
		)
		return
	}

	var activity asns.AnyActivity
	err = gojson.Unmarshal(body, &activity)
	if nil != err {
		http.Error(responseWriter, "400 Bad Request", http.StatusBadRequest)
		log.Error(
			stringly.S("failed to unmarshal activity"),
			stringly.E(err),
		)
		return
	}

	activityType := activity.Type.Strings()
	log.Trace(
		stringly.String("activity.type", strings.Join(activityType, ", ")),
		stringly.String("activity.id", activity.ID.GetElse("")),
	)

	if !hasType(activityType, asns.TypeCreate) {
		http.Error(responseWriter, "202 Accepted", http.StatusAccepted)
		log.Trace(stringly.S("ignoring non-Create activity"))
		return
	}

	if nil == activity.Object {
		http.Error(responseWriter, "400 Bad Request", http.StatusBadRequest)
		log.Error(stringly.S("Create activity has nil object"))
		return
	}

	obj, ok := activity.Object.(asns.ProtoObject)
	if !ok {
		http.Error(responseWriter, "400 Bad Request", http.StatusBadRequest)
		log.Error(stringly.S("Create activity object does not implement ProtoObject"))
		return
	}

	anyObj := obj.ProtoObject()

	err = libdb.InsertObject(request.Context(), dbsrv.Conn, anyObj)
	if nil != err {
		http500.InternalServerError(responseWriter, request)
		log.Error(
			stringly.S("failed to insert object into ClickHouse"),
			stringly.E(err),
		)
		return
	}

	log.Inform(
		stringly.S("inserted object from Create activity"),
		stringly.String("object.id", anyObj.ID.GetElse("")),
	)

	responseWriter.WriteHeader(http.StatusAccepted)
}

func hasType(types []string, target string) bool {
	for _, t := range types {
		if t == target {
			return true
		}
	}
	return false
}

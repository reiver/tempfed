package verboten

import (
	"flag"
	"io"
	"net/http"
	"os"
	"strings"

	gojson "encoding/json"

	"codeberg.org/reiver/go-activitypub"
	"codeberg.org/reiver/go-field"
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
	log := logsrv.Begin(field.String("www.path", path))
	defer log.End()

	if nil == responseWriter {
		log.Error(field.S("nil HTTP response-writer"))
		return
	}
	if nil == request {
		http500.InternalServerError(responseWriter, request)
		log.Error(field.S("nil HTTP request"))
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
			field.S("failed to read request body"),
			field.String("error", err.Error()),
		)
		return
	}

	var activity activitypub.AnyActivity
	err = gojson.Unmarshal(body, &activity)
	if nil != err {
		http.Error(responseWriter, "400 Bad Request", http.StatusBadRequest)
		log.Error(
			field.S("failed to unmarshal activity"),
			field.String("error", err.Error()),
		)
		return
	}

	activityType := activity.Type.Strings()
	log.Trace(
		field.String("activity.type", strings.Join(activityType, ", ")),
		field.String("activity.id", activity.ID.GetElse("")),
	)

	if !hasType(activityType, activitypub.TypeCreate) {
		http.Error(responseWriter, "202 Accepted", http.StatusAccepted)
		log.Trace(field.S("ignoring non-Create activity"))
		return
	}

	if nil == activity.Object {
		http.Error(responseWriter, "400 Bad Request", http.StatusBadRequest)
		log.Error(field.S("Create activity has nil object"))
		return
	}

	obj, ok := activity.Object.(activitypub.ProtoObject)
	if !ok {
		http.Error(responseWriter, "400 Bad Request", http.StatusBadRequest)
		log.Error(field.S("Create activity object does not implement ProtoObject"))
		return
	}

	anyObj := obj.ProtoObject()

	err = libdb.InsertObject(request.Context(), dbsrv.Conn, anyObj)
	if nil != err {
		http500.InternalServerError(responseWriter, request)
		log.Error(
			field.S("failed to insert object into ClickHouse"),
			field.String("error", err.Error()),
		)
		return
	}

	log.Inform(
		field.S("inserted object from Create activity"),
		field.String("object.id", anyObj.ID.GetElse("")),
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

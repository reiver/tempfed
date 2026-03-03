package verboten

import (
	"net/http"

	"codeberg.org/reiver/go-asns"
	"codeberg.org/reiver/go-field/stringly"
	"github.com/reiver/go-http404"
	"github.com/reiver/go-http500"
	"github.com/reiver/go-opt"
	"github.com/reiver/go-pathmux"

	"tempfed/lib/actors"
	"tempfed/lib/refs"
	"tempfed/srv/http"
	"tempfed/srv/log"
)

const pattern string = "/gozaar/{actorname}"

func init() {
	var handler pathmux.PatternHandler = pathmux.PatternHandlerFunc(serveHTTP)

	err := httpsrv.Mux.HandlePattern(handler, pattern)
	if nil != err {
		panic(err)
	}
}

func serveHTTP(responseWriter http.ResponseWriter, request *pathmux.ParameterizedRequest) {
	log := logsrv.Begin(stringly.String("www.pattern", pattern))
	defer log.End()

	if nil == responseWriter {
		log.Error(stringly.S("nil HTTP response-writer"))
		return
	}
	if nil == request {
		http500.InternalServerError(responseWriter, nil)
		log.Error(stringly.S("nil HTTP path-mux request"))
		return
	}

	actorName, found := request.ParameterByIndex(0)
	if !found {
		http500.InternalServerError(responseWriter, request.HTTPRequest())
		log.Error(stringly.S("missing 'actorname' (this should never happen)"))
		return
	}
	log.Trace(stringly.String("actor-name", actorName))

	if !libactors.IsValidUserName(actorName) {
		log.Warn(
			stringly.S("not found because invalid actor user-name"),
			stringly.String("actor-name", actorName),
		)
		http404.NotFound(responseWriter, request.HTTPRequest())
		return
	}

	{
		var host string = request.HTTPRequest().Host

		var (
			name    opt.Optional[string] = opt.Something(actorName)
			summary opt.Optional[string] = opt.Something("Search for: " + actorName)
		)

		var service = asns.Service{
			ID: opt.Something(librefs.Actor(host, actorName)),

			Name:    name,
			Summary: summary,

			EndPoints: asns.EndPoints{
				SharedInBox: opt.Something(librefs.SharedInBox(host)),
			},
			InBox:  opt.Something(librefs.ActorInBox(host, actorName)),
			OutBox: opt.Something(librefs.ActorOutBox(host, actorName)),
		}


		bytes, err := asns.Marshal(service)
		if nil != err {
			http500.InternalServerError(responseWriter, request.HTTPRequest())
			log.Error(
				stringly.S("failed to jsonld-marshal ActivityPub / ActivityStreams 'Service'"),
				stringly.E(err),
			)
			return
		}

		asns.ServeHTTP(responseWriter, request.HTTPRequest(), bytes)
	}
}

package verboten

import (
	"errors"
	"net/http"
	"strings"

	"codeberg.org/reiver/go-accturi"
	"codeberg.org/reiver/go-field/stringly"
	"codeberg.org/reiver/go-webfinger"
	"github.com/reiver/go-errhttp"
	"github.com/reiver/go-opt"

	"tempfed/lib/refs"
	"tempfed/srv/http"
	"tempfed/srv/log"
)

const path string = webfinger.DefaultPath // "/.well-known/webfinger"

func init() {
	var webFingerHandler webfinger.Handler = webfinger.HandlerFunc(serveWebFinger)
	var handler http.Handler = webfinger.HTTPHandler(webFingerHandler)

	err := httpsrv.Mux.HandlePath(handler, path)
	if nil != err {
		panic(err)
	}
}

func serveWebFinger(resource string, rels ...string) ([]byte, error) {
	log := logsrv.Begin()
	defer log.End()

	{
		actor, host, err := accturi.Split(resource)
		if nil == err {
			return serveActorHost(resource, actor, host)
		}
		if !errors.Is(err, accturi.ErrAcctURISchemeNotFound) {
			return nil, errhttp.Return(http.StatusBadRequest)
		}

		log.Tracef("actor, host = %q, %q", actor, host)
	}

	{
		//@TODO: handle other types of IRI/URI/URL schemes.
	}

	return nil, errhttp.Return(http.StatusNotFound)
}

func serveActorHost(resource string, actor string, host string) ([]byte, error) {
	log := logsrv.Begin()
	defer log.End()

	log.Trace(
		stringly.String("actor", actor),
		stringly.String("host", host),
	)

	if !strings.HasPrefix(actor, "search:") && !strings.HasPrefix(actor, "search-") {
		return nil, errhttp.Return(http.StatusNotFound)
	}

	//@TODO: put the `actor` into a canonical form.

	var (
		self   string = librefs.Actor(host, actor)
		outbox string = librefs.ActorOutBox(host, actor)
	)
	log.Trace(
		stringly.String("JRD-self", self),
		stringly.String("JRD-outbox", outbox),
	)

	// Return JRD (JSON Resource Descriptor) document,
	// that is expected to be returned in a WebFinger response.
	//
	// For example, for the resource "acct:joeblow:something@host.example",
	// this could return
	//
	//	{
	//		"subject" : "joeblow:something@host.example",
	//		"aliases" :
	//		[
	//			"https://host.example/gozaar/something"
	//		],
	//		"links"   :
	//		[
	//			{
	//				"rel"  : "self",
	//				"type" : "application/activity+json",
	//				"href" : "https://host.example/gozaar/something",
	//			},
	//			{
	//				"rel"  : "outbox",
	//				"type" : "application/activity+json",
	//				"href" : "https://host.example/gozaar/something/outbox",
	//			},
	//			{
	//				"rel"  : "outbox",
	//				"type" : "text/event-stream",
	//				"href" : "https://host.example/gozaar/something/outbox",
	//			}
	//		],
	//	}
	{
		var response webfinger.Response = webfinger.Response{
			Subject: opt.Something(resource),
			Aliases: []string{
				self,
			},
			Links: []webfinger.Link{
				webfinger.Link{
					Rel:  opt.Something("self"),
					Type: opt.Something("application/activity+json"),
					HRef: opt.Something(self),
				},
				webfinger.Link{
					Rel:  opt.Something("outbox"),
					Type: opt.Something("application/activity+json"),
					HRef: opt.Something(outbox),
				},
				webfinger.Link{
					Rel:  opt.Something("outbox"),
					Type: opt.Something("text/event-stream"),
					HRef: opt.Something(outbox),
				},
			},
		}

		return response.MarshalJSON()
	}
}

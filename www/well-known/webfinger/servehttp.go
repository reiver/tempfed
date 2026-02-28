package verboten

import (
	"errors"
	"net/http"
	gourl "net/url"
	"strings"

	"codeberg.org/reiver/go-accturi"
	"codeberg.org/reiver/go-field/stringly"
	"codeberg.org/reiver/go-webfinger"
	"github.com/reiver/go-errhttp"
	"github.com/reiver/go-opt"

	"tempfed/srv/http"
	"tempfed/srv/log"
)

const path string = webfinger.DefaultPath

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
	}

	{
		//@TODO: handle other types of IRI/URI/URL schemes.
	}

	return nil, errhttp.Return(http.StatusNotFound)
}

func serveActorHost(resource string, actor string, host string) ([]byte, error) {
	log := logsrv.Begin()
	defer log.End()

	if !strings.HasPrefix(actor, "search:") && !strings.HasPrefix(actor, "search-") {
		return nil, errhttp.Return(http.StatusNotFound)
	}

	var url = gourl.URL{
		Scheme: "https",
		Host: host,
		Path: "/" + actor,
	}

	var href string = url.String()
	log.Trace(stringly.String("href", href))

	url.Path += "/outbox"

	var outbox string = url.String()

	var response webfinger.Response = webfinger.Response{
		Subject: opt.Something(resource),
		Aliases: []string{
			href,
		},
		Links: []webfinger.Link{
			webfinger.Link{
				Rel:  opt.Something("self"),
				Type: opt.Something("application/activity+json"),
				HRef: opt.Something(href),
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

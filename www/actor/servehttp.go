package verboten

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"codeberg.org/reiver/go-asns"
	"codeberg.org/reiver/go-asns/ns/sec1"
	"codeberg.org/reiver/go-field/stringly"
	"github.com/reiver/go-http500"
	"github.com/reiver/go-opt"

	"tempfed/cfg"
	"tempfed/lib/refs"
	"tempfed/srv/http"
	"tempfed/srv/log"
)

const path string = "/actor"

var publicKeyPem string

func init() {
	// Skip this if we are running inside of a Go test.
	if nil != flag.Lookup("test.v") || strings.HasSuffix(os.Args[0], ".test") {
		return
	}

	{
		var keyFile string = cfg.RelayPublicKeyFileName()
		if "" != keyFile {
			data, err := os.ReadFile(keyFile)
			if nil != err {
				panic("failed to read relay public key file: " + err.Error())
			}
			publicKeyPem = string(data)
		}
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

	var host string = request.Host

	var application = asns.Application{
		ID: opt.Something(librefs.RelayActor(host)),

		PreferredUserName: asns.SomeString("relay"),

		InBox: opt.Something(librefs.RelayActorInBox(host)),

		Followers: opt.Something(librefs.RelayActorFollowers(host)),
		Following: opt.Something(librefs.RelayActorFollowing(host)),

		EndPoints: asns.EndPoints{
			SharedInBox: opt.Something(librefs.RelayActorInBox(host)),
		},
	}

	var marshalArgs []any
	marshalArgs = append(marshalArgs, application)

	if "" != publicKeyPem {
		var security = sec1.Security{
			PublicKey: opt.Something(sec1.PublicKey{
				ID:           opt.Something(librefs.RelayActor(host) + "#main-key"),
				Owner:        opt.Something(librefs.RelayActor(host)),
				PublicKeyPem: opt.Something(publicKeyPem),
			}),
		}
		marshalArgs = append(marshalArgs, security)
	}

	bytes, err := asns.Marshal(marshalArgs...)
	if nil != err {
		http500.InternalServerError(responseWriter, request)
		log.Error(
			stringly.S("failed to jsonld-marshal ActivityPub / ActivityStreams 'Application'"),
			stringly.E(err),
		)
		return
	}

	asns.ServeHTTP(responseWriter, request, bytes)
}

package verboten

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"codeberg.org/reiver/go-activitypub"
	"codeberg.org/reiver/go-field"
	"codeberg.org/reiver/go-log"
	"github.com/reiver/go-http404"
	"github.com/reiver/go-http500"
	"github.com/reiver/go-opt"
	"github.com/reiver/go-nul"
	"github.com/reiver/go-pathmux"

	"tempfed/lib/actors"
	"tempfed/lib/refs"
	"tempfed/srv/db"
	"tempfed/srv/http"
	"tempfed/srv/log"
)

const pattern string = "/gozaar/{actorname}/outbox"

const pageSize int = 20

func init() {
	// Skip this if we are running inside of a Go test.
	if nil != flag.Lookup("test.v") || strings.HasSuffix(os.Args[0], ".test") {
		return
	}

	var handler pathmux.PatternHandler = pathmux.PatternHandlerFunc(serveHTTP)

	err := httpsrv.Mux.HandlePattern(handler, pattern)
	if nil != err {
		panic(err)
	}
}

func serveHTTP(responseWriter http.ResponseWriter, request *pathmux.ParameterizedRequest) {
	logger := logsrv.Begin(field.String("www.pattern", pattern))
	defer logger.End()

	if nil == responseWriter {
		logger.Error(field.S("nil HTTP response-writer"))
		return
	}
	if nil == request {
		http500.InternalServerError(responseWriter, nil)
		logger.Error(field.S("nil HTTP path-mux request"))
		return
	}

	actorName, found := request.ParameterByIndex(0)
	if !found {
		http500.InternalServerError(responseWriter, request.HTTPRequest())
		logger.Error(field.S("missing 'actorname' (this should never happen)"))
		return
	}
	logger.Trace(field.String("actor-name", actorName))


	var keyword string
	{
		serviceName, parameter, split := libactors.Split(actorName)
		if !split {
			http404.NotFound(responseWriter, request.HTTPRequest())
			logger.Error(field.S("not found because invalid actor user-name"))
			return
		}

		logger.Trace(
			field.String("service-name", serviceName),
			field.String("parameter", parameter),
		)

		keyword = parameter
	}
	logger.Trace(field.String("keyword", keyword))

	var host string = request.HTTPRequest().Host
	var outboxURL string = librefs.ActorOutBox(host, actorName)

	pageParam := request.HTTPRequest().URL.Query().Get("page")
	logger.Trace(field.String("page", pageParam))

	switch {
	case "" == pageParam :
		logger.Tracef("collection")
		serveCollection(responseWriter, request, logger, keyword, outboxURL)
	default:
		page, err := strconv.Atoi(pageParam)
		if nil != err || page < 1 {
			http404.NotFound(responseWriter, request.HTTPRequest())
			return
		}
		logger.Tracef("collection-page")
		servePage(responseWriter, request, logger, keyword, outboxURL, page)
	}
}

func serveCollection(responseWriter http.ResponseWriter, request *pathmux.ParameterizedRequest, logger log.Logger, keyword string, outboxURL string) {
	logger = logsrv.Begin()
	defer logger.End()

	ctx := request.HTTPRequest().Context()

	var count uint64
	err := dbsrv.Conn.QueryRow(ctx,
		`SELECT count() FROM data_nodes WHERE as_content LIKE concat('%', ?, '%') OR has(hashtags, ?)`,
		keyword, keyword,
	).Scan(&count)
	if nil != err {
		http500.InternalServerError(responseWriter, request.HTTPRequest())
		logger.Error(
			field.S("failed to query count from ClickHouse"),
			field.String("error", err.Error()),
		)
		return
	}

	lastPage := int(count) / pageSize
	if int(count)%pageSize != 0 {
		lastPage++
	}
	if lastPage < 1 {
		lastPage = 1
	}

	var collection = activitypub.OrderedCollection{
		ID: opt.Something(outboxURL),
		CoreCollection: activitypub.CoreCollection{
			TotalItems: nul.Something(activitypub.WholeNumber(uint64(count))),
			First:      activitypub.HRef(fmt.Sprintf("%s?page=1", outboxURL)),            //@TODO: construct URL in safer way
			Last:       activitypub.HRef(fmt.Sprintf("%s?page=%d", outboxURL, lastPage)), //@TODO: construct URL in safer way
		},
	}

	bytes, err := activitypub.Marshal(collection)
	if nil != err {
		http500.InternalServerError(responseWriter, request.HTTPRequest())
		logger.Error(
			field.S("failed to jsonld-marshal OrderedCollection"),
			field.String("error", err.Error()),
		)
		return
	}

	activitypub.ServeHTTP(responseWriter, request.HTTPRequest(), bytes)
}

func servePage(responseWriter http.ResponseWriter, request *pathmux.ParameterizedRequest, logger log.Logger, keyword string, outboxURL string, page int) {
	logger = logsrv.Begin()
	defer logger.End()

	ctx := request.HTTPRequest().Context()

	var count uint64
	err := dbsrv.Conn.QueryRow(ctx,
		`SELECT count() FROM data_nodes WHERE as_content LIKE concat('%', ?, '%') OR has(hashtags, ?)`,
		keyword, keyword,
	).Scan(&count)
	if nil != err {
		http500.InternalServerError(responseWriter, request.HTTPRequest())
		logger.Error(
			field.S("failed to query count from ClickHouse"),
			field.String("error", err.Error()),
		)
		return
	}

	lastPage := int(count) / pageSize
	if 0 != int(count)%pageSize {
		lastPage++
	}
	if lastPage < 1 {
		lastPage = 1
	}

	offset := (page - 1) * pageSize

	rows, err := dbsrv.Conn.Query(ctx,
		`SELECT as_id, as_type,
		        as_name, as_summary, as_content, as_media_type, as_url,
		        as_attributed_to, as_to, as_cc, as_audience,
		        as_published, as_updated, as_start_time, as_end_time, as_duration,
		        hashtags, as_in_reply_to, as_also_known_as, as_moved_to
		 FROM data_nodes
		 WHERE as_content LIKE concat('%', ?, '%') OR has(hashtags, ?)
		 ORDER BY as_published DESC
		 LIMIT ? OFFSET ?`,
		keyword, keyword, pageSize, offset,
	)
	if nil != err {
		http500.InternalServerError(responseWriter, request.HTTPRequest())
		logger.Error(
			field.S("failed to query data_nodes from ClickHouse"),
			field.String("error", err.Error()),
		)
		return
	}
	defer rows.Close()

	var orderedItems []any

	for rows.Next() {
		var (
			asID           string
			asType         []string
			asName         *string
			asSummary      *string
			asContent      *string
			asMediaType    *string
			asURL          *string
			asAttributedTo []string
			asTo           []string
			asCC           []string
			asAudience     []string
			asPublished    *time.Time
			asUpdated      *time.Time
			asStartTime    *time.Time
			asEndTime      *time.Time
			asDuration     *string
			hashtags       []string
			asInReplyTo    []string
			asAlsoKnownAs  []string
			asMovedTo      *string
		)

		err := rows.Scan(
			&asID, &asType,
			&asName, &asSummary, &asContent, &asMediaType, &asURL,
			&asAttributedTo, &asTo, &asCC, &asAudience,
			&asPublished, &asUpdated, &asStartTime, &asEndTime, &asDuration,
			&hashtags, &asInReplyTo, &asAlsoKnownAs, &asMovedTo,
		)
		if nil != err {
			http500.InternalServerError(responseWriter, request.HTTPRequest())
			logger.Error(
				field.S("failed to scan row from ClickHouse"),
				field.String("error", err.Error()),
			)
			return
		}

		obj := activitypub.AnyObject{
			ID:           opt.Something(asID),
//@TODO: Tyoe
			CoreEntity: activitypub.CoreEntity{
				Name: optFromPtrFromOptional(asName),
			},
			CoreObject: activitypub.CoreObject{
				Summary:      optFromPtrFromNullable(asSummary),
				Content:      optFromPtrFromOptional(asContent),
				MediaType:    optFromPtrFromOptional(asMediaType),
				URL:          optFromPtrFromOptional(asURL),
				AttributedTo: stringsToObjectIDs(asAttributedTo),
				To:           activitypub.SomeStrings(asTo...),
				CC:           activitypub.SomeStrings(asCC...),
				Audiences:    stringsToObjectIDs(asAudience),
				Published:    optTimeToString(asPublished),
				Updated:      optTimeToString(asUpdated),
				StartTime:    optTimeToString(asStartTime),
				EndTime:      optTimeToString(asEndTime),
				Duration:     optFromPtrFromOptional(asDuration),
				InReplyTo:    activitypub.SomeStrings(asInReplyTo...),
				AlsoKnownAs:  activitypub.SomeStrings(asAlsoKnownAs...),
				MovedTo:      optFromPtrFromOptional(asMovedTo),
			},
		}

		orderedItems = append(orderedItems, obj)
	}

	if nil != rows.Err() {
		http500.InternalServerError(responseWriter, request.HTTPRequest())
		logger.Error(
			field.S("error iterating ClickHouse rows"),
			field.String("error", rows.Err().Error()),
		)
		return
	}

	collectionPage := activitypub.OrderedCollectionPage{
		ID: opt.Something(fmt.Sprintf("%s?page=%d", outboxURL, page)),
		CoreCollectionPage: activitypub.CoreCollectionPage{
			PartOf: activitypub.HRef(outboxURL),
		},
		CoreOrderedCollection: activitypub.CoreOrderedCollection{
			OrderedItems: orderedItems,
		},
	}

	if page > 1 {
		collectionPage.Prev = activitypub.HRef(fmt.Sprintf("%s?page=%d", outboxURL, page-1))
	}
	if page < lastPage {
		collectionPage.Next = activitypub.HRef(fmt.Sprintf("%s?page=%d", outboxURL, page+1))
	}

	bytes, err := activitypub.Marshal(collectionPage)
	if nil != err {
		http500.InternalServerError(responseWriter, request.HTTPRequest())
		logger.Error(
			field.S("failed to jsonld-marshal OrderedCollectionPage"),
			field.String("error", err.Error()),
		)
		return
	}

	activitypub.ServeHTTP(responseWriter, request.HTTPRequest(), bytes)
}

func optFromPtrFromNullable(p *string) nul.Nullable[string] {
	if nil == p {
		return nul.Nothing[string]()
	}
	return nul.Something(*p)
}

func optFromPtrFromOptional(p *string) opt.Optional[string] {
	if nil == p {
		return opt.Nothing[string]()
	}
	return opt.Something(*p)
}

func optTimeToString(p *time.Time) opt.Optional[string] {
	if nil == p {
		return opt.Nothing[string]()
	}
	return opt.Something(p.Format(time.RFC3339))
}

func stringsToObjectIDs(ss []string) []activitypub.ProtoObjectOrProtoLink {
	var result []activitypub.ProtoObjectOrProtoLink
	for _, s := range ss {
		result = append(result, activitypub.ObjectID(s))
	}
	return result
}

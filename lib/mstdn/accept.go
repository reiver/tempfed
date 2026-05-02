package libmstdn

import (
	"context"
	"fmt"
	"strings"
	"time"

	"codeberg.org/reiver/go-activitypub"
	"codeberg.org/reiver/go-field"
	"codeberg.org/reiver/go-log"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/reiver/go-mstdn/api/v1/streaming/public"
	"github.com/reiver/go-nul"
	"github.com/reiver/go-opt"
)

const insertSQL = `INSERT INTO data_nodes (
	as_id, as_type,
	as_name, as_summary, as_content, as_media_type, as_url,
	as_attributed_to, as_to, as_cc, as_audience,
	as_published, as_updated, as_start_time, as_end_time, as_duration,
	hashtags,
	as_in_reply_to, as_also_known_as, as_moved_to
)`

func Accept(ctx context.Context, logger log.Logger, host string, conn driver.Conn, batchSize int, flushInterval time.Duration, statsInterval time.Duration) {
	log := logger.Begin()
	defer log.End()

	client, err := public.DialHost(host)
	if nil != err {
		log.Error(
			field.String("", "failed to connect to mstdn server"),
			field.String("host", host),
			field.String("error", err.Error()),
		)
		return
	}
	defer func() {
		err := client.Close()
		if nil != err {
			log.Error(
				field.String("", "failed to close mstdn local stream"),
				field.String("error", err.Error()),
			)
		}
	}()

	batch, err := conn.PrepareBatch(ctx, insertSQL)
	if nil != err {
		log.Error(
			field.String("", "failed to prepare batch"),
			field.String("error", err.Error()),
		)
		return
	}

	rowCount := 0
	lastFlush := time.Now()

	var postsReceived int
	var postsInserted int
	var postsDropped int
	var batchesFlushed int
	lastStatsLog := time.Now()

	var totalReceived int
	var totalInserted int
	var totalDropped int
	var totalBatches int

	logStats := func() {
		if time.Since(lastStatsLog) < statsInterval {
			return
		}
		log.Inform(
			field.String("", "stats"),
			field.String("host", host),
			field.String("received", fmt.Sprintf("%d", postsReceived)),
			field.String("inserted", fmt.Sprintf("%d", postsInserted)),
			field.String("dropped", fmt.Sprintf("%d", postsDropped)),
			field.String("batches", fmt.Sprintf("%d", batchesFlushed)),
		)
		totalReceived += postsReceived
		totalInserted += postsInserted
		totalDropped += postsDropped
		totalBatches += batchesFlushed
		postsReceived = 0
		postsInserted = 0
		postsDropped = 0
		batchesFlushed = 0
		lastStatsLog = time.Now()
	}

	for client.Next() {
		if nil != ctx.Err() {
			break
		}

		var event public.Event

		postsReceived++

		err := client.Decode(&event)
		if nil != err {
			log.Error(
				field.String("", "failed to decode event"),
				field.String("error", err.Error()),
			)
			postsDropped++
			logStats()
			continue
		}

		var note activitypub.Note
		err = event.Status.ActivityNote(&note)
		if nil != err {
			postsDropped++
			logStats()
			continue
		}

		logger.Trace(
			field.String("event.Name", event.Name),
			field.String("event.Status.ID", event.Status.ID.GetElse("")),
			field.String("event.Status.CreatedAt", event.Status.CreatedAt.GetElse("")),
			field.String("event.Status.URL", event.Status.URL.GetElse("")),
			field.String("event.Status.URI", event.Status.URI.GetElse("")),
			field.String("note.ID", note.ID.GetElse("")),
			field.String("note.AttributedTo", strings.Join(stringifySlice(note.AttributedTo), ", ")),
			field.String("note.Content", note.Content.GetElse("")),
		)

		var tagNames []string
		for _, tag := range note.Tags {
			hashtag, casted := tag.(activitypub.HashTag)
			if !casted {
				continue
			}

			name, found := hashtag.Name.Get()
			if !found {
				continue
			}

			tagNames = append(tagNames, name)
		}

		err = batch.Append(
			note.ID.GetElse(""),
			[]string{"Note"},
			nullableStrFromOptional(note.Name),
			nullableStrFromNullable(note.Summary),
			nullableStrFromOptional(note.Content),
			nullableStrFromOptional(note.MediaType),
			nullableStrFromOptional(note.URL),
			stringifySlice(note.AttributedTo),
			note.To.Strings(),
			note.CC.Strings(),
			stringifySlice(note.Audiences),
			nullableTime(note.Published),
			nullableTime(note.Updated),
			nullableTime(note.StartTime),
			nullableTime(note.EndTime),
			nullableStrFromOptional(note.Duration),
			tagNames,
			note.InReplyTo.Strings(),
			note.AlsoKnownAs.Strings(),
			nullableStrFromOptional(note.MovedTo),
		)
		if nil != err {
			log.Error(
				field.String("", "failed to append row"),
				field.String("error", err.Error()),
			)
			postsDropped++
			logStats()
			continue
		}

		rowCount++

		if rowCount >= batchSize || time.Since(lastFlush) >= flushInterval {
			flushedCount := rowCount
			var ok bool
			batch, rowCount, lastFlush, ok = flush(ctx, log, conn, batch, rowCount)
			if ok {
				postsInserted += flushedCount
				batchesFlushed++
			} else {
				postsDropped += flushedCount
			}
		}

		logStats()
	}

	// flush any remaining buffered rows
	if rowCount > 0 {
		flushedCount := rowCount
		_, _, _, ok := flush(ctx, log, conn, batch, flushedCount)
		if ok {
			postsInserted += flushedCount
			batchesFlushed++
		} else {
			postsDropped += flushedCount
		}
	}

	// log final cumulative stats on shutdown
	totalReceived += postsReceived
	totalInserted += postsInserted
	totalDropped += postsDropped
	totalBatches += batchesFlushed
	log.Inform(
		field.String("", "total-stats"),
		field.String("host", host),
		field.String("received", fmt.Sprintf("%d", totalReceived)),
		field.String("inserted", fmt.Sprintf("%d", totalInserted)),
		field.String("dropped", fmt.Sprintf("%d", totalDropped)),
		field.String("batches", fmt.Sprintf("%d", totalBatches)),
	)

	if err := client.Err(); nil != err {
		log.Error(
			field.String("", "post-stream error"),
			field.String("error", err.Error()),
		)
	}
}

func flush(ctx context.Context, log log.Logger, conn driver.Conn, batch driver.Batch, rowCount int) (driver.Batch, int, time.Time, bool) {
	err := batch.Send()
	if nil != err {
		log.Error(
			field.String("", "failed to send batch"),
			field.String("error", err.Error()),
		)
	}
	sendOK := nil == err

	newBatch, err := conn.PrepareBatch(ctx, insertSQL)
	if nil != err {
		log.Error(
			field.String("", "failed to prepare batch"),
			field.String("error", err.Error()),
		)
		return batch, 0, time.Now(), sendOK
	}

	return newBatch, 0, time.Now(), sendOK
}

func stringifySlice(items []activitypub.ProtoObjectOrProtoLink) []string {
	var result []string
	for _, item := range items {
		if s, ok := item.(fmt.Stringer); ok {
			result = append(result, s.String())
		}
	}
	return result
}

func nullableStrFromOptional(o opt.Optional[string]) *string {
	v, ok := o.Get()
	if !ok {
		return nil
	}
	return &v
}

func nullableStrFromNullable(o nul.Nullable[string]) *string {
	v, ok := o.Get()
	if !ok {
		return nil
	}
	return &v
}

func nullableTime(o opt.Optional[string]) *time.Time {
	v, ok := o.Get()
	if !ok {
		return nil
	}
	t, err := time.Parse(time.RFC3339, v)
	if nil != err {
		return nil
	}
	return &t
}

package libmstdn

import (
	"context"
	"time"

	"codeberg.org/reiver/go-asns"
	"codeberg.org/reiver/go-field/stringly"
	"codeberg.org/reiver/go-log"
	"github.com/reiver/go-mstdn/api/v1/streaming/public"
	"github.com/reiver/go-opt"

	"tempfed/srv/db"
)

func Accept(ctx context.Context, logger log.Logger, host string) {
	log := logger.Begin()
	defer log.End()

	client, err := public.DialHost(host)
	if nil != err {
		log.Error(
			stringly.String("", "failed to connect to mstdn server"),
			stringly.String("host", host),
			stringly.Error("error", err),
		)
		return
	}
	defer func() {
		err := client.Close()
		if nil != err {
			log.Error(
				stringly.String("", "failed to close mstdn local stream"),
				stringly.Error("error", err),
			)
		}
	}()

	for client.Next() {
		if nil != ctx.Err() {
			return
		}
		var event public.Event

		err := client.Decode(&event)
		if nil != err {
			log.Error(
				stringly.String("", "failed to decode event"),
				stringly.Error("error", err),
			)
			continue
		}

		var note asns.Note
		err = event.Status.ActivityNote(&note)
		if nil != err {
			continue
		}

		logger.Trace(
			stringly.String("event.Name", event.Name),
			stringly.String("event.Status.ID", event.Status.ID.GetElse("")),
			stringly.String("event.Status.CreatedAt", event.Status.CreatedAt.GetElse("")),
			stringly.String("event.Status.URL", event.Status.URL.GetElse("")),
			stringly.String("event.Status.URI", event.Status.URI.GetElse("")),
			stringly.String("note.ID", note.ID.GetElse("")),
			stringly.String("note.Content", note.Content.GetElse("")),
		)

		var tagNames []string
		for _, tag := range note.Tags {
			hashtag, casted := tag.(asns.HashTag)
			if !casted {
				continue
			}

			name, found := hashtag.Name.Get()
			if !found {
				continue
			}

			tagNames = append(tagNames, name)
		}

		batch, err := dbsrv.Conn.PrepareBatch(ctx, `INSERT INTO data_nodes (
			as_id, as_type,
			as_name, as_summary, as_content, as_media_type, as_url,
			as_attributed_to, as_to, as_cc, as_audience,
			as_published, as_updated, as_start_time, as_end_time, as_duration,
			hashtags,
			as_in_reply_to, as_also_known_as, as_moved_to
		)`)
		if nil != err {
			log.Error(
				stringly.String("", "failed to prepare batch"),
				stringly.Error("error", err),
			)
			continue
		}

		err = batch.Append(
			note.ID.GetElse(""),
			[]string{"Note"},
			nullableStr(note.Name),
			nullableStr(note.Summary),
			nullableStr(note.Content),
			nullableStr(note.MediaType),
			nullableStr(note.URL),
			note.AttributedTo.Strings(),
			note.To.Strings(),
			note.CC.Strings(),
			note.Audiences.Strings(),
			nullableTime(note.Published),
			nullableTime(note.Updated),
			nullableTime(note.StartTime),
			nullableTime(note.EndTime),
			nullableStr(note.Duration),
			tagNames,
			note.InReplyTo.Strings(),
			note.AlsoKnownAs.Strings(),
			nullableStr(note.MovedTo),
		)
		if nil != err {
			log.Error(
				stringly.String("", "failed to append row"),
				stringly.Error("error", err),
			)
			continue
		}

		err = batch.Send()
		if nil != err {
			log.Error(
				stringly.String("", "failed to send batch"),
				stringly.Error("error", err),
			)
			continue
		}
	}
	if err := client.Err(); nil != err {
		log.Error(
			stringly.String("", "post-stream error"),
			stringly.Error("error", err),
		)
	}
}

func nullableStr(o opt.Optional[string]) *string {
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

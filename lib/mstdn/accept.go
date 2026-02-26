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

func Accept(logger log.Logger, host string) {
	log := logger.Begin()
	defer log.End()

	client, err := public.DialHost(host)
	if nil != err {
		log.Fatal(
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

	ctx := context.Background()

	for client.Next() {
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

		logger.Debug(
			stringly.String("note.Content", note.Content.GetElse("")),
		)

		batch, err := dbsrv.Conn.PrepareBatch(ctx, `INSERT INTO data_nodes (
			id, type,
			name, summary, content, media_type, url,
			attributed_to, to, cc, audience,
			published, updated, start_time, end_time, duration,
			in_reply_to, also_known_as, moved_to
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

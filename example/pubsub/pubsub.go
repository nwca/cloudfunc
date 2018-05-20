package hello

import (
	"context"
	"os"

	"time"

	"cloud.google.com/go/pubsub"
)

var (
	cli  *pubsub.Client
	gerr error
)

func init() {
	cli, gerr = pubsub.NewClient(context.Background(), os.Getenv("GCLOUD_PROJECT"))
}

var start = time.Now()

func HandleTopic(ctx context.Context, m *pubsub.Message) error {
	if gerr != nil {
		return gerr
	}
	m.Attributes["uptime"] = time.Since(start).String()
	_, err := cli.Topic("test").Publish(ctx, m).Get(ctx)
	return err
}

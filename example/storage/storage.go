package hello

import (
	"context"
	"os"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
)

var (
	cli  *pubsub.Client
	gerr error
)

func init() {
	cli, gerr = pubsub.NewClient(context.Background(), os.Getenv("GCLOUD_PROJECT"))
}

func HandleStorage(ctx context.Context, attrs *storage.ObjectAttrs) error {
	_, err := cli.Topic("test").Publish(ctx, &pubsub.Message{
		Data: attrs.MD5, Attributes: map[string]string{
			"name": attrs.Name,
		},
	}).Get(ctx)
	return err
}

package gcp

import (
	"context"
	"fmt"
	"io"
	"math/rand"

	"cloud.google.com/go/longrunning"
	longauto "cloud.google.com/go/longrunning/autogen"
	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	funcs "google.golang.org/genproto/googleapis/cloud/functions/v1beta2"
	longpb "google.golang.org/genproto/googleapis/longrunning"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const region = "us-central1"

const scope = "https://www.googleapis.com/auth/cloudfunctions"

func defaultFuncsClientOptions() []option.ClientOption {
	return []option.ClientOption{
		option.WithEndpoint("cloudfunctions.googleapis.com:443"),
		option.WithScopes([]string{
			"https://www.googleapis.com/auth/cloud-platform",
			scope,
		}...),
	}
}

func NewClient(project string) (*Client, error) {
	ctx := context.Background()
	conn, err := transport.DialGRPC(ctx, defaultFuncsClientOptions()...)
	if err != nil {
		return nil, err
	}
	cli := funcs.NewCloudFunctionsServiceClient(conn)
	scli, err := storage.NewClient(ctx)
	if err != nil {
		conn.Close()
		return nil, err
	}
	long, err := longauto.NewOperationsClient(ctx, defaultFuncsClientOptions()...)
	if err != nil {
		conn.Close()
		scli.Close()
		return nil, err
	}
	return &Client{
		project: project,
		funcs:   cli,
		storage: scli,
		long:    long,
		conn:    conn,
		region:  region,
	}, nil
}

type Client struct {
	project string
	region  string
	conn    io.Closer
	funcs   funcs.CloudFunctionsServiceClient
	storage *storage.Client
	long    *longauto.OperationsClient
	staging string
}

func (c *Client) Close() error {
	c.conn.Close()
	c.storage.Close()
	c.long.Close()
	return nil
}

func (c *Client) getBucket(ctx context.Context) (string, error) {
	if c.staging != "" {
		return c.staging, nil
	}
	name := c.project + "-staging"
	_, err := c.storage.Bucket(name).Attrs(ctx)
	if err == nil {
		c.staging = name
		return name, nil
	} else if err != nil && err != storage.ErrBucketNotExist {
		return "", err
	}
	err = c.storage.Bucket(name).Create(ctx, c.project, &storage.BucketAttrs{
		Name: name, StorageClass: "REGIONAL", Location: c.region,
	})
	if err != nil {
		return "", err
	}
	c.staging = name
	return name, nil
}

func (c *Client) functionID(name string) string {
	return "projects/" + c.project + "/locations/" + c.region + "/functions/" + name
}

func (c *Client) Deploy(ctx context.Context, name string, tr Trigger, r io.Reader) error {
	f, err := c.funcs.GetFunction(ctx, &funcs.GetFunctionRequest{
		Name: c.functionID(name),
	})
	create := false
	if status.Code(err) == codes.NotFound {
		f = &funcs.CloudFunction{
			Name:       c.functionID(name),
			EntryPoint: "helloWorld",
		}
		create = true
	} else if err != nil {
		return err
	}
	tr.setOn(c.project, f)

	staging, err := c.getBucket(ctx)
	if err != nil {
		return err
	}
	b := c.storage.Bucket(staging)
	fname := fmt.Sprintf("%s-%d.zip", name, rand.Int())

	wctx, cancel := context.WithCancel(ctx)
	w := b.Object(fname).NewWriter(wctx)
	_, err = io.Copy(w, r)
	if err != nil {
		cancel()
		return err
	}
	err = w.Close()
	cancel()
	if err != nil {
		return err
	}
	defer b.Object(fname).Delete(ctx)
	f.SourceCode = &funcs.CloudFunction_SourceArchiveUrl{
		SourceArchiveUrl: "gs://" + staging + "/" + fname,
	}

	var (
		oppb *longpb.Operation
	)
	if create {
		oppb, err = c.funcs.CreateFunction(ctx, &funcs.CreateFunctionRequest{
			Location: "projects/" + c.project + "/locations/" + c.region,
			Function: f,
		})
	} else {
		oppb, err = c.funcs.UpdateFunction(ctx, &funcs.UpdateFunctionRequest{
			Name:     c.functionID(name),
			Function: f,
		})
	}
	if err != nil {
		return fmt.Errorf("cannot update function: %v", err)
	}
	op := longrunning.InternalNewOperation(c.long, oppb)
	if err = op.Wait(ctx, nil); err != nil {
		return fmt.Errorf("cannot check operation state: %v", err)
	}
	return nil
}

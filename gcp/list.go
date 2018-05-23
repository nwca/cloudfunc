package gcp

import (
	"context"

	"google.golang.org/api/transport"
	funcs "google.golang.org/genproto/googleapis/cloud/functions/v1beta2"
)

type CloudFunction = funcs.CloudFunction

func (c *Client) ListFuncs() ([]*CloudFunction, error) {
	ctx := context.Background()
	conn, err := transport.DialGRPC(ctx, defaultFuncsClientOptions()...)
	if err != nil {
		return nil, err
	}
	cli := funcs.NewCloudFunctionsServiceClient(conn)
	resp, err := cli.ListFunctions(ctx, &funcs.ListFunctionsRequest{
		Location: "projects/" + c.project + "/locations/-",
	})
	if err != nil {
		return nil, err
	}
	return resp.Functions, nil
}

// +build pubsub

package main

import (
	"context"
	"encoding/json"
	"net/http"

	"cloud.google.com/go/pubsub"
)

type PubSubFunc func(ctx context.Context, m *pubsub.Message) error

func HandlePubSub(fnc PubSubFunc) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		// /execute/_ah/push-handlers/pubsub/projects/[PROJECT]/topics/[TOPIC]
		// {
		// 		"context":{
		// 			"eventId":"[ID]",
		// 			"timestamp":"yyyy-mm-ddThh:mm:ss.000Z",
		// 			"eventType":"google.pubsub.topic.publish",
		// 			"resource":{
		// 				"service":"pubsub.googleapis.com",
		// 				"name":"projects/[PROJECT]/topics/[TOPIC]",
		// 				"type":"type.googleapis.com/google.pubsub.v1.PubsubMessage"
		// 			}
		// 		},
		// 		"data":{
		// 			"@type":"type.googleapis.com/google.pubsub.v1.PubsubMessage",
		// 			"attributes":{...},"data":"..."
		// 		}
		// }
		var m struct {
			Ctx  Context `json:"context"`
			Data struct {
				Type string            `json:"@type"`
				Attr map[string]string `json:"attributes"`
				Data []byte            `json:"data"`
			} `json:"data"`
		}
		err := json.NewDecoder(r.Body).Decode(&m)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		msg := &pubsub.Message{
			Attributes: m.Data.Attr,
			Data:       m.Data.Data,
			//PublishTime: m.Ctx.TS,
		}
		if msg.Attributes == nil {
			msg.Attributes = make(map[string]string)
		}
		ctx := r.Context()
		err = fnc(ctx, msg)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	})
}

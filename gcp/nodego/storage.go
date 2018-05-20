// +build storage

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	raw "google.golang.org/api/storage/v1"
)

type StorageFunc func(ctx context.Context, attrs *storage.ObjectAttrs) error

func HandleStorage(fnc StorageFunc) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		// /execute/_ah/push-handlers/pubsub/projects/xxxx/topics/cloud-functions-yyyy
		// {
		// 		"context":{
		// 			"eventId":"[ID]",
		// 			"timestamp":"yyyy-mm-ddThh:mm:ss.000Z",
		// 			"eventType":"google.storage.object.[EVENT_TYPE]",
		// 			"resource":{
		// 				"service":"storage.googleapis.com",
		// 				"name":"projects/_/buckets/[BUCKET]/objects/[PATH]",
		// 				"type":"storage#object"
		// 			}
		// 		},
		// 		"data":{...}
		// }
		var m struct {
			Ctx  Context    `json:"context"`
			Data raw.Object `json:"data"`
		}
		err := json.NewDecoder(r.Body).Decode(&m)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		obj := newObject(&m.Data)
		ctx := r.Context()
		err = fnc(ctx, obj)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	})
}

func newObject(o *raw.Object) *storage.ObjectAttrs {
	if o == nil {
		return nil
	}
	acl := make([]storage.ACLRule, len(o.Acl))
	for i, rule := range o.Acl {
		acl[i] = storage.ACLRule{
			Entity: storage.ACLEntity(rule.Entity),
			Role:   storage.ACLRole(rule.Role),
		}
	}
	owner := ""
	if o.Owner != nil {
		owner = o.Owner.Entity
	}
	md5, _ := base64.StdEncoding.DecodeString(o.Md5Hash)
	crc32c, _ := decodeUint32(o.Crc32c)
	var sha256 string
	if o.CustomerEncryption != nil {
		sha256 = o.CustomerEncryption.KeySha256
	}
	return &storage.ObjectAttrs{
		Bucket:             o.Bucket,
		Name:               o.Name,
		ContentType:        o.ContentType,
		ContentLanguage:    o.ContentLanguage,
		CacheControl:       o.CacheControl,
		ACL:                acl,
		Owner:              owner,
		ContentEncoding:    o.ContentEncoding,
		ContentDisposition: o.ContentDisposition,
		Size:               int64(o.Size),
		MD5:                md5,
		CRC32C:             crc32c,
		MediaLink:          o.MediaLink,
		Metadata:           o.Metadata,
		Generation:         o.Generation,
		Metageneration:     o.Metageneration,
		StorageClass:       o.StorageClass,
		CustomerKeySHA256:  sha256,
		KMSKeyName:         o.KmsKeyName,
		Created:            convertTime(o.TimeCreated),
		Deleted:            convertTime(o.TimeDeleted),
		Updated:            convertTime(o.Updated),
	}
}

// Decode a uint32 encoded in Base64 in big-endian byte order.
func decodeUint32(b64 string) (uint32, error) {
	d, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return 0, err
	}
	if len(d) != 4 {
		return 0, fmt.Errorf("storage: %q does not encode a 32-bit value", d)
	}
	return uint32(d[0])<<24 + uint32(d[1])<<16 + uint32(d[2])<<8 + uint32(d[3]), nil
}

// convertTime converts a time in RFC3339 format to time.Time.
// If any error occurs in parsing, the zero-value time.Time is silently returned.
func convertTime(t string) time.Time {
	var r time.Time
	if t != "" {
		r, _ = time.Parse(time.RFC3339, t)
	}
	return r
}

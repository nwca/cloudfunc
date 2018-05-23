package gcp

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	funcs "google.golang.org/genproto/googleapis/cloud/functions/v1beta2"
)

func ParseTarget(path string) (Target, error) {
	var err error
	path, err = toPackage(path)
	if err != nil {
		return Target{}, err
	}
	base := filepath.Base(path)
	i := strings.LastIndex(base, ".")
	if i < 0 {
		return Target{Package: path}, nil
	}
	fnc := base[i+1:]
	pkg := strings.TrimSuffix(path, "."+fnc)
	return Target{
		Func: fnc, Package: pkg,
	}, nil
}

type Trigger interface {
	writeSource(w io.Writer) error
	buildTags() []string
	gcloudArgs() []string
	setOn(proj string, f *funcs.CloudFunction)
}

type Target struct {
	Package string
	Func    string
}

func (t Target) target() Target {
	return t
}

type HTTPTrigger struct {
	Target
}

func (t HTTPTrigger) buildTags() []string { return nil }

func (t HTTPTrigger) writeSource(w io.Writer) error {
	var err error
	if t.Func == "" {
		_, err = fmt.Fprintf(w, `package main

import _ %q
`, t.Package)
	} else {
		_, err = fmt.Fprintf(w, `package main

import p %q

func init(){
	HandleHTTP(p.%s)
}
`, t.Package, t.Func)
	}
	return err
}

func (t HTTPTrigger) setOn(proj string, f *funcs.CloudFunction) {
	f.Trigger = &funcs.CloudFunction_HttpsTrigger{
		HttpsTrigger: &funcs.HTTPSTrigger{},
	}
}
func (t HTTPTrigger) gcloudArgs() []string {
	return []string{"--trigger-http"}
}

type TopicTrigger struct {
	Target
	Topic string
}

func (t TopicTrigger) buildTags() []string { return []string{"pubsub"} }

func (t TopicTrigger) writeSource(w io.Writer) error {
	_, err := fmt.Fprintf(w, `package main

import p %q

func init(){
	HandlePubSub(p.%s)
}
`, t.Package, t.Func)
	return err
}

func (t TopicTrigger) setOn(proj string, f *funcs.CloudFunction) {
	f.Trigger = &funcs.CloudFunction_EventTrigger{
		EventTrigger: &funcs.EventTrigger{
			EventType: "providers/cloud.pubsub/eventTypes/topic.publish",
			Resource:  "projects/" + proj + "/topics/" + t.Topic,
		},
	}
}

func (t TopicTrigger) gcloudArgs() []string {
	return []string{"--trigger-topic", t.Topic}
}

func toPackage(pkg string) (string, error) {
	if !strings.HasPrefix(pkg, ".") && !strings.HasPrefix(pkg, "/") {
		return pkg, nil
	}
	abs, err := filepath.Abs(pkg)
	if err != nil {
		return pkg, err
	}
	gopath, err := goPath()
	if err != nil {
		return pkg, err
	}
	for _, pref := range strings.Split(gopath, ":") {
		if strings.HasPrefix(abs, pref) {
			pkg = strings.TrimPrefix(abs, filepath.Join(pref, "src"))
			pkg = strings.Trim(pkg, "/")
			break
		}
	}
	return pkg, nil
}

type StorageEvent string

const (
	StorageEventPref      = "google.storage.object."
	StorageFinalize       = StorageEvent(StorageEventPref + "finalize")
	StorageDelete         = StorageEvent(StorageEventPref + "delete")
	StorageArchive        = StorageEvent(StorageEventPref + "archive")
	StorageMetadataUpdate = StorageEvent(StorageEventPref + "metadataUpdate")
)

type StorageTrigger struct {
	Target
	Bucket string
	//Event  StorageEvent
}

func (t StorageTrigger) buildTags() []string { return []string{"storage"} }

func (t StorageTrigger) writeSource(w io.Writer) error {
	_, err := fmt.Fprintf(w, `package main

import p %q

func init(){
	HandleStorage(p.%s)
}
`, t.Package, t.Func)
	return err
}

func (t StorageTrigger) setOn(proj string, f *funcs.CloudFunction) {
	f.Trigger = &funcs.CloudFunction_EventTrigger{
		EventTrigger: &funcs.EventTrigger{
			EventType: "providers/cloud.storage/eventTypes/object.change",
			Resource:  "projects/" + proj + "/buckets/" + t.Bucket,
		},
	}
}

func (t StorageTrigger) gcloudArgs() []string {
	//if t.Event != "" {
	//	return []string{"--trigger-resource", t.Bucket, "--trigger-event", string(t.Event)}
	//}
	return []string{"--trigger-bucket", t.Bucket}
}

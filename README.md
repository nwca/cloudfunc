# cloudfunc
Google Cloud Functions for Go

## Prerequisites:

- Go 1.10+
- Google Cloud SDK (`gcloud`, `gsutil`)

## Build and install the binary:

```
go get -u github.com/nwca/cloudfunc
go install github.com/nwca/cloudfunc/cmd/cloudfunc
```

## Build and deploy a cloud function

Package that registers HTTP handlers in `init()`:

```
cloudfunc deploy http -s my-staging-bucket hello ./example/hello
```

Specific handler function in the package:

```
cloudfunc deploy http -s my-staging-bucket hello ./example/hellofnc.HelloFunc
```

PubSub trigger function:

```
cloudfunc deploy pubsub -s my-staging-bucket -t my-topic hello ./example/pubsub.HandleTopic
```

Storage trigger function:

```
cloudfunc deploy storage -s my-staging-bucket -b my-bucket hello ./example/storage.HandleStorage
```

or with a specific event type:

```
cloudfunc deploy storage -s my-staging-bucket -b my-bucket -e delete hello ./example/storage.HandleStorage
```
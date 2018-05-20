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
cloudfunc deploy http -b my-staging-bucket hello ./example/hello
```

Specific handler function in the package:

```
cloudfunc deploy http -b my-staging-bucket hello ./example/hellofnc.HelloFunc
```

PubSub handler function:

```
cloudfunc deploy pubsub -b my-staging-bucket -t my-topic hello ./example/pubsub.HandleTopic
```

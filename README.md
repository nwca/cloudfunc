# cloudfunc
Google Cloud Functions for Go

## Prerequisites:

- Go 1.10+

## Build and install the binary:

```
go get -u github.com/nwca/cloudfunc
go install github.com/nwca/cloudfunc/cmd/cloudfunc
```

## Login to Google Cloud

```
gcloud auth application-default login
```

## Build and deploy a cloud function

Package that registers HTTP handlers in `init()`:

```
cloudfunc deploy http -p my-project hello ./example/hello
```

Specific handler function in the package:

```
cloudfunc deploy http -p my-project hello ./example/hellofnc.HelloFunc
```

PubSub trigger function:

```
cloudfunc deploy pubsub -p my-project -t my-topic hello ./example/pubsub.HandleTopic
```

Storage trigger function:

```
cloudfunc deploy storage -p my-project -b my-bucket hello ./example/storage.HandleStorage
```

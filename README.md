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

Register a handler using `init()`:

```
cloudfunc deploy -b my-staging-bucket hello ./example/hello
```

Or to register specific handler function:

```
cloudfunc deploy -b my-staging-bucket hello ./example/hellofnc.HelloFunc
```

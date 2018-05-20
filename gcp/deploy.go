package gcp

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func DeployZIP(name string, zip string, bucket string) error {
	return deployFile(name, zip, bucket, HTTPTrigger{}.gcloudArgs())
}

func Deploy(name string, tr Trigger, bucket string) error {
	f, err := ioutil.TempFile("", "cloudfunc-")
	if err != nil {
		return err
	}
	fname := f.Name()
	f.Close()

	if err := os.Rename(fname, fname+".zip"); err != nil {
		os.Remove(fname)
		return err
	}
	fname += ".zip"
	defer os.Remove(fname)

	if err := Build(tr, fname); err != nil {
		return err
	}
	return deployFile(name, fname, bucket, tr.gcloudArgs())
}

func deployFile(name string, file string, bucket string, flags []string) error {
	cfile := "gs://" + bucket + "/" + filepath.Base(file)
	const entrypoint = "helloWorld"

	run := func(name string, args ...string) error {
		cmd := exec.Command(name, args...)
		cmd.Stderr = os.Stderr
		log.Println(append([]string{name}, args...))
		return cmd.Run()
	}

	if err := run("gsutil", "cp", file, cfile); err != nil {
		return err
	}
	defer run("gsutil", "rm", cfile)

	args := []string{
		"beta", "functions", "deploy", name,
		"--entry-point", entrypoint,
		"--source", cfile,
	}
	args = append(args, flags...)

	return run("gcloud", args...)
}

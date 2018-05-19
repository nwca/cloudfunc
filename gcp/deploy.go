package gcp

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Deploy(name string, pkg string, bucket string) error {
	var file string
	if strings.HasSuffix(pkg, ".zip") {
		file = pkg
	} else {
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

		if err := Build(pkg, fname); err != nil {
			return err
		}
		file = fname
	}

	cfile := "gs://" + bucket + "/" + filepath.Base(file)
	const entrypoint = "helloWorld"

	run := func(name string, args ...string) error {
		cmd := exec.Command(name, args...)
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	if err := run("gsutil", "cp", file, cfile); err != nil {
		return err
	}
	defer run("gsutil", "rm", cfile)

	return run("gcloud", "beta", "functions", "deploy", name,
		"--entry-point", entrypoint, "--trigger-http", "--source", cfile)
}

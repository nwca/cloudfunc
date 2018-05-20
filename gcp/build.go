package gcp

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"strings"

	"log"

	"github.com/nwca/cloudfunc/gcp/bindata"
)

func goPath() (string, error) {
	data, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(data)), nil
}

func Build(tr Trigger, out string) error {
	dir, err := ioutil.TempDir("", "cloudfunc-")
	if err != nil {
		return err
	}
	log.Println("buid dir:", dir)
	defer os.RemoveAll(dir)

	if err := unpackBindata("nodego", dir); err != nil {
		return fmt.Errorf("cannot unpack template: %v", err)
	}
	err = writeImpl(dir, tr.writeSource)
	if err != nil {
		return fmt.Errorf("cannot write import: %v", err)
	}
	bin := filepath.Join(dir, "main")
	if err := goBuild(bin, dir, tr.buildTags()); err != nil {
		return fmt.Errorf("cannot build binary: %v", err)
	}
	if err := repackTar2ZipWith("function.tar", out, bin); err != nil {
		return err
	}
	return nil
}

func unpackBindata(from, to string) error {
	files, err := bindata.AssetDir(from)
	if err != nil {
		return err
	}
	for _, name := range files {
		data, err := bindata.Asset(filepath.Join(from, name))
		if err != nil {
			return err
		}
		if err = ioutil.WriteFile(filepath.Join(to, name), data, 0644); err != nil {
			return err
		}
	}
	return nil
}

func writeImpl(dir string, fnc func(w io.Writer) error) error {
	f, err := os.Create(filepath.Join(dir, "impl.go"))
	if err != nil {
		return err
	}
	defer f.Close()
	if err = fnc(f); err != nil {
		return err
	}
	return f.Close()
}

func goBuild(out string, dir string, tags []string) error {
	gopath, err := goPath()
	if err != nil {
		return err
	}
	tags = append(tags, "node")
	cmd := exec.Command("go", "build", "-tags", strings.Join(tags, " "), "-o", out)
	cmd.Env = append(cmd.Env,
		`GOARCH=amd64`,
		`GOOS=linux`,
		`CGO_ENABLED=0`,
		`GOPATH=`+gopath,
	)
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	return cmd.Run()
}

func repackTar2ZipWith(from, to string, add ...string) error {
	arch, err := bindata.Asset(from)
	if err != nil {
		return err
	}
	tr := tar.NewReader(bytes.NewReader(arch))
	zf, err := os.Create(to)
	if err != nil {
		return err
	}
	defer zf.Close()

	zw := zip.NewWriter(zf)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		fi := h.FileInfo()

		zh, err := zip.FileInfoHeader(fi)
		if err != nil {
			return err
		}
		zh.Name = h.Name
		w, err := zw.CreateHeader(zh)
		if err != nil {
			return fmt.Errorf("cannot copy zip file: %v", err)
		}
		_, err = io.Copy(w, tr)
		if err != nil {
			return fmt.Errorf("cannot copy zip file: %v", err)
		}
	}
	for _, fname := range add {
		f, err := os.Open(fname)
		if err != nil {
			return err
		}
		st, err := f.Stat()
		if err != nil {
			f.Close()
			return err
		}
		h, err := zip.FileInfoHeader(st)
		if err != nil {
			f.Close()
			return err
		}
		h.Name = filepath.Base(fname)

		w, err := zw.CreateHeader(h)
		if err != nil {
			f.Close()
			return err
		}
		_, err = io.Copy(w, f)
		if err != nil {
			f.Close()
			return err
		}
	}
	if err = zw.Close(); err != nil {
		return fmt.Errorf("flush failed: %v", err)
	}
	return zf.Close()
}

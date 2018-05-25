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

func BuildTmp(tr Trigger, env map[string]string) (io.ReadCloser, error) {
	buf := bytes.NewBuffer(nil)
	if err := Build(tr, env, buf); err != nil {
		return nil, err
	}
	return ioutil.NopCloser(buf), nil
}

func Build(tr Trigger, env map[string]string, out io.Writer) error {
	dir, err := ioutil.TempDir("", "cloudfunc-")
	if err != nil {
		return err
	}
	log.Println("build dir:", dir)
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
	if err := testBin(bin); err != nil {
		return err
	}
	envjs := filepath.Join(dir, "env.js")
	err = writeEnvJS(envjs, env)
	if err != nil {
		return fmt.Errorf("cannot write env: %v", err)
	}
	if err := repackTar2ZipWith("function.tar", out, bin, envjs); err != nil {
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

func writeSource(file string, fnc func(w io.Writer) error) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	if err = fnc(f); err != nil {
		return err
	}
	return f.Close()
}

func writeImpl(dir string, fnc func(w io.Writer) error) error {
	return writeSource(filepath.Join(dir, "impl.go"), fnc)
}

func writeEnvJS(dst string, env map[string]string) error {
	log.Printf("writing %d environment variables", len(env))
	return writeSource(dst, func(w io.Writer) error {
		var last error
		for k, v := range env {
			_, last = fmt.Fprintf(w, "process.env[%q] = %q;\n", k, v)
		}
		return last
	})
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

func testBin(bin string) error {
	cmd := exec.Command(bin, "-h")
	out, err := cmd.CombinedOutput()
	if _, ok := err.(*exec.ExitError); !ok && err != nil {
		return err
	}
	if bytes.Contains(out, []byte("panic")) {
		return fmt.Errorf("binary panics on start:\n%s", string(out))
	}
	return nil
}

func repackTar2ZipWith(from string, to io.Writer, add ...string) error {
	arch, err := bindata.Asset(from)
	if err != nil {
		return err
	}
	tr := tar.NewReader(bytes.NewReader(arch))

	zw := zip.NewWriter(to)
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
	return nil
}

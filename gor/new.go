package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func CmdInit(path string) {
	_, err := os.Stat(path)
	if err == nil || !os.IsNotExist(err) {
		log.Fatal("Path Exist?!")
	}

	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(INIT_ZIP))
	b, _ := ioutil.ReadAll(decoder)

	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Unpack init content zip")

	for _, zf := range z.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		dst := path + "/" + zf.Name
		os.MkdirAll(filepath.Dir(dst), os.ModePerm)
		f, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		rc, err := zf.Open()
		if err != nil {
			log.Fatal(err)
		}
		_, err = io.Copy(f, rc)
		if err != nil {
			log.Fatal(err)
		}
		f.Sync()
		f.Close()
		rc.Close()
	}
	log.Println("Done")
}

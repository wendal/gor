package gor

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	INIT_ZIP = "https://raw.github.com/wendal/gor/master/gor/gor-content.zip"
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

	log.Println("Download init content zip")

	resp, err := http.Get(INIT_ZIP)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatal("Network error")
	}

	tmp, err := os.Create(path + "/tmp.zip")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Downloading init content zip")
	sz, err := io.Copy(tmp, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	tmp.Sync()
	tmp.Seek(0, os.SEEK_SET)

	z, err := zip.NewReader(tmp, sz)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Unpack init content zip")

	for _, zf := range z.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		dst := path + "/" + zf.FileInfo().Name()
		os.MkdirAll(filepath.Dir(dst), os.ModePerm)
		f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, os.ModePerm)
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
	tmp.Close()
	os.Remove(path + "/tmp.zip")
	log.Println("Done")
}

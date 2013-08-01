package gor

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

const (
	TPL_NEW_POST = `---
title: %s
date: '%s'
description:
categories:

tags:

---

`
	IMG_TAG = `<img src="%s" alt="img: " width="600" class="imgbox">`
	IMG_LOCATION = `http://iwood.qiniudn.com/`
)

// 创建一个新post
// TODO 移到到其他地方?
func CreateNewPost(title string) (path string){
	if !IsGorDir(".") {
		log.Fatal("Not Gor Dir, need config.yml")
	}
	path = "posts/" + strings.Replace(title, " ", "-", -1) + ".md"
	_, err := os.Stat(path)
	if err == nil || !os.IsNotExist(err) {
		log.Fatal("Post File Exist?!", path)
	}
	err = ioutil.WriteFile(path, []byte(fmt.Sprintf(TPL_NEW_POST, title, time.Now().Format("2006-01-02"))), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Create Post at " + path)
	return
}


func CreateNewPostWithImgs(title, imgsrc string) {

	path := CreateNewPost(title)

	start := strings.LastIndex(path, "/") + 1
	end := strings.LastIndex(path, ".")
	if start < 0 || end < 0 {
		log.Fatal("path not complate? ", path)
	}
	post := path[start:end]

	// 如果创建失败直接exit，所以不用检查
	imgs := cpPostImgs(post, imgsrc)
	tags := generateImgLinks(imgs)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for _, tag := range tags {
		if _, err = f.WriteString("\n" + tag + "\n"); err != nil {
			panic(err)
		}
	}
}

func cpPostImgs(post, imgsrc string) (imgtag []string) {
	files, err := ioutil.ReadDir(imgsrc)
	if files == nil || err != nil {
		log.Println("no img file exists.");
		return nil;
	}

	if !strings.HasSuffix(imgsrc, "/") {
		imgsrc += "/"
	}

	imgdst := "rc/media/" + post
	_, err = os.Stat(imgdst)
 	if os.IsNotExist(err) {
		os.MkdirAll(imgdst, 0777)
	}

	imgtag = make([]string, len(files))
	i := 0
	for idx, f := range files {
		err := cp(imgdst + "/" + f.Name(), imgsrc + f.Name())
		if err != nil {
			log.Println(idx, "resouce file cp error: ", f.Name())
			continue
		}
		//log.Println(idx, imgdst + "/" + f.Name(), imgsrc + f.Name())
		imgtag[i] = imgdst + "/" + f.Name()
		i++
	}
	imgtag = imgtag[:i]
	return
}

func cp(dst, src string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	// no need to check errors on read only file, we already got everything
	// we need from the filesystem, so nothing can go wrong now.
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}

func generateImgLinks(files []string) (links []string){
	links = make([]string, len(files))
	for i, f := range files {
		tmp := strings.TrimLeft(f, "rc/")
		links[i] = fmt.Sprintf(IMG_TAG, IMG_LOCATION + tmp)
		println(i, links[i])
	}
		
	return
}


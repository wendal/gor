package gor

import (
	"fmt"
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
---

`
)

// 创建一个新post
// TODO 移到到其他地方?
func CreateNewPost(title string) {
	if !IsGorDir(".") {
		log.Fatal("Not Gor Dir, need config.yml")
	}
	path := "posts/" + strings.Replace(title, " ", "-", -1) + ".md"
	_, err := os.Stat(path)
	if err == nil || !os.IsNotExist(err) {
		log.Fatal("Post File Exist?!", path)
	}
	err = ioutil.WriteFile(path, []byte(fmt.Sprintf(TPL_NEW_POST, title, time.Now().Format("2006-01-02"))), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Create Post at " + path)
}

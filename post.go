package gor

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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
	IMG_TAG       = `<img src="%s" alt="img: " width="600">`
	IMG_URLPERFIX = `{{urls.media}}/`
	IMG_LOCALDIR  = `media/`
)

var reservedNames = map[string]struct{} {
	"CON":{}, 
	"PRN":{}, 
	"AUX":{}, 
	"NUL":{},
	"COM1":{}, 
	"COM2":{}, 
	"COM3":{}, 
	"COM4":{}, 
	"COM5":{}, 
	"COM6":{}, 
	"COM7":{}, 
	"COM8":{}, 
	"COM9":{},
	"LPT1":{}, 
	"LPT2":{}, 
	"LPT3":{}, 
	"LPT4":{}, 
	"LPT5":{}, 
	"LPT6":{}, 
	"LPT7":{}, 
	"LPT8":{}, 
	"LPT9":{},
	".":{},
	"..":{},
	"'":{},
	";":{},
	",":{},
	" ":{},
}

var reservedChars = map[string]struct{} {
	" ":{},
	">":{}, 
	"<":{}, 
	":":{}, 
	"\"":{}, 
	"\\":{}, 
	"/":{}, 
	"|":{}, 
	"?":{}, 
	"*":{}, 
	"'":{}, 
	".":{},
	";":{},
	",":{},
	"x0": {}, // ascii 0
	"x1": {}, 
	"x2": {}, 
	"x3": {}, 
	"x4": {}, 
	"x5": {}, 
	"x6": {}, 
	"x7": {}, 
	"x8": {}, 
	"x9": {}, 
	"xa": {}, 
	"xb": {}, 
	"xc": {}, 
	"xd": {}, 
	"xe": {}, 
	"xf": {}, 
	"x10": {}, 
	"x11": {}, 
	"x12": {}, 
	"x13": {}, 
	"x14": {}, 
	"x15": {}, 
	"x16": {}, 
	"x17": {}, 
	"x18": {}, 
	"x19": {}, 
	"x1a": {}, 
	"x1b": {}, 
	"x1c": {}, 
	"x1d": {}, 
	"x1e": {}, 
	"x1f": {},  // ascii 31
}

// convert post title to path,
// print fatal messages for invalid title and exit or
// return path.
func postPath(title string) (path string) {
	// replace invalid characters
	// https://stackoverflow.com/questions/1976007/what-characters-are-forbidden-in-windows-and-linux-directory-names
	/*
		The following reserved characters:

		< (less than)
		> (greater than)
		: (colon)
		" (double quote)
		/ (forward slash)
		\ (backslash)
		| (vertical bar or pipe)
		? (question mark)
		* (asterisk)
		ASCII 0-31 (ASCII control characters)

		===
		The following filenames are reserved:

		Windows:

		CON, PRN, AUX, NUL 
		COM1, COM2, COM3, COM4, COM5, COM6, COM7, COM8, COM9
		LPT1, LPT2, LPT3, LPT4, LPT5, LPT6, LPT7, LPT8, LPT9
	*/
	if _, ok := reservedNames[title]; ok {
		log.Printf("Reserved title: %s\n", title)
		log.Println("Reserved list:")
		for val := range reservedNames {
			log.Printf("%s, ", val)
		}
		log.Println("")
		log.Fatalf("Invalid title (reserved): %s\n", title)
	}
	for val := range reservedChars {
		title = strings.Replace(title, val, "-", -1)
	}
	path = "posts/" + time.Now().Format("2006-01-02") + "-" + title + ".md"
	return path
}

// 创建一个新post
// TODO 移到到其他地方?
func CreateNewPost(title string) (path string) {
	if !IsGorDir(".") {
		log.Fatal("Not Gor Dir, need config.yml")
	}
	path = postPath(title)
	_, err := os.Stat(path)
	if err == nil || !os.IsNotExist(err) {
		fmt.Printf("Create Post File Error: \n %v\n", err)
		log.Fatal("Post File Exist ?: ", path)
	}
	err = ioutil.WriteFile(path, []byte(fmt.Sprintf(TPL_NEW_POST, title, time.Now().Format("2006-01-02"))), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Create Post at " + path)
	return
}

func CreateNewPostWithImgs(title, imgsrc string) {

	cfg := loadConfig(".")
	for k, v := range cfg {
		log.Println(k, "=", v)
	}
	path := CreateNewPost(title)

	start := strings.LastIndex(path, "/") + 1
	end := strings.LastIndex(path, ".")
	if start < 0 || end < 0 {
		log.Fatal("path not complate? ", path)
	}
	post := path[start:end]

	// 如果创建失败直接exit，所以不用检查
	imgs := cpPostImgs(post, imgsrc, cfg)
	tags := generateImgLinks(imgs, cfg)

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

func AddImgs(title, imgsrc string, date string) {
	
	cfg := loadConfig(".")
	for k, v := range cfg {
		log.Println(k, "=", v)
	}
	
	if !IsGorDir(".") {
		log.Fatal("Not Gor Dir, need config.yml")
	}
	
	if (date == "") {
		date = time.Now().Format("2006-01-02")
	}
	
	path := "posts/" + date  + "-" + strings.Replace(title, " ", "-", -1) + ".md"
	_, err := os.Stat(path)
	if err != nil || os.IsNotExist(err) {
		log.Fatal("Post File Not Exist?!", path)
	}

	start := strings.LastIndex(path, "/") + 1
	end := strings.LastIndex(path, ".")
	if start < 0 || end < 0 {
		log.Fatal("path not complate? ", path)
	}
	post := path[start:end]

	// 如果创建失败直接exit，所以不用检查
	imgs := cpPostImgs(post, imgsrc, cfg)
	tags := generateImgLinks(imgs, cfg)

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

func cpPostImgs(post string, imgsrc string, cfg Mapper) (imgtag []string) {
	files, err := ioutil.ReadDir(imgsrc)
	if files == nil || err != nil {
		log.Println("no img file exists.")
		return nil
	}

	if !strings.HasSuffix(imgsrc, "/") {
		imgsrc += "/"
	}

	imgdst := cfg.GetString("localdir") + post
	_, err = os.Stat(imgdst)
	if os.IsNotExist(err) {
		os.MkdirAll(imgdst, 0777)
	}

	imgtag = make([]string, len(files))
	i := 0
	for idx, f := range files {
		err := cp(imgdst+"/"+f.Name(), imgsrc+f.Name())
		if err != nil {
			log.Println(idx, "resouce file cp error: ", f.Name())
			continue
		}
		//log.Println(idx, imgdst + "/" + f.Name(), imgsrc + f.Name())
		imgtag[i] = post + "/" + f.Name()
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

func generateImgLinks(files []string, cfg Mapper) (links []string) {
	links = make([]string, len(files))
	for i, f := range files {
		//tmp := strings.TrimLeft(f, "rc/")
		links[i] = fmt.Sprintf(cfg.GetString("imgtag"), cfg.GetString("urlperfix")+f)
		println(i, links[i])
	}

	return
}

func loadConfig(root string) (imgs_cfg Mapper) {
	var cfg Mapper
	var err error

	if root == "" {
		root = "."
	}
	root, err = filepath.Abs(root)
	root += "/"
	log.Println("root=", root)

	cfg, err = ReadYml(root + CONFIG_YAML)
	if err != nil {
		log.Println("Fail to read ", root+CONFIG_YAML, err)
		return
	}

	if cfg["imgs"] == nil {
		imgs_cfg = make(Mapper)
	} else {
		imgs_cfg = cfg["imgs"].(map[string]interface{})
	}

	if imgs_cfg["imgtag"] == nil {
		imgs_cfg["imgtag"] = IMG_TAG
	}
	if imgs_cfg["urlperfix"] == nil {
		imgs_cfg["urlperfix"] = IMG_URLPERFIX
	}
	if imgs_cfg["localdir"] == nil {
		imgs_cfg["localdir"] = IMG_LOCALDIR
	}

	return
}

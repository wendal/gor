package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"github.com/wendal/gor"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
)

const (
	VER = "3.6.1"
)

var (
	http_addr = flag.String("http", ":8080", "Http addr for Preview or Server")
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	log.Println("gor ver " + VER)
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 || len(args) > 3 {
		PrintUsage()
		os.Exit(1)
	}
	switch args[0] {
	default:
		PrintUsage()
		os.Exit(1)
	case "config":
		cnf, err := gor.ReadConfig(".")
		if err != nil {
			log.Fatal(err)
		}
		log.Println("RuhohSpec: ", cnf["RuhohSpec"])
		buf, err := json.MarshalIndent(cnf, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		log.Println("global config\n", string(buf))
	case "new":
		fallthrough
	case "init":
		if len(args) == 1 {
			log.Fatalln(os.Args[0], "new", "<dir>")
		}
		CmdInit(args[1])
	case "posts":
		gor.ListPosts()
	case "payload":
		payload, err := gor.BuildPlayload("./")
		if err != nil {
			log.Fatal(err)
		}
		buf, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		log.Println(string(buf))
	case "compile":
		err := gor.Compile()
		if err != nil {
			log.Fatal(err)
		}
	case "post":
		if len(args) == 1 {
			log.Fatal("gor post <title>")
		} else if len(args) == 2 {
			gor.CreateNewPost(args[1])
		} else {
			gor.CreateNewPostWithImgs(args[1], args[2])
		}
	case "http":
		log.Println("Listen at ", *http_addr)
		log.Println(http.ListenAndServe(*http_addr, http.FileServer(http.Dir("compiled"))))
	case "pprof":
		f, _ := os.OpenFile("gor.pprof", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
		for i := 0; i < 100; i++ {
			err := gor.Compile()
			if err != nil {
				log.Fatal(err)
			}
		}
	case ".update.zip.go":
		d, _ := ioutil.ReadFile("gor-content.zip")
		_zip, _ := os.OpenFile("zip.go", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
		header := `package main
const INIT_ZIP="`
		_zip.Write([]byte(header))
		encoder := base64.NewEncoder(base64.StdEncoding, _zip)
		encoder.Write(d)
		encoder.Close()
		_zip.Write([]byte(`"`))
		_zip.Sync()
		_zip.Close()
	}
}

package main

import (
	"encoding/json"
	"flag"
	"github.com/wendal/gor"
	"log"
	"net/http"
	"os"
)

const (
	VER = "1.0.1"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	log.Println("gor ver " + VER)
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 || len(args) > 2 {
		os.Exit(1)
	}
	switch args[0] {
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
		if len(args) == 1 {
			log.Fatalln(os.Args[0], "new", "<dir>")
		}
		gor.CmdInit(args[1])
	case "posts":
		gor.ListPosts()
	case "payload":
		payload, err := gor.BuildPlayload()
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
		}
		gor.CreateNewPost(args[1])
	case "http":
		log.Println("Listen at 0.0.0.0:8080")
		http.ListenAndServe(":8080", http.FileServer(http.Dir("compiled")))
	}
}

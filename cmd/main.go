package main

import (
	"encoding/json"
	"github.com/wendal/gor"
	"log"
	"os"
)

func main() {
	if len(os.Args) == 1 || len(os.Args) > 3 {
		os.Exit(1)
	}
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	switch os.Args[1] {
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
		if len(os.Args) == 2 {
			log.Fatal(os.Args[0], "new", "<dir>")
		}
		gor.CmdInit(os.Args[1])
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
	}
}

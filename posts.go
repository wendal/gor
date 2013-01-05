package gor

import (
	"fmt"
	"log"
)

func ListPosts() {
	var payload Mapper
	payload, err := BuildPlayload()
	if err != nil {
		log.Fatal(err)
	}
	posts := payload["db"].(map[string]interface{})["posts"].(map[string]interface{})["chronological"].([]string)
	fmt.Printf("Posts Count=%d", len(posts))
	for _, id := range posts {
		fmt.Println("-", id)
	}
}

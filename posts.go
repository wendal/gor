package gor

import (
	"fmt"
	"log"
)

// 列出全部post -- 纯属无聊?
func ListPosts() {
	var payload Mapper
	payload, err := BuildPayload("./")
	if err != nil {
		log.Fatal(err)
	}
	posts := payload["db"].(map[string]interface{})["posts"].(map[string]interface{})["chronological"].([]string)
	fmt.Printf("Posts Count=%d\n", len(posts))
	for _, id := range posts {
		fmt.Println("-", id)
	}
}

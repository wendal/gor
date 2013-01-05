package gor

import (
	"encoding/json"
	"log"
	"os"
)

const (
	CONFIG_YAML = "config.yml"
	CONFIG_JSON = "config.json"
)

func IsGorDir() bool {
	_, err := os.Stat(CONFIG_YAML)
	if err != nil {
		if os.IsNotExist(err) {
			_, err = os.Stat(CONFIG_JSON)
			if err != nil {
				return false
			}
		}
	}
	return true
}

func PrintJson(v interface{}) {
	buf, err := json.Marshal(v)
	if err != nil {
		log.Println("ERR Json Marshal : " + err.Error())
	} else {
		log.Println(">>\n" + string(buf))
	}
}

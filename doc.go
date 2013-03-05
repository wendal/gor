// Gor静态模块引擎
package gor

import (
	"encoding/json"
	"log"
	"os"
)

const (
	//配置文件的标准命名
	CONFIG_YAML = "config.yml"
	SITE_YAML   = "site.yml"
)

const (
	KEY_CONFIG = "config"
	KEY_LAYOUT = "layout"
)

// 存在核心配置文件的路径,才可能是Gor的目录
func IsGorDir(path string) bool {
	_, err := os.Stat(path + "/" + CONFIG_YAML)
	return err == nil
}

// 以Json方式打印对象,方便调试
func PrintJson(v interface{}) {
	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Println("ERR Json Marshal : " + err.Error())
	} else {
		log.Println(">>\n" + string(buf))
	}
}

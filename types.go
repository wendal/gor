package gor

import (
	"fmt"
	"log"
)

type Mapper map[string]interface{}

func (m Mapper) Get(key string) interface{} {
	return m[key]
}

func (m Mapper) GetString(key string) string {
	val := m[key]
	if val == nil {
		return ""
	}
	return val.(string)
}

func (m Mapper) GetInt(key string) int64 {
	val := m[key]
	if val == nil {
		return 0
	}
	i, ok := val.(int64)
	if ok {
		return i
	}
	i2, ok := val.(int)
	if ok {
		return int64(i2)
	}
	return 0
}

func (m Mapper) Id() string {
	return m.GetString("id")
}

func (m Mapper) Url() string {
	return m.GetString("url")
}

func (m Mapper) Layout() string {
	return m.GetString("layout")
}

func (m Mapper) Permalink() string {
	return m.GetString("permalink")
}

func (m Mapper) Tags() []string {
	return m.GetStrings("tags")
}

func (m Mapper) Categories() []string {
	return m.GetStrings("categories")
}

func (m Mapper) GetStrings(key string) (strs []string) {
	v := m[key]
	strs = make([]string, 0)
	if v == nil {
		return
	}
	switch v.(type) {
	case string:
		strs = []string{v.(string)}
	case []interface{}:
		for _, v2 := range v.([]interface{}) {
			strs = append(strs, fmt.Sprintf("%v", v2))
		}
	case []string:
		strs = v.([]string)
	default:
		log.Println(">>", v)
	}
	return
}

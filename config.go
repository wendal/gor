package gor

import (
	"encoding/json"
	"github.com/wendal/goyaml"
	"io"
	"io/ioutil"
)

func ReadConfig(root string) (cnf map[string]interface{}, err error) {
	cnf, err = ReadYmlCnf(root)
	if err != nil {
		cnf, err = ReadJsonCnf(root)
	}
	return
}

func ReadYmlCnf(root string) (map[string]interface{}, error) {
	return ReadYml(root + "/config.yml")
}

func ReadJsonCnf(root string) (cnf map[string]interface{}, err error) {
	return ReadJson(root + "/config.json")
}

func ReadYml(path string) (cnf map[string]interface{}, err error) {
	err = nil
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = goyaml.Unmarshal(buf, &cnf)
	return
}

func ReadJson(path string) (cnf map[string]interface{}, err error) {
	err = nil
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = json.Unmarshal(buf, &cnf)
	return
}

func ReadYmlReader(r io.Reader) (cnf map[string]interface{}, err error) {
	err = nil
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}
	err = goyaml.Unmarshal(buf, &cnf)
	return
}

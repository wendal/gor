package gor

import (
	"bytes"
	"encoding/json"
	//"github.com/wendal/goyaml"
	"github.com/wendal/goyaml2"
	"io"
	"io/ioutil"
	"log"
	"os"
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
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	cnf, err = ReadYmlReader(f)
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
	if err != nil || len(buf) < 3 {
		return
	}
	//err = goyaml.Unmarshal(buf, &cnf)

	if string(buf[0:1]) == "{" {
		log.Println("Look lile a Json, try it")
		err = json.Unmarshal(buf, &cnf)
		if err == nil {
			log.Println("It is Json Map")
			return
		}
	}

	_map, _err := goyaml2.Read(bytes.NewBuffer(buf))
	if _err != nil {
		log.Println("Goyaml2 ERR>", string(buf), _err)
		//err = goyaml.Unmarshal(buf, &cnf)
		err = _err
		return
	}
	if _map == nil {
		log.Println("Goyaml2 output nil? Pls report bug\n" + string(buf))
	}
	cnf, ok := _map.(map[string]interface{})
	if !ok {
		log.Println("Not a Map? >> ", string(buf), _map)
		cnf = nil
	}
	return
}

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

func read_condig(file string) *Setting {
	var setting Setting
	config_file, err := os.Open(*conf)
	config, err := ioutil.ReadAll(config_file)
	if err != nil {
		log.Println(err)
		return nil
	}
	config_file.Close()
	if err := json.Unmarshal(config, &setting); err != nil {
		log.Println(err)
		return nil
	}
	return &setting
}

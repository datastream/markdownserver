package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	conf = flag.String("-c", "markdown.json", "markdown server config")
)

type Setting struct {
	Bind  string            `json:"bind"`
	Port  string            `json:"port"`
	Hosts map[string]string `json:"hosts"`
}

var setting Setting
var errorTemplate, _ = template.ParseFiles("error.html")

func main() {
	flag.Parse()
	config_file, err := os.Open(*conf)
	config, err := ioutil.ReadAll(config_file)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	config_file.Close()
	if err := json.Unmarshal(config, &setting); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	log.Println(setting)
	http.HandleFunc("/", routing)
	http.ListenAndServe(setting.Bind+":"+setting.Port, nil)
}

func routing(w http.ResponseWriter, r *http.Request) {
	name := r.Host
	if conf, ok := setting.Hosts[name]; ok {
		log.Println(conf)
	} else {
		w.WriteHeader(http.StatusNotFound)
		errorTemplate.Execute(w, nil)
	}
}

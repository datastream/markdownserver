package main

import (
	"encoding/json"
	"flag"
	"github.com/russross/blackfriday"
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
	host_name := r.Host
	if _, ok := setting.Hosts[host_name]; ok {
		servermarkdown(w, r, host_name)
	} else {
		w.WriteHeader(http.StatusNotFound)
		errorTemplate.Execute(w, nil)
	}
}

func servermarkdown(w http.ResponseWriter, r *http.Request, host string) {
	// need additon path check
	path, ismarkdow := get_file(setting.Hosts[host] + r.URL.Path)
	if f, err := os.Open(path); err != nil {
		w.WriteHeader(http.StatusNotFound)
		// return 404
	} else {
		body, _ := ioutil.ReadAll(f)
		if ismarkdow {
			w.Write(blackfriday.MarkdownCommon(body))
		} else {
			w.Write(body)
		}
	}
}

func get_file(path string) (string, bool) {
	rawpath := path
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			path += ".html"
		}
	} else {
		return path, false
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			path = rawpath + ".md"
		}
	}
	return path, true
}

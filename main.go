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
	if root, ok := setting.Hosts[host_name]; ok {
		root_dir := http.Dir(root)
		servermarkdown(w, r, root_dir)
	} else {
		w.WriteHeader(http.StatusNotFound)
		errorTemplate.Execute(w, nil)
	}
}

func servermarkdown(w http.ResponseWriter, r *http.Request, root_dir http.Dir) {
	// need additon path check
	file_path, ismarkdow := get_file(root_dir, r.URL.Path)
	if f, err := root_dir.Open(file_path); err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		body, _ := ioutil.ReadAll(f)
		if ismarkdow {
			w.Write(blackfriday.MarkdownCommon(body))
		} else {
			w.Write(body)
		}
	}
}

func get_file(root http.Dir, path string) (string, bool) {
	if _, err := root.Open(path); err == nil {
		return path, false
	}
	if _, err := root.Open(path + "/index.html"); err == nil {
		return path + "/index.html", false
	}
	if _, err := root.Open(path + ".md"); err == nil {
		return path + ".md", true
	}
	if _, err := root.Open(path + "/index.md"); err == nil {
		return path + "/index.md", true
	}
	return "", false
}

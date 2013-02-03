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
	"strings"
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
	path := get_real_uri(r.URL.Path)
	file_path, ismarkdow := get_file(setting.Hosts[host], path)
	if f, err := os.Open(file_path); err != nil {
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

func get_file(base_root string, path string) (string, bool) {
	if path[len(path)-1] == '/' {
		path += "index"
	}
	file_base := base_root + path
	if _, err := os.Stat(file_base); err == nil {
		return file_base, false
	}
	if _, err := os.Stat(file_base + ".html"); err == nil {
		return file_base + ".html", false
	}
	if _, err := os.Stat(file_base + ".md"); err == nil {
		return file_base + ".md", true
	}
	return "", false
}

func get_real_uri(path string) string {
	tokens := strings.Split(path, "/")
	var path_tokens []string
	var real_path string
	for i := range tokens {
		if tokens[i] == ".." {
			l := len(path_tokens)
			path_tokens = path_tokens[:l-1]
			if i == 1 {
				break
			}
		} else {
			path_tokens = append(path_tokens, tokens[i])
		}
	}
	for _, token := range path_tokens {
		real_path += "/" + token
	}
	return real_path
}

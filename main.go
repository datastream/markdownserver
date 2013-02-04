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
	http.HandleFunc("/images", static_img)
	http.HandleFunc("/javascripts", static_js)
	http.HandleFunc("/stylesheets", static_css)
	http.ListenAndServe(setting.Bind+":"+setting.Port, nil)
}

func static_img(w http.ResponseWriter, r *http.Request) {
	if root, ok := setting.Hosts[r.Host]; ok {
		http.ServeFile(w, r, root+"/images")
	}
}

func static_js(w http.ResponseWriter, r *http.Request) {
	if root, ok := setting.Hosts[r.Host]; ok {
		http.ServeFile(w, r, root+"/javascripts")
	}
}

func static_css(w http.ResponseWriter, r *http.Request) {
	if root, ok := setting.Hosts[r.Host]; ok {
		http.ServeFile(w, r, root+"/stylesheets")
	}
}

func routing(w http.ResponseWriter, r *http.Request) {
	if root, ok := setting.Hosts[r.Host]; ok {
		root_dir := http.Dir(root)
		servermarkdown(w, r, root_dir)
	} else {
		w.WriteHeader(http.StatusNotFound)
		errorTemplate.Execute(w, nil)
	}
}

func header(root_dir http.Dir) []byte {
	var body []byte
	fd, err := root_dir.Open("templates/header.tpl")
	if err == nil {
		body, _ = ioutil.ReadAll(fd)
	}
	return body
}

func footer(root_dir http.Dir) []byte {
	var body []byte
	fd, err := root_dir.Open("templates/footer.tpl")
	if err == nil {
		body, _ = ioutil.ReadAll(fd)
	}
	return body
}

func servermarkdown(w http.ResponseWriter, r *http.Request, root_dir http.Dir) {
	// need additon path check
	file, ismarkdow := get_file(root_dir, r.URL.Path)
	if file == nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		defer file.Close()
		body, _ := ioutil.ReadAll(file)
		w.Write(header(root_dir))
		if ismarkdow {
			w.Write(blackfriday.MarkdownCommon(body))
		} else {
			w.Write(body)
		}
		w.Write(footer(root_dir))
	}
}

func get_file(root http.Dir, path string) (http.File, bool) {
	if fd, err := root.Open(path); err == nil {
		return fd, false
	}
	if fd, err := root.Open(path + "/index.html"); err == nil {
		return fd, false
	}
	if fd, err := root.Open(path + ".md"); err == nil {
		return fd, true
	}
	if fd, err := root.Open(path + "/index.md"); err == nil {
		return fd, true
	}
	return nil, false
}

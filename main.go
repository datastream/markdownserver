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
	"regexp"
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
	http.HandleFunc("/", routing)
	http.ListenAndServe(setting.Bind+":"+setting.Port, nil)
}

//url routing
func routing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", detect_file_type(r.URL.Path))
	if root, ok := setting.Hosts[r.Host]; ok {
		root_dir := http.Dir(root)
		servermarkdown(w, r, root_dir)
	} else {
		w.WriteHeader(http.StatusNotFound)
		errorTemplate.Execute(w, nil)
	}
}

// return file mime
func detect_file_type(path string) string {
	var f_t string
	reg, _ := regexp.Compile("\\.(css|js|jpg|jpeg|png|xml|rss|gif|svg)$")
	rst := reg.FindString(path)
	switch rst {
	case ".css":
		f_t = "text/css"
	case ".js":
		f_t = "application/javascript"
	case ".jpg":
		f_t = "image/jpeg"
	case ".jpeg":
		f_t = "image/jpeg"
	case ".png":
		f_t = "image/png"
	case ".xml":
		f_t = "text/xml"
	case ".rss":
		f_t = "text/xml"
	case ".gif":
		f_t = "image/gif"
	case ".svg":
		f_t = "image/svg+xml"
	default:
		f_t = "text/html"
	}
	return f_t
}

//templates for header part
func header(root_dir http.Dir) []byte {
	var body []byte
	fd, err := root_dir.Open("templates/header.tpl")
	if err == nil {
		body, _ = ioutil.ReadAll(fd)
	}
	return body
}

//templates for footer part
func footer(root_dir http.Dir) []byte {
	var body []byte
	fd, err := root_dir.Open("templates/footer.tpl")
	if err == nil {
		body, _ = ioutil.ReadAll(fd)
	}
	return body
}

//returm markdown file
func servermarkdown(w http.ResponseWriter, r *http.Request, root_dir http.Dir) {
	file, ismarkdow := get_file(root_dir, r.URL.Path)
	if file == nil {
		w.WriteHeader(http.StatusNotFound)
		error_page, err := template.ParseFiles(string(root_dir) +
			"/error.html")
		if err == nil {
			error_page.Execute(w, nil)
		}
	} else {
		defer file.Close()
		body, _ := ioutil.ReadAll(file)
		if ismarkdow {
			w.Write(header(root_dir))
			w.Write(blackfriday.MarkdownCommon(body))
			w.Write(footer(root_dir))
		} else {
			w.Write(body)
		}
	}
}

//try get file
func get_file(root http.Dir, path string) (http.File, bool) {
	if path[len(path)-1] == '/' {
		path += "index"
	}
	if fd, err := root.Open(path); err == nil {
		return fd, false
	}
	if fd, err := root.Open(path + ".html"); err == nil {
		return fd, false
	}
	if fd, err := root.Open(path + ".md"); err == nil {
		return fd, true
	}
	return nil, false
}

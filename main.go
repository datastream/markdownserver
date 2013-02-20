package main

import (
	"flag"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	conf     = flag.String("f", "markdown.json", "server config")
	mimeconf = flag.String("m", "mime.json", "mime config")
)

type Setting struct {
	Bind  string            `json:"bind"`
	Port  string            `json:"port"`
	Hosts map[string]string `json:"hosts"`
}

var setting *Setting
var errorTemplate, _ = template.ParseFiles("error.html")
var mime MIMETypes

func main() {
	flag.Parse()
	reloadchan := make(chan os.Signal, 1)
	signal.Notify(reloadchan, syscall.SIGHUP)
	go func() {
		<-reloadchan
		log.Println("reload config")
		m := NewMIME(*mimeconf)
		if m != nil {
			mime = m
		}
		s := read_condig(*conf)
		if s != nil {
			setting = s
		}
	}()

	mime = NewMIME(*mimeconf)
	if mime == nil {
		log.Println("mime config not found")
		os.Exit(1)
	}
	setting = read_condig(*conf)
	if setting == nil {
		log.Println("server config not found")
		os.Exit(1)
	}
	http.HandleFunc("/", routing)
	http.ListenAndServe(setting.Bind+":"+setting.Port, nil)
}

//url routing
func routing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", mime.DetectMIME(r.URL.Path))
	if root, ok := setting.Hosts[r.Host]; ok {
		root_dir := http.Dir(root)
		servermarkdown(w, r, root_dir)
	} else {
		w.WriteHeader(http.StatusNotFound)
		errorTemplate.Execute(w, nil)
	}
	log.Print(r.RemoteAddr + " " + r.Method + " " + r.RequestURI +
		" " + r.Host)
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
			headertemplate, err := template.ParseFiles(
				string(root_dir) + "/templates/header.tpl")
			if err == nil {
				headertemplate.Execute(w, nil)
			} else {
				log.Println("header error", err)
			}
			w.Write(blackfriday.MarkdownCommon(body))
			footertemplate, err := template.ParseFiles(
				string(root_dir) + "/templates/footer.tpl")
			if err == nil {
				footertemplate.Execute(w, nil)
			} else {
				log.Println("footer error", err)
			}
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

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

type MIMETypes map[string]string

func NewMIME(file string) MIMETypes {
	fd, err := os.Open(file)
	defer fd.Close()
	if err != nil {
		return nil
	}
	body, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil
	}
	mime := make(map[string]string)
	if err := json.Unmarshal(body, &mime); err != nil {
		return nil
	}
	return mime
}

// return mime
func (this MIMETypes) DetectMIME(path string) string {
	reg, _ := regexp.Compile("\\.([a-zA-Z0-9]+)$")
	rst := reg.FindString(path)
	if len(rst) > 0 {
		k := strings.ToLower(rst)
		if v, ok := this[k[1:]]; ok {
			return v
		}
	}
	return "text/html"
}

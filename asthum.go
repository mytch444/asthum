package main

import (
	"io"
	"os"
	"log"
	"flag"
	"strings"
	"net/url"
	"net/http"
	"os/exec"
	"text/template"
)

const (
	PageTemplateName = ".PAGE.tmpl"
	DirTemplateName = ".DIR.tmpl"
	IndexPrefix = "index"
)

var interpreter, port *string
var root string

type TemplateData struct {
	Name string
	Link string
	Content string
	Links map[string]string
}

func stringInList(name string, list []string) bool {
	for _, h := range list {
		if h == name {
			return true
		}
	}
	return false
}

func parseMarkdown(file *os.File) string {
	cmd := exec.Command(*interpreter)
	cmd.Stdin = file
	output, err := cmd.Output()
	
	if err != nil {
		log.Print("Error parsing markdown: ", err)
		return ""
	} else {
		return string(output)
	}
}

func findTemplate(path string, tmplName string) *template.Template {
	path += "/"
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			file, err := os.Open(path[:i])
			if err != nil {
				log.Print("Error finding template: ", err)
				return nil
			}
			
			names, err := file.Readdirnames(0)
			if err != nil {
				continue
			}
			
			file.Close()
			
			if stringInList(tmplName, names) {
				tmpl, err := template.ParseFiles(path[:i+1] + tmplName)
				if err != nil {
					log.Print("Error opening template " + path[:i+1] + tmplName + ": ", err)
					return nil
				} else {
					return tmpl
				}
			}
		}
	}

	return nil
}

func dirIndex(file *os.File) *os.File {
	names, err := file.Readdirnames(0)
	if err != nil {
		return nil
	}
	file.Seek(0, 0)
		
	for _, name := range names {
		if strings.HasPrefix(name, IndexPrefix) {
			f, err := os.Open(file.Name() + "/" + name)
			if err == nil {
				return f
			}
		}
	}
	
	return nil
}

func processDir(w http.ResponseWriter, data *TemplateData, file *os.File) {
	file, err := os.Open(file.Name())
	names, err := file.Readdirnames(-1)
	if err != nil {
		return
	}
	
	data.Links = make(map[string]string)
	
	for _, name := range names {
		if ! strings.HasPrefix(name, ".") {
			name = strings.TrimSuffix(name, ".md")
			data.Links[name] = data.Link + name
		}
	}
	
	tmpl := findTemplate(file.Name(), DirTemplateName)
	if tmpl == nil {
		for name, link := range data.Links {
			io.WriteString(w, "<a href=\"" + link + "\">" + name + "</a><br/>")
		}
	} else {
		tmpl.Execute(w, data)
	}
}

func executePage(w http.ResponseWriter, link *url.URL, data *TemplateData, file *os.File) {
	values := link.Query()
	
	args := make([]string, len(values) + 1)
	args[0] = file.Name()
	
	i := 1
	for name, value := range values {
		args[i] = name + "=" + value[0]
		i++
	}
	
	l := strings.LastIndex(file.Name(), "/")
	dir := file.Name()[:l]
	base := file.Name()[l:]
	
	cmd := exec.Command("." + base)
	cmd.Stdout = w
	cmd.Args = args
	cmd.Dir = dir
	
	err := cmd.Run()
	if err != nil {
		log.Print("Error executing: ", file.Name(), err)
	}
}

func processPage(w http.ResponseWriter, link *url.URL, data *TemplateData, file *os.File) {
	fi, err := file.Stat()
	if err != nil {
		log.Print("Error stating: ", err)
		return
	}
	
	if strings.Contains(fi.Mode().String(), "x") {
		executePage(w, link, data, file)
	} else {
		data.Content = parseMarkdown(file)
		tmpl := findTemplate(file.Name(), PageTemplateName)
		if tmpl == nil {
			io.WriteString(w, data.Content)
		} else {
			tmpl.Execute(w, data)
		}
	}
}

func processFile(w http.ResponseWriter, link *url.URL, file *os.File) {
	fi, err := file.Stat()
	if err != nil {
		log.Print("Error stating: ", err)
		return
	}

	data := new(TemplateData)
	data.Link = link.Path
	data.Name = strings.TrimSuffix(fi.Name(), ".md")
	
	if fi.IsDir() {
		index := dirIndex(file)
		
		if ! strings.HasSuffix(data.Link, "/") {
			data.Link += "/"
		}
		
		if index == nil {
			processDir(w, data, file)
			return
		} else {
			defer index.Close()
			file = index
		}
	}
	
	processPage(w, link, data, file)
}

func handler(w http.ResponseWriter, req *http.Request) {
	var file *os.File
	var err error
	
	log.Print(req.URL.String())
	
	path := root + req.URL.Path

	file, err = os.Open(path)
	if err != nil {
		path += ".md"
		file, err = os.Open(path)
		
		if err != nil {
			io.WriteString(w, "404: " + req.URL.Path)
			return
		}
	}

	lsl := strings.LastIndex(req.URL.Path, "/")
	if lsl > 0 && strings.Contains(req.URL.Path[lsl+1:], ".") {
		io.Copy(w, file)
	} else {
		processFile(w, req.URL, file)
	
	}
	file.Close()
}

func main() {
	interpreter = flag.String("m", "markdown", "Name/Path of/to executable used to parse markdown")
	port = flag.String("p", "80", "Port to listen on")
	
	flag.Parse()
	
	if flag.NArg() > 0 {
		root = flag.Args()[0]
	} else {
		root = "."
	}
	
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":" + *port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
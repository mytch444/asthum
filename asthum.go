package main

import (
	"io"
	"os"
	"flag"
	"sort"
	"log"
	"strings"
	"net/http"
	"os/exec"
	"text/template"
)

const (
	PageTemplateName = "PAGE.tmpl"
	DirTemplateName = "DIR.tmpl"
)

var hiddenNames []string = []string{
	"PAGE.tmpl",
	"DIR.tmpl",
}

var interpreter *string
var port *string
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
					log.Print("Error finding template: ", err)
					return nil
				} else {
					return tmpl
				}
			}
		}
	}

	return nil
}

func processIndex(w http.ResponseWriter, file *os.File) {
	fi, err := file.Stat()
	if err != nil {
		log.Print("Error stating index: ", err)
		return
	}
	
	if strings.Contains(fi.Mode().String(), "x") {
		/* File executable */
		cmd := exec.Command(file.Name())
		output, err := cmd.Output()
		if err != nil {
			log.Print("Error executing: ", err)
		} else {
			io.WriteString(w, string(output))
		}
	} else {
		/* File not executable */
		io.Copy(w, file)
	}
}

func processDir(w http.ResponseWriter, data *TemplateData, file *os.File, fi os.FileInfo) {
	names, err := file.Readdirnames(0)
	if err != nil {
		io.WriteString(w, "Error reading dirnames for " + file.Name())
		return
	}
	
	if !strings.HasSuffix(data.Link, "/") {
		data.Link += "/"
	}
	
	data.Links = make(map[string]string)
	sort.Strings(names)
	
	for _, name := range names {
		if strings.HasPrefix(name, "index") {
			in, err := os.Open(file.Name() + "/" + name)
			if err == nil {
				processIndex(w, in)
				return
			} else {
				log.Print("Error opening index file: ", err)
				return
			}
		} else if !stringInList(name, hiddenNames) {
			n := strings.TrimSuffix(name, ".md")
			data.Links[n] = data.Link + n
		}
	}
	
	tmpl := findTemplate(file.Name(), DirTemplateName)
	if tmpl == nil {
		io.WriteString(w, "<html><head><title>" + fi.Name() + "</title></head><body>")
		for name, link := range data.Links {
			io.WriteString(w, "<a href=\"" + link + "\">" + name + "</a><br/>")
		}
		io.WriteString(w, "</body></html>")
	} else {
		tmpl.Execute(w, data)
	}
}

func processPage(w http.ResponseWriter, data *TemplateData, file *os.File, fi os.FileInfo) {
	data.Content = parseMarkdown(file)
	
	tmpl := findTemplate(file.Name(), PageTemplateName)
	if tmpl == nil {
		io.WriteString(w, "<html><head><title>" + data.Name + "</title></head><body>")
		io.WriteString(w, data.Content)
		io.WriteString(w, "</body></html>")
	} else {
		tmpl.Execute(w, data)
	}
}

func handler(w http.ResponseWriter, req *http.Request) {
	var file *os.File
	var fi os.FileInfo
	var err error
	var path string
	
	path = root + req.URL.Path

	fi, err = os.Stat(path)
	if err != nil {
		if !strings.HasSuffix(path, ".md") {
			path += ".md"
			fi, err = os.Stat(path)
		}
		
		if err != nil {
			io.WriteString(w, "Error stating " + path)
			return
		}
	}

	file, err = os.Open(path)
	if err != nil {
		io.WriteString(w, "Error opening " + path)
		return
	}
	defer file.Close()
	
	/* If they requested the markdown version then give them it */
	if strings.HasSuffix(req.URL.Path, ".md") {
		io.Copy(w, file)
		return
	}

	data := new(TemplateData)
	data.Link = req.URL.Path
	data.Name = strings.TrimSuffix(fi.Name(), ".md")

	if fi.IsDir() {
		processDir(w, data, file, fi)
	} else {
		processPage(w, data, file, fi)
	}
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
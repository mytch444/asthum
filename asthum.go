package main

import (
	"os"
	"io"
	"log"
	"flag"
	"strings"
	"net/url"
	"net/http"
	"os/exec"
	"io/ioutil"
	"text/template"
	"html"
)

type TemplateData struct {
	Name string
	Link string
	Content string
}

const (
	PageTemplateName = ".page.tmpl"
	InterpreterName = ".interpreters"
)

var siteRoot *string = flag.String("r", ".", "Path to files")
var siteName *string = flag.String("n", "debug", "Name of the site # Will be suffixed to all pages")
var serverPort *string = flag.String("p", "80", "Port to listen on")

/*
 * Split s on last occurence of pattern, so returns (most, suffix).
 * If no matches of pattern were found then returns (s, "").
 */
func splitSuffix(s string, pattern string) (string, string) {
	l := strings.LastIndex(s, pattern)
	if l > 0 {
		return s[:l], s[l+1:]
	} else {
		return s, ""
	}
}

func findFile(path string, name string) string {
	for {
		path, _ = splitSuffix(path, "/")
		if path == "" {
			return os.DevNull
		}
		p := path + "/" + name
		_, err := os.Stat(p)
		if err == nil {
			return p
		}
	}
}

func dirIndex(file *os.File) string {
	names, err := file.Readdirnames(0)
	if err != nil {
		return ""
	}
	file.Seek(0, 0)
	
	dir := file.Name()
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
		
	for _, name := range names {
		if strings.HasPrefix(name, "index") {
			return name
		}
	}
	
	return ""
}

func readLine(file *os.File, bytes []byte) (string, error) {
	n, err := file.Read(bytes)
	if err != nil {
		return "", err
	}
	
	s := string(bytes[:n])
	l := strings.IndexByte(s, '\n') + 1
	
	if l > 0 {
		file.Seek(int64(l - n), 1)
		return s[:l-1], nil
	} else {
		return "", nil
	}
}

func findInterpreter(path string) (bool, []string) {
	intPath := findFile(path, InterpreterName)
	if intPath == "" {
		return false, []string{}
	}
	
	file, err := os.Open(intPath)
	if err != nil {
		log.Print(err)
		return false, []string{}
	}
	
	bytes := make([]byte, 256)
	
	for {
		line, err := readLine(file, bytes)
		if err != nil {
			return false, []string{}
		} else if len(line) == 0 || line[0] == '#' {
			continue
		}
		
		suffix := strings.SplitN(line, " ", 2)
		if len(suffix) > 0 && strings.HasSuffix(path, suffix[0]) {
			parts := strings.Split(line, " ")
			if len(parts) < 2 {
				log.Print("Error in interpreter file. ", path)
				continue
			} else {
				return strings.HasPrefix(parts[1], "y"), parts[2:]
			}
		}
	}
}

func runInterpreter(interpreter []string, values map[string][]string, file *os.File) ([]byte, error) {
	l := strings.LastIndex(file.Name(), "/")
	dir := file.Name()[:l]
	base := file.Name()[l+1:]
	
	l = len(interpreter) + len(values) + 1
	args := make([]string, l)
	copy(args, interpreter)
	args[len(interpreter)] = base
	
	i := len(interpreter) + 1
	for name, value := range values {
		args[i] = name + "=" + value[0]
		i++
	}
	
	cmd := exec.Command(interpreter[0])
	cmd.Args = args
	cmd.Dir = dir
	return cmd.Output()
}

func processFile(w http.ResponseWriter, link *url.URL, data *TemplateData, file *os.File) {
	var err error
	var bytes []byte
	
	useTemplate, interpreter := findInterpreter(file.Name())
	
	if len(interpreter) == 0 {
		bytes, err = ioutil.ReadAll(file)
	} else {
		bytes, err = runInterpreter(interpreter, link.Query(), file)
	}
	
	if err != nil {
		log.Print(err)
		io.WriteString(w, "ERROR")
		return
	} else {
		data.Content = string(bytes)
	}
	
	if useTemplate {
		tmplPath := findFile(file.Name(), PageTemplateName)
		tmpl, err := template.ParseFiles(tmplPath)
		if err == nil {
			tmpl.Execute(w, data)
			return
		}
	}
	
	/* No template/error opening template */
	io.WriteString(w, data.Content)
}

func handler(w http.ResponseWriter, req *http.Request) {
	var file *os.File
	var err error
	
	path := "." + html.EscapeString(req.URL.Path)

	file, err = os.Open(path)
	if err != nil {
		log.Print("404 ", path + " -- ", err)
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "404: " + html.EscapeString(req.URL.Path))
		return
	}
	defer file.Close()
	
	log.Print(req.RemoteAddr, " request: ", path)

	fi, err := file.Stat()
	if err != nil {
		log.Print(err)
		return
	}

	data := new(TemplateData)
	data.Link = req.URL.Path
	
	if path == "./" {
		data.Name = *siteName
	} else {
		name, _ := splitSuffix(fi.Name(), ".")
		data.Name = name + " # " + *siteName
	}

	if fi.IsDir() {
		index := dirIndex(file)

		if index == "" {
			w.WriteHeader(http.StatusNotFound)
			io.WriteString(w, "404: " + html.EscapeString(req.URL.Path))
		} else {
			/* Redirect to index */
			url := req.URL.Path
			if !strings.HasSuffix(url, "/") {
				url += "/"
			}
			
			url += index
			
			if req.URL.RawQuery != "" {
				url += "?" + req.URL.RawQuery
			}

			http.Redirect(w, req, url, http.StatusSeeOther)
		}
	} else {
		processFile(w, req.URL, data, file)
	}
}

func main() {
	flag.Parse()
	
	os.Chdir(*siteRoot)
	
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":" + *serverPort, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

package main

import (
	"os"
	"io"
	"log"
	"fmt"
	"flag"
	"strings"
	"net/http"
	"os/exec"
	"text/template"
	"html"
	"sync"
	"regexp"
)

type TemplateData struct {
	Name string
	Link string
	Content string
}

const (
	TemplateName = ".tmpl"
	RulesName = ".rules"
)

var siteRoot *string = flag.String("r", ".", 
	"Path to serve.")

var siteName *string = flag.String("n", "Asthum Site",
	"Site name.")

var nameFormat *string = flag.String("f", "%s - %s", 
	"String used by fmt to get name to give to template." +
	"The first substitution is the page name, second the site name.")

var serverPortNormal *string = flag.String("p", "80", 
	"Port to listen on for normal connections. Set to 0 to disable.")

var serverPortTLS *string = flag.String("t", "0", 
	"Port to listen on for TLS connections. Set to 0 to disable.")

var certFilePath *string = flag.String("c", "/dev/null", 
	"TLS certificate.")

var keyFilePath *string = flag.String("k", "/dev/null", 
	"TLS key file.")

var maxBytes *int = flag.Int("m", 1024 * 1024,
	"Max file size that will be given to templates. Also the chunk size " + 
	"that is read in before writing to the stream")

func splitSuffix(s string, pattern string) (string, string) {
	l := strings.LastIndex(s, pattern)
	if l > 0 {
		return s[:l], s[l+1:]
	} else {
		return "", s
	}
}

func findFile(path string, name string) string {
	for {
		path, _ = splitSuffix(path, "/")
	
		p := "./" + path + "/" + name

		_, err := os.Stat(p)
		if err == nil {
			return p
		}
	
		if path == "" {
			return ""
		}
	}
}

func readLine(file *os.File, bytes []byte) (int, error) {
	var i int
	b := make([]byte, 1)
	escaped := false
	
	for i = 0; i < len(bytes); i++ {
		_, err := file.Read(b)
		if err != nil {
			return i, err 
		}

		if rune(b[0]) == '\\' {
			escaped = true
		} else if rune(b[0]) == '\n' {
			if escaped {
				escaped = false
				i -= 2
				continue
			} else {
				break
			}
		} else {
			escaped = false
		}

		bytes[i] = b[0]
	}

	return i, nil
}

func parseRule(strings []string) (bool, bool, []string) {
	i := 0
	templated := false

	if strings[i] == "hidden" {
		return true, false, []string{}
	} else if strings[i] == "templated" {
		templated = true
		i++
	}

	return false, templated, strings[i:]
}

func findApplicableRule(file *os.File, name string) ([]string, error) {
	bytes := make([]byte, 256)
	
	for {
		n, err := readLine(file, bytes)
		if err != nil {
			return []string{}, err
		} else if n < 1 {
			continue
		}

		line := strings.Split(string(bytes[:n]), " ")
		
		if len(line) == 0 || line[0][0] == '#' {
			continue
		}

		matched, err := regexp.MatchString(line[0], name)

		if matched {
			return line[1:], nil
		}
	}
}

func readRules(path string) (bool, bool, []string) {
	var file *os.File = nil
	hidden, templated := false, false
	interpreter := []string{}
	
	parts := strings.Split(path, "/")
	
	spath := "./"

	for _, part := range parts {
		_, err := os.Stat(spath + RulesName)
		if err == nil {
			if file != nil {
				file.Close()
			}

			file, err = os.Open(spath + RulesName)
			if err != nil {
				panic(err)
			}
		}

		if file != nil {
			file.Seek(0, 0)

			rule, _ := findApplicableRule(file, part)
			if len(rule) > 0 {
				hidden, templated, interpreter = parseRule(rule)
				/* If any parent directories are hidden then it will be hidden. */
				if hidden {
					break
				}
			}
		}

		spath += path + "/"
	}
	
	if file != nil {
		file.Close()
	}

	return hidden, templated, interpreter
}

func runInterpreter(interpreter []string, 
		values map[string][]string, file *os.File) ([]byte, error) {
	dir, base := splitSuffix(file.Name(), "/")

	cmd := exec.Command(interpreter[0])
	cmd.Args = append(interpreter, base)
	cmd.Dir = "./" + dir

	l := len(cmd.Env) + len(values) + 1
	env := make([]string, l)
	copy(env, cmd.Env)
	
	i := len(cmd.Env) + 1
	for name, value := range values {
		env[i] = name + "=" + value[0]
		i++
	}
	
	cmd.Env = env
	return cmd.Output()
}

func processFile(w http.ResponseWriter, req *http.Request,
		data *TemplateData, file *os.File, fi os.FileInfo) {
	var err error
	var bytes []byte
	var n int
	
	hidden, templated, interpreter := readRules(file.Name())

	if hidden {
		log.Print("Hidden file requested:", req.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "404")
		return
	}

	if len(interpreter) > 0 {
		bytes, err = runInterpreter(interpreter, 
				req.URL.Query(), file)
		n = len(bytes)
	} else {
		bytes = make([]byte, *maxBytes)
		n, err = file.Read(bytes)
	}

	if err != nil {
		log.Print("Error: ", err)
		io.WriteString(w, 
			"An error occured. " +
			"Please contact the administrator.")
		return
	}

	if templated {
		processTemplatedData(w, req, data, bytes[:n])
	} else {
		processRawData(w, req, bytes, n, fi.Size(), file)
	}
}

func processTemplatedData(w http.ResponseWriter, req *http.Request, 
		data *TemplateData, bytes []byte) {

	tmplPath := findFile(req.URL.Path[1:], TemplateName)

	if tmplPath == "" {
		log.Print("Error: No template found!!")
		io.WriteString(w, 
			"An error occured. " +
			"Please contact the administrator.")
		return
	}

	tmpl, err := template.ParseFiles(tmplPath)
	if err == nil {
		data.Content = string(bytes)
		tmpl.Execute(w, data)
	}
}

func processRawData(w http.ResponseWriter, req *http.Request, 
		bytes []byte, n int, size int64, file *os.File) {
	var err error
	req.ContentLength = size

	for {
		_, err = w.Write(bytes[:n])
		if err != nil {
			break
		}
		n, err = file.Read(bytes)
		if err != nil {
			break
		}
	}
}

func findDirIndex(req string) string {
	file, err := os.Open("." + req)
	if err != nil {
		log.Print("Error finding index: ", err)
		return ""
	}
	defer file.Close()

	names, err := file.Readdirnames(0)
	if err != nil {
		log.Print("Error: ", err)
		return ""
	}
	
	if !strings.HasSuffix(req, "/") {
		req += "/"
	}
		
	for _, name := range names {
		if strings.HasPrefix(name, "index") {
			return req + name
		}
	}
	
	return ""
}

func handleDir(w http.ResponseWriter, req *http.Request) {
	index := findDirIndex(req.URL.Path)

	if index == "" {
		log.Print("Error:", req.URL.Path, "has no index")
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "404")
		return
	}
		
	url := index + req.URL.RawQuery
	log.Print("Redirect to: ", url)
	http.Redirect(w, req, url, http.StatusMovedPermanently)
}

func handler(w http.ResponseWriter, req *http.Request) {
	var file *os.File
	var err error
	var name string
	
	log.Print(req.RemoteAddr, " requested: ", req.URL.String())
	
	path := html.EscapeString(req.URL.Path[1:])

	if len(path) == 0 {
		handleDir(w, req)
		return
	}

	file, err = os.Open(path)
	if err != nil {
		log.Print("Error: ", err)
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "404")
		return
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		log.Print("Error: ", err)
		return
	}

	if fi.IsDir() {
		handleDir(w, req)
		return
	}

	data := new(TemplateData)
	data.Link = req.URL.String()
	
	if strings.HasPrefix(fi.Name(), "index") {
		path, _ = splitSuffix(path, "/")
		_, name = splitSuffix(path, "/")
		path += "/"
	} else {
		name, _ = splitSuffix(fi.Name(), ".")
	}
	
	if path == "/" {
		data.Name = *siteName
	} else {
		data.Name = fmt.Sprintf(*nameFormat, name, *siteName)
	}

	processFile(w, req, data, file, fi)
}

func main() {
	var wg sync.WaitGroup
	
	flag.Parse()

	err := os.Chdir(*siteRoot)
	if err != nil {
		panic(err)
	}
	
	http.HandleFunc("/", handler)
	
	wg.Add(2)

	go func() {
		defer wg.Done()

		if *serverPortNormal == "0" {
			return
		}
		
		err := http.ListenAndServe(":" + *serverPortNormal, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	go func() {
		defer wg.Done()

		if *serverPortTLS == "0" {
			return
		}

		err := http.ListenAndServeTLS(":" + *serverPortTLS, 
			*certFilePath, *keyFilePath, nil)
		if err != nil {
			log.Fatal("ListenAndServeTLS: ", err)
		}
	}()

	wg.Wait()
}

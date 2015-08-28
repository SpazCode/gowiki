package main

import (
	"regexp"
	"html/template"
	"io/ioutil"
	"net/http"
	"errors"
	"fmt"
	"strings"
)

var templates = template.Must(template.ParseFiles("templates/edit.html", "templates/view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var fileNames = []string{}
var fileNameStrings = ""

func setup() {
	var files,_ = ioutil.ReadDir("./files")
	for _,file := range files {
		fmt.Println("File Found: " + file.Name())
		name := strings.Replace(file.Name(), ".txt", "", 1)
		fileNameStrings = fileNameStrings + name + " "
		fileNames = append(fileNames, name)
	}
	var tplFuncMap = make(template.FuncMap)
	tplFuncMap["Split"] = strings.Split
	templates = template.Must(template.New("").Funcs(tplFuncMap).ParseFiles("templates/edit.html", "templates/view.html"))
}

type Page struct {
	Title string
	Body []byte
}

func (p *Page) save() error {
	filename := "files/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := "files/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // the title is the second subexpression.
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	fmt.Println(title)
	if title != "FrontPage" {
		p, err :=loadPage(title)
		if err != nil {
			http.Redirect(w, r, "/edit/"+title, http.StatusFound)
			return
		}
		renderTemplate(w, "view",p)
	} else {
		p := &Page{Title: "Root", Body: []byte(fileNameStrings)}	
		renderTemplate(w, "root",p)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	if title != "FrontPage" {
		p, err :=loadPage(title)
		if err != nil {
			p = &Page{Title: title}
		}
		renderTemplate(w, "edit",p)
	} else {
		http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, "templates/" + tmpl + ".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	setup()
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.ListenAndServe(":8080", nil)
}


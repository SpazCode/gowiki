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

var index = &Index{Pages: []string{}}
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var templates = template.Must(template.ParseGlob("templates/*"))

func setup() {
	index = &Index{Pages: []string{}}
	var files,_ = ioutil.ReadDir("./files")
	for _,file := range files {
		fmt.Println("- File Found: " + file.Name())
		name := strings.Replace(file.Name(), ".txt", "", 1)
		index.Pages = append(index.Pages, name)
	}
}

type Page struct {
	Title string
	Body []byte
}

type Index struct {
	Pages []string
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
	fmt.Println("Viewing " + title)
	p, err :=loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view",p)
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Creating New")
	err := templates.ExecuteTemplate(w, "newPage", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Loading Index")
	setup()
	err := templates.ExecuteTemplate(w, "index", index)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	fmt.Println("Editing: " + title)
	p, err :=loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit",p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	pageTitle := title
	if title == "new" {
		pageTitle = r.FormValue("title")
	}
	body := r.FormValue("body")
	fmt.Println("Saving: " + pageTitle)
	p := &Page{Title: pageTitle, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+pageTitle, http.StatusFound)
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
	err := templates.ExecuteTemplate(w, tmpl + "Page", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/new", newHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.ListenAndServe(":8080", nil)
}

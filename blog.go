package main

import (
        "github.com/knieriem/markdown"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"bufio"
	"bytes"
	"regexp"
	"strconv"
	"text/template"
	"errors"
)

type Page struct {
	Title string
	Outfilename string
	Body  []byte
	BodyParsed []byte
	Year int
	Month int
	Day int
	Index int
}

var Pages []Page

var BlogTemplate *template.Template
var FrontTemplate *template.Template

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(filename string) (*Page, error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	p := Page{Title: filename, Body: body}

	r, err := regexp.Compile(`/(\d+)_(.*)\.txt`)
	if err != nil {
		fmt.Println("Problem with the loadPage regexp")
		os.Exit(1)
	}
	res := r.FindStringSubmatch(filename)

	if res != nil {
		p.Year = 0 //strconv.Atoi(res[1])
		p.Month = 0 //strconv.Atoi(res[2])
		p.Day = 0 //strconv.Atoi(res[3])
		p.Index, _ = strconv.Atoi(res[1])
		p.Title = res[2]
	} else {
		fmt.Println("Entry does not match format:"+filename)
		return nil, errors.New("Couldn't load Page")
	}
	return &p, nil
}

func (p *Page) parseAndSave() error {
	p.Outfilename = p.Title+".html"
	filename := "output/"+p.Outfilename
	fo, err := os.Create(filename)
	if err != nil {
		return err
	}

	var b []byte
	buff := bytes.NewBuffer(b)
	writer := bufio.NewWriter(buff)
	reader := bufio.NewReader(bytes.NewReader(p.Body))
	parser := markdown.NewParser(nil)
	fmt.Println("Writing to : "+filename)
	parser.Markdown( reader, markdown.ToHTML(writer))
	writer.Flush()

	p.BodyParsed = buff.Bytes()
	fw := bufio.NewWriter(fo)
	err = BlogTemplate.Execute(fw, p)
	if err != nil {
		panic(err)
	}
	fw.Flush()
	return nil
}

func parseAllFiles() error {
	dirname := "./input"
	d, err := os.Open(dirname)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	files, err := d.Readdir(-1)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, fi := range files {
		fmt.Println("Processing: "+fi.Name())
		page, err := loadPage(dirname+"/"+fi.Name())
		if err != nil {
			fmt.Println(err)
		} else {
			page.parseAndSave()
			Pages = append(Pages, *page)
		}
	}
	return nil
}

func makeIndex() {
	fo, err := os.Create("output/index.html")

	if err != nil {
		fmt.Println("Cannot write index.html")
		os.Exit(1)
	}

	var b []byte
	buff := bytes.NewBuffer(b)
	for _,p := range Pages {
		buff.WriteString("<a href=\""+p.Outfilename+"\" >")
		buff.WriteString(strconv.Itoa(p.Index) )
		buff.WriteString(" "+p.Title+"</a><br/>")
	}

//	p := buff.Bytes()
	fw := bufio.NewWriter(fo)
	err = FrontTemplate.Execute(fw, Pages)
	if err != nil {
		panic(err)
	}
	fw.Flush()
	fo.Close()
}

func initTemplate(filename string) *template.Template {
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	 }

	template, err := template.New(filename).Parse(string(f))
	if err != nil { panic(err) }
	return template
}

func initTemplates() {
	BlogTemplate = initTemplate("templates/blogentry.html")
	FrontTemplate = initTemplate("templates/front.html")
}

func copyStaticFiles() {
	exec.Command("mkdir", "output").Run()
	exec.Command("cp", "templates/styles.css", "output").Run()
	exec.Command("cp", "templates/404.html", "output").Run()
	exec.Command("cp", "templates/not_found.html", "output").Run()
}

func main() {
	copyStaticFiles()

	initTemplates()

	parseAllFiles()

	makeIndex()

}

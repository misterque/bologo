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

	r, err := regexp.Compile(`(\d+)_(\d+)_(\d+)_(.*)\.txt`)
	if err != nil {
		fmt.Println("Problem with the loadPage regexp")
		os.Exit(1)
	}
	res := r.FindStringSubmatch(filename)

	if res != nil {
		p.Year, _ = strconv.Atoi(res[1])
		p.Month, _ = strconv.Atoi(res[2])
		p.Day, _ = strconv.Atoi(res[3])
		p.Title = res[4]
	} else {
		fmt.Println("Entry does not match format:"+filename)
		os.Exit(1)
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
			os.Exit(1)
		}
		page.parseAndSave()
		Pages = append(Pages, *page)
	}
	return nil
}

func MakeIndex() {
	fo, err := os.Create("output/index.html")

	if err != nil {
		fmt.Println("Cannot write index.html")
		os.Exit(1)
	}
	for _,p := range Pages {
		fo.WriteString("<a href=\""+p.Outfilename+"\" >")
		fo.WriteString(strconv.Itoa(p.Day)+"."+strconv.Itoa(p.Month)+"."+strconv.Itoa(p.Year) )
		fo.WriteString(" "+p.Title+"</a><br/>")
	}
	fo.Close()
}

func initTemplates() {
	f, err := ioutil.ReadFile("templates/blogentry.html") 
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	 }

	BlogTemplate, err = template.New("blogentry").Parse(string(f))
	if err != nil { panic(err) }
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

	MakeIndex()

}

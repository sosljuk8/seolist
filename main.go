package main

import (
	"errors"
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"regexp"
)

var count int

type Job struct {
	AllowedDomains []string
	StartingURL    string
	OnLink         func(e *colly.HTMLElement) error
	OnPage         func(e *colly.Response) (Page, error)
}

type Page struct {
	Url  string
	HTML string
}

func (p Page) GetURL() string {
	return p.Url
}

func (p Page) GetHTML() string {
	return p.HTML
}

type IPage interface {
	GetURL() string
	GetHTML() string
}

func NewDefaultJob() Job {
	return Job{
		AllowedDomains: []string{},
		StartingURL:    "",
		OnLink: func(e *colly.HTMLElement) error {
			return nil
		},
		OnPage: func(e *colly.Response) (Page, error) {
			return Page{
				Url:  e.Request.URL.String(),
				HTML: string(e.Body),
			}, nil
		},
	}
}

func NewHydacJob() Job {

	// if url following mask /shop/en/{number}{end of string} with regex
	exp, err := regexp.Compile(`\/shop\/en\/\d+$`)
	if err != nil {
		log.Fatal(err)
	}

	job := NewDefaultJob()
	job.AllowedDomains = []string{"www.hydac.com"}
	job.StartingURL = "https://www.hydac.com/shop/en"
	job.OnLink = func(e *colly.HTMLElement) error {

		url := e.Attr("href")
		if exp.MatchString(url) {
			return nil
		}

		return errors.New("url not following mask /shop/en/{number}")
	}
	job.OnPage = func(e *colly.Response) (Page, error) {

		// TODO: Save body to db

		fmt.Println("OnPage")
		return Page{}, nil
	}

	return job
}

func main() {

	jobs := []Job{
		NewHydacJob(),
	}

	p := NewProcessor(nil)

	for _, j := range jobs {
		p.Process(j)
	}

}

type PageStore interface {
	Save(IPage) error
}
type Processor struct {
	PageStore PageStore
}

func NewProcessor(s PageStore) Processor {
	return Processor{
		PageStore: s,
	}

}

func (p Processor) Process(j Job) {

	count = 0
	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains(j.AllowedDomains...),
	)

	// On every an element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {

		err := j.OnLink(e)
		if err != nil {
			fmt.Println("OnLink failed", err.Error())
			return
		}

		link := e.Attr("href")
		err = c.Visit(e.Request.AbsoluteURL(link))
		if err != nil {
			fmt.Println("Visiting failed", err.Error())
		}
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Visited", r.Request.URL)

		page, err := j.OnPage(r)
		if err != nil {
			fmt.Println("OnResponse failed", err.Error())
			return
		}

		err = p.PageStore.Save(page)
		if err != nil {
			fmt.Println("Page Save failed", err.Error())
			return
		}
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
		count++
		fmt.Println("Visited = ", count)
	})

	err := c.Visit(j.StartingURL)
	if err != nil {
		fmt.Println("Visiting failed!!!!! to ", err.Error())
	}
}

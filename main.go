package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gocolly/colly/v2"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"sync"
)

type CrawlOutput struct {
	Paths []string `yaml:"paths"`
}

var urls []string

func main() {
	var baseDomain string
	var allowedDomains string

	flag.StringVar(&baseDomain, "base-domain", "https://google.de", "Base domain for crawler")
	flag.StringVar(&allowedDomains, "allowed-domains", "google.de", "Allowed domain list")
	flag.Parse()

	mutex := sync.Mutex{}

	c := colly.NewCollector(
		colly.AllowedDomains(strings.Split(allowedDomains, ",")...),
		colly.Async(),
	)

	urls = make([]string, 0)

	_ = c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 8})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		err := c.Visit(e.Request.AbsoluteURL(link))
		if err != nil {
			if errors.Is(err, colly.ErrAlreadyVisited) {
				return
			}
			fmt.Printf("Error occured while visiting %s, %s\n", e.Request.AbsoluteURL(link), err.Error())
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Adding", r.URL.String())

		mutex.Lock()
		urls = append(urls, r.URL.String())
		mutex.Unlock()
	})

	err := c.Visit(baseDomain)
	if err != nil {
		fmt.Println("Error while visiting: ", err.Error())
		return
	}
	c.Wait()

	data := CrawlOutput{Paths: urls}

	d, err := yaml.Marshal(data)

	fmt.Printf("%s\n", d)

	if err != nil {
		fmt.Println("Error while marshalling: ", err.Error())
		return
	}

	err = os.WriteFile("spider.yaml", d, 0777)

	if err != nil {
		fmt.Println("Error while saving: ", err.Error())
	}
}

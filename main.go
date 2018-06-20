// CLI tool for extracting links for a provided base URL
package main

import (
	"flag"
	"fmt"

	"github.com/tamarakaufler/go-crawler/crawler"
)

var baseURL string
var depth int

func init() {
	flag.StringVar(&baseURL, "url", "https://docs.docker.com", "Base URL where the crawler starts. Default is https://docs.docker.com .")
	flag.IntVar(&depth, "depth", 3, "How deep the crawler goes. Up to 10 levels are supported. Default is 3.")
}

func main() {
	flag.Parse()

	var cc crawler.Crawler

	c := &crawler.Creeper{
		BaseURL: baseURL,
		Depth:   int8(depth),
	}

	cc = c

	err := cc.Run()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
}

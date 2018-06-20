package crawler

import (
	"io/ioutil"
	"strings"
)

// mock fetch using content in the mock dir
func mockFetch(testBase string) func(string) (string, error) {
	return func(url string) (string, error) {

		var file string
		if url == testBase {
			file = "./mock/basePage.html"
		} else {

			switch {
			case strings.Contains(url, "/faq"):
				file = "./mock/faq.html"
			case strings.Contains(url, "/about"):
				file = "./mock/about.html"
			case strings.Contains(url, "/careers"):
				file = "./mock/careers.html"
			case strings.Contains(url, "/info"):
				file = "./mock/info.html"
			case strings.Contains(url, "/generic"):
				file = "./mock/generic.html"
			}
		}

		body, err := ioutil.ReadFile(file)
		if err != nil {
			return "", err
		}
		return string(body), nil
	}
}

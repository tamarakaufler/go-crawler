package crawler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	ErrNoUrlProvided      = errors.New("No URL provided")
	ErrIncorrectUrlFormat = errors.New("Wrong URL format provided")
)

// Crawly interface must be satisfied to do the crawling
type Crawler interface {
	Run() error
	extractLinks(body string) []string
	process(int8, string, func(string) (string, error))
	display(int8, string)
}

type Creeper struct {
	BaseURL string
	Depth   int8

	pageScanner   *regexp.Regexp
	baseURLParsed *url.URL
	seen          chan *page
	fail          chan error
	sig           chan os.Signal
	done          chan struct{}
	seenLinks     map[string][]string
	wg            sync.WaitGroup
	muSeen        sync.Mutex
}

type page struct {
	url   string
	links []string
}

func (cc *Creeper) Run() error {
	start := time.Now()
	var elapsed time.Duration

	if err := inputCheck(cc); err != nil {
		return err
	}
	crawlerInit(cc)

	fmt.Println("\n--- Starting to crawl ---\n")

	// launch gouroutine to collect results, catch errors and trigger
	// sitemap display
	go func() {
		for {
			select {
			case page := <-cc.seen:
				cc.muSeen.Lock()
				cc.seenLinks[page.url] = page.links
				cc.muSeen.Unlock()
			case <-cc.done:
				elapsed = time.Since(start)
				break
			case err := <-cc.fail:
				fmt.Printf("\nfailure!: %v\n\n", err)
				os.Exit(1)
			case sig := <-cc.sig:
				elapsed = time.Since(start)
				cc.display(cc.Depth, "   ")
				fmt.Printf("\nInterrupted by %v after %s\n\n", sig, elapsed)
				os.Exit(1)
			}
		}
	}()

	// start processing the base URL
	//		concurrent processing of links
	cc.wg.Add(1)
	go func() {
		defer cc.wg.Done()
		fetch := fetch()
		cc.process(0, cc.BaseURL, fetch)
	}()

	cc.wg.Wait()
	cc.done <- struct{}{}

	cc.display(cc.Depth, "   ")
	fmt.Printf(">> The crawler took %s <<\n\n", elapsed)

	return nil
}

// inputCheck checks user setup
func inputCheck(cc *Creeper) error {
	if cc.BaseURL == "" {
		return ErrNoUrlProvided
	}
	trimmed := strings.Trim(cc.BaseURL, " ")
	trimmed = strings.TrimSuffix(trimmed, "/")
	cc.BaseURL = trimmed

	err := checkURL(cc.BaseURL)
	if err != nil {
		return err
	}
	baseURLParsed, err := url.ParseRequestURI(cc.BaseURL)
	if err != nil {
		return err
	}
	cc.baseURLParsed = baseURLParsed

	if cc.Depth > int8(10) {
		fmt.Printf("Up to 10 levels of crawling are allowed. Capping at 10.\n\n")
		cc.Depth = int8(10)
	}

	return err
}

func checkURL(url string) error {
	regStr := fmt.Sprint(`^http(s)?://[a-zA-Z0-9\-_.]+/?$`)
	regex := regexp.MustCompile(regStr)
	if regex.MatchString(url) {
		return nil
	}
	return ErrIncorrectUrlFormat
}

func crawlerInit(cc *Creeper) {
	pageScannerSetup(cc)
	cc.seenLinks = make(map[string][]string)
	cc.seen = make(chan *page)
	cc.fail = make(chan error)
	cc.done = make(chan struct{})

	cc.sig = make(chan os.Signal, 1)
	signal.Notify(cc.sig, syscall.SIGINT, syscall.SIGTERM)
}

func pageScannerSetup(cc *Creeper) {
	if cc.pageScanner == nil {
		cc.pageScanner = regexSetup(cc.BaseURL)
	}
}

func regexSetup(s string) *regexp.Regexp {
	regStr := fmt.Sprintf(`<a\s+(?:[a-zA-Z0-9_="\- ]+)?href="((?:%s)?/[a-zA-Z_0-9\-/&?]+)"\s*([a-z=]*)?(\s*/?>)?`, s)
	return regexp.MustCompile(regStr)
}

// process processes a page at a given url
// to find links for the given criteria
func (cc *Creeper) process(depth int8, url string, fetch func(string) (string, error)) {
	if depth > cc.Depth {
		return
	}
	if url == "" {
		cc.fail <- errors.New("incorrect input to process")
		return
	}

	cc.muSeen.Lock()
	if _, ok := cc.seenLinks[url]; ok {
		cc.muSeen.Unlock()
		return
	}
	cc.muSeen.Unlock()

	body, err := fetch(url)
	if err != nil {
		log.Printf("Error while fetching url: [%s]\n", url)
		return
	}
	links := cc.extractLinks(body)

	cc.seen <- &page{
		url:   url,
		links: links,
	}

	depth = depth + 1
	for _, link := range links {
		cc.muSeen.Lock()
		if _, ok := cc.seenLinks[link]; ok {
			cc.muSeen.Unlock()
			continue
		}
		cc.muSeen.Unlock()

		if url == link {
			continue
		}

		cc.wg.Add(1)
		go func(l string) {
			defer cc.wg.Done()
			cc.process(depth, l, fetch)
		}(link)
	}
}

// fetch retrieves content at the given URL
func fetch() func(string) (string, error) {
	return func(url string) (string, error) {
		res, err := http.Get(url)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", fmt.Errorf("Error retrieving content for url %s: %v", url, err)
		}

		return string(body), nil
	}
}

// extractLinks returns a list of urls
// - number of retrieved links is hardcoded to the maximum of 30
func (cc *Creeper) extractLinks(body string) []string {
	links := []string{}
	seen := map[string]struct{}{}

	matches := cc.pageScanner.FindAllStringSubmatch(body, 30)
	for _, m := range matches {
		l := m[1]

		u, err := url.Parse(l)
		if err != nil {
			log.Printf("%v", err)
			continue
		}
		if strings.Contains(l, "redirect") {
			continue
		}

		if !u.IsAbs() {
			ru := cc.baseURLParsed.ResolveReference(u)
			l = fmt.Sprintf("%s", ru)
		}
		if _, ok := seen[l]; ok {
			continue
		}
		seen[l] = struct{}{}
		links = append(links, l)
	}
	return links
}

// display displays the sitemap to the given depth
func (cc *Creeper) display(depth int8, offset string) {
	fmt.Println("👍 SiteMap display 👍\n")

	displayedPages := make(map[string]struct{})
	displayPageMap(cc.seenLinks, cc.Depth, displayedPages, offset, int8(0), cc.BaseURL)

	fmt.Println("\n👍 The END 👍")
}

func createOffset(offset string, depth int8) string {
	i := int8(0)
	off := ""
	for i <= depth {
		off = off + offset
		i++
	}
	return off
}

// displayPageMap provides recursive display for links
func displayPageMap(seenLinks map[string][]string, maxDepth int8, displayedPages map[string]struct{}, offset string, depth int8, url string) {
	urlOfs := createOffset(offset, depth)
	linkOfs := fmt.Sprintf("%s%s", urlOfs, offset)

	links := seenLinks[url]

	fmt.Println("================================")
	fmt.Printf("%s* %s (depth %d)\n", urlOfs, url, depth)
	fmt.Printf("%s number of links = %d\n", linkOfs, len(links))
	fmt.Println("--------------------------------")

	depth = depth + 1
	if depth > maxDepth {
		return
	}
	for i, l := range links {
		fmt.Printf("%s- %d - [%s]\n", linkOfs, i, l)
		if url == l {
			continue
		}
		if _, ok := displayedPages[url]; ok {
			fmt.Printf("%s (links displayed before)\n", linkOfs)
			continue
		}
		displayPageMap(seenLinks, maxDepth, displayedPages, offset, depth, l)
	}
	displayedPages[url] = struct{}{}
	fmt.Println("--------------------------------")
}

# go-crawler

 - Golang
 - CLI
 - concurrency
 - Docker
 - Makefile
 - unit testing

## REQUIREMENTS

Write a simple web crawler in Go. The crawler should be limited to one domain - so when you start with https://docs.docker.com/, it would crawl all pages within docs.docker.com, but not follow external links, for example to the Facebook and Twitter accounts. Given a URL, it should print a simple site map, showing the links between pages.

## IMPLEMENTATION

The implementation provides a CLI tool written in Go. The command accepts two flags:
  - url      in the form of http(s)://domain(/). Default is https://docs.docker.com
  - depth    indicating how deep the crawler should go. Maximum of 10 levels are accepted, default is 3

The crawler concurrently processes links, stopping when the given depth is exceeded. Number of retrieved links on a page is currenly hardcoded to 30. The crawler then prints
out the sitemap and shows how long the crawling took (excluding the display). 

## USAGE

a)
go run main.go (default values used)
go run main.go -url=.... -depth=...

b)
go build -o creepycrawly .
./creepycrawly
./creepycrawly  -url=.... -depth=...

c)
make dev
docker run --rm --name=creepycrawly quay.io/tamarakaufler/creepycrawly:v1alpha1  -url=https://docs.docker.com -depth=4

d)
make dev
CRAWLER_URL=https://docs.docker.com CRAWLER_DEPTH=2 make run

## CAVEATS

- Hardcoded number of retrieved links on a page : 30

## IMPROVEMENTS

- Provide a flag for setting the retrieved number of links on a page
- Provide a help flag showing basic command info and its usage
- Improve the sitemap display

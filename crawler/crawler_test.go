package crawler

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"sync"
	"testing"
)

var testBaseURL = "https://mmmmm.com"
var testBaseURLParsed, _ = url.Parse(testBaseURL)
var testPageScanner = regexSetup(testBaseURL)

func TestCreeper_extractLinks(t *testing.T) {
	type fields struct {
		BaseURL       string
		pageScanner   *regexp.Regexp
		baseURLParsed *url.URL
	}
	type args struct {
		body string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			// test 1
			name: "web page contains absolute links",
			fields: fields{
				BaseURL:       testBaseURL,
				pageScanner:   testPageScanner,
				baseURLParsed: testBaseURLParsed,
			},
			args: args{
				body: `<p><br /> We’ve had a fantastic time catching up with the Mmmmm community in <a href="https://mmmmm.com/t/mmmm-hitting-meetup/1111/22">Edinburgh, Bristol and Manchester</a> over the last few months. These meetups have cemented our commitment to keep spreading <a href="https://mmmmm.com/static/info">community forum</a> the word across the UK this year.</p>

				<p><a href="https://notthisone.com/about">community forum</a></p>
				<p>But planning events from 100 miles away isn’t easy and so we’d love to help you organise meetups in your city! There’s tons of potential for awesome events led by you and supported by us.<a href="https://mmmmm.com/static/images/favicon.png">community forum</a></p>

				<p>Create a platform to showcase the most exciting and inspiring community focused projects and ideas your city has to offer, with our help <a href="https://mmmmm.com/blog/2017/06/08/host-a-mmmm-meetup/">community forum</a> or a couple of Mmmmm team members to come along and help — talk to us about your idea and what you need to make it happen!</p><a href="https://mmmmm.com/static/info">community forum</a> 

				<p>In our experience <a class="c-footer__link" href="https://mmmmm.com/careers">careers</a>, some of the most successful events have begun with a conversation. With this in mind, the <a href="http://mmmmm.com/about">community forum</a> is a great place to head for some initial research <a href="https://mmmmm.com/community" class="o-button-text u-width-full u-width-auto-md u-margin-top">Join our community today &rsaquo;</a> and inspiration.</p>`,
			},
			want: []string{
				"https://mmmmm.com/t/mmmm-hitting-meetup/1111/22",
				"https://mmmmm.com/static/info",
				"https://mmmmm.com/blog/2017/06/08/host-a-mmmm-meetup/",
				"https://mmmmm.com/careers",
				"https://mmmmm.com/community",
			},
		},
		{
			// test 2
			name: "web page contains absolute https and http variations of given base URL",
			fields: fields{
				BaseURL:       testBaseURL,
				pageScanner:   testPageScanner,
				baseURLParsed: testBaseURLParsed,
			},
			args: args{
				body: `<p><br /> We’ve had a fantastic time catching up with the Mmmmm community in <a href="https://mmmmm.com/t/mmmm-hitting-meetup/1111/22">Edinburgh, Bristol and Manchester</a> over the last few months. These meetups have cemented our commitment to keep spreading <a href="http://mmmmm.com/static/info">community forum</a> the word across the UK this year.</p>

				<p><a href="https://notthisone.com/about">community forum</a></p>
				<p>But planning events from 100 miles away isn’t easy and so we’d love to help you organise meetups in your city! There’s tons of potential for awesome events led by you and supported by us.<a href="https://mmmmm.com/static/images/favicon.png">community forum</a></p>

				<p>Create a platform to showcase the most exciting and inspiring community focused projects and ideas your city has to offer, with our help <a href="https://mmmmm.com/blog/2017/06/08/host-a-mmmm-meetup/">community forum</a> or a couple of Mmmmm team members to come along and help — talk to us about your idea and what you need to make it happen!</p>

				<p>In our experience, some of the most successful events have begun with a conversation. With this in mind, the <a href="http://mmmmm.com/about">community forum</a>  is a great place to head for some initial research and inspiration.</p>`,
			},
			want: []string{
				"https://mmmmm.com/t/mmmm-hitting-meetup/1111/22",
				"https://mmmmm.com/blog/2017/06/08/host-a-mmmm-meetup/",
			},
		},
		{
			// test 3
			name: "web page contains absolute and relative links",
			fields: fields{
				BaseURL:       testBaseURL,
				pageScanner:   testPageScanner,
				baseURLParsed: testBaseURLParsed,
			},
			args: args{
				body: `<p><br /> We’ve had a fantastic time catching up with the Mmmmm community in <a href="https://mmmmm.com/t/mmmm-hitting-meetup/1111/22">Edinburgh, Bristol and Manchester</a> over the last few months. These meetups have cemented our commitment to keep spreading <a href="http://mmmmm.com/static/info">community forum</a> the word across the UK this year.</p>

				<p><a href="https://notthisone.com/about">community forum</a></p>
				<p>But planning events from 100 miles away isn’t easy and so we’d love to help you organise meetups in your city! There’s tons of potential for awesome events led by you and supported by us.<a href="https://mmmmm.com/static/images/favicon.png">community forum</a></p>

				<p>Create a platform to showcase the most exciting and inspiring community focused projects and ideas your city has to offer, with our help <a href="https://mmmmm.com/blog/2017/06/08/host-a-mmmm-meetup/">community forum</a> or a couple of Mmmmm team members to come along and help — talk to us about your idea and what you need to make it happen!</p>

				<p>In our experience, some of the most successful events have begun with a conversation. With this in mind, the <a href="/about">community forum</a>  is a great place to head for some initial research and inspiration.  <a  href="/blog/authors/naji-esiri">Our blog</a></p>`,
			},
			want: []string{
				"https://mmmmm.com/t/mmmm-hitting-meetup/1111/22",
				"https://mmmmm.com/blog/2017/06/08/host-a-mmmm-meetup/",
				"https://mmmmm.com/about",
				"https://mmmmm.com/blog/authors/naji-esiri",
			},
		},
		{
			// test 4
			name: "web page contains redirect URL",
			fields: fields{
				BaseURL:       testBaseURL,
				pageScanner:   testPageScanner,
				baseURLParsed: testBaseURLParsed,
			},
			args: args{
				body: `<p><br /> We’ve had a fantastic time catching up with the Mmmmm community in <a href="https://mmmmm.com/t/mmmm-hitting-meetup/1111/22">Edinburgh, Bristol and Manchester</a> over the last few months. These meetups have cemented our commitment to keep spreading <a href="http://mmmmm.com/static/info">community forum</a> the word across the UK this year.</p>

				<p><a href="https://notthisone.com/about">community forum</a></p>
				<p>But planning events from 100 miles away isn’t easy and so we’d love to help you organise meetups in your city! There’s tons of potential for awesome events led by you and supported by us.<a href="https://mmmmm.com/static/images/favicon.png">community forum</a></p>

				<p>Create a platform to showcase the most exciting and inspiring community focused projects and ideas your city has to offer, with our help <a href="/-play-store-redirect">community forum</a> or a couple of Mmmmm team members to come along and help — talk to us about your idea and what you need to make it happen!</p>

				<p>In our experience, some of the most successful events have begun with a conversation. With this in mind, the <a href="https://mmmmm.com/t/about">community forum</a>  is a great place to head for some initial research and inspiration.</p>`,
			},
			want: []string{
				"https://mmmmm.com/t/mmmm-hitting-meetup/1111/22",
				"https://mmmmm.com/t/about",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &Creeper{
				BaseURL:       tt.fields.BaseURL,
				pageScanner:   tt.fields.pageScanner,
				baseURLParsed: tt.fields.baseURLParsed,
			}
			got := cc.extractLinks(tt.args.body)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TestCreeper.extractLinks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreeper_Run(t *testing.T) {
	type fields struct {
		BaseURL string
		Depth   int8
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Incorrect user input: empty url",
			fields: fields{
				BaseURL: "",
				Depth:   int8(0),
			},
			wantErr: true,
		},
		{
			name: "Incorrect user input: incorrect schema",
			fields: fields{
				BaseURL: "htttp://aaa.com",
				Depth:   int8(3),
			},
			wantErr: true,
		},
		{
			name: "Incorrect user input: missing schema",
			fields: fields{
				BaseURL: "aaa.com",
				Depth:   int8(3),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &Creeper{
				BaseURL: tt.fields.BaseURL,
				Depth:   tt.fields.Depth,
			}
			if err := cc.Run(); (err != nil) != tt.wantErr {
				t.Errorf("TestCreeper.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreeper_inputCheck(t *testing.T) {
	type fields struct {
		BaseURL       string
		Depth         int8
		baseURLParsed *url.URL
	}

	tests := []struct {
		name    string
		fields  fields
		want    *Creeper
		wantErr bool
	}{
		// test 1
		{
			name: "Incorrect user input: empty url",
			fields: fields{
				BaseURL: "",
				Depth:   int8(0),
			},
			wantErr: true,
		},
		// test 2
		{
			name: "Incorrect user input: incorrect schema",
			fields: fields{
				BaseURL: "htttp://aaa.com",
				Depth:   int8(3),
			},
			wantErr: true,
		},
		// test 3
		{
			name: "Incorrect user input: missing schema",
			fields: fields{
				BaseURL: "aaa.com",
				Depth:   int8(3),
			},
			wantErr: true,
		},
		// test 4
		{
			name: "User input: depth exceeds permitted max",
			fields: fields{
				BaseURL: "https://aaa.com",
				Depth:   int8(15),
			},
			want: &Creeper{
				BaseURL: "https://aaa.com",
				Depth:   10,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &Creeper{
				BaseURL: tt.fields.BaseURL,
				Depth:   tt.fields.Depth,
			}
			if err := inputCheck(cc); (err != nil) != tt.wantErr {
				t.Errorf("TestCreeper_inputCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				cc.baseURLParsed = nil
				if !reflect.DeepEqual(cc, tt.want) {
					t.Errorf("TestCreeper.extractLinks() = %+v, want %+v", cc, tt.want)
				}
			}
		})
	}
}

func TestCreeper_process(t *testing.T) {
	type fields struct {
		BaseURL       string
		pageScanner   *regexp.Regexp
		Depth         int8
		baseURLParsed *url.URL
		seen          chan *page
		fail          chan error
		done          chan struct{}
		seenLinks     map[string][]string
		wg            sync.WaitGroup
		muSeen        sync.Mutex
	}
	type args struct {
		depth int8
		url   string
		fetch func(string) (string, error)
	}
	fetch := mockFetch(testBaseURL)
	basePage := "https://mmmmm.com"

	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string][]string
	}{
		// test 1
		{
			name: "processing to the depth of 1",
			fields: fields{
				BaseURL: basePage,
				Depth:   int8(1),
			},
			args: args{
				depth: int8(1),
				url:   basePage,
				fetch: fetch,
			},
			want: map[string][]string{
				"https://mmmmm.com":       []string{"https://mmmmm.com/faq", "https://mmmmm.com/about"},
				"https://mmmmm.com/faq":   []string{"https://mmmmm.com/about", "https://mmmmm.com/info"},
				"https://mmmmm.com/about": []string{"https://mmmmm.com/careers", "https://mmmmm.com/faq"},
			},
		},
		// test 2
		{
			name: "processing to the depth of 2",
			fields: fields{
				BaseURL: basePage,
				Depth:   int8(2),
			},
			args: args{
				depth: int8(1),
				url:   basePage,
				fetch: fetch,
			},
			want: map[string][]string{
				"https://mmmmm.com":         []string{"https://mmmmm.com/faq", "https://mmmmm.com/about"},
				"https://mmmmm.com/faq":     []string{"https://mmmmm.com/about", "https://mmmmm.com/info"},
				"https://mmmmm.com/about":   []string{"https://mmmmm.com/careers", "https://mmmmm.com/faq"},
				"https://mmmmm.com/info":    []string{"https://mmmmm.com/about", "https://mmmmm.com/generic"},
				"https://mmmmm.com/careers": []string{"https://mmmmm.com/generic"},
			},
		},
		// test 3
		{
			name: "processing to the depth of 8",
			fields: fields{
				BaseURL: basePage,
				Depth:   int8(8),
			},
			args: args{
				depth: int8(8),
				url:   basePage,
				fetch: fetch,
			},
			want: map[string][]string{
				"https://mmmmm.com":         []string{"https://mmmmm.com/faq", "https://mmmmm.com/about"},
				"https://mmmmm.com/faq":     []string{"https://mmmmm.com/about", "https://mmmmm.com/info"},
				"https://mmmmm.com/about":   []string{"https://mmmmm.com/careers", "https://mmmmm.com/faq"},
				"https://mmmmm.com/info":    []string{"https://mmmmm.com/about", "https://mmmmm.com/generic"},
				"https://mmmmm.com/careers": []string{"https://mmmmm.com/generic"},
				"https://mmmmm.com/generic": []string{},
			},
		},
		// test 5
		{
			name: "processing to the depth of 0",
			fields: fields{
				BaseURL: basePage,
				Depth:   int8(0),
			},
			args: args{
				depth: int8(0),
				url:   basePage,
				fetch: fetch,
			},
			want: map[string][]string{
				"https://mmmmm.com": []string{"https://mmmmm.com/faq", "https://mmmmm.com/about"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &Creeper{
				BaseURL:       tt.fields.BaseURL,
				Depth:         tt.fields.Depth,
				baseURLParsed: tt.fields.baseURLParsed,
				seen:          tt.fields.seen,
				fail:          tt.fields.fail,
				done:          tt.fields.done,
				seenLinks:     tt.fields.seenLinks,
			}
			inputCheck(cc)
			crawlerInit(cc)

			go func() {
				for {
					select {
					case page := <-cc.seen:
						cc.muSeen.Lock()
						cc.seenLinks[page.url] = page.links
						cc.muSeen.Unlock()
					case <-cc.done:
						break
					case err := <-cc.fail:
						fmt.Printf("\nfailure!: %v\n\n", err)
						// if !tt.wantErr {
						// 	t.Errorf("TestCreeper_process error = %v, wantErr %v", err, tt.wantErr)
						// }
						break
					default:
					}
				}
			}()

			cc.wg.Add(1)
			go func() {
				defer cc.wg.Done()
				cc.process(0, tt.args.url, tt.args.fetch)
			}()
			cc.wg.Wait()
			cc.done <- struct{}{}

			if !reflect.DeepEqual(cc.seenLinks, tt.want) {
				t.Errorf("TestCreeper_process = %+v, want %+v", cc.seenLinks, tt.want)
			}
		})
	}
}

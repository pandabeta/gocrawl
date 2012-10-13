package gocrawl

import (
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestAllSameHost(t *testing.T) {
	opts := NewOptions(nil, nil)
	opts.SameHostOnly = true
	opts.CrawlDelay = DefaultTestCrawlDelay
	spyv, spyu, _ := runFileFetcherWithOptions(opts, []string{"*"}, []string{"http://hosta/page1.html", "http://hosta/page4.html"})

	assertCallCount(spyv, 5, t)
	assertCallCount(spyu, 13, t)
}

func TestAllNotSameHost(t *testing.T) {
	opts := NewOptions(nil, nil)
	opts.SameHostOnly = false
	opts.CrawlDelay = DefaultTestCrawlDelay
	opts.LogFlags = LogError | LogTrace
	spyv, spyu, _ := runFileFetcherWithOptions(opts, []string{"*"}, []string{"http://hosta/page1.html", "http://hosta/page4.html"})

	assertCallCount(spyv, 10, t)
	assertCallCount(spyu, 24, t)
}

func TestSelectOnlyPage1s(t *testing.T) {
	opts := NewOptions(nil, nil)
	opts.SameHostOnly = false
	opts.CrawlDelay = DefaultTestCrawlDelay
	opts.LogFlags = LogError | LogTrace
	spyv, spyu, _ := runFileFetcherWithOptions(opts,
		[]string{"http://hosta/page1.html", "http://hostb/page1.html", "http://hostc/page1.html", "http://hostd/page1.html"},
		[]string{"http://hosta/page1.html", "http://hosta/page4.html", "http://hostb/pageunlinked.html"})

	assertCallCount(spyv, 3, t)
	assertCallCount(spyu, 11, t)
}

func TestRunTwiceSameInstance(t *testing.T) {
	spyv := newVisitorSpy(0, nil, true)
	spyu := newUrlSelectorSpy(0, "*")

	opts := NewOptions(nil, nil)
	opts.SameHostOnly = true
	opts.CrawlDelay = DefaultTestCrawlDelay
	opts.URLVisitor = spyv.f
	opts.URLSelector = spyu.f
	opts.LogFlags = LogNone
	opts.Fetcher = newFileFetcher("./testdata/")
	c := NewCrawlerWithOptions(opts)
	c.Run("http://hosta/page1.html", "http://hosta/page4.html")

	assertCallCount(spyv, 5, t)
	assertCallCount(spyu, 13, t)

	spyv = newVisitorSpy(0, nil, true)
	spyu = newUrlSelectorSpy(0, "http://hosta/page1.html", "http://hostb/page1.html", "http://hostc/page1.html", "http://hostd/page1.html")
	opts.URLVisitor = spyv.f
	opts.URLSelector = spyu.f
	opts.SameHostOnly = false
	c.Run("http://hosta/page1.html", "http://hosta/page4.html", "http://hostb/pageunlinked.html")

	assertCallCount(spyv, 3, t)
	assertCallCount(spyu, 11, t)
}

func TestIdleTimeOut(t *testing.T) {
	opts := NewOptions(nil, nil)
	opts.SameHostOnly = false
	opts.WorkerIdleTTL = 200 * time.Millisecond
	opts.CrawlDelay = DefaultTestCrawlDelay
	opts.LogFlags = LogInfo
	_, _, b := runFileFetcherWithOptions(opts,
		[]string{"*"},
		[]string{"http://hosta/page1.html", "http://hosta/page4.html", "http://hostb/pageunlinked.html"})

	assertIsInLog(*b, "worker for host hostd cleared on idle policy\n", t)
	assertIsInLog(*b, "worker for host hostunknown cleared on idle policy\n", t)
}

func TestReadBodyInVisitor(t *testing.T) {
	var err error
	var b []byte

	c := NewCrawler(func(res *http.Response, doc *goquery.Document) ([]*url.URL, bool) {
		b, err = ioutil.ReadAll(res.Body)
		return nil, false
	}, nil)

	c.Options.Fetcher = newFileFetcher("./testdata/")
	c.Options.CrawlDelay = DefaultTestCrawlDelay
	c.Options.LogFlags = LogAll
	c.Run("http://hostc/page3.html")

	if err != nil {
		t.Error(err)
	} else if len(b) == 0 {
		t.Error("Empty body")
	}
}

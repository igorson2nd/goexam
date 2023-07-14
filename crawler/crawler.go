package crawler

import "goexam/page"
import "goexam/href"
import "goexam/database"
import "sync"
import "sync/atomic"
// import "strings"
import "time"
import "os"
import "fmt"

type Crawler struct {
	BaseUrl string
	hrefs_visited sync.Map
	hrefs_new sync.Map
	Pages sync.Map
	channel chan bool

	Concurrency uint
	Crawl_limit uint32
	crawl_count uint32

	initialized bool
	sigterm bool
	SignalChannel chan os.Signal

	DB database.Database
}

func (c *Crawler) Crawl() {
	if ! c.initialized {
		c.init()
	}

	ticker := time.NewTicker(3 * time.Second)
	go func(){
		for ; ! c.sigterm; {
			<-ticker.C
			fmt.Println("  page count:", c.page_count())
		}
	}()

	for {
		c.channel <- true
		go c.search_new_page()

		if c.is_crawling_finished() {
			break
		}
	}
}

func sync_map_pop(sm *sync.Map) any {
	var k any
	sm.Range(func(key any, val any) bool {
		k = key
		return false
	})
	sm.Delete(k)
	return k
}

func (c *Crawler) search_new_page() {
	defer func() {<-c.channel}()
	any_ref := sync_map_pop(&c.hrefs_new)
	if any_ref == nil {
		return
	}
	ref := any_ref.(href.Href)
	
	c.hrefs_visited.Store(ref, true)
	atomic.AddUint32(&c.crawl_count, 1)
	// fmt.Println("Searching: ", ref.ToStr())
	time.Sleep(600 * time.Millisecond)

	p, page_refs := page.SearchPage(c.BaseUrl + ref.ToStr())
	page_id := p.ID.GetStrID()
	if (page_id != "") && (p.Title != "") {
		c.Pages.Store(p.ID, p)
		// fmt.Println(p.ID, p.Title)
		c.DB.SavePage(p)
	}

	for _, next_href := range page_refs {
		if visited, _ := c.hrefs_visited.Load(next_href); visited == true {
			continue
		}
		c.hrefs_new.Store(next_href, true)
	}
}

func (c *Crawler) init() {
	c.hrefs_new.Store(href.New(""), true)
	c.channel = make(chan bool, c.Concurrency)
	c.initialized = true

	if c.Concurrency > 5 {
		fmt.Println("Warning: number of concurrent goroutines is set to", c.Concurrency)
		fmt.Println("Recommended value is 5.")
	}

	go func(){
		<-c.SignalChannel
		c.sigterm = true
		fmt.Println(">>> Crawling interrupted !")
	}()
}

func (c *Crawler) is_crawling_finished() bool {
	if c.sigterm {
		return true
	}
	if atomic.LoadUint32(&c.crawl_count) > c.Crawl_limit {
		return true
	}

	any_elements := false
	c.hrefs_new.Range(func(id any, val any) bool {
		any_elements = true
		return false
	})

	if any_elements {
		return false
	}

	time.Sleep(3600 * time.Millisecond)
	// wait for http response
	// fmt.Println("second try")
	c.hrefs_new.Range(func(key any, val any) bool {
		any_elements = true
		return false
	})
	return ! any_elements
}

func (c *Crawler) page_count() uint {
	var pages_count uint
	pages_count = 0 
	c.Pages.Range(func(key any, val any) bool {
		pages_count++
		return true
	})
	return pages_count
}

func (c *Crawler) Interrupted() bool {
	return c.sigterm
}
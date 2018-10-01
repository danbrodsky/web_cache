package page_scraper

import (
        "testing"
        //"time"
        "fmt"
)

var (
        ps PS
)

func TestNewPageScraperAndExecute(t *testing.T) {
	ps = NewPageScraper("http://vaastavanand.com/")
	_= ps
	page, err := ps.execute()
	fmt.Println(page.Links)
	fmt.Println(page.Scripts)
	fmt.Println(page.Images)
	if err != nil {
                t.Errorf("%+v", err)
        }
}


package page_scraper

import (
        "testing"
        //"time"
        //"fmt"
)

var (
        rs RS
)

func TestInitialize(t *testing.T) {
	rs = NewResourceScraper()
	_= rs
}

func TestScrapeFile(t *testing.T) {
	_,err := rs.ScrapeResource("http://vaastavanand.com/favicon.ico")
	if err != nil {
                t.Errorf("%+v", err)
        }

	_,err = rs.ScrapeResource("https://www.cs.ubc.ca/~wolf/")
        if err == nil {
                t.Errorf("should have failed")
        }
}


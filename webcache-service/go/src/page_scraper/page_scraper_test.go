package page_scraper

import (
        "testing"
        //"time"
        "fmt"
	"os"
)

var (
        ps PS
)

func TestNewPageScraperAndExecute(t *testing.T) {
	h := os.Getenv("DEPLOY_HOST_IP") + ":" + os.Getenv("DEPLOY_HOST_PORT") + "/" + os.Getenv("RES_ENTRYPOINT")
	fmt.Println("source path: " + h)
	ps = NewPageScraper("http://vaastavanand.com/favicon.ico")
	_= ps
	page, err := ps.GetPage()
	fmt.Println(page.Links)
	fmt.Println(page.Scripts)
	fmt.Println(page.Images)
//        fmt.Println(page.Html)
	if err != nil {
                t.Errorf("%+v", err)
        }
        scrapedPage, err := ps.ScrapePage(page)
	fmt.Println(scrapedPage.Html)
        fmt.Println(scrapedPage.Links)
        fmt.Println(scrapedPage.Scripts)
        fmt.Println(scrapedPage.Images)
}


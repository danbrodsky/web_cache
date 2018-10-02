package page_scraper

import (
    "bytes"
    "disk_client"
    "fmt"
    "golang.org/x/net/html"
    "io"
    "io/ioutil"
    "net/http"
    "os"
    "strings"
    "unsafe"
)


type PageScraper struct {
	Url string
}

type PS interface {
	// gets page and caches resources on disk for tags html link, script, img and updates the tags according to it
	Execute() (page disk_client.Page, err error)
}

func NewPageScraper(url string) (pageScraper PS) {
        pageScraper = PageScraper{ Url: url}
        return pageScraper
}


func getBody(doc *html.Node, page *disk_client.Page) {
    var f func(*html.Node)
    f = func(n *html.Node) {
        if n.Type == html.ElementNode && n.Data == "link" {
                res := getHref(n.Attr)
                ref,err := scrapeResource(formatUrl(res.Val, page.Url))
                if err == nil{
			page.Links = append(page.Links, ref)
                        res.Val = ref
                }
        } else if n.Type == html.ElementNode && n.Data == "script" {
                res := getSrc(n.Attr)
                ref,err := scrapeResource(formatUrl(res.Val, page.Url))
                if err == nil{
			page.Scripts = append(page.Scripts, ref)
                        res.Val = ref
                }
        } else if n.Type == html.ElementNode && n.Data == "img" {
		res := getSrc(n.Attr)
		ref,err := scrapeResource(formatUrl(res.Val, page.Url))
		if err == nil{
			page.Images = append(page.Images, ref)
			res.Val = ref
		}
        }

        for c := n.FirstChild; c != nil; c = c.NextSibling {
            f(c)
        }
    }
    f(doc)
    //fmt.Println(renderNode(doc))
    page.Html = renderNode(doc)
}

func getHref(attr []html.Attribute) (att *html.Attribute) {
        for i,a := range attr {
                if a.Key == "href" {
                        return &attr[i]
                }

        }
        return &html.Attribute{}

}

func getSrc(attr []html.Attribute) (att *html.Attribute) {
        for i,a := range attr {
                if a.Key == "src" {
                        return &attr[i]
                }

        }
        return &html.Attribute{}

}


func scrapeResource(url string) (ref string,err error){
        rs := NewResourceScraper()
        ref, err = rs.ScrapeResource(url)
        fmt.Println(ref)
        if err != nil{
		return "",err
        }
	return ref,nil
}

func formatUrl(uri string,url string) string {
        if(!strings.Contains(uri, "http")){
                return url+uri
        } else{
                return uri
        }

}

func renderNode(n *html.Node) string {
    var buf bytes.Buffer
    w := io.Writer(&buf)
    html.Render(w, n)
    return buf.String()
}

func GetHtml(url string) (text string, err error) {
    var bytes []byte
    resp, err := http.Get(url)
    if err != nil {
        fmt.Println("There seems to ben an error with the page")
    }
    bytes, err = ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("Cannot read byte response")
    }
    text = string(bytes)

    return text, err
}

func getPageSize(p disk_client.Page) (int64) {

    var size int64 = 0
    for _, i := range p.Images {
        fi, _ := os.Stat(i)
        size += fi.Size()
    }
    for _, l := range p.Links {
        fi, _ := os.Stat(l)
        size += fi.Size()
    }
    for _, s := range p.Scripts {
        fi, _ := os.Stat(s)
        size += fi.Size()
    }
    size += int64(unsafe.Sizeof(p.Html))

    return size
}

func (ps PageScraper) Execute() (page disk_client.Page, err error) {
    htmlSrc, err := GetHtml(ps.Url)
    if err != nil {
        fmt.Println("Cannot read HTML source code.")
	return page,err
    }

    page = disk_client.Page{ Url: ps.Url}
    page.Html = htmlSrc
    page.Size = getPageSize(page)
    if err != nil {
        return page,err
    }
    return page,nil
}


package page_scraper

import (
    "os"
    "errors"
    "bytes"
    "fmt"
    "golang.org/x/net/html"
    "io"
    "io/ioutil"
    "net/http"
    "strings"
    "diskclient"
    "sync"
)


type PageScraper struct {
	Url string
}

type PS interface {
	GetPage() (page diskclient.Page, err error)
        ScrapePage(page diskclient.Page) (ScrapedPage diskclient.Page, err error)
}

var (
	HOSTPORT string
)

func NewPageScraper(url string) (pageScraper PS) {
        pageScraper = PageScraper{ Url: url}
        return pageScraper
}


func getHtml(doc *html.Node, page *diskclient.Page) {
    var f func(*html.Node)
    //var wg sync.WaitGroup
    f = func(n *html.Node) {
        //defer wg.Done()
        if n.Type == html.ElementNode && n.Data == "link" {
                res := getHref(n.Attr)
		if(res != nil){
			url := formatUrl(res.Val, page.Url)
                        //ref,err := scrapeResource(formatUrl(res.Val, page.Url))
                        //if err == nil{
			        page.Links = append(page.Links, url)
                                //res.Val = HOSTPORT + ref
                        //}
		}
        } else if n.Type == html.ElementNode && n.Data == "script" {
                res := getSrc(n.Attr)
		if(res != nil){
			url := formatUrl(res.Val, page.Url)
                        //ref,err := scrapeResource(formatUrl(res.Val, page.Url))
                        //if err == nil{
			        page.Scripts = append(page.Scripts, url)
                                //res.Val = HOSTPORT + ref
                        //}
		}
        } else if n.Type == html.ElementNode && n.Data == "img" {
		res := getSrc(n.Attr)
		if(res != nil){
			url := formatUrl(res.Val, page.Url)
		        //ref,err := scrapeResource(formatUrl(res.Val, page.Url))
		        //if err == nil{
			       page.Images = append(page.Images, url)
			       //res.Val = HOSTPORT + ref
		        //}
		}
        }

        for c := n.FirstChild; c != nil; c = c.NextSibling {
            //wg.Add(1)
            f(c)
        }
    }
    //wg.Add(1)
    f(doc)
    //fmt.Println(renderNode(doc))
    //wg.Wait()
    fmt.Println("done!")
    page.Html = renderNode(doc)
}

func scrapeHtml(doc *html.Node, page *diskclient.Page) {
    var f func(*html.Node)
    var wg sync.WaitGroup
    f = func(n *html.Node) {
        defer wg.Done()
        if n.Type == html.ElementNode && n.Data == "link" {
                res := getHref(n.Attr)
                if(res != nil){
                        ref,err := scrapeResource(formatUrl(res.Val, page.Url))
                        if err == nil{
                                page.Links = append(page.Links, ref)
                                res.Val = HOSTPORT + ref
                        }
                }
        } else if n.Type == html.ElementNode && n.Data == "script" {
                res := getSrc(n.Attr)
                if(res != nil){
                        ref,err := scrapeResource(formatUrl(res.Val, page.Url))
                        if err == nil{
                                page.Scripts = append(page.Scripts, ref)
                                res.Val = HOSTPORT + ref
                        }
                }
        } else if n.Type == html.ElementNode && n.Data == "img" {
                res := getSrc(n.Attr)
                if(res != nil){
                        ref,err := scrapeResource(formatUrl(res.Val, page.Url))
			fmt.Println(ref)
                        if err == nil{
                               page.Images = append(page.Images, ref)
                               res.Val = HOSTPORT + ref
                        }
                }
        }

        for c := n.FirstChild; c != nil; c = c.NextSibling {
            wg.Add(1)
            go f(c)
        }
    }
    wg.Add(1)
    go f(doc)
  //  fmt.Println(renderNode(doc))
    wg.Wait()
    fmt.Println("done!")
    page.Html = renderNode(doc)
}

func getHref(attr []html.Attribute) (att *html.Attribute) {
	fmt.Println(attr)
        for i,a := range attr {
                if a.Key == "href" {
                        return &attr[i]
                }

        }
        return att

}

func getSrc(attr []html.Attribute) (att *html.Attribute) {
	fmt.Println(attr)
        for i,a := range attr {
                if a.Key == "src" {
                        return &attr[i]
                }

        }
        return att

}


func scrapeResource(url string) (ref string,err error){
        rs := NewResourceScraper()
        ref, err = rs.ScrapeResource(url)
        fmt.Println(ref)
        if err != nil{
		return ref,err
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
    fmt.Println(url)
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

func (ps PageScraper) GetPage() (page diskclient.Page, err error) {
    HOSTPORT = "http://" + os.Getenv("DEPLOY_HOST_IP") + ":" + "8000"
    if(len(HOSTPORT) < 1){
	return page,errors.New("host port environment variables not set")
    }
    htmlSrc, err := GetHtml(ps.Url)
    if err != nil {
        fmt.Println("Cannot read HTML source code.")
	return page,err
    }

    page = diskclient.Page{ Url: ps.Url}
    //fmt.Println(strings.NewReader(htmlSrc))
    doc, _ := html.Parse(strings.NewReader(htmlSrc))
    getHtml(doc, &page)
    if err != nil {
        return page,err
    }

    //page2 := diskclient.Page{ Url: ps.Url}
    //doc2,_ := html.Parse(strings.NewReader(page.Html))
    //getBody(doc2, &page2)
    return page,nil
}

func (ps PageScraper) ScrapePage(page diskclient.Page) (ScrapedPage diskclient.Page, err error){
    HOSTPORT = "http://" + os.Getenv("DEPLOY_HOST_IP") + ":" + "8000"
    if(len(HOSTPORT) < 1){
        return page,errors.New("host port environment variables not set")
    }
    ScrapedPage = diskclient.Page{ Url: ps.Url}
    doc ,_ := html.Parse(strings.NewReader(page.Html))
    scrapeHtml(doc, &ScrapedPage)
    if err != nil {
        return ScrapedPage,err
    }
    return ScrapedPage,err
}


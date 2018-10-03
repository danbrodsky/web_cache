package diskclient

import (
	"testing"
	"time"
	"fmt"
)

var (
	dc DC
	testPage Page
)

func TestInitialize(t *testing.T) {
        var cacheCapacity int64 = 128
        var err error = nil
        dc, err = Initialize(cacheCapacity)
        if err != nil {
                t.Errorf("%+v", err)
        }
        _,  err = Initialize(cacheCapacity)
        if err != nil {
                t.Errorf("")
        }
}

func TestAddPage(t *testing.T) {
	arr := []string{ "123","456","79"}
	testPage = Page{ Url:"http://example.com", Timestamp: time.Now().Unix(), Images:arr, Links:arr, Scripts:arr, Html:"html" }
	dc.DeletePage(testPage.Url) // in case previous test failed
	if dc.AddPage(testPage) != 1 {
		t.Errorf("some went wrong return code should be 1")
	}

//	if dc.AddPage(testPage) != -1 {
//                t.Errorf("some went wrong adding duplicate pages to the set should return -1")
//        }
}

func TestGetPage(t *testing.T) {
	page, err := dc.GetPage(testPage.Url)
	if(err != nil){
		 t.Errorf("%+v", err)
	}
	if(testPage.Timestamp != page.Timestamp && testPage.Url != page.Url && page.Html != page.Html){
		t.Fail()
	}

}

func TestRefreshPageTimestamp(t *testing.T) {
	time.Sleep(1 * time.Second)
        timestamp ,_ := dc.RefreshPageTimestamp(testPage.Url)
	fmt.Println(timestamp)
	fmt.Println(testPage.Timestamp)
        if(timestamp == -1) {
                t.Errorf("err timestamp should of been updated")
        }
        if timestamp == testPage.Timestamp {
                t.Errorf("something went wrong test timestamp and page timestamp should not be equal")
        }

        page, err := dc.GetPage(testPage.Url)
        if(err != nil){
                 t.Errorf("%+v", err)
        }
        if(testPage.Timestamp != page.Timestamp && testPage.Url != page.Url && page.Html != page.Html){
                t.Fail()
        }
}


func TestDeletePage(t *testing.T) {
	if dc.DeletePage(testPage.Url) != 1 {
                t.Errorf("some went wrong return code should be 1")
        }
	if dc.DeletePage(testPage.Url) != -1 {
                t.Errorf("some went wrong return code should be -1")
        }

}


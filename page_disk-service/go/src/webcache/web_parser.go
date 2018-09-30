package main

import "golang.org/x/net/html"
import "fmt"

func main() {
	response, err := http.Get("https://www.cs.ubc.ca/~wolf/")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	htmlTokens := html.NewTokenizer(response.Body)
	loop:
		for {
			tt := htmlTokens.Next()
			fmt.Printf("%T", tt)
			switch tt {
				case html.ErrorToken:
					fmt.Println("End")
					break loop
				case html.TextToken:
					fmt.Println(tt)
				case html.StartTagToken:
					t := htmlTokens.Token()
					isAnchor := t.Data == "a"
				if isAnchor {
					fmt.Println("We found an anchor!")
				}
			}
		}
}


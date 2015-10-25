package main

import (
	"os"
	"fmt"
	"net/http"
	"strings"
	"code.google.com/p/go.net/html"
)

func PrintSlice(slice []string) {
    fmt.Printf("Slice length = %d\r\n", len(slice))
    for i := 0; i < len(slice); i++ {
        fmt.Printf("[%d] := %s\r\n", i, slice[i])
    }
}

func parse_html(slice *[]string,n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, element := range n.Attr {			
			if element.Key == "class" && element.Val == "torrent" {
				parse_torrent(slice,n)				
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		parse_html(slice,c)
	}
}

func parse_torrent(slice *[]string,n *html.Node) {	
	var id string
	finished:= false
	for c := n.FirstChild; c != nil; c = c.NextSibling {		
		if c.Type == html.ElementNode && c.Data == "div" {
			for _, element := range c.Attr {				
				if element.Key == "class" && element.Val == "torrentDetails" {
					for d:= c.FirstChild; d != nil; d = d.NextSibling {						
						if len(d.Attr) == 0 && strings.Contains(d.FirstChild.Data,"Seeding") {
							finished = true
						}
					}
					
				}
			}
		}	
		if c.Type == html.ElementNode && c.Data == "form" {			
			for _, element := range c.Attr {
				if element.Key == "action" {
					val := element.Val;
					values := strings.Split(val,"/");
					if( len(values) >= 4){
						if(values[2] == "delete"){
							id = values[3]					
						}
					}
				}
			}
		}
	}
	if(finished){
		*slice = append(*slice,id);
	}
	
}

func main(){
	fmt.Println(len(os.Args), os.Args)
	var slice []string
	if(len(os.Args) >= 2){
		base_url := os.Args[1];
		res, err := http.Get(base_url+"/torrents")
		if err != nil {
			panic(err)
		}

		defer res.Body.Close();
		doc, err := html.Parse(res.Body)
	  	parse_html(&slice,doc)
	  	PrintSlice(slice);

	  	for i := 0; i < len(slice); i++ {
	        url := base_url+"/cmd/remove/"+slice[i]
	        req, err := http.NewRequest("POST", url, nil);    	
	    	client := &http.Client{}
	    	resp, err := client.Do(req)
	    	 if err != nil {
	        	panic(err)
	    	}
	    	defer resp.Body.Close()
	    }
	}
  	

}
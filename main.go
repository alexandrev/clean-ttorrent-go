package main

import (
	"os"
	"fmt"
	"net/http"
	"strings"
	"golang.org/x/net/html"
	"time"
	"strconv"
	"log"
)

func printSlice(slice []string) {
    log.Printf("Number of recovered items = %d\r\n", len(slice))
    for i := 0; i < len(slice); i++ {
        log.Printf("[%d] := %s\r\n", i, slice[i])
    }
}

func parseHTML(slice *[]string,n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, element := range n.Attr {			
			if element.Key == "class" && element.Val == "torrent" {
				parseTorrent(slice,n)				
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		parseHTML(slice,c)
	}
}

func parseTorrent(slice *[]string,n *html.Node) {	
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

func getBaseURL(initialBaseURL string) string {
    baseURL:= initialBaseURL;
    found := false;
    idx := 0;    
    factor := 1;
    initialLastIPFragment := strings.LastIndex(baseURL,".");
    endLastIPFragment := strings.LastIndex(baseURL,":");
    baseURLFragment := baseURL[0:initialLastIPFragment];
    lastIPFragment := baseURL[initialLastIPFragment+1:endLastIPFragment];
    port, _:= strconv.Atoi(lastIPFragment);
    
    
    for( !found ) {
        portString  := strconv.Itoa(port);
        baseURL = baseURLFragment + "." + portString + ":1080";
        log.Printf("Trying to reach tTorrent at the IP: %s", baseURL)
        _, err := http.Get(baseURL +"/torrents")
        if err == nil {
            found = true;
        }
        idx = idx + 1;
        port = port + idx*factor;
        factor = factor * (-1);        
    }
    log.Printf("Final value: %s", baseURL)
    return baseURL;
        
}

func main(){
	fmt.Println(len(os.Args), os.Args)
	var slice []string
	if(len(os.Args) >= 3){
		baseURL := os.Args[1];
		log.Printf("Using the following URL as tTorrent PRO Web Interface: %s ",baseURL);
		sleepTime, _ := strconv.Atoi(os.Args[2]);
		log.Printf("Refresh Rate: %d secs.",sleepTime);
		exit := true;
		
        baseURL = getBaseURL(baseURL);
        
		for(exit == true){
            slice = slice[:0]
			log.Printf("New attept to clean the downloaded items");
			res, err := http.Get(baseURL+"/torrents")
			if err != nil {
				panic(err)
			}
	
			defer res.Body.Close();
			doc, err := html.Parse(res.Body)
			parseHTML(&slice,doc)
			printSlice(slice);
	
			for i := 0; i < len(slice); i++ {
				url := baseURL+"/cmd/remove/"+slice[i]
				req, err := http.NewRequest("POST", url, nil);    	
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
			}
				time.Sleep(time.Duration(sleepTime) * time.Second);
		}
	}
}
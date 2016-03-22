package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

func hello(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", c.URLParams["name"])
}

func getRoomByName(c web.C, w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms?search="+c.URLParams["name"], nil)
	check(err)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	check(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	check(err)

	fmt.Fprintf(w, "%s", body)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	goji.Get("/hello/:name", hello)
	goji.Get("/room/:name", getRoomByName)
	goji.Serve()
}

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"github.com/ziutek/telnet"
)

func hello(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", c.URLParams["name"])
}

func getRoomByName(c web.C, w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms?search="+c.URLParams["name"], nil)
	checkErr(err)

	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	checkErr(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	checkErr(err)

	fmt.Fprintf(w, "%s", body)
}

func sendCommand(t *telnet.Conn, command string) {
	buf := make([]byte, len(command)+1)
	copy(buf, command)
	buf[len(command)] = '\n'
	_, err := t.Write(buf)
	checkErr(err)
}

func telnetDial(c web.C, w http.ResponseWriter, r *http.Request) {
	command := c.URLParams["command"]

	t, err := telnet.Dial("tcp", "10.6.36.53:23")
	checkErr(err)

	t.SetUnixWriteMode(true)

	command = command + "\nhostname"
	sendCommand(t, command)
	//sendCommand(t, "hostname")
	t.SkipUntil(">")
	var stringy []byte
	stringy, err = t.ReadUntil("TSW-750>")
	checkErr(err)
	fmt.Printf("%s\n", stringy)

	//checkErr(err)
	fmt.Printf("doneskyes")

	t.Close() // Close the telnet session
	fmt.Fprintf(w, "%s", err)
}

func log(message string) {
	fmt.Printf("%s\n", message)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	goji.Get("/hello/:name", hello)
	goji.Get("/room/:name", getRoomByName)
	goji.Get("/telnet/:command", telnetDial)
	goji.Serve()
}

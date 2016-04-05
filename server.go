package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"github.com/ziutek/telnet"
)

func health(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Uh, we had a slight weapons malfunction, but uh... everything's perfectly all right now. We're fine. We're all fine here now, thank you. How are you?")
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
	command = command + "\nhostname" // Send two commands so we get a second prompt to use as a delimiter
	buf := make([]byte, len(command)+1)
	copy(buf, command)
	buf[len(command)] = '\n'
	_, err := t.Write(buf)
	checkErr(err)
}

func getTelnetOutput(c web.C, w http.ResponseWriter, r *http.Request) {
	command := c.URLParams["command"]
	address := c.URLParams["address"]
	port := c.URLParams["port"]

	t, err := telnet.Dial("tcp", address+":"+port)
	checkErr(err)

	t.SetUnixWriteMode(true) // Convert any '\n' (LF) to '\r\n' (CR LF)

	sendCommand(t, command)
	t.SkipUntil("TSW-750>") // Skip to the first prompt delimiter
	var output []byte
	output, err = t.ReadUntil("TSW-750>") // Read until the second prompt delimiter (provided by sending two commands in sendCommand)
	checkErr(err)

	t.Close() // Close the telnet session

	output = output[:len(output)-10] // Ghetto trim the prompt off the response

	fmt.Fprintf(w, "%s", output)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	goji.Get("/health", health)
	goji.Get("/room/:name", getRoomByName)
	goji.Get("/telnet/address/:address/port/:port/command/:command", getTelnetOutput)
	goji.Serve()
}

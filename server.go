package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"github.com/ziutek/telnet"
)

func checkErr(err error) {
	if err != nil {
		panic(err) // Don't forget your towel
	}
}

func health(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Uh, we had a slight weapons malfunction, but uh... everything's perfectly all right now. We're fine. We're all fine here now, thank you. How are you?")
}

func getTelnetOutput(c web.C, w http.ResponseWriter, r *http.Request) {
	command := c.URLParams["command"]
	address := c.URLParams["address"]
	port := c.URLParams["port"]

	t, err := telnet.Dial("tcp", address+":"+port)
	checkErr(err)

	t.SetUnixWriteMode(true) // Convert any '\n' (LF) to '\r\n' (CR LF)

	command = command + "\nhostname" // Send two commands so we get a second prompt to use as a delimiter
	buf := make([]byte, len(command)+1)
	copy(buf, command)
	buf[len(command)] = '\n'
	_, err = t.Write(buf)
	checkErr(err)

	t.SkipUntil("TSW-750>") // Skip to the first prompt delimiter
	var output []byte
	output, err = t.ReadUntil("TSW-750>") // Read until the second prompt delimiter (provided by sending two commands in sendCommand)
	checkErr(err)

	t.Close() // Close the telnet session

	output = output[:len(output)-10] // Ghetto trim the prompt off the response

	fmt.Fprintf(w, "%s", output)
}

func fusionRequest(requestType string, url string) string {
	client := &http.Client{}
	req, err := http.NewRequest(requestType, url, nil)
	checkErr(err)

	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	checkErr(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	checkErr(err)

	return string(body) // Convert the bytes to a string before returning
}

func getRooms(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", fusionRequest("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms"))
}

func getRoomByName(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", fusionRequest("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms?search="+c.URLParams["room"]))
}

func main() {
	// Endpoints for debugging
	goji.Get("/telnet/address/:address/port/:port/command/:command", getTelnetOutput)

	// Production endpoints
	goji.Get("/health", health)

	goji.Get("/rooms", getRooms)
	goji.Get("/rooms/:room", getRoomByName)
	// goji.Get("/buildings", ...)
	// goji.Get("/buildings/:building", ...)
	// goji.Get("/buildings/:building/rooms/:room", ...)
	// goji.Get("/buildings/:building/rooms/:room/signals", ...)
	// goji.Get("/buildings/:building/rooms/:room/signals/:signal", ...)
	//
	// goji.Post("/rooms", ...)
	// goji.Post("/buildings", ...)
	// goji.Post("/buildings/:building/rooms/:room/signals", ...)
	//
	// goji.Put("/rooms/:room", ...)
	// goji.Put("/buildings/:building", ...)
	// goji.Put("/buildings/:building/rooms/:room", ...)
	// goji.Put("/buildings/:building/rooms/:room/signals/:signal", ...)
	//
	// goji.Delete("/rooms/:room", ...)
	// goji.Delete("/buildings/:building", ...)
	// goji.Delete("/buildings/:building/rooms/:room", ...)
	// goji.Delete("/buildings/:building/rooms/:room/signals/:signal", ...)

	goji.Serve() // Serve that puppy
}

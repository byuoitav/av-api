package helpers

import "github.com/ziutek/telnet"

func GetTelnetOutput(address string, port string, command string) (string, error) {
	t, err := telnet.Dial("tcp", address+":"+port)
	if err != nil {
		return "", err
	}

	t.SetUnixWriteMode(true) // Convert any '\n' (LF) to '\r\n' (CR LF)

	command = command + "\nhostname" // Send two commands so we get a second prompt to use as a delimiter
	buf := make([]byte, len(command)+1)
	copy(buf, command)
	buf[len(command)] = '\n'
	_, err = t.Write(buf)
	if err != nil {
		return "", err
	}

	t.SkipUntil("TSW-750>") // Skip to the first prompt delimiter
	var output []byte
	output, err = t.ReadUntil("TSW-750>") // Read until the second prompt delimiter (provided by sending two commands in sendCommand)
	if err != nil {
		return "", err
	}

	t.Close() // Close the telnet session

	output = output[:len(output)-10] // Ghetto trim the prompt off the response

	return string(output), nil
}

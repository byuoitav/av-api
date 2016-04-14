package helpers

import "fmt"

// CheckAvailability checks room availability by consulting with the EMS API and trying to ping the room via telnet
func CheckAvailability(building string, room string) bool {
	telnet := CheckTelnetAvailability()
	scheduling := CheckEMSAvailability(building, room)

	if telnet && scheduling {
		return true
	}

	return false
}

// CheckTelnetAvailability pings the room via telnet to see if the room is currently in use
func CheckTelnetAvailability() bool {
	return true // Temporary for debugging and placeholding
}

// CheckEMSAvailability consults the EMS API to see if the room in question is scheduled to be in use currently
func CheckEMSAvailability(building string, room string) bool {
	availability, err := GetRoomID(building, room)
	CheckErr(err)

	fmt.Printf("%v", availability)

	return true
}

package init

import "github.com/byuoitav/configuration-database-microservice/accessors"

//DMPSInitializer implements the Initializer interface
type DMPSInitializer struct {
}

/*
Initialize fulfils the interface requirements.

We need to get all the DMPS in our room,
*/
func (i *DMPSInitializer) Initialize(room accessors.Room) error {

	return nil
}

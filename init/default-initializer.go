package init

import (
	"fmt"

	"github.com/byuoitav/configuration-database-microservice/accessors"
)

//DefaultInitializer implements the Initializer interface
type DefaultInitializer struct {
}

//Initialize fulfills the initializers for the Initializer interface
func (i *DefaultInitializer) Initialize(room accessors.Room) error {
	fmt.Printf("Yay! I work.\n")
	return nil
}

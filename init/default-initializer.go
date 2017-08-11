package init

import (
	"fmt"

	"github.com/byuoitav/configuration-database-microservice/structs"
)

//DefaultInitializer implements the Initializer interface
type DefaultInitializer struct {
}

//Initialize fulfills the initializers for the Initializer interface
func (i *DefaultInitializer) Initialize(room structs.Room) error {
	fmt.Printf("Yay! I work.\n")
	return nil
}

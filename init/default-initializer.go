package init

import (
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
)

//DefaultInitializer implements the Initializer interface
type DefaultInitializer struct {
}

//Initialize fulfills the initializers for the Initializer interface
func (i *DefaultInitializer) Initialize(room structs.Room) error {
	log.L.Info("[init] Yay! I work.\n")
	return nil
}

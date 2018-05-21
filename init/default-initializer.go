package init

import (
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/structs"
)

//DefaultInitializer implements the Initializer interface
type DefaultInitializer struct {
}

//Initialize fulfills the initializers for the Initializer interface
func (i *DefaultInitializer) Initialize(room structs.Room) error {
	base.Log("Yay! I work.\n")
	return nil
}

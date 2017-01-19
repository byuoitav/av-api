package init

import "fmt"

//DefaultInitializer implements the Initializer interface
type DefaultInitializer struct {
}

//Initialize fulfills the initializers for the Initializer interface
func (i *DefaultInitializer) Initialize() error {
	fmt.Printf("Yay! I work.\n")
	return nil
}

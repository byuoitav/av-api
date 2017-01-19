package init

//DefaultInitializer implements the Initializer interface
type DefaultInitializer struct {}

//Initialize fulfills the initializers for the Initializer interface
(i *defaultInitializer) Initialize() {
  fmt.Printf("Yay! I work.")
}

package helpers

func CheckErr(err error) {
	if err != nil {
		panic(err) // Don't forget your towel
	}
}

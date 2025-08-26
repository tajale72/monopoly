package types

// Define a struct
type Players struct {
	Name       string
	Balance    int
	Position   int
	Properties []Property
}

type Property struct {
	PropertyName string
	Price        int
	Rent         int
	Owner        string
}

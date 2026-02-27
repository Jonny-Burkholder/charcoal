package tests

type simple struct {
	Name string
	Age  int
}

type nested struct {
	ID     string
	Simple simple
}

type pointer struct {
	ID     string
	Simple *simple
}

// TODO: get this one working
// type multipleIndirects struct {
// 	ID     string
// 	Simple **simple
// }

type unexported struct {
	name string
	age  int
}

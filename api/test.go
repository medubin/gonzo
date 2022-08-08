package api

type UserID string

type User struct {
  Id UserID
  Name string
}

type TestBody struct {
  Message string
  Count int
  Many []string
  User User
  Users []User
}

type Server interface {
	Test(TestBody) User
}
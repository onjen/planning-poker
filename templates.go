package main

type templateData struct {
	IsModerator     bool
	IsAuthenticated bool
	Users           []*User
	Poll            string
}

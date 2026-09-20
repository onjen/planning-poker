package main

type templateData struct {
	IsModerator     bool
	IsAuthenticated bool
	Users           []User
	Poll            string
	PointValues     []int
	Vote            int
	CurrentUserID   int
	Revealed        bool
	CanVote         bool
	VotedCount      int
}

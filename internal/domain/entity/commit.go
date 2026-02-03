package entity

import "time"

type Commit struct {
	SHA         string
	Message     string
	Author      string
	AuthorEmail string
	Date        time.Time
	HTMLURL     string
	Additions   int
	Deletions   int
	Repository  string
	Branch      string
}

type CommitFilter struct {
	Repository string
	Branch     string
	Since      *time.Time
	Limit      int
}

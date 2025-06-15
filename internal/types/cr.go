package types

import "time"

type Config struct {
	Registry  string
	RepoName  string
	Username  string
	Password  string
	MaxImages int
}

type TagInfo struct {
	Tag     string
	Created time.Time
}

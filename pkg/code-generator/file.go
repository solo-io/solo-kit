package code_generator

import "os"

type File struct {
	Filename   string
	Content    string
	Permission os.FileMode
}

type Files []File

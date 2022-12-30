package storage

import (
	"path"
	"strconv"
)

type PutOptions struct {
	Bucket      string
	Category    string
	CategoryID  int
	Filename    string
	ContentType string
	Size        int64
}

func (o *PutOptions) BuildPath() string {
	return path.Join(o.Category, strconv.Itoa(o.CategoryID), o.Filename)
}

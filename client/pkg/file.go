package pkg

import (
	"cc.tim/client/model"
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

const (
	LOCAL = iota
)

var FILETYPE int = 0

func init() {
	sendtype := os.Getenv("FILE_WHERE")
	switch sendtype {
	case "LOCAL":
		FILETYPE = LOCAL
	}
}

type FileC interface {
	Save(file *multipart.FileHeader) error
	Load() string
	Delete() error
}

type LoadType struct {
	filename string
}

func NewFileType(filename string) FileC {
	filepath := fmt.Sprintf(model.Filepath, filename)
	switch FILETYPE {
	case LOCAL:
		return &LoadType{filename: filepath}
	}
	return nil
}

func (f *LoadType) Save(file *multipart.FileHeader) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(f.filename)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, src); err != nil {
		return err
	}

	return nil
}

func (f *LoadType) Load() string {
	return f.filename
}
func (f *LoadType) Delete() error {
	if _, err := os.Stat(f.filename); os.IsNotExist(err) {
		return err
	}

	if err := os.Remove(f.filename); err != nil {
		return err
	}

	return nil
}

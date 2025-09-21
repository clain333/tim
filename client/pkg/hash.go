package pkg

import (
	model2 "cc.tim/client/model"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

func HashFile(filename string) (sha256Str string, err error) {
	filepath := fmt.Sprintf(model2.Filepath, filename)
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer file.Close()

	sha256Hash := sha256.New()

	// 将文件内容写入哈希函数
	buf := make([]byte, 4096)
	for {
		n, errRead := file.Read(buf)
		if n > 0 {
			sha256Hash.Write(buf[:n])
		}
		if errRead != nil {
			if errRead == io.EOF {
				break
			}
			err = errRead
			return
		}
	}

	sha256Str = fmt.Sprintf("%x", sha256Hash.Sum(nil))
	return
}

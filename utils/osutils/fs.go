package osutils

import (
	"bufio"
	"io/ioutil"
	"os"

	"github.com/dingkegithub/com.dk.user/utils/osutils"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func IsFile(path string) bool {
	return !IsDir(path)
}

func Mkdir(p string, reverse bool) error {
	if reverse {
		return os.MkdirAll(p, 0766)
	} else {
		return os.Mkdir(p, 0766)
	}
}

func Touch(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()
	return nil
}

func Read(file string) ([]byte, error) {
	if !(osutils.Exists(file) && osutils.IsFile(file)) {
		return nil, os.ErrNotExist
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	return ioutil.ReadAll(f)
}

func Write(file string, c []byte) error {
	wfd, err := os.Create(file)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(wfd)
	defer w.Flush()
	w.Write(c)
	return nil
}

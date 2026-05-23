package validator

import "os"

func closeFile(file *os.File) {
	if err := file.Close(); err != nil {
		return
	}
}

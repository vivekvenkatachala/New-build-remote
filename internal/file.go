package internal

import "fmt"

func ValidateFileExtension(fileExtension string) (string, error) {
	if fileExtension != "zip" && fileExtension != "tar.gz" && fileExtension != "file" && fileExtension != "tar" {
		return "", fmt.Errorf("file extension not supported")
	}

	return "", nil
}
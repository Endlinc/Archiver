package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ReadFileContent(filePath string) (content []string, err error) {
	content = make([]string, 0, 10)
	fi, fErr := os.Open(filePath)
	if fErr != nil {
		fmt.Printf("%s open with a failrue, %s", filePath, fErr.Error())
		return content, fmt.Errorf("bad path: %s", filePath)
	}
	defer fi.Close()

	scanner := bufio.NewScanner(fi)
	for scanner.Scan() {
		content = append(content, strings.TrimSpace(scanner.Text()))
	}

	return content, scanner.Err()
}

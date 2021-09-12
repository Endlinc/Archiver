package function

import (
	"archive/tar"
	"fmt"
	"os"
)

func Compress(dest string, paths []string) error {
	fileWriter, fErr := os.Create(dest)
	if fErr != nil {
		fmt.Println(fErr.Error())
		return fmt.Errorf("bad path for output archive: %s", dest)
	}
	defer func() {
		if err := fileWriter.Close(); err != nil {
			fmt.Println(err.Error())
		}
	}()

	tarWriter := tar.NewWriter(fileWriter)
	defer tarWriter.Close()

	for _, path := range paths {
		fi, pErr := os.Lstat(path)
		if pErr != nil {
			fmt.Println(pErr.Error())
			return fmt.Errorf("%s is not valid file paht", path)
		}

		realPath, err := os.Readlink(path)
		if err != nil {
			fmt.Println(err.Error())
		}
		hdr, err := tar.FileInfoHeader(fi, realPath)
		err = tarWriter.WriteHeader(hdr)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	return nil
}

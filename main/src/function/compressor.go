package function

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type PathInfo struct {
	path  string
	isRec bool
}

type SortablePath []PathInfo

func (a SortablePath) Len() int           { return len(a) }
func (a SortablePath) Less(i, j int) bool { return a[i].path < a[j].path }
func (a SortablePath) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type Compressor struct {
	incl    []PathInfo
	excl    []PathInfo
	relRoot string
	fw      *os.File
	gw      *gzip.Writer
	tw      *tar.Writer
	hasher  hash.Hash
	n       int
}

func (c *Compressor) Init(dest string) error {
	c.relRoot = string(filepath.Separator)
	var err error
	c.fw, err = os.Create(dest)
	if err != nil {
		return err
	}
	c.hasher = sha256.New()
	mulWrt := io.MultiWriter(c.fw, c.hasher)
	c.gw = gzip.NewWriter(mulWrt)
	c.tw = tar.NewWriter(c.gw)
	return nil
}

func (c *Compressor) Close() error {
	fErr := c.fw.Close()
	if fErr != nil {
		fmt.Println(fErr.Error())
	}
	gErr := c.gw.Close()
	if gErr != nil {
		fmt.Println(gErr.Error())
	}
	tErr := c.tw.Close()
	if tErr != nil {
		fmt.Println(tErr.Error())
	}
	if fErr != nil || gErr != nil || tErr != nil {
		return fmt.Errorf("failed during closing writer")
	}
	return nil
}

func (c *Compressor) FormSha256sum() string {
	return hex.EncodeToString(c.hasher.Sum(nil))
}

func (c *Compressor) SetRelRoot(path string) {
	c.relRoot = path
}

func (c *Compressor) LoadPaths(paths []string, isIncl bool) {
	uniqPaths := unique(paths)
	sort.Strings(uniqPaths)
	uniqPaths = removeChildPath(uniqPaths)
	if isIncl {
		for _, path := range uniqPaths {
			c.incl = append(c.incl, PathInfo{path: path, isRec: true})
		}
	} else {
		for _, path := range uniqPaths {
			c.excl = append(c.excl, PathInfo{path: path, isRec: true})
		}
	}
}

func (c *Compressor) AddAllPredecessors() {
	predecessors := make([]string, 0, 10)
	for _, pathInfo := range c.incl {
		tp := pathInfo.path
		prd := filepath.Dir(tp)
		for !strings.EqualFold(prd, string(filepath.Separator)) {
			predecessors = append(predecessors, prd)
			prd = filepath.Dir(prd)
		}
	}
	predecessors = unique(predecessors)
	for _, predecessor := range predecessors {
		c.incl = append(c.incl, PathInfo{path: predecessor, isRec: false})
	}
	sort.Sort(SortablePath(c.incl))
}

func (c *Compressor) Archive() error {
	for _, inclPth := range c.incl {
		if err := c.addContentToTar(inclPth); err != nil {
			fmt.Printf("error occured during archive: %s", err.Error())
			return err
		}
	}
	return nil
}

func (c *Compressor) addContentToTar(incl PathInfo) error {
	if !incl.isRec {
		info, err := os.Lstat(incl.path)
		if err != nil {
			return err
		}
		return c.writePathInfo(incl.path, info)
	} else {
		return filepath.Walk(incl.path, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				fmt.Println(err.Error())
				return nil
			}
			if err = c.writePathInfo(path, info); err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			return c.writeRegularFileContent(path)
		})
	}
}

func (c *Compressor) writePathInfo(path string, info fs.FileInfo) error {
	if c.isOmittingPath(path, info) {
		return nil
	}
	hdr, err := c.makeHeader(path, info)
	if err != nil {
		fmt.Println(err.Error())
		return fmt.Errorf("cannot make header for %s", path)
	}
	if err := c.writeHeader(hdr); err != nil {
		fmt.Println(err.Error())
		return fmt.Errorf("cannot write header for %s", path)
	}
	return nil
}

func (c *Compressor) writeRegularFileContent(path string) error {
	fr, err := os.Open(path)
	if err != nil {
		fmt.Printf("cannot open file: %s", err.Error())
		return fmt.Errorf("cannot open file %s", path)
	}
	_, err = io.Copy(c.tw, fr)
	return err
}

func (c *Compressor) isOmittingPath(path string, info fs.FileInfo) bool {
	if !(info.IsDir() || info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0) {
		return true
	}
	return c.isExcluded(path)
}

func (c *Compressor) isExcluded(path string) bool {
	for c.n < len(c.excl) {
		if strings.Compare(path, c.excl[c.n].path) < 0 {
			return false
		} else if strings.HasPrefix(path, c.excl[c.n].path) {
			if isParDir(c.excl[c.n].path, path) {
				return true
			} else {
				c.n++
				continue
			}
		} else {
			c.n++
		}
	}
	return false
}

func (c *Compressor) makeHeader(path string, info fs.FileInfo) (hdr *tar.Header, err error) {
	pointee := ""
	if info.Mode()&os.ModeSymlink != 0 {
		pointee, err = os.Readlink(path)
		if err != nil {
			return nil, err
		}
	}
	return tar.FileInfoHeader(info, pointee)
}

func (c *Compressor) writeHeader(hdr *tar.Header) error {
	return c.tw.WriteHeader(hdr)
}

func unique(strSlice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func removeChildPath(paths []string) (parPaths []string) {
	parPaths = make([]string, 0, 10)
	parPaths = append(parPaths, paths...)
	lastIdx := 0
	for i := 1; i < len(parPaths); {
		if strings.HasPrefix(parPaths[i], filepath.Join(parPaths[lastIdx], string(filepath.Separator))) {
			parPaths = append(parPaths[:i], parPaths[i+1:]...)
		} else {
			lastIdx = i
			i++
		}
	}
	return parPaths
}

func isParDir(base, target string) bool {
	relPath, err := filepath.Rel(base, target)
	if err != nil {
		fmt.Printf("cannot determine parent path: %s", err.Error())
		return false
	}
	return strings.HasPrefix(relPath, "..")
}

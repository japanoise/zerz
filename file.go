package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	mmap "github.com/edsrzf/mmap-go"
	termutil "github.com/japanoise/termbox-util"
	homedir "github.com/mitchellh/go-homedir"
)

type ZerzFile struct {
	File          *os.File
	Filename      string
	FilenameWidth int
	Filepath      string
	Bytes         mmap.MMap
	Size          int64
}

func OpenFile(filename string) (*ZerzFile, error) {
	absname, err := AbsPath(filename)
	if err != nil {
		return nil, fmt.Errorf("Can't get abspath: %s", err.Error())
	}

	file, err := os.OpenFile(absname, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("Can't open file: %s", err.Error())
	}

	fs, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("Can't stat file: %s", err.Error())
	}

	mm, err := mmap.Map(file, mmap.RDWR, 0)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("Can't mmap file: %s", err.Error())
	}

	ret := ZerzFile{Filename: filepath.Base(absname), Filepath: absname,
		Size: fs.Size(), Bytes: mm, File: file}
	ret.FilenameWidth = termutil.RunewidthStr(ret.Filename)
	return &ret, nil
}

func (zfile *ZerzFile) Close() {
	zfile.Bytes.Unmap()
	zfile.File.Close()
}

func AbsPath(filename string) (string, error) {
	hdpath, perr := homedir.Expand(filename)
	if perr != nil {
		return filename, perr
	}
	if len(hdpath) > 0 && hdpath[0] == '/' {
		return hdpath, nil
	}
	cwd, cerr := os.Getwd()
	if cerr != nil {
		return filename, cerr
	}
	return path.Join(cwd, filename), nil
}

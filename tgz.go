package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func tgzCreate(fsys fs.FS, name string, dataFilter dateFilter) error {
	out, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("failed to open %s, error: %v", name, err)
	}
	defer out.Close()

	err = tgz(fsys, out, dataFilter)
	if err != nil {
		return fmt.Errorf("failed to create %s, error: %v", name, err)
	}

	return nil
}

type dateFilter interface {
	Support(name string) bool
	Filter(name string, data []byte) ([]byte, error)
}

func tgz(fsys fs.FS, buf io.Writer, dataFilter dateFilter) error {
	// tar > gzip > buf
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	// is file a folder?
	fi, err := fs.Stat(fsys, ".")
	if err != nil {
		return err
	}
	mode := fi.Mode()
	if mode.IsRegular() {
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		data, err := fsys.Open(fi.Name())
		if err != nil {
			return err
		}
		if _, err := io.Copy(tw, data); err != nil {
			return err
		}
	} else if mode.IsDir() {
		err = fs.WalkDir(fsys, ".", func(file string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			info, err := d.Info()
			if err != nil {
				return err
			}

			header, err := tar.FileInfoHeader(info, file)
			if err != nil {
				return err
			}

			// must provide real name (see https://golang.org/src/archive/tar/common.go?#L626)
			header.Name = filepath.ToSlash(file)

			if !d.IsDir() {
				if _, err = copyFile(fsys, file, dataFilter, tw, header); err != nil {
					return err
				}
			} else {
				if err := tw.WriteHeader(header); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("error: file type not supported")
	}

	if err := tw.Close(); err != nil {
		return err
	}
	if err := zr.Close(); err != nil {
		return err
	}
	return nil
}

func copyFile(fsys fs.FS, file string, dataFilter dateFilter, tw *tar.Writer, header *tar.Header) (n int64, err error) {
	var dataFile fs.File
	dataFile, err = fsys.Open(file)
	if err != nil {
		return 0, err
	}
	defer dataFile.Close()

	if dataFilter != nil && dataFilter.Support(file) {
		var b bytes.Buffer
		if _, err = io.Copy(&b, dataFile); err != nil {
			return 0, err
		}
		filteredData, err := dataFilter.Filter(file, b.Bytes())
		if err != nil {
			return 0, err
		}

		header.Size = int64(len(filteredData))
		if err := tw.WriteHeader(header); err != nil {
			return 0, err
		}

		var ni int
		ni, err = tw.Write(filteredData)
		return int64(ni), err
	}

	if err := tw.WriteHeader(header); err != nil {
		return 0, err
	}
	return io.Copy(tw, dataFile)
}

// check for path traversal and correct forward slashes
func validRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
}

func unTgz(src io.Reader, dst string) error {
	zr, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	tr := tar.NewReader(zr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}
		target := header.Name

		// validate name against path traversal
		if !validRelPath(header.Name) {
			return fmt.Errorf("tar contained invalid name error %q", target)
		}

		// add dst + re-format slashes according to system
		target = filepath.Join(dst, header.Name)
		// if no join is needed, replace with ToSlash:
		// target = filepath.ToSlash(header.Name)

		switch header.Typeflag {
		case tar.TypeDir: // if its a dir and it doesn't exist create it (with 0755 permission)
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0o755); err != nil {
					return err
				}
			}

		case tar.TypeReg: // if it's a file create it (with same permission)
			fileToWrite, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(fileToWrite, tr); err != nil {
				return err
			}
			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			fileToWrite.Close()
		}
	}

	return nil
}

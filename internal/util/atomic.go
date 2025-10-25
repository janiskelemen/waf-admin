package util

import (
	"bufio"
	"io/fs"
	"os"
)

func AtomicWrite(path string, data []byte, mode fs.FileMode) error {
	tmp := path+".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode); if err != nil { return err }
	bw := bufio.NewWriter(f); if _, err := bw.Write(data); err != nil { _=f.Close(); _=os.Remove(tmp); return err }
	if err := bw.Flush(); err != nil { _=f.Close(); _=os.Remove(tmp); return err }
	if err := f.Close(); err != nil { _=os.Remove(tmp); return err }
	return os.Rename(tmp, path)
}

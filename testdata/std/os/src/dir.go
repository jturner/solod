package main

import (
	"solod.dev/so/fmt"
	"solod.dev/so/os"
)

func dirTest() {
	{
		// ReadDir on a directory with known contents.
		dirName := "test_readdir"
		os.Mkdir(dirName, 0o755)

		os.WriteFile(dirName+"/aaa.txt", []byte("hello"), 0o666)

		os.WriteFile(dirName+"/bbb.txt", []byte("world"), 0o666)

		os.Mkdir(dirName+"/subdir", 0o755)

		entries, err := os.ReadDir(nil, dirName)
		if err != nil {
			panic("ReadDir failed")
		}

		if len(entries) != 3 {
			fmt.Printf("ReadDir: expected 3 entries, got %d\n", len(entries))
			panic("ReadDir: wrong count")
		}

		entry := entries[0]
		if entry.Name != "aaa.txt" || entry.IsDir {
			panic("ReadDir: want 1st = aaa.txt")
		}
		entry = entries[1]
		if entry.Name != "bbb.txt" || entry.IsDir {
			panic("ReadDir: want 2nd = bbb.txt")
		}
		entry = entries[2]
		if entry.Name != "subdir" || !entry.IsDir {
			panic("ReadDir: want 3rd = subdir")
		}
		if entry.Type&os.ModeDir == 0 {
			panic("ReadDir: subdir should have ModeDir")
		}

		os.FreeDirEntry(nil, entries)
		os.Remove(dirName + "/subdir")
		os.Remove(dirName + "/bbb.txt")
		os.Remove(dirName + "/aaa.txt")
		os.Remove(dirName)
	}
	{
		// ReadDir on nonexistent directory.
		_, err := os.ReadDir(nil, "nonexistent_dir_xyz")
		if err != os.ErrNotExist {
			panic("ReadDir nonexistent: wrong error")
		}
	}
}

package main

import (
	"solod.dev/so/mem"
	"solod.dev/so/os"
)

func basicTest() {
	{
		// WriteFile, ReadFile.
		name := "test_rw.txt"
		data := []byte("hello world")
		err := os.WriteFile(name, data, 0o666)
		if err != nil {
			panic("WriteFile failed")
		}
		b, err := os.ReadFile(nil, name)
		if err != nil {
			panic("ReadFile failed")
		}
		if string(b) != string(data) {
			panic("ReadFile: wrong data")
		}
		mem.FreeSlice(nil, b)
		os.Remove(name)
	}
	{
		// Create, Write, Close.
		name := "test_file.txt"
		f, err := os.Create(name)
		if err != nil {
			panic("Create failed")
		}
		// Write.
		n, err := f.Write([]byte("abcdef"))
		if err != nil {
			panic("Write failed")
		}
		if n != 6 {
			panic("Write: wrong count")
		}

		// Close.
		err = f.Close()
		if err != nil {
			panic("Close failed")
		}
		os.Remove(name)
	}
	{
		// Open, Read, Close.
		name := "test_file.txt"
		data := []byte("abcdef")
		err := os.WriteFile(name, data, 0o666)
		if err != nil {
			panic("WriteFile failed")
		}
		// Open.
		f, err := os.Open(name)
		if err != nil {
			panic("Open failed")
		}

		// Read.
		buf := make([]byte, 10)
		n, err := f.Read(buf)
		if err != nil {
			panic("Read failed")
		}
		if n != 6 {
			panic("Read: wrong count")
		}
		if string(buf[:n]) != "abcdef" {
			panic("Read: wrong data")
		}

		// Close.
		err = f.Close()
		if err != nil {
			panic("Close failed")
		}
		os.Remove(name)
	}
	{
		// WriteString.
		name := "test_writestr.txt"
		f, err := os.Create(name)
		if err != nil {
			panic("Create failed")
		}
		n, err := f.WriteString("hello")
		if err != nil {
			panic("WriteString failed")
		}
		if n != 5 {
			panic("WriteString: wrong count")
		}
		f.Close()

		b, err := os.ReadFile(nil, name)
		if err != nil {
			panic("ReadFile failed")
		}
		if string(b) != "hello" {
			panic("WriteString: wrong data")
		}
		mem.FreeSlice(nil, b)
		os.Remove(name)
	}
	{
		// Stdout, Stderr.
		n, err := os.Stdout.WriteString("hello")
		if err != nil {
			panic("Stdout failed")
		}
		if n != 5 {
			panic("Stdout: wrong count")
		}
		n, err = os.Stderr.WriteString("goodbye")
		if err != nil {
			panic("Stderr failed")
		}
		if n != 7 {
			panic("Stderr: wrong count")
		}
		println()
	}
}

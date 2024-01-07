package main

// Author: MarceloUrbanoLima@gmail.com

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Status define the possible status of an Entry and give a hint of what action
// should be taken on and Entry.
// It's zero value is not used
type Status int

const (
	StatusError  Status = 1
	StatusAbort  Status = 2
	StatusRename Status = 3
	StatusDelete Status = 4
	StatusIgnore Status = 5
)

type Entry struct {
	// The ID on a new list of files starts from 1 and is equivalent to their
	// line number as the input data is not supposed to have comments, blank
	// lines, etc, it's just a list of paths.
	// The list after edition though may come with invalid lines, deleted lines, etc
	// then that will link the edited list with the input list is their ID.
	// And ID absent in the edited list mean that the file should be deleted or ignored
	// (depending on this program logic)
	// Lines that cannot be parsed won't have an ID, they will be ignored, so
	// the ID's zero-value is unused
	ID      int
	Path    string
	NewPath string
	Error   error
	Status  Status
}

func (e Entry) String() string {
	return fmt.Sprintf("%d '%s' -> '%s'", e.ID, e.Path, e.NewPath)
}

type Entries map[int]Entry

func main() {
	var err error
	dir := "."
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(fmt.Sprintf("error reading directory %s: %s\n", dir, err))
	}
	entries := make(Entries, len(files))
	for pos, file := range files {
		line := pos + 1
		e := Entry{ID: line, Path: file.Name()}
		entries[line] = e
	}

	//	tmpFile, err := os.CreateTemp("", "editdir.tmp")
	if err != nil {
		panic(fmt.Sprintf("error creating temporary file: %s\n", err))
	}

	output, err := BuildEditList(entries)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		entry := ParseLine(scanner.Text())
		switch entry.Status {
		case StatusAbort:
			fmt.Println("Aborting execution without modifying any files")
			return
		case StatusIgnore:
			continue
		case StatusRename:
			e := entries[entry.ID]
			e.NewPath = entry.Path
			entries[entry.ID] = e
			fmt.Println(e)
		default:
			panic(fmt.Sprintf("Invalid status: %d\n", entry.Status))
		}
	}

	// Load CLI options
	// Receive the path where dirs reside and store in a temporary file
	// add line numbers to each path in the temporary file
	// add to the temp file a line showing how this works
	// Run the editor on that temporary file and wait for it to finish
	// when the editor exists check its return code, if rc!= exit
	// read all lines of the temp file, if any of them have the word cancel or abort then exit
	// check if there are no lines with the same number, exit with error
	// check if there are are new files with the same name (duplicates). Warn

	// options to:
	// delete file if line is deleted from file
	// enable check for duplicate names in output
	// enable check for when the file already exists
	// include dot files (call it 'hidden' if I can work with hidden files in windows)
	// load files recursively, this implies: needs to know what to do with directories, what to do with directory entries, etc
}

func BuildEditList(entries Entries) (string, error) {
	var output strings.Builder
	keys := sort.IntSlice{}
	for k, _ := range entries {
		keys = append(keys, k)
	}
	keys.Sort()
	for _, key := range keys {
		e := entries[key]
		_, err := output.WriteString(fmt.Sprintf("%d %s\n", e.ID, e.Path))
		if err != nil {
			return "", fmt.Errorf("error building edit list: %v", err)
		}
	}
	return output.String(), nil
}

func ParseLine(line string) Entry {
	var entry Entry
	// TODO: Trim other kinds of unicode space
	trimmed := strings.Trim(line, " \r\n")
	if trimmed == "" {
		entry.Status = StatusIgnore
		return entry
	}

	if trimmed[0] == '#' {
		entry.Status = StatusIgnore
		return entry
	}

	if strings.HasPrefix(strings.ToLower(trimmed), "cancel") ||
		strings.HasPrefix(strings.ToLower(trimmed), "abort") {
		entry.Status = StatusAbort
		return entry
	}
	parts := strings.SplitN(line, " ", 2)
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		entry.Error = fmt.Errorf("error reading line number: %v (%v)", err, line)
		entry.Status = StatusError
		return entry
	}
	entry.ID = id
	// TODO: Trim other kinds of unicode space
	entry.Path = strings.Trim(parts[1], " \r\n")
	entry.Status = StatusRename
	return entry
}

package main

// Author: MarceloUrbanoLima@gmail.com

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Status define the possible status of an Entry and give a hint of what action
// should be taken on and Entry.
// It's zero value is not used
type Status string

const (
	StatusIgnore Status = "Ignore" // The default status
	StatusError  Status = "Error"
	StatusCancel Status = "Cancel"
	StatusRename Status = "Rename"
	StatusDelete Status = "Delete"
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
	return fmt.Sprintf("%d '%s' -> '%s' Status: %v", e.ID, e.Path, e.NewPath, e.Status)
}

type Entries map[int]Entry

func (es Entries) String() string {
	keys := sort.IntSlice{}
	for k := range es {
		keys = append(keys, k)
	}
	keys.Sort()
	sb := strings.Builder{}
	for _, key := range keys {
		e := es[key]
		_, err := sb.WriteString(fmt.Sprintf("%d %s\n", e.ID, e.Path))
		if err != nil {
			panic(fmt.Sprintf("error building Entries' String(): %v", err))
		}
	}
	return sb.String()
}

// Update updates Entries from other entries, updating their NewPath and Status
func (es *Entries) Update(from Entries) {
	for i, _ := range *es {
		e := (*es)[i]
		ee, ok := from[e.ID]
		if !ok {
			e.Status = StatusDelete
			(*es)[i] = e
			continue
		}
		e.NewPath = ee.Path
		e.Status = StatusRename
		(*es)[i] = e
	}
}

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// TODO:
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

func run() error {
	// Things hardcoded only for now
	dir := "." // get from CLI args or a list from standard input
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	addHint := true
	// End of hardcoded things I need to turn into variables

	var err error
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("error reading directory %s: %s", dir, err)
	}

	entries := make(Entries, len(files))
	for pos, file := range files {
		line := pos + 1
		e := Entry{ID: line, Path: file.Name()}
		entries[line] = e
	}

	// Create temporary file to be edited
	editableList := entries.String()
	tmpFile, err := os.CreateTemp("", "editdir.")
	if err != nil {
		return fmt.Errorf("Error creating temporary file for edition: %s", err)
	}
	defer os.Remove(tmpFile.Name())

	if addHint {
		tmpFile.WriteString("# Lines starting with # or empty are ignored.\n")
		tmpFile.WriteString("# Lines starting with 'cancel' or 'abort' tell the program to do nothing on exit. As well as exiting the editor with non-zero return code.\n")
		tmpFile.WriteString("# Valid path lines have the format 'ID<space>NewPath'. They must start with a number (no spaces before) followed by a space and the new path for that file.\n")
		tmpFile.WriteString("# The number is the ID referencing to the original list so you can reorder the lines as long as the ID is kept.\n")
		tmpFile.WriteString("# Deleted lines tell the program that it must delete that file from disk.\n")
	}
	tmpFile.WriteString(editableList)
	log.Debug().Str("tmpFile", tmpFile.Name()).Msg("Temporary file created for edition")

	// Run the editor
	editorBin := "vi"
	editorArgs := []string{tmpFile.Name()}
	cmd := exec.Command(editorBin, editorArgs...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Canceling due to no clean execution and exit of editor: Exit Message=%s (%T), Return Code=%d\n", err, err, cmd.ProcessState.ExitCode())
	}

	// Load edited entries
	tmpContent, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("Error reading file with edited paths: %v", err)
	}

	editedEntries, err := LoadEditedList(string(tmpContent))
	if err != nil {
		// If it was cancelled does not tell it's an error
		for _, e := range editedEntries {
			if e.Status == StatusCancel {
				return e.Error // contains the line in which the cancelation was present
			}
		}
		return fmt.Errorf("Error parsing the edited paths: %v", err)
	}

	// Merge edited paths
	entries.Update(editedEntries)
	for _, e := range entries {
		fmt.Println(e)
	}

	return nil
}

// ParseEditedLine will read a line string and try to extract an Entry from it.
// It expects the line to start with and ID and a Path
// The ID will reference the entry in the input line and will be used to match
// the old name with the new/edited name for the program to rename or delete
// that file
func ParseEditedLine(line string) Entry {
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
		entry.Status = StatusCancel
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

// LoadEditedList receives a string containing multiple lines, each line with
// a possible Entry in text form (before being parsed into and Entry).
// It returns Entries, error and/or indication of cancellation
func LoadEditedList(text string) (Entries, error) {
	editedEntries := Entries{}
	scanner := bufio.NewScanner(strings.NewReader(text))
	lineNum := 0
	for scanner.Scan() {
		entry := ParseEditedLine(scanner.Text())
		lineNum++
		switch entry.Status {
		case StatusCancel:
			// Return a single entry to indicate that the program should be cancelled
			err := fmt.Errorf("Cancelling execution without modifying any files due to cancel/abort clause in line %d", lineNum)
			ent := Entries{}
			ent[0] = Entry{
				ID:     0,
				Error:  err,
				Status: StatusCancel,
			}
			return ent, err
		case StatusIgnore:
			log.Debug().Str("line", scanner.Text()).Msg("Ignored line")
			continue
		case StatusRename:
			editedEntries[entry.ID] = entry
		default:
			return editedEntries, fmt.Errorf("Invalid status: %v. Error=%s\n", entry.Status, entry.Error)
		}
	}
	return editedEntries, nil
}

// entriesFromLines receives a string with a file path in each line and
// return an Entries struct with the line number as the ID and the line text
// as Path of each entry
func entriesFromLines(text string) Entries {
	entries := Entries{}
	scanner := bufio.NewScanner(strings.NewReader(text))
	line := 1
	for scanner.Scan() {
		e := Entry{ID: line,
			Path:   scanner.Text(),
			Status: StatusIgnore}
		entries[e.ID] = e
		line++
	}
	return entries
}

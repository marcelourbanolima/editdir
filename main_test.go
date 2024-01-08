package main

import (
	"testing"
)

func TestParseLineNonRenameStatus(t *testing.T) {
	testCases := []struct {
		Line       string
		WantStatus Status
	}{
		{Line: "", WantStatus: StatusIgnore},
		{Line: "#", WantStatus: StatusIgnore},
		{Line: "###", WantStatus: StatusIgnore},
		{Line: " ##", WantStatus: StatusIgnore},
		{Line: "cancel", WantStatus: StatusAbort},
		{Line: "abort", WantStatus: StatusAbort},
		{Line: "bad line", WantStatus: StatusError},
	}

	for _, tc := range testCases {
		t.Run(tc.Line, func(t *testing.T) {
			e := ParseLine(tc.Line)
			if e.Status != tc.WantStatus {
				t.Fatalf("Expected Status %q, got %q for line: %v", tc.WantStatus, e.Status, tc.Line)
			}
		})
	}
}

func TestParseLineTrimmed(t *testing.T) {
	e := ParseLine("2   file   a.txt   ")
	if e.Status != StatusRename {
		t.Fatalf("Entry should be StatusRename, got: %v ", e.Status)
	}
	if e.Error != nil {
		t.Fatalf("Entry should not have an error: %s ", e.Error)
	}
	if e.ID != 2 {
		t.Fatalf("Entry's line number is not 2, is %d", e.ID)
	}
	if e.Path != "file   a.txt" {
		t.Fatalf("Entry's path is not 'file   a.txt', it is %q", e.Path)
	}
}

func TestParseLineUnixNewLineEnd(t *testing.T) {
	e := ParseLine("2 file   a.txt\n")
	if e.Status != StatusRename {
		t.Fatalf("Entry should be StatusRename, got: %v", e.Status)
	}
	if e.Error != nil {
		t.Fatalf("Entry should not have an error: %s ", e.Error)
	}
	if e.ID != 2 {
		t.Fatalf("Entry's line number is not 2, is %d", e.ID)
	}
	if e.Path != "file   a.txt" {
		t.Fatalf("Entry's path is not 'file   a.txt', it is %q", e.Path)
	}
}

func TestParseLineWindowsNewLineEnd(t *testing.T) {
	e := ParseLine("2 file   a.txt\r\n")
	if e.Status != StatusRename {
		t.Fatalf("Entry should be StatusRename, got: %v", e.Status)
	}
	if e.Error != nil {
		t.Fatalf("Entry should not have an error: %s ", e.Error)
	}
	if e.ID != 2 {
		t.Fatalf("Entry's line number is not 2, is %d", e.ID)
	}
	if e.Path != "file   a.txt" {
		t.Fatalf("Entry's path is not 'file   a.txt', it is %q", e.Path)
	}
}

func TestEntriesFromLines(t *testing.T) {
	es := entriesFromLines(`file1.txt
file 2.txt
`)
	if len(es) != 2 {
		t.Fatalf("Invalid number of entries from lines: %d!=2", len(es))
	}
	e1, ok := es[1]
	if !ok {
		t.Fatalf("Should have found ID 1")
	}
	if e1.ID != 1 {
		t.Fatalf("Entry 1 should have ID 1")
	}
	if e1.Path != "file1.txt" {
		t.Fatalf("Entry 1 has bad Path: %v", e1.Path)
	}

	e2, ok := es[2]
	if !ok {
		t.Fatalf("Should have found ID 2")
	}
	if e2.ID != 2 {
		t.Fatalf("Entry 2 should have ID 2")
	}
	if e2.Path != "file 2.txt" {
		t.Fatalf("Entry 2 has bad Path: %v", e2.Path)
	}

}

func TestEditedList(t *testing.T) {
	entries := entriesFromLines(`file1.txt
file 2.txt
file 3.txt
file 4.txt
`)
	edited, err := LoadEditedList(`1 file1.txt
2 file TWO.txt
4 file 4.txt
`)
	if err != nil {
		t.Fatalf("Error preparing edited list in this test")
	}

	entries.Update(edited)
	e1, ok := entries[1]
	if !ok {
		t.Fatalf("Should have found ID 1")
	}
	if e1.Path != "file1.txt" {
		t.Fatalf("Entry 1 has bad Path: %v", e1.Path)
	}
	if e1.Status != StatusRename {
		t.Fatalf("Entry 1 should have StatusRename: %v != %v", StatusRename, e1.Status)
	}
	if e1.NewPath == "" {
		t.Fatalf("Entry 1 should have a NewPath")
	}

	e2, ok := entries[2]
	if !ok {
		t.Fatalf("Should have found ID 2")
	}
	if e2.Path != "file 2.txt" {
		t.Fatalf("Entry 2 has bad Path: %v", e2.Path)
	}
	if e2.Status != StatusRename {
		t.Fatalf("Entry 2 should have StatusRename: %v != %v", StatusRename, e2.Status)
	}
	if e2.NewPath == "" {
		t.Fatalf("Entry 2 should have a NewPath")
	}

	e3, ok := entries[3]
	if !ok {
		t.Fatalf("Should have found ID 3")
	}
	if e3.ID != 3 {
		t.Fatalf("Entry 3 should have ID 3, has %d", e3.ID)
	}
	if e3.Status != StatusDelete {
		t.Fatalf("Entry 3 should have StatusDelete: %v != %v", StatusDelete, e3.Status)
	}

	e4, ok := entries[4]
	if !ok {
		t.Fatalf("Should have found ID 4")
	}
	if e4.Path != "file 4.txt" {
		t.Fatalf("Entry 4 has bad Path: %v", e4.Path)
	}
	if e4.Status != StatusRename {
		t.Fatalf("Entry 4 should have StatusRename: %v != %v", StatusRename, e4.Status)
	}
	if e4.NewPath == "" {
		t.Fatalf("Entry 4 should have a NewPath")
	}
}

/* more tests to add
4 file with tricky char ?
5 file with tricky char *
6 file with tricky char : on windows
*/

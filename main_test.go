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
		t.Fatalf("Entry should be StatusRename, got: %d ", e.Status)
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
		t.Fatalf("Entry should be StatusRename, got: %d ", e.Status)
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
		t.Fatalf("Entry should be StatusRename, got: %d ", e.Status)
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

/* more tests to add
4 file with tricky char ?
5 file with tricky char *
6 file with tricky char : on windows
*/

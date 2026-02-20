package next

import (
	"testing"
	"time"
)

func TestTableLog_AddAndLength(t *testing.T) {
	log := NewTableLog(3)
	if log.Length() != 0 {
		t.Fatal("initial length not 0")
	}
	log.Add("src", "INFO", "msg1")
	if log.Length() != 1 {
		t.Fatalf("expected length 1, got %d", log.Length())
	}
	log.Add("src", "INFO", "msg2")
	if log.Length() != 2 {
		t.Fatalf("expected length 2, got %d", log.Length())
	}
}

func TestTableLog_RingBuffer(t *testing.T) {
	log := NewTableLog(2)
	log.Add("s1", "INFO", "first")
	log.Add("s2", "INFO", "second")
	// Row0 should be second (newest)
	if log.Str(0, 3) != "second" {
		t.Fatalf("row0 message: got %s", log.Str(0, 3))
	}
	if log.Str(1, 3) != "first" {
		t.Fatalf("row1 message: got %s", log.Str(1, 3))
	}
	// Add third, overwrites oldest
	log.Add("s3", "INFO", "third")
	// Now rows: row0=third, row1=second
	if log.Str(0, 3) != "third" {
		t.Fatalf("after overwrite row0: got %s", log.Str(0, 3))
	}
	if log.Str(1, 3) != "second" {
		t.Fatalf("after overwrite row1: got %s", log.Str(1, 3))
	}
	if log.Length() != 2 {
		t.Fatalf("expected length 2, got %d", log.Length())
	}
}

func TestTableLog_Columns(t *testing.T) {
	log := NewTableLog(10)
	cols := log.Columns()
	if len(cols) != 4 {
		t.Fatalf("expected 4 columns, got %d", len(cols))
	}
	if cols[0].Header != "Time" {
		t.Error("column 0 header should be Time")
	}
	if cols[1].Header != "Level" {
		t.Error("column 1 header should be Level")
	}
	if cols[2].Header != "Source" {
		t.Error("column 2 header should be Source")
	}
	if cols[3].Header != "Message" {
		t.Error("column 3 header should be Message")
	}
}

func TestTableLog_Iter(t *testing.T) {
	log := NewTableLog(5)
	log.Add("s", "INFO", "a")
	log.Add("s", "INFO", "b")
	log.Add("s", "INFO", "c")
	ch := log.Iter()
	var msgs []string
	for entry := range ch {
		msgs = append(msgs, entry.Message)
	}
	if len(msgs) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(msgs))
	}
	if msgs[0] != "a" || msgs[1] != "b" || msgs[2] != "c" {
		t.Fatalf("expected oldest to newest order, got %v", msgs)
	}
}

func TestTableLog_TimeFormat(t *testing.T) {
	log := NewTableLog(1)
	log.Add("src", "DEBUG", "test")
	timeStr := log.Str(0, 0)
	// Expect time in HH:MM:SS format (time.TimeOnly = "15:04:05")
	_, err := time.Parse(time.TimeOnly, timeStr)
	if err != nil {
		t.Fatalf("time format invalid: %s", err)
	}
}

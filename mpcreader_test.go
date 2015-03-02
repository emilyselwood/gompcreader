package gompcreader

import (
	"testing"
	"time"
)

func TestReadString(t *testing.T) {
	buffer := []byte("fish and chips ")
	var result = readString(buffer)
	if result != "fish and chips" {
		t.Errorf("readString = %s, want \"fish and chips\"", result)
	}
}

func TestReadFloat(t *testing.T) {
	buffer := []byte("12.8567 ")
	var result = readFloat(buffer)
	if result != 12.8567 {
		t.Errorf("readFloat = %d, want %d", result, 12.8567)
	}
}

func TestReadInt(t *testing.T) {
	buffer := []byte("128 ")
	var result = readInt(buffer)
	if result != 128 {
		t.Errorf("readInt = %d, want %d", result, 128)
	}
}

func TestReadPackedInt(t *testing.T) {
	buffer := []byte("a128 ")
	var result = readPackedInt(buffer)
	if result != 36128 {
		t.Errorf("readInt = %d, want %d", result, 36128)
	}
}

func TestReadPackedDate(t *testing.T) {
	buffer := []byte("I23AP")
	var result = readPackedTime(buffer)
	if !time.Date(1823, 10, 25, 0, 0, 0, 0, time.UTC).Equal(result) {
		t.Errorf("readPackedTime = %s", result.Format("2006-01-02T03:04:00"))
	}
}

func TestReadPackedIdentifier(t *testing.T) {
	buffer := []byte("PLS2040")
	var result = readPackedIdentifier(buffer)
	if result != "2040 P-L" {
		t.Errorf("readPackedIdentifier = %s should be 2040 P-L", result)
	}

	buffer = []byte("T1S3138")
	result = readPackedIdentifier(buffer)
	if result != "3138 T-1" {
		t.Errorf("readPackedIdentifier = %s should be 3138 T-1", result)
	}

	buffer = []byte("J95X00A")
	result = readPackedIdentifier(buffer)
	if result != "1995 XA" {
		t.Errorf("readPackedIdentifier = %s should be 1995 XA", result)
	}

	buffer = []byte("A0001")
	result = readPackedIdentifier(buffer)
	if result != "100001" {
		t.Errorf("readPackedIdentifier = %s should be 100001", result)
	}
}

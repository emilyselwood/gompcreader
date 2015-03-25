package gompcreader

import (
	"testing"
	"time"
)

func TestReadString(t *testing.T) {
	buffer := "fish and chips "
	var result = readString(buffer)
	if result != "fish and chips" {
		t.Errorf("readString = %s, want \"fish and chips\"", result)
	}
}

func TestReadFloat(t *testing.T) {
	buffer := "12.8567 "
	var result = readFloat(buffer)
	if result != 12.8567 {
		t.Errorf("readFloat = %f, want %f", result, 12.8567)
	}
}

func TestReadInt(t *testing.T) {
	buffer := "128 "
	var result = readInt(buffer)
	if result != 128 {
		t.Errorf("readInt = %d, want %d", result, 128)
	}
}

func TestReadPackedInt(t *testing.T) {
	buffer := "a128 "
	var result = readPackedInt(buffer)
	if result != 36128 {
		t.Errorf("readInt = %d, want %d", result, 36128)
	}
}

var packedDateTests = []struct {
	in  string
	out time.Time
}{
	{"I23AP", time.Date(1823, 10, 25, 0, 0, 0, 0, time.UTC)},
	{" J2319 ", time.Date(1923, 1, 9, 0, 0, 0, 0, time.UTC)},
}

func TestReadPackedDate(t *testing.T) {
	for _, tt := range packedDateTests {
		var result = readPackedTime(tt.in)
		if !tt.out.Equal(result) {
			t.Errorf(
				"readPackedTime(%s) = %s expected %s",
				tt.in,
				result.Format("2006-01-02T03:04:00 -0700"),
				tt.out.Format("2006-01-02T03:04:00 -0700"))
		}
	}
}

var packedIdentifierTests = []struct {
	in  string
	out string
}{
	{"PLS2040", "2040 P-L"},
	{"T1S3138", "3138 T-1"},
	{"J95X00A", "1995 XA"},
	{"A0001", "100001"},
	{"0000054 ", "54"},
}

func TestReadPackedIdentifier(t *testing.T) {
	for _, tt := range packedIdentifierTests {
		var result = readPackedIdentifier(tt.in)
		if result != tt.out {
			t.Errorf("readPackedIdentifier(%s) == %s expected %s",
				tt.in,
				result,
				tt.out)
		}
	}
}

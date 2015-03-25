package gompcreader

import (
	"testing"
	"time"
)

var readStringTests = []struct {
	in  string
	out string
}{
	{"fish and chips ", "fish and chips"},
	{" fish and chips", "fish and chips"},
	{"   000 000 000    ", "000 000 000"},
	{"       ", ""},
	{"", ""},
}

func TestReadString(t *testing.T) {
	for _, tt := range readStringTests {
		var result = readString(tt.in)
		if result != tt.out {
			t.Errorf("readString(%s) = %s expected %s",
				tt.in,
				result,
				tt.out)
		}
	}
}

var readFloatTests = []struct {
	in  string
	out float64
}{
	{"12.8567 ", 12.8567},
	{" -1.4553 ", -1.4553},
	{"0.000000000", 0},
	{"2.33E4", 23300},
	{"", 0}, // Probably shouldn't parse or should have an error path
}

func TestReadFloat(t *testing.T) {
	for _, tt := range readFloatTests {
		var result = readFloat(tt.in)
		if result != tt.out {
			t.Errorf("readFloat(%s) = %f expected %f",
				tt.in,
				result,
				tt.out)
		}
	}
}

var readIntTests = []struct {
	in  string
	out int64
}{
	{"128 ", 128},
	{" 34533  ", 34533},
	{" -345", -345},
}

func TestReadInt(t *testing.T) {
	for _, tt := range readIntTests {
		var result = readInt(tt.in)
		if result != tt.out {
			t.Errorf("readInt(%s) = %d expected %d",
				tt.in,
				result,
				tt.out)
		}
	}
}

var readPackedIntTests = []struct {
	in  string
	out int64
}{
	{" a128 ", 36128},
	{" 1234  ", 1234},
	{"z123", 61123},
	{"00", 0},
	{"A", 10},
	{"Z", 35},
	{"Z4", 354},
}

func TestReadPackedInt(t *testing.T) {
	for _, tt := range readPackedIntTests {
		var result = readPackedInt(tt.in)
		if result != tt.out {
			t.Errorf("readPackedInt(%s) = %d expected %d",
				tt.in,
				result,
				tt.out)
		}
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
				result.Format("2006-01-02T15:04:00 -0700"),
				tt.out.Format("2006-01-02T15:04:00 -0700"))
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

func TestConvert(t *testing.T) {
	var entry = "00001    3.34  0.12 K13B4  10.55761   72.29213   80.32762   10.59398  0.0757973  0.21415869   2.7668073  0 MPO286777  6502 105 1802-2014 0.82 M-v 30h MPCLINUX   0000      (1) Ceres              20140307"
	var result = convertToMinorPlanet(entry)
	if result.ID != "1" {
		t.Errorf("convertToMinorPlanet ID %s expected 1", result.ID)
	}

	if result.ReadableDesignation != "(1) Ceres" {
		t.Errorf("convertToMinorPlanet ReadableDesignation %s expected 1",
			result.ReadableDesignation)
	}

	var expected = time.Date(2014, time.March, 7, 0, 0, 0, 0, time.UTC)

	if !expected.Equal(result.DateOfLastObservation) {
		t.Errorf("convertToMinorPlanet DateOfLastObservation %s expected %s",
			result.DateOfLastObservation.Format("2006-01-02T15:04:00 -0700"),
			expected.Format("2006-01-02T15:04:00 -0700"))
	}
}

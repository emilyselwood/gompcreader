package gompcreader

import (
	"fmt"
	"testing"
	"time"
)

type stringTestCase struct {
	in  string
	out string
}

type doStringTest func(stringTestCase) (bool, string)

func stringTest(t *testing.T, n string, cases []stringTestCase, test doStringTest) {
	for _, tt := range cases {
		var r, a = test(tt)
		if !r {
			t.Errorf("%s(%s) was %s expected %s", n, tt.in, a, tt.out)
		}
	}
}

type intTestCase struct {
	in  string
	out int64
}

type doIntTest func(intTestCase) (bool, int64, error)

func intTest(t *testing.T, n string, cases []intTestCase, test doIntTest) {
	for _, tt := range cases {
		var r, a, e = test(tt)
		if !r {
			if e != nil {
				t.Errorf("%s(%s) failed with %s. Expected %d", n, tt.in, e, tt.out)
			} else {
				t.Errorf("%s(%s) was %d expected %d", n, tt.in, a, tt.out)
			}
		}
	}
}

type floatTestCase struct {
	in  string
	out float64
}

type doFloatTest func(floatTestCase) (bool, float64, error)

func floatTest(t *testing.T, n string, cases []floatTestCase, test doFloatTest) {
	for _, tt := range cases {
		var r, a, e = test(tt)
		if !r {
			if e != nil {
				t.Errorf("%s(%s) failed with %s. Expected %f", n, tt.in, e, tt.out)
			} else {
				t.Errorf("%s(%s) was %f expected %f", n, tt.in, a, tt.out)
			}
		}
	}
}

var readStringTests = []stringTestCase{
	{"fish and chips ", "fish and chips"},
	{" fish and chips", "fish and chips"},
	{"   000 000 000    ", "000 000 000"},
	{"       ", ""},
	{"", ""},
}

func TestReadString(t *testing.T) {
	stringTest(t,
		"readString",
		readStringTests,
		func(t stringTestCase) (bool, string) {
			var r = readString(t.in)
			return r == t.out, r
		})
}

var readFloatTests = []floatTestCase{
	{"12.8567 ", 12.8567},
	{" -1.4553 ", -1.4553},
	{"0.000000000", 0},
	{"2.33E4", 23300},
}

func TestReadFloat(t *testing.T) {
	floatTest(t,
		"readFloat",
		readFloatTests,
		func(c floatTestCase) (bool, float64, error) {
			var r, e = readFloat(c.in)
			return e == nil && r == c.out, r, e
		})
}

var errorFloatTests = []stringTestCase{
	{"", "strconv.ParseFloat: parsing \"\": invalid syntax"},
	{"jshdjkghgjk", "strconv.ParseFloat: parsing \"jshdjkghgjk\": invalid syntax"},
	{"54767 64", "strconv.ParseFloat: parsing \"54767 64\": invalid syntax"},
}

func TestErrorReadFloat(t *testing.T) {
	stringTest(t,
		"readFloat",
		errorFloatTests,
		func(c stringTestCase) (bool, string) {
			var r, e = readFloat(c.in)
			if e == nil {
				return false, fmt.Sprintf("%f", r)
			}
			return e.Error() == c.out, e.Error()
		})
}

var readIntTests = []intTestCase{
	{"128 ", 128},
	{" 34533  ", 34533},
	{" -345", -345},
}

func TestReadInt(t *testing.T) {
	intTest(t,
		"readInt",
		readIntTests,
		func(t intTestCase) (bool, int64, error) {
			var r, e = readInt(t.in)
			return e == nil && r == t.out, r, e
		})
}

var errorIntTests = []stringTestCase{
	{"", "strconv.ParseInt: parsing \"\": invalid syntax"},
	{"jshdjkghgjk", "strconv.ParseInt: parsing \"jshdjkghgjk\": invalid syntax"},
	{"54767 64", "strconv.ParseInt: parsing \"54767 64\": invalid syntax"},
	{"54767.64", "strconv.ParseInt: parsing \"54767.64\": invalid syntax"},
}

func TestReadIntErrors(t *testing.T) {
	stringTest(t,
		"readInt",
		errorIntTests,
		func(c stringTestCase) (bool, string) {
			var r, e = readInt(c.in)
			if e == nil {
				return false, fmt.Sprintf("%d", r)
			}
			return e.Error() == c.out, e.Error()
		})
}

var readHexIntTests = []intTestCase{
	{"128 ", 296},
	{" 34533  ", 214323},
	{" -345", -837},
	{" A ", 10},
	{" FF", 255},
}

func TestReadHexInt(t *testing.T) {
	intTest(t,
		"readHexInt",
		readHexIntTests,
		func(t intTestCase) (bool, int64, error) {
			var r, e = readHexInt(t.in)
			return e == nil && r == t.out, r, e
		})
}

var errorHexIntTests = []stringTestCase{
	{"", "strconv.ParseInt: parsing \"\": invalid syntax"},
	{"jshdjkghgjk", "strconv.ParseInt: parsing \"jshdjkghgjk\": invalid syntax"},
	{"54767 64", "strconv.ParseInt: parsing \"54767 64\": invalid syntax"},
	{"54767.64", "strconv.ParseInt: parsing \"54767.64\": invalid syntax"},
}

func TestReadHexIntErrors(t *testing.T) {
	stringTest(t,
		"readHexInt",
		errorHexIntTests,
		func(c stringTestCase) (bool, string) {
			var r, e = readHexInt(c.in)
			if e == nil {
				return false, fmt.Sprintf("%d", r)
			}
			return e.Error() == c.out, e.Error()
		})
}

var readPackedIntTests = []intTestCase{
	{" a128 ", 36128},
	{" 1234  ", 1234},
	{"z123", 61123},
	{"00", 0},
	{"A", 10},
	{"Z", 35},
	{"Z4", 354},
}

func TestReadPackedInt(t *testing.T) {
	intTest(t,
		"readPackedInt",
		readPackedIntTests,
		func(t intTestCase) (bool, int64, error) {
			var r = readPackedInt(t.in)
			return r == t.out, r, nil
		})
}

var packedIdentifierTests = []stringTestCase{
	{"PLS2040", "2040 P-L"},
	{"T1S3138", "3138 T-1"},
	{"J95X00A", "1995 XA"},
	{"J95X45A", "1995 XA45"},
	{"A0001", "100001"},
	{"0000054 ", "54"},
}

func TestReadPackedIdentifier(t *testing.T) {
	stringTest(t,
		"readPackedIdentifier",
		packedIdentifierTests,
		func(c stringTestCase) (bool, string) {
			var r = readPackedIdentifier(c.in)
			return r == c.out, r
		})
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

var dateTests = []struct {
	in  string
	out time.Time
}{
	{"19950000", time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC)},
	{" 19951123 ", time.Date(1995, 11, 23, 0, 0, 0, 0, time.UTC)},
}

func TestReadDate(t *testing.T) {
	for _, tt := range dateTests {
		var r, e = readTime(tt.in)
		if e != nil {
			t.Errorf("readTime(%s) errored %s, expected %s",
				tt.in,
				e.Error(),
				tt.out.Format("2006-01-02T15:04:00 -0700"))
		}
		if !tt.out.Equal(r) {
			t.Errorf(
				"readTime(%s) = %s expected %s",
				tt.in,
				r.Format("2006-01-02T15:04:00 -0700"),
				tt.out.Format("2006-01-02T15:04:00 -0700"))
		}
	}
}

func TestConvert(t *testing.T) {
	var entry = "00001    3.34  0.12 K13B4  10.55761   72.29213   80.32762   10.59398  0.0757973  0.21415869   2.7668073  0 MPO286777  6502 105 1802-2014 0.82 M-v 30h MPCLINUX   0000      (1) Ceres              20140307"
	var result, err = convertToMinorPlanet(entry)
	if err != nil {
		t.Fatalf("convertToMinorPlanet returned an error %s", err)
	}
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

var convertErrorsTests = []stringTestCase{
	{"ajghfjhsdfjkhgjfkghjfhgjfhgjsfhgjhfjghfdjkh",
		"strconv.ParseFloat: parsing \"gjsfhgjhf\": invalid syntax"},
	{"00001    3.34  0.12 K13B4  10.55761  sjhagjkfhgjkshfgjl",
		"strconv.ParseFloat: parsing \"sjhagjkfhg\": invalid syntax"},
	{"00001    3.34  0.12 K13B4  dshgsh  sjhagjkfhgjkshfgjl",
		"strconv.ParseFloat: parsing \"dshgsh\": invalid syntax"},
	{"00001    3.34  0.12 K13B4  10.55761   72.29213  jajhkhs  hfdsjkgjkh",
		"strconv.ParseFloat: parsing \"jajhkhs\": invalid syntax"},
}

func TestConverErrors(t *testing.T) {
	stringTest(t,
		"convertToMinorPlanet",
		convertErrorsTests,
		func(c stringTestCase) (bool, string) {
			var _, e = convertToMinorPlanet(c.in)
			return e.Error() == c.out, e.Error()
		})
}

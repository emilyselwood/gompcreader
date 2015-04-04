package gompcreader

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type stringTestCase struct {
	in  string
	out string
}

type doStringTest func(string) string

func stringTest(t *testing.T, n string, cases []stringTestCase, test doStringTest) {
	for _, tt := range cases {
		a := test(tt.in)
		assert.Equal(t, tt.out, a, "%s(%s) was %s expected %s", n, tt.in, a, tt.out)
	}
}

type intTestCase struct {
	in  string
	out int64
}

type doIntTest func(string) (int64, error)

func intTest(t *testing.T, n string, cases []intTestCase, test doIntTest) {
	for _, tt := range cases {
		a, e := test(tt.in)
		assert.Nil(t, e)
		assert.Equal(t, tt.out, a, "int not as expected")
	}
}

type floatTestCase struct {
	in  string
	out float64
}

type doFloatTest func(string) (float64, error)

func floatTest(t *testing.T, n string, cases []floatTestCase, test doFloatTest) {
	for _, tt := range cases {
		var a, e = test(tt.in)
		assert.Nil(t, e)
		assert.Equal(t, tt.out, a, "%s(%s) was %f expected %f", n, tt.in, a, tt.out)
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
		readString)
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
		readFloat)
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
		func(c string) string {
			var r, e = readFloat(c)
			if e == nil {
				return fmt.Sprintf("%f", r)
			}
			return e.Error()
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
		readInt)
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
		func(c string) string {
			var r, e = readInt(c)
			if e == nil {
				return fmt.Sprintf("%d", r)
			}
			return e.Error()
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
		readHexInt)
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
		func(c string) string {
			var r, e = readHexInt(c)
			if e == nil {
				return fmt.Sprintf("%d", r)
			}
			return e.Error()
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
		func(t string) (int64, error) {
			var r = readPackedInt(t)
			return r, nil
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
		readPackedIdentifier)
}

var arcLengthTests = []intTestCase{
	{"  4 days", 4},
	{"34 days ", 34},
}

func TestArcLength(t *testing.T) {
	intTest(t,
		"readArcLength",
		arcLengthTests,
		readArcLength)
}

var arcLengthErrorTests = []stringTestCase{
	{" dsfhsdj dsfjhdhsj", "strconv.ParseInt: parsing \"dsfhsdj\": invalid syntax"},
	{"sghkjsdf", "Arc length didn't have enough parts"},
	{"", "Arc length didn't have enough parts"},
}

func TestArcLengthErrors(t *testing.T) {
	stringTest(t,
		"readArcLength",
		arcLengthErrorTests,
		func(c string) string {
			_, e := readArcLength(c)
			return e.Error()
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

func TestConvertCeres(t *testing.T) {
	var entry = "00001    3.34  0.12 K13B4  10.55761   72.29213   80.32762   10.59398  0.0757973  0.21415869   2.7668073  0 MPO286777  6502 105 1802-2014 0.82 M-v 30h MPCLINUX   0000      (1) Ceres              20140307"
	var result, err = convertToMinorPlanet(entry)
	assert.Nil(t, err, "convertToMinorPlanet returned an error %s", err)

	assert.Equal(t,
		"1",
		result.ID,
		"convertToMinorPlanet ID %s expected 1",
		result.ID)

	assert.Equal(t,
		"(1) Ceres",
		result.ReadableDesignation,
		"convertToMinorPlanet ReadableDesignation %s expected (1) Ceres",
		result.ReadableDesignation)

	expected := time.Date(2014, time.March, 7, 0, 0, 0, 0, time.UTC)
	assert.Equal(t,
		expected,
		result.DateOfLastObservation,
		"convertToMinorPlanet DateOfLastObservation %s expected %s",
		result.DateOfLastObservation.Format("2006-01-02T15:04:00 -0700"),
		expected.Format("2006-01-02T15:04:00 -0700"))

	assert.Equal(t,
		0,
		result.ArcLength,
		"convertToMinorPlanet ReadableDesignation %s expected 0",
		result.ArcLength)

	assert.Equal(t,
		1802,
		result.YearOfFirstObservation,
		"convertToMinorPlanet YearOfFirstObservation %s expected 1802",
		result.YearOfFirstObservation)

	assert.Equal(t,
		2014,
		result.YearOfLastObservation,
		"convertToMinorPlanet YearOfLastObservation %s expected 2014",
		result.YearOfLastObservation)
}

func TestConvertT3S5154(t *testing.T) {
	var entry = "T3S5154 17.1   0.15 J77AO  17.78418  247.82110  104.38071    9.61380  0.2757131  0.18128053   3.0919701    MPC 12559     8   1    6 days              Bardwell   2000          5154 T-3           19771017"
	var result, err = convertToMinorPlanet(entry)
	if err != nil {
		t.Fatalf("convertToMinorPlanet returned an error %s", err)
	}
	if result.ID != "5154 T-3" {
		t.Errorf("convertToMinorPlanet ID %s expected 5154 T-3", result.ID)
	}

	if result.ReadableDesignation != "5154 T-3" {
		t.Errorf("convertToMinorPlanet ReadableDesignation %s expected 5154 T-3",
			result.ReadableDesignation)
	}

	var expected = time.Date(1977, time.October, 17, 0, 0, 0, 0, time.UTC)

	if !expected.Equal(result.DateOfLastObservation) {
		t.Errorf("convertToMinorPlanet DateOfLastObservation %s expected %s",
			result.DateOfLastObservation.Format("2006-01-02T15:04:00 -0700"),
			expected.Format("2006-01-02T15:04:00 -0700"))
	}

	if result.ArcLength != 6 {
		t.Errorf("convertToMinorPlanet ArcLength %s expected 6",
			result.ArcLength)
	}

	if result.YearOfFirstObservation != 0 {
		t.Errorf("convertToMinorPlanet YearOfFirstObservation %s expected 0",
			result.YearOfFirstObservation)
	}

	if result.YearOfLastObservation != 0 {
		t.Errorf("convertToMinorPlanet YearOfLastObservation %s expected 0",
			result.YearOfLastObservation)
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
		func(c string) string {
			var _, e = convertToMinorPlanet(c)
			return e.Error()
		})
}

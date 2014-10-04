/*
	Provides a simple method to read the Minor Planet Center data files.

*/
package gompcreader

/**
	TODO:
  opposition
  year of observations
  arc length
  perterbers translation
	Split into package and main.
  tests
  documentation
  Readme
  publish
*/

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
MinorPlanet is the result of reading a record from a file
*/
type MinorPlanet struct {
	Id                           string
	AbsoluteMagnitude            float64
	Slope                        float64
	Epoch                        time.Time
	MeanAnomalyEpoch             float64
	ArgumentOfPerihelion         float64
	LongitudeOfTheAscendingNode  float64
	InclinationToTheEcliptic     float64
	OrbitalEccentricity          float64
	MeanDailyMotion              float64
	SemimajorAxis                float64
	UncertaintyParameter         string
	Reference                    string
	NumberOfObservations         int64
	NumberOfOppositions          int64
	RMSResidual                  float64
	CoarseIndicatorOfPerturbers  string
	PreciseIndicatorOfPerturbers string
	ComputerName                 string
	HexDigitFlags                int64
	ReadableDesignation          string
	DateOfLastObservation        time.Time
	YearOfFirstObservation       int64
	YearOfLastObservation        int64
	ArcLength                    int64
}

/*
Use this to create a new minor planet center reader.

This takes a path as a string to the file and returns the reader structure.

If there is a problem opening the file it will return nil for the reader and an
error indicating what went wrong.
*/
func NewMpcReader(filePath string) (*MpcReader, error) {
	var reader *MpcReader = new(MpcReader)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	reader.F = bufio.NewReader(file)
	return reader, nil

}

/*
Simple wrapper around a bufio.Reader used to read the file. Should be constructed
using NewMpcReader(string)
*/
type MpcReader struct {
	F *bufio.Reader
}

/*
ReadEntry returns the next minor planet from the file or error if there is a problem
reading the record.

Note: this will return an io.EOF when the end of the file is reached.
*/
func (reader *MpcReader) ReadEntry() (*MinorPlanet, error) {
	buffer, err := reader.findLine()
	if err != nil {
		return nil, err
	}

	result := convertToMinorPlanet(buffer)
	return result, nil
}

/*
Takes a chunk of the buffer and returns it as a string
*/
func readString(buffer []byte, start int, end int) string {
	s := string(buffer[start:end])
	return strings.Trim(s, " ")
}

/*
Takes a chunk of the buffer and reads it as a float

Note returns zero on error. This may not be ideal.
*/
func readFloat(buffer []byte, start int, end int) float64 {
	s := readString(buffer, start, end)
	result, error := strconv.ParseFloat(s, 64)
	if error == nil {
		return result
	}
	return 0.0
}

func readInt(buffer []byte, start int, end int) int64 {
	s := readString(buffer, start, end)
	result, error := strconv.ParseInt(s, 10, 64)
	if error == nil {
		return result
	}
	return 0
}

func readHexInt(buffer []byte, start int, end int) int64 {
	s := readString(buffer, start, end)
	result, error := strconv.ParseInt(s, 16, 64)
	if error == nil {
		return result
	}
	return 0
}

func readTime(buffer []byte, start int, end int) time.Time {
	s := readString(buffer, start, end)
	t, _ := time.Parse("20060102", s)
	return t
}

/*
Reads a packed int from the buffer

Packed ints encode the most significant digit using 0-9A-Za-z to cover 0 to 62
This is used as a base for the packed identifier and the packed date.
*/
func readPackedInt(buffer []byte, start int, end int) int64 {
	var result int64 = 0
	var decimal int64 = 1
	var localBuffer = readString(buffer, start, end)
	if len(localBuffer) > 0 {

		for i := len(localBuffer) - 1; i > 0; i = i - 1 {
			var working = localBuffer[i]
			if working >= '0' && working <= '9' {
				result = result + (int64(working-'0') * decimal)
				decimal = decimal * 10
			}
		}


		var working = localBuffer[0]
		if working >= 'a' && working <= 'z' {
			result = result + (int64(working-'a'+36) * decimal)
		} else if working >= 'A' && working <= 'Z' {
			result = result + (int64(working-'A'+10) * decimal)
		} else if working >= '0' && working <= '9' {
			result = result + (int64(working-'0') * decimal)
		}
	}
	return result
}

/*
Read a packed identifier from the buffer and return it as a string.

There are three different types of identifier in the file.

The first is a simple packed int. There are identifiable by only having numbers
after the first digit.

The second starts with a packed int then has a two digit code stored in
positions 3 and 6, position 5 can be ignored.

The third starts with a two character code and has a packed int on the end.
These should be swapped around to build the final identifier.
*/
func readPackedIdentifier(buffer []byte, start int, end int) string {
	if onlyNumbers(buffer, start+1, end) {
		return strconv.FormatInt(readPackedInt(buffer, start, end), 10)
	} else if buffer[start+2] >= '0' && buffer[start+2] <= '9' {
		result := strconv.FormatInt(readPackedInt(buffer, start, start+3), 10) + " " + string(buffer[start+3]) + string(buffer[start+6])
		number := readPackedInt(buffer, start+4, start+6)
		if number > 0  {
			return result + strconv.FormatInt(number, 10)
		}
		return result
	} else {
		number := readPackedInt(buffer, start+3, start+7)
		return strconv.FormatInt(number, 10) + " " + string(buffer[start]) + "-" + string(buffer[start+1])
	}
}

/*
Packed time fields are simply three packed int representing year, month and day
*/
func readPackedTime(buffer []byte, start int, end int) time.Time {
	year := int(readPackedInt(buffer, start, start+3))
	month := int(readPackedInt(buffer, start+3, start+4))
	day := int(readPackedInt(buffer, start+4, start+5))
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

/*
Helper function to check if a section of the buffer only contains numbers and
spaces. Used for decoding packed ints.
*/
func onlyNumbers(buffer []byte, start int, end int) bool {
	for i := start; i < end; i = i + 1 {
		if buffer[i] != ' ' && (buffer[i] < '0' || buffer[i] > '9') {
			return false
		}
	}
	return true
}

/*
Read a line from the file.
It will keep reading more lines until it finds one that is 203 characters long
(The length of a data record).

If it gets to the end of the file it will return io.EOF for error
*/
func (reader *MpcReader) findLine() ([]byte, error) {
	var result []byte
	var err error
	for len(result) != 203 {
		result, err = reader.F.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

/*
Convert a byte buffer into a minor planet. This takes apart the buffer and
populates. The MinorPlanet struct
*/
func convertToMinorPlanet(buffer []byte) *MinorPlanet {
	var result = new(MinorPlanet)

	result.Id = readPackedIdentifier(buffer, 0, 6)
	result.AbsoluteMagnitude = readFloat(buffer, 8, 13)
	result.Slope = readFloat(buffer, 14, 19)
	result.Epoch = readPackedTime(buffer, 20, 25)
	result.MeanAnomalyEpoch = readFloat(buffer, 26, 35)
	result.ArgumentOfPerihelion = readFloat(buffer, 37, 47)
	result.LongitudeOfTheAscendingNode = readFloat(buffer, 48, 57)
	result.InclinationToTheEcliptic = readFloat(buffer, 59, 68)
	result.OrbitalEccentricity = readFloat(buffer, 70, 79)
	result.MeanDailyMotion = readFloat(buffer, 80, 91)
	result.SemimajorAxis = readFloat(buffer, 92, 103)
	result.UncertaintyParameter = readString(buffer, 105, 106)
	result.Reference = readString(buffer, 107, 116)
	result.NumberOfObservations = readInt(buffer, 117, 122)
	result.NumberOfOppositions = readInt(buffer, 123, 126)
	// ignore opposition for a second.
	result.RMSResidual = readFloat(buffer, 137, 141)
	result.CoarseIndicatorOfPerturbers = readString(buffer, 142, 145)
	result.PreciseIndicatorOfPerturbers = readString(buffer, 146, 149)
	result.ComputerName = readString(buffer, 150, 160)
	result.HexDigitFlags = readHexInt(buffer, 161, 165)
	result.ReadableDesignation = readString(buffer, 166, 194)
	result.DateOfLastObservation = readTime(buffer, 194, 202)

	// optional parts depending on number of observations.
	//result.yearOfFirstObservation = readInt(buffer, xxx, yyy)
	//result.yearOfLastObservation = readInt(buffer, xxx, yyy)
	//result.arcLength = readInt(buffer, xxx, yyy)

	return result
}

func main() {

	mpcReader, err := NewMpcReader("/home/wselwood/MinorPlanets/MPCORB.DAT")
	if err != nil {
		fmt.Println("error creating mpcReader " + err.Error())
		panic(err)
	}

  var count int64 = 0
	result, err := mpcReader.ReadEntry()
	for err == nil {
		fmt.Println(result.Id + ":" + result.ReadableDesignation)
		result, err = mpcReader.ReadEntry()
		count = count + 1
	}

	if err != nil && err != io.EOF {
		fmt.Println("error reading line " + err.Error())
	}

	fmt.Println("read " + strconv.FormatInt(count, 10) + " records")
}

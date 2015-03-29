/*
Package gompcreader provides a simple method to read the Minor Planet Center
data files.

Designed to load the files from http://www.minorplanetcenter.net/iau/MPCORB.html
and http://www.minorplanetcenter.net/iau/ECS/MPCAT/MPCAT.html

This will not read the comets files or the observations files. If you try you
will probably find the first call to ReadEntry() returns io.EOF as it has skipped
all the records in the file.

This can handle both the gzipped and uncompressed versions of the files. If you
need speed and don't care about space use the uncompressed versions.

To read a file first construct a MpcReader using the NewMpcReader(string) function.
This requires the path to a file.

Each record is read using the ReadEntry() function. Note this may consume more
than one line from the file if the next line is not the correct length. These
are usually the comments at the top of the file or the blank section sperators.
When you get to the end of the file this will return io.EOF.

Finaly when you are done with the reader you should Close() it. This should
normally be defered.
*/
package gompcreader

/**
	TODO:
  perterbers translation
*/

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
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
	ID                           string
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
NewMpcReader is used to create a new minor planet center reader.

This takes a path as a string to the file and returns the reader structure.

If there is a problem opening the file it will return nil for the reader and an
error indicating what went wrong.

This will automatically detect if the file extension suggests a gziped version
of the file and open it correctly.
*/
func NewMpcReader(filePath string) (*MpcReader, error) {
	var reader = new(MpcReader)
	var err error
	reader.f, err = os.Open(filePath)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(filePath, ".gz") {
		reader.g, err = gzip.NewReader(reader.f)
		if err != nil {
			return nil, err
		}

		reader.s = bufio.NewScanner(reader.g)
	} else {
		reader.g = nil
		reader.s = bufio.NewScanner(reader.f)
	}

	return reader, nil

}

/*
MpcReader is a simple data structure to keep track of readers and scanners for
reading the requested file. Should be constructed using NewMpcReader(string) and
closed using Close() before disposal.
*/
type MpcReader struct {
	f *os.File
	g *gzip.Reader
	s *bufio.Scanner
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

	result, err := convertToMinorPlanet(buffer)
	return result, err
}

/*
Close the reader down. This will clean up the open file handle.

This should be defered just after NewMpcReader has been called
*/
func (reader *MpcReader) Close() {
	if reader.g != nil {
		reader.g.Close()
	}
	reader.f.Close()
}

/*
Takes a chunk of the buffer and returns it as a string
*/
func readString(buffer string) string {
	return strings.TrimFunc(buffer, cutSec)
}

func cutSec(input rune) bool {
	return input == ' '
}

/*
Takes a chunk of the buffer and reads it as a float
*/
func readFloat(buffer string) (float64, error) {
	s := readString(buffer)
	return strconv.ParseFloat(s, 64)
}

func readInt(buffer string) (int64, error) {
	s := readString(buffer)
	return strconv.ParseInt(s, 10, 64)
}

func readHexInt(buffer string) (int64, error) {
	s := readString(buffer)
	return strconv.ParseInt(s, 16, 64)
}

// Some times we get dates with out months and days.
// For the sake of not losing data we make them the first of janurary.
// Strangely most of these appear to be around 1995 1996
// Other dates are handled normally.
func readTime(buffer string) (time.Time, error) {
	s := readString(buffer)
	if strings.HasSuffix(s, "0000") {
		s = fmt.Sprint(s[:4], "0101")
	}
	return time.ParseInLocation("20060102", s, time.UTC)
}

/*
Reads a packed int from the buffer

Packed ints encode the most significant digit using 0-9A-Za-z to cover 0 to 61
This is used as a base for the packed identifier and the packed date.
*/
func readPackedInt(buffer string) int64 {
	var result int64
	var decimal int64 = 1
	var localBuffer = strings.TrimFunc(buffer, cutSec)
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
func readPackedIdentifier(buffer string) string {
	if onlyNumbers(buffer[1:]) {
		return strconv.FormatInt(readPackedInt(buffer), 10)
	} else if buffer[2] >= '0' && buffer[2] <= '9' {
		var output bytes.Buffer
		output.WriteString(strconv.FormatInt(readPackedInt(buffer[0:3]), 10))
		output.WriteRune(' ')
		output.WriteByte(buffer[3])
		output.WriteByte(buffer[6])
		number := readPackedInt(buffer[4:6])
		if number > 0 {
			output.WriteString(strconv.FormatInt(number, 10))
		}
		return output.String()
	} else {
		var output bytes.Buffer
		output.WriteString(strconv.FormatInt(readPackedInt(buffer[3:7]), 10))
		output.WriteRune(' ')
		output.WriteByte(buffer[0])
		output.WriteRune('-')
		output.WriteByte(buffer[1])
		return output.String()
	}
}

/*
Packed time fields are simply three packed int representing year, month and day
*/
func readPackedTime(buffer string) time.Time {
	var tb = readString(buffer)
	year := int(readPackedInt(tb[0:3]))
	month := int(readPackedInt(tb[3:4]))
	day := int(readPackedInt(tb[4:5]))
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

/*
The Arc length column contains strings like "  4 days"

First we use readString to trim the string.
Then we split the string on spaces. If we didn't trim the split would include empty strings
Then convert the result to an int
*/
func readArcLength(buffer string) (int64, error) {
	var s = readString(buffer)
	var np = strings.SplitN(s, " ", 2)
	if np == nil || len(np) < 2 {
		return 0, errors.New("Arc length didn't have enough parts")
	}
	return readInt(np[0])
}

/*
Helper function to check if a section of the buffer only contains numbers and
spaces. Used for decoding packed ints.
*/
func onlyNumbers(buffer string) bool {
	for _, v := range buffer {
		if v != ' ' && (v < '0' || v > '9') {
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
func (reader *MpcReader) findLine() (string, error) {
	var result string
	var err error

	for len(result) != 202 {
		if reader.s.Scan() {
			result = reader.s.Text()
		} else {
			err = reader.s.Err()
			if err != nil {
				return "", err
			}
			return "", io.EOF
		}
	}
	return result, nil
}

/*
Convert a byte buffer into a minor planet. This takes apart the buffer and
populates. The MinorPlanet struct
*/
func convertToMinorPlanet(buffer string) (*MinorPlanet, error) {
	var r = new(MinorPlanet)
	var err error

	r.ID = readPackedIdentifier(buffer[0:7])

	// the following two columns are alowed to be blank
	r.AbsoluteMagnitude, _ = readFloat(buffer[8:13])
	r.Slope, _ = readFloat(buffer[14:19])

	r.Epoch = readPackedTime(buffer[20:25])

	r.MeanAnomalyEpoch, err = readFloat(buffer[26:35])
	if err != nil {
		return nil, err
	}

	r.ArgumentOfPerihelion, err = readFloat(buffer[37:47])
	if err != nil {
		return nil, err
	}

	r.LongitudeOfTheAscendingNode, err = readFloat(buffer[48:57])
	if err != nil {
		return nil, err
	}

	r.InclinationToTheEcliptic, err = readFloat(buffer[59:68])
	if err != nil {
		return nil, err
	}

	r.OrbitalEccentricity, err = readFloat(buffer[70:79])
	if err != nil {
		return nil, err
	}

	r.MeanDailyMotion, err = readFloat(buffer[80:91])
	if err != nil {
		return nil, err
	}

	r.SemimajorAxis, err = readFloat(buffer[92:103])
	if err != nil {
		return nil, err
	}

	r.UncertaintyParameter = readString(buffer[105:106])
	r.Reference = readString(buffer[107:116])
	r.NumberOfObservations, _ = readInt(buffer[117:122])
	r.NumberOfOppositions, _ = readInt(buffer[123:126])

	// The next column has different values depending on the NumberOfOppositions
	// When there has been more than one there are two years showing the first and last observations
	// When there is less than or equal one then there is the amount of orbit we have seen.
	if r.NumberOfOppositions > 1 {
		r.YearOfFirstObservation, err = readInt(buffer[127:131])
		if err != nil {
			return nil, err
		}

		r.YearOfLastObservation, err = readInt(buffer[132:136])
		if err != nil {
			return nil, err
		}
	} else {
		r.ArcLength, err = readArcLength(buffer[127:136])
		if err != nil {
			return nil, err
		}
	}

	// This column is optional. Some times it is blank
	r.RMSResidual, _ = readFloat(buffer[137:141])

	r.CoarseIndicatorOfPerturbers = readString(buffer[142:145])
	r.PreciseIndicatorOfPerturbers = readString(buffer[146:149])
	r.ComputerName = readString(buffer[150:160])

	r.HexDigitFlags, err = readHexInt(buffer[161:165])
	if err != nil {
		return nil, err
	}

	r.ReadableDesignation = readString(buffer[166:194])

	r.DateOfLastObservation, err = readTime(buffer[194:202])
	if err != nil {
		return nil, err
	}

	return r, nil
}

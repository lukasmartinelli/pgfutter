package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	csv "github.com/JensRantil/go-csv"
	"github.com/cheggaaa/pb"
)

func containsDelimiter(col string) bool {
	return strings.Contains(col, ";") || strings.Contains(col, ",") ||
		strings.Contains(col, "|") || strings.Contains(col, "\t") ||
		strings.Contains(col, "^") || strings.Contains(col, "~")
}

// Parse the delimiter for an escape sequence. This allows windows users to pass
// in \t since they cannot pass "`t" or "$Tab" to the program.
func parseDelimiter(delim string, skip bool) string {
	if !strings.HasPrefix(delim, "\\") || skip {
		return delim
	}
	switch delim {
	case "\\t":
		{
			return "\t"
		}
	default:
		{
			return delim
		}
	}
}

// Parse columns from first header row or from flags
func parseColumns(reader *csv.Reader, skipHeader bool, fields string) ([]string, error) {
	var err error
	var columns []string
	if fields != "" {
		columns = strings.Split(fields, ",")

		if skipHeader {
			reader.Read() //Force consume one row
		}
	} else {
		columns, err = reader.Read()
		fmt.Printf("%v columns\n%v\n", len(columns), columns)
		if err != nil {
			fmt.Printf("FOUND ERR\n")
			return nil, err
		}
	}

	for _, col := range columns {
		if containsDelimiter(col) {
			return columns, errors.New("Please specify the correct delimiter with -d.\n" +
				"Header column contains a delimiter character: " + col)
		}
	}

	for i, col := range columns {
		columns[i] = postgresify(col)
	}

	return columns, nil
}

func copyCSVRows(i *Import, reader *csv.Reader, ignoreErrors bool,
	delimiter string, columns []string, nullDelimiter string) (error, int, int) {
	success := 0
	failed := 0

	for {
		cols := make([]interface{}, len(columns))
		record, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			line := strings.Join(record, delimiter)
			failed++

			if ignoreErrors {
				os.Stderr.WriteString(string(line))
				continue
			} else {
				err = fmt.Errorf("%s: %s", err, line)
				return err, success, failed
			}
		}

		//Loop ensures we don't insert too many values and that
		//values are properly converted into empty interfaces
		for i, col := range record {
			cols[i] = strings.Replace(col, "\x00", "", -1)
			// bytes.Trim(b, "\x00")
			// cols[i] = col
		}

		err = i.AddRow(nullDelimiter, cols...)

		if err != nil {
			line := strings.Join(record, delimiter)
			failed++

			if ignoreErrors {
				os.Stderr.WriteString(string(line))
				continue
			} else {
				err = fmt.Errorf("%s: %s", err, line)
				return err, success, failed
			}
		}

		success++
	}

	return nil, success, failed
}

func importCSV(filename string, connStr string, schema string, tableName string, ignoreErrors bool,
	skipHeader bool, fields string, delimiter string, excel bool, nullDelimiter string) error {

	db, err := connect(connStr, schema)
	if err != nil {
		return err
	}
	defer db.Close()

	dialect := csv.Dialect{}
	dialect.Delimiter, _ = utf8.DecodeRuneInString(delimiter)

	// Excel 2008 and 2011 and possibly other versions uses a carriage return \r
	// rather than a line feed \n as a newline
	if excel {
		dialect.LineTerminator = "\r"
	} else {
		dialect.LineTerminator = "\n"
	}

	var reader *csv.Reader
	var bar *pb.ProgressBar
	if filename != "" {
		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()

		bar = NewProgressBar(file)
		reader = csv.NewDialectReader(io.TeeReader(file, bar), dialect)
	} else {
		reader = csv.NewDialectReader(os.Stdin, dialect)
	}

	columns, err := parseColumns(reader, skipHeader, fields)
	if err != nil {
		return err
	}

	i, err := NewCSVImport(db, schema, tableName, columns)
	if err != nil {
		return err
	}

	var success, failed int
	if filename != "" {
		bar.Start()
		err, success, failed = copyCSVRows(i, reader, ignoreErrors, delimiter, columns, nullDelimiter)
		bar.Finish()
	} else {
		err, success, failed = copyCSVRows(i, reader, ignoreErrors, delimiter, columns, nullDelimiter)
	}

	if err != nil {
		lineNumber := success + failed
		if !skipHeader {
			lineNumber++
		}
		return fmt.Errorf("line %d: %s", lineNumber, err)
	}

	fmt.Println(fmt.Sprintf("%d rows imported into %s.%s", success, schema, tableName))

	if ignoreErrors && failed > 0 {
		fmt.Println(fmt.Sprintf("%d rows could not be imported into %s.%s and have been written to stderr.", failed, schema, tableName))
	}

	return i.Commit()
}

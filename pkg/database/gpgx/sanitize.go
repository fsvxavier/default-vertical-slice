package gpgx

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Part is either a string or an int. A string is raw SQL. An int is a
// argument placeholder.
type Part any

type Query struct {
	Parts []Part
}

// utf.DecodeRune returns the utf8.RuneError for errors. But that is actually rune U+FFFD -- the unicode replacement
// character. utf8.RuneError is not an error if it is also width 3.
//
// https://github.com/jackc/pgx/issues/1380
const replacementcharacterwidth = 3

func (q *Query) Sanitize(args ...any) (string, error) {
	argUse := make([]bool, len(args))
	buf := &bytes.Buffer{}

	for _, part := range q.Parts {
		var str string
		switch part := part.(type) {
		case string:
			str = part
		case int:
			argIdx := part - 1
			if argIdx >= len(args) {
				return "", fmt.Errorf("insufficient arguments")
			}
			arg := args[argIdx]
			switch arg := arg.(type) {
			case nil:
				str = "null"
			case int64:
				str = strconv.FormatInt(arg, 10)
			case float64:
				str = strconv.FormatFloat(arg, 'f', -1, 64)
			case bool:
				str = strconv.FormatBool(arg)
			case []byte:
				str = QuoteBytes(arg)
			case string:
				str = QuoteString(arg)
			case time.Time:
				str = arg.Truncate(time.Microsecond).Format("'2006-01-02 15:04:05.999999999Z07:00:00'")
			default:
				return "", fmt.Errorf("invalid arg type: %T", arg)
			}
			argUse[argIdx] = true
		default:
			return "", fmt.Errorf("invalid Part type: %T", part)
		}
		buf.WriteString(str)
	}

	for i, used := range argUse {
		if !used {
			return "", fmt.Errorf("unused argument: %d", i)
		}
	}
	return buf.String(), nil
}

func NewQuery(sql string) (*Query, error) {
	slx := &sqlLexer{
		src:     sql,
		stateFn: rawState,
	}

	for slx.stateFn != nil {
		slx.stateFn = slx.stateFn(slx)
	}

	query := &Query{Parts: slx.parts}

	return query, nil
}

func QuoteString(str string) string {
	return "'" + strings.ReplaceAll(str, "'", "''") + "'"
}

func QuoteBytes(buf []byte) string {
	return `'\x` + hex.EncodeToString(buf) + "'"
}

type sqlLexer struct {
	stateFn stateFn
	src     string
	parts   []Part
	start   int
	pos     int
	nested  int
}

type stateFn func(*sqlLexer) stateFn

func rawState(slx *sqlLexer) stateFn {
	for {
		runer, width := utf8.DecodeRuneInString(slx.src[slx.pos:])
		slx.pos += width

		switch runer {
		case 'e', 'E':
			nextRune, width := utf8.DecodeRuneInString(slx.src[slx.pos:])
			if nextRune == '\'' {
				slx.pos += width
				return escapeStringState
			}
		case '\'':
			return singleQuoteState
		case '"':
			return doubleQuoteState
		case '$':
			nextRune, _ := utf8.DecodeRuneInString(slx.src[slx.pos:])
			if '0' <= nextRune && nextRune <= '9' {
				if slx.pos-slx.start > 0 {
					slx.parts = append(slx.parts, slx.src[slx.start:slx.pos-width])
				}
				slx.start = slx.pos
				return placeholderState
			}
		case '-':
			nextRune, width := utf8.DecodeRuneInString(slx.src[slx.pos:])
			if nextRune == '-' {
				slx.pos += width
				return oneLineCommentState
			}
		case '/':
			nextRune, width := utf8.DecodeRuneInString(slx.src[slx.pos:])
			if nextRune == '*' {
				slx.pos += width
				return multilineCommentState
			}
		case utf8.RuneError:
			if width != replacementcharacterwidth {
				if slx.pos-slx.start > 0 {
					slx.parts = append(slx.parts, slx.src[slx.start:slx.pos])
					slx.start = slx.pos
				}
				return nil
			}
		}
	}
}

func singleQuoteState(slx *sqlLexer) stateFn {
	for {
		runer, width := utf8.DecodeRuneInString(slx.src[slx.pos:])
		slx.pos += width

		switch runer {
		case '\'':
			nextRune, width := utf8.DecodeRuneInString(slx.src[slx.pos:])
			if nextRune != '\'' {
				return rawState
			}
			slx.pos += width
		case utf8.RuneError:
			if width != replacementcharacterwidth {
				if slx.pos-slx.start > 0 {
					slx.parts = append(slx.parts, slx.src[slx.start:slx.pos])
					slx.start = slx.pos
				}
				return nil
			}
		}
	}
}

func doubleQuoteState(slx *sqlLexer) stateFn {
	for {
		runer, width := utf8.DecodeRuneInString(slx.src[slx.pos:])
		slx.pos += width

		switch runer {
		case '"':
			nextRune, width := utf8.DecodeRuneInString(slx.src[slx.pos:])
			if nextRune != '"' {
				return rawState
			}
			slx.pos += width
		case utf8.RuneError:
			if width != replacementcharacterwidth {
				if slx.pos-slx.start > 0 {
					slx.parts = append(slx.parts, slx.src[slx.start:slx.pos])
					slx.start = slx.pos
				}
				return nil
			}
		}
	}
}

// placeholderState consumes a placeholder value. The $ must have already has
// already been consumed. The first rune must be a digit.
func placeholderState(slx *sqlLexer) stateFn {
	num := 0

	for {
		runer, width := utf8.DecodeRuneInString(slx.src[slx.pos:])
		slx.pos += width

		if '0' <= runer && runer <= '9' {
			num *= 10
			num += int(runer - '0')
		} else {
			slx.parts = append(slx.parts, num)
			slx.pos -= width
			slx.start = slx.pos
			return rawState
		}
	}
}

func escapeStringState(slx *sqlLexer) stateFn {
	for {
		runer, width := utf8.DecodeRuneInString(slx.src[slx.pos:])
		slx.pos += width

		switch runer {
		case '\\':
			_, width = utf8.DecodeRuneInString(slx.src[slx.pos:])
			slx.pos += width
		case '\'':
			nextRune, width := utf8.DecodeRuneInString(slx.src[slx.pos:])
			if nextRune != '\'' {
				return rawState
			}
			slx.pos += width
		case utf8.RuneError:
			if width != replacementcharacterwidth {
				if slx.pos-slx.start > 0 {
					slx.parts = append(slx.parts, slx.src[slx.start:slx.pos])
					slx.start = slx.pos
				}
				return nil
			}
		}
	}
}

func oneLineCommentState(slx *sqlLexer) stateFn {
	for {
		runer, width := utf8.DecodeRuneInString(slx.src[slx.pos:])
		slx.pos += width

		switch runer {
		case '\\':
			_, width = utf8.DecodeRuneInString(slx.src[slx.pos:])
			slx.pos += width
		case '\n', '\r':
			return rawState
		case utf8.RuneError:
			if width != replacementcharacterwidth {
				if slx.pos-slx.start > 0 {
					slx.parts = append(slx.parts, slx.src[slx.start:slx.pos])
					slx.start = slx.pos
				}
				return nil
			}
		}
	}
}

func multilineCommentState(slx *sqlLexer) stateFn {
	for {
		runer, width := utf8.DecodeRuneInString(slx.src[slx.pos:])
		slx.pos += width

		switch runer {
		case '/':
			nextRune, width := utf8.DecodeRuneInString(slx.src[slx.pos:])
			if nextRune == '*' {
				slx.pos += width
				slx.nested++
			}
		case '*':
			nextRune, width := utf8.DecodeRuneInString(slx.src[slx.pos:])
			if nextRune != '/' {
				continue
			}

			slx.pos += width
			if slx.nested == 0 {
				return rawState
			}
			slx.nested--

		case utf8.RuneError:
			if width != replacementcharacterwidth {
				if slx.pos-slx.start > 0 {
					slx.parts = append(slx.parts, slx.src[slx.start:slx.pos])
					slx.start = slx.pos
				}
				return nil
			}
		}
	}
}

// SanitizeSQL replaces placeholder values with args. It quotes and escapes args
// as necessary. This function is only safe when standard_conforming_strings is
// on.
func SanitizeSQL(sql string, args ...any) (string, error) {
	query, err := NewQuery(sql)
	if err != nil {
		return "", err
	}
	return query.Sanitize(args...)
}

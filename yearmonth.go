package epochdate

import (
	"errors"
	"time"
)

const rfc3339YearMonth = "2006-01"

// ClampYearMonth returns a YearMonth from its constituent year and month
// parts. If the result is out of the representable range, it'll be clamped
// to the nearest representable extreme. Most applications should just get
// YearMonth values via the Date.YearMonth or YearMonth.UnmarshalText
// methods (i.e. JSON decoding).
//
func ClampYearMonth(year int, month time.Month) YearMonth {
	ym, _ := newYearMonth(year, month)
	return ym
}

var errYearMonthOutOfRange = errors.New("epochdate: YearMonth input must be in range [1970-01,TODO]")

func newYearMonth(year int, month time.Month) (YearMonth, error) {
	ym := 12*(year-minYear) + int(month-1)
	if ym < 0 {
		return 0, errYearMonthOutOfRange
	}
	if ym > maxDate {
		return maxDate, errYearMonthOutOfRange
	}
	return YearMonth(ym), nil
}

// YearMonth represents an ordinal year-month combination, such that
// incrementing the value that represents December 2019 yields a value that
// represents January 2020. Each ordinal value semantically covers a range
// of dates, e.g. the value 0 semantically covers "1970-01" (or the range of
// dates from 1970-01-01 through 1970-01-31, inclusive.
//
// The total representable range is 1970-01 through 7431-04 inclusive,
// though only values 1970-01 through 2149-04 inclusive are fully compatible
// with Date, and 2149-06 is only partially compatible with Date.
//
// The representable range of YearMonth is greater than Date: calling
// OrdinalDate will clamp to the maximum Date value if needed.
//
type YearMonth uint16

// IsZero returns true if the receiver holds the minimum representable
// YearMonth value.
func (ym YearMonth) IsZero() bool {
	return ym == 0
}

// IsMax returns true if the receiver holds the maximum representable
// YearMonth value.
func (ym YearMonth) IsMax() bool {
	return ym == maxDate
}

// StartTime returns the first inclusive time instant covered by the
// receiver, relative to the given location, i.e. the zeroth nanosecond of
// the first day of the month.
//
func (ym YearMonth) StartTime(loc *time.Location) time.Time {
	y := int(minYear + ym/12)
	m := time.Month(ym%12 + 1)
	return time.Date(y, m, 1, 0, 0, 0, 0, loc)
}

// EndTime returns the last inclusive time instant (last nanosecond) covered
// by the receiver, relative to the given location, i.e. the last
// representable time.Time moment of the last day of the month.
//
func (ym YearMonth) EndTime(loc *time.Location) time.Time {
	// we don't specify nanoseconds, since we always want the minimum
	// representable unit, in case that's ever smaller than nanoseconds.
	return ym.StartTime(loc).AddDate(0, 1, 0).Add(-1)
}

// StartDate returns the Date representing the first day of the full month
// represented by the receiver. If the result is out of range for Date, the
// maximum Date value will be returned instead.
//
func (ym YearMonth) StartDate() Date {
	return ClampFromTime(ym.StartTime(time.UTC))
}

// EndDate returns the Date representing the last day of the full month
// represented by the receiver. If the result is out of range for Date, the
// maximum Date value will be returned instead.
//
func (ym YearMonth) EndDate() Date {
	return ClampFromTime(ym.EndTime(time.UTC))
}

// String returns a representation of the receiver in the form year-month,
// for example, "2020-01".
//
func (ym YearMonth) String() string {
	return ym.Format(rfc3339YearMonth)
}

// Format is shorthand for `ym.StartTime(time.UTC).Format`.
func (ym YearMonth) Format(layout string) string {
	return ym.StartTime(time.UTC).Format(layout)
}

// MarshalText implements a TextMarshaler for encoding YearMonth values as
// strings, always of the form year-month ("2020-01").
//
func (ym YearMonth) MarshalText() ([]byte, error) {
	b := ym.StartTime(time.UTC).AppendFormat(nil, rfc3339YearMonth)
	return b, nil
}

// UnmarshalText implements a TextUnmarshaler for decoding YearMonth values
// from JSON strings or other textual inputs, using one of the forms
// year-month ("2020-01") or year-month-day ("2020-01-01"). If using the
// year-month-day form, the day is validated but then discarded.
//
// An error will be returned if the input is out of range.
//
func (ym *YearMonth) UnmarshalText(b []byte) error {
	s := string(b)
	t, err := time.Parse(RFC3339, s)
	if err != nil {
		t, err = time.Parse(rfc3339YearMonth, s)
	}
	if err != nil {
		return err
	}
	y, m, _ := t.Date()
	v, err := newYearMonth(y, m)
	if err != nil {
		return err
	}
	*ym = v
	return nil
}

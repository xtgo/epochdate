// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package epochdate provides date handling in a compact format.
//
// Represents dates from Jan 1 1970 - Jun 6 2149 as days since the Unix epoch.
// This format requires 2 bytes (it's a uint16), in contrast to the 16 or 20
// byte representations (on 32 or 64-bit systems, respectively) used by the
// standard time package.
//
// Timezone is accounted for when applicable; when converting from standard
// time format into a Date, the date relative to the time value's zone is
// retained.  Times at any point during a given day (relative to timezone) are
// normalized to the same date.
//
// Conversely, conversions back to standard time format may be done using the
// Local, UTC, and In methods (semantically corresponding to the same-named
// Time methods), but with the result normalized to midnight (the beginning of
// the day) relative to that timezone.
//
// All functions and methods with the same names as those found in the stdlib
// time package have identical semantics in epochdate, with the exception that
// epochdate truncates time-of-day information.
//
// epochdate does not handle leap seconds, and is thus generally follows the
// behavior of the standard library time package, which it uses for
// translating to and from Date and YearMonth values, parsing, and
// formatting.
//
// epochdate does not handle leap seconds, which may result in a low risk of
// date inconsistencies when calling New/ClampFromUnix directly, if that
// timestamp falls on the zeroth second of a day more than six months in the
// future. In particular, if that second was selected as a leap second, the
// initial call may result in a value one lower than if that same call were
// made after the leap second had been amended into the standard library
// time package. Using New/ClampFromTime and New/ClampFromDate should
// prevent such a discrepancy.
//
package epochdate

import (
	"bytes"
	"errors"
	"time"
)

// Clamp will use the nearest representable date if the input is out of
// range but otherwise a correct time.Time or parseable string. When false,
// parsing or conversion will return an error when the input represents an
// out-of-range date. This does not affect the wrap-around behavior of
// arithmetic on the underlying uint16 value. For compatibility reasons, the
// default value of this package is false.
//
// Consider using the ClampFrom* functions instead of NewFrom* when clamping
// behavior is desired, as the ClampFrom* variants do not depend on the
// value of this variable.
var Clamp = false

const (
	day      = 60 * 60 * 24
	nsPerSec = 1e9
	bits     = 16
	minYear  = 1970
	maxUnix  = (1<<bits)*day - 1
	maxDate  = 1<<bits - 1
)

// Format constants, for use with Parse and the Date.Format method.
const (
	RFC3339        = "2006-01-02"
	AmericanShort  = "1-2-06"
	AmericanCommon = "01-02-06"
)

// ErrOutOfRange is returned if the input date is not a representable Date.
var ErrOutOfRange = errors.New("epochdate: dates must be in the range [1970-01-01,2149-06-06]")

// Today returns the local date at this instant. If the local date does not
// fall within the representable range, then then zero value will be returned
// (1970-01-01).
func Today() Date {
	date, _ := NewFromTime(time.Now())
	return date
}

// TodayUTC returns the date at this instant, relative to UTC. If the UTC
// date does not fall within the representable range, then then zero value
// will be returned (1970-01-01).
func TodayUTC() Date {
	date, _ := NewFromTime(time.Now().UTC())
	return date
}

// Parse follows the same semantics as time.Parse, but ignores time-of-day
// information and returns a Date value.
func Parse(layout, value string) (Date, error) {
	t, err := time.Parse(layout, value)
	if err != nil {
		return 0, err
	}
	return NewFromTime(t)
}

// ParseRFC is like Parse, except that the layout is fixed to RFC3339.
func ParseRFC(value string) (Date, error) {
	t, err := time.Parse(RFC3339, value)
	if err != nil {
		return 0, err
	}
	return NewFromTime(t)
}

// MustParse is like Parse, except that it panics if an error occurs.
func MustParse(layout, value string) Date {
	d, err := Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return d
}

// MustParseRFC is like ParseRFC, except that it panics if an error occurs.
func MustParseRFC(value string) Date {
	d, err := ParseRFC(value)
	if err != nil {
		panic(err)
	}
	return d
}

// ClampFromTime behaves like NewFromTime, except that it clamps
// out-of-range dates rather than returning an error. This means that either
// range errors are undetectable, or the representable date range must be
// reduced by one value on each extreme (this 1970-01-01 and 2149-06-06
// could be considered errors, if error handling is needed).
//
func ClampFromTime(t time.Time) Date {
	s := t.Unix()
	_, offset := t.Zone()
	return ClampFromUnix(s + int64(offset))
}

// NewFromTime returns a Date equivalent to NewFromDate(t.Date()),
// where t is a time.Time object.
//
func NewFromTime(t time.Time) (Date, error) {
	s := t.Unix()
	_, offset := t.Zone()
	return NewFromUnix(s + int64(offset))
}

// ClampFromDate behaves like NewFromDate, except that it clamps
// out-of-range dates rather than returning an error. This means that either
// range errors are undetectable, or the representable date range must be
// reduced by one value on each extreme (this 1970-01-01 and 2149-06-06
// could be considered errors, if error handling is needed).
//
func ClampFromDate(year int, month time.Month, day int) Date {
	return ClampFromUnix(time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix())
}

// NewFromDate returns a Date value corresponding to the supplied
// year, month, and day.
//
func NewFromDate(year int, month time.Month, day int) (Date, error) {
	return NewFromUnix(time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix())
}

// ClampFromUnix behaves like NewFromUnix, except that it clamps
// out-of-range dates rather than returning an error. This means that either
// range errors are undetectable, or the representable date range must be
// reduced by one value on each extreme (this 1970-01-01 and 2149-06-06
// could be considered errors, if error handling is needed).
//
// Please see package documentation regarding leap seconds when using this
// function.
//
func ClampFromUnix(seconds int64) Date {
	switch {
	case seconds < 0:
		return 0

	case seconds > maxUnix:
		return maxDate
	}

	return Date(seconds / day)
}

// NewFromUnix creates a Date from a Unix timestamp, relative to any location
// Specifically, if you pass in t.Unix(), where t is a time.Time value with a
// non-UTC zone, you may receive an unexpected Date. Unless this behavior is
// specifically desired (returning the date in one location at the given time
// instant in another location), it's best to use epochdate.NewFromTime(t),
// which normalizes the resulting Date value by adjusting for zone offsets.
//
// Please see package documentation regarding leap seconds when using this
// function.
//
func NewFromUnix(seconds int64) (Date, error) {
	switch {
	case UnixInRange(seconds):
		return Date(seconds / day), nil

	case Clamp && seconds < 0:
		return 0, nil

	case Clamp && seconds > maxUnix:
		return maxDate, nil
	}

	return 0, ErrOutOfRange
}

// UnixInRange is true if the provided Unix timestamp is in Date's
// representable range. The timestamp is interpreted according to the semantics
// used by NewFromUnix. You probably won't need to use this, since this will
// only return false if NewFromUnix returns an error of ErrOutOfRange.
//
func UnixInRange(seconds int64) bool {
	return seconds >= 0 && seconds <= maxUnix
}

// Date stores the number of days since Jan 1, 1970. The last representable
// date is June 6, 2149.
type Date uint16

// Returns an RFC3339/ISO-8601 date string, of the form "2006-01-02".
func (d Date) String() string {
	return d.Format(RFC3339)
}

// Unix returns the number of seconds elapsed since Jan 1 1970 UTC, from the
// start of the given date value. In this case, the date is considered to be
// a UTC date, rather than a location-independent date.
//
func (d Date) Unix() int64 {
	return int64(d) * day
}

// UnixNano is semantically identical to the Unix method, except that it
// returns elapsed nanoseconds.
//
func (d Date) UnixNano() int64 {
	return int64(d) * day * nsPerSec
}

// YearMonth returns the YearMonth that corresponds to the receiver.
func (d Date) YearMonth() YearMonth {
	y, m, _ := d.Date()
	return ClampYearMonth(y, m)
}

// IsZero returns true if d represents the zero value for the Date type.
func (d Date) IsZero() bool {
	return d == 0
}

// IsMin returns true if d represents the minimum representable date.
// Equivalent to IsZero.
//
func (d Date) IsMin() bool {
	return d == 0
}

// IsMax returns true if d represents the maximum representable date.
func (d Date) IsMax() bool {
	return d == maxDate
}

// Format is identical to time.Time.Format, except that any time-of-day format
// specifiers that are used will be equivalent to "00:00:00Z".
//
func (d Date) Format(layout string) string {
	return d.UTC().Format(layout)
}

// Date is semantically identical to the behavior of t.Date(), where t is a
// time.Time value.
//
func (d Date) Date() (year int, month time.Month, day int) {
	return d.UTC().Date()
}

// UTC returns a UTC Time object set to 00:00:00 on the given date.
func (d Date) UTC() time.Time {
	return time.Unix(int64(d)*day, 0).UTC()
}

// Local returns a local Time object set to 00:00:00 on the given date.
func (d Date) Local() time.Time {
	return d.In(time.Local)
}

// In returns a location-relative Time object set to 00:00:00 on the given date.
func (d Date) In(loc *time.Location) time.Time {
	t := time.Unix(int64(d)*day, 0).In(loc)
	_, offset := t.Zone()
	return t.Add(time.Duration(-offset) * time.Second)
}

// MarshalText implements encoding.TextMarshaler.
func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.Format(RFC3339)), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (d *Date) UnmarshalText(data []byte) error {
	v, err := ParseRFC(string(data))
	if err != nil {
		return err
	}
	*d = v
	return nil
}

// MarshalJSON implements json.Marshaler.
func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(d.Format(`"` + RFC3339 + `"`)), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Date) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, jsonNull) {
		return nil
	}
	data = bytes.Trim(data, `"`)
	return d.UnmarshalText(data)
}

var jsonNull = []byte(`null`)

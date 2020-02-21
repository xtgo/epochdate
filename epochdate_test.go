// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package epochdate

import (
	"encoding"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

var (
	_ encoding.TextMarshaler   = Date(0)
	_ encoding.TextUnmarshaler = new(Date)
	_ json.Marshaler           = Date(0)
	_ json.Unmarshaler         = new(Date)
)

func try(fn func()) (panicked bool) {
	defer func() { panicked = recover() != nil }()
	fn()
	return false
}

func TestMustParse(t *testing.T) {
	tests := []struct {
		name      string
		layout    string
		input     string
		want      Date
		wantPanic bool
	}{
		{
			name:      "bad_layout",
			layout:    "blah",
			input:     "2020-02-15",
			wantPanic: true,
		},
		{
			name:      "bad_input",
			layout:    RFC3339,
			input:     "blah",
			wantPanic: true,
		},
		{
			name:   "american_slash",
			layout: "1/2/06",
			input:  "3/26/19",
			want:   ClampFromDate(2019, 3, 26),
		},
		{
			name:   "zero",
			layout: RFC3339,
			input:  "1970-01-01",
			want:   0,
		},
		{
			name:   "max",
			layout: RFC3339,
			input:  "2149-06-06",
			want:   maxDate,
		},
		{
			name:      "underflow",
			layout:    RFC3339,
			input:     "1969-12-31",
			wantPanic: true,
		},
		{
			name:      "overflow",
			layout:    RFC3339,
			input:     "2149-06-07",
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Date

			panicked := try(func() { got = MustParse(tt.layout, tt.input) })
			switch {
			case !panicked && tt.wantPanic:
				t.Errorf("MustParse(%q, %q) should not panicked", tt.layout, tt.input)

			case panicked && !tt.wantPanic:
				t.Errorf("MustParse(%q, %q) should not have panicked", tt.layout, tt.input)

			case got != tt.want:
				t.Errorf("MustParse(%q, %q) = %d, want %d", tt.layout, tt.input, got, tt.want)
			}

			if tt.layout != RFC3339 {
				return
			}

			panicked = try(func() { got = MustParseRFC(tt.input) })
			switch {
			case !panicked && tt.wantPanic:
				t.Errorf("MustParseRFC(%q) should not panicked", tt.input)

			case panicked && !tt.wantPanic:
				t.Errorf("MustParseRFC(%q) should not have panicked", tt.input)

			case got != tt.want:
				t.Errorf("MustParseRFC(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestClampFromUnix(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  Date
	}{
		{
			name:  "underflow",
			input: -1,
			want:  0,
		},
		{
			name:  "zero",
			input: 0,
			want:  0,
		},
		{
			name:  "last_moment_of_first_day",
			input: 86399,
			want:  0,
		},
		{
			name:  "first_moment_of_second_day",
			input: 86400,
			want:  1,
		},
		{
			name:  "max_minus_one",
			input: maxUnix - 1,
			want:  maxDate,
		},
		{
			name:  "max",
			input: maxUnix,
			want:  maxDate,
		},
		{
			name:  "overflow",
			input: maxUnix + 1,
			want:  maxDate,
		},
		{
			name:  "ultra_max",
			input: 1<<63 - 1,
			want:  maxDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClampFromUnix(tt.input)
			if got != tt.want {
				t.Errorf("ClampFromUnix(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func isLastMinuteOfDay(t time.Time) bool {
	return t.Hour() == 23 && t.Minute() == 59
}

func TestToday(t *testing.T) {
	now := time.Now()
	if isLastMinuteOfDay(now) {
		t.Skip("skipping time-sensitive test near end of day")
	}

	got := Today()
	want := ClampFromDate(now.Date())

	if got != want {
		t.Errorf("Today() = %q, want %q", got, want)
	}
}

func TestTodayUTC(t *testing.T) {
	now := time.Now().UTC()
	if isLastMinuteOfDay(now) {
		t.Skip("skipping time-sensitive test near end of day")
	}

	got := TodayUTC()
	want := ClampFromDate(now.Date())

	if got != want {
		t.Errorf("Today() = %q, want %q", got, want)
	}
}

func TestDate_String(t *testing.T) {
	tests := []struct {
		name  string
		input Date
		want  string
	}{
		{
			name:  "zero",
			input: 0,
			want:  "1970-01-01",
		},
		{
			name:  "max",
			input: 65535,
			want:  "2149-06-06",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.String()
			if got != tt.want {
				t.Errorf("%d.String() = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

type triple struct {
	year  int
	month time.Month
	day   int
}

type equiv struct {
	date Date
	unix int64
	trip triple
	str  string
}

var equivs = []equiv{
	{0, 0, triple{1970, 1, 1}, "1970-01-01"},
	{0, day - 1, triple{1970, 1, 1}, "1970-01-01"},
	{366, 366 * day, triple{1971, 1, 2}, "1971-01-02"},
	{366, 367*day - 1, triple{1971, 1, 2}, "1971-01-02"},
	{65535, 65535 * day, triple{2149, 6, 6}, "2149-06-06"},
	{65535, 65536*day - 1, triple{2149, 6, 6}, "2149-06-06"},
}

var extrema = []struct {
	unix  int64
	valid bool
}{
	{-1, false},
	{0, true},
	{65536*day - 1, true},
	{65536 * day, false},
}

func TestDate_equivalences(t *testing.T) {
	for _, e := range equivs {
		if unix, err := NewFromUnix(e.unix); err != nil {
			t.Fatal(err)
		} else if trip, err := NewFromDate(e.trip.year, e.trip.month, e.trip.day); err != nil {
			t.Fatal(err)
		} else if str, err := Parse(RFC3339, e.str); err != nil {
			t.Fatal(err)
		} else if e.date != unix || e.date != trip || e.date != str {
			t.Fatal("Unexpected non-equivalence:", e.date, unix, trip, str)
		}
	}
}

func TestDate_extrema(t *testing.T) {
	var desc string
	for _, e := range extrema {
		if UnixInRange(e.unix) != e.valid {
			if e.valid {
				desc = "valid"
			} else {
				desc = "invalid"
			}
			t.Fatal("Unix timestamp", e.unix, "should be", desc)
		}
	}
}

func TestDate_timezone_irrelevance(t *testing.T) {
	const hour = 60 * 60
	min := time.FixedZone("min", -12*hour)
	max := time.FixedZone("max", +14*hour)
	t1 := time.Date(2149, 06, 06, 0, 0, 0, 0, min)
	t2 := time.Date(2149, 06, 06, 0, 0, 0, 0, max)
	var (
		d1, d2 Date
		err    error
	)
	if d1, err = NewFromTime(t1); err != nil {
		t.Fatal(err)
	}
	if d2, err = NewFromTime(t2); err != nil {
		t.Fatal(err)
	}
	if d1 != d2 {
		t.Fatal("Expected", t1, "and", t2, "to result in same date; got", d1, "and", d2)
	}
}

func TestDate_UTC(t *testing.T) {
	var date Date
	local := date.Local()
	utc := date.UTC()
	prefix := "1970-01-01T00:00:00"
	if f := local.Format(time.RFC3339); !strings.HasPrefix(f, prefix) {
		t.Fatalf("Expected local time to %q; got %q\n", prefix, f)
	} else if f := utc.Format(time.RFC3339); !strings.HasPrefix(f, prefix) {
		t.Fatalf("Expected universal time to %q; got %q\n", prefix, f)
	}
}

func TestDate_Unix(t *testing.T) {
	var d Date = 1
	const (
		dayInSecs     = 60 * 60 * 24
		dayInNanosecs = dayInSecs * 1e9
	)
	if s := d.Unix(); s != dayInSecs {
		t.Error("Expected Date(1).Unix() to return", dayInSecs, "but got", s)
	}
	if ns := d.UnixNano(); ns != dayInNanosecs {
		t.Error("Expected Date(1).UnixNano() to return", dayInNanosecs, "but got", ns)
	}
}

func TestDate_MarshalText(t *testing.T) {
	const (
		unquoted = "1970-01-02"
		quoted   = `"` + unquoted + `"`
		n        = 1
	)
	d := Date(n)
	b, err := d.MarshalText()
	if err != nil {
		t.Error("Unexpected MarshalText error:", err)
	} else if string(b) != unquoted {
		t.Errorf("Expected Date(%d).MarshalText() to return %#q but got %#q", n, unquoted, b)
	}
	b, err = d.MarshalJSON()
	if err != nil {
		t.Error("Unexpected MarshalJSON error:", err)
	} else if string(b) != quoted {
		t.Errorf("Expected Date(%d).MarshalJSON() to return %#q but got %#q", n, quoted, b)
	}
	err = d.UnmarshalText([]byte(unquoted))
	if err != nil {
		t.Error("Unexpected UnmarshalText error:", err)
	} else if d != n {
		t.Errorf("Expected Date(%d).UnmarshalText() to return %#q but got %#q", n, unquoted, b)
	}
	err = d.UnmarshalJSON([]byte(quoted))
	if err != nil {
		t.Error("Unexpected UnmarshalJSON error:", err)
	} else if d != n {
		t.Errorf("Expected Date(%d).UnmarshalJSON() to return %#q but got %#q", n, quoted, b)
	}
}

func TestDate_UnmarshalText_error(t *testing.T) {
	input := []byte("blah")
	var d Date
	err := d.UnmarshalText(input)
	if err == nil {
		t.Errorf("Date.UnmarshalText(%q) = nil, want error", input)
	}
}

func TestDate_UnmarshalJSON_null(t *testing.T) {
	data := []byte("null")
	input := Date(123)
	date := input
	want := input

	err := json.Unmarshal(data, &date)
	if err != nil {
		t.Errorf("json.Unmarshal(%q) = %q, want nil", data, err)
	}

	if date != want {
		t.Errorf("json.Unmarshal(%q, %v) -> %v, want %v",
			data, input, date, want)
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		input   string
		want    string // unused by non-clamping error results
		wantErr bool   // unused by clamping results
	}{
		{
			input:   "1969-12-31",
			want:    "1970-01-01",
			wantErr: true,
		},
		{
			input:   "1970-01-01",
			want:    "1970-01-01",
			wantErr: false,
		},
		{
			input:   "2019-07-11",
			want:    "2019-07-11",
			wantErr: false,
		},
		{
			input:   "2149-06-06",
			want:    "2149-06-06",
			wantErr: false,
		},
		{
			input:   "2149-06-07",
			want:    "2149-06-06",
			wantErr: true,
		},
	}

	t.Run("normal", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				got, err := ParseRFC(tt.input)
				switch {
				case tt.wantErr && err == nil:
					t.Errorf("ParseRFC(%q) = nil [err], want error", tt.input)

				case !tt.wantErr && err != nil:
					t.Errorf("ParseRFC(%q) = %q [err], want nil", tt.input, err)

				case err != nil && got != 0:
					t.Errorf("ParseRFC(%q) = %q, want %q", tt.input, got, Date(0))

				case err == nil && got != MustParseRFC(tt.want):
					t.Errorf("ParseRFC(%q) = %q, want %q", tt.input, got, tt.want)
				}
			})
		}
	})

	t.Run("clamp", func(t *testing.T) {
		Clamp = true
		defer func() { Clamp = false }()

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				got, err := ParseRFC(tt.input)
				switch {
				case err != nil:
					t.Errorf("ParseRFC(%q) = %q, want nil", tt.input, err)

				case got != MustParseRFC(tt.want):
					t.Errorf("ParseRFC(%q) = %q, want %q", tt.input, got, tt.want)

				case got == 0 && !got.IsZero():
					t.Errorf("%q.IsZero() = false, want true", tt.input)

				case got.IsZero() != got.IsMin():
					t.Errorf("%q.IsZero() != %[1]q.IsMin()", tt.input)

				case got.IsMin() && got.IsMax():
					t.Errorf("%q.IsMin() && %[1]q.IsMax()", tt.input)

				case got == maxDate && !got.IsMax():
					t.Errorf("%q.IsMax() = false, want true", tt.input)
				}
			})
		}
	})
}

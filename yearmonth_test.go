package epochdate

import (
	"testing"
	"time"
)

func TestClampYearMonth(t *testing.T) {
	tests := []struct {
		name  string
		year  int
		month time.Month
		want  YearMonth
	}{
		{
			name:  "minus_one_year",
			year:  1969,
			month: time.January,
			want:  0,
		},
		{
			name:  "minus_one_month",
			year:  1969,
			month: time.December,
			want:  0,
		},
		{
			name:  "zero",
			year:  1970,
			month: time.January,
			want:  0,
		},
		{
			name:  "one_month",
			year:  1970,
			month: time.February,
			want:  1,
		},
		{
			name:  "thirteen_months",
			year:  1970,
			month: time.December + 1,
			want:  12,
		},
		{
			name:  "one_year",
			year:  1971,
			month: time.January,
			want:  12,
		},
		{
			name:  "max",
			year:  7431,
			month: time.April,
			want:  65535,
		},
		{
			name:  "overflow",
			year:  7431,
			month: time.May,
			want:  65535,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClampYearMonth(tt.year, tt.month)
			if got != tt.want {
				t.Errorf("ClampYearMonth(%d, %d) = %v, want %v", tt.year, tt.month, got, tt.want)
			}
		})
	}
}

func TestYearMonth_properties(t *testing.T) {
	today := TodayUTC()

	tests := []struct {
		name   string
		ym     YearMonth
		isZero bool
		isMax  bool
	}{
		{
			name:   "zero",
			ym:     0,
			isZero: true,
			isMax:  false,
		},
		{
			name:   "one",
			ym:     1,
			isZero: false,
			isMax:  false,
		},
		{
			name:   "today",
			ym:     today.YearMonth(),
			isZero: false,
			isMax:  false,
		},
		{
			name:   "max_minus_one",
			ym:     65534,
			isZero: false,
			isMax:  false,
		},
		{
			name:   "max",
			ym:     65535,
			isZero: false,
			isMax:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ym.IsZero() != tt.isZero {
				t.Errorf("%q.IsZero() = %v, want %v", tt.ym, tt.ym.IsZero(), tt.isZero)
			}

			if tt.ym.IsMax() != tt.isMax {
				t.Errorf("%q.IsMax() = %v, want %v", tt.ym, tt.ym.IsMax(), tt.isMax)
			}
		})
	}
}

func TestYearMonth_StartTime(t *testing.T) {
	loc := time.FixedZone("UTC-1", -3600)
	now := time.Now().UTC()
	today, _ := NewFromTime(now)

	tests := []struct {
		name  string
		ym    YearMonth
		loc   *time.Location
		start time.Time
		end   time.Time
	}{
		{
			name:  "zero_utc",
			ym:    0,
			loc:   time.UTC,
			start: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(1970, 1, 31, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:  "zero_utc_minus_one",
			ym:    0,
			loc:   time.UTC,
			start: time.Date(1969, 12, 31, 23, 0, 0, 0, loc),
			end:   time.Date(1970, 1, 31, 22, 59, 59, 999999999, loc),
		},
		{
			name:  "today",
			ym:    today.YearMonth(),
			loc:   time.UTC,
			start: func() time.Time { y, m, _ := now.Date(); return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC) }(),
			end:   func() time.Time { y, m, _ := now.Date(); return time.Date(y, m+1, 1, 0, 0, 0, -1, time.UTC) }(),
		},
		{
			name:  "max_utc",
			ym:    65535,
			loc:   time.UTC,
			start: time.Date(7431, 4, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(7431, 4, 30, 23, 59, 59, 999999999, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.StartTime(tt.loc)
			if !got.Equal(tt.start) {
				t.Errorf("%q.StartTime() = %q, want %q", tt.ym, got, tt.start)
			}

			got = tt.ym.EndTime(tt.loc)
			if !got.Equal(tt.end) {
				t.Errorf("%q.EndTime() = %q, want %q", tt.ym, got, tt.end)
			}
		})
	}
}

func TestYearMonth_StartDate(t *testing.T) {
	today := TodayUTC()

	tests := []struct {
		name  string
		ym    YearMonth
		start Date
		end   Date
	}{
		{
			name:  "zero",
			ym:    0,
			start: ClampFromDate(1970, 1, 1),
			end:   ClampFromDate(1970, 1, 31),
		},
		{
			name:  "today",
			ym:    today.YearMonth(),
			start: func() Date { y, m, _ := today.Date(); return ClampFromDate(y, m, 1) }(),
			end:   func() Date { y, m, _ := today.Date(); return ClampFromDate(y, m+1, 0) }(),
		},
		{
			name:  "max",
			ym:    65535,
			start: maxDate,
			end:   maxDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.StartDate()
			if got != tt.start {
				t.Errorf("%q.StartDate() = %q, want %q", tt.ym, got, tt.start)
			}

			got = tt.ym.EndDate()
			if got != tt.end {
				t.Errorf("%q.EndDate() = %q, want %q", tt.ym, got, tt.end)
			}
		})
	}
}

func TestYearMonth_String(t *testing.T) {
	today := TodayUTC()

	tests := []struct {
		name string
		ym   YearMonth
		want string
	}{
		{
			name: "zero",
			ym:   0,
			want: "1970-01",
		},
		{
			name: "one",
			ym:   1,
			want: "1970-02",
		},
		{
			name: "today",
			ym:   today.YearMonth(),
			want: today.Format("2006-01"),
		},
		{
			name: "max",
			ym:   65535,
			want: "7431-04",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ym.String()
			if got != tt.want {
				t.Errorf("%d.String() = %q, want %q", tt.ym, got, tt.want)
			}

			const format = "2006-01"
			got = tt.ym.Format(format)
			if got != tt.want {
				t.Errorf("%d.Format(%q) = %q, want %q", tt.ym, format, got, tt.want)
			}

			buf, err := tt.ym.MarshalText()
			if err != nil {
				t.Errorf("%d.MarshalText() = %q [error], want nil", tt.ym, err)
			}

			got = string(buf)
			if got != tt.want {
				t.Errorf("%d.MarshalText() = %q, want %q", tt.ym, got, tt.want)
			}
		})
	}
}

func TestYearMonth_UnmarshalText(t *testing.T) {
	today := TodayUTC()

	tests := []struct {
		name    string
		input   string
		want    YearMonth
		wantErr bool
	}{
		{
			name:    "empty_input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "bad_input",
			input:   "blah",
			wantErr: true,
		},
		{
			name:  "full_date",
			input: "1970-02-15",
			want:  1,
		},
		{
			name:  "zero",
			input: "1970-01",
			want:  0,
		},
		{
			name:  "one",
			input: "1970-02",
			want:  1,
		},
		{
			name:  "2020",
			input: "2020-06",
			want:  ClampYearMonth(2020, 6),
		},
		{
			name:  "today",
			input: today.Format("2006-01"),
			want:  today.YearMonth(),
		},
		{
			name:  "max",
			input: "7431-04",
			want:  65535,
		},
		{
			name:    "underflow",
			input:   "1969-12",
			wantErr: true,
		},
		{
			name:    "overflow",
			input:   "7431-05",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ym YearMonth
			err := ym.UnmarshalText([]byte(tt.input))

			switch {
			case !tt.wantErr && err != nil:
				t.Errorf("YearMonth.UnmarshalText(%q) = %q [error], want nil", tt.input, err)

			case tt.wantErr && err == nil:
				t.Errorf("YearMonth.UnmarshalText(%q) = nil, want error", tt.input)

			case ym != tt.want:
				t.Errorf("YearMonth.UnmarshalText(%q) -> %d, want %d", tt.input, ym, tt.want)
			}
		})
	}
}

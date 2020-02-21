# epochdate

epochdate is a small library for storing contemporary dates (without
time-of-day information) in 2 bytes per date, instead of the 16-20 bytes that
are used by a time.Time value from the Go standard library.

epochdate.Date is a uint16 value representing the number of days since Jan 1
1970, with each value representing a whole calendar day. The maximum
representable Date is Jun 6 2149. epochdate.YearMonth is a uint16 value
representing the number of whole months Jan 1 1970, with each value
representing a whole calendar month. The maximum representable YearMonth is Apr
7431, though for interoperability with Date, values above May or June 2149
should not be used.

Arithmetical operations on these types react predictably, for example,
incrementing a Date is equivalent to "the following day," while adding 12 to a
YearMonth is equivalent to "same month of the following year."

## YearMonth operations

YearMonth enables simple date or time-alignment around month boundaries, for example:

    // first day of current month
    TodayUTC().YearMonth().StartDate()

    // last day of current month
    TodayUTC().YearMonth().EndDate()

    // first day of next month
    TodayUTC().YearMonth().EndDate() + 1

    // ... or equivalently
    (TodayUTC().YearMonth() + 1).StartDate()

    // last time instant of the current month
    TodayUTC().YearMonth.EndTime()

## Encoding/Decoding

Both Date and YearMonth can be encoded to and from JSON, XML, and other string inputs.

Date values can be encoded to or from RFC-3339 dates (i.e. "2020-01-26")

YearDate values encode to a partial RFC-3339 date (i.e. "2020-01"), and can be
decoded from either that same partial format, or a full date (i.e.
"2020-01-26"), in which case the day portion of the input will be validated and
discarded.

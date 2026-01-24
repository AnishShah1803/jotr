package options

import (
	"time"

	"github.com/spf13/cobra"
)

// DateOption provides date-related command options for operations
// that need to target specific dates (e.g., daily notes).
type DateOption struct {
	Yesterday bool
	Today     bool
	Date      time.Time
}

// OutputOption controls output formatting for search and list commands.
type OutputOption struct {
	CountOnly bool
	FilesOnly bool
	PathOnly  bool
	Quiet     bool
	JSON      bool
}

// TimeRangeOption specifies a time range for queries like stats and summaries.
type TimeRangeOption struct {
	Week  bool
	Month bool
	Year  bool
	All   bool
	Days  int
}

func NewDateOption() DateOption {
	return DateOption{
		Date: time.Now(),
	}
}

func (d *DateOption) SetTargetDate() {
	if d.Yesterday {
		d.Date = d.Date.AddDate(0, 0, -1)
	}
}

func (d *DateOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&d.Yesterday, "yesterday", false, "Operate on yesterday's date")
	cmd.Flags().BoolVar(&d.Today, "today", false, "Operate on today's date (default)")
}

func NewOutputOption() OutputOption {
	return OutputOption{}
}

func (o *OutputOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.CountOnly, "count", false, "Show only the count of matches")
	cmd.Flags().BoolVar(&o.FilesOnly, "files", false, "Show only file names without content")
	cmd.Flags().BoolVar(&o.PathOnly, "path", false, "Show only the path")
	cmd.Flags().BoolVar(&o.Quiet, "quiet", false, "Suppress normal output")
	cmd.Flags().BoolVar(&o.JSON, "json", false, "Output in JSON format")
}

func NewTimeRangeOption() TimeRangeOption {
	return TimeRangeOption{
		All: true,
	}
}

func (r *TimeRangeOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&r.Week, "week", false, "Show stats for the last 7 days")
	cmd.Flags().BoolVar(&r.Month, "month", false, "Show stats for the last 30 days")
	cmd.Flags().BoolVar(&r.Year, "year", false, "Show stats for the last 365 days")
	cmd.Flags().BoolVar(&r.All, "all", true, "Show all-time stats (default)")
	cmd.Flags().IntVar(&r.Days, "days", 0, "Show stats for the last N days")
}

func (r *TimeRangeOption) GetDays() int {
	if r.Days > 0 {
		return r.Days
	}
	if r.Week {
		return 7
	}
	if r.Month {
		return 30
	}
	if r.Year {
		return 365
	}
	return 0
}

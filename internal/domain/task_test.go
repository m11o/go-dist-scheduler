package domain

import (
	"testing"
	"time"
)

func TestTask_NextRunTime(t *testing.T) {
	now := time.Date(2024, time.April, 1, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name           string
		cronExpression string
		expected       time.Time
		expectErr      bool
	}{
		{
			name:           "every minute",
			cronExpression: "* * * * *",
			expected:       time.Date(2024, time.April, 1, 0, 1, 0, 0, time.UTC),
			expectErr:      false,
		},
		{
			name:           "every hour",
			cronExpression: "0 * * * *",
			expected:       time.Date(2024, time.April, 1, 1, 0, 0, 0, time.UTC),
			expectErr:      false,
		},
		{
			name:           "every day at midnight",
			cronExpression: "0 0 * * *",
			expected:       time.Date(2024, time.April, 2, 0, 0, 0, 0, time.UTC),
			expectErr:      false,
		},
		{
			name:           "invalid cron expression",
			cronExpression: "invalid",
			expected:       time.Time{},
			expectErr:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			task := &Task{
				CronExpression: tc.cronExpression,
			}

			nextRunTime, err := task.NextRunTime(now)

			if tc.expectErr {
				if err == nil {
					t.Errorf("expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !nextRunTime.Equal(tc.expected) {
					t.Errorf("expected next run time to be %v, but got %v", tc.expected, nextRunTime)
				}
			}
		})
	}
}

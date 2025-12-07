package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJob_MarkAsRunning(t *testing.T) {
	job := &Job{}
	job.MarkAsRunning()

	assert.Equal(t, JobStatusRunning, job.Status)
	assert.NotZero(t, job.StartedAt)
	assert.NotZero(t, job.UpdatedAt)
}

func TestJob_MarkAsSuccess(t *testing.T) {
	job := &Job{}
	job.MarkAsSuccess()

	assert.Equal(t, JobStatusSuccess, job.Status)
	assert.NotZero(t, job.FinishedAt)
	assert.NotZero(t, job.UpdatedAt)
}

func TestJob_MarkAsFailed(t *testing.T) {
	job := &Job{}
	job.MarkAsFailed()

	assert.Equal(t, JobStatusFailed, job.Status)
	assert.NotZero(t, job.FinishedAt)
	assert.NotZero(t, job.UpdatedAt)
}

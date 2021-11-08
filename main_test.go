package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAtLeastTwoTimes(t *testing.T) {
	assert.Equal(t, []string{"1", "2"}, atLeastTwoTimes([]string{"1", "2"}))
	assert.Equal(t, []string{"1", "2"}, atLeastTwoTimes([]string{"1", "2"}))
	assert.Equal(t, []string{"1", "3"}, atLeastTwoTimes([]string{"1", "3"}))
	assert.Equal(t, []string{"1", "3"}, atLeastTwoTimes([]string{"1", "3"}))
	assert.Equal(t, []string{"1", "3"}, atLeastTwoTimes([]string{"1", "2", "3"}))
	assert.Equal(t, []string{"1", "3"}, atLeastTwoTimes([]string{"1", "3"}))
	assert.Equal(t, []string{"1", "3"}, atLeastTwoTimes([]string{"1", "2", "3"}))
	assert.Equal(t, []string{"1", "2", "3"}, atLeastTwoTimes([]string{"1", "2", "3"}))
	assert.Equal(t, []string{"1", "2", "3"}, atLeastTwoTimes([]string{"1", "2", "3"}))
}

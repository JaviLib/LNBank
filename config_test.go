package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateConfig(t *testing.T) {
	assert.NoError(t, CreateConfig("test", "test_service", "random_value"))
	row := DB.QueryRow("select service, value from config where service='test_service' limit 1")
	var serv, value string
	assert.NoError(t, row.Scan(&serv, &value))
	assert.Equal(t, "test_service", serv)
	assert.Equal(t, "random_value", value)
}

func TestReadConfig(t *testing.T) {
	c, err := ReadConfig("test", "test_read_config", "testing_value")
	assert.NoError(t, err)
	assert.Equal(t, "testing_value", c)
	// Now ensure the previously stored value is read
	c, err = ReadConfig("test", "test_read_config", "this should not be showing")
	assert.NoError(t, err)
	assert.Equal(t, "testing_value", c)
}

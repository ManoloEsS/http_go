package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid spacing header before colon
	headers = NewHeaders()
	data = []byte("Host : localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid single header with extra white space
	headers = NewHeaders()
	data = []byte("Host:     localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 34, n)
	assert.False(t, done)

	// Test: Valid 2 headers with existing headers
	headers = NewHeaders()
	data = []byte("Host:     localhost:42069       \r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 34, n)
	assert.False(t, done)

	data = []byte("Hostest: with the mostest\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "with the mostest", headers["hostest"])
	assert.Equal(t, 27, n)
	assert.False(t, done)

	// Test: Valid done after 2 headers
	headers = NewHeaders()
	data = []byte("Host:     localhost:42069       \r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 34, n)
	assert.False(t, done)

	data = []byte("Hostest: with the mostest\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "with the mostest", headers["hostest"])
	assert.Equal(t, 27, n)
	assert.False(t, done)

	data = data[n:]
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: Invalid single header with invalid chars
	headers = NewHeaders()
	data = []byte("H©st:     localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)

	// Test: Multiple values per key
	headers = NewHeaders()
	data = []byte("Set-person: me\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "me", headers["set-person"])
	data = []byte("Set-person: not me\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "me, not me", headers["set-person"])
	data = []byte("Set-person: other me\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "me, not me, other me", headers["set-person"])

}

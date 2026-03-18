package ui

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRGB16Bit(t *testing.T) {
	c, ok := parseRGB("rgb:ffff/0000/8080")
	require.True(t, ok)
	assert.InDelta(t, 1.0, c.R, 0.001)
	assert.InDelta(t, 0.0, c.G, 0.001)
	assert.InDelta(t, 0.502, c.B, 0.01)
}

func TestParseRGB8Bit(t *testing.T) {
	c, ok := parseRGB("rgb:ff/00/80")
	require.True(t, ok)
	assert.InDelta(t, 1.0, c.R, 0.001)
	assert.InDelta(t, 0.0, c.G, 0.001)
	assert.InDelta(t, 0.502, c.B, 0.01)
}

func TestParseRGB1Digit(t *testing.T) {
	c, ok := parseRGB("rgb:f/0/8")
	require.True(t, ok)
	assert.InDelta(t, 1.0, c.R, 0.001)
	assert.InDelta(t, 0.0, c.G, 0.001)
	assert.InDelta(t, 0.533, c.B, 0.01)
}

func TestParseRGB3Digit(t *testing.T) {
	c, ok := parseRGB("rgb:fff/000/800")
	require.True(t, ok)
	assert.InDelta(t, 1.0, c.R, 0.001)
	assert.InDelta(t, 0.0, c.G, 0.001)
	assert.InDelta(t, 0.500, c.B, 0.01)
}

func TestParseRGBInvalid(t *testing.T) {
	_, ok := parseRGB("not-rgb")
	assert.False(t, ok)

	_, ok = parseRGB("rgb:ff/gg/00")
	assert.False(t, ok)

	_, ok = parseRGB("rgb:ff/00")
	assert.False(t, ok)
}

func newTestDetector(data string) *detector {
	return &detector{reader: bufio.NewReader(strings.NewReader(data))}
}

func TestReadForegroundColor(t *testing.T) {
	d := newTestDetector("\x1b]10;rgb:c0c0/caca/f5f5\x07")
	da1, err := d.readNext()
	require.NoError(t, err)
	assert.False(t, da1)
	assert.True(t, d.colors.Detected[sampleForeground])
	assert.InDelta(t, 0.753, d.colors.Colors[sampleForeground].R, 0.01)
}

func TestReadBackgroundColor(t *testing.T) {
	d := newTestDetector("\x1b]11;rgb:1a1a/1b1b/2626\x07")
	da1, err := d.readNext()
	require.NoError(t, err)
	assert.False(t, da1)
	assert.True(t, d.colors.Detected[sampleBackground])
	assert.InDelta(t, 0.102, d.colors.Colors[sampleBackground].R, 0.01)
}

func TestReadANSIColor(t *testing.T) {
	d := newTestDetector("\x1b]4;4;rgb:7a7a/a2a2/f7f7\x07")
	da1, err := d.readNext()
	require.NoError(t, err)
	assert.False(t, da1)
	assert.True(t, d.colors.Detected[4]) // blue
	assert.InDelta(t, 0.478, d.colors.Colors[4].R, 0.01)
}

func TestReadColorWithSTTerminator(t *testing.T) {
	d := newTestDetector("\x1b]10;rgb:ffff/ffff/ffff\x1b\\")
	da1, err := d.readNext()
	require.NoError(t, err)
	assert.False(t, da1)
	assert.True(t, d.colors.Detected[sampleForeground])
	assert.InDelta(t, 1.0, d.colors.Colors[sampleForeground].R, 0.001)
}

func TestReadDA1(t *testing.T) {
	d := newTestDetector("\x1b[?62;c")
	da1, err := d.readNext()
	require.NoError(t, err)
	assert.True(t, da1)
}

func TestReadMultipleResponses(t *testing.T) {
	d := newTestDetector(
		"\x1b]10;rgb:c0c0/caca/f5f5\x07" +
			"\x1b]11;rgb:1a1a/1b1b/2626\x07" +
			"\x1b]4;2;rgb:5050/fafa/7b7b\x07",
	)

	for range 3 {
		da1, err := d.readNext()
		require.NoError(t, err)
		assert.False(t, da1)
	}

	assert.True(t, d.colors.Detected[sampleForeground])
	assert.True(t, d.colors.Detected[sampleBackground])
	assert.True(t, d.colors.Detected[2])
}

func TestDetectedColorsDefaultEmpty(t *testing.T) {
	d := DetectedColors{}
	for i := range sampleCount {
		assert.False(t, d.Detected[i])
		assert.Equal(t, colorful.Color{}, d.Colors[i])
	}
}

// mockTTY simulates a terminal that responds to OSC queries.
// It discards writes (the query) and feeds back the canned response.
type mockTTY struct {
	io.Reader
	io.Writer
}

func newMockTTY(response []byte) *mockTTY {
	return &mockTTY{
		Reader: bytes.NewReader(response),
		Writer: io.Discard,
	}
}

func (m *mockTTY) Read(p []byte) (int, error)  { return m.Reader.Read(p) }
func (m *mockTTY) Write(p []byte) (int, error) { return m.Writer.Write(p) }

func TestDetectFromDarkTheme(t *testing.T) {
	// Simulate a Tokyo Night-like dark terminal responding with
	// foreground, background, blue, green, bright blue, then DA1.
	response := "" +
		"\x1b]10;rgb:c0c0/caca/f5f5\x07" + // foreground
		"\x1b]11;rgb:1a1a/1b1b/2626\x07" + // background
		"\x1b]4;4;rgb:7a7a/a2a2/f7f7\x07" + // blue
		"\x1b]4;2;rgb:9e9e/cece/6a6a\x07" + // green
		"\x1b]4;12;rgb:7d7d/cfcf/ffff\x07" + // bright blue
		"\x1b[?62;22c" // DA1 sentinel

	mock := newMockTTY([]byte(response))
	d := detectFrom(mock)

	assert.True(t, d.Detected[sampleForeground])
	assert.True(t, d.Detected[sampleBackground])
	assert.True(t, d.Detected[4])  // blue
	assert.True(t, d.Detected[2])  // green
	assert.True(t, d.Detected[12]) // bright blue

	assert.InDelta(t, 0.753, d.Colors[sampleForeground].R, 0.01)
	assert.InDelta(t, 0.102, d.Colors[sampleBackground].R, 0.01)
	assert.InDelta(t, 0.478, d.Colors[4].R, 0.01) // blue
}

func TestDetectFromLightTheme(t *testing.T) {
	response := "" +
		"\x1b]10;rgb:3434/3b3b/5858\x07" + // dark foreground
		"\x1b]11;rgb:d5d5/d6d6/dbdb\x07" + // light background
		"\x1b]4;4;rgb:2e2e/7d7d/e9e9\x07" + // blue
		"\x1b[?62;c" // DA1

	mock := newMockTTY([]byte(response))
	d := detectFrom(mock)

	assert.True(t, d.Detected[sampleForeground])
	assert.True(t, d.Detected[sampleBackground])
	assert.True(t, d.Detected[4])

	// Light background should have high lightness
	bgL, _, _ := d.Colors[sampleBackground].OkLch()
	assert.Greater(t, bgL, 0.8)

	// Dark foreground should have low lightness
	fgL, _, _ := d.Colors[sampleForeground].OkLch()
	assert.Less(t, fgL, 0.4)
}

func TestDetectFromPartialResponse(t *testing.T) {
	// Only background responds before DA1
	response := "" +
		"\x1b]11;rgb:1a1a/1b1b/2626\x07" +
		"\x1b[?62;c"

	mock := newMockTTY([]byte(response))
	d := detectFrom(mock)

	assert.True(t, d.Detected[sampleBackground])
	assert.False(t, d.Detected[sampleForeground])
	assert.False(t, d.Detected[4]) // blue not detected
}

func TestDetectFromNoOSCSupport(t *testing.T) {
	// Terminal that doesn't support OSC queries — reader returns EOF
	// immediately (no response data at all).
	mock := newMockTTY([]byte{})
	d := detectFrom(mock)

	for i := range sampleCount {
		assert.False(t, d.Detected[i])
	}
}

func TestDetectFromTimeout(t *testing.T) {
	// Simulate a terminal that hangs — use a pipe that never writes.
	pr, pw := io.Pipe()
	defer pw.Close()

	rw := &mockTTY{Reader: pr, Writer: io.Discard}

	start := time.Now()
	time.AfterFunc(50*time.Millisecond, func() { pr.Close() })
	d := detectFrom(rw)
	elapsed := time.Since(start)

	for i := range sampleCount {
		assert.False(t, d.Detected[i])
	}
	assert.Less(t, elapsed, 200*time.Millisecond, "should respect timeout")
}

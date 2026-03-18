package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/x/term"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	sampleForeground = 16
	sampleBackground = 17
	sampleCount      = 18
)

// DetectedColors holds optional RGB values detected from the terminal.
type DetectedColors struct {
	Colors   [sampleCount]colorful.Color
	Detected [sampleCount]bool
}

// SupportsTrueColor reports whether the terminal is likely to support
// 24-bit color output. COLORTERM is the authoritative signal; when it is
// absent (common over SSH), we infer support from successful OSC palette
// detection.
func (d DetectedColors) SupportsTrueColor() bool {
	colorterm := os.Getenv("COLORTERM")
	if colorterm == "truecolor" || colorterm == "24bit" {
		return true
	}

	return slices.Contains(d.Detected[:], true)
}

// DetectTerminalColors queries the terminal for foreground, background,
// and all 16 ANSI palette colors via OSC sequences. A DA1 request is
// appended as a sentinel. The function returns after the DA1 sentinel
// is received or the timeout expires.
func DetectTerminalColors(timeout time.Duration) DetectedColors {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return DetectedColors{}
	}

	fd := tty.Fd()
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		tty.Close()
		return DetectedColors{}
	}

	cleanup := sync.OnceFunc(func() {
		term.Restore(fd, oldState)
		tty.Close()
	})
	defer cleanup()

	time.AfterFunc(timeout, cleanup)

	return detectFrom(tty)
}

// detectFrom runs the OSC query/response cycle over any ReadWriter.
// It reads responses until DA1 is seen or a read error occurs. The
// caller can close the reader to interrupt.
func detectFrom(rw io.ReadWriter) DetectedColors {
	var query strings.Builder
	oscQuery := func(code string) { fmt.Fprintf(&query, "\x1b]%s;?\x07", code) }
	csiQuery := func(code string) { fmt.Fprintf(&query, "\x1b[%s", code) }

	oscQuery("10") // foreground
	oscQuery("11") // background
	for i := range 16 {
		oscQuery(fmt.Sprintf("4;%d", i))
	}
	csiQuery("c") // DA1 sentinel

	if _, err := rw.Write([]byte(query.String())); err != nil {
		return DetectedColors{}
	}

	d := detector{reader: bufio.NewReader(rw)}
	for {
		da1, err := d.readNext()
		if da1 || err != nil {
			return d.colors
		}
	}
}

// Private

type detector struct {
	colors DetectedColors
	reader *bufio.Reader
}

// readNext scans for the next escape sequence and dispatches to the
// appropriate reader. Returns true when DA1 is seen.
func (d *detector) readNext() (da1 bool, err error) {
	for {
		b, err := d.reader.ReadByte()
		if err != nil {
			return false, err
		}
		if b != 0x1b {
			continue
		}

		b, err = d.reader.ReadByte()
		if err != nil {
			return false, err
		}

		switch b {
		case ']':
			if err := d.readOSC(); err != nil {
				return false, err
			}
			return false, nil
		case '[':
			return d.readCSI()
		}
	}
}

// readOSC reads an OSC sequence body (after ESC ]) until the BEL or
// ST terminator, then stores any color result.
func (d *detector) readOSC() error {
	var body []byte
	for {
		b, err := d.reader.ReadByte()
		if err != nil {
			return err
		}
		switch b {
		case 0x07: // BEL
			d.collectColor(body)
			return nil
		case 0x1b: // possible ST
			next, err := d.reader.ReadByte()
			if err != nil {
				return err
			}
			if next == '\\' {
				d.collectColor(body)
				return nil
			}
			body = append(body, b, next)
		default:
			body = append(body, b)
		}
	}
}

// readCSI reads a CSI sequence (after ESC [) until a final byte.
// Returns true if the final byte is 'c' (DA1).
func (d *detector) readCSI() (da1 bool, err error) {
	for {
		b, err := d.reader.ReadByte()
		if err != nil {
			return false, err
		}
		if b >= 0x40 && b <= 0x7e {
			return b == 'c', nil
		}
	}
}

// collectColor parses an OSC color response body and stores the result.
// The body is the content between ESC ] and the terminator, e.g.
// "10;rgb:c0c0/caca/f5f5" or "4;2;rgb:5050/fafa/7b7b".
func (d *detector) collectColor(body []byte) {
	s := string(body)

	rgbIdx := strings.Index(s, "rgb:")
	if rgbIdx < 0 {
		return
	}
	prefix := strings.TrimRight(s[:rgbIdx], ";")
	rgbStr := s[rgbIdx:]

	clr, ok := parseRGB(rgbStr)
	if !ok {
		return
	}

	index := -1
	switch {
	case prefix == "10":
		index = sampleForeground
	case prefix == "11":
		index = sampleBackground
	case strings.HasPrefix(prefix, "4;"):
		n, err := strconv.Atoi(strings.TrimPrefix(prefix, "4;"))
		if err == nil && n >= 0 && n < 16 {
			index = n
		}
	}

	if index >= 0 {
		d.colors.Colors[index] = clr
		d.colors.Detected[index] = true
	}
}

// Helpers

// parseRGB parses "rgb:RRRR/GGGG/BBBB" into a colorful.Color.
// Handles 1 to 4 hex digits per component (XParseColor format).
func parseRGB(s string) (colorful.Color, bool) {
	s = strings.TrimPrefix(s, "rgb:")
	parts := strings.Split(s, "/")
	if len(parts) != 3 {
		return colorful.Color{}, false
	}

	parse := func(hex string) (float64, bool) {
		if len(hex) < 1 || len(hex) > 4 {
			return 0, false
		}
		v, err := strconv.ParseUint(hex, 16, 16)
		if err != nil {
			return 0, false
		}
		maxVal := [5]float64{0, 0xF, 0xFF, 0xFFF, 0xFFFF}
		return float64(v) / maxVal[len(hex)], true
	}

	r, ok := parse(parts[0])
	if !ok {
		return colorful.Color{}, false
	}
	g, ok := parse(parts[1])
	if !ok {
		return colorful.Color{}, false
	}
	b, ok := parse(parts[2])
	if !ok {
		return colorful.Color{}, false
	}

	return colorful.Color{R: r, G: g, B: b}, true
}

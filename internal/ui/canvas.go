package ui

import (
	"slices"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
)

// zoneLayer pairs a lipgloss Layer with its raw content string, so that
// zone markers can be extracted before canvas compositing strips them.
type zoneLayer struct {
	*lipgloss.Layer
	content string
}

func newZoneLayer(content string) zoneLayer {
	return zoneLayer{Layer: lipgloss.NewLayer(content), content: content}
}

func centeredZoneLayer(content string, containerWidth, containerHeight int) zoneLayer {
	x := max((containerWidth-lipgloss.Width(content))/2, 0)
	y := max((containerHeight-lipgloss.Height(content))/2, 0)
	return zoneLayer{
		Layer:   lipgloss.NewLayer(content).X(x).Y(y),
		content: content,
	}
}

// renderPreservingZones composites layers using lipgloss's canvas, preserving
// any bubblezone markers that the canvas would otherwise strip during its
// cell-based compositing. Zone markers are scanned from each layer before
// rendering, then re-inserted at the correct absolute positions afterward.
func renderPreservingZones(layers ...zoneLayer) string {
	var allMarkers []zoneMarker

	for _, l := range layers {
		for _, m := range findZoneMarkers(l.content) {
			m.row += l.GetY()
			m.col += l.GetX()
			allMarkers = append(allMarkers, m)
		}
	}

	lipLayers := make([]*lipgloss.Layer, len(layers))
	for i, l := range layers {
		lipLayers[i] = l.Layer
	}
	result := lipgloss.NewCanvas(lipLayers...).Render()
	return insertZoneMarkers(result, allMarkers)
}

type zoneMarker struct {
	marker string
	row    int
	col    int
	index  int
}

func findZoneMarkers(content string) []zoneMarker {
	var markers []zoneMarker

	for row, line := range strings.Split(content, "\n") {
		col := 0
		i := 0
		for i < len(line) {
			if line[i] == '\x1b' {
				if i+1 < len(line) && line[i+1] == '[' {
					start := i
					end := i + 2
					for end < len(line) && line[end] >= 0x20 && line[end] <= 0x3F {
						end++
					}
					if end < len(line) {
						end++
					}
					if line[end-1] == 'z' {
						markers = append(markers, zoneMarker{
							marker: line[start:end],
							row:    row,
							col:    col,
							index:  len(markers),
						})
					}
					i = end
				} else if i+1 < len(line) {
					i += 2
				} else {
					i++
				}
				continue
			}

			_, size := utf8.DecodeRuneInString(line[i:])
			col++
			i += size
		}
	}

	return markers
}

func insertZoneMarkers(output string, markers []zoneMarker) string {
	if len(markers) == 0 {
		return output
	}

	lines := strings.Split(output, "\n")

	// Process from bottom-right to top-left so insertions don't shift
	// positions of markers not yet processed.
	slices.SortFunc(markers, func(a, b zoneMarker) int {
		if a.row != b.row {
			return b.row - a.row
		}
		if a.col != b.col {
			return b.col - a.col
		}
		return b.index - a.index
	})

	for _, m := range markers {
		if m.row >= 0 && m.row < len(lines) {
			lines[m.row] = insertAtVisualCol(lines[m.row], m.col, m.marker)
		}
	}

	return strings.Join(lines, "\n")
}

func insertAtVisualCol(line string, col int, insert string) string {
	visualCol := 0
	i := 0

	for i < len(line) {
		if visualCol == col {
			return line[:i] + insert + line[i:]
		}

		if line[i] == '\x1b' {
			if i+1 < len(line) && line[i+1] == '[' {
				end := i + 2
				for end < len(line) && line[end] >= 0x20 && line[end] <= 0x3F {
					end++
				}
				if end < len(line) {
					end++
				}
				i = end
			} else if i+1 < len(line) {
				i += 2
			} else {
				i++
			}
			continue
		}

		_, size := utf8.DecodeRuneInString(line[i:])
		visualCol++
		i += size
	}

	if visualCol == col {
		return line + insert
	}

	return line
}

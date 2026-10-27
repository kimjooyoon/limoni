package widgets_test

import (
	"strings"
	"testing"

	"github.com/thebanri/limoni/core/buffer"
	"github.com/thebanri/limoni/core/cell"
	"github.com/thebanri/limoni/core/terminal"
	"github.com/thebanri/limoni/testkit"
	"github.com/thebanri/limoni/widgets"
)

func choiceWidgets() []struct {
	name string
	make func(string, cell.Style, cell.Style) widgets.Widget
} {
	return []struct {
		name string
		make func(string, cell.Style, cell.Style) widgets.Widget
	}{
		{"checkbox", func(label string, base, focus cell.Style) widgets.Widget {
			return widgets.Checkbox{ID: "choice", Label: label, Style: base, FocusedStyle: focus}
		}},
		{"radio", func(label string, base, focus cell.Style) widgets.Widget {
			return widgets.RadioButton{ID: "choice", Label: label, Style: base, FocusedStyle: focus}
		}},
	}
}

func TestChoiceFrameTheme(t *testing.T) {
	custom := cell.Style{Fg: cell.NewColorRGB(240, 80, 120), Bg: cell.NewColorRGB(40, 50, 60)}
	fallback := cell.Style{Modifier: cell.ModifierReverse | cell.ModifierBold}
	for _, widget := range choiceWidgets() {
		t.Run(widget.name, func(t *testing.T) {
			var focusColours []cell.Color
			for _, preset := range []struct {
				name  string
				theme widgets.Theme
			}{
				{"dark", widgets.DarkTheme()}, {"light", widgets.LightTheme()},
			} {
				t.Run(preset.name, func(t *testing.T) {
					surface := preset.theme.RoleStyle("surface")
					for _, tc := range []struct {
						name    string
						focused bool
						base    cell.Style
						focus   cell.Style
						want    cell.Style
					}{
						{name: "normal", want: surface},
						{name: "focused", focused: true, want: surface.Merge(preset.theme.Focus)},
						{name: "explicit base", focused: true, base: custom, want: surface.Merge(custom).Merge(fallback)},
						{name: "foreground only", focused: true, base: cell.Style{Fg: custom.Fg}, want: surface.Merge(cell.Style{Fg: custom.Fg}).Merge(fallback)},
						{name: "background only", focused: true, base: cell.Style{Bg: custom.Bg}, want: surface.Merge(cell.Style{Bg: custom.Bg}).Merge(fallback)},
						{name: "modifier only", focused: true, base: cell.Style{Modifier: cell.ModifierItalic}, want: surface.Merge(cell.Style{Modifier: cell.ModifierItalic}).Merge(fallback)},
						{name: "explicit focus", focused: true, focus: custom, want: surface.Merge(custom)},
						{name: "both explicit", focused: true, base: cell.Style{Fg: cell.NewColorRGB(1, 2, 3)}, focus: custom, want: surface.Merge(custom)},
					} {
						t.Run(tc.name, func(t *testing.T) {
							term := testkit.NewTerminal(20, 1)
							term.Draw(func(f *terminal.Frame) {
								f.SetTheme(preset.theme)
								if tc.focused {
									f.FocusManager.SetFocused("choice")
								}
								f.RenderWidget(widget.make("Choice", tc.base, tc.focus), term.Area())
							})
							for _, x := range []uint16{1, 4} {
								if got := term.Frame().Buffer.CellAt(x, 0).Style; got != tc.want {
									t.Errorf("cell %d style = %+v; want %+v", x, got, tc.want)
								}
							}
							if tc.name == "focused" {
								focusColours = append(focusColours, term.Frame().Buffer.CellAt(1, 0).Style.Fg)
							}
						})
					}
				})
			}
			if len(focusColours) != 2 || focusColours[0] == focusColours[1] {
				t.Fatal("dark and light frames must draw different focus colours")
			}
		})
	}
}

func TestChoiceEmptyFocusTheme(t *testing.T) {
	for _, widget := range choiceWidgets() {
		for _, themed := range []bool{false, true} {
			term := testkit.NewTerminal(20, 1)
			term.Draw(func(f *terminal.Frame) {
				if themed {
					f.SetTheme(widgets.Theme{})
				}
				f.FocusManager.SetFocused("choice")
				f.RenderWidget(widget.make("Choice", cell.Style{}, cell.Style{}), term.Area())
			})
			want := cell.Style{Modifier: cell.ModifierReverse | cell.ModifierBold}
			if got := term.Frame().Buffer.CellAt(1, 0).Style; got != want {
				t.Errorf("%s themed=%v: style = %+v; want %+v", widget.name, themed, got, want)
			}
		}
	}
}

func TestChoiceLabelClipping(t *testing.T) {
	for _, tc := range []struct {
		width uint16
		label string
		want  string
	}{
		{3, "Choice", "[ ]"},
		{4, "Choice", "[ ]"},
		{7, "Choice", "[ ] Cho"},
		{7, "界面", "[ ] 界"},
		{60, strings.Repeat("Choice", 8), "[ ] " + strings.Repeat("Choice", 8)},
	} {
		for _, widget := range choiceWidgets() {
			area := cell.Rect{Width: tc.width, Height: 1}
			buf := buffer.NewBuffer(area)
			widget.make(tc.label, cell.Style{}, cell.Style{}).Draw(cell.NewContext(area, cell.Style{}), buf)
			want := tc.want
			if widget.name == "radio" {
				want = strings.Replace(want, "[ ]", "( )", 1)
			}
			if got := strings.TrimRight(buf.Snapshot(), " \n"); got != want {
				t.Errorf("%s width=%d label=%q: got %q; want %q", widget.name, tc.width, tc.label, got, want)
			}
		}
	}
}

func TestRadioLeadingLabelClusters(t *testing.T) {
	for _, tc := range []struct {
		label     string
		separator string
		nextX     uint16
	}{
		{"\u0301e", " \u0301", 4},
		{"\u093ee", " \u093e", 4},
		{"\U0001f3fbe", " \U0001f3fb", 5},
		{"\u200de", " \u200d", 4},
	} {
		t.Run(tc.label, func(t *testing.T) {
			area := cell.Rect{Width: 20, Height: 1}
			buf := buffer.NewBuffer(area)
			widgets.RadioButton{ID: "choice", Label: tc.label}.Draw(cell.NewContext(area, cell.Style{}), buf)
			if got := cell.ClusterText(buf.CellAt(3, 0).Content); got != tc.separator {
				t.Errorf("separator cluster = %q; want %q", got, tc.separator)
			}
			if got := buf.CellAt(tc.nextX, 0).Content; got != 'e' {
				t.Errorf("cell %d = %q; want e", tc.nextX, got)
			}
		})
	}
}

func BenchmarkChoiceDrawFocusedTheme(b *testing.B) {
	for _, widget := range choiceWidgets() {
		b.Run(widget.name, func(b *testing.B) {
			area := cell.Rect{Width: 20, Height: 1}
			buf := buffer.NewBuffer(area)
			theme := widgets.DarkTheme()
			ctx := cell.NewContext(area, theme.RoleStyle("surface"))
			ctx.ThemeStyle = theme.RoleStyle
			ctx.FocusedID = "choice"
			w := widget.make("Choice", cell.Style{}, cell.Style{})
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				w.Draw(ctx, buf)
			}
		})
	}
}

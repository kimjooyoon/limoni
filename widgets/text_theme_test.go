package widgets_test

import (
	"testing"

	"github.com/thebanri/limoni/core/buffer"
	"github.com/thebanri/limoni/core/cell"
	"github.com/thebanri/limoni/core/terminal"
	"github.com/thebanri/limoni/testkit"
	"github.com/thebanri/limoni/widgets"
)

func themedTextWidgets() []struct {
	name string
	make func(cell.Style, cell.Style, bool) widgets.Widget
} {
	return []struct {
		name string
		make func(cell.Style, cell.Style, bool) widgets.Widget
	}{
		{"input", func(base, focus cell.Style, forced bool) widgets.Widget {
			state := widgets.NewTextInputState()
			state.SetValue("Hello")
			return widgets.TextInput{ID: "text", State: state, Style: base, FocusedStyle: focus, Focused: forced}
		}},
		{"area", func(base, focus cell.Style, forced bool) widgets.Widget {
			state := widgets.NewTextAreaState()
			state.SetValue("Hello")
			return widgets.TextArea{ID: "text", State: state, Style: base, FocusedStyle: focus, Focused: forced}
		}},
	}
}

func TestTextFrameTheme(t *testing.T) {
	custom := cell.Style{Fg: cell.NewColorRGB(240, 80, 120), Bg: cell.NewColorRGB(40, 50, 60), Modifier: cell.ModifierItalic}
	for _, widget := range themedTextWidgets() {
		t.Run(widget.name, func(t *testing.T) {
			var colours []cell.Color
			for _, preset := range []struct {
				name  string
				theme widgets.Theme
			}{{"dark", widgets.DarkTheme()}, {"light", widgets.LightTheme()}} {
				t.Run(preset.name, func(t *testing.T) {
					surface := preset.theme.RoleStyle("surface")
					for _, tc := range []struct {
						name            string
						managed, forced bool
						base, focus     cell.Style
						want            cell.Style
					}{
						{name: "normal", want: surface},
						{name: "focused", managed: true, want: surface.Merge(preset.theme.Focus)},
						{name: "forced", forced: true, want: surface.Merge(preset.theme.Focus)},
						{name: "explicit base", managed: true, base: custom, want: surface.Merge(custom)},
						{name: "foreground only", managed: true, base: cell.Style{Fg: custom.Fg}, want: surface.Merge(cell.Style{Fg: custom.Fg})},
						{name: "background only", managed: true, base: cell.Style{Bg: custom.Bg}, want: surface.Merge(cell.Style{Bg: custom.Bg})},
						{name: "modifier only", managed: true, base: cell.Style{Modifier: cell.ModifierItalic}, want: surface.Merge(cell.Style{Modifier: cell.ModifierItalic})},
						{name: "explicit focus", managed: true, focus: custom, want: surface.Merge(custom)},
						{name: "forced explicit focus", forced: true, focus: custom, want: surface.Merge(custom)},
						{name: "both explicit", managed: true, base: cell.Style{Fg: cell.NewColorRGB(1, 2, 3)}, focus: custom, want: surface.Merge(custom)},
					} {
						t.Run(tc.name, func(t *testing.T) {
							term := testkit.NewTerminal(10, 2)
							term.Draw(func(f *terminal.Frame) {
								f.SetTheme(preset.theme)
								if tc.managed {
									f.FocusManager.SetFocused("text")
								}
								f.RenderWidget(widget.make(tc.base, tc.focus, tc.forced), term.Area())
							})
							for _, x := range []uint16{0, 4, 7} {
								if got := term.Frame().Buffer.CellAt(x, 0).Style; got != tc.want {
									t.Errorf("cell %d style = %+v; want %+v", x, got, tc.want)
								}
							}
							cursor := tc.want
							if tc.managed || tc.forced {
								cursor.Modifier |= cell.ModifierReverse
							}
							if got := term.Frame().Buffer.CellAt(5, 0).Style; got != cursor {
								t.Errorf("cursor style = %+v; want %+v", got, cursor)
							}
							if tc.name == "focused" {
								colours = append(colours, term.Frame().Buffer.CellAt(0, 0).Style.Fg)
							}
						})
					}
				})
			}
			if len(colours) != 2 || colours[0] == colours[1] {
				t.Fatal("dark and light frames must draw different focus colours")
			}
		})
	}
}

func TestTextEmptyFocusTheme(t *testing.T) {
	for _, widget := range themedTextWidgets() {
		for _, themed := range []bool{false, true} {
			for _, forced := range []bool{false, true} {
				term := testkit.NewTerminal(10, 2)
				term.Draw(func(f *terminal.Frame) {
					if themed {
						f.SetTheme(widgets.Theme{})
					}
					if !forced {
						f.FocusManager.SetFocused("text")
					}
					f.RenderWidget(widget.make(cell.Style{}, cell.Style{}, forced), term.Area())
				})
				for _, x := range []uint16{0, 7} {
					if got := term.Frame().Buffer.CellAt(x, 0).Style; got != (cell.Style{}) {
						t.Errorf("%s themed=%v forced=%v: cell %d style = %+v", widget.name, themed, forced, x, got)
					}
				}
				if got := term.Frame().Buffer.CellAt(5, 0).Style; got != (cell.Style{Modifier: cell.ModifierReverse}) {
					t.Errorf("%s themed=%v forced=%v: cursor style = %+v", widget.name, themed, forced, got)
				}
			}
		}
	}
}

func BenchmarkTextDrawFocusedTheme(b *testing.B) {
	for _, widget := range themedTextWidgets() {
		b.Run(widget.name, func(b *testing.B) {
			area := cell.Rect{Width: 20, Height: 3}
			buf := buffer.NewBuffer(area)
			theme := widgets.DarkTheme()
			ctx := cell.NewContext(area, theme.RoleStyle("surface"))
			ctx.ThemeStyle, ctx.FocusedID = theme.RoleStyle, "text"
			w := widget.make(cell.Style{}, cell.Style{}, false)
			w.Draw(ctx, buf)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				w.Draw(ctx, buf)
			}
		})
	}
}

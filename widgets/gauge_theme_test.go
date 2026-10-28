package widgets_test

import (
	"testing"

	"github.com/thebanri/limoni/core/buffer"
	"github.com/thebanri/limoni/core/cell"
	"github.com/thebanri/limoni/core/terminal"
	"github.com/thebanri/limoni/testkit"
	"github.com/thebanri/limoni/widgets"
)

func TestGaugeFrameTheme(t *testing.T) {
	custom := cell.Style{Fg: cell.NewColorRGB(240, 80, 120), Bg: cell.NewColorRGB(40, 50, 60)}
	for _, preset := range []struct {
		name  string
		theme widgets.Theme
	}{
		{"dark", widgets.DarkTheme()}, {"light", widgets.LightTheme()}, {"empty", widgets.Theme{}},
	} {
		t.Run(preset.name, func(t *testing.T) {
			surface := preset.theme.RoleStyle("surface")
			for _, tc := range []struct {
				name       string
				gauge      widgets.Gauge
				fill, base cell.Style
			}{
				{"default", widgets.Gauge{}, surface.Merge(preset.theme.RoleStyle("primary")), surface},
				{"explicit base", widgets.Gauge{Style: custom}, surface.Merge(custom), surface.Merge(custom)},
				{"explicit fill", widgets.Gauge{GaugeStyle: custom}, surface.Merge(custom), surface},
				{"modifier fill", widgets.Gauge{GaugeStyle: cell.Style{Modifier: cell.ModifierBold}}, surface.Bold(), surface},
			} {
				t.Run(tc.name, func(t *testing.T) {
					term := testkit.NewTerminal(8, 2)
					tc.gauge.Ratio, tc.gauge.Label = 0.5625, "ABCDEFGH"
					term.Draw(func(f *terminal.Frame) { f.SetTheme(preset.theme); f.RenderWidget(tc.gauge, term.Area()) })
					partial := tc.base
					partial.Fg = tc.fill.Fg
					label := partial.Reverse()
					for _, want := range []struct {
						x, y  uint16
						ch    rune
						style cell.Style
					}{
						{0, 0, '█', tc.fill}, {4, 0, '▌', partial}, {7, 0, ' ', tc.base},
						{0, 1, 'A', label}, {7, 1, 'H', tc.base},
					} {
						got := term.Frame().Buffer.CellAt(want.x, want.y)
						if got.Content != want.ch || got.Style != want.style {
							t.Errorf("cell(%d,%d) = %+v; want %q with %+v", want.x, want.y, got, want.ch, want.style)
						}
					}
				})
			}
		})
	}
}

func TestLineGaugeFrameTheme(t *testing.T) {
	custom := cell.Style{Fg: cell.NewColorRGB(240, 80, 120), Bg: cell.NewColorRGB(40, 50, 60)}
	for _, preset := range []struct {
		name  string
		theme widgets.Theme
	}{
		{"dark", widgets.DarkTheme()}, {"light", widgets.LightTheme()}, {"empty", widgets.Theme{}},
	} {
		t.Run(preset.name, func(t *testing.T) {
			surface := preset.theme.RoleStyle("surface")
			primary, muted := surface.Merge(preset.theme.RoleStyle("primary")), surface.Merge(preset.theme.RoleStyle("muted"))
			for _, tc := range []struct {
				name                  string
				gauge                 widgets.LineGauge
				label, fill, unfilled cell.Style
			}{
				{"default", widgets.LineGauge{}, surface, primary, muted},
				{"explicit label", widgets.LineGauge{Style: custom}, surface.Merge(custom), primary, muted},
				{"explicit filled", widgets.LineGauge{FilledStyle: custom}, surface, surface.Merge(custom), muted},
				{"explicit unfilled", widgets.LineGauge{UnfilledStyle: custom}, surface, primary, surface.Merge(custom)},
				{"modifier filled", widgets.LineGauge{FilledStyle: cell.Style{Modifier: cell.ModifierBold}}, surface, surface.Bold(), muted},
				{"modifier unfilled", widgets.LineGauge{UnfilledStyle: cell.Style{Modifier: cell.ModifierDim}}, surface, primary, surface.Merge(cell.Style{Modifier: cell.ModifierDim})},
			} {
				t.Run(tc.name, func(t *testing.T) {
					term := testkit.NewTerminal(10, 1)
					tc.gauge.Ratio, tc.gauge.Label = 0.5, "OK"
					term.Draw(func(f *terminal.Frame) { f.SetTheme(preset.theme); f.RenderWidget(tc.gauge, term.Area()) })
					for _, want := range []struct {
						x     uint16
						ch    rune
						style cell.Style
					}{{0, 'O', tc.label}, {3, '━', tc.fill}, {9, '─', tc.unfilled}} {
						got := term.Frame().Buffer.CellAt(want.x, 0)
						if got.Content != want.ch || got.Style != want.style {
							t.Errorf("cell(%d,0) = %+v; want %q with %+v", want.x, got, want.ch, want.style)
						}
					}
				})
			}
		})
	}
}

func BenchmarkGaugeDrawTheme(b *testing.B) {
	area := cell.Rect{Width: 80, Height: 1}
	buf := buffer.NewBuffer(area)
	theme := widgets.DarkTheme()
	ctx := cell.NewContext(area, theme.RoleStyle("surface"))
	ctx.ThemeStyle = theme.RoleStyle
	gauge := widgets.Gauge{Ratio: 0.42}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gauge.Draw(ctx, buf)
	}
}

func BenchmarkLineGaugeDrawTheme(b *testing.B) {
	area := cell.Rect{Width: 80, Height: 1}
	buf := buffer.NewBuffer(area)
	theme := widgets.DarkTheme()
	ctx := cell.NewContext(area, theme.RoleStyle("surface"))
	ctx.ThemeStyle = theme.RoleStyle
	gauge := widgets.LineGauge{Ratio: 0.42}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gauge.Draw(ctx, buf)
	}
}

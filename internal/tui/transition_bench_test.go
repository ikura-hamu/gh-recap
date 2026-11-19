package tui

import "testing"

func BenchmarkPageTransition(b *testing.B) {
	for _, size := range []struct {
		name          string
		width, height int
	}{{"120x28", 120, 28}, {"200x60", 200, 60}} {
		b.Run(size.name, func(b *testing.B) {
			m := readyModel()
			m.width, m.height = size.width, size.height
			m = update(m, pressKey("l"))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				m.transition.frame = i % transitionFrames
				_ = m.View()
			}
		})
	}
}

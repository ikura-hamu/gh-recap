package tui

import "math"

var dpgkPalette = [...]string{
	"#FF2BD6", "#9B5CFF", "#30C5FF", "#54FFE3", "#C6FF39", "#FFE34D", "#FF704D",
}

func dpgkColor(phase int) string {
	return dpgkPalette[phase%len(dpgkPalette)]
}

func (e backgroundEffect) dpgkCell(x, y int) (string, string) {
	hue := e.page + e.frame/2 + x/9 + y/4
	t := float64(e.frame) * 0.28
	cx, cy := float64(e.width)/2, float64(e.height)/2
	px, py := float64(x), float64(y)

	switch e.page {
	case pageOverview: // Crossing aurora ribbons.
		first := cy + cy*0.58*math.Sin(px*0.05+t)
		second := cy + cy*0.52*math.Sin(px*0.06-t+2)
		return dpgkGlow(min(math.Abs(py-first), math.Abs(py-second)), hue)
	case pageCalendar: // A beam lights up columns of the activity grid.
		scan := (x + e.width - e.frame*5%e.width) % e.width
		switch {
		case scan < 6:
			return dpgkGlow(float64(scan)*0.5, hue)
		case x%5 == 0 && y%2 == 0:
			return "▪", dpgkColor(hue + 2)
		}
	case pageMonths: // Opposing waveforms with bright peaks.
		first := cy*0.7 + cy*0.44*math.Sin(px*0.27-t*2)
		second := cy*1.3 + cy*0.44*math.Sin(px*0.19+t*2)
		distance := min(math.Abs(py-first), math.Abs(py-second))
		if distance < 0.75 {
			return "●", dpgkColor(hue)
		}
		if distance < 2 {
			return "·", dpgkColor(hue + 1)
		}
	case pageStreak: // Comets race along diagonal lanes.
		if (x+2*y)%8 < 2 {
			tail := (x + y*3 + e.frame*4) % 28
			switch {
			case tail < 3:
				return "◆", dpgkColor(hue)
			case tail < 10:
				return "╱", dpgkColor(hue + 1)
			case tail < 16:
				return "·", dpgkColor(hue + 2)
			}
		}
	case pageCommits: // Staggered columns of luminous rain.
		if x%5 == 0 {
			column, period := x/5, e.height+11
			head := (e.frame*(2+column%3) + column*13) % period
			tail := (head + period - y) % period
			switch {
			case tail == 0:
				return "✦", dpgkColor(hue)
			case tail < 6:
				return "│", dpgkColor(hue + 1)
			case tail < 11:
				return "·", dpgkColor(hue + 2)
			}
		}
	case pageCoauthors: // Two expanding rings light up their intersections.
		left := math.Hypot((px-cx*0.55)*0.5, py-cy)
		right := math.Hypot((px-cx*1.45)*0.5, py-cy)
		a := math.Mod(left+13-math.Mod(t*5, 13), 13)
		b := math.Mod(right+13-math.Mod(t*5+4, 13), 13)
		if a < 2 && b < 2 {
			return "✦", dpgkColor(hue + 3)
		}
		return dpgkGlow(min(a, b), hue)
	case pagePullRequests: // Chevron bands converge toward the center.
		distance := math.Abs(px-cx) + 2*math.Abs(py-cy)
		phase := (int(distance) + e.frame*3) % 22
		if phase < 3 {
			if px < cx {
				return "›", dpgkColor(hue)
			}
			return "‹", dpgkColor(hue)
		}
		if phase < 6 {
			return "░", dpgkColor(hue + 1)
		}
	case pagePopularPR: // Scattered stars brighten in a traveling spotlight.
		if (x+y*3)%11 == 0 {
			twinkle := (x*3 + y + e.frame*2) % 17
			switch {
			case twinkle < 4:
				return "✦", dpgkColor(hue)
			case twinkle < 8:
				return "◇", dpgkColor(hue + 1)
			default:
				return "·", dpgkColor(hue + 2)
			}
		}
	case pageReviews: // Mirrored horizontal lasers sweep toward each other.
		travel := max(1, e.height-1)
		beam := float64(travel) - math.Abs(float64(e.frame*2%(travel*2)-travel))
		distance := min(math.Abs(py-beam), math.Abs(py-(float64(travel)-beam)))
		if distance == 0 {
			return "━", dpgkColor(hue)
		}
		return dpgkGlow(distance+0.5, hue)
	case pageRepositories: // An equalizer skyline climbs from the bottom.
		if x%7 < 4 {
			top := e.height - 1 - int(cy*(0.9+0.65*math.Sin(float64(x/7)*1.6+t*2)))
			switch {
			case y == top:
				return "▄", dpgkColor(hue)
			case y > top && y%3 != 0:
				return "▓", dpgkColor(hue + (y-top)/3)
			}
		}
	case pageLanguages: // Streams of code symbols run in alternating directions.
		if y%3 == 1 {
			symbols := [...]string{"<", ">", "{", "}", "+", "*"}
			offset := x
			if y%2 == 0 {
				offset = e.width - 1 - x
			}
			offset += e.frame*3 + y*11
			switch offset % 6 {
			case 0:
				return symbols[(offset/6+y)%len(symbols)], dpgkColor(hue)
			case 1, 2:
				return "·", dpgkColor(hue + 2)
			}
		}
	case pageSummary: // Staggered fireworks throw sparks out from two centers.
		center, phase := cx*0.55, t
		if px >= cx {
			center, phase = cx*1.48, t+2.3
		}
		dx, dy := (px-center)*0.45, py-cy
		radius := math.Hypot(dx, dy)
		ray := math.Cos(math.Atan2(dy, dx)*12 - phase)
		burst := math.Mod(radius+18-math.Mod(phase*5+7, 18), 18)
		switch {
		case ray > 0.73 && burst < 2.8:
			return "✦", dpgkColor(hue)
		case ray > 0.55 && burst < 5:
			return "◆", dpgkColor(hue + 1)
		case ray > 0.45 && burst < 8:
			return "·", dpgkColor(hue + 2)
		}
	}
	return "", ""
}

func dpgkGlow(distance float64, hue int) (string, string) {
	switch {
	case distance < 0.8:
		return "▓", dpgkColor(hue)
	case distance < 1.7:
		return "▒", dpgkColor(hue + 1)
	case distance < 2.8:
		return "░", dpgkColor(hue + 2)
	default:
		return "", ""
	}
}

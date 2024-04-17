package color

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Color struct {
	focused string
	blurred string
}

func (c Color) Darken(factor float64) Color {
	var res Color

	d, err := darken(factor, c.focused)
	if err != nil {
		panic(err)
	}

	res.focused = d

	if c.blurred != c.focused {
		b, err := darken(factor, c.blurred)
		if err != nil {
			panic(err)
		}

		res.blurred = b
	} else {
		res.blurred = res.focused
	}

	return res
}

func (c Color) Brighten(percent float64) Color {
	var res Color

	b, err := brighten(percent, c.focused)
	if err != nil {
		panic(err)
	}

	res.focused = b

	if c.blurred != c.focused {
		b, err := brighten(percent, c.blurred)
		if err != nil {
			panic(err)
		}

		res.blurred = b
	} else {
		res.blurred = res.focused
	}

	return res
}

func (c *Color) Focused() lipgloss.Color {
	return lipgloss.Color(c.focused)
}

func (c *Color) Blurred() lipgloss.Color {
	return lipgloss.Color(c.blurred)
}

func Simple(color string) Color {
	return Focusable(color, color)
}

func Focusable(focused, blurred string) Color {
	c := Color{
		focused: focused,
	}

	if blurred != "" {
		c.blurred = blurred
	} else {
		c.blurred = c.focused
	}

	return c
}

func darken(factor float64, color string) (string, error) {
	if factor < 0 || factor > 1 {
		return "", fmt.Errorf("factor must be between 0 and 1")
	}

	if len(color) == 4 {
		color = expandHexColor(color)
	}

	// Convert hex string to integers
	r, err := strconv.ParseInt(color[1:3], 16, 64)
	if err != nil {
		return "", err
	}
	g, err := strconv.ParseInt(color[3:5], 16, 64)
	if err != nil {
		return "", err
	}
	b, err := strconv.ParseInt(color[5:7], 16, 64)
	if err != nil {
		return "", err
	}

	// Calculate the new darker color
	r = int64(float64(r) * (1 - factor))
	g = int64(float64(g) * (1 - factor))
	b = int64(float64(b) * (1 - factor))

	// Convert back to hex string
	return fmt.Sprintf("#%02x%02x%02x", r, g, b), nil
}

func brighten(factor float64, color string) (string, error) {
	if factor < 0 || factor > 1 {
		return "", fmt.Errorf("factor must be between 0 and 1")
	}

	if len(color) == 4 {
		color = expandHexColor(color)
	}

	// Convert hex string to integers
	r, err := strconv.ParseInt(color[1:3], 16, 64)
	if err != nil {
		return "", err
	}
	g, err := strconv.ParseInt(color[3:5], 16, 64)
	if err != nil {
		return "", err
	}
	b, err := strconv.ParseInt(color[5:7], 16, 64)
	if err != nil {
		return "", err
	}

	// Calculate the new brighter color
	r = int64(float64(255-r)*factor) + r
	g = int64(float64(255-g)*factor) + g
	b = int64(float64(255-b)*factor) + b

	// Ensure RGB values do not exceed 255
	if r > 255 {
		r = 255
	}
	if g > 255 {
		g = 255
	}
	if b > 255 {
		b = 255
	}

	// Convert back to hex string
	return fmt.Sprintf("#%02x%02x%02x", r, g, b), nil
}

func expandHexColor(shortHex string) string {
	// Repeat each character in the color code
	r := strings.Repeat(string(shortHex[1]), 2)
	g := strings.Repeat(string(shortHex[2]), 2)
	b := strings.Repeat(string(shortHex[3]), 2)

	return fmt.Sprintf("#%s%s%s", r, g, b)
}

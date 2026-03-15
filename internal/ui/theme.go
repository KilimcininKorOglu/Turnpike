package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// OLEDTheme implements a true black OLED-optimized dark theme
type OLEDTheme struct{}

var _ fyne.Theme = (*OLEDTheme)(nil)

// Color definitions for OLED dark theme
var (
	ColorPureBlack   = color.NRGBA{R: 0, G: 0, B: 0, A: 255}       // #000000
	ColorNearBlack   = color.NRGBA{R: 26, G: 26, B: 26, A: 255}    // #1A1A1A
	ColorDarker      = color.NRGBA{R: 13, G: 13, B: 13, A: 255}    // #0D0D0D
	ColorBorder      = color.NRGBA{R: 51, G: 51, B: 51, A: 255}    // #333333
	ColorWhite       = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // #FFFFFF
	ColorSecondary   = color.NRGBA{R: 204, G: 204, B: 204, A: 255} // #CCCCCC
	ColorDisabled    = color.NRGBA{R: 136, G: 136, B: 136, A: 255} // #888888
	ColorSuccess     = color.NRGBA{R: 0, G: 170, B: 0, A: 255}     // #00AA00
	ColorError       = color.NRGBA{R: 204, G: 0, B: 0, A: 255}     // #CC0000
	ColorWarning     = color.NRGBA{R: 255, G: 152, B: 0, A: 255}   // #FF9800
	ColorInfo        = color.NRGBA{R: 52, G: 152, B: 219, A: 255}  // #3498DB
	ColorTurquoise   = color.NRGBA{R: 26, G: 188, B: 156, A: 255}  // #1ABC9C
	ColorOrange      = color.NRGBA{R: 230, G: 126, B: 34, A: 255}  // #E67E22
	ColorDarkAmber   = color.NRGBA{R: 42, G: 31, B: 0, A: 255}     // #2A1F00
	ColorAmberBorder = color.NRGBA{R: 204, G: 136, B: 0, A: 255}   // #CC8800
)

// Color returns the color for the given theme color name
func (t *OLEDTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return ColorPureBlack
	case theme.ColorNameButton:
		return ColorNearBlack
	case theme.ColorNameDisabledButton:
		return ColorBorder
	case theme.ColorNameDisabled:
		return ColorDisabled
	case theme.ColorNameError:
		return ColorError
	case theme.ColorNameForeground:
		return ColorWhite
	case theme.ColorNameHover:
		return ColorBorder
	case theme.ColorNameInputBackground:
		return ColorNearBlack
	case theme.ColorNameInputBorder:
		return ColorBorder
	case theme.ColorNamePlaceHolder:
		return ColorDisabled
	case theme.ColorNamePressed:
		return ColorDarker
	case theme.ColorNamePrimary:
		return ColorSuccess
	case theme.ColorNameScrollBar:
		return ColorBorder
	case theme.ColorNameSeparator:
		return ColorBorder
	case theme.ColorNameShadow:
		return ColorPureBlack
	case theme.ColorNameSuccess:
		return ColorSuccess
	case theme.ColorNameWarning:
		return ColorWarning
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

// Font returns the font for the given text style
func (t *OLEDTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon returns the icon for the given icon name
func (t *OLEDTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns the size for the given size name
func (t *OLEDTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameText:
		return 13
	case theme.SizeNameInputBorder:
		return 1
	default:
		return theme.DefaultTheme().Size(name)
	}
}

package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestOLEDTheme_ImplementsInterface(t *testing.T) {
	var _ fyne.Theme = &OLEDTheme{}
}

func TestOLEDTheme_BackgroundIsPureBlack(t *testing.T) {
	th := &OLEDTheme{}
	bg := th.Color(theme.ColorNameBackground, theme.VariantDark)

	r, g, b, a := bg.RGBA()
	if r != 0 || g != 0 || b != 0 {
		t.Errorf("expected pure black background (#000000), got RGBA(%d,%d,%d,%d)", r, g, b, a)
	}
}

func TestOLEDTheme_ForegroundIsWhite(t *testing.T) {
	th := &OLEDTheme{}
	fg := th.Color(theme.ColorNameForeground, theme.VariantDark)

	expected := ColorWhite
	if fg != expected {
		t.Errorf("expected white foreground, got %v", fg)
	}
}

func TestOLEDTheme_PrimaryIsSuccess(t *testing.T) {
	th := &OLEDTheme{}
	primary := th.Color(theme.ColorNamePrimary, theme.VariantDark)

	if primary != ColorSuccess {
		t.Errorf("expected Success green for primary, got %v", primary)
	}
}

func TestOLEDTheme_ErrorColor(t *testing.T) {
	th := &OLEDTheme{}
	errColor := th.Color(theme.ColorNameError, theme.VariantDark)

	if errColor != ColorError {
		t.Errorf("expected error red, got %v", errColor)
	}
}

func TestOLEDTheme_InputBackgroundIsNearBlack(t *testing.T) {
	th := &OLEDTheme{}
	inputBg := th.Color(theme.ColorNameInputBackground, theme.VariantDark)

	if inputBg != ColorNearBlack {
		t.Errorf("expected near black input background, got %v", inputBg)
	}
}

func TestOLEDTheme_DisabledColor(t *testing.T) {
	th := &OLEDTheme{}
	disabled := th.Color(theme.ColorNameDisabled, theme.VariantDark)

	if disabled != ColorDisabled {
		t.Errorf("expected disabled gray, got %v", disabled)
	}
}

func TestOLEDTheme_SizeText(t *testing.T) {
	th := &OLEDTheme{}
	size := th.Size(theme.SizeNameText)

	if size != 13 {
		t.Errorf("expected text size 13, got %f", size)
	}
}

func TestOLEDTheme_SizePadding(t *testing.T) {
	th := &OLEDTheme{}
	size := th.Size(theme.SizeNamePadding)

	if size != 6 {
		t.Errorf("expected padding size 6, got %f", size)
	}
}

func TestOLEDTheme_FontNotNil(t *testing.T) {
	th := &OLEDTheme{}
	font := th.Font(fyne.TextStyle{})

	if font == nil {
		t.Error("expected non-nil font")
	}
}

func TestOLEDTheme_IconNotNil(t *testing.T) {
	th := &OLEDTheme{}
	icon := th.Icon(theme.IconNameInfo)

	if icon == nil {
		t.Error("expected non-nil icon")
	}
}

func TestColorConstants(t *testing.T) {
	tests := []struct {
		name     string
		color    color.NRGBA
		expected color.NRGBA
	}{
		{"PureBlack", ColorPureBlack, color.NRGBA{R: 0, G: 0, B: 0, A: 255}},
		{"NearBlack", ColorNearBlack, color.NRGBA{R: 26, G: 26, B: 26, A: 255}},
		{"Darker", ColorDarker, color.NRGBA{R: 13, G: 13, B: 13, A: 255}},
		{"Border", ColorBorder, color.NRGBA{R: 51, G: 51, B: 51, A: 255}},
		{"White", ColorWhite, color.NRGBA{R: 255, G: 255, B: 255, A: 255}},
		{"Secondary", ColorSecondary, color.NRGBA{R: 204, G: 204, B: 204, A: 255}},
		{"Disabled", ColorDisabled, color.NRGBA{R: 136, G: 136, B: 136, A: 255}},
		{"Success", ColorSuccess, color.NRGBA{R: 0, G: 170, B: 0, A: 255}},
		{"Error", ColorError, color.NRGBA{R: 204, G: 0, B: 0, A: 255}},
	}

	for _, tt := range tests {
		if tt.color != tt.expected {
			t.Errorf("Color %s: expected %v, got %v", tt.name, tt.expected, tt.color)
		}
	}
}

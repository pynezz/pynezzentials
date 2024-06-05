package ansi

import (
	"fmt"
	"strconv"
	"strings"
)

// Ansi colors
const (
	Reset  = "\033[0m" // Reset the escape sequence
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
)

type U8Color string

const (
	SetColor  U8Color = "38"
	SetNormal U8Color = "5"
	SetHex    U8Color = "2"
	SetBgHex          = "48"
)

var Bit8Color U8Color

// Ansi styles
const (
	Bold      = "\033[1m"
	Underline = "\033[4m"
	Inverse   = "\033[7m"
)

// Ansi 256 light colors
const (
	LightRed    = "\033[91m"
	LightGreen  = "\033[92m"
	LightYellow = "\033[93m"
	LightBlue   = "\033[94m"
	LightPurple = "\033[95m"
	LightCyan   = "\033[96m"
)

// Ansi 256 dark colors
const (
	DarkRed    = "\033[31m"
	DarkGreen  = "\033[32m"
	DarkYellow = "\033[33m"
	DarkBlue   = "\033[34m"
	DarkPurple = "\033[35m"
	DarkCyan   = "\033[36m"
)

// Background colors
const (
	BgRed    = "\033[41m"
	BgGreen  = "\033[42m"
	BgYellow = "\033[43m"
	BgBlue   = "\033[44m"
	BgPurple = "\033[45m"
	BgCyan   = "\033[46m"
	BgGray   = "\033[47m"
)

// Cursor movement
const (
	CursorUp    = "\033[A" // Move the cursor up
	CursorDown  = "\033[B" // Move the cursor down
	CursorRight = "\033[C" // Move the cursor right
	CursorLeft  = "\033[D" // Move the cursor left
)

// Terminal control
const (
	ClearScreen = "\033[2J" // Clear the screen
	ClearLine   = "\033[K"  // Clear the current line
	Backspace   = "\b"      // Backspace key
	Delete      = "\033[3~" // Delete key
	Enter       = "\r"      // Return carriage
	Tab         = "\t"      // Tab

	// Cursor positioning
	Home     = "\033[H"
	Position = "\033[%d;%dH"
	GetPos   = "\033[6n"

	// Save and restore cursor position
	SaveCursor    = "\033[s"
	RestoreCursor = "\033[u"

	// Hide and show cursor
	HideCursor = "\033[?25l"
	ShowCursor = "\033[?25h"

	// Overwrite the current line
	Overwrite = "\033[1A\033[2K"
)

// Rounded box characters
const (
	RoundedTopLeft     = "╭" // Top left corner of a rounded box
	RoundedTopRight    = "╮" // Top right corner of a rounded box
	RoundedBottomLeft  = "╰" // Bottom left corner of a rounded box
	RoundedBottomRight = "╯" // Bottom right corner of a rounded box
	RoundedHoriz       = "─" // Horizontal line of a rounded box
	RoundedVert        = "│" // Vertical line of a rounded box
)

// SetCursorPos sets the cursor position
func SetCursorPos(x, y int) {
	fmt.Printf(Position, x, y)
}

func GetCursorPos() string {
	return "\033[6n"
}

// SPrintRoundedTop returns a string with a rounded top
func SPrintRoundedTop(width int) string {
	rtop := ""
	for i := 0; i < width-2; i++ {
		rtop += RoundedHoriz
	}
	return fmt.Sprintf("%s%s%s", RoundedTopLeft, rtop, RoundedTopRight)
}

// SPrintRoundedBottom returns a string with a rounded bottom
func SPrintRoundedBottom(width int) string {
	rbottom := ""
	for i := 0; i < width-2; i++ {
		rbottom += RoundedHoriz
	}
	return fmt.Sprintf("%s%s%s", RoundedBottomLeft, rbottom, RoundedBottomRight)
}

// PrintRoundedTop prints a rounded top
func PrintRoundedTop(width int) {
	fmt.Print(RoundedTopLeft)
	for i := 0; i < width-2; i++ {
		fmt.Print(RoundedHoriz)
	}
	fmt.Print(RoundedTopRight)
}

// PrintRoundedBottom prints a rounded bottom
func PrintRoundedBottom(width int) {
	fmt.Print(RoundedBottomLeft)
	for i := 0; i < width-2; i++ {
		fmt.Print(RoundedHoriz)
	}
	fmt.Print(RoundedBottomRight)
}

// AddPadding adds padding to a string to make it a certain width
func AddPadding(content string, width int) string {
	padding := width - len(content)
	for i := 0; i < padding+1; i++ {
		content += " "
	}
	return content
}

// Get the width of a multiline string
func GetWidth(content string) int {
	width := 0
	tmpW := 0
	for _, c := range content {
		if c == '\n' {
			if width < tmpW {
				width = tmpW
			}
			tmpW = 0
		} else {
			tmpW++
		}
	}
	if width == 0 {
		width = tmpW
	}
	return width
}

// FormatRoundedBox formats a string into a rounded box with proper padding
func FormatRoundedBox(content string) string {
	tmpW := 0 // Temporary width
	w := 0    // Actual final width
	result := ""

	lines := []string{}

	for i, c := range content {
		if c == '\n' {
			if w < tmpW {
				w = tmpW
			}
			// result += fmt.Sprintf("%s %s %s\n", RoundedVert, content[i-tmpW:i], RoundedVert) // Should be │ content │
			lines = append(lines, fmt.Sprintf("%s %s ", RoundedVert, content[i-tmpW:i])) // Add to the line slice

			tmpW = 0
		} else {
			tmpW++
		}
	}

	if w == 0 {
		w = tmpW // If there are no newlines
	}

	finres := ""

	w += 4 // Add 4 for the corners + padding

	for _, l := range lines {
		finres += AddPadding(l, w) + RoundedVert + "\n"
	}

	result = SPrintRoundedTop(w) + "\n" + finres + SPrintRoundedBottom(w)

	return result
}

// PrintSuccess prints a success message to the console
func PrintSuccess(msg string) {
	fmt.Printf("%s[+]%s %s\n", Green, Reset, msg)
}

// PrintError prints an error message to the console
func PrintError(msg string) {
	fmt.Printf("%s[!]%s %s\n", Red, Reset, msg)
}

// PrintErrorf prints a formatted error message to the console
func PrintErrorf(format string, a ...interface{}) {
	fmt.Printf("%s[!]%s %s\n", Red, Reset, fmt.Sprintf(format, a...))
}

// ColorF returns a formatted string with a color
func ColorF(color, format string, a ...interface{}) string {
	return fmt.Sprintf("%s%s%s", color, fmt.Sprintf(format, a...), Reset)
}

// ErrorF returns a formatted error message
func Errorf(format string, a ...interface{}) error {
	return fmt.Errorf("%s[!]%s %s", Red, Reset, fmt.Sprintf(format, a...))
}

// ItalicF returns a formatted string with italic style
func ItalicF(format string, a ...interface{}) string {
	return fmt.Sprintf("%s%s%s", "\033[3m", fmt.Sprintf(format, a...), Reset)
}

// PrintInfo prints an info message to the console
func PrintInfo(msg string) {
	fmt.Printf("%s[i]%s %s\n", Cyan, Reset, msg)
}

// PrintWarning prints a warning message to the console
func PrintWarning(msg string) {
	fmt.Printf("%s⚠️%s %s\n", Yellow, Reset, msg)
}

// PrintDebug prints a debug message to the console
func PrintDebug(msg string) {
	fmt.Printf("%s[DEBUG]%s %s\n", Gray, Reset, msg)
}

// PrintBold prints a bold message to the console
func PrintBold(msg string) {
	fmt.Printf("%s%s%s\n", Bold, msg, Reset)
}

// PrintItalic prints an italic message to the console
func PrintItalic(msg string) {
	fmt.Printf("%s%s%s\n", "\033[3m", msg, Reset)
}

// PrintUnderline prints an underlined message to the console
func PrintUnderline(msg string) {
	fmt.Printf("%s%s%s\n", Underline, msg, Reset)
}

// PrintInverse prints an inverted message to the console
func PrintInverse(msg string) {
	fmt.Printf("%s%s%s\n", Inverse, msg, Reset)
}

// PrintColor prints a colored message to the console
func PrintColor(color, msg string) {
	fmt.Printf("%s%s%s\n", color, msg, Reset)
}

// PrintColorf prints a colored formatted message to the console
func PrintColorf(color, format string, a ...interface{}) {
	fmt.Printf("%s%s%s\n", color, fmt.Sprintf(format, a...), Reset)
}

// PrintColorBold prints a colored bold message to the console
func PrintColorBold(color, msg string) {
	fmt.Printf("%s%s%s\n", color+Bold, msg, Reset)
}

// PrintColorUnderline prints a colored underlined message to the console
func PrintColorUnderline(color, msg string) {
	fmt.Printf("%s%s%s\n", color+Underline, msg, Reset)
}

// PrintColorAndBg prints a colored message with a background to the console
func PrintColorAndBg(color, bg, msg string) {
	fmt.Printf("%s%s%s\n", color+bg, msg, Reset)
}

// PrintColorAndBgBold prints a colored bold message with a background to the console
func PrintColorAndBgBold(color, bg, msg string) {
	fmt.Printf("%s%s%s\n", color+bg+Bold, msg, Reset)
}

func (s U8Color) Color256(color U8Color, msg string) string {
	return fmt.Sprintf("\033[%s;5;%sm%s%s", SetColor, string(color), msg, Reset)
}

// HexToRGB converts a hex color string to an RGB string
//
// Example:
//
//	U8color.HexColor256("#3498db") or U8Color.HexToColor("3498d")
//
// HexToRGB converts a hex color string to an RGB string
func HexToRGB(hex string) (int, int, int, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color")
	}

	r, err := strconv.ParseInt(hex[0:2], 16, 0)
	if err != nil {
		return 0, 0, 0, err
	}
	g, err := strconv.ParseInt(hex[2:4], 16, 0)
	if err != nil {
		return 0, 0, 0, err
	}
	b, err := strconv.ParseInt(hex[4:6], 16, 0)
	if err != nil {
		return 0, 0, 0, err
	}

	return int(r), int(g), int(b), nil
}

// RGBToColor256 converts RGB values to the nearest xterm 256 color code
func RGBToColor256(r, g, b int) int {
	return 16 + 36*(r/51) + 6*(g/51) + (b / 51)
}

// HexColor256 converts RGB values to an ANSI 256 color escape sequence and applies it to a message
func HexColor256(r, g, b int, msg string) string {
	color := RGBToColor256(r, g, b)
	return fmt.Sprintf("\033[%s;%s;%dm%s%s", SetColor, SetNormal, color, msg, Reset)
}

// SprintHexf returns a string with a hex color
func SprintHexf(hex string, msg string) (string, error) {
	r, g, b, err := HexToRGB(hex)
	if err != nil {
		return "", err
	}
	color := RGBToColor256(r, g, b)
	return fmt.Sprintf("\033[%s;%s;%dm%s%s", SetColor, SetNormal, color, msg, Reset), nil
}

// HexToBg converts a hex color to a background color
func HexToBg(hex string) (string, error) {
	r, g, b, err := HexToRGB(hex)
	if err != nil {
		return "", err
	}
	color := RGBToColor256(r, g, b)
	return fmt.Sprintf("\033[%s;%s;%dm", SetBgHex, SetNormal, color), nil
}

// HexBgAndFg returns a string with a background and foreground color
func HexBgAndFg(hexFg, hexBg, msg string) (string, error) {
	rFg, gFg, bFg, err := HexToRGB(hexFg)
	if err != nil {
		return "", err
	}
	colorFg := RGBToColor256(rFg, gFg, bFg)

	rBg, gBg, bBg, err := HexToRGB(hexBg)
	if err != nil {
		return "", err
	}
	colorBg := RGBToColor256(rBg, gBg, bBg)

	return fmt.Sprintf("\033[%s;%s;%dm\033[%s;%s;%dm%s%s", SetColor, SetNormal, colorFg, SetBgHex, SetNormal, colorBg, msg, Reset), nil
}

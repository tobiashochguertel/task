package transparent

import (
	"bytes"
	"os"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/fatih/color"
)

// syntaxHighlight applies chroma syntax highlighting to JSON content.
// Returns the original content unchanged when color output is disabled.
// When showWhitespaces is true, whitespace visibility markers (·, →, ↵) are
// temporarily reversed before highlighting and re-applied afterwards.
// ANSI escape codes do not contain space or tab characters, so the
// re-application is safe.
func syntaxHighlight(content string, showWhitespaces bool) string {
	if color.NoColor || os.Getenv("NO_COLOR") != "" {
		return content
	}

	toHighlight := content
	if showWhitespaces {
		toHighlight = undoWhitespaceVisible(content)
	}

	// Only highlight content that looks like JSON
	trimmed := strings.TrimSpace(toHighlight)
	if len(trimmed) == 0 || (trimmed[0] != '{' && trimmed[0] != '[') {
		return content
	}

	highlighted := chromaHighlight(toHighlight)
	if highlighted == "" {
		return content
	}

	if showWhitespaces {
		highlighted = reapplyWhitespaceVisible(highlighted)
	}

	return highlighted
}

// undoWhitespaceVisible reverses the makeWhitespaceVisible transformation
// so that content can be properly lexed by chroma.
func undoWhitespaceVisible(s string) string {
	s = strings.ReplaceAll(s, "\u2190\u21b5\n", "\r\n")
	s = strings.ReplaceAll(s, "\u21b5\n", "\n")
	s = strings.ReplaceAll(s, "\u2190", "\r")
	s = strings.ReplaceAll(s, "\u00b7", " ")
	s = strings.ReplaceAll(s, "\u2192", "\t")
	s = strings.ReplaceAll(s, "[ESC]", "")
	return s
}

// reapplyWhitespaceVisible re-applies whitespace visibility markers after
// syntax highlighting. ANSI escape sequences do not contain literal space
// or tab characters, so the replacement is safe.
func reapplyWhitespaceVisible(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\u2190\u21b5\n")
	s = strings.ReplaceAll(s, "\n", "\u21b5\n")
	s = strings.ReplaceAll(s, "\r", "\u2190")
	s = strings.ReplaceAll(s, " ", "\u00b7")
	s = strings.ReplaceAll(s, "\t", "\u2192")
	return s
}

// chromaHighlight tokenises content as JSON and formats it with 256-color
// ANSI terminal codes using the Monokai style.
func chromaHighlight(content string) string {
	lexer := lexers.Get("json")
	if lexer == nil {
		return ""
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		return ""
	}

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return ""
	}

	result := strings.TrimRight(buf.String(), "\n")
	// Ensure ANSI reset at end to prevent color bleeding
	if !strings.HasSuffix(result, ansiReset) {
		result += ansiReset
	}
	return result
}

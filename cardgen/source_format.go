package cardgen

import (
	"fmt"
	"go/format"
	"strings"
)

func formatGeneratedSource(source string) (string, error) {
	formatted, err := format.Source([]byte(source))
	if err != nil {
		return "", fmt.Errorf("formatting generated source: %w", err)
	}
	return string(formatted), nil
}

func writeRawTextField(b *strings.Builder, indent, field, text string) {
	if strings.ContainsRune(text, '`') {
		_, _ = fmt.Fprintf(b, "%s%s: %q,\n", indent, field, text)
		return
	}
	_, _ = fmt.Fprintf(b, "%s%s: `\n", indent, field)
	for line := range strings.SplitSeq(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		_, _ = fmt.Fprintf(b, "%s\t%s\n", indent, line)
	}
	_, _ = fmt.Fprintf(b, "%s`,\n", indent)
}

func indentContinuation(literal, indent string) string {
	return strings.ReplaceAll(literal, "\n", "\n"+indent)
}

func layoutToLiteral(layout string) string {
	switch layout {
	case "transform":
		return "game.LayoutTransform"
	case "modal_dfc":
		return "game.LayoutModalDFC"
	case "meld":
		return "game.LayoutMeld"
	case "double_faced_token":
		return "game.LayoutDoubleFacedToken"
	case "reversible_card":
		return "game.LayoutReversibleCard"
	case "adventure":
		return "game.LayoutAdventure"
	case "split":
		return "game.LayoutSplit"
	case "prepare":
		return "game.LayoutPrepare"
	default:
		return ""
	}
}

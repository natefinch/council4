package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// splitGrantKeywordItems splits a keyword-grant body into its top-level granted
// keyword items. It mirrors the Oxford-comma and "and" splitting used for simple
// keyword lists but keeps a "protection from <...>" phrase intact, including any
// "and from <...>" continuation that names additional protected predicates
// ("protection from black and from red", "protection from artifacts and from
// the color of your choice"). Without this, the bare "from <...>" tail would be
// split off into its own item and fail the grantable-keyword check.
func splitGrantKeywordItems(text string) []string {
	text = strings.ReplaceAll(strings.ToLower(text), ", and ", ", ")
	text = strings.ReplaceAll(text, " and ", ", ")
	var items []string
	for part := range strings.SplitSeq(text, ", ") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "from ") &&
			len(items) > 0 &&
			(strings.HasPrefix(items[len(items)-1], "protection from") ||
				strings.HasPrefix(items[len(items)-1], "hexproof from")) {
			items[len(items)-1] += " and " + part
			continue
		}
		items = append(items, part)
	}
	return items
}

// grantableProtectionPhrase reports whether phrase is a complete "protection
// from <...>" grant body the executable backend can express. It re-lexes the
// phrase and runs the canonical protection-parameter parser, so the exactness
// gate accepts exactly the protection variants the keyword parser structures
// (a color list, the each-color/everything/monocolored/multicolored/chosen-color
// quantifiers, a card-type list, or a creature/land subtype list) and the
// lowering reduces to a static protection mechanic. It returns false for any
// protection phrase the parser cannot fully structure (a dynamic or
// predicate-qualified protection such as "protection from the colors of
// permanents you control"), which keeps those grants fail-closed.
func grantableProtectionPhrase(phrase string) bool {
	if !strings.HasPrefix(strings.ToLower(phrase), "protection from ") {
		return false
	}
	tokens, diagnostics := lexAll(phrase)
	if len(diagnostics) != 0 {
		return false
	}
	atoms := collectAtoms(tokens, nil, nil, "", false)
	fromIndex := slices.IndexFunc(tokens, func(token shared.Token) bool {
		return equalWord(token, "from")
	})
	if fromIndex < 0 {
		return false
	}
	parameter, end := parseProtectionKeywordParameter(tokens, fromIndex, atoms)
	if end <= fromIndex {
		return false
	}
	for _, token := range tokens[end:] {
		if token.Kind != shared.EOF {
			return false
		}
	}
	return runtimeExpressibleProtection(parameter.Protection())
}

// grantableHexproofFromPhrase reports whether phrase is a complete "hexproof
// from <colors>" grant body the executable backend can express. It mirrors
// grantableProtectionPhrase but uses the hexproof color-list parameter parser,
// so the exactness gate accepts exactly the color-qualified hexproof the keyword
// parser structures and the lowering reduces to a HexproofFromKeyword grant. It
// returns false for any hexproof phrase the parser cannot fully structure.
func grantableHexproofFromPhrase(phrase string) bool {
	if !strings.HasPrefix(strings.ToLower(phrase), "hexproof from ") {
		return false
	}
	tokens, diagnostics := lexAll(phrase)
	if len(diagnostics) != 0 {
		return false
	}
	atoms := collectAtoms(tokens, nil, nil, "", false)
	fromIndex := slices.IndexFunc(tokens, func(token shared.Token) bool {
		return equalWord(token, "from")
	})
	if fromIndex < 0 {
		return false
	}
	parameter, end := parseHexproofKeywordParameter(tokens, fromIndex, atoms)
	if end <= fromIndex {
		return false
	}
	for _, token := range tokens[end:] {
		if token.Kind != shared.EOF {
			return false
		}
	}
	return len(parameter.Protection().FromColors) != 0
}

// protected predicate the lowering reduces to a static protection mechanic. It
// mirrors the cases staticAbilityFromProtectionKeyword handles, so the parser
// only marks a protection grant exact when the executable backend can build it.
func runtimeExpressibleProtection(protection ProtectionParameter) bool {
	return protection.Everything ||
		protection.EachColor ||
		protection.ChosenColor ||
		protection.Multicolored ||
		protection.Monocolored ||
		len(protection.FromColors) != 0 ||
		len(protection.FromTypes) != 0 ||
		len(protection.FromSubtypes) != 0
}

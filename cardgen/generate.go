package cardgen

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type generatedCardFields struct {
	Name       string
	Layout     string
	ManaCost   string
	ManaValue  int
	TypeLine   string
	OracleText string
	Colors     []string

	ColorIdentity []string

	Power     *string
	Toughness *string
	Loyalty   *string
	Defense   *string

	EntersTapped bool
}

// GenerateCardSource generates a Go source file for a CardDef from Scryfall data.
// The file belongs to the given package name (e.g., "l" for the l/ subdirectory).
func GenerateCardSource(card *ScryfallCard, pkgName string) (string, error) {
	var b strings.Builder

	root := rootFields(card)
	faces := generatedFaces(card)
	if card.Layout == "reversible_card" && len(card.CardFaces) > 0 {
		faces = facesFromAllCardFaces(card)
	}
	needsMana := fieldsNeedMana(root) || anyFaceNeedsMana(faces)
	needsOpt := fieldsNeedOpt(root) || anyFaceNeedsOpt(faces)

	b.WriteString(fmt.Sprintf("package %s\n\n", pkgName))
	b.WriteString("import (\n")
	if needsOpt {
		b.WriteString("\t\"github.com/natefinch/council4/opt\"\n")
	}
	b.WriteString("\t\"github.com/natefinch/council4/mtg/game\"\n")
	if needsMana {
		b.WriteString("\t\"github.com/natefinch/council4/mtg/game/mana\"\n")
	}
	b.WriteString(")\n\n")

	if card.Layout == "reversible_card" && len(card.CardFaces) > 0 {
		for i, face := range facesFromAllCardFaces(card) {
			if i > 0 {
				b.WriteString("\n")
			}
			writeSingleFaceComment(&b, face)
			if err := writeCardDef(&b, face, card.Layout, nil); err != nil {
				return "", err
			}
		}
		return b.String(), nil
	}

	writeCardComment(&b, card, root, faces)
	if err := writeCardDef(&b, root, card.Layout, faces); err != nil {
		return "", err
	}

	return b.String(), nil
}

func rootFields(card *ScryfallCard) generatedCardFields {
	if len(card.CardFaces) > 0 && faceLayoutUsesFrontAsRoot(card.Layout) {
		root := fieldsFromFace(card.CardFaces[0])
		root.Layout = card.Layout
		root.ColorIdentity = append([]string(nil), card.ColorIdentity...)
		return root
	}
	return generatedCardFields{
		Name:          card.Name,
		Layout:        card.Layout,
		ManaCost:      card.ManaCost,
		ManaValue:     int(card.CMC),
		TypeLine:      card.TypeLine,
		OracleText:    card.OracleText,
		Colors:        append([]string(nil), card.Colors...),
		ColorIdentity: append([]string(nil), card.ColorIdentity...),
		Power:         card.Power,
		Toughness:     card.Toughness,
		Loyalty:       card.Loyalty,
		Defense:       card.Defense,
	}
}

func fieldsFromFace(face ScryfallCardFace) generatedCardFields {
	return generatedCardFields{
		Name:         face.Name,
		ManaCost:     face.ManaCost,
		ManaValue:    ManaValueFromCost(face.ManaCost),
		TypeLine:     face.TypeLine,
		OracleText:   face.OracleText,
		Colors:       append([]string(nil), face.Colors...),
		Power:        face.Power,
		Toughness:    face.Toughness,
		Loyalty:      face.Loyalty,
		Defense:      face.Defense,
		EntersTapped: oracleMeansEntersTapped(face.OracleText),
	}
}

func generatedFaces(card *ScryfallCard) []generatedCardFields {
	if len(card.CardFaces) == 0 || !layoutEmitsFaces(card.Layout) {
		return nil
	}
	faces := make([]generatedCardFields, 0, len(card.CardFaces))
	for _, face := range card.CardFaces {
		faces = append(faces, fieldsFromFace(face))
	}
	return faces
}

func facesFromAllCardFaces(card *ScryfallCard) []generatedCardFields {
	faces := make([]generatedCardFields, 0, len(card.CardFaces))
	for _, face := range card.CardFaces {
		fields := fieldsFromFace(face)
		fields.Layout = card.Layout
		fields.ColorIdentity = append([]string(nil), card.ColorIdentity...)
		faces = append(faces, fields)
	}
	return faces
}

func faceLayoutUsesFrontAsRoot(layout string) bool {
	switch layout {
	case "transform", "modal_dfc", "meld", "double_faced_token", "reversible_card":
		return true
	default:
		return false
	}
}

func layoutEmitsFaces(layout string) bool {
	switch layout {
	case "transform", "modal_dfc", "double_faced_token":
		return true
	default:
		return false
	}
}

func fieldsNeedMana(fields generatedCardFields) bool {
	return fields.ManaCost != "" || len(fields.Colors) > 0 || len(fields.ColorIdentity) > 0
}

func fieldsNeedOpt(fields generatedCardFields) bool {
	return fields.ManaCost != "" || fields.Power != nil || fields.Toughness != nil || fields.Loyalty != nil || fields.Defense != nil
}

func anyFaceNeedsMana(faces []generatedCardFields) bool {
	for _, face := range faces {
		if fieldsNeedMana(face) {
			return true
		}
	}
	return false
}

func anyFaceNeedsOpt(faces []generatedCardFields) bool {
	for _, face := range faces {
		if fieldsNeedOpt(face) {
			return true
		}
	}
	return false
}

func writeCardComment(b *strings.Builder, card *ScryfallCard, root generatedCardFields, faces []generatedCardFields) {
	b.WriteString(fmt.Sprintf("// %s\n", card.Name))
	b.WriteString("//\n")
	b.WriteString(fmt.Sprintf("// Type: %s\n", card.TypeLine))
	if card.ManaCost != "" {
		b.WriteString(fmt.Sprintf("// Cost: %s\n", card.ManaCost))
	}
	if len(faces) > 0 {
		for _, face := range faces {
			b.WriteString(fmt.Sprintf("// Face: %s — %s", face.Name, face.TypeLine))
			if face.ManaCost != "" {
				b.WriteString(fmt.Sprintf(" (%s)", face.ManaCost))
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("//\n")
	b.WriteString("// Oracle text:\n")
	oracle := card.OracleText
	if oracle == "" {
		oracle = root.OracleText
	}
	if oracle != "" {
		for _, line := range strings.Split(oracle, "\n") {
			b.WriteString(fmt.Sprintf("//   %s\n", line))
		}
	} else {
		for i, face := range faces {
			if i > 0 {
				b.WriteString("//   ---\n")
			}
			b.WriteString(fmt.Sprintf("//   %s\n", face.Name))
			for _, line := range strings.Split(face.OracleText, "\n") {
				b.WriteString(fmt.Sprintf("//   %s\n", line))
			}
		}
	}
	b.WriteString("//\n")
	b.WriteString("// TODO: Fill in Abilities from oracle text.\n")
}

func writeSingleFaceComment(b *strings.Builder, fields generatedCardFields) {
	b.WriteString(fmt.Sprintf("// %s\n", fields.Name))
	b.WriteString("//\n")
	b.WriteString(fmt.Sprintf("// Type: %s\n", fields.TypeLine))
	if fields.ManaCost != "" {
		b.WriteString(fmt.Sprintf("// Cost: %s\n", fields.ManaCost))
	}
	b.WriteString("//\n")
	b.WriteString("// Oracle text:\n")
	for _, line := range strings.Split(fields.OracleText, "\n") {
		b.WriteString(fmt.Sprintf("//   %s\n", line))
	}
	b.WriteString("//\n")
	b.WriteString("// TODO: Fill in Abilities from oracle text.\n")
}

func writeCardDef(b *strings.Builder, root generatedCardFields, layout string, faces []generatedCardFields) error {
	varName := CardNameToVarName(root.Name)
	b.WriteString(fmt.Sprintf("\nvar %s = &game.CardDef{\n", varName))
	if err := writeFields(b, root, "\t", true, true); err != nil {
		return err
	}
	if layoutLiteral := layoutToLiteral(layout); layoutLiteral != "" {
		b.WriteString(fmt.Sprintf("\tLayout: %s,\n", layoutLiteral))
	}
	if len(faces) > 0 {
		b.WriteString("\tFaces: []game.CardFace{\n")
		for _, face := range faces {
			b.WriteString("\t\t{\n")
			if err := writeFields(b, face, "\t\t\t", true, false); err != nil {
				return err
			}
			b.WriteString("\t\t},\n")
		}
		b.WriteString("\t},\n")
	}
	b.WriteString("}\n")
	return nil
}

func writeFields(b *strings.Builder, fields generatedCardFields, indent string, includeName bool, includeColorIdentity bool) error {
	if includeName {
		b.WriteString(fmt.Sprintf("%sName: %q,\n", indent, fields.Name))
	}
	if fields.ManaCost != "" {
		costLiteral, err := ParseManaCostLiteral(fields.ManaCost)
		if err != nil {
			return fmt.Errorf("parsing mana cost for %s: %w", fields.Name, err)
		}
		if costLiteral != "" {
			b.WriteString(fmt.Sprintf("%sManaCost: opt.Val(%s),\n", indent, indentContinuation(costLiteral, indent)))
		}
	}
	b.WriteString(fmt.Sprintf("%sManaValue: %d,\n", indent, fields.ManaValue))
	if len(fields.Colors) > 0 {
		b.WriteString(fmt.Sprintf("%sColors: []mana.Color{%s},\n", indent, colorLiterals(fields.Colors)))
	}
	if includeColorIdentity && len(fields.ColorIdentity) > 0 {
		b.WriteString(fmt.Sprintf("%sColorIdentity: mana.NewColorIdentity(%s),\n", indent, colorLiterals(fields.ColorIdentity)))
	}
	parsed := ParseTypeLine(fields.TypeLine)
	if len(parsed.Supertypes) > 0 {
		var literals []string
		for _, st := range parsed.Supertypes {
			literals = append(literals, SupertypeToLiteral(st))
		}
		b.WriteString(fmt.Sprintf("%sSupertypes: []game.Supertype{%s},\n", indent, strings.Join(literals, ", ")))
	}
	if len(parsed.Types) > 0 {
		var literals []string
		for _, t := range parsed.Types {
			literals = append(literals, CardTypeToLiteral(t))
		}
		b.WriteString(fmt.Sprintf("%sTypes: []game.CardType{%s},\n", indent, strings.Join(literals, ", ")))
	}
	if len(parsed.Subtypes) > 0 {
		var literals []string
		for _, s := range parsed.Subtypes {
			literals = append(literals, SubtypeToLiteral(s, parsed.Types))
		}
		b.WriteString(fmt.Sprintf("%sSubtypes: []string{%s},\n", indent, strings.Join(literals, ", ")))
	}
	if fields.Power != nil {
		b.WriteString(fmt.Sprintf("%sPower: opt.Val(%s),\n", indent, ptLiteral(*fields.Power)))
	}
	if fields.Toughness != nil {
		b.WriteString(fmt.Sprintf("%sToughness: opt.Val(%s),\n", indent, ptLiteral(*fields.Toughness)))
	}
	if fields.Loyalty != nil {
		if n, err := strconv.Atoi(*fields.Loyalty); err == nil {
			b.WriteString(fmt.Sprintf("%sLoyalty: opt.Val(%d),\n", indent, n))
		}
	}
	if fields.Defense != nil {
		if n, err := strconv.Atoi(*fields.Defense); err == nil {
			b.WriteString(fmt.Sprintf("%sDefense: opt.Val(%d),\n", indent, n))
		}
	}
	if fields.EntersTapped {
		b.WriteString(fmt.Sprintf("%sEntersTapped: true,\n", indent))
	}
	if fields.OracleText != "" {
		b.WriteString(fmt.Sprintf("%sOracleText: %q,\n", indent, fields.OracleText))
	}
	b.WriteString(fmt.Sprintf("%s// Abilities: filled in by LLM from oracle text.\n", indent))
	b.WriteString(fmt.Sprintf("%sAbilities: []game.AbilityDef{},\n", indent))
	return nil
}

func indentContinuation(literal string, indent string) string {
	return strings.ReplaceAll(literal, "\n", "\n"+indent)
}

func colorLiterals(colors []string) string {
	var literals []string
	for _, c := range colors {
		literals = append(literals, ColorToLiteral(c))
	}
	return strings.Join(literals, ", ")
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
	default:
		return ""
	}
}

// ManaValueFromCost returns the mana value of a mana-cost string.
func ManaValueFromCost(cost string) int {
	matches := manaSymbolRe.FindAllStringSubmatch(cost, -1)
	total := 0
	for _, match := range matches {
		symbol := match[1]
		total += manaSymbolValue(symbol)
	}
	return total
}

func manaSymbolValue(symbol string) int {
	if n, err := strconv.Atoi(symbol); err == nil {
		return n
	}
	if strings.HasPrefix(symbol, "X") {
		return 0
	}
	if strings.HasPrefix(symbol, "2/") {
		return 2
	}
	return 1
}

func oracleMeansEntersTapped(oracle string) bool {
	normalized := strings.ToLower(oracle)
	return strings.Contains(normalized, "enters the battlefield tapped.") ||
		strings.Contains(normalized, "enters tapped.")
}

func ptLiteral(val string) string {
	if val == "*" {
		return "game.PT{IsStar: true}"
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Sprintf("game.PT{} /* unparseable: %q */", val)
	}
	return fmt.Sprintf("game.PT{Value: %d}", n)
}

// CardNameToVarName converts a card name to a Go exported variable name.
// e.g., "Lightning Bolt" -> "LightningBolt", "Sol Ring" -> "SolRing"
func CardNameToVarName(name string) string {
	var b strings.Builder
	capitalize := true
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			capitalize = true
			continue
		}
		if capitalize {
			b.WriteRune(unicode.ToUpper(r))
			capitalize = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// CardNameToFileName converts a card name to a snake_case file name.
// e.g., "Lightning Bolt" -> "lightning_bolt", "Sol Ring" -> "sol_ring"
func CardNameToFileName(name string) string {
	var b strings.Builder
	prevWasUnderscore := false
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			if !prevWasUnderscore && b.Len() > 0 {
				b.WriteRune('_')
				prevWasUnderscore = true
			}
			continue
		}
		b.WriteRune(unicode.ToLower(r))
		prevWasUnderscore = false
	}
	result := b.String()
	return strings.TrimSuffix(result, "_")
}

// CardNameToPackageLetter returns the lowercase first letter of the card name.
func CardNameToPackageLetter(name string) string {
	for _, r := range name {
		if unicode.IsLetter(r) {
			return string(unicode.ToLower(r))
		}
	}
	return "other"
}

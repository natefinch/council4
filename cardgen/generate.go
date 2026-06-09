package cardgen

import (
	"fmt"
	"go/format"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type generatedCardFields struct {
	Name       string
	Layout     string
	ManaCost   string
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

// GenerateCardSource generates a canonically formatted CardDef source file from
// Scryfall data. The file belongs to the given package name (e.g., "l"). It
// always emits a partial definition with a TODO; executable cards use
// GenerateExecutableCardSource instead.
func GenerateCardSource(card *ScryfallCard, pkgName string) (string, error) {
	return genCardSource(card, pkgName)
}

func genCardSource(
	card *ScryfallCard,
	pkgName string,
) (string, error) {
	var b strings.Builder

	root := rootFields(card)
	faces := generatedFaces(card)
	if card.Layout == "reversible_card" && len(card.CardFaces) > 0 {
		faces = facesFromAllCardFaces(card)
	}
	needsCost := fieldsNeedCost(root) || anyFaceNeedsCost(faces)
	needsColor := fieldsNeedColor(root) || anyFaceNeedsColor(faces)
	needsMana := fieldsNeedMana(root) || anyFaceNeedsMana(faces)
	needsOpt := fieldsNeedOpt(root) || anyFaceNeedsOpt(faces) || len(faces) > 0
	needsTypes := fieldsNeedTypes(root) || slices.ContainsFunc(faces, fieldsNeedTypes)

	_, _ = fmt.Fprintf(&b, "package %s\n\n", pkgName)
	_, _ = b.WriteString("import (\n")
	_, _ = b.WriteString("\t\"github.com/natefinch/council4/mtg/game\"\n")
	if needsColor {
		_, _ = b.WriteString("\t\"github.com/natefinch/council4/mtg/game/color\"\n")
	}
	if needsCost {
		_, _ = b.WriteString("\t\"github.com/natefinch/council4/mtg/game/cost\"\n")
	}
	if needsTypes {
		_, _ = b.WriteString("\t\"github.com/natefinch/council4/mtg/game/types\"\n")
	}
	if needsMana {
		_, _ = b.WriteString("\t\"github.com/natefinch/council4/mtg/game/mana\"\n")
	}
	if needsOpt {
		_, _ = b.WriteString("\t\"github.com/natefinch/council4/opt\"\n")
	}
	_, _ = b.WriteString(")\n\n")

	if card.Layout == "reversible_card" && len(card.CardFaces) > 0 {
		for i, face := range faces {
			if i > 0 {
				_, _ = b.WriteString("\n")
			}
			writeSingleFaceComment(&b, face)
			if err := writeCardDef(&b, face, card.Layout, nil); err != nil {
				return "", err
			}
		}
		return formatGeneratedSource(b.String())
	}

	writeCardComment(&b, card, root, faces)
	if err := writeCardDef(&b, root, card.Layout, faces); err != nil {
		return "", err
	}

	return formatGeneratedSource(b.String())
}

func formatGeneratedSource(source string) (string, error) {
	formatted, err := format.Source([]byte(source))
	if err != nil {
		return "", fmt.Errorf("formatting generated source: %w", err)
	}
	return string(formatted), nil
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
	if len(card.CardFaces) < 2 || !layoutEmitsFaces(card.Layout) {
		return nil
	}
	faces := make([]generatedCardFields, 0, len(card.CardFaces)-1)
	for _, face := range card.CardFaces[1:] {
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

func fieldsNeedCost(fields generatedCardFields) bool {
	return fields.ManaCost != ""
}

func fieldsNeedColor(fields generatedCardFields) bool {
	return len(fields.Colors) > 0 || len(fields.ColorIdentity) > 0
}

func fieldsNeedMana(fields generatedCardFields) bool {
	return costLiteralNeedsManaPackage(fields.ManaCost)
}

func costLiteralNeedsManaPackage(manaCost string) bool {
	for _, match := range manaSymbolRe.FindAllStringSubmatch(manaCost, -1) {
		if strings.Contains(match[1], "/") {
			return true
		}
	}
	return false
}

func fieldsNeedOpt(fields generatedCardFields) bool {
	return fields.ManaCost != "" ||
		fields.Power != nil ||
		fields.Toughness != nil ||
		fields.Loyalty != nil ||
		fields.Defense != nil
}

func fieldsNeedTypes(fields generatedCardFields) bool {
	parsed := ParseTypeLine(fields.TypeLine)
	return len(parsed.Supertypes) > 0 ||
		len(parsed.Types) > 0 ||
		len(parsed.Subtypes) > 0
}

func anyFaceNeedsCost(faces []generatedCardFields) bool {
	return slices.ContainsFunc(faces, fieldsNeedCost)
}

func anyFaceNeedsColor(faces []generatedCardFields) bool {
	return slices.ContainsFunc(faces, fieldsNeedColor)
}

func anyFaceNeedsMana(faces []generatedCardFields) bool {
	return slices.ContainsFunc(faces, fieldsNeedMana)
}

func anyFaceNeedsOpt(faces []generatedCardFields) bool {
	return slices.ContainsFunc(faces, fieldsNeedOpt)
}

func writeCardComment(b *strings.Builder, card *ScryfallCard, root generatedCardFields, faces []generatedCardFields) {
	_, _ = fmt.Fprintf(b, "// %s\n", card.Name)
	_, _ = b.WriteString("//\n")
	_, _ = fmt.Fprintf(b, "// Type: %s\n", card.TypeLine)
	if card.ManaCost != "" {
		_, _ = fmt.Fprintf(b, "// Cost: %s\n", card.ManaCost)
	}
	if len(faces) > 0 {
		for _, face := range faces {
			_, _ = fmt.Fprintf(b, "// Face: %s — %s", face.Name, face.TypeLine)
			if face.ManaCost != "" {
				_, _ = fmt.Fprintf(b, " (%s)", face.ManaCost)
			}
			_, _ = b.WriteString("\n")
		}
	}
	_, _ = b.WriteString("//\n")
	_, _ = b.WriteString("// Oracle text:\n")
	oracle := card.OracleText
	if oracle == "" {
		oracle = root.OracleText
	}
	if oracle != "" {
		for line := range strings.SplitSeq(oracle, "\n") {
			_, _ = fmt.Fprintf(b, "//   %s\n", line)
		}
	} else {
		for i, face := range faces {
			if i > 0 {
				_, _ = b.WriteString("//   ---\n")
			}
			_, _ = fmt.Fprintf(b, "//   %s\n", face.Name)
			for line := range strings.SplitSeq(face.OracleText, "\n") {
				_, _ = fmt.Fprintf(b, "//   %s\n", line)
			}
		}
	}
	_, _ = b.WriteString("//\n")
	_, _ = b.WriteString("// TODO: Fill in ability fields from oracle text using categorized CardFace fields.\n")
}

func writeSingleFaceComment(b *strings.Builder, fields generatedCardFields) {
	_, _ = fmt.Fprintf(b, "// %s\n", fields.Name)
	_, _ = b.WriteString("//\n")
	_, _ = fmt.Fprintf(b, "// Type: %s\n", fields.TypeLine)
	if fields.ManaCost != "" {
		_, _ = fmt.Fprintf(b, "// Cost: %s\n", fields.ManaCost)
	}
	_, _ = b.WriteString("//\n")
	_, _ = b.WriteString("// Oracle text:\n")
	for line := range strings.SplitSeq(fields.OracleText, "\n") {
		_, _ = fmt.Fprintf(b, "//   %s\n", line)
	}
	_, _ = b.WriteString("//\n")
	_, _ = b.WriteString("// TODO: Fill in ability fields from oracle text using categorized CardFace fields.\n")
}

func writeCardDef(b *strings.Builder, root generatedCardFields, layout string, faces []generatedCardFields) error {
	varName := CardNameToVarName(root.Name)
	_, _ = fmt.Fprintf(b, "\nvar %s = &game.CardDef{\n", varName)
	if len(root.ColorIdentity) > 0 {
		_, _ = fmt.Fprintf(b, "\tColorIdentity: color.NewIdentity(%s),\n", colorLiterals(root.ColorIdentity))
	}
	_, _ = b.WriteString("\tCardFace: game.CardFace{\n")
	if err := writeFields(b, root, "\t\t", true); err != nil {
		return err
	}
	_, _ = b.WriteString("\t},\n")
	if layoutLiteral := layoutToLiteral(layout); layoutLiteral != "" {
		_, _ = fmt.Fprintf(b, "\tLayout: %s,\n", layoutLiteral)
	}
	if len(faces) > 0 {
		if len(faces) > 1 {
			return fmt.Errorf("%s has %d non-front faces; CardDef supports one optional Back face", root.Name, len(faces))
		}
		_, _ = b.WriteString("\tBack: opt.Val(game.CardFace{\n")
		if err := writeFields(b, faces[0], "\t\t", true); err != nil {
			return err
		}
		_, _ = b.WriteString("\t}),\n")
	}
	_, _ = b.WriteString("}\n")
	return nil
}

func writeFields(b *strings.Builder, fields generatedCardFields, indent string, includeName bool) error {
	if err := writePrintedScalarFields(b, fields, indent, includeName); err != nil {
		return err
	}
	if fields.EntersTapped {
		_, _ = fmt.Fprintf(b, "%sReplacementAbilities: []game.ReplacementAbility{\n", indent)
		_, _ = fmt.Fprintf(b, "%s\tgame.EntersTappedReplacement(%q),\n", indent, "This permanent enters tapped.")
		_, _ = fmt.Fprintf(b, "%s},\n", indent)
	}
	if fields.OracleText != "" {
		writeRawTextField(b, indent, "OracleText", fields.OracleText)
	}
	_, _ = fmt.Fprintf(b, "%s// TODO: Fill in ability fields from oracle text.\n", indent)
	_, _ = fmt.Fprintf(b, "%s// Use categorized CardFace fields (SpellAbility, ActivatedAbilities,\n", indent)
	_, _ = fmt.Fprintf(b, "%s// ManaAbilities, LoyaltyAbilities, TriggeredAbilities, StaticAbilities).\n", indent)
	_, _ = fmt.Fprintf(b, "%s// Follow mtg/cards/k/karplusan_forest.go: raw multiline Text values and\n", indent)
	_, _ = fmt.Fprintf(b, "%s// vertically expanded ability bodies. For mixed categories, use an initializer\n", indent)
	_, _ = fmt.Fprintf(b, "%s// function and append in oracle order.\n", indent)
	return nil
}

// writePrintedScalarFields emits the printed CardFace fields parsed from
// Scryfall data (name, mana cost, colors, types, and power/toughness/loyalty/
// defense). It is shared by the non-executable generator and the executable
// Renderer so both paths render identical printed fields.
func writePrintedScalarFields(b *strings.Builder, fields generatedCardFields, indent string, includeName bool) error {
	if includeName {
		_, _ = fmt.Fprintf(b, "%sName: %q,\n", indent, fields.Name)
	}
	if fields.ManaCost != "" {
		costLiteral, err := ParseManaCostLiteral(fields.ManaCost)
		if err != nil {
			return fmt.Errorf("parsing mana cost for %s: %w", fields.Name, err)
		}
		if costLiteral != "" {
			_, _ = fmt.Fprintf(b, "%sManaCost: opt.Val(%s),\n", indent, indentContinuation(costLiteral, indent))
		}
	}
	if len(fields.Colors) > 0 {
		_, _ = fmt.Fprintf(b, "%sColors: []color.Color{%s},\n", indent, colorLiterals(fields.Colors))
	}
	parsed := ParseTypeLine(fields.TypeLine)
	if len(parsed.Supertypes) > 0 {
		var literals []string
		for _, st := range parsed.Supertypes {
			literals = append(literals, SupertypeToLiteral(st))
		}
		_, _ = fmt.Fprintf(b, "%sSupertypes: []types.Super{%s},\n", indent, strings.Join(literals, ", "))
	}
	if len(parsed.Types) > 0 {
		var literals []string
		for _, t := range parsed.Types {
			literals = append(literals, CardTypeToLiteral(t))
		}
		_, _ = fmt.Fprintf(b, "%sTypes: []types.Card{%s},\n", indent, strings.Join(literals, ", "))
	}
	if len(parsed.Subtypes) > 0 {
		var literals []string
		for _, s := range parsed.Subtypes {
			literals = append(literals, SubtypeToLiteral(s, parsed.Types))
		}
		_, _ = fmt.Fprintf(b, "%sSubtypes: []types.Sub{%s},\n", indent, strings.Join(literals, ", "))
	}
	if fields.Power != nil {
		_, _ = fmt.Fprintf(b, "%sPower: opt.Val(%s),\n", indent, ptLiteral(*fields.Power))
	}
	if fields.Toughness != nil {
		_, _ = fmt.Fprintf(b, "%sToughness: opt.Val(%s),\n", indent, ptLiteral(*fields.Toughness))
	}
	if fields.Loyalty != nil {
		if n, err := strconv.Atoi(*fields.Loyalty); err == nil {
			_, _ = fmt.Fprintf(b, "%sLoyalty: opt.Val(%d),\n", indent, n)
		}
	}
	if fields.Defense != nil {
		if n, err := strconv.Atoi(*fields.Defense); err == nil {
			_, _ = fmt.Fprintf(b, "%sDefense: opt.Val(%d),\n", indent, n)
		}
	}
	return nil
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
// e.g., "Lightning Bolt" -> "LightningBolt", "Sol Ring" -> "SolRing".
func CardNameToVarName(name string) string {
	var b strings.Builder
	capitalize := true
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			capitalize = true
			continue
		}
		if capitalize {
			_, _ = b.WriteRune(unicode.ToUpper(r))
			capitalize = false
		} else {
			_, _ = b.WriteRune(r)
		}
	}
	return b.String()
}

// CardNameToFileName converts a card name to a snake_case file name.
// e.g., "Lightning Bolt" -> "lightning_bolt", "Sol Ring" -> "sol_ring".
func CardNameToFileName(name string) string {
	var b strings.Builder
	prevWasUnderscore := false
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			if !prevWasUnderscore && b.Len() > 0 {
				_, _ = b.WriteRune('_')
				prevWasUnderscore = true
			}
			continue
		}
		_, _ = b.WriteRune(unicode.ToLower(r))
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

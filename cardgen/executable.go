package cardgen

import (
	"fmt"
	"strconv"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GenerateExecutableCardSource generates a CardDef only when every Oracle-text
// ability can be lowered completely by the typed executable backend. It lowers
// each face into typed game values, assembles and validates a game.CardDef, and
// renders deterministic Go source from the typed values. Unsupported cards
// return diagnostics and an empty source string.
func GenerateExecutableCardSource(
	card *ScryfallCard,
	pkgName string,
) (string, []shared.Diagnostic, error) {
	return ExecutableGenerator{}.GenerateCardSource(card, pkgName)
}

// ExecutableGenerator configures executable CardDef source generation.
type ExecutableGenerator struct {
	IdentifierSuffix string
}

// GenerateCardSource generates one executable card source file.
func (g ExecutableGenerator) GenerateCardSource(
	card *ScryfallCard,
	pkgName string,
) (string, []shared.Diagnostic, error) {
	if !supportedLayouts[card.Layout] {
		return "", []shared.Diagnostic{{
			Severity: shared.SeverityWarning,
			Summary:  "unsupported card layout",
			Detail:   fmt.Sprintf("the source generator does not support Scryfall layout %q", card.Layout),
		}}, nil
	}
	if layoutEmitsAlternate(card.Layout) && len(card.CardFaces) > 2 {
		return "", []shared.Diagnostic{{
			Severity: shared.SeverityWarning,
			Summary:  "unsupported card layout",
			Detail:   fmt.Sprintf("the source generator supports at most 2 faces for %q layout cards, found %d", card.Layout, len(card.CardFaces)),
		}}, nil
	}

	faceAbilities, diagnostics := lowerExecutableFaces(card)
	if len(diagnostics) > 0 {
		return "", diagnostics, nil
	}

	defs, err := assembleCardDefs(card, faceAbilities)
	if err != nil {
		return "", nil, err
	}
	var validationDiagnostics []shared.Diagnostic
	for _, def := range defs {
		for _, issue := range game.ValidateCardDef(def) {
			validationDiagnostics = append(validationDiagnostics, shared.Diagnostic{
				Severity: shared.SeverityWarning,
				Summary:  "validation failed: " + string(issue.Code),
				Detail:   issue.Message,
			})
		}
	}
	if len(validationDiagnostics) > 0 {
		return "", validationDiagnostics, nil
	}

	source, err := (Renderer{IdentifierSuffix: g.IdentifierSuffix}).RenderCardSource(card, defs, faceHintsFrom(faceAbilities), pkgName)
	if err != nil {
		return "", nil, err
	}
	return source, nil, nil
}

// faceHintsFrom converts the typed lowering results into narrow rendering hints.
// Only presentation metadata (the package-level variable reference for a static
// ability) crosses into the renderer; every mechanical value comes from the
// validated game.CardDef. The expected body travels with each hint so the
// renderer can verify it against the CardDef before using the VarName.
func faceHintsFrom(faceAbilities []loweredFaceAbilities) []faceRenderHints {
	hints := make([]faceRenderHints, len(faceAbilities))
	for i := range faceAbilities {
		h := faceRenderHints{}
		for j := range faceAbilities[i].StaticAbilities {
			sa := &faceAbilities[i].StaticAbilities[j]
			h.StaticVarNames = append(h.StaticVarNames, staticVarHint{
				VarName: sa.VarName,
				Body:    sa.Body,
			})
		}
		hints[i] = h
	}
	return hints
}

// executableFaces returns the printed fields for each face in the positional
// order used by the typed lowering: reversible cards expose every face, while
// other layouts expose the root face followed by any additional emitted faces.
func executableFaces(card *ScryfallCard) []scryfallFaceFields {
	if card.Layout == "reversible_card" && len(card.CardFaces) > 0 {
		return facesFromAllCardFaces(card)
	}
	faces := []scryfallFaceFields{rootFields(card)}
	faces = append(faces, generatedFaces(card)...)
	faces = append(faces, alternateFields(card)...)
	return faces
}

// assembleCardDefs builds the typed game.CardDef values validated before
// rendering. Reversible cards yield one CardDef per face; other layouts yield a
// single CardDef with an optional Back face.
func assembleCardDefs(card *ScryfallCard, faceAbilities []loweredFaceAbilities) ([]*game.CardDef, error) {
	faces := executableFaces(card)
	if len(faces) != len(faceAbilities) {
		return nil, fmt.Errorf("face count mismatch: %d printed faces, %d lowered faces", len(faces), len(faceAbilities))
	}

	if card.Layout == "reversible_card" && len(card.CardFaces) > 0 {
		defs := make([]*game.CardDef, 0, len(faces))
		for i, fields := range faces {
			face, err := buildCardFace(fields, faceAbilities[i])
			if err != nil {
				return nil, err
			}
			defs = append(defs, &game.CardDef{
				CardFace:      face,
				Layout:        cardLayoutValue(card.Layout),
				ColorIdentity: identityValue(fields.ColorIdentity),
			})
		}
		return defs, nil
	}

	rootFace, err := buildCardFace(faces[0], faceAbilities[0])
	if err != nil {
		return nil, err
	}
	def := &game.CardDef{
		CardFace:      rootFace,
		Layout:        cardLayoutValue(card.Layout),
		ColorIdentity: identityValue(faces[0].ColorIdentity),
	}
	if len(faces) > 1 {
		otherFace, err := buildCardFace(faces[1], faceAbilities[1])
		if err != nil {
			return nil, err
		}
		if layoutEmitsAlternate(card.Layout) {
			def.Alternate = opt.Val(otherFace)
		} else {
			def.Back = opt.Val(otherFace)
		}
	}
	return []*game.CardDef{def}, nil
}

func buildCardFace(fields scryfallFaceFields, abilities loweredFaceAbilities) (game.CardFace, error) {
	face := game.CardFace{
		Name:       fields.Name,
		OracleText: fields.OracleText,
	}
	if fields.ManaCost != "" {
		manaCost, err := parseManaCostValue(fields.ManaCost)
		if err != nil {
			return game.CardFace{}, fmt.Errorf("parsing mana cost for %s: %w", fields.Name, err)
		}
		if len(manaCost) > 0 {
			face.ManaCost = opt.Val(manaCost)
		}
	}
	for _, letter := range fields.Colors {
		if value, ok := parseColorValue(letter); ok {
			face.Colors = append(face.Colors, value)
		}
	}
	parsed := ParseTypeLine(fields.TypeLine)
	for _, supertype := range parsed.Supertypes {
		face.Supertypes = append(face.Supertypes, types.Super(supertype))
	}
	for _, cardType := range parsed.Types {
		face.Types = append(face.Types, types.Card(cardType))
	}
	for _, subtype := range parsed.Subtypes {
		face.Subtypes = append(face.Subtypes, types.Sub(subtype))
	}
	if fields.Power != nil {
		face.Power = opt.Val(parsePTValue(*fields.Power))
	}
	if fields.Toughness != nil {
		face.Toughness = opt.Val(parsePTValue(*fields.Toughness))
	}
	if abilities.DynamicPower.Exists {
		face.DynamicPower = abilities.DynamicPower
	}
	if abilities.DynamicToughness.Exists {
		face.DynamicToughness = abilities.DynamicToughness
	}
	if fields.Loyalty != nil {
		if n, err := strconv.Atoi(*fields.Loyalty); err == nil {
			face.Loyalty = opt.Val(n)
		}
	}
	if fields.Defense != nil {
		if n, err := strconv.Atoi(*fields.Defense); err == nil {
			face.Defense = opt.Val(n)
		}
	}
	face.EntersPrepared = abilities.EntersPrepared
	for i := range abilities.StaticAbilities {
		face.StaticAbilities = append(face.StaticAbilities, abilities.StaticAbilities[i].Body)
	}
	if faceHasKeyword(&face, game.Devoid) {
		face.Colors = nil
	}
	face.ActivatedAbilities = abilities.ActivatedAbilities
	face.ManaAbilities = abilities.ManaAbilities
	face.LoyaltyAbilities = abilities.LoyaltyAbilities
	face.TriggeredAbilities = abilities.TriggeredAbilities
	face.ChapterAbilities = abilities.ChapterAbilities
	face.ReplacementAbilities = abilities.ReplacementAbilities
	face.SpellAbility = abilities.SpellAbility
	face.Overload = abilities.Overload
	face.AdditionalCosts = abilities.AdditionalCosts
	face.AlternativeCosts = abilities.AlternativeCosts
	return face, nil
}

func faceHasKeyword(face *game.CardFace, keyword game.Keyword) bool {
	for i := range face.StaticAbilities {
		if game.BodyHasKeyword(&face.StaticAbilities[i], keyword) {
			return true
		}
	}
	return false
}

func identityValue(letters []string) color.Identity {
	colors := make([]color.Color, 0, len(letters))
	for _, letter := range letters {
		if value, ok := parseColorValue(letter); ok {
			colors = append(colors, value)
		}
	}
	if len(colors) == 0 {
		return color.Identity{}
	}
	return color.NewIdentity(colors...)
}

func cardLayoutValue(layout string) game.CardLayout {
	switch layout {
	case "transform":
		return game.LayoutTransform
	case "modal_dfc":
		return game.LayoutModalDFC
	case "meld":
		return game.LayoutMeld
	case "double_faced_token":
		return game.LayoutDoubleFacedToken
	case "reversible_card":
		return game.LayoutReversibleCard
	case "adventure":
		return game.LayoutAdventure
	case "split":
		return game.LayoutSplit
	case "prepare":
		return game.LayoutPrepare
	default:
		return game.LayoutNormal
	}
}

// parseColorValue converts a Scryfall color letter (e.g., "W") into a typed
// color.Color. It reports false for unrecognized letters.
func parseColorValue(letter string) (color.Color, bool) {
	switch letter {
	case "W":
		return color.White, true
	case "U":
		return color.Blue, true
	case "B":
		return color.Black, true
	case "R":
		return color.Red, true
	case "G":
		return color.Green, true
	default:
		return "", false
	}
}

// parsePTValue converts a Scryfall power/toughness string into a typed game.PT.
func parsePTValue(value string) game.PT {
	if value == "*" {
		return game.PT{IsStar: true}
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return game.PT{}
	}
	return game.PT{Value: n}
}

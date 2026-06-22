package cardgen

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
)

// Import paths that the renderer may emit. The game package is always needed.
const (
	importGame    = "github.com/natefinch/council4/mtg/game"
	importColor   = "github.com/natefinch/council4/mtg/game/color"
	importCompare = "github.com/natefinch/council4/mtg/game/compare"
	importCounter = "github.com/natefinch/council4/mtg/game/counter"
	importCost    = "github.com/natefinch/council4/mtg/game/cost"
	importMana    = "github.com/natefinch/council4/mtg/game/mana"
	importTypes   = "github.com/natefinch/council4/mtg/game/types"
	importZone    = "github.com/natefinch/council4/mtg/game/zone"
	importOpt     = "github.com/natefinch/council4/opt"
)

// renderCtx accumulates import paths needed during one rendering pass.
// It is not safe for concurrent use.
type renderCtx struct {
	imports   map[string]struct{}
	tokenBase string
	tokenDefs []tokenDefEntry
	tokenKeys map[string]string
}

// tokenDefEntry is a synthesized token CardDef to emit as a package-level var
// alongside the card that creates it.
type tokenDefEntry struct {
	varName string
	def     *game.CardDef
}

func newRenderCtx() *renderCtx {
	return &renderCtx{imports: map[string]struct{}{importGame: {}}}
}

func (c *renderCtx) need(path string) { c.imports[path] = struct{}{} }

// tokenDefVar registers a synthesized token CardDef for emission and returns the
// package-level var name to reference it by. Structurally identical token defs
// share one var; names are unique within the generated file.
func (c *renderCtx) tokenDefVar(def *game.CardDef) string {
	key := tokenDefKey(def)
	if name, ok := c.tokenKeys[key]; ok {
		return name
	}
	base := c.tokenBase
	if base == "" {
		base = lowerFirst(CardNameToVarName(def.Name))
	}
	if base == "" {
		base = "token"
	}
	name := base + "Token"
	for i := 2; tokenNameTaken(c.tokenDefs, name); i++ {
		name = fmt.Sprintf("%sToken%d", base, i)
	}
	if c.tokenKeys == nil {
		c.tokenKeys = map[string]string{}
	}
	c.tokenKeys[key] = name
	c.tokenDefs = append(c.tokenDefs, tokenDefEntry{varName: name, def: def})
	return name
}

func tokenNameTaken(defs []tokenDefEntry, name string) bool {
	for _, d := range defs {
		if d.varName == name {
			return true
		}
	}
	return false
}

// tokenDefKey is a structural identity for a synthesized token def so identical
// tokens reuse one emitted var.
func tokenDefKey(def *game.CardDef) string {
	return fmt.Sprintf("%s|%v|%v|%v|%v|%v", def.Name, def.Types, def.Subtypes, def.Colors,
		def.Power, def.Toughness)
}

func (c *renderCtx) sortedImports() []string {
	paths := make([]string, 0, len(c.imports))
	for p := range c.imports {
		paths = append(paths, p)
	}
	slices.Sort(paths)
	return paths
}

// faceRenderHints carries presentation metadata for rendering one card face.
// The renderer uses hints only to select rendering style; all mechanical values
// come from the validated game.CardDef. A hint is verified against the CardDef
// before use; a mismatch returns an error.
type faceRenderHints struct {
	// StaticVarNames is indexed parallel to game.CardFace.StaticAbilities.
	// An empty VarName means "render as struct literal".
	StaticVarNames []staticVarHint
}

// staticVarHint carries an optional package-level variable reference and the
// expected StaticAbility body for divergence verification before use.
type staticVarHint struct {
	VarName string
	Body    game.StaticAbility
}

// renderPTValue renders a typed game.PT as a Go literal.
func renderPTValue(pt game.PT) string {
	if pt.IsStar {
		return "game.PT{IsStar: true}"
	}
	return fmt.Sprintf("game.PT{Value: %d}", pt.Value)
}

// renderDynamicValue renders a typed game.DynamicValue as a Go literal. It is
// used for characteristic-defining power/toughness values ("equal to the number
// of cards in your hand").
func renderDynamicValue(value game.DynamicValue) string {
	kind := dynamicValueKindLiteral(value.Kind)
	fields := ""
	if value.Value != 0 {
		fields += fmt.Sprintf(", Value: %d", value.Value)
	}
	if value.Offset != 0 {
		fields += fmt.Sprintf(", Offset: %d", value.Offset)
	}
	return fmt.Sprintf("game.DynamicValue{Kind: %s%s}", kind, fields)
}

func dynamicValueKindLiteral(kind game.DynamicValueKind) string {
	switch kind {
	case game.DynamicValueConstant:
		return "game.DynamicValueConstant"
	case game.DynamicValueControllerHandSize:
		return "game.DynamicValueControllerHandSize"
	case game.DynamicValueControllerGraveyardSize:
		return "game.DynamicValueControllerGraveyardSize"
	case game.DynamicValueControllerCreatureCount:
		return "game.DynamicValueControllerCreatureCount"
	case game.DynamicValueControllerLandCount:
		return "game.DynamicValueControllerLandCount"
	case game.DynamicValueControllerArtifactCount:
		return "game.DynamicValueControllerArtifactCount"
	case game.DynamicValueAllBattlefieldCreatureCount:
		return "game.DynamicValueAllBattlefieldCreatureCount"
	case game.DynamicValueAllGraveyardsSize:
		return "game.DynamicValueAllGraveyardsSize"
	case game.DynamicValueCreatureCardsInAllGraveyards:
		return "game.DynamicValueCreatureCardsInAllGraveyards"
	case game.DynamicValueCardTypesAmongAllGraveyards:
		return "game.DynamicValueCardTypesAmongAllGraveyards"
	default:
		return "game.DynamicValueNone"
	}
}

// Renderer renders typed game ability values and complete CardDef values as
// deterministic Go source. IdentifierSuffix disambiguates distinct cards that
// share a printed name without changing CardDef.Name. A zero-value Renderer is
// ready to use. Every method renders from typed values using exported accessors
// so that repeated calls with identical input produce byte-identical output.
type Renderer struct {
	IdentifierSuffix string
}

// RenderCardSource renders a complete Go source file for executable CardDefs.
// The validated game.CardDef values in defs are the sole source of every
// mechanical and ability value. The original ScryfallCard provides only
// comment and variable-name metadata and the layout used to map defs to faces.
// The hints carry presentation metadata (such as static-ability variable
// references) verified against the CardDef values before use.
func (r Renderer) RenderCardSource(
	card *ScryfallCard,
	defs []*game.CardDef,
	hints []faceRenderHints,
	pkgName string,
) (string, error) {
	if len(defs) == 0 {
		return "", errors.New("render: no CardDef to render")
	}

	ctx := newRenderCtx()
	if len(defs) > 0 {
		ctx.tokenBase = lowerFirst(CardNameToVarName(defs[0].Name))
	}
	reversible := card.Layout == "reversible_card" && len(card.CardFaces) > 0

	var body strings.Builder
	if reversible {
		commentFaces := facesFromAllCardFaces(card)
		for i, def := range defs {
			if i > 0 {
				_, _ = body.WriteString("\n")
			}
			if i < len(commentFaces) {
				r.writeFaceComment(&body, commentFaces[i])
			}
			if err := r.writeReversibleFaceDef(&body, ctx, def, card.Layout, hintAt(hints, i)); err != nil {
				return "", err
			}
		}
	} else {
		root := rootFields(card)
		faces := generatedFaces(card)
		if len(faces) == 0 {
			faces = alternateFields(card)
		}
		r.writeCardComment(&body, card, root, faces)
		if err := r.writeCardDef(&body, ctx, defs[0], card.Layout, hints); err != nil {
			return "", err
		}
	}

	for _, entry := range ctx.tokenDefs {
		_, _ = body.WriteString("\n")
		if err := r.writeTokenDefVar(&body, ctx, entry); err != nil {
			return "", err
		}
	}

	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "package %s\n\n", pkgName)
	r.writeImports(&b, ctx)
	_, _ = b.WriteString(body.String())
	return formatGeneratedSource(b.String())
}

func hintAt(hints []faceRenderHints, i int) faceRenderHints {
	if i < len(hints) {
		return hints[i]
	}
	return faceRenderHints{}
}

func (Renderer) writeImports(b *strings.Builder, ctx *renderCtx) {
	_, _ = b.WriteString("import (\n")
	for _, path := range ctx.sortedImports() {
		_, _ = fmt.Fprintf(b, "\t%q\n", path)
	}
	_, _ = b.WriteString(")\n\n")
}

func (Renderer) writeCardComment(b *strings.Builder, card *ScryfallCard, root scryfallFaceFields, faces []scryfallFaceFields) {
	_, _ = fmt.Fprintf(b, "// %s\n", card.Name)
	_, _ = b.WriteString("//\n")
	_, _ = fmt.Fprintf(b, "// Type: %s\n", card.TypeLine)
	if card.ManaCost != "" {
		_, _ = fmt.Fprintf(b, "// Cost: %s\n", card.ManaCost)
	}
	for _, face := range faces {
		_, _ = fmt.Fprintf(b, "// Face: %s — %s", face.Name, face.TypeLine)
		if face.ManaCost != "" {
			_, _ = fmt.Fprintf(b, " (%s)", face.ManaCost)
		}
		_, _ = b.WriteString("\n")
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
}

func (Renderer) writeFaceComment(b *strings.Builder, fields scryfallFaceFields) {
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
}

func (r Renderer) writeCardDef(
	b *strings.Builder,
	ctx *renderCtx,
	def *game.CardDef,
	layout string,
	hints []faceRenderHints,
) error {
	varName := CardNameToVarName(def.Name) + r.IdentifierSuffix
	if r.IdentifierSuffix != "" {
		_, _ = fmt.Fprintf(b, "\n// %s is the card definition for %s.\n", varName, def.Name)
	}
	_, _ = fmt.Fprintf(b, "var %s = &game.CardDef{\n", varName)
	if cols := def.ColorIdentity.Colors(); len(cols) > 0 {
		ctx.need(importColor)
		colorLits, err := colorValueLiterals(cols)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(b, "\tColorIdentity: color.NewIdentity(%s),\n", colorLits)
	}
	_, _ = b.WriteString("\tCardFace: game.CardFace{\n")
	if err := r.writeFaceFields(b, ctx, &def.CardFace, "\t\t", hintAt(hints, 0)); err != nil {
		return err
	}
	_, _ = b.WriteString("\t},\n")
	if layoutLiteral := layoutToLiteral(layout); layoutLiteral != "" {
		_, _ = fmt.Fprintf(b, "\tLayout: %s,\n", layoutLiteral)
	}
	if def.Back.Exists {
		ctx.need(importOpt)
		_, _ = b.WriteString("\tBack: opt.Val(game.CardFace{\n")
		if err := r.writeFaceFields(b, ctx, &def.Back.Val, "\t\t", hintAt(hints, 1)); err != nil {
			return err
		}
		_, _ = b.WriteString("\t}),\n")
	}
	if def.Alternate.Exists {
		ctx.need(importOpt)
		_, _ = b.WriteString("\tAlternate: opt.Val(game.CardFace{\n")
		if err := r.writeFaceFields(b, ctx, &def.Alternate.Val, "\t\t", hintAt(hints, 1)); err != nil {
			return err
		}
		_, _ = b.WriteString("\t}),\n")
	}
	_, _ = b.WriteString("}\n")
	return nil
}

// writeTokenDefVar emits a synthesized token CardDef as a package-level var. The
// token def is a plain creature face (name, types, subtypes, colors, P/T) with no
// abilities, referenced by a CreateToken primitive via game.TokenDef.
func (r Renderer) writeTokenDefVar(b *strings.Builder, ctx *renderCtx, entry tokenDefEntry) error {
	_, _ = fmt.Fprintf(b, "var %s = &game.CardDef{\n", entry.varName)
	_, _ = b.WriteString("\tCardFace: game.CardFace{\n")
	if err := r.writeFaceFields(b, ctx, &entry.def.CardFace, "\t\t", tokenFaceHints(entry.def)); err != nil {
		return err
	}
	_, _ = b.WriteString("\t},\n")
	_, _ = b.WriteString("}\n")
	return nil
}

// tokenFaceHints reconstructs render hints for a synthesized token's static
// abilities by matching each typed body against the keywordStaticBodies catalog,
// so a keyword like flying renders as game.FlyingStaticBody rather than an
// unrenderable struct literal. Bodies with no catalog match get an empty hint and
// fall back to structural rendering.
func tokenFaceHints(def *game.CardDef) faceRenderHints {
	if len(def.StaticAbilities) == 0 {
		return faceRenderHints{}
	}
	hints := faceRenderHints{StaticVarNames: make([]staticVarHint, len(def.StaticAbilities))}
	for i := range def.StaticAbilities {
		hints.StaticVarNames[i] = staticVarHint{Body: def.StaticAbilities[i]}
		hints.StaticVarNames[i].VarName = tokenStaticBodyVarName(&def.StaticAbilities[i])
	}
	return hints
}

// tokenStaticBodyVarName returns the package-level variable reference the Renderer
// emits for a synthesized token's typed static-ability body, or "" if the body has
// no reusable variable and must fall back to structural rendering. It mirrors the
// keyword bodies lowered onto synthesized tokens in lower_token.go.
func tokenStaticBodyVarName(body *game.StaticAbility) string {
	for kw := range keywordStaticBodies {
		if reflect.DeepEqual(keywordStaticBodies[kw].Body, *body) {
			return keywordStaticBodies[kw].VarName
		}
	}
	return ""
}

func (r Renderer) writeReversibleFaceDef(b *strings.Builder, ctx *renderCtx, def *game.CardDef, layout string, hints faceRenderHints) error {
	varName := CardNameToVarName(def.Name) + r.IdentifierSuffix
	if r.IdentifierSuffix != "" {
		_, _ = fmt.Fprintf(b, "\n// %s is the card definition for %s.\n", varName, def.Name)
	}
	_, _ = fmt.Fprintf(b, "var %s = &game.CardDef{\n", varName)
	if cols := def.ColorIdentity.Colors(); len(cols) > 0 {
		ctx.need(importColor)
		colorLits, err := colorValueLiterals(cols)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(b, "\tColorIdentity: color.NewIdentity(%s),\n", colorLits)
	}
	_, _ = b.WriteString("\tCardFace: game.CardFace{\n")
	if err := r.writeFaceFields(b, ctx, &def.CardFace, "\t\t", hints); err != nil {
		return err
	}
	_, _ = b.WriteString("\t},\n")
	if layoutLiteral := layoutToLiteral(layout); layoutLiteral != "" {
		_, _ = fmt.Fprintf(b, "\tLayout: %s,\n", layoutLiteral)
	}
	_, _ = b.WriteString("}\n")
	return nil
}

func (r Renderer) writeFaceFields(b *strings.Builder, ctx *renderCtx, face *game.CardFace, indent string, hints faceRenderHints) error {
	if err := r.writeFaceScalarFields(b, ctx, face, indent); err != nil {
		return err
	}
	block, err := r.renderFaceAbilityFields(ctx, face, hints)
	if err != nil {
		return err
	}
	for _, field := range block {
		for line := range strings.SplitSeq(field, "\n") {
			_, _ = fmt.Fprintf(b, "%s%s\n", indent, line)
		}
	}
	if face.OracleText != "" {
		writeRawTextField(b, indent, "OracleText", face.OracleText)
	}
	return nil
}

// writeFaceScalarFields renders the printed scalar CardFace fields (name, mana
// cost, colors, types, power/toughness/loyalty/defense) directly from the
// validated typed values on face.
func (Renderer) writeFaceScalarFields(b *strings.Builder, ctx *renderCtx, face *game.CardFace, indent string) error {
	_, _ = fmt.Fprintf(b, "%sName: %q,\n", indent, face.Name)
	if face.ManaCost.Exists {
		ctx.need(importOpt)
		rawCostLit, err := renderManaCostMultiline(ctx, face.ManaCost.Val)
		if err != nil {
			return err
		}
		costLiteral := indentContinuation(rawCostLit, indent)
		_, _ = fmt.Fprintf(b, "%sManaCost: opt.Val(%s),\n", indent, costLiteral)
	}
	if len(face.Colors) > 0 {
		ctx.need(importColor)
		colorLits, err := colorValueLiterals(face.Colors)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(b, "%sColors: []color.Color{%s},\n", indent, colorLits)
	}
	if face.EntersPrepared {
		_, _ = fmt.Fprintf(b, "%sEntersPrepared: true,\n", indent)
	}
	if len(face.Supertypes) > 0 {
		ctx.need(importTypes)
		literals := make([]string, 0, len(face.Supertypes))
		for _, st := range face.Supertypes {
			lit, err := supertypeLiteral(st)
			if err != nil {
				return err
			}
			literals = append(literals, lit)
		}
		_, _ = fmt.Fprintf(b, "%sSupertypes: []types.Super{%s},\n", indent, strings.Join(literals, ", "))
	}
	if len(face.Types) > 0 {
		ctx.need(importTypes)
		literals := make([]string, 0, len(face.Types))
		for _, t := range face.Types {
			lit, err := cardTypeLiteral(t)
			if err != nil {
				return err
			}
			literals = append(literals, lit)
		}
		_, _ = fmt.Fprintf(b, "%sTypes: []types.Card{%s},\n", indent, strings.Join(literals, ", "))
	}
	if len(face.Subtypes) > 0 {
		ctx.need(importTypes)
		cardTypeStrings := make([]string, 0, len(face.Types))
		for _, t := range face.Types {
			cardTypeStrings = append(cardTypeStrings, string(t))
		}
		literals := make([]string, 0, len(face.Subtypes))
		for _, sub := range face.Subtypes {
			literals = append(literals, SubtypeToLiteral(string(sub), cardTypeStrings))
		}
		_, _ = fmt.Fprintf(b, "%sSubtypes: []types.Sub{%s},\n", indent, strings.Join(literals, ", "))
	}
	if face.Power.Exists {
		ctx.need(importOpt)
		_, _ = fmt.Fprintf(b, "%sPower: opt.Val(%s),\n", indent, renderPTValue(face.Power.Val))
	}
	if face.Toughness.Exists {
		ctx.need(importOpt)
		_, _ = fmt.Fprintf(b, "%sToughness: opt.Val(%s),\n", indent, renderPTValue(face.Toughness.Val))
	}
	if face.DynamicPower.Exists {
		ctx.need(importOpt)
		ctx.need(importGame)
		_, _ = fmt.Fprintf(b, "%sDynamicPower: opt.Val(%s),\n", indent, renderDynamicValue(face.DynamicPower.Val))
	}
	if face.DynamicToughness.Exists {
		ctx.need(importOpt)
		ctx.need(importGame)
		_, _ = fmt.Fprintf(b, "%sDynamicToughness: opt.Val(%s),\n", indent, renderDynamicValue(face.DynamicToughness.Val))
	}
	if face.Loyalty.Exists {
		ctx.need(importOpt)
		_, _ = fmt.Fprintf(b, "%sLoyalty: opt.Val(%d),\n", indent, face.Loyalty.Val)
	}
	if face.Defense.Exists {
		ctx.need(importOpt)
		_, _ = fmt.Fprintf(b, "%sDefense: opt.Val(%d),\n", indent, face.Defense.Val)
	}
	return nil
}

// colorValueLiterals renders a slice of typed color.Color values as a
// comma-separated list of Go constant references. Returns an error for any
// unrecognised color value.
func colorValueLiterals(colors []color.Color) (string, error) {
	literals := make([]string, 0, len(colors))
	for _, c := range colors {
		lit, err := colorValueToLiteral(c)
		if err != nil {
			return "", err
		}
		literals = append(literals, lit)
	}
	return strings.Join(literals, ", "), nil
}

func colorValueToLiteral(c color.Color) (string, error) {
	switch c {
	case color.White:
		return "color.White", nil
	case color.Blue:
		return "color.Blue", nil
	case color.Black:
		return "color.Black", nil
	case color.Red:
		return "color.Red", nil
	case color.Green:
		return "color.Green", nil
	default:
		return "", fmt.Errorf("render: unsupported color %q", string(c))
	}
}

// renderFaceAbilityFields renders the categorized ability fields for one face in
// canonical order. Each returned element is a complete "Field: value," fragment.
// All values come from the validated face; hints only select rendering style and
// are verified against the face before use.
func (r Renderer) renderFaceAbilityFields(ctx *renderCtx, face *game.CardFace, hints faceRenderHints) ([]string, error) {
	var fields []string

	if len(face.StaticAbilities) > 0 {
		elements := make([]string, 0, len(face.StaticAbilities))
		for i := range face.StaticAbilities {
			hint := staticHintAt(hints, i)
			if hint != nil && hint.VarName != "" && !reflect.DeepEqual(hint.Body, face.StaticAbilities[i]) {
				return nil, fmt.Errorf("render: hint VarName %q for static ability %d does not match CardDef value (divergence)", hint.VarName, i)
			}
			rendered, err := r.renderStaticAbility(ctx, &face.StaticAbilities[i], hint)
			if err != nil {
				return nil, err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("StaticAbilities", "game.StaticAbility", elements))
	}

	if len(face.ActivatedAbilities) > 0 {
		elements := make([]string, 0, len(face.ActivatedAbilities))
		for i := range face.ActivatedAbilities {
			rendered, err := r.renderActivatedAbility(ctx, &face.ActivatedAbilities[i])
			if err != nil {
				return nil, err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("ActivatedAbilities", "game.ActivatedAbility", elements))
	}

	if len(face.ManaAbilities) > 0 {
		elements := make([]string, 0, len(face.ManaAbilities))
		for i := range face.ManaAbilities {
			rendered, err := r.renderManaAbility(ctx, &face.ManaAbilities[i])
			if err != nil {
				return nil, err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("ManaAbilities", "game.ManaAbility", elements))
	}

	if len(face.TriggeredAbilities) > 0 {
		elements := make([]string, 0, len(face.TriggeredAbilities))
		for i := range face.TriggeredAbilities {
			rendered, err := r.renderTriggeredAbility(ctx, &face.TriggeredAbilities[i])
			if err != nil {
				return nil, err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("TriggeredAbilities", "game.TriggeredAbility", elements))
	}

	if len(face.ChapterAbilities) > 0 {
		elements := make([]string, 0, len(face.ChapterAbilities))
		for i := range face.ChapterAbilities {
			rendered, err := r.renderChapterAbility(ctx, &face.ChapterAbilities[i])
			if err != nil {
				return nil, err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("ChapterAbilities", "game.ChapterAbility", elements))
	}

	if len(face.LoyaltyAbilities) > 0 {
		elements := make([]string, 0, len(face.LoyaltyAbilities))
		for i := range face.LoyaltyAbilities {
			rendered, err := r.renderLoyaltyAbility(ctx, &face.LoyaltyAbilities[i])
			if err != nil {
				return nil, err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("LoyaltyAbilities", "game.LoyaltyAbility", elements))
	}

	if len(face.ReplacementAbilities) > 0 {
		elements := make([]string, 0, len(face.ReplacementAbilities))
		for i := range face.ReplacementAbilities {
			rendered, err := r.renderReplacementAbility(ctx, &face.ReplacementAbilities[i])
			if err != nil {
				return nil, err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("ReplacementAbilities", "game.ReplacementAbility", elements))
	}

	if len(face.AdditionalCosts) > 0 {
		rendered, err := r.renderAdditionalCosts(ctx, face.AdditionalCosts)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("AdditionalCosts: %s,", rendered))
	}

	if len(face.AlternativeCosts) > 0 {
		rendered, err := r.renderAlternativeCosts(ctx, face.AlternativeCosts)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("AlternativeCosts: %s,", rendered))
	}

	if face.Overload.Exists {
		ctx.need(importOpt)
		manaCost, err := r.renderManaCost(ctx, face.Overload.Val.Cost)
		if err != nil {
			return nil, err
		}
		content, err := r.renderAbilityContent(ctx, face.Overload.Val.SpellAbility)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf(
			"Overload: opt.Val(game.OverloadAbility{Cost: %s, SpellAbility: %s}),",
			manaCost,
			content,
		))
	}

	if face.SpellAbility.Exists {
		ctx.need(importOpt)
		content, err := r.renderAbilityContent(ctx, face.SpellAbility.Val)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("SpellAbility: opt.Val(%s),", content))
	}

	return fields, nil
}

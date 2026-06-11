package cardgen

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
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
	imports map[string]struct{}
}

func newRenderCtx() *renderCtx {
	return &renderCtx{imports: map[string]struct{}{importGame: {}}}
}

func (c *renderCtx) need(path string) { c.imports[path] = struct{}{} }

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

func staticHintAt(hints faceRenderHints, i int) *staticVarHint {
	if i < len(hints.StaticVarNames) {
		return &hints.StaticVarNames[i]
	}
	return nil
}

func (r Renderer) renderStaticAbility(ctx *renderCtx, body *game.StaticAbility, hint *staticVarHint) (string, error) {
	if hint != nil && hint.VarName != "" {
		return hint.VarName, nil
	}
	if protectedColors := game.StaticBodyProtectionColors(body); len(protectedColors) > 0 {
		renderedColors, err := renderColorArguments(ctx, protectedColors)
		if err != nil {
			return "", err
		}
		if reflect.DeepEqual(*body, game.ProtectionFromColorsStaticAbility(protectedColors...)) {
			return fmt.Sprintf("game.ProtectionFromColorsStaticAbility(%s)", renderedColors), nil
		}
	}
	if target, ok := game.StaticBodyEnchantTarget(body); ok &&
		reflect.DeepEqual(*body, game.EnchantStaticAbility(&target)) {
		renderedTarget, err := r.renderTargetSpec(ctx, &target)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.EnchantStaticAbility(&%s)", renderedTarget), nil
	}
	if manaCost, ok := game.StaticBodyWardCost(body); ok &&
		reflect.DeepEqual(*body, game.WardStaticAbility(manaCost)) {
		renderedCost, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.WardStaticAbility(%s)", renderedCost), nil
	}
	var fields []string
	if body.Text != "" {
		fields = append(fields, fmt.Sprintf("Text: %s,", renderText(body.Text)))
	}
	if len(body.KeywordAbilities) > 0 {
		elements := make([]string, 0, len(body.KeywordAbilities))
		for _, keyword := range body.KeywordAbilities {
			rendered, err := r.renderKeywordAbility(ctx, keyword)
			if err != nil {
				return "", err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("KeywordAbilities", "game.KeywordAbility", elements))
	}
	if body.Condition.Exists {
		rendered, err := r.renderStaticAbilityCondition(ctx, &body.Condition.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Condition: opt.Val(%s),", rendered))
	}
	if len(body.ContinuousEffects) > 0 {
		elements := make([]string, 0, len(body.ContinuousEffects))
		for i := range body.ContinuousEffects {
			rendered, err := r.renderContinuousEffect(ctx, &body.ContinuousEffects[i])
			if err != nil {
				return "", err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("ContinuousEffects", "game.ContinuousEffect", elements))
	}
	if len(body.RuleEffects) > 0 {
		elements := make([]string, 0, len(body.RuleEffects))
		for i := range body.RuleEffects {
			rendered, err := r.renderRuleEffect(ctx, &body.RuleEffects[i])
			if err != nil {
				return "", err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("RuleEffects", "game.RuleEffect", elements))
	}
	return structLit("game.StaticAbility", fields), nil
}

func (r Renderer) renderContinuousEffect(ctx *renderCtx, effect *game.ContinuousEffect) (string, error) {
	var fields []string
	if len(effect.RemoveKeywords) > 0 || len(effect.AddAbilities) > 0 {
		return "", errors.New("render: unsupported ability-layer continuous effect fields")
	}
	if effect.AffectedSource && !effect.Group.Empty() {
		return "", errors.New("render: continuous effect cannot set both AffectedSource and Group")
	}
	switch effect.Layer {
	case game.LayerAbility:
		if effect.PowerDelta != 0 || effect.ToughnessDelta != 0 {
			return "", errors.New("render: power/toughness fields require a power/toughness layer")
		}
	case game.LayerPowerToughnessModify:
		if len(effect.AddKeywords) > 0 {
			return "", errors.New("render: keyword fields require the ability layer")
		}
	default:
	}
	layerLit, err := renderContinuousLayer(effect.Layer)
	if err != nil {
		return "", err
	}
	fields = append(fields, fmt.Sprintf("Layer: %s,", layerLit))
	if effect.AffectedSource {
		fields = append(fields, "AffectedSource: true,")
	}
	if effect.Group.Valid() {
		groupLit, err := r.renderGroupReference(ctx, effect.Group)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Group: %s,", groupLit))
	}
	if effect.PowerDelta != 0 {
		fields = append(fields, fmt.Sprintf("PowerDelta: %d,", effect.PowerDelta))
	}
	if effect.ToughnessDelta != 0 {
		fields = append(fields, fmt.Sprintf("ToughnessDelta: %d,", effect.ToughnessDelta))
	}
	if len(effect.AddKeywords) > 0 {
		elements := make([]string, 0, len(effect.AddKeywords))
		for _, keyword := range effect.AddKeywords {
			literal, err := renderKeyword(keyword)
			if err != nil {
				return "", err
			}
			elements = append(elements, literal+",")
		}
		fields = append(fields, sliceField("AddKeywords", "game.Keyword", elements))
	}
	return structLit("game.ContinuousEffect", fields), nil
}

func renderContinuousLayer(layer game.ContinuousLayer) (string, error) {
	switch layer {
	case game.LayerAbility:
		return "game.LayerAbility", nil
	case game.LayerPowerToughnessModify:
		return "game.LayerPowerToughnessModify", nil
	default:
		return "", fmt.Errorf("render: unsupported continuous layer %d", layer)
	}
}

func (r Renderer) renderRuleEffect(ctx *renderCtx, effect *game.RuleEffect) (string, error) {
	kind, err := renderRuleEffectKind(effect.Kind)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Kind: %s,", kind)}
	if effect.AffectedPlayer != game.PlayerAny {
		player, err := renderPlayerRelation(effect.AffectedPlayer)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("AffectedPlayer: %s,", player))
	}
	if !effect.CardSelection.Empty() {
		selection, err := r.renderSelection(ctx, effect.CardSelection)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("CardSelection: %s,", selection))
	}
	if effect.Kind == game.RuleEffectGrantHandCardAbility {
		if !game.BodyHasKeyword(effect.GrantedAbility, game.Cycling) {
			return "", errors.New("render: hand-card ability grant must grant Cycling")
		}
		ability, err := r.renderActivatedAbility(ctx, &effect.GrantedAbility)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("GrantedAbility: %s,", ability))
	}
	if effect.Kind == game.RuleEffectCostModifier {
		modifier, err := r.renderCostModifier(ctx, effect.CostModifier)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("CostModifier: %s,", modifier))
	}
	return structLit("game.RuleEffect", fields), nil
}

func renderRuleEffectKind(kind game.RuleEffectKind) (string, error) {
	switch kind {
	case game.RuleEffectCostModifier:
		return "game.RuleEffectCostModifier", nil
	case game.RuleEffectGrantHandCardAbility:
		return "game.RuleEffectGrantHandCardAbility", nil
	default:
		return "", fmt.Errorf("render: unsupported rule effect kind %d", kind)
	}
}

func (r Renderer) renderCostModifier(ctx *renderCtx, modifier game.CostModifier) (string, error) {
	kind, err := renderCostModifierKind(modifier.Kind)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Kind: %s,", kind)}
	if modifier.MatchCardType {
		cardType, err := cardTypeLiteral(modifier.CardType)
		if err != nil {
			return "", err
		}
		fields = append(fields, "MatchCardType: true,", fmt.Sprintf("CardType: %s,", cardType))
	}
	if modifier.AbilityKeyword != game.KeywordNone {
		keyword, err := renderKeyword(modifier.AbilityKeyword)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("AbilityKeyword: %s,", keyword))
	}
	if modifier.GenericIncrease != 0 {
		fields = append(fields, fmt.Sprintf("GenericIncrease: %d,", modifier.GenericIncrease))
	}
	if modifier.GenericReduction != 0 {
		fields = append(fields, fmt.Sprintf("GenericReduction: %d,", modifier.GenericReduction))
	}
	if modifier.SetGeneric.Exists {
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("SetGeneric: opt.Val(%d),", modifier.SetGeneric.Val))
	}
	if modifier.SetManaCost.Exists {
		ctx.need(importOpt)
		manaCost, err := r.renderManaCost(ctx, modifier.SetManaCost.Val)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("SetManaCost: opt.Val(%s),", manaCost))
	}
	if modifier.MinimumGeneric != 0 {
		fields = append(fields, fmt.Sprintf("MinimumGeneric: %d,", modifier.MinimumGeneric))
	}
	if modifier.FirstCycleEachTurn {
		fields = append(fields, "FirstCycleEachTurn: true,")
	}
	return structLit("game.CostModifier", fields), nil
}

func renderCostModifierKind(kind game.CostModifierKind) (string, error) {
	switch kind {
	case game.CostModifierSpell:
		return "game.CostModifierSpell", nil
	case game.CostModifierAbility:
		return "game.CostModifierAbility", nil
	case game.CostModifierAttack:
		return "game.CostModifierAttack", nil
	default:
		return "", fmt.Errorf("render: unsupported cost modifier kind %d", kind)
	}
}

func (r Renderer) renderActivatedAbility(ctx *renderCtx, ability *game.ActivatedAbility) (string, error) {
	if manaCost, ok := game.ActivatedBodyEquipCost(ability); ok &&
		reflect.DeepEqual(*ability, game.EquipActivatedAbility(manaCost)) {
		renderedCost, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.EquipActivatedAbility(%s)", renderedCost), nil
	}
	if manaCost, ok := game.ActivatedBodyCyclingCost(ability); ok &&
		reflect.DeepEqual(*ability, game.CyclingActivatedAbility(manaCost)) {
		renderedCost, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.CyclingActivatedAbility(%s)", renderedCost), nil
	}

	var fields []string
	if ability.Text != "" {
		fields = append(fields, fmt.Sprintf("Text: %s,", renderText(ability.Text)))
	}
	if ability.ManaCost.Exists {
		ctx.need(importOpt)
		manaCostLit, err := r.renderManaCost(ctx, ability.ManaCost.Val)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ManaCost: opt.Val(%s),", manaCostLit))
	}
	if len(ability.AdditionalCosts) > 0 {
		rendered, err := r.renderAdditionalCosts(ctx, ability.AdditionalCosts)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("AdditionalCosts: %s,", rendered))
	}
	if ability.ZoneOfFunction != zone.None {
		ctx.need(importZone)
		zoneLiteral, err := renderZone(ability.ZoneOfFunction)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ZoneOfFunction: %s,", zoneLiteral))
	}
	if ability.Timing != game.NoTimingRestriction {
		timing, err := renderTimingRestriction(ability.Timing)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Timing: %s,", timing))
	}
	if len(ability.KeywordAbilities) > 0 {
		elements := make([]string, 0, len(ability.KeywordAbilities))
		for _, keyword := range ability.KeywordAbilities {
			rendered, err := r.renderKeywordAbility(ctx, keyword)
			if err != nil {
				return "", err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("KeywordAbilities", "game.KeywordAbility", elements))
	}
	content, err := r.renderAbilityContent(ctx, ability.Content)
	if err != nil {
		return "", err
	}
	fields = append(fields, fmt.Sprintf("Content: %s,", content))
	return structLit("game.ActivatedAbility", fields), nil
}

func (r Renderer) renderManaAbility(ctx *renderCtx, ability *game.ManaAbility) (string, error) {
	for _, manaColor := range []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G, mana.C} {
		if !reflect.DeepEqual(*ability, game.TapManaAbility(manaColor)) {
			continue
		}
		ctx.need(importMana)
		colorLiteral, err := renderManaColor(manaColor)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.TapManaAbility(%s)", colorLiteral), nil
	}
	if colors, ok := tapManaChoiceColors(ability); ok &&
		reflect.DeepEqual(*ability, game.TapManaChoiceAbility(colors...)) {
		ctx.need(importMana)
		colorLiterals := make([]string, 0, len(colors))
		for _, manaColor := range colors {
			colorLiteral, err := renderManaColor(manaColor)
			if err != nil {
				return "", err
			}
			colorLiterals = append(colorLiterals, colorLiteral)
		}
		return fmt.Sprintf("game.TapManaChoiceAbility(%s)", strings.Join(colorLiterals, ", ")), nil
	}

	var fields []string
	if ability.Text != "" {
		fields = append(fields, fmt.Sprintf("Text: %s,", renderText(ability.Text)))
	}
	if ability.ManaCost.Exists {
		ctx.need(importOpt)
		manaCostLit, err := r.renderManaCost(ctx, ability.ManaCost.Val)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ManaCost: opt.Val(%s),", manaCostLit))
	}
	if len(ability.AdditionalCosts) > 0 {
		rendered, err := r.renderAdditionalCosts(ctx, ability.AdditionalCosts)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("AdditionalCosts: %s,", rendered))
	}
	if ability.Timing != game.NoTimingRestriction {
		timing, err := renderTimingRestriction(ability.Timing)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Timing: %s,", timing))
	}
	content, err := r.renderAbilityContent(ctx, ability.Content)
	if err != nil {
		return "", err
	}
	fields = append(fields, fmt.Sprintf("Content: %s,", content))
	return structLit("game.ManaAbility", fields), nil
}

func renderTimingRestriction(timing game.TimingRestriction) (string, error) {
	switch timing {
	case game.NoTimingRestriction:
		return "game.NoTimingRestriction", nil
	case game.SorceryOnly:
		return "game.SorceryOnly", nil
	case game.OncePerTurn:
		return "game.OncePerTurn", nil
	case game.SorceryOncePerTurn:
		return "game.SorceryOncePerTurn", nil
	case game.DuringCombat:
		return "game.DuringCombat", nil
	case game.DuringUpkeep:
		return "game.DuringUpkeep", nil
	default:
		return "", fmt.Errorf("unsupported timing restriction %d", timing)
	}
}

func tapManaChoiceColors(ability *game.ManaAbility) ([]mana.Color, bool) {
	content := ability.Content
	if content.IsModal() || len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 2 {
		return nil, false
	}
	choose, ok := content.Modes[0].Sequence[0].Primitive.(game.Choose)
	if !ok || len(choose.Choice.Colors) < 2 || len(choose.Choice.Colors) > 6 {
		return nil, false
	}
	seen := make(map[mana.Color]struct{}, len(choose.Choice.Colors))
	for _, manaColor := range choose.Choice.Colors {
		switch manaColor {
		case mana.W, mana.U, mana.B, mana.R, mana.G, mana.C:
		default:
			return nil, false
		}
		if _, duplicate := seen[manaColor]; duplicate {
			return nil, false
		}
		seen[manaColor] = struct{}{}
	}
	return choose.Choice.Colors, true
}

func (r Renderer) renderTriggeredAbility(ctx *renderCtx, ability *game.TriggeredAbility) (string, error) {
	var fields []string
	if ability.Text != "" {
		fields = append(fields, fmt.Sprintf("Text: %s,", renderText(ability.Text)))
	}
	trigger, err := r.renderTriggerCondition(ctx, &ability.Trigger)
	if err != nil {
		return "", err
	}
	fields = append(fields, fmt.Sprintf("Trigger: %s,", trigger))
	if ability.Optional {
		fields = append(fields, "Optional: true,")
	}
	content, err := r.renderAbilityContent(ctx, ability.Content)
	if err != nil {
		return "", err
	}
	fields = append(fields, fmt.Sprintf("Content: %s,", content))
	return structLit("game.TriggeredAbility", fields), nil
}

func (r Renderer) renderChapterAbility(ctx *renderCtx, ability *game.ChapterAbility) (string, error) {
	content, err := r.renderAbilityContent(ctx, ability.Content)
	if err != nil {
		return "", err
	}
	return structLit("game.ChapterAbility", []string{
		fmt.Sprintf("Text: %s,", renderText(ability.Text)),
		fmt.Sprintf("Chapters: %#v,", ability.Chapters),
		fmt.Sprintf("Content: %s,", content),
	}), nil
}

func (r Renderer) renderLoyaltyAbility(ctx *renderCtx, ability *game.LoyaltyAbility) (string, error) {
	var fields []string
	if ability.Text != "" {
		fields = append(fields, fmt.Sprintf("Text: %s,", renderText(ability.Text)))
	}
	fields = append(fields, fmt.Sprintf("LoyaltyCost: %d,", ability.LoyaltyCost))
	content, err := r.renderAbilityContent(ctx, ability.Content)
	if err != nil {
		return "", err
	}
	fields = append(fields, fmt.Sprintf("Content: %s,", content))
	return structLit("game.LoyaltyAbility", fields), nil
}

func (r Renderer) renderTriggerCondition(ctx *renderCtx, trigger *game.TriggerCondition) (string, error) {
	triggerType, err := renderTriggerType(trigger.Type)
	if err != nil {
		return "", err
	}
	pattern, err := r.renderTriggerPattern(ctx, &trigger.Pattern)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Type: %s,", triggerType),
		fmt.Sprintf("Pattern: %s,", pattern),
	}
	if trigger.InterveningIf != "" {
		fields = append(fields, fmt.Sprintf("InterveningIf: %q,", trigger.InterveningIf))
	}
	if trigger.InterveningCondition.Exists {
		condition, err := r.renderControlledPermanentInterveningCondition(ctx, &trigger.InterveningCondition.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("InterveningCondition: opt.Val(%s),", condition))
	}
	if trigger.InterveningIfEventPermanentHadNoCounterKind.Exists {
		kind, err := renderCounterKind(trigger.InterveningIfEventPermanentHadNoCounterKind.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importCounter)
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("InterveningIfEventPermanentHadNoCounterKind: opt.Val(%s),", kind))
	}
	if trigger.InterveningIfEventPermanentWasKicked {
		fields = append(fields, "InterveningIfEventPermanentWasKicked: true,")
	}
	if trigger.InterveningIfEventPermanentWasCast {
		fields = append(fields, "InterveningIfEventPermanentWasCast: true,")
	}
	return structLit("game.TriggerCondition", fields), nil
}

func (r Renderer) renderControlledPermanentInterveningCondition(ctx *renderCtx, condition *game.Condition) (string, error) {
	if condition == nil || !condition.ControlsMatching.Exists {
		return "", errors.New("render: unsupported trigger intervening condition")
	}
	unsupported := *condition
	unsupported.Text = ""
	unsupported.ControlsMatching = opt.V[game.SelectionCount]{}
	if !unsupported.Empty() || unsupported.Negate {
		return "", errors.New("render: unsupported trigger intervening condition")
	}
	count := condition.ControlsMatching.Val
	if count.Selection.Empty() || count.MinCount != 0 || count.TotalPower.Exists {
		return "", errors.New("render: unsupported trigger intervening controls condition")
	}
	selection, err := r.renderSelection(ctx, count.Selection)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Text: %q,", condition.Text),
		fmt.Sprintf("ControlsMatching: opt.Val(game.SelectionCount{Selection: %s}),", selection),
	}
	ctx.need(importOpt)
	return structLit("game.Condition", fields), nil
}

func (Renderer) renderTriggerPattern(ctx *renderCtx, pattern *game.TriggerPattern) (string, error) {
	if (pattern.Event == game.EventBeginningOfStep) != (pattern.Step != game.StepNone) {
		return "", errors.New("render: beginning-of-step trigger pattern must set exactly one supported step")
	}
	if !pattern.SubjectSelection.Empty() ||
		len(pattern.RequireCardTypes) != 0 ||
		len(pattern.ExcludeCardTypes) != 0 ||
		pattern.MatchFromZone ||
		pattern.MatchToZone ||
		pattern.MatchStackObjectKind ||
		pattern.DamageRecipientCombatState != game.CombatStateAny ||
		pattern.SpellTargetsSource ||
		pattern.SpellTargetAllow != game.TargetAllowUnspecified ||
		pattern.SpellTargetPattern.Exists {
		return "", errors.New("render: unsupported trigger pattern fields")
	}
	if !pattern.CardSelection.Empty() && pattern.Event != game.EventSpellCast {
		return "", errors.New("render: CardSelection is only supported for EventSpellCast trigger patterns")
	}
	if !pattern.CardSelection.Empty() {
		unsupported := pattern.CardSelection
		unsupported.RequiredTypes = nil
		unsupported.RequiredTypesAny = nil
		unsupported.ExcludedTypes = nil
		unsupported.ColorsAny = nil
		unsupported.Colorless = false
		unsupported.Multicolored = false
		if !unsupported.Empty() {
			return "", errors.New("render: unsupported CardSelection fields in cast trigger pattern")
		}
	}
	event, err := renderEventKind(pattern.Event)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Event: %s,", event)}
	if pattern.Source != game.TriggerSourceAny {
		source, err := renderTriggerSource(pattern.Source)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Source: %s,", source))
	}
	if pattern.Controller != game.TriggerControllerAny {
		controller, err := renderTriggerController(pattern.Controller)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Controller: %s,", controller))
	}
	if pattern.Step != game.StepNone {
		step, err := renderStep(pattern.Step)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Step: %s,", step))
	}
	if pattern.Subject != game.TriggerSubjectDefault {
		subject, err := renderTriggerSubject(pattern.Subject)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Subject: %s,", subject))
	}
	if pattern.ExcludeSelf {
		fields = append(fields, "ExcludeSelf: true,")
	}
	if len(pattern.RequirePermanentTypes) > 0 {
		rpt, err := renderTypesCardSlice(ctx, pattern.RequirePermanentTypes)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("RequirePermanentTypes: %s,", rpt))
	}
	if len(pattern.ExcludePermanentTypes) > 0 {
		ept, err := renderTypesCardSlice(ctx, pattern.ExcludePermanentTypes)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ExcludePermanentTypes: %s,", ept))
	}
	if pattern.RequireNonToken {
		fields = append(fields, "RequireNonToken: true,")
	}
	if pattern.OneOrMore {
		fields = append(fields, "OneOrMore: true,")
	}
	if pattern.Player != game.TriggerPlayerAny {
		player, err := renderTriggerPlayer(pattern.Player)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Player: %s,", player))
	}
	if pattern.DamageRecipient != game.DamageRecipientNone {
		recipient, err := renderDamageRecipient(pattern.DamageRecipient)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("DamageRecipient: %s,", recipient))
	}
	if len(pattern.DamageRecipientTypes) > 0 {
		recipientTypes, err := renderTypesCardSlice(ctx, pattern.DamageRecipientTypes)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("DamageRecipientTypes: %s,", recipientTypes))
	}
	if pattern.RequireCombatDamage {
		fields = append(fields, "RequireCombatDamage: true,")
	}
	if !pattern.CardSelection.Empty() {
		sel, err := (Renderer{}).renderSelection(ctx, pattern.CardSelection)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("CardSelection: %s,", sel))
	}
	return structLit("game.TriggerPattern", fields), nil
}

func (r Renderer) renderReplacementAbility(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	if len(ability.Replacement.EntersWithCounters) > 0 {
		if ability.Replacement.EntersTapped ||
			ability.UnlessPaid.Exists ||
			ability.Replacement.Condition.Exists {
			return "", errors.New("render: ETB counter replacement cannot also tap, require payment, or have a condition")
		}
		placements, err := renderCounterPlacements(ctx, ability.Replacement.EntersWithCounters)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.EntersWithCountersReplacement(%q, %s)", ability.Text, strings.Join(placements, ", ")), nil
	}
	if ability.Replacement.EntersTapped && ability.UnlessPaid.Exists {
		if ability.Replacement.Condition.Exists {
			return "", errors.New("render: paid ETB replacement cannot also have a condition")
		}
		payment, err := r.renderResolutionPayment(ctx, ability.UnlessPaid.Val)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.EntersTappedUnlessPaidReplacement(%q, %s)", ability.Text, payment), nil
	}
	if ability.Replacement.EntersTapped && !ability.UnlessPaid.Exists {
		if !ability.Replacement.Condition.Exists {
			return fmt.Sprintf("game.EntersTappedReplacement(%q)", ability.Text), nil
		}
		condStr, err := r.renderConditionForETBReplacement(ctx, &ability.Replacement.Condition.Val)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.EntersTappedIfReplacement(%q, %s)", ability.Text, condStr), nil
	}
	if ability.Replacement.ReplaceToZone != zone.None {
		replacement, err := renderZoneDestinationReplacement(ctx, ability)
		if err != nil {
			return "", err
		}
		return replacement, nil
	}
	if ability.Replacement.TokenMultiplier > 0 {
		replacement, err := renderTokenCreationReplacement(ability)
		if err != nil {
			return "", err
		}
		return replacement, nil
	}
	if ability.Replacement.DamageMultiplier > 0 || ability.Replacement.DamageAddend != 0 {
		replacement, err := renderDamageReplacement(ctx, ability)
		if err != nil {
			return "", err
		}
		return replacement, nil
	}
	if ability.Replacement.CounterMultiplier > 0 {
		replacement, err := renderCounterPlacementReplacement(ctx, ability)
		if err != nil {
			return "", err
		}
		return replacement, nil
	}
	return "", fmt.Errorf("render: unsupported replacement ability %q", ability.Text)
}

func renderDamageReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	if replacement.EntersTapped ||
		len(replacement.EntersWithCounters) != 0 ||
		ability.UnlessPaid.Exists ||
		replacement.Condition.Exists ||
		replacement.MatchEvent != game.EventDamageDealt ||
		replacement.ControllerFilter == game.TriggerControllerAny ||
		(replacement.DamageMultiplier <= 1 && replacement.DamageAddend == 0) {
		return "", errors.New("render: unsupported damage replacement shape")
	}
	controller, err := renderTriggerController(replacement.ControllerFilter)
	if err != nil {
		return "", err
	}
	colors := "nil"
	if len(replacement.DamageSourceColors) > 0 {
		colors, err = renderColorSlice(ctx, replacement.DamageSourceColors)
		if err != nil {
			return "", err
		}
	}
	constructor := "game.DamageReplacement"
	if replacement.DamageExcludeSource {
		constructor = "game.DamageReplacementExcludingSource"
	}
	return fmt.Sprintf("%s(%q, %d, %d, %s, %s)",
		constructor,
		ability.Text,
		replacement.DamageMultiplier,
		replacement.DamageAddend,
		colors,
		controller,
	), nil
}

func renderCounterPlacementReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	if replacement.EntersTapped ||
		len(replacement.EntersWithCounters) != 0 ||
		ability.UnlessPaid.Exists ||
		replacement.Condition.Exists ||
		replacement.MatchEvent != game.EventCountersAdded ||
		replacement.ControllerFilter == game.TriggerControllerAny ||
		replacement.CounterMultiplier <= 1 {
		return "", errors.New("render: unsupported counter-placement replacement shape")
	}
	controller, err := renderTriggerController(replacement.ControllerFilter)
	if err != nil {
		return "", err
	}
	if !replacement.MatchCounterKind {
		return fmt.Sprintf("game.AnyCounterPlacementReplacement(%q, %d, %s)",
			ability.Text,
			replacement.CounterMultiplier,
			controller,
		), nil
	}
	kind, err := renderCounterKind(replacement.CounterKindFilter)
	if err != nil {
		return "", err
	}
	ctx.need(importCounter)
	return fmt.Sprintf("game.CounterPlacementReplacement(%q, %d, %s, %s)",
		ability.Text,
		replacement.CounterMultiplier,
		kind,
		controller,
	), nil
}

func renderTokenCreationReplacement(ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	if replacement.EntersTapped ||
		len(replacement.EntersWithCounters) != 0 ||
		ability.UnlessPaid.Exists ||
		replacement.Condition.Exists ||
		replacement.MatchEvent != game.EventTokenCreated ||
		replacement.ControllerFilter == game.TriggerControllerAny ||
		replacement.TokenMultiplier <= 1 {
		return "", errors.New("render: unsupported token-creation replacement shape")
	}
	controller, err := renderTriggerController(replacement.ControllerFilter)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("game.TokenCreationReplacement(%q, %d, %s)",
		ability.Text,
		replacement.TokenMultiplier,
		controller,
	), nil
}

func renderZoneDestinationReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	if replacement.EntersTapped ||
		len(replacement.EntersWithCounters) != 0 ||
		ability.UnlessPaid.Exists ||
		replacement.Condition.Exists ||
		replacement.MatchEvent != game.EventZoneChanged ||
		!replacement.MatchToZone ||
		replacement.ToZone == zone.None {
		return "", errors.New("render: unsupported zone-destination replacement shape")
	}
	toZone, err := renderZone(replacement.ToZone)
	if err != nil {
		return "", err
	}
	replaceToZone, err := renderZone(replacement.ReplaceToZone)
	if err != nil {
		return "", err
	}
	fields := []string{
		"MatchEvent: game.EventZoneChanged,",
		"MatchToZone: true,",
		fmt.Sprintf("ToZone: %s,", toZone),
		fmt.Sprintf("ReplaceToZone: %s,", replaceToZone),
		"Duration: game.DurationPermanent,",
	}
	if replacement.ShuffleIntoLibrary {
		if replacement.ReplaceToZone != zone.Library {
			return "", errors.New("render: shuffle-into-library replacement must replace to library")
		}
		fields = append(fields, "ShuffleIntoLibrary: true,")
	}
	if replacement.RevealSource {
		fields = append(fields, "RevealSource: true,")
	}
	if replacement.MatchFromZone {
		fromZone, err := renderZone(replacement.FromZone)
		if err != nil {
			return "", err
		}
		fields = append(fields, "MatchFromZone: true,", fmt.Sprintf("FromZone: %s,", fromZone))
	}
	ctx.need(importZone)
	return fmt.Sprintf("game.ReplacementAbility{Text: %q, Replacement: %s}",
		ability.Text,
		structLit("game.ReplacementEffect", fields),
	), nil
}

func renderCounterPlacements(ctx *renderCtx, placements []game.CounterPlacement) ([]string, error) {
	rendered := make([]string, 0, len(placements))
	for _, placement := range placements {
		if placement.Amount <= 0 {
			return nil, fmt.Errorf("render: invalid ETB counter amount %d", placement.Amount)
		}
		kind, err := renderCounterKind(placement.Kind)
		if err != nil {
			return nil, err
		}
		ctx.need(importCounter)
		rendered = append(rendered, fmt.Sprintf("game.CounterPlacement{Kind: %s, Amount: %d}", kind, placement.Amount))
	}
	return rendered, nil
}

func (r Renderer) renderResolutionPayment(ctx *renderCtx, payment game.ResolutionPayment) (string, error) {
	var fields []string
	hasCost := payment.ManaCost.Exists || len(payment.AdditionalCosts) > 0
	if !hasCost {
		return "", errors.New("render: resolution payment has no cost")
	}
	if payment.Prompt != "" {
		fields = append(fields, fmt.Sprintf("Prompt: %q,", payment.Prompt))
	}
	if payment.ManaCost.Exists {
		manaCost, err := renderManaCostMultiline(ctx, payment.ManaCost.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("ManaCost: opt.Val(%s),", manaCost))
	}
	if len(payment.AdditionalCosts) > 0 {
		additionalCosts, err := r.renderAdditionalCosts(ctx, payment.AdditionalCosts)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("AdditionalCosts: %s,", additionalCosts))
	}
	if payment.XValue != 0 {
		fields = append(fields, fmt.Sprintf("XValue: %d,", payment.XValue))
	}
	return structLit("game.ResolutionPayment", fields), nil
}

// renderConditionForETBReplacement renders a game.Condition for use in a
// conditional enters-tapped replacement. Only the exact supported shape is
// accepted; any other combination returns an error.
func (r Renderer) renderConditionForETBReplacement(ctx *renderCtx, cond *game.Condition) (string, error) {
	rendered, err := r.renderControllerControlsCondition(ctx, cond, "ETB replacement")
	if err != nil {
		return "", err
	}
	return "&" + rendered, nil
}

func (r Renderer) renderStaticAbilityCondition(ctx *renderCtx, cond *game.Condition) (string, error) {
	return r.renderControllerControlsCondition(ctx, cond, "static ability")
}

func (r Renderer) renderControllerControlsCondition(ctx *renderCtx, cond *game.Condition, context string) (string, error) {
	if cond.ControllerLifeAtLeast < 0 ||
		cond.ControllerHandSizeAtLeast < 0 ||
		cond.AnyPlayerLifeAtMost < 0 ||
		cond.OpponentCountAtLeast < 0 {
		return "", fmt.Errorf("render: %s condition has a negative threshold", context)
	}
	// Reject unsupported condition fields.
	if cond.ControlsMatching.Exists ||
		cond.Object.Exists ||
		len(cond.Types) != 0 ||
		cond.EventPermanentNameUniqueAmongControlledAndGraveyardCreatures ||
		cond.SourceClassLevelAtLeast != 0 ||
		cond.SourceClassLevelLessThan != 0 ||
		cond.SourceNotMonstrous ||
		cond.ControllerHasMaxSpeed ||
		cond.TargetEnteredThisTurn.Exists ||
		cond.CastFromZone.Exists {
		return "", fmt.Errorf("render: unsupported condition shape for %s", context)
	}
	var fields []string
	if cond.Text != "" {
		fields = append(fields, fmt.Sprintf("Text: %s,", renderText(cond.Text)))
	}
	if cond.Negate {
		fields = append(fields, "Negate: true,")
	}
	hasPredicate := false
	if !cond.ControllerControls.Empty() {
		filter := cond.ControllerControls
		if filter.Power.Exists ||
			filter.Toughness.Exists ||
			filter.TotalPower.Exists {
			return "", fmt.Errorf("render: unsupported PermanentFilter shape for %s condition", context)
		}
		filterStr, err := r.renderPermanentFilterForCondition(ctx, filter)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ControllerControls: %s,", filterStr))
		hasPredicate = true
	}
	if cond.ControllerLifeAtLeast > 0 {
		fields = append(fields, fmt.Sprintf("ControllerLifeAtLeast: %d,", cond.ControllerLifeAtLeast))
		hasPredicate = true
	}
	if cond.ControllerHandSizeAtLeast > 0 {
		fields = append(fields, fmt.Sprintf("ControllerHandSizeAtLeast: %d,", cond.ControllerHandSizeAtLeast))
		hasPredicate = true
	}
	if cond.AnyPlayerLifeAtMost > 0 {
		fields = append(fields, fmt.Sprintf("AnyPlayerLifeAtMost: %d,", cond.AnyPlayerLifeAtMost))
		hasPredicate = true
	}
	if cond.OpponentCountAtLeast > 0 {
		fields = append(fields, fmt.Sprintf("OpponentCountAtLeast: %d,", cond.OpponentCountAtLeast))
		hasPredicate = true
	}
	if cond.AnyOpponentControls.Exists {
		rendered, err := r.renderSelectionCountForCondition(ctx, cond.AnyOpponentControls.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("AnyOpponentControls: opt.Val(%s),", rendered))
		hasPredicate = true
	}
	if cond.OpponentsControl.Exists {
		rendered, err := r.renderSelectionCountForCondition(ctx, cond.OpponentsControl.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("OpponentsControl: opt.Val(%s),", rendered))
		hasPredicate = true
	}
	if !hasPredicate {
		return "", fmt.Errorf("render: %s condition has no supported predicate", context)
	}
	return structLit("game.Condition", fields), nil
}

func (r Renderer) renderSelectionCountForCondition(ctx *renderCtx, count game.SelectionCount) (string, error) {
	if count.MinCount < 0 {
		return "", errors.New("render: condition permanent-count threshold cannot be negative")
	}
	selection, err := r.renderSelection(ctx, count.Selection)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Selection: %s,", selection)}
	if count.MinCount != 0 {
		fields = append(fields, fmt.Sprintf("MinCount: %d,", count.MinCount))
	}
	if count.TotalPower.Exists {
		ctx.need(importOpt)
		cmp, err := renderCompareInt(ctx, count.TotalPower.Val)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("TotalPower: opt.Val(%s),", cmp))
	}
	return structLit("game.SelectionCount", fields), nil
}

func (Renderer) renderPermanentFilterForCondition(ctx *renderCtx, filter game.PermanentFilter) (string, error) {
	if filter.MinCount < 0 {
		return "", errors.New("render: condition permanent-count threshold cannot be negative")
	}
	var fields []string
	if len(filter.Types) > 0 {
		lits, err := renderTypesCardSlice(ctx, filter.Types)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Types: %s,", lits))
	}
	if len(filter.Supertypes) > 0 {
		ctx.need(importTypes)
		literals := make([]string, 0, len(filter.Supertypes))
		for _, st := range filter.Supertypes {
			lit, err := supertypeLiteral(st)
			if err != nil {
				return "", err
			}
			literals = append(literals, lit)
		}
		fields = append(fields, fmt.Sprintf("Supertypes: []types.Super{%s},", strings.Join(literals, ", ")))
	}
	if len(filter.SubtypesAny) > 0 {
		ctx.need(importTypes)
		literals := make([]string, 0, len(filter.SubtypesAny))
		cardTypes := make([]string, 0, len(filter.Types))
		for _, cardType := range filter.Types {
			cardTypes = append(cardTypes, string(cardType))
		}
		for _, subtype := range filter.SubtypesAny {
			literals = append(literals, SubtypeToLiteral(string(subtype), cardTypes))
		}
		fields = append(fields, fmt.Sprintf("SubtypesAny: []types.Sub{%s},", strings.Join(literals, ", ")))
	}
	if len(filter.ColorsAny) > 0 {
		literals, err := renderColorSlice(ctx, filter.ColorsAny)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ColorsAny: %s,", literals))
	}
	if len(filter.ExcludedColors) > 0 {
		literals, err := renderColorSlice(ctx, filter.ExcludedColors)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ExcludedColors: %s,", literals))
	}
	if filter.MinCount != 0 {
		fields = append(fields, fmt.Sprintf("MinCount: %d,", filter.MinCount))
	}
	if filter.ExcludeSource {
		fields = append(fields, "ExcludeSource: true,")
	}
	return structLit("game.PermanentFilter", fields), nil
}

func (r Renderer) renderAbilityContent(ctx *renderCtx, content game.AbilityContent) (string, error) {
	if !content.IsModal() {
		mode, err := r.renderMode(ctx, content.Modes[0])
		if err != nil {
			return "", err
		}
		return mode + ".Ability()", nil
	}
	return r.renderModalAbilityContent(ctx, content)
}

// renderModalAbilityContent renders a modal game.AbilityContent with multiple
// modes, MinModes, and MaxModes as a game.AbilityContent struct literal.
func (r Renderer) renderModalAbilityContent(ctx *renderCtx, content game.AbilityContent) (string, error) {
	if len(content.Modes) == 0 {
		return "", errors.New("render: modal ability content has no modes")
	}
	modeElements := make([]string, 0, len(content.Modes))
	for i := range content.Modes {
		rendered, err := r.renderMode(ctx, content.Modes[i])
		if err != nil {
			return "", err
		}
		modeElements = append(modeElements, rendered+",")
	}
	fields := []string{sliceField("Modes", "game.Mode", modeElements)}
	if content.MinModes != 0 {
		fields = append(fields, fmt.Sprintf("MinModes: %d,", content.MinModes))
	}
	if content.MaxModes != 0 {
		fields = append(fields, fmt.Sprintf("MaxModes: %d,", content.MaxModes))
	}
	return structLit("game.AbilityContent", fields), nil
}

func (r Renderer) renderMode(ctx *renderCtx, mode game.Mode) (string, error) {
	var fields []string
	if mode.Text != "" {
		fields = append(fields, fmt.Sprintf("Text: %s,", renderText(mode.Text)))
	}
	if len(mode.Targets) > 0 {
		elements := make([]string, 0, len(mode.Targets))
		for i := range mode.Targets {
			rendered, err := r.renderTargetSpec(ctx, &mode.Targets[i])
			if err != nil {
				return "", err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("Targets", "game.TargetSpec", elements))
	}
	elements := make([]string, 0, len(mode.Sequence))
	for i := range mode.Sequence {
		rendered, err := r.renderInstruction(ctx, &mode.Sequence[i])
		if err != nil {
			return "", err
		}
		elements = append(elements, rendered+",")
	}
	fields = append(fields, sliceField("Sequence", "game.Instruction", elements))
	return structLit("game.Mode", fields), nil
}

func (r Renderer) renderInstruction(ctx *renderCtx, instruction *game.Instruction) (string, error) {
	primitive, err := r.renderPrimitive(ctx, instruction.Primitive)
	if err != nil {
		return "", err
	}
	return structLit("", []string{fmt.Sprintf("Primitive: %s,", primitive)}), nil
}

func (r Renderer) renderPrimitive(ctx *renderCtx, primitive game.Primitive) (string, error) {
	if primitive == nil {
		return "", errors.New("render: nil primitive")
	}
	switch primitive.Kind() {
	case game.PrimitiveDamage:
		return r.renderDamagePrimitive(ctx, primitive)
	case game.PrimitiveDraw, game.PrimitiveDiscard, game.PrimitiveMill,
		game.PrimitiveScry, game.PrimitiveSurveil, game.PrimitiveGainLife,
		game.PrimitiveLoseLife:
		return r.renderPlayerAmountPrimitive(ctx, primitive)
	case game.PrimitiveInvestigate, game.PrimitiveProliferate, game.PrimitiveManifest:
		return r.renderStandalonePrimitive(ctx, primitive)
	case game.PrimitiveDestroy, game.PrimitiveBounce, game.PrimitiveUntap,
		game.PrimitiveExile:
		return r.renderObjectOrGroupPrimitive(ctx, primitive)
	case game.PrimitiveTap, game.PrimitiveRegenerate, game.PrimitiveExplore:
		return r.renderObjectPrimitive(primitive)
	case game.PrimitiveAddMana:
		value, ok := primitive.(game.AddMana)
		if !ok {
			return "", errors.New("render: internal error: AddMana kind has unexpected concrete type")
		}
		return r.renderAddMana(ctx, &value)
	case game.PrimitiveAddCounter:
		value, ok := primitive.(game.AddCounter)
		if !ok {
			return "", errors.New("render: internal error: AddCounter kind has unexpected concrete type")
		}
		return r.renderAddCounter(ctx, &value)
	case game.PrimitiveAddPlayerCounter:
		value, ok := primitive.(game.AddPlayerCounter)
		if !ok {
			return "", errors.New("render: internal error: AddPlayerCounter kind has unexpected concrete type")
		}
		return r.renderAddPlayerCounter(ctx, &value)
	case game.PrimitiveModifyPT:
		value, ok := primitive.(game.ModifyPT)
		if !ok {
			return "", errors.New("render: internal error: ModifyPT kind has unexpected concrete type")
		}
		return r.renderModifyPT(ctx, &value)
	case game.PrimitiveFight:
		return r.renderFightPrimitive(primitive)
	case game.PrimitiveChoose:
		value, ok := primitive.(game.Choose)
		if !ok {
			return "", errors.New("render: internal error: Choose kind has unexpected concrete type")
		}
		return r.renderChoose(ctx, value)
	case game.PrimitivePutOnBattlefield:
		value, ok := primitive.(game.PutOnBattlefield)
		if !ok {
			return "", errors.New("render: internal error: PutOnBattlefield kind has unexpected concrete type")
		}
		return r.renderPutOnBattlefield(ctx, value)
	case game.PrimitiveMoveCard:
		value, ok := primitive.(game.MoveCard)
		if !ok {
			return "", errors.New("render: internal error: MoveCard kind has unexpected concrete type")
		}
		return r.renderMoveCard(ctx, value)
	case game.PrimitiveGrantCastPermission:
		value, ok := primitive.(game.GrantCastPermission)
		if !ok {
			return "", errors.New("render: internal error: GrantCastPermission kind has unexpected concrete type")
		}
		return r.renderGrantCastPermission(ctx, value)
	default:
		return "", fmt.Errorf("render: unsupported primitive kind %d", primitive.Kind())
	}
}

func (r Renderer) renderPutOnBattlefield(ctx *renderCtx, value game.PutOnBattlefield) (string, error) {
	source, err := renderBattlefieldSource(value.Source)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Source: %s,", source)}
	if value.Recipient.Exists {
		recipient, err := r.renderPlayerReference(value.Recipient.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Recipient: opt.Val(%s),", recipient))
	}
	if len(value.ContinuousEffects) > 0 {
		return "", errors.New("render: unsupported PutOnBattlefield continuous effects")
	}
	if value.EntryTapped {
		fields = append(fields, "EntryTapped: true,")
	}
	if len(value.EntryCounters) > 0 {
		counters, err := renderCounterPlacements(ctx, value.EntryCounters)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("EntryCounters: []game.CounterPlacement{%s},", strings.Join(counters, ", ")))
	}
	return structLit("game.PutOnBattlefield", fields), nil
}

func renderBattlefieldSource(source game.BattlefieldSource) (string, error) {
	if ref, ok := source.CardRef(); ok {
		rendered, err := renderCardReference(ref)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.CardBattlefieldSource(%s)", rendered), nil
	}
	if key, ok := source.LinkedKey(); ok {
		return fmt.Sprintf("game.LinkedBattlefieldSource(game.LinkedKey(%q))", string(key)), nil
	}
	return "", errors.New("render: unsupported battlefield source")
}

func (Renderer) renderMoveCard(ctx *renderCtx, value game.MoveCard) (string, error) {
	card, err := renderCardReference(value.Card)
	if err != nil {
		return "", err
	}
	fromZone, err := renderZone(value.FromZone)
	if err != nil {
		return "", err
	}
	destination, err := renderZone(value.Destination)
	if err != nil {
		return "", err
	}
	ctx.need(importZone)
	fields := []string{
		fmt.Sprintf("Card: %s,", card),
		fmt.Sprintf("FromZone: %s,", fromZone),
		fmt.Sprintf("Destination: %s,", destination),
	}
	if value.DestinationBottom {
		fields = append(fields, "DestinationBottom: true,")
	}
	return structLit("game.MoveCard", fields), nil
}

func (Renderer) renderGrantCastPermission(ctx *renderCtx, value game.GrantCastPermission) (string, error) {
	card, err := renderCardReference(value.Card)
	if err != nil {
		return "", err
	}
	fromZone, err := renderZone(value.FromZone)
	if err != nil {
		return "", err
	}
	duration, err := renderDuration(value.Duration)
	if err != nil {
		return "", err
	}
	if value.Face != game.FaceAlternate {
		return "", fmt.Errorf("render: unsupported cast-permission face %d", value.Face)
	}
	ctx.need(importZone)
	return structLit("game.GrantCastPermission", []string{
		fmt.Sprintf("Card: %s,", card),
		fmt.Sprintf("FromZone: %s,", fromZone),
		"Face: game.FaceAlternate,",
		fmt.Sprintf("Duration: %s,", duration),
	}), nil
}

func renderCardReference(reference game.CardReference) (string, error) {
	switch reference.Kind {
	case game.CardReferenceEvent:
		if reference.LinkID != "" {
			return "", errors.New("render: event card reference has LinkID")
		}
		return "game.CardReference{Kind: game.CardReferenceEvent}", nil
	case game.CardReferenceSource:
		if reference.LinkID != "" {
			return "", errors.New("render: source card reference has LinkID")
		}
		return "game.CardReference{Kind: game.CardReferenceSource}", nil
	case game.CardReferenceTarget:
		if reference.LinkID != "" {
			return "", errors.New("render: target card reference has LinkID")
		}
		if reference.TargetIndex < 0 {
			return "", errors.New("render: target card reference has negative TargetIndex")
		}
		if reference.TargetIndex != 0 {
			return fmt.Sprintf("game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: %d}", reference.TargetIndex), nil
		}
		return "game.CardReference{Kind: game.CardReferenceTarget}", nil
	case game.CardReferenceLinked:
		if reference.LinkID == "" {
			return "", errors.New("render: linked card reference has no LinkID")
		}
		return fmt.Sprintf("game.CardReference{Kind: game.CardReferenceLinked, LinkID: %q}", reference.LinkID), nil
	default:
		return "", fmt.Errorf("render: unsupported card reference kind %d", reference.Kind)
	}
}

func (r Renderer) renderAddCounter(ctx *renderCtx, value *game.AddCounter) (string, error) {
	object, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	kind, err := renderCounterKind(value.CounterKind)
	if err != nil {
		return "", err
	}
	ctx.need(importCounter)
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	return structLit("game.AddCounter", []string{
		fmt.Sprintf("Amount: %s,", amount),
		fmt.Sprintf("Object: %s,", object),
		fmt.Sprintf("CounterKind: %s,", kind),
	}), nil
}

func (r Renderer) renderAddPlayerCounter(ctx *renderCtx, value *game.AddPlayerCounter) (string, error) {
	player, err := r.renderPlayerReference(value.Player)
	if err != nil {
		return "", err
	}
	kind, err := renderCounterKind(value.CounterKind)
	if err != nil {
		return "", err
	}
	ctx.need(importCounter)
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	return structLit("game.AddPlayerCounter", []string{
		fmt.Sprintf("Amount: %s,", amount),
		fmt.Sprintf("Player: %s,", player),
		fmt.Sprintf("CounterKind: %s,", kind),
	}), nil
}

func (r Renderer) renderDamagePrimitive(ctx *renderCtx, primitive game.Primitive) (string, error) {
	value, ok := primitive.(game.Damage)
	if !ok {
		return "", errors.New("render: internal error: Damage kind has unexpected concrete type")
	}
	recipient, err := r.renderDamageRecipient(ctx, value.Recipient)
	if err != nil {
		return "", err
	}
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Amount: %s,", amount),
		fmt.Sprintf("Recipient: %s,", recipient),
	}
	if value.DamageSource.Exists {
		source, err := r.renderObjectReference(value.DamageSource.Val)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("DamageSource: opt.Val(%s),", source))
		ctx.need(importOpt)
	}
	return structLit("game.Damage", fields), nil
}

func (r Renderer) renderPlayerAmountPrimitive(ctx *renderCtx, primitive game.Primitive) (string, error) {
	var typeName string
	var amount game.Quantity
	var player game.PlayerReference
	switch primitive.Kind() {
	case game.PrimitiveDraw:
		value, ok := primitive.(game.Draw)
		if !ok {
			return "", errors.New("render: internal error: Draw kind has unexpected concrete type")
		}
		typeName, amount, player = "game.Draw", value.Amount, value.Player
	case game.PrimitiveDiscard:
		value, ok := primitive.(game.Discard)
		if !ok {
			return "", errors.New("render: internal error: Discard kind has unexpected concrete type")
		}
		typeName, amount, player = "game.Discard", value.Amount, value.Player
	case game.PrimitiveMill:
		value, ok := primitive.(game.Mill)
		if !ok {
			return "", errors.New("render: internal error: Mill kind has unexpected concrete type")
		}
		typeName, amount, player = "game.Mill", value.Amount, value.Player
	case game.PrimitiveScry:
		value, ok := primitive.(game.Scry)
		if !ok {
			return "", errors.New("render: internal error: Scry kind has unexpected concrete type")
		}
		typeName, amount, player = "game.Scry", value.Amount, value.Player
	case game.PrimitiveSurveil:
		value, ok := primitive.(game.Surveil)
		if !ok {
			return "", errors.New("render: internal error: Surveil kind has unexpected concrete type")
		}
		typeName, amount, player = "game.Surveil", value.Amount, value.Player
	case game.PrimitiveGainLife:
		value, ok := primitive.(game.GainLife)
		if !ok {
			return "", errors.New("render: internal error: GainLife kind has unexpected concrete type")
		}
		typeName, amount, player = "game.GainLife", value.Amount, value.Player
	case game.PrimitiveLoseLife:
		value, ok := primitive.(game.LoseLife)
		if !ok {
			return "", errors.New("render: internal error: LoseLife kind has unexpected concrete type")
		}
		typeName, amount, player = "game.LoseLife", value.Amount, value.Player
	default:
		return "", fmt.Errorf("render: unsupported player amount primitive kind %d", primitive.Kind())
	}
	rendered, err := r.renderPlayerReference(player)
	if err != nil {
		return "", err
	}
	return r.renderAmountPlayer(ctx, typeName, amount, rendered)
}

func (r Renderer) renderStandalonePrimitive(ctx *renderCtx, primitive game.Primitive) (string, error) {
	switch primitive.Kind() {
	case game.PrimitiveInvestigate:
		value, ok := primitive.(game.Investigate)
		if !ok {
			return "", errors.New("render: internal error: Investigate kind has unexpected concrete type")
		}
		amount, err := r.renderQuantity(ctx, value.Amount)
		if err != nil {
			return "", err
		}
		return structLit("game.Investigate", []string{fmt.Sprintf("Amount: %s,", amount)}), nil
	case game.PrimitiveProliferate:
		value, ok := primitive.(game.Proliferate)
		if !ok {
			return "", errors.New("render: internal error: Proliferate kind has unexpected concrete type")
		}
		amount, err := r.renderQuantity(ctx, value.Amount)
		if err != nil {
			return "", err
		}
		return structLit("game.Proliferate", []string{fmt.Sprintf("Amount: %s,", amount)}), nil
	case game.PrimitiveManifest:
		value, ok := primitive.(game.Manifest)
		if !ok {
			return "", errors.New("render: internal error: Manifest kind has unexpected concrete type")
		}
		var fields []string
		if value.Dread {
			fields = append(fields, "Dread: true,")
		}
		return structLit("game.Manifest", fields), nil
	default:
		return "", fmt.Errorf("render: unsupported standalone primitive kind %d", primitive.Kind())
	}
}

func (r Renderer) renderObjectOrGroupPrimitive(ctx *renderCtx, primitive game.Primitive) (string, error) {
	switch primitive.Kind() {
	case game.PrimitiveDestroy:
		value, ok := primitive.(game.Destroy)
		if !ok {
			return "", errors.New("render: internal error: Destroy kind has unexpected concrete type")
		}
		return r.renderObjectOrGroup(ctx, "game.Destroy", value.Object, value.Group)
	case game.PrimitiveBounce:
		value, ok := primitive.(game.Bounce)
		if !ok {
			return "", errors.New("render: internal error: Bounce kind has unexpected concrete type")
		}
		return r.renderObjectOrGroup(ctx, "game.Bounce", value.Object, value.Group)
	case game.PrimitiveUntap:
		value, ok := primitive.(game.Untap)
		if !ok {
			return "", errors.New("render: internal error: Untap kind has unexpected concrete type")
		}
		return r.renderObjectOrGroup(ctx, "game.Untap", value.Object, value.Group)
	case game.PrimitiveExile:
		value, ok := primitive.(game.Exile)
		if !ok {
			return "", errors.New("render: internal error: Exile kind has unexpected concrete type")
		}
		return r.renderObjectOrGroup(ctx, "game.Exile", value.Object, value.Group)
	default:
		return "", fmt.Errorf("render: unsupported object or group primitive kind %d", primitive.Kind())
	}
}

func (r Renderer) renderObjectPrimitive(primitive game.Primitive) (string, error) {
	var typeName string
	fieldName := "Object"
	var object game.ObjectReference
	switch primitive.Kind() {
	case game.PrimitiveTap:
		value, ok := primitive.(game.Tap)
		if !ok {
			return "", errors.New("render: internal error: Tap kind has unexpected concrete type")
		}
		typeName, object = "game.Tap", value.Object
	case game.PrimitiveRegenerate:
		value, ok := primitive.(game.Regenerate)
		if !ok {
			return "", errors.New("render: internal error: Regenerate kind has unexpected concrete type")
		}
		typeName, object = "game.Regenerate", value.Object
	case game.PrimitiveExplore:
		value, ok := primitive.(game.Explore)
		if !ok {
			return "", errors.New("render: internal error: Explore kind has unexpected concrete type")
		}
		fieldName = "Creature"
		typeName, object = "game.Explore", value.Creature
	default:
		return "", fmt.Errorf("render: unsupported object primitive kind %d", primitive.Kind())
	}
	rendered, err := r.renderObjectReference(object)
	if err != nil {
		return "", err
	}
	return structLit(typeName, []string{fmt.Sprintf("%s: %s,", fieldName, rendered)}), nil
}

func (r Renderer) renderFightPrimitive(primitive game.Primitive) (string, error) {
	value, ok := primitive.(game.Fight)
	if !ok {
		return "", errors.New("render: internal error: Fight kind has unexpected concrete type")
	}
	object, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	related, err := r.renderObjectReference(value.RelatedObject)
	if err != nil {
		return "", err
	}
	return structLit("game.Fight", []string{
		fmt.Sprintf("Object: %s,", object),
		fmt.Sprintf("RelatedObject: %s,", related),
	}), nil
}

func (r Renderer) renderAmountPlayer(
	ctx *renderCtx,
	typeName string,
	amount game.Quantity,
	player string,
) (string, error) {
	renderedAmount, err := r.renderQuantity(ctx, amount)
	if err != nil {
		return "", err
	}
	return structLit(typeName, []string{
		fmt.Sprintf("Amount: %s,", renderedAmount),
		fmt.Sprintf("Player: %s,", player),
	}), nil
}

func (r Renderer) renderObjectOrGroup(ctx *renderCtx, typeName string, object game.ObjectReference, group game.GroupReference) (string, error) {
	if group.Domain() != 0 {
		rendered, err := r.renderGroupReference(ctx, group)
		if err != nil {
			return "", err
		}
		return structLit(typeName, []string{fmt.Sprintf("Group: %s,", rendered)}), nil
	}
	rendered, err := r.renderObjectReference(object)
	if err != nil {
		return "", err
	}
	return structLit(typeName, []string{fmt.Sprintf("Object: %s,", rendered)}), nil
}

func (r Renderer) renderAddMana(ctx *renderCtx, value *game.AddMana) (string, error) {
	amount, err := r.renderQuantity(ctx, value.Amount)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Amount: %s,", amount)}
	if value.ManaColor != "" {
		ctx.need(importMana)
		colorLiteral, err := renderManaColor(value.ManaColor)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ManaColor: %s,", colorLiteral))
	}
	if value.ChoiceFrom != "" {
		fields = append(fields, fmt.Sprintf("ChoiceFrom: game.ChoiceKey(%q),", string(value.ChoiceFrom)))
	}
	return structLit("game.AddMana", fields), nil
}

func (r Renderer) renderModifyPT(ctx *renderCtx, value *game.ModifyPT) (string, error) {
	object, err := r.renderObjectReference(value.Object)
	if err != nil {
		return "", err
	}
	duration, err := renderDuration(value.Duration)
	if err != nil {
		return "", err
	}
	power, err := r.renderQuantity(ctx, value.PowerDelta)
	if err != nil {
		return "", err
	}
	toughness, err := r.renderQuantity(ctx, value.ToughnessDelta)
	if err != nil {
		return "", err
	}
	return structLit("game.ModifyPT", []string{
		fmt.Sprintf("Object: %s,", object),
		fmt.Sprintf("PowerDelta: %s,", power),
		fmt.Sprintf("ToughnessDelta: %s,", toughness),
		fmt.Sprintf("Duration: %s,", duration),
	}), nil
}

func (r Renderer) renderChoose(ctx *renderCtx, value game.Choose) (string, error) {
	choice, err := r.renderResolutionChoice(ctx, value.Choice)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Choice: %s,", choice)}
	if value.PublishChoice != "" {
		fields = append(fields, fmt.Sprintf("PublishChoice: game.ChoiceKey(%q),", string(value.PublishChoice)))
	}
	return structLit("game.Choose", fields), nil
}

func (Renderer) renderResolutionChoice(ctx *renderCtx, choice game.ResolutionChoice) (string, error) {
	kind, err := renderResolutionChoiceKind(choice.Kind)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Kind: %s,", kind)}
	if len(choice.Colors) > 0 {
		ctx.need(importMana)
		colors, err := renderManaColorSlice(ctx, choice.Colors)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Colors: %s,", colors))
	}
	return structLit("game.ResolutionChoice", fields), nil
}

func (r Renderer) renderTargetSpec(ctx *renderCtx, spec *game.TargetSpec) (string, error) {
	fields := []string{
		fmt.Sprintf("MinTargets: %d,", spec.MinTargets),
		fmt.Sprintf("MaxTargets: %d,", spec.MaxTargets),
	}
	if spec.Constraint != "" {
		fields = append(fields, fmt.Sprintf("Constraint: %q,", spec.Constraint))
	}
	if spec.Allow != game.TargetAllowUnspecified {
		fields = append(fields, fmt.Sprintf("Allow: %s,", renderTargetAllow(spec.Allow)))
	}
	if spec.TargetZone != zone.None {
		targetZone, err := renderZone(spec.TargetZone)
		if err != nil {
			return "", err
		}
		ctx.need(importZone)
		fields = append(fields, fmt.Sprintf("TargetZone: %s,", targetZone))
	}
	if spec.Selection.Exists {
		selection, err := r.renderSelection(ctx, spec.Selection.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Selection: opt.Val(%s),", selection))
	}
	if predicate, ok, err := r.renderTargetPredicate(ctx, spec.Predicate); err != nil {
		return "", err
	} else if ok {
		fields = append(fields, fmt.Sprintf("Predicate: %s,", predicate))
	}
	return structLit("game.TargetSpec", fields), nil
}

func (Renderer) renderTargetPredicate(ctx *renderCtx, predicate game.TargetPredicate) (lit string, ok bool, err error) {
	var fields []string
	if len(predicate.PermanentTypes) > 0 {
		ctx.need(importTypes)
		lits, err := renderTypesCardSlice(ctx, predicate.PermanentTypes)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("PermanentTypes: %s,", lits))
	}
	if len(predicate.ExcludedTypes) > 0 {
		lits, err := renderTypesCardSlice(ctx, predicate.ExcludedTypes)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("ExcludedTypes: %s,", lits))
	}
	if len(predicate.Colors) > 0 {
		colors, err := renderColorSlice(ctx, predicate.Colors)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("Colors: %s,", colors))
	}
	if len(predicate.ExcludedColors) > 0 {
		colors, err := renderColorSlice(ctx, predicate.ExcludedColors)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("ExcludedColors: %s,", colors))
	}
	if predicate.Player != game.PlayerAny {
		pr, err := renderPlayerRelation(predicate.Player)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("Player: %s,", pr))
	}
	if predicate.Controller != game.ControllerAny {
		cr, err := renderControllerRelation(predicate.Controller)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("Controller: %s,", cr))
	}
	if predicate.Tapped != game.TriAny {
		ts, err := renderTriState(predicate.Tapped)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("Tapped: %s,", ts))
	}
	if predicate.CombatState != game.CombatStateAny {
		cs, err := renderCombatStateFilter(predicate.CombatState)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("CombatState: %s,", cs))
	}
	if predicate.Keyword != game.KeywordNone {
		kw, err := renderKeyword(predicate.Keyword)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("Keyword: %s,", kw))
	}
	if predicate.ExcludedKeyword != game.KeywordNone {
		kw, err := renderKeyword(predicate.ExcludedKeyword)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("ExcludedKeyword: %s,", kw))
	}
	if predicate.ManaValue.Exists {
		ctx.need(importOpt)
		cmp, err := renderCompareInt(ctx, predicate.ManaValue.Val)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("ManaValue: opt.Val(%s),", cmp))
	}
	if predicate.Power.Exists {
		ctx.need(importOpt)
		cmp, err := renderCompareInt(ctx, predicate.Power.Val)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("Power: opt.Val(%s),", cmp))
	}
	if predicate.Toughness.Exists {
		ctx.need(importOpt)
		cmp, err := renderCompareInt(ctx, predicate.Toughness.Val)
		if err != nil {
			return "", false, err
		}
		fields = append(fields, fmt.Sprintf("Toughness: opt.Val(%s),", cmp))
	}
	if predicate.Another {
		fields = append(fields, "Another: true,")
	}
	if len(fields) == 0 {
		return "", false, nil
	}
	return structLit("game.TargetPredicate", fields), true, nil
}

func (r Renderer) renderGroupReference(ctx *renderCtx, group game.GroupReference) (string, error) {
	selection, err := r.renderSelection(ctx, group.Selection())
	if err != nil {
		return "", err
	}
	exclude, hasExclude := group.Exclusion()
	switch group.Domain() {
	case game.GroupDomainBattlefield:
		if hasExclude {
			rendered, err := r.renderObjectReference(exclude)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("game.BattlefieldGroupExcluding(%s, %s)", selection, rendered), nil
		}
		return fmt.Sprintf("game.BattlefieldGroup(%s)", selection), nil
	case game.GroupDomainAttachedObject:
		anchor, _ := group.Anchor()
		rendered, err := r.renderObjectReference(anchor)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.AttachedObjectGroup(%s)", rendered), nil
	case game.GroupDomainObjectControlled:
		anchor, _ := group.Anchor()
		renderedAnchor, err := r.renderObjectReference(anchor)
		if err != nil {
			return "", err
		}
		if hasExclude {
			renderedExclude, err := r.renderObjectReference(exclude)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("game.ObjectControlledGroupExcluding(%s, %s, %s)", renderedAnchor, selection, renderedExclude), nil
		}
		return fmt.Sprintf("game.ObjectControlledGroup(%s, %s)", renderedAnchor, selection), nil
	default:
		return "", fmt.Errorf("render: unsupported group reference domain %d", group.Domain())
	}
}

func (Renderer) renderSelection(ctx *renderCtx, selection game.Selection) (string, error) {
	var fields []string

	if len(selection.RequiredTypes) > 0 {
		ctx.need(importTypes)
		lits, err := renderTypesCardSlice(ctx, selection.RequiredTypes)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("RequiredTypes: %s,", lits))
	}
	if len(selection.RequiredTypesAny) > 0 {
		ctx.need(importTypes)
		lits, err := renderTypesCardSlice(ctx, selection.RequiredTypesAny)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("RequiredTypesAny: %s,", lits))
	}
	if len(selection.ExcludedTypes) > 0 {
		ctx.need(importTypes)
		lits, err := renderTypesCardSlice(ctx, selection.ExcludedTypes)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ExcludedTypes: %s,", lits))
	}

	if len(selection.Supertypes) > 0 {
		ctx.need(importTypes)
		literals := make([]string, 0, len(selection.Supertypes))
		for _, st := range selection.Supertypes {
			lit, err := supertypeLiteral(st)
			if err != nil {
				return "", err
			}
			literals = append(literals, lit)
		}
		fields = append(fields, fmt.Sprintf("Supertypes: []types.Super{%s},", strings.Join(literals, ", ")))
	}
	if len(selection.SubtypesAny) > 0 {
		ctx.need(importTypes)
		literals := make([]string, 0, len(selection.SubtypesAny))
		for _, sub := range selection.SubtypesAny {
			literals = append(literals, SubtypeToLiteral(string(sub), nil))
		}
		fields = append(fields, fmt.Sprintf("SubtypesAny: []types.Sub{%s},", strings.Join(literals, ", ")))
	}

	if len(selection.ColorsAny) > 0 {
		colorLits, err := renderColorSlice(ctx, selection.ColorsAny)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ColorsAny: %s,", colorLits))
	}
	if len(selection.ExcludedColors) > 0 {
		colorLits, err := renderColorSlice(ctx, selection.ExcludedColors)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ExcludedColors: %s,", colorLits))
	}
	if selection.Colorless {
		fields = append(fields, "Colorless: true,")
	}
	if selection.Multicolored {
		fields = append(fields, "Multicolored: true,")
	}

	if selection.Controller != game.ControllerAny {
		cr, err := renderControllerRelation(selection.Controller)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Controller: %s,", cr))
	}
	if selection.Player != game.PlayerAny {
		pr, err := renderPlayerRelation(selection.Player)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Player: %s,", pr))
	}

	if selection.Tapped != game.TriAny {
		ts, err := renderTriState(selection.Tapped)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Tapped: %s,", ts))
	}
	if selection.CombatState != game.CombatStateAny {
		cs, err := renderCombatStateFilter(selection.CombatState)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("CombatState: %s,", cs))
	}

	if selection.Keyword != game.KeywordNone {
		kw, err := renderKeyword(selection.Keyword)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Keyword: %s,", kw))
	}
	if selection.ExcludedKeyword != game.KeywordNone {
		kw, err := renderKeyword(selection.ExcludedKeyword)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ExcludedKeyword: %s,", kw))
	}

	if selection.ManaValue.Exists {
		ctx.need(importOpt)
		cmp, err := renderCompareInt(ctx, selection.ManaValue.Val)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ManaValue: opt.Val(%s),", cmp))
	}
	if selection.Power.Exists {
		ctx.need(importOpt)
		cmp, err := renderCompareInt(ctx, selection.Power.Val)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Power: opt.Val(%s),", cmp))
	}
	if selection.Toughness.Exists {
		ctx.need(importOpt)
		cmp, err := renderCompareInt(ctx, selection.Toughness.Val)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Toughness: opt.Val(%s),", cmp))
	}

	if selection.ExcludeSource {
		fields = append(fields, "ExcludeSource: true,")
	}
	if selection.NonToken {
		fields = append(fields, "NonToken: true,")
	}
	if selection.TokenOnly {
		fields = append(fields, "TokenOnly: true,")
	}

	for i := range fields {
		fields[i] = strings.TrimSuffix(fields[i], ",")
	}
	return compactStructLit("game.Selection", fields), nil
}

func renderColorSlice(ctx *renderCtx, colors []color.Color) (string, error) {
	ctx.need(importColor)
	literals := make([]string, 0, len(colors))
	for _, c := range colors {
		lit, err := colorValueToLiteral(c)
		if err != nil {
			return "", err
		}
		literals = append(literals, lit)
	}
	return "[]color.Color{" + strings.Join(literals, ", ") + "}", nil
}

func renderColorArguments(ctx *renderCtx, colors []color.Color) (string, error) {
	ctx.need(importColor)
	literals := make([]string, 0, len(colors))
	seen := make(map[color.Color]struct{}, len(colors))
	for _, c := range colors {
		if _, ok := seen[c]; ok {
			return "", fmt.Errorf("render: duplicate color %q", c)
		}
		seen[c] = struct{}{}
		literal, err := colorValueToLiteral(c)
		if err != nil {
			return "", err
		}
		literals = append(literals, literal)
	}
	return strings.Join(literals, ", "), nil
}

func renderControllerRelation(cr game.ControllerRelation) (string, error) {
	switch cr {
	case game.ControllerAny:
		return "game.ControllerAny", nil
	case game.ControllerYou:
		return "game.ControllerYou", nil
	case game.ControllerOpponent:
		return "game.ControllerOpponent", nil
	case game.ControllerNotYou:
		return "game.ControllerNotYou", nil
	default:
		return "", fmt.Errorf("render: unsupported controller relation %d", cr)
	}
}

func renderTriState(ts game.TriState) (string, error) {
	switch ts {
	case game.TriAny:
		return "game.TriAny", nil
	case game.TriTrue:
		return "game.TriTrue", nil
	case game.TriFalse:
		return "game.TriFalse", nil
	default:
		return "", fmt.Errorf("render: unsupported tri-state %d", ts)
	}
}

func renderCombatStateFilter(cs game.CombatStateFilter) (string, error) {
	switch cs {
	case game.CombatStateAny:
		return "game.CombatStateAny", nil
	case game.CombatStateAttacking:
		return "game.CombatStateAttacking", nil
	case game.CombatStateBlocking:
		return "game.CombatStateBlocking", nil
	case game.CombatStateAttackingOrBlocking:
		return "game.CombatStateAttackingOrBlocking", nil
	default:
		return "", fmt.Errorf("render: unsupported combat state filter %d", cs)
	}
}

func renderKeyword(kw game.Keyword) (string, error) {
	switch kw {
	case game.KeywordNone:
		return "game.KeywordNone", nil
	case game.Devoid:
		return "game.Devoid", nil
	case game.Deathtouch:
		return "game.Deathtouch", nil
	case game.Defender:
		return "game.Defender", nil
	case game.DoubleStrike:
		return "game.DoubleStrike", nil
	case game.FirstStrike:
		return "game.FirstStrike", nil
	case game.Flash:
		return "game.Flash", nil
	case game.Flying:
		return "game.Flying", nil
	case game.Haste:
		return "game.Haste", nil
	case game.Hexproof:
		return "game.Hexproof", nil
	case game.Indestructible:
		return "game.Indestructible", nil
	case game.Lifelink:
		return "game.Lifelink", nil
	case game.Menace:
		return "game.Menace", nil
	case game.Protection:
		return "game.Protection", nil
	case game.Reach:
		return "game.Reach", nil
	case game.Shroud:
		return "game.Shroud", nil
	case game.Trample:
		return "game.Trample", nil
	case game.Vigilance:
		return "game.Vigilance", nil
	case game.Ward:
		return "game.Ward", nil
	case game.SplitSecond:
		return "game.SplitSecond", nil
	case game.Equip:
		return "game.Equip", nil
	case game.Enchant:
		return "game.Enchant", nil
	case game.Cycling:
		return "game.Cycling", nil
	case game.Flashback:
		return "game.Flashback", nil
	case game.Kicker:
		return "game.Kicker", nil
	case game.Madness:
		return "game.Madness", nil
	case game.Morph:
		return "game.Morph", nil
	case game.Disguise:
		return "game.Disguise", nil
	case game.Convoke:
		return "game.Convoke", nil
	case game.Delve:
		return "game.Delve", nil
	case game.Suspend:
		return "game.Suspend", nil
	case game.Storm:
		return "game.Storm", nil
	case game.Cascade:
		return "game.Cascade", nil
	case game.Prowess:
		return "game.Prowess", nil
	case game.Mutate:
		return "game.Mutate", nil
	case game.Companion:
		return "game.Companion", nil
	case game.Ninjutsu:
		return "game.Ninjutsu", nil
	case game.Escape:
		return "game.Escape", nil
	case game.Foretell:
		return "game.Foretell", nil
	case game.Craft:
		return "game.Craft", nil
	case game.Discover:
		return "game.Discover", nil
	case game.Eternalize:
		return "game.Eternalize", nil
	case game.Affinity:
		return "game.Affinity", nil
	case game.Improvise:
		return "game.Improvise", nil
	case game.Emerge:
		return "game.Emerge", nil
	case game.Undying:
		return "game.Undying", nil
	case game.Persist:
		return "game.Persist", nil
	case game.Wither:
		return "game.Wither", nil
	case game.Infect:
		return "game.Infect", nil
	case game.Toxic:
		return "game.Toxic", nil
	case game.Annihilator:
		return "game.Annihilator", nil
	case game.Exalted:
		return "game.Exalted", nil
	case game.ReadAhead:
		return "game.ReadAhead", nil
	default:
		return "", fmt.Errorf("render: unsupported keyword %d", kw)
	}
}

func renderCompareInt(ctx *renderCtx, cmp compare.Int) (string, error) {
	ctx.need(importCompare)
	op, err := renderCompareOp(cmp.Op)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("compare.Int{Op: %s, Value: %d}", op, cmp.Value), nil
}

func renderCompareOp(op compare.Op) (string, error) {
	switch op {
	case compare.Any:
		return "compare.Any", nil
	case compare.Equal:
		return "compare.Equal", nil
	case compare.LessOrEqual:
		return "compare.LessOrEqual", nil
	case compare.GreaterOrEqual:
		return "compare.GreaterOrEqual", nil
	case compare.LessThan:
		return "compare.LessThan", nil
	case compare.GreaterThan:
		return "compare.GreaterThan", nil
	default:
		return "", fmt.Errorf("render: unsupported compare op %d", op)
	}
}

func (r Renderer) renderDamageRecipient(ctx *renderCtx, recipient game.DamageRecipient) (string, error) {
	if object, ok := recipient.AnyTargetObjectReference(); ok {
		return fmt.Sprintf("game.AnyTargetDamageRecipient(%d)", object.TargetIndex()), nil
	}
	if object, ok := recipient.ObjectReference(); ok {
		rendered, err := r.renderObjectReference(object)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.ObjectDamageRecipient(%s)", rendered), nil
	}
	if player, ok := recipient.PlayerReference(); ok {
		rendered, err := r.renderPlayerReference(player)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.PlayerDamageRecipient(%s)", rendered), nil
	}
	if group, ok := recipient.GroupReference(); ok {
		rendered, err := r.renderGroupReference(ctx, group)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.GroupDamageRecipient(%s)", rendered), nil
	}
	return "", errors.New("render: unsupported damage recipient")
}

func (Renderer) renderObjectReference(reference game.ObjectReference) (string, error) {
	switch reference.Kind() {
	case game.ObjectReferenceTargetPermanent:
		return fmt.Sprintf("game.TargetPermanentReference(%d)", reference.TargetIndex()), nil
	case game.ObjectReferenceTargetStackObject:
		return fmt.Sprintf("game.TargetStackObjectReference(%d)", reference.TargetIndex()), nil
	case game.ObjectReferenceSourcePermanent:
		return "game.SourcePermanentReference()", nil
	case game.ObjectReferenceSourceAttachedPermanent:
		return "game.SourceAttachedPermanentReference()", nil
	case game.ObjectReferenceTargetAttachedPermanent:
		return fmt.Sprintf("game.TargetAttachedPermanentReference(%d)", reference.TargetIndex()), nil
	case game.ObjectReferenceLinkedObject:
		return fmt.Sprintf("game.LinkedObjectReference(%q)", reference.LinkID()), nil
	case game.ObjectReferenceEventPermanent:
		return "game.EventPermanentReference()", nil
	default:
		return "", fmt.Errorf("render: unsupported object reference kind %d", reference.Kind())
	}
}

func (r Renderer) renderPlayerReference(reference game.PlayerReference) (string, error) {
	switch reference.Kind() {
	case game.PlayerReferenceController:
		return "game.ControllerReference()", nil
	case game.PlayerReferenceTargetPlayer:
		return fmt.Sprintf("game.TargetPlayerReference(%d)", reference.TargetIndex()), nil
	case game.PlayerReferenceObjectController:
		object, _ := reference.Object()
		rendered, err := r.renderObjectReference(object)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.ObjectControllerReference(%s)", rendered), nil
	case game.PlayerReferenceObjectOwner:
		object, _ := reference.Object()
		rendered, err := r.renderObjectReference(object)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.ObjectOwnerReference(%s)", rendered), nil
	default:
		return "", fmt.Errorf("render: unsupported player reference kind %d", reference.Kind())
	}
}

func (r Renderer) renderKeywordAbility(ctx *renderCtx, keyword game.KeywordAbility) (string, error) {
	if ward, ok := keyword.(game.WardKeyword); ok {
		wardCost, err := r.renderManaCost(ctx, ward.Cost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.WardKeyword{Cost: %s}", wardCost), nil
	}
	if cycling, ok := keyword.(game.CyclingKeyword); ok {
		cyclingCost, err := r.renderManaCost(ctx, cycling.Cost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.CyclingKeyword{Cost: %s}", cyclingCost), nil
	}
	if ninjutsu, ok := keyword.(game.NinjutsuKeyword); ok {
		ninjutsuCost, err := r.renderManaCost(ctx, ninjutsu.Cost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.NinjutsuKeyword{Cost: %s}", ninjutsuCost), nil
	}
	if mutate, ok := keyword.(game.MutateKeyword); ok {
		mutateCost, err := r.renderManaCost(ctx, mutate.Cost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.MutateKeyword{Cost: %s}", mutateCost), nil
	}
	if kicker, ok := keyword.(game.KickerKeyword); ok {
		kickerCost, err := r.renderManaCost(ctx, kicker.Cost)
		if err != nil {
			return "", err
		}
		if len(kicker.BonusContent.Modes) != 0 {
			return "", errors.New("render: Kicker bonus content must be rendered by its owning ability")
		}
		return fmt.Sprintf("game.KickerKeyword{Cost: %s}", kickerCost), nil
	}
	if madness, ok := keyword.(game.MadnessKeyword); ok {
		madnessCost, err := r.renderManaCost(ctx, madness.Cost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.MadnessKeyword{Cost: %s}", madnessCost), nil
	}
	if morph, ok := keyword.(game.MorphKeyword); ok {
		morphCost, err := r.renderManaCost(ctx, morph.Cost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.MorphKeyword{Cost: %s}", morphCost), nil
	}
	if disguise, ok := keyword.(game.DisguiseKeyword); ok {
		disguiseCost, err := r.renderManaCost(ctx, disguise.Cost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.DisguiseKeyword{Cost: %s}", disguiseCost), nil
	}
	if toxic, ok := keyword.(game.ToxicKeyword); ok {
		return fmt.Sprintf("game.ToxicKeyword{Amount: %d}", toxic.Amount), nil
	}
	return "", fmt.Errorf("render: unsupported keyword ability %T", keyword)
}

func (Renderer) renderManaCost(ctx *renderCtx, manaCost cost.Mana) (string, error) {
	ctx.need(importCost)
	if len(manaCost) == 0 {
		return "cost.Mana{}", nil
	}
	symbols := make([]string, 0, len(manaCost))
	for _, symbol := range manaCost {
		sym, err := renderManaSymbol(ctx, symbol)
		if err != nil {
			return "", err
		}
		symbols = append(symbols, sym)
	}
	return "cost.Mana{" + strings.Join(symbols, ", ") + "}", nil
}

// renderManaCostMultiline renders a printed face ManaCost as a multi-line
// cost.Mana literal so gofmt preserves the canonical generated-card layout.
func renderManaCostMultiline(ctx *renderCtx, manaCost cost.Mana) (string, error) {
	ctx.need(importCost)
	if len(manaCost) == 0 {
		return "cost.Mana{}", nil
	}
	symbols := make([]string, 0, len(manaCost))
	for _, symbol := range manaCost {
		sym, err := renderManaSymbol(ctx, symbol)
		if err != nil {
			return "", err
		}
		symbols = append(symbols, sym)
	}
	return "cost.Mana{\n\t\t\t" + strings.Join(symbols, ",\n\t\t\t") + ",\n\t\t}", nil
}

func (Renderer) renderAdditionalCosts(ctx *renderCtx, costs []cost.Additional) (string, error) {
	ctx.need(importCost)
	if len(costs) == 1 &&
		costs[0].Kind == cost.AdditionalTap &&
		costs[0].Text == "" &&
		costs[0].Amount == 0 &&
		costs[0].Source == zone.None {
		return "cost.Tap", nil
	}
	elements := make([]string, 0, len(costs))
	for _, additional := range costs {
		rendered, err := renderAdditional(ctx, additional)
		if err != nil {
			return "", err
		}
		elements = append(elements, rendered+",")
	}
	return sliceLit("cost.Additional", elements), nil
}

func renderAdditional(ctx *renderCtx, additional cost.Additional) (string, error) {
	ctx.need(importCost)
	kind, err := renderAdditionalKind(additional.Kind)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Kind: %s,", kind)}
	if additional.Text != "" {
		fields = append(fields, fmt.Sprintf("Text: %q,", additional.Text))
	}
	if additional.Amount != 0 {
		fields = append(fields, fmt.Sprintf("Amount: %d,", additional.Amount))
	}
	if additional.AmountFromX {
		fields = append(fields, "AmountFromX: true,")
	}
	if additional.Source != zone.None {
		ctx.need(importZone)
		zoneLiteral, err := renderZone(additional.Source)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Source: %s,", zoneLiteral))
	}
	if additional.MatchPermanentType {
		cardType, err := cardTypeLiteral(additional.PermanentType)
		if err != nil {
			return "", err
		}
		ctx.need(importTypes)
		fields = append(fields,
			"MatchPermanentType: true,",
			fmt.Sprintf("PermanentType: %s,", cardType),
		)
	}
	if additional.MatchCardType {
		cardType, err := cardTypeLiteral(additional.CardType)
		if err != nil {
			return "", err
		}
		ctx.need(importTypes)
		fields = append(fields,
			"MatchCardType: true,",
			fmt.Sprintf("CardType: %s,", cardType),
		)
	}
	if additional.MatchCardColor {
		colorLiteral, err := colorValueToLiteral(additional.CardColor)
		if err != nil {
			return "", err
		}
		ctx.need(importColor)
		fields = append(fields,
			"MatchCardColor: true,",
			fmt.Sprintf("CardColor: %s,", colorLiteral),
		)
	}
	if additional.RequireTapped {
		fields = append(fields, "RequireTapped: true,")
	}
	if additional.RequireSupertype != "" {
		supertype, err := supertypeLiteral(additional.RequireSupertype)
		if err != nil {
			return "", err
		}
		ctx.need(importTypes)
		fields = append(fields, fmt.Sprintf("RequireSupertype: %s,", supertype))
	}
	if additional.SubtypesAny != (cost.SubtypeSet{}) {
		ctx.need(importTypes)
		literals := make([]string, 0, len(additional.SubtypesAny))
		for _, subtype := range additional.SubtypesAny {
			if subtype == "" {
				continue
			}
			literals = append(literals, SubtypeToLiteral(string(subtype), []string{"Land", "Creature"}))
		}
		fields = append(fields, fmt.Sprintf("SubtypesAny: cost.SubtypeSet{%s},", strings.Join(literals, ", ")))
	}
	if additional.Kind == cost.AdditionalRemoveCounter || additional.Kind == cost.AdditionalPutCounter {
		counterKind, err := renderCounterKind(additional.CounterKind)
		if err != nil {
			return "", err
		}
		ctx.need(importCounter)
		fields = append(fields, fmt.Sprintf("CounterKind: %s,", counterKind))
	}
	return structLit("", fields), nil
}

func (r Renderer) renderQuantity(ctx *renderCtx, quantity game.Quantity) (string, error) {
	dynamic := quantity.DynamicAmount()
	if !dynamic.Exists {
		return fmt.Sprintf("game.Fixed(%d)", quantity.Value()), nil
	}
	kind, err := renderDynamicAmountKind(dynamic.Val.Kind)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Kind: %s,", kind)}
	if dynamic.Val.Constant != 0 {
		fields = append(fields, fmt.Sprintf("Constant: %d,", dynamic.Val.Constant))
	}
	if dynamic.Val.Multiplier != 0 {
		fields = append(fields, fmt.Sprintf("Multiplier: %d,", dynamic.Val.Multiplier))
	}
	if dynamic.Val.Kind == game.DynamicAmountTargetCounters || dynamic.Val.CounterKind != 0 {
		counterKind, err := renderCounterKind(dynamic.Val.CounterKind)
		if err != nil {
			return "", err
		}
		ctx.need(importCounter)
		fields = append(fields, fmt.Sprintf("CounterKind: %s,", counterKind))
	}
	if !dynamic.Val.Group.Empty() {
		group, err := r.renderGroupReference(ctx, dynamic.Val.Group)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Group: %s,", group))
	}
	if dynamic.Val.Object.Kind() != game.ObjectReferenceNone {
		object, err := r.renderObjectReference(dynamic.Val.Object)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Object: %s,", object))
	}
	if dynamic.Val.Player != nil && dynamic.Val.Player.Kind() != game.PlayerReferenceNone {
		player, err := r.renderPlayerReference(*dynamic.Val.Player)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Player: func() *game.PlayerReference { ref := %s; return &ref }(),", player))
	}
	if dynamic.Val.CardZone != zone.None {
		cardZone, err := renderZone(dynamic.Val.CardZone)
		if err != nil {
			return "", err
		}
		ctx.need(importZone)
		fields = append(fields, fmt.Sprintf("CardZone: %s,", cardZone))
	}
	if dynamic.Val.Selection != nil && !dynamic.Val.Selection.Empty() {
		selection, err := r.renderSelection(ctx, *dynamic.Val.Selection)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Selection: &%s,", selection))
	}
	if dynamic.Val.ResultKey != "" {
		fields = append(fields, fmt.Sprintf("ResultKey: game.ResultKey(%q),", string(dynamic.Val.ResultKey)))
	}
	return fmt.Sprintf("game.Dynamic(%s)", structLit("game.DynamicAmount", fields)), nil
}

func renderDynamicAmountKind(kind game.DynamicAmountKind) (string, error) {
	switch kind {
	case game.DynamicAmountConstant:
		return "game.DynamicAmountConstant", nil
	case game.DynamicAmountX:
		return "game.DynamicAmountX", nil
	case game.DynamicAmountTargetPower:
		return "game.DynamicAmountTargetPower", nil
	case game.DynamicAmountTargetToughness:
		return "game.DynamicAmountTargetToughness", nil
	case game.DynamicAmountTargetManaValue:
		return "game.DynamicAmountTargetManaValue", nil
	case game.DynamicAmountTargetCounters:
		return "game.DynamicAmountTargetCounters", nil
	case game.DynamicAmountControllerLife:
		return "game.DynamicAmountControllerLife", nil
	case game.DynamicAmountControllerHandSize:
		return "game.DynamicAmountControllerHandSize", nil
	case game.DynamicAmountControllerGraveyardSize:
		return "game.DynamicAmountControllerGraveyardSize", nil
	case game.DynamicAmountCountSelector:
		return "game.DynamicAmountCountSelector", nil
	case game.DynamicAmountCountCardsInZone:
		return "game.DynamicAmountCountCardsInZone", nil
	case game.DynamicAmountPreviousEffectResult:
		return "game.DynamicAmountPreviousEffectResult", nil
	case game.DynamicAmountOpponentCount:
		return "game.DynamicAmountOpponentCount", nil
	case game.DynamicAmountEventDamage:
		return "game.DynamicAmountEventDamage", nil
	case game.DynamicAmountPreviousEffectExcessDamage:
		return "game.DynamicAmountPreviousEffectExcessDamage", nil
	case game.DynamicAmountObjectPower:
		return "game.DynamicAmountObjectPower", nil
	default:
		return "", fmt.Errorf("render: unsupported dynamic amount kind %d", kind)
	}
}

func renderManaSymbol(ctx *renderCtx, symbol cost.Symbol) (string, error) {
	ctx.need(importCost)
	switch symbol.Kind {
	case cost.ColoredSymbol:
		switch symbol.Color {
		case mana.W:
			return "cost.W", nil
		case mana.U:
			return "cost.U", nil
		case mana.B:
			return "cost.B", nil
		case mana.R:
			return "cost.R", nil
		case mana.G:
			return "cost.G", nil
		default:
			return "", fmt.Errorf("render: unsupported colored mana symbol %q", string(symbol.Color))
		}
	case cost.GenericSymbol:
		return fmt.Sprintf("cost.O(%d)", symbol.Generic), nil
	case cost.ColorlessSymbol:
		return "cost.C", nil
	case cost.VariableSymbol:
		return "cost.X", nil
	case cost.SnowSymbol:
		return "cost.S", nil
	case cost.HybridSymbol:
		ctx.need(importMana)
		first, err := renderManaColor(symbol.Color)
		if err != nil {
			return "", fmt.Errorf("render: unsupported hybrid mana color: %w", err)
		}
		second, err := renderManaColor(symbol.AltColor)
		if err != nil {
			return "", fmt.Errorf("render: unsupported hybrid mana alt color: %w", err)
		}
		return fmt.Sprintf("cost.HybridMana(%s, %s)", first, second), nil
	case cost.TwobridSymbol:
		ctx.need(importMana)
		c, err := renderManaColor(symbol.Color)
		if err != nil {
			return "", fmt.Errorf("render: unsupported twobrid mana color: %w", err)
		}
		return fmt.Sprintf("cost.Twobrid(%s)", c), nil
	case cost.PhyrexianSymbol:
		ctx.need(importMana)
		c, err := renderManaColor(symbol.Color)
		if err != nil {
			return "", fmt.Errorf("render: unsupported phyrexian mana color: %w", err)
		}
		return fmt.Sprintf("cost.PhyrexianMana(%s)", c), nil
	default:
		return "", fmt.Errorf("render: unsupported mana symbol kind %d", symbol.Kind)
	}
}

func renderManaColor(c mana.Color) (string, error) {
	switch c {
	case mana.W:
		return "mana.W", nil
	case mana.U:
		return "mana.U", nil
	case mana.B:
		return "mana.B", nil
	case mana.R:
		return "mana.R", nil
	case mana.G:
		return "mana.G", nil
	case mana.C:
		return "mana.C", nil
	default:
		return "", fmt.Errorf("render: unsupported mana color %q", string(c))
	}
}

func renderManaColorSlice(ctx *renderCtx, colors []mana.Color) (string, error) {
	ctx.need(importMana)
	literals := make([]string, 0, len(colors))
	for _, c := range colors {
		literal, err := renderManaColor(c)
		if err != nil {
			return "", err
		}
		literals = append(literals, literal)
	}
	return "[]mana.Color{" + strings.Join(literals, ", ") + "}", nil
}

func renderTypesCardSlice(ctx *renderCtx, cardTypes []types.Card) (string, error) {
	ctx.need(importTypes)
	literals := make([]string, 0, len(cardTypes))
	for _, cardType := range cardTypes {
		lit, err := cardTypeLiteral(cardType)
		if err != nil {
			return "", err
		}
		literals = append(literals, lit)
	}
	return "[]types.Card{" + strings.Join(literals, ", ") + "}", nil
}

// cardTypeLiteral returns the Go constant for a types.Card value. It errors for
// any card type not known to the renderer's supported subset, preventing silent
// emission of comment fallbacks.
func cardTypeLiteral(t types.Card) (string, error) {
	lit := CardTypeToLiteral(string(t))
	if strings.HasPrefix(lit, "/*") {
		return "", fmt.Errorf("render: unsupported card type %q", string(t))
	}
	return lit, nil
}

// supertypeLiteral returns the Go constant for a types.Super value. It errors
// for any supertype not known to the renderer's supported subset.
func supertypeLiteral(st types.Super) (string, error) {
	lit := SupertypeToLiteral(string(st))
	if strings.HasPrefix(lit, "/*") {
		return "", fmt.Errorf("render: unsupported supertype %q", string(st))
	}
	return lit, nil
}

func renderAdditionalKind(kind cost.AdditionalKind) (string, error) {
	switch kind {
	case cost.AdditionalSacrifice:
		return "cost.AdditionalSacrifice", nil
	case cost.AdditionalSacrificeSource:
		return "cost.AdditionalSacrificeSource", nil
	case cost.AdditionalDiscard:
		return "cost.AdditionalDiscard", nil
	case cost.AdditionalPayLife:
		return "cost.AdditionalPayLife", nil
	case cost.AdditionalExile:
		return "cost.AdditionalExile", nil
	case cost.AdditionalReveal:
		return "cost.AdditionalReveal", nil
	case cost.AdditionalTap:
		return "cost.AdditionalTap", nil
	case cost.AdditionalExileSource:
		return "cost.AdditionalExileSource", nil
	case cost.AdditionalUntap:
		return "cost.AdditionalUntap", nil
	case cost.AdditionalRemoveCounter:
		return "cost.AdditionalRemoveCounter", nil
	case cost.AdditionalReturnUnblockedAttacker:
		return "cost.AdditionalReturnUnblockedAttacker", nil
	case cost.AdditionalTapPermanents:
		return "cost.AdditionalTapPermanents", nil
	case cost.AdditionalEnergy:
		return "cost.AdditionalEnergy", nil
	case cost.AdditionalReturnToHand:
		return "cost.AdditionalReturnToHand", nil
	case cost.AdditionalExert:
		return "cost.AdditionalExert", nil
	case cost.AdditionalMill:
		return "cost.AdditionalMill", nil
	case cost.AdditionalPutCounter:
		return "cost.AdditionalPutCounter", nil
	default:
		return "", fmt.Errorf("render: unsupported additional cost kind %d", kind)
	}
}

func renderCounterKind(kind counter.Kind) (string, error) {
	switch kind {
	case counter.PlusOnePlusOne:
		return "counter.PlusOnePlusOne", nil
	case counter.MinusOneMinusOne:
		return "counter.MinusOneMinusOne", nil
	case counter.Charge:
		return "counter.Charge", nil
	case counter.Loyalty:
		return "counter.Loyalty", nil
	case counter.Time:
		return "counter.Time", nil
	case counter.Defense:
		return "counter.Defense", nil
	case counter.Poison:
		return "counter.Poison", nil
	case counter.Lore:
		return "counter.Lore", nil
	case counter.Verse:
		return "counter.Verse", nil
	case counter.Shield:
		return "counter.Shield", nil
	case counter.Stun:
		return "counter.Stun", nil
	case counter.Finality:
		return "counter.Finality", nil
	case counter.Brick:
		return "counter.Brick", nil
	case counter.Page:
		return "counter.Page", nil
	case counter.Enlightened:
		return "counter.Enlightened", nil
	case counter.Oil:
		return "counter.Oil", nil
	case counter.Blood:
		return "counter.Blood", nil
	case counter.Indestructible:
		return "counter.Indestructible", nil
	case counter.Deathtouch:
		return "counter.Deathtouch", nil
	case counter.Flying:
		return "counter.Flying", nil
	case counter.FirstStrike:
		return "counter.FirstStrike", nil
	case counter.Hexproof:
		return "counter.Hexproof", nil
	case counter.Lifelink:
		return "counter.Lifelink", nil
	case counter.Menace:
		return "counter.Menace", nil
	case counter.Reach:
		return "counter.Reach", nil
	case counter.Trample:
		return "counter.Trample", nil
	case counter.Vigilance:
		return "counter.Vigilance", nil
	case counter.Energy:
		return "counter.Energy", nil
	case counter.Experience:
		return "counter.Experience", nil
	default:
		return "", fmt.Errorf("render: unsupported counter kind %d", kind)
	}
}

func renderTargetAllow(allow game.TargetAllow) string {
	var parts []string
	if allow&game.TargetAllowPermanent != 0 {
		parts = append(parts, "game.TargetAllowPermanent")
	}
	if allow&game.TargetAllowPlayer != 0 {
		parts = append(parts, "game.TargetAllowPlayer")
	}
	if allow&game.TargetAllowStackObject != 0 {
		parts = append(parts, "game.TargetAllowStackObject")
	}
	if allow&game.TargetAllowCard != 0 {
		parts = append(parts, "game.TargetAllowCard")
	}
	if len(parts) == 0 {
		return "game.TargetAllowUnspecified"
	}
	return strings.Join(parts, " | ")
}

func renderPlayerRelation(relation game.PlayerRelation) (string, error) {
	switch relation {
	case game.PlayerAny:
		return "game.PlayerAny", nil
	case game.PlayerYou:
		return "game.PlayerYou", nil
	case game.PlayerOpponent:
		return "game.PlayerOpponent", nil
	case game.PlayerNotYou:
		return "game.PlayerNotYou", nil
	default:
		return "", fmt.Errorf("render: unsupported player relation %d", relation)
	}
}

func renderTriggerType(triggerType game.TriggerType) (string, error) {
	switch triggerType {
	case game.TriggerWhen:
		return "game.TriggerWhen", nil
	case game.TriggerWhenever:
		return "game.TriggerWhenever", nil
	case game.TriggerAt:
		return "game.TriggerAt", nil
	case game.TriggerState:
		return "game.TriggerState", nil
	default:
		return "", fmt.Errorf("render: unsupported trigger type %d", triggerType)
	}
}

func renderStep(step game.Step) (string, error) {
	switch step {
	case game.StepUpkeep:
		return "game.StepUpkeep", nil
	case game.StepDraw:
		return "game.StepDraw", nil
	case game.StepBeginningOfCombat:
		return "game.StepBeginningOfCombat", nil
	case game.StepEnd:
		return "game.StepEnd", nil
	default:
		return "", fmt.Errorf("render: unsupported step %d", step)
	}
}

func renderTriggerSource(source game.TriggerSourceFilter) (string, error) {
	switch source {
	case game.TriggerSourceSelf:
		return "game.TriggerSourceSelf", nil
	case game.TriggerSourceAttachedPermanent:
		return "game.TriggerSourceAttachedPermanent", nil
	default:
		return "", fmt.Errorf("render: unsupported trigger source %d", source)
	}
}

func renderTriggerSubject(subject game.TriggerSubjectObject) (string, error) {
	switch subject {
	case game.TriggerSubjectPermanent:
		return "game.TriggerSubjectPermanent", nil
	case game.TriggerSubjectBlockedAttacker:
		return "game.TriggerSubjectBlockedAttacker", nil
	case game.TriggerSubjectDamageSource:
		return "game.TriggerSubjectDamageSource", nil
	default:
		return "", fmt.Errorf("render: unsupported trigger subject %d", subject)
	}
}

func renderTriggerController(controller game.TriggerControllerFilter) (string, error) {
	switch controller {
	case game.TriggerControllerYou:
		return "game.TriggerControllerYou", nil
	case game.TriggerControllerOpponent:
		return "game.TriggerControllerOpponent", nil
	default:
		return "", fmt.Errorf("render: unsupported trigger controller filter %d", controller)
	}
}

func renderTriggerPlayer(player game.TriggerPlayerFilter) (string, error) {
	switch player {
	case game.TriggerPlayerYou:
		return "game.TriggerPlayerYou", nil
	case game.TriggerPlayerOpponent:
		return "game.TriggerPlayerOpponent", nil
	default:
		return "", fmt.Errorf("render: unsupported trigger player filter %d", player)
	}
}

func renderEventKind(event game.EventKind) (string, error) {
	switch event {
	case game.EventDamageDealt:
		return "game.EventDamageDealt", nil
	case game.EventAttackerBecameBlocked:
		return "game.EventAttackerBecameBlocked", nil
	case game.EventAttackerDeclared:
		return "game.EventAttackerDeclared", nil
	case game.EventBlockerDeclared:
		return "game.EventBlockerDeclared", nil
	case game.EventSpellCast:
		return "game.EventSpellCast", nil
	case game.EventLifeGained:
		return "game.EventLifeGained", nil
	case game.EventLifeLost:
		return "game.EventLifeLost", nil
	case game.EventPermanentEnteredBattlefield:
		return "game.EventPermanentEnteredBattlefield", nil
	case game.EventPermanentDied:
		return "game.EventPermanentDied", nil
	case game.EventCardDiscarded:
		return "game.EventCardDiscarded", nil
	case game.EventCycled:
		return "game.EventCycled", nil
	case game.EventPermanentMutated:
		return "game.EventPermanentMutated", nil
	case game.EventBeginningOfStep:
		return "game.EventBeginningOfStep", nil
	default:
		return "", fmt.Errorf("render: unsupported event kind %d", event)
	}
}

func renderDamageRecipient(recipient game.DamageRecipientKind) (string, error) {
	switch recipient {
	case game.DamageRecipientPlayer:
		return "game.DamageRecipientPlayer", nil
	case game.DamageRecipientPermanent:
		return "game.DamageRecipientPermanent", nil
	default:
		return "", fmt.Errorf("render: unsupported damage recipient %d", recipient)
	}
}

func renderDuration(duration game.EffectDuration) (string, error) {
	switch duration {
	case game.DurationUntilEndOfTurn:
		return "game.DurationUntilEndOfTurn", nil
	case game.DurationUntilEndOfYourNextTurn:
		return "game.DurationUntilEndOfYourNextTurn", nil
	default:
		return "", fmt.Errorf("render: unsupported effect duration %d", duration)
	}
}

func renderResolutionChoiceKind(kind game.ResolutionChoiceKind) (string, error) {
	switch kind {
	case game.ResolutionChoiceMana:
		return "game.ResolutionChoiceMana", nil
	case game.ResolutionChoiceCardType:
		return "game.ResolutionChoiceCardType", nil
	case game.ResolutionChoicePlayer:
		return "game.ResolutionChoicePlayer", nil
	case game.ResolutionChoiceCard:
		return "game.ResolutionChoiceCard", nil
	default:
		return "", fmt.Errorf("render: unsupported resolution choice kind %d", kind)
	}
}

func renderZone(zoneType zone.Type) (string, error) {
	switch zoneType {
	case zone.Battlefield:
		return "zone.Battlefield", nil
	case zone.Hand:
		return "zone.Hand", nil
	case zone.Graveyard:
		return "zone.Graveyard", nil
	case zone.Library:
		return "zone.Library", nil
	case zone.Exile:
		return "zone.Exile", nil
	default:
		return "", fmt.Errorf("render: unsupported zone %d", zoneType)
	}
}

// renderText renders a string field value, preferring a raw backtick literal for
// multi-line text and falling back to a quoted literal when the text already
// contains a backtick.
func renderText(text string) string {
	if strings.ContainsRune(text, '`') {
		return strconv.Quote(text)
	}
	if strings.ContainsRune(text, '\n') {
		return "`" + text + "`"
	}
	return strconv.Quote(text)
}

func structLit(typeName string, fields []string) string {
	if len(fields) == 0 {
		return typeName + "{}"
	}
	return typeName + "{\n" + strings.Join(fields, "\n") + "\n}"
}

func sliceLit(elementType string, elements []string) string {
	if len(elements) == 0 {
		return "[]" + elementType + "{}"
	}
	return "[]" + elementType + "{\n" + strings.Join(elements, "\n") + "\n}"
}

func sliceField(fieldName, elementType string, elements []string) string {
	return fieldName + ": " + sliceLit(elementType, elements) + ","
}

// compactStructLit renders a struct literal on a single line so that gofmt
// preserves it inline. Each field must be a "Key: value" fragment without a
// trailing comma.
func compactStructLit(typeName string, fields []string) string {
	return typeName + "{" + strings.Join(fields, ", ") + "}"
}

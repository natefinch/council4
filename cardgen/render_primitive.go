package cardgen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/natefinch/council4/mtg/game"
)

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
	case game.CardReferenceCaptured:
		if reference.LinkID != "" {
			return "", errors.New("render: captured card reference has LinkID")
		}
		if reference.TargetIndex != 0 {
			return "", errors.New("render: captured card reference has TargetIndex")
		}
		return "game.CardReference{Kind: game.CardReferenceCaptured}", nil
	default:
		return "", fmt.Errorf("render: unsupported card reference kind %d", reference.Kind)
	}
}

// renderTokenSource renders a CreateToken's TokenSource: either a synthesized
// token CardDef var or a copy of the effect's target object. Copy specs may
// carry characteristic-overriding exceptions ("except it's a 1/1 green Frog",
// "except it's an artifact in addition to its other types"); their power,
// toughness, color, card-type, subtype, and keyword overrides are rendered
// below. Eternalize-style specs that rename, drop mana cost, or drop printed
// text are built directly in card code, never rendered, so they fail closed.
func (r Renderer) renderTokenSource(ctx *renderCtx, source game.TokenSource) (string, error) {
	if def, ok := source.TokenDefRef(); ok {
		return fmt.Sprintf("game.TokenDef(%s)", ctx.tokenDefVar(def)), nil
	}
	spec, ok := source.TokenCopy()
	if !ok || spec.SetName != "" || spec.NoManaCost || spec.NoPrintedText {
		return "", errors.New("render: unsupported CreateToken token source")
	}
	switch spec.Source {
	case game.TokenCopySourceObject:
		return r.renderTokenCopyObjectSource(ctx, spec)
	case game.TokenCopySourceEachInGroup:
		return r.renderTokenCopyForEachSource(ctx, spec)
	case game.TokenCopySourceChosenFromTriggerBatch:
		return renderTokenCopyTriggeringSetSource(ctx, spec)
	case game.TokenCopySourceChosenControlledCreatureToken:
		return renderTokenCopyPopulateSource(ctx, spec)
	case game.TokenCopySourceChosenFromGroup:
		return r.renderTokenCopyChosenFromGroupSource(ctx, spec)
	case game.TokenCopySourceLinkedExiledCard:
		return renderTokenCopyLinkedExiledCardSource(ctx, spec)
	default:
		return "", errors.New("render: unsupported CreateToken token source")
	}
}

func renderTokenCopyLinkedExiledCardSource(ctx *renderCtx, spec game.TokenCopySpec) (string, error) {
	fields := []string{
		"Source: game.TokenCopySourceLinkedExiledCard,",
		fmt.Sprintf("LinkID: game.LinkedKey(%q),", string(spec.LinkID)),
	}
	fields, err := appendTokenCopyModifierFields(ctx, fields, spec)
	if err != nil {
		return "", err
	}
	rendered, err := renderTokenCopyKeywordField(fields, spec)
	if err != nil {
		return "", err
	}
	return "game.TokenCopyOf(" + structLit("game.TokenCopySpec", rendered) + ")", nil
}

func (r Renderer) renderTokenCopyChosenFromGroupSource(ctx *renderCtx, spec game.TokenCopySpec) (string, error) {
	if spec.Group == nil {
		return "", errors.New("render: chosen-group token copy has no group")
	}
	group, err := r.renderGroupReference(ctx, *spec.Group)
	if err != nil {
		return "", err
	}
	fields := []string{
		"Source: game.TokenCopySourceChosenFromGroup,",
		fmt.Sprintf("Group: game.GroupRef(%s),", group),
	}
	fields, err = appendTokenCopyModifierFields(ctx, fields, spec)
	if err != nil {
		return "", err
	}
	rendered, err := renderTokenCopyKeywordField(fields, spec)
	if err != nil {
		return "", err
	}
	return structLit("game.TokenCopyOf(game.TokenCopySpec", rendered) + ")", nil
}

func renderTokenCopyPopulateSource(ctx *renderCtx, spec game.TokenCopySpec) (string, error) {
	fields := []string{
		"Source: game.TokenCopySourceChosenControlledCreatureToken,",
	}
	fields, err := appendTokenCopyModifierFields(ctx, fields, spec)
	if err != nil {
		return "", err
	}
	rendered, err := renderTokenCopyKeywordField(fields, spec)
	if err != nil {
		return "", err
	}
	return "game.TokenCopyOf(" + structLit("game.TokenCopySpec", rendered) + ")", nil
}

func renderTokenCopyTriggeringSetSource(ctx *renderCtx, spec game.TokenCopySpec) (string, error) {
	fields := []string{
		"Source: game.TokenCopySourceChosenFromTriggerBatch,",
	}
	fields, err := appendTokenCopyModifierFields(ctx, fields, spec)
	if err != nil {
		return "", err
	}
	rendered, err := renderTokenCopyKeywordField(fields, spec)
	if err != nil {
		return "", err
	}
	return structLit("game.TokenCopyOf(game.TokenCopySpec", rendered) + ")", nil
}

func (r Renderer) renderTokenCopyObjectSource(ctx *renderCtx, spec game.TokenCopySpec) (string, error) {
	object, err := r.renderObjectReference(spec.Object)
	if err != nil {
		return "", err
	}
	fields := []string{
		"Source: game.TokenCopySourceObject,",
		fmt.Sprintf("Object: %s,", object),
	}
	fields, err = appendTokenCopyModifierFields(ctx, fields, spec)
	if err != nil {
		return "", err
	}
	rendered, err := renderTokenCopyKeywordField(fields, spec)
	if err != nil {
		return "", err
	}
	return structLit("game.TokenCopyOf(game.TokenCopySpec", rendered) + ")", nil
}

func (r Renderer) renderTokenCopyForEachSource(ctx *renderCtx, spec game.TokenCopySpec) (string, error) {
	group, err := r.renderGroupReference(ctx, *spec.Group)
	if err != nil {
		return "", err
	}
	fields := []string{
		"Source: game.TokenCopySourceEachInGroup,",
		fmt.Sprintf("Group: game.GroupRef(%s),", group),
	}
	fields, err = appendTokenCopyModifierFields(ctx, fields, spec)
	if err != nil {
		return "", err
	}
	rendered, err := renderTokenCopyKeywordField(fields, spec)
	if err != nil {
		return "", err
	}
	return structLit("game.TokenCopyOf(game.TokenCopySpec", rendered) + ")", nil
}

// appendTokenCopyModifierFields renders the copy-token spec's
// characteristic-overriding exception fields: the legendary drop, the
// power/toughness override, the replacing color/type/subtype overrides ("except
// it's a 1/1 green Frog"), and the additive color/type/subtype overrides
// ("except it's an artifact in addition to its other types").
func appendTokenCopyModifierFields(ctx *renderCtx, fields []string, spec game.TokenCopySpec) ([]string, error) {
	if spec.SetNotLegendary {
		fields = append(fields, "SetNotLegendary: true,")
	}
	if spec.SetPower.Exists {
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("SetPower: opt.Val(%s),", renderPTValue(spec.SetPower.Val)))
	}
	if spec.SetToughness.Exists {
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("SetToughness: opt.Val(%s),", renderPTValue(spec.SetToughness.Val)))
	}
	if spec.HalvePowerToughnessRoundUp {
		fields = append(fields, "HalvePowerToughnessRoundUp: true,")
	}
	if len(spec.SetColors) != 0 {
		literal, err := renderColorSlice(ctx, spec.SetColors)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("SetColors: %s,", literal))
	}
	if len(spec.SetTypes) != 0 {
		literal, err := renderTypesCardSlice(ctx, spec.SetTypes)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("SetTypes: %s,", literal))
	}
	if len(spec.SetSubtypes) != 0 {
		literal, err := renderSubtypeSlice(ctx, spec.SetSubtypes)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("SetSubtypes: %s,", literal))
	}
	if len(spec.AddColors) != 0 {
		literal, err := renderColorSlice(ctx, spec.AddColors)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("AddColors: %s,", literal))
	}
	if len(spec.AddTypes) != 0 {
		literal, err := renderTypesCardSlice(ctx, spec.AddTypes)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("AddTypes: %s,", literal))
	}
	if len(spec.AddSubtypes) != 0 {
		literal, err := renderSubtypeSlice(ctx, spec.AddSubtypes)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("AddSubtypes: %s,", literal))
	}
	return fields, nil
}

func renderTokenCopyKeywordField(fields []string, spec game.TokenCopySpec) ([]string, error) {
	if len(spec.AddKeywords) == 0 {
		return fields, nil
	}
	rendered := make([]string, 0, len(spec.AddKeywords))
	for _, keyword := range spec.AddKeywords {
		literal, err := renderKeyword(keyword)
		if err != nil {
			return nil, err
		}
		rendered = append(rendered, literal)
	}
	return append(fields, fmt.Sprintf("AddKeywords: []game.Keyword{%s},", strings.Join(rendered, ", "))), nil
}

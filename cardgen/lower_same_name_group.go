package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// This file holds the composable "same name as that <noun>" group building
// blocks shared by every consumer of the same-name group family: the trailing
// "and all other <group> with the same name as that <noun>" clause the parser
// records on a single-permanent target's selector (Maelstrom Pulse and the
// Echoing cycle destroy verbs, Bile Blight and the Echoing pumps). Selection
// describes WHAT the group matches; game.SameNamePermanentGroup binds it to the
// chosen target so a single mass instruction affects the target together with
// every other battlefield permanent sharing its name (the anchor is included
// because it shares its own name). Each consuming verb (destroy, power/toughness
// modify, …) reuses these helpers so the group model is derived once.

// sameNameGroupTargetSpec builds the shared TargetSpec and same-name group
// Selection for a "target <noun> and all other <group> with the same name as
// that <noun>" effect. The trailing group clause is stripped from the target so
// the cleaned target reconstructs the bare "target <noun>" spec, and the group
// Selection carries the printed group noun's card-type restriction (empty for
// the bare "permanents" noun). It fails closed when the target carries no
// same-name group or the cleaned target is not a representable permanent target.
func sameNameGroupTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, game.Selection, bool) {
	group := target.Selector.SameNameGroup
	if group == nil {
		return game.TargetSpec{}, game.Selection{}, false
	}
	cleaned := target
	cleaned.Selector.SameNameGroup = nil
	spec, ok := permanentTargetSpec(cleaned)
	if !ok {
		return game.TargetSpec{}, game.Selection{}, false
	}
	groupSelection := game.Selection{}
	if len(group.GroupTypes) > 0 {
		groupSelection.RequiredTypes = append([]types.Card(nil), group.GroupTypes...)
	}
	return spec, groupSelection, true
}

// sameNameGroupBackReferencesSupported reports whether every reference in a
// same-name group body is the tolerated "that <noun>" back-reference that binds
// to the chosen target. The mass instruction reads the target through the group
// anchor, so the back-reference carries no additional runtime meaning; any other
// reference shape is rejected so unsupported wordings stay fail-closed.
func sameNameGroupBackReferencesSupported(references []compiler.CompiledReference) bool {
	for i := range references {
		if references[i].Kind != compiler.ReferenceThatObject ||
			references[i].Binding != compiler.ReferenceBindingTarget {
			return false
		}
	}
	return true
}

// lowerSameNameGroupModifyPTSpell lowers an exact fixed until-end-of-turn
// power/toughness change applied to a target creature and every other creature
// sharing its name ("Target creature and all other creatures with the same name
// as that creature get -3/-3 until end of turn.", Bile Blight; the Echoing
// pumps). The parser records the trailing "and all other <group> with the same
// name as that <noun>" clause as the target selector's same-name group, and the
// pump addresses the group's subject through the "that creature" back-reference,
// so the effect's context is the referenced object. The lowering emits one
// until-end-of-turn LayerPowerToughnessModify continuous effect over a
// SameNamePermanentGroup anchored on the chosen target, which includes the
// target itself, so the whole same-name group is modified in one move — the
// group analogue of the single-target ModifyPT pump. It returns false for any
// other shape (a variable/dynamic amount, a rider, a condition, or a non-creature
// pump target) so the caller falls through to the generic pump paths and the
// fail-closed diagnostic.
func lowerSameNameGroupModifyPTSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 1 {
		return game.AbilityContent{}, false
	}
	if ctx.content.Targets[0].Selector.SameNameGroup == nil {
		return game.AbilityContent{}, false
	}
	effect := &ctx.content.Effects[0]
	if len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		effect.Negated ||
		effect.Context != parser.EffectContextReferencedObject ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		effect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		!effect.PowerDelta.Known ||
		!effect.ToughnessDelta.Known ||
		!pumpTargetSelector(ctx.content.Targets[0].Selector) ||
		!sameNameGroupBackReferencesSupported(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	spec, groupSelection, ok := sameNameGroupTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	continuous := game.ContinuousEffect{
		Layer: game.LayerPowerToughnessModify,
		Group: game.SameNamePermanentGroup(
			game.TargetPermanentReference(0),
			groupSelection,
		),
		PowerDelta:     compiledSignedAmountValue(effect.PowerDelta),
		ToughnessDelta: compiledSignedAmountValue(effect.ToughnessDelta),
	}
	return game.Mode{
		Targets: []game.TargetSpec{spec},
		Sequence: []game.Instruction{
			{
				Primitive: game.ApplyContinuous{
					ContinuousEffects: []game.ContinuousEffect{continuous},
					Duration:          game.DurationUntilEndOfTurn,
				},
			},
		},
	}.Ability(), true
}

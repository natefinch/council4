package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// lowerCantBlockSpell lowers the temporary combat-restriction effect "<targets>
// can't block this turn." into one ApplyRule instruction per target slot, each
// placing an unconditional RuleEffectCantBlock restriction on a targeted
// creature for the turn (game.DurationThisTurn, removed during cleanup). It
// accepts the single-target form and the optional/plural multi-target
// cardinalities ("Up to three target creatures can't block this turn.") the
// parser recognizes; every other recipient, duration, condition, mode, or
// reference fails closed so the broader "can't block this turn" family stays
// faithful and bounded.
func lowerCantBlockSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported can't-block effect",
			"the executable source backend supports only exact \"<targets> can't block this turn.\"",
		)
	}
	effect := ctx.content.Effects[0]
	if effect.Context == parser.EffectContextController {
		return lowerGroupCantBlockSpell(ctx)
	}
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextTarget ||
		effect.Duration != compiler.DurationThisTurn ||
		ctx.optional ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Selector.Kind != compiler.SelectorCreature ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return unsupported()
	}
	targetSpec, ok := permanentTargetSpecWithCardinality(ctx.content.Targets[0])
	if !ok {
		return unsupported()
	}
	sequence := make([]game.Instruction, 0, targetSpec.MaxTargets)
	for i := range targetSpec.MaxTargets {
		sequence = append(sequence, game.Instruction{
			Primitive: game.ApplyRule{
				Object: opt.Val(game.TargetPermanentReference(i)),
				RuleEffects: []game.RuleEffect{
					{Kind: game.RuleEffectCantBlock},
				},
				Duration: game.DurationThisTurn,
			},
		})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{targetSpec},
		Sequence: sequence,
	}.Ability(), nil
}

// lowerGroupCantBlockSpell lowers the group-scoped combat-restriction effect
// "<group> can't block this turn." (Falter, Magmatic Chasm, Seismic Stomp:
// "Creatures without flying can't block this turn."; Cosmotronic Wave:
// "Creatures your opponents control can't block this turn.") into a single
// object-less ApplyRule that places an unconditional RuleEffectCantBlock
// restriction on every creature matching the affected group for the turn
// (game.DurationThisTurn, removed during cleanup). The affected group is scoped
// by controller relation, creature type, and an optional color/keyword filter
// carried on the rule effect. Any subject the group builder cannot faithfully
// represent, or any additional target, condition, mode, keyword, or reference,
// fails closed so the family stays bounded.
func lowerGroupCantBlockSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported can't-block effect",
			"the executable source backend supports only exact \"<group> can't block this turn.\" with a representable group",
		)
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		ctx.optional ||
		effect.Duration != compiler.DurationThisTurn ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return unsupported()
	}
	ruleEffect, ok := groupCantBlockRuleEffect(&effect)
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyRule{
				RuleEffects: []game.RuleEffect{ruleEffect},
				Duration:    game.DurationThisTurn,
			},
		}},
	}.Ability(), nil
}

// groupCantBlockRuleEffect builds the this-turn can't-block rule effect for a
// recognized group subject, scoping it by controller relation and creature type
// and folding any color/keyword filter onto the affected Selection. Only the
// controller-scoped whole-creature groups are representable; every other subject
// (subtype-, counter-, or power-filtered groups, the source-excluding "other
// creatures" form) fails closed.
func groupCantBlockRuleEffect(effect *compiler.CompiledEffect) (game.RuleEffect, bool) {
	var controller game.ControllerRelation
	switch effect.StaticSubject {
	case compiler.StaticSubjectAllCreatures:
		controller = game.ControllerAny
	case compiler.StaticSubjectControlledCreatures:
		controller = game.ControllerYou
	case compiler.StaticSubjectOpponentControlledCreatures:
		controller = game.ControllerOpponent
	default:
		return game.RuleEffect{}, false
	}
	selection, ok := groupCantBlockSelection(effect)
	if !ok {
		return game.RuleEffect{}, false
	}
	return game.RuleEffect{
		Kind:               game.RuleEffectCantBlock,
		AffectedController: controller,
		PermanentTypes:     []types.Card{types.Creature},
		AffectedSelection:  selection,
	}, true
}

// groupCantBlockSelection maps the recognized group subject's optional color and
// keyword filters onto a runtime Selection. It leaves the controller and
// creature-type scoping to the rule effect's AffectedController/PermanentTypes,
// so the returned Selection carries only the refinement predicates and is empty
// for an unfiltered group. Counter-, chosen-color-, and unmappable color/keyword
// filters fail closed.
func groupCantBlockSelection(effect *compiler.CompiledEffect) (game.Selection, bool) {
	if effect.StaticSubjectChosenColorFromEntry() {
		return game.Selection{}, false
	}
	if _, _, present := effect.StaticSubjectCounter(); present {
		return game.Selection{}, false
	}
	selection := game.Selection{
		Colorless:    effect.StaticSubjectColorless(),
		Multicolored: effect.StaticSubjectMulticolored(),
	}
	for _, parserColor := range effect.StaticSubjectColorsAny() {
		runtimeColor, ok := animateSelfColor(parserColor)
		if !ok {
			return game.Selection{}, false
		}
		selection.ColorsAny = append(selection.ColorsAny, runtimeColor)
	}
	if keyword, excluded, present := effect.StaticSubjectKeyword(); present {
		runtimeKw, ok := runtimeKeyword(keyword)
		if !ok {
			return game.Selection{}, false
		}
		if excluded {
			selection.ExcludedKeyword = runtimeKw
		} else {
			selection.Keyword = runtimeKw
		}
	}
	return selection, true
}

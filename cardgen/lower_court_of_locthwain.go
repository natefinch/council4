package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// courtOfLocthwainExileLink is the source-keyed linked set under which Court of
// Locthwain remembers each card it exiles from an opponent's library, so its
// monarch-gated free cast can act on "cards exiled with this enchantment".
const courtOfLocthwainExileLink = game.LinkedKey("court-of-locthwain-exile")

// lowerCourtOfLocthwainUpkeep lowers Court of Locthwain's upkeep trigger: "At the
// beginning of your upkeep, exile the top card of target opponent's library. You
// may play that card for as long as it remains exiled, and mana of any type can
// be spent to cast it. If you're the monarch, until end of turn, you may cast a
// spell from among cards exiled with this enchantment without paying its mana
// cost."
//
// The compiled body is three ordered effects sharing one target opponent and one
// monarch condition: the exile of the top card of that opponent's library, an
// impulse-style play permission carrying the any-type-mana rider ("that card ...
// for as long as it remains exiled, and mana of any type can be spent to cast
// it"), and a monarch-gated free cast from the accumulated pool of cards this
// enchantment exiled. Lowering merges the exile and its play permission into one
// ImpulseExile (top of the target opponent's library, any mana, remembered under
// the source-keyed linked set) and emits a monarch-gated ApplyRule that installs
// the until-end-of-turn free cast from that same linked pool. It fails closed for
// any shape it does not fully model, so the generic optional/ordered routes never
// silently drop a rider or the monarch gate.
func lowerCourtOfLocthwainUpkeep(ctx contentCtx) (game.AbilityContent, bool) {
	content := ctx.content
	if ctx.optional ||
		len(content.Effects) != 3 ||
		len(content.Targets) != 1 ||
		len(content.Conditions) != 1 ||
		len(content.Keywords) != 0 ||
		len(content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	if content.Targets[0].Cardinality.Min != 1 ||
		content.Targets[0].Cardinality.Max != 1 {
		return game.AbilityContent{}, false
	}
	condition := content.Conditions[0]
	if condition.Predicate != compiler.ConditionPredicateControllerIsMonarch ||
		condition.Negated {
		return game.AbilityContent{}, false
	}
	if !locthwainExileTopOfTargetLibrary(content.Effects[0]) ||
		!locthwainPlayFromExilePermission(content.Effects[1]) ||
		!locthwainMonarchFreeCast(content.Effects[2]) {
		return game.AbilityContent{}, false
	}
	monarch := opt.Val(game.EffectCondition{
		Condition: opt.Val(game.Condition{ControllerIsMonarch: true}),
	})
	return game.Mode{
		Targets: []game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Constraint: "target opponent",
			Allow:      game.TargetAllowPlayer,
			Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
		}},
		Sequence: []game.Instruction{
			{Primitive: game.ImpulseExile{
				Player:        game.TargetPlayerReference(0),
				Amount:        game.Fixed(1),
				Duration:      game.DurationPermanent,
				SpendAnyMana:  true,
				PublishLinked: courtOfLocthwainExileLink,
			}},
			{
				Primitive: game.ApplyRule{
					RuleEffects: []game.RuleEffect{{
						Kind:           game.RuleEffectCastLinkedExileForFree,
						AffectedPlayer: game.PlayerYou,
						ExiledLinkKey:  courtOfLocthwainExileLink,
					}},
					Duration: game.DurationUntilEndOfTurn,
				},
				Condition: monarch,
			},
		},
	}.Ability(), true
}

// locthwainExileTopOfTargetLibrary reports whether the effect is Court of
// Locthwain's controller-context exile of the single target opponent's top
// library card. The library owner is the ability's lone player target, carried
// as the effect's sole target rather than a reference.
func locthwainExileTopOfTargetLibrary(effect compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectExile &&
		!effect.Negated &&
		!effect.Optional &&
		effect.Context == parser.EffectContextController &&
		effect.Selector.Kind == compiler.SelectorCard &&
		len(effect.Targets) == 1 &&
		effect.Targets[0].Cardinality.Min == 1 &&
		effect.Targets[0].Cardinality.Max == 1 &&
		len(effect.References) == 0
}

// locthwainPlayFromExilePermission reports whether the effect is the impulse-
// style permission to play the just-exiled card ("You may play that card for as
// long as it remains exiled, and mana of any type can be spent to cast it."). Its
// references are the target-bound "that card"/"it" and the prior-result "it"; it
// takes no target and carries no effect context of its own.
func locthwainPlayFromExilePermission(effect compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectCast &&
		!effect.Negated &&
		!effect.Optional &&
		effect.Context == parser.EffectContextUnknown &&
		len(effect.Targets) == 0 &&
		len(effect.References) == 3 &&
		effect.References[0].Kind == compiler.ReferenceThatObject &&
		effect.References[0].Binding == compiler.ReferenceBindingTarget &&
		effect.References[1].Kind == compiler.ReferencePronoun &&
		effect.References[1].Binding == compiler.ReferenceBindingTarget &&
		effect.References[2].Kind == compiler.ReferencePronoun &&
		effect.References[2].Binding == compiler.ReferenceBindingPriorInstructionResult
}

// locthwainMonarchFreeCast reports whether the effect is the monarch-gated,
// until-end-of-turn optional free cast of a spell from among the cards this
// enchantment exiled ("If you're the monarch, until end of turn, you may cast a
// spell from among cards exiled with this enchantment without paying its mana
// cost."). Its references are the source-bound "this enchantment" and "its".
func locthwainMonarchFreeCast(effect compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectCast &&
		effect.Optional &&
		effect.CastWithoutPayingManaCost &&
		!effect.Negated &&
		effect.Context == parser.EffectContextController &&
		effect.Duration == compiler.DurationUntilEndOfTurn &&
		effect.Selector.Kind == compiler.SelectorSpell &&
		len(effect.Targets) == 0 &&
		len(effect.References) == 2 &&
		effect.References[0].Kind == compiler.ReferenceThisObject &&
		effect.References[0].Binding == compiler.ReferenceBindingSource &&
		effect.References[1].Kind == compiler.ReferencePronoun &&
		effect.References[1].Binding == compiler.ReferenceBindingSource
}

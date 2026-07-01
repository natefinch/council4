package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerTargetPlayerHandExile lowers the "target <player> exiles N card(s) from
// their hand" family (Mindmelter, Skullcap Snail, Unscrupulous Agent, Vessel of
// Malignity, Kyoki, and the modal Aim for the Head / Perfect Intimidation
// options) into one ChooseFromZone instruction that has the targeted player
// exile a fixed number of cards from their own hand. The targeted player is the
// chooser, mirroring how the rules engine resolves a "that player exiles ..."
// instruction: the candidate pool is the chosen player's hand, so the exile is a
// hidden-information choice that player makes.
//
// It reuses the existing ExileFromHandChoice runtime envelope (Chrome Mox's
// imprint exiles the controller's own hand the same way) with no parser or
// runtime change, supplying the single target player as the choosing player.
//
// It returns ok=false for any shape it does not fully model so the unmodeled
// wording falls through to the generic exile path's diagnostic rather than
// lowering to a silently-wrong instruction: a body-level optional, a modal or
// multi-effect body, a condition, a keyword rider, a non-hand source, a
// non-fixed or zero amount, a target that is not a single player, or a reference
// other than the redundant "their" possessive naming the targeted player.
func lowerTargetPlayerHandExile(ctx contentCtx) (game.AbilityContent, bool) {
	// lowerFixedExileSpell reaches this only through lowerImmediateSingleEffectSpell's
	// EffectExile arm, which guarantees single-effect content whose sole effect is an
	// EffectExile; a different count or kind is a dispatch bug, not an unsupported card.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerTargetPlayerHandExile: reached with %d effects; the EffectExile dispatch is single-effect", len(ctx.content.Effects)))
	}
	if ctx.optional ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Targets) != 1 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectExile {
		panic(fmt.Sprintf("lowerTargetPlayerHandExile: reached with effect kind %v; the EffectExile dispatch guarantees EffectExile", effect.Kind))
	}
	if effect.Negated ||
		effect.Optional ||
		effect.Divided ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextTarget ||
		effect.FromZone != zone.Hand {
		return game.AbilityContent{}, false
	}
	selector := effect.Selector
	if selector.Kind != compiler.SelectorCard ||
		selector.Zone != zone.Hand ||
		selector.Controller != compiler.ControllerAny ||
		selector.All ||
		selector.Another ||
		selector.Other ||
		selector.Attacking ||
		selector.Blocking ||
		selector.Tapped ||
		selector.Untapped {
		return game.AbilityContent{}, false
	}
	if !effect.Amount.Known ||
		effect.Amount.RangeKnown ||
		effect.Amount.VariableX ||
		effect.Amount.DynamicKind != 0 ||
		effect.Amount.Value < 1 {
		return game.AbilityContent{}, false
	}
	if !targetPlayerExileReferencesOnly(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := playerTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	selection, ok := cardSelectionForSelector(selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.ExileFromHandChoice(
				game.TargetPlayerReference(0),
				selection,
				game.Fixed(effect.Amount.Value),
				"",
			),
		}},
	}.Ability(), true
}

// targetPlayerExileReferencesOnly reports whether every reference of a
// target-player hand exile is the redundant "their"/"them" possessive naming the
// player the effect already targets ("... exiles a card from their hand"). That
// pronoun is bound to the sole target and is already expressed by scoping the
// choice to the target player, so it is the only reference tolerated; every other
// reference fails closed. A reference-free body always passes.
func targetPlayerExileReferencesOnly(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Kind != compiler.ReferencePronoun ||
			reference.Binding != compiler.ReferenceBindingTarget {
			return false
		}
		if reference.Pronoun != compiler.ReferencePronounTheir &&
			reference.Pronoun != compiler.ReferencePronounThem {
			return false
		}
	}
	return true
}

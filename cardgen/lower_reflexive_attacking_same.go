package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// reflexiveAttackingSameRiderPresent reports whether the effect carries the
// reflexive attacking-opponent rider — the anaphoric "Each opponent attacking
// that player does the same." (Curse of Opulence, Curse of Disturbance, Curse of
// Vitality, Curse of Verbosity) or the explicit "Each opponent attacking that
// player untaps all nonland permanents they control." (Curse of Bounty). The
// parser folds either rider onto an enchanted-player combat trigger's lone
// controller effect and records the rider sentence span here. Lowering reads the
// span to emit the extra group instruction that mirrors the controller effect
// onto each opponent attacking the enchanted player.
func reflexiveAttackingSameRiderPresent(effect *compiler.CompiledEffect) bool {
	return effect.EachOpponentAttackingSameRiderSpan != (shared.Span{})
}

// reflexiveAttackingSameControllerFixed reports whether a gain-life or draw
// effect carrying the anaphoric "does the same." rider is the exact plain
// controller fixed-amount shape the rider can widen to the attacking-opponent
// group: a controller-recipient effect with a known fixed count of at least one,
// no dynamic or variable amount, no targets, references, conditions, keywords, or
// modes, and neither negated nor optional. The parser credits the rider only onto
// a controller create-token, gain-life, or draw effect, so lowering re-checks the
// gain-life and draw shape here and fails closed on any other, refusing to widen
// a shape whose group mirror the runtime cannot attribute per recipient.
func reflexiveAttackingSameControllerFixed(ctx contentCtx, effect *compiler.CompiledEffect) bool {
	return effect.Context == parser.EffectContextController &&
		effect.Amount.Known && effect.Amount.Value >= 1 &&
		effect.Amount.DynamicKind == compiler.DynamicAmountNone &&
		!effect.Amount.VariableX &&
		!effect.Negated && !effect.Optional && !ctx.optional &&
		len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 0 &&
		len(ctx.content.Conditions) == 0 &&
		len(ctx.content.Keywords) == 0 &&
		len(ctx.content.Modes) == 0
}

// reflexiveAttackingGroupReference is the player group the reflexive
// attacking-opponent rider mirrors the controller effect onto: the distinct
// opponents of the enchanted player's Aura controller who are attacking that
// enchanted player this combat. The resolver excludes the controller, so pairing
// this group instruction with the controller instruction never doubles an effect
// for the same player.
func reflexiveAttackingGroupReference() game.PlayerGroupReference {
	return game.OpponentsAttackingTriggerPlayerReference()
}

// lowerReflexiveAttackingUntapSpell lowers Curse of Bounty's enchanted-player
// combat trigger — "untap all nonland permanents you control. Each opponent
// attacking that player untaps all nonland permanents they control." The parser
// folds the explicit rider onto the controller untap and records its span, so
// this widens the single controller untap into two instructions: the
// controller's own untap of the nonland permanents they control, then a
// per-player untap that runs once for each opponent attacking the enchanted
// player, each untapping that opponent's own nonland permanents. The per-player
// instruction binds the attacking-opponent group through ForEachPlayerGroup and
// scopes its untap with a PlayerControlledGroup anchored on the bound group-offer
// member, so selection context resolves against each recipient — not the Aura
// controller — and every recipient untaps only the permanents they control.
//
// It reuses the base controller untap's group unchanged and derives the
// per-player group from the same mass selection with its "you control" filter
// dropped, leaving the member anchor to supply each recipient's control scope. It
// fails closed on any shape other than the exact "untap all nonland permanents
// you control" mass group the rider mirrors.
func lowerReflexiveAttackingUntapSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if _, ok := exactMassGroup(ctx); !ok ||
		effect.Context != parser.EffectContextController ||
		effect.Selector.Controller != compiler.ControllerYou {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported untap spell",
			"the reflexive attacking-opponent rider widens only an exact untap of all nonland permanents you control",
		)
	}
	controllerSelection, ok := massGroupSelection(effect.Selector)
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported untap spell",
			"the reflexive attacking-opponent untap group is not representable",
		)
	}
	memberSelection := controllerSelection
	memberSelection.Controller = game.ControllerAny
	controllerGroup := game.BattlefieldGroup(controllerSelection)
	memberGroup := game.PlayerControlledGroup(game.GroupOfferMemberReference(), memberSelection)
	if len(controllerGroup.Validate()) != 0 || len(memberGroup.Validate()) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported untap spell",
			"the reflexive attacking-opponent untap group is not representable",
		)
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.Untap{Group: controllerGroup}},
			{
				Primitive:          game.Untap{Group: memberGroup},
				ForEachPlayerGroup: opt.Val(reflexiveAttackingGroupReference()),
			},
		},
	}.Ability(), nil
}

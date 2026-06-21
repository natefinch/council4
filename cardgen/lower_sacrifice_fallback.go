package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerSacrificeWithInabilityFallback folds an edict followed by a "who can't"
// rider ("each player sacrifices a creature or planeswalker of their choice.
// Each player who can't discards a card.") into one SacrificePermanents
// instruction carrying a per-player Fallback. The rider applies only to players
// who couldn't satisfy the edict, which the runtime resolves by checking each
// player's eligible permanents. It returns false unless the sequence is exactly
// a group sacrifice edict plus a supported inability fallback (discard or lose
// life) targeting the same player group.
func lowerSacrificeWithInabilityFallback(ctx contentCtx) (game.AbilityContent, bool) {
	content := ctx.content
	if len(content.Effects) != 2 ||
		len(content.Targets) != 0 ||
		len(content.Conditions) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	edict := &content.Effects[0]
	rider := &content.Effects[1]
	if edict.Kind != compiler.EffectSacrifice || !rider.FallbackOnInability {
		return game.AbilityContent{}, false
	}
	fallback, ok := inabilityFallback(rider)
	if !ok || !sameEachPlayerContext(edict.Context, rider.Context) {
		return game.AbilityContent{}, false
	}
	edictCtx := ctx
	edictContent := content
	edictContent.Effects = content.Effects[:1]
	edictContent.References = edict.References
	edictContent.Targets = edict.Targets
	edictCtx.content = edictContent
	ability, diagnostic := lowerSacrificeSpell(edictCtx)
	if diagnostic != nil || len(ability.Modes) != 1 || len(ability.Modes[0].Sequence) != 1 {
		return game.AbilityContent{}, false
	}
	prim, ok := ability.Modes[0].Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok || prim.PlayerGroup.Kind == game.PlayerGroupReferenceNone {
		return game.AbilityContent{}, false
	}
	prim.Fallback = fallback
	return game.Mode{Sequence: []game.Instruction{{Primitive: prim}}}.Ability(), true
}

// inabilityFallback maps a "who can't" rider effect to its typed SacrificeFallback.
// Only a plain own-choice discard of N cards or a loss of N life is supported.
func inabilityFallback(rider *compiler.CompiledEffect) (game.SacrificeFallback, bool) {
	if !rider.Exact || !rider.Amount.Known || rider.Amount.Value < 1 ||
		len(rider.References) != 0 || len(rider.Targets) != 0 {
		return game.SacrificeFallback{}, false
	}
	amount := game.Fixed(rider.Amount.Value)
	switch rider.Kind {
	case compiler.EffectDiscard:
		if rider.DiscardEntireHand || rider.HandDiscard.AtRandom {
			return game.SacrificeFallback{}, false
		}
		return game.SacrificeFallback{Kind: game.SacrificeFallbackDiscard, Amount: amount}, true
	case compiler.EffectLose:
		if !rider.LifeObject {
			return game.SacrificeFallback{}, false
		}
		return game.SacrificeFallback{Kind: game.SacrificeFallbackLoseLife, Amount: amount}, true
	default:
		return game.SacrificeFallback{}, false
	}
}

// sameEachPlayerContext reports whether both effects address the same each-player
// group, so a who-can't rider can pair with its edict.
func sameEachPlayerContext(a, b parser.EffectContextKind) bool {
	switch a {
	case parser.EffectContextEachPlayer,
		parser.EffectContextEachOpponent,
		parser.EffectContextEachOtherPlayer:
		return a == b
	default:
		return false
	}
}

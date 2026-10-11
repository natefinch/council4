package parser

import "github.com/natefinch/council4/mtg/game/zone"

func emitTargetOwnerLibraryProducerClauses(effects []*EffectSyntax) {
	for i, effect := range effects {
		if i == 0 || effect.Kind != EffectReveal || !effect.Exact ||
			effect.Context != EffectContextPriorSubject || effect.Player != EffectPlayerTargetOwner ||
			effect.CardSource != EffectCardSourceTopOfPlayerLibrary ||
			!effect.Amount.Known || effect.Amount.Value != 1 {
			continue
		}
		owner := effects[i-1]
		if owner.Kind == EffectShuffle && owner.Exact && owner.Context == EffectContextTarget &&
			owner.Player == EffectPlayerTargetOwner && owner.ToZone == zone.Library &&
			!owner.Optional && !owner.Negated && len(owner.Targets) == 1 {
			effect.LibraryOwnerClauseID = owner.ClauseID
		}
	}
}

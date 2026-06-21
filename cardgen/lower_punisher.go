package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerPunisherLoseLifeSpell lowers the "punisher" family ("Each opponent loses
// N life unless that player sacrifices a permanent of their choice or discards
// a card.") to a single PunisherEachLoseLife instruction. The affected group is
// taken from the effect's player context, the life amount from its Amount, and
// the sacrifice filter from its Selector. It fails closed for any targets,
// conditions, keywords, or modes.
func lowerPunisherLoseLifeSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported punisher effect",
			"the executable source backend does not yet lower this punisher effect",
		)
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		effect.Negated ||
		(!effect.PunisherSacrifice && !effect.PunisherDiscard) {
		return unsupported()
	}
	var group game.PlayerGroupReference
	switch effect.Context {
	case parser.EffectContextEachOpponent, parser.EffectContextEachOtherPlayer:
		group = game.OpponentsReference()
	case parser.EffectContextEachPlayer:
		group = game.AllPlayersReference()
	default:
		return unsupported()
	}
	prim := game.PunisherEachLoseLife{
		PlayerGroup:    group,
		Amount:         game.Fixed(effect.Amount.Value),
		AllowSacrifice: effect.PunisherSacrifice,
		AllowDiscard:   effect.PunisherDiscard,
	}
	if effect.PunisherSacrifice {
		selection, ok := massGroupSelection(effect.Selector)
		if !ok || selection.Controller != game.ControllerAny {
			return unsupported()
		}
		prim.SacrificeSelection = selection
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: prim}}}.Ability(), nil
}

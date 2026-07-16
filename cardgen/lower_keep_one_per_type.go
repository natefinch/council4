package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// lowerKeepOnePerTypeSacrifice lowers a recognized "keep one of each type"
// sacrifice effect (Liliana, Dreadhorde General's −9; Cataclysm; Cataclysmic
// Gearhulk) into the generic game.KeepOnePerType primitive. It fails closed if
// the effect is not a standalone sacrifice — any surrounding target, condition,
// keyword, mode, or sibling effect the primitive cannot express keeps the card
// unsupported rather than lowering an approximation.
func lowerKeepOnePerTypeSacrifice(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported keep-one-per-type sacrifice",
			"the executable source backend does not yet lower this keep-one-per-type sacrifice effect",
		)
	}

	keep := ctx.content.Effects[0].KeepOnePerType
	if keep == nil ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(keep.Types) == 0 {
		return unsupported()
	}

	var players game.PlayerGroupReference
	switch keep.Scope {
	case parser.KeepScopeOpponents:
		players = game.OpponentsReference()
	case parser.KeepScopeAllPlayers:
		players = game.AllPlayersReference()
	default:
		return unsupported()
	}

	var affected game.Selection
	if keep.NonlandOnly {
		// The nonland variants ("... from among the nonland permanents they
		// control ...") leave each player's lands untouched; every other permanent
		// they control is in the pool.
		affected = game.Selection{ExcludedTypes: []types.Card{types.Land}}
	}

	primitive := game.KeepOnePerType{
		Players:                 players,
		Types:                   keep.Types,
		AffectedSelection:       affected,
		ControllerChoosesForAll: keep.ControllerChoosesForAll,
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: primitive}}}.Ability(), nil
}

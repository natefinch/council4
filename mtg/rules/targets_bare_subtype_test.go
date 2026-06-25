package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// bareSubtypeCreatureDef builds a vanilla creature carrying a single subtype,
// matching the board shape the cardgen lowerer's bare-subtype target selects.
func bareSubtypeCreatureDef(name string, subtype types.Sub) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{subtype},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

// TestBareSubtypeTargetSpecMatchesBySubtype exercises the runtime target
// legality of the predicate shape cardgen produces for a bare-subtype target
// such as "target Beast you control": empty PermanentTypes with a Subtypes
// filter. Without the permanentTypeMatchesSpec subtype short-circuit these
// specs match no permanent at all, so this fails before that fix and passes
// after it.
func TestBareSubtypeTargetSpecMatchesBySubtype(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ownBeast := addCombatPermanent(g, game.Player1, bareSubtypeCreatureDef("Own Beast", types.Beast))
	ownSoldier := addCombatPermanent(g, game.Player1, bareSubtypeCreatureDef("Own Soldier", types.Soldier))
	opponentBeast := addCombatPermanent(g, game.Player2, bareSubtypeCreatureDef("Opponent Beast", types.Beast))

	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
		Constraint: "target Beast you control",
		Selection: opt.Val(game.Selection{
			SubtypesAny: []types.Sub{types.Beast},
			Controller:  game.ControllerYou,
		}),
	}

	if !permanentTargetMatchesSpec(g, game.Player1, 0, &spec, ownBeast.ObjectID) {
		t.Fatal("controlled Beast is not a legal target, want legal")
	}
	if permanentTargetMatchesSpec(g, game.Player1, 0, &spec, ownSoldier.ObjectID) {
		t.Fatal("controlled non-Beast is a legal target, want illegal")
	}
	if permanentTargetMatchesSpec(g, game.Player1, 0, &spec, opponentBeast.ObjectID) {
		t.Fatal("opponent's Beast is a legal target for a you-control spec, want illegal")
	}
}

// TestBareSubtypeAnotherTargetExcludesSource confirms that the "another" form
// ("another target Soldier you control") excludes the source permanent itself
// while still matching a different controlled permanent of that subtype.
func TestBareSubtypeAnotherTargetExcludesSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, bareSubtypeCreatureDef("Source Soldier", types.Soldier))
	otherSoldier := addCombatPermanent(g, game.Player1, bareSubtypeCreatureDef("Other Soldier", types.Soldier))

	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
		Constraint: "another target Soldier you control",
		Selection: opt.Val(game.Selection{
			SubtypesAny:   []types.Sub{types.Soldier},
			Controller:    game.ControllerYou,
			ExcludeSource: true,
		}),
	}

	if permanentTargetMatchesSpec(g, game.Player1, source.ObjectID, &spec, source.ObjectID) {
		t.Fatal("source permanent is a legal target for an \"another\" spec, want excluded")
	}
	if !permanentTargetMatchesSpec(g, game.Player1, source.ObjectID, &spec, otherSoldier.ObjectID) {
		t.Fatal("a different controlled Soldier is not a legal target, want legal")
	}
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestMentorTargetRequiresLesserPowerThanSource covers the source-relative
// "with lesser power" target filter that Mentor's attack trigger uses: a
// TargetPredicate.PowerLessThanSource requires the candidate permanent's power
// to be strictly less than the ability source's power. The mentoring creature
// (power 3) may target a weaker attacker (power 2) but not one of equal (3) or
// greater (4) power.
func TestMentorTargetRequiresLesserPowerThanSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mentor := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Legion Warboss",
		Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 3}),
	}})
	weaker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Weaker",
		Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 2}),
	}})
	equal := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Equal",
		Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 3}),
	}})
	stronger := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Stronger",
		Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 4}),
	}})

	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
		Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, PowerLessThanSource: true}),
	}

	if !permanentTargetMatchesSpec(g, game.Player1, mentor.ObjectID, &spec, weaker.ObjectID) {
		t.Fatal("creature with lesser power than the source should be a legal target")
	}
	if permanentTargetMatchesSpec(g, game.Player1, mentor.ObjectID, &spec, equal.ObjectID) {
		t.Fatal("creature with equal power must not match a lesser-power filter")
	}
	if permanentTargetMatchesSpec(g, game.Player1, mentor.ObjectID, &spec, stronger.ObjectID) {
		t.Fatal("creature with greater power must not match a lesser-power filter")
	}

	greaterSpec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
		Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, PowerGreaterThanSource: true}),
	}

	if !permanentTargetMatchesSpec(g, game.Player1, mentor.ObjectID, &greaterSpec, stronger.ObjectID) {
		t.Fatal("creature with greater power than the source should be a legal target for a greater-power filter")
	}
	if permanentTargetMatchesSpec(g, game.Player1, mentor.ObjectID, &greaterSpec, weaker.ObjectID) {
		t.Fatal("creature with lesser power must not match a greater-power filter")
	}
}

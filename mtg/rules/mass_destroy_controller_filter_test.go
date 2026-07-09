package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestMassDestroyControllerScopedUnionSparesCasterPermanents is the runtime proof
// for In Garruk's Wake ("Destroy all creatures you don't control and all
// planeswalkers you don't control."). The controller-scoped union destroys every
// creature or planeswalker an opponent controls while leaving the caster's own
// creature and planeswalker untouched, and never touches noncreature/nonwalker
// permanents such as artifacts on either side.
func TestMassDestroyControllerScopedUnionSparesCasterPermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	planeswalkerDef := func() *game.CardDef {
		return &game.CardDef{CardFace: game.CardFace{Name: "Test Planeswalker",
			Types: []types.Card{types.Planeswalker}},
		}
	}
	artifactDef := func() *game.CardDef {
		return &game.CardDef{CardFace: game.CardFace{Name: "Test Artifact",
			Types: []types.Card{types.Artifact}},
		}
	}

	// Caster (Player1) permanents that must survive.
	ownCreature := addCreaturePermanent(g, game.Player1)
	ownPlaneswalker := addCombatPermanent(g, game.Player1, planeswalkerDef())
	ownArtifact := addCombatPermanent(g, game.Player1, artifactDef())

	// Opponent creatures and planeswalkers that must be destroyed.
	oppCreature2 := addCreaturePermanent(g, game.Player2)
	oppPlaneswalker2 := addCombatPermanent(g, game.Player2, planeswalkerDef())
	oppCreature3 := addCreaturePermanent(g, game.Player3)
	oppPlaneswalker4 := addCombatPermanent(g, game.Player4, planeswalkerDef())

	// An opponent artifact must survive: it is neither creature nor planeswalker.
	oppArtifact := addCombatPermanent(g, game.Player2, artifactDef())

	addEffectSpellToStack(g, game.Player1, game.Destroy{
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker},
			Controller:       game.ControllerNotYou,
		}),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	survivors := map[string]*game.Permanent{
		"caster creature":     ownCreature,
		"caster planeswalker": ownPlaneswalker,
		"caster artifact":     ownArtifact,
		"opponent artifact":   oppArtifact,
	}
	for label, permanent := range survivors {
		if _, ok := permanentByObjectID(g, permanent.ObjectID); !ok {
			t.Errorf("%s was destroyed but should have survived", label)
		}
	}

	destroyed := map[string]*game.Permanent{
		"opponent creature (Player2)":     oppCreature2,
		"opponent planeswalker (Player2)": oppPlaneswalker2,
		"opponent creature (Player3)":     oppCreature3,
		"opponent planeswalker (Player4)": oppPlaneswalker4,
	}
	for label, permanent := range destroyed {
		if _, ok := permanentByObjectID(g, permanent.ObjectID); ok {
			t.Errorf("%s survived but should have been destroyed", label)
		}
	}
}

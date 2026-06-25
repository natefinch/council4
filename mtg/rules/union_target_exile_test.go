package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// addTypedPermanent puts a permanent with the given card types and subtypes onto
// the battlefield under controller, returning it for target-matching assertions.
func addTypedPermanent(g *game.Game, controller game.PlayerID, cardTypes []types.Card, subtypes []types.Sub) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:     "Test Permanent",
			Types:    cardTypes,
			Subtypes: subtypes,
		}},
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

// TestUnionTargetExileMatchesEachUnionMember proves the card-type and subtype
// union TargetSpecs the cardgen backend emits for "Exile target artifact,
// creature, or enchantment." and "Exile target Skeleton, Vampire, or Zombie."
// admit every union member and reject permanents outside the union, so only the
// intended permanents are legal exile targets.
func TestUnionTargetExileMatchesEachUnionMember(t *testing.T) {
	t.Run("card-type union", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		spec := game.TargetSpec{
			MinTargets: 1,
			MaxTargets: 1,
			Allow:      game.TargetAllowPermanent,
			Constraint: "target artifact, creature, or enchantment",
			Selection: opt.Val(game.Selection{
				RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment},
			}),
		}
		artifact := addTypedPermanent(g, game.Player2, []types.Card{types.Artifact}, nil)
		creature := addTypedPermanent(g, game.Player2, []types.Card{types.Creature}, nil)
		enchantment := addTypedPermanent(g, game.Player2, []types.Card{types.Enchantment}, nil)
		land := addTypedPermanent(g, game.Player2, []types.Card{types.Land}, nil)

		for _, member := range []*game.Permanent{artifact, creature, enchantment} {
			if !permanentTargetMatchesSpec(g, game.Player1, member.ObjectID, &spec, member.ObjectID) {
				t.Errorf("union member %v was not a legal target", member.CardInstanceID)
			}
		}
		if permanentTargetMatchesSpec(g, game.Player1, land.ObjectID, &spec, land.ObjectID) {
			t.Error("land outside the union was a legal target")
		}
	})

	t.Run("subtype union", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		spec := game.TargetSpec{
			MinTargets: 1,
			MaxTargets: 1,
			Allow:      game.TargetAllowPermanent,
			Constraint: "target Skeleton, Vampire, or Zombie",
			Selection: opt.Val(game.Selection{
				SubtypesAny: []types.Sub{types.Skeleton, types.Vampire, types.Zombie},
			}),
		}
		skeleton := addTypedPermanent(g, game.Player2, []types.Card{types.Creature}, []types.Sub{types.Skeleton})
		vampire := addTypedPermanent(g, game.Player2, []types.Card{types.Creature}, []types.Sub{types.Vampire})
		zombie := addTypedPermanent(g, game.Player2, []types.Card{types.Creature}, []types.Sub{types.Zombie})
		goblin := addTypedPermanent(g, game.Player2, []types.Card{types.Creature}, []types.Sub{types.Goblin})

		for _, member := range []*game.Permanent{skeleton, vampire, zombie} {
			if !permanentTargetMatchesSpec(g, game.Player1, member.ObjectID, &spec, member.ObjectID) {
				t.Errorf("union member %v was not a legal target", member.CardInstanceID)
			}
		}
		if permanentTargetMatchesSpec(g, game.Player1, goblin.ObjectID, &spec, goblin.ObjectID) {
			t.Error("Goblin outside the subtype union was a legal target")
		}
	})
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// coatOfArmsCard models Coat of Arms: each creature on the battlefield gets
// +1/+1 for each other creature that shares a creature type with it.
func coatOfArmsCard() *game.CardDef {
	bonus := opt.Val(game.DynamicAmount{
		Kind:       game.DynamicAmountSharedCreatureTypeCountInGroup,
		Multiplier: 1,
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		}),
	})
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Coat of Arms",
		Types: []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerPowerToughnessModify,
				Group: game.BattlefieldGroup(game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				}),
				PowerDeltaDynamic:     bonus,
				ToughnessDeltaDynamic: bonus,
			}},
		}},
	}}
}

func goblinWithPT(name string, power, toughness int) *game.CardDef {
	def := creatureWithPT(name, power, toughness)
	def.Subtypes = []types.Sub{types.Goblin}
	return def
}

// TestSharedCreatureTypeAnthemBuffsPerSharedType verifies that the Coat of Arms
// bonus equals, per affected creature, the number of other creatures sharing at
// least one creature type with it. A Changeling shares every creature type, so
// it counts toward and benefits from every other creature.
func TestSharedCreatureTypeAnthemBuffsPerSharedType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, coatOfArmsCard())
	goblin1 := addCombatPermanent(g, game.Player1, goblinWithPT("Goblin One", 1, 1))
	goblin2 := addCombatPermanent(g, game.Player1, goblinWithPT("Goblin Two", 1, 1))
	goblinWizard := addCombatPermanent(g, game.Player2, goblinWithPT("Goblin Wizard", 1, 1))

	zombie := creatureWithPT("Zombie", 2, 2)
	zombie.Subtypes = []types.Sub{types.Zombie}
	zombiePermanent := addCombatPermanent(g, game.Player1, zombie)

	changeling := creatureWithPT("Mistform Ultimus", 1, 1)
	changeling.StaticAbilities = []game.StaticAbility{game.ChangelingStaticBody}
	changelingPermanent := addCombatPermanent(g, game.Player2, changeling)

	cases := []struct {
		name      string
		permanent *game.Permanent
		base      int
		shared    int
	}{
		// Two other Goblins (the other Goblin and the Goblin Wizard) share, plus the
		// Changeling, which has every creature type.
		{"goblin shares with goblins and changeling", goblin1, 1, 3},
		{"second goblin matches", goblin2, 1, 3},
		{"goblin wizard matches across controllers", goblinWizard, 1, 3},
		// The Zombie shares only with the Changeling.
		{"zombie shares only with changeling", zombiePermanent, 2, 1},
		// The Changeling shares with every other creature: three Goblins and the
		// Zombie.
		{"changeling shares with everything", changelingPermanent, 1, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := tc.base + tc.shared
			if got := effectivePower(g, tc.permanent); got != want {
				t.Fatalf("effective power = %d, want %d", got, want)
			}
			toughness, ok := effectiveToughness(g, tc.permanent)
			if !ok || toughness != want {
				t.Fatalf("effective toughness = %d (ok=%v), want %d", toughness, ok, want)
			}
		})
	}
}

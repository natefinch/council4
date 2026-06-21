package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func tapLandForManaTriggerSource(subtype types.Sub, color mana.Color) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Crypt Ghast",
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:                game.EventPermanentTapped,
					Controller:           game.TriggerControllerYou,
					RequireTappedForMana: true,
					SubjectSelection:     game.Selection{SubtypesAny: []types.Sub{subtype}},
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: color},
			}}}.Ability(),
		}},
	}}
}

func swampManaLand(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:          "Swamp",
		Types:         []types.Card{types.Land},
		Subtypes:      []types.Sub{types.Swamp},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.B)},
	}})
}

// TestTapLandForManaTriggerAddsAdditionalMana proves the "Whenever you tap a
// Swamp for mana, add an additional {B}" mana trigger (Crypt Ghast, Nirkana
// Revenant): tapping a Swamp for mana yields the land's {B} plus the trigger's
// additional {B}.
func TestTapLandForManaTriggerAddsAdditionalMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, tapLandForManaTriggerSource(types.Swamp, mana.B))
	swamp := swampManaLand(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(swamp.ObjectID, 0, nil, 0)) {
		t.Fatal("activating swamp mana ability = false, want true")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].ManaPool.Amount(mana.B); got != 2 {
		t.Fatalf("black mana = %d, want 2 (1 from Swamp + 1 from trigger)", got)
	}
}

// TestTapLandForManaTriggerIgnoresOtherLandTypes proves the subtype filter: a
// non-Swamp land tapped for mana does not fire the "tap a Swamp for mana"
// trigger, so no additional {B} is added.
func TestTapLandForManaTriggerIgnoresOtherLandTypes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, tapLandForManaTriggerSource(types.Swamp, mana.B))
	forest := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:          "Forest",
		Types:         []types.Card{types.Land},
		Subtypes:      []types.Sub{types.Forest},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.G)},
	}})

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(forest.ObjectID, 0, nil, 0)) {
		t.Fatal("activating forest mana ability = false, want true")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].ManaPool.Amount(mana.B); got != 0 {
		t.Fatalf("black mana = %d, want 0 (Forest tap must not fire the Swamp trigger)", got)
	}
}

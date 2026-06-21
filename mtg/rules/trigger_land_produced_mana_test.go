package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func triggerLandProducedManaSource() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Mirari's Wake",
		Types: []types.Card{types.Enchantment},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:                game.EventPermanentTapped,
					Controller:           game.TriggerControllerYou,
					RequireTappedForMana: true,
					SubjectSelection:     game.Selection{RequiredTypes: []types.Card{types.Land}},
				},
			},
			Content: game.TriggerLandProducedManaContent(),
		}},
	}}
}

// TestTriggerLandProducedManaDoublesProducedType proves the "Whenever you tap a
// land for mana, add one mana of any type that land produced." mana doubler
// (Mirari's Wake, Zendikar Resurgent): tapping a Forest for {G} yields the
// land's {G} plus one additional mana of the same type the land produced.
func TestTriggerLandProducedManaDoublesProducedType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, triggerLandProducedManaSource())
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

	if got := g.Players[game.Player1].ManaPool.Amount(mana.G); got != 2 {
		t.Fatalf("green mana = %d, want 2 (1 from Forest + 1 of the same type from the trigger)", got)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 2 {
		t.Fatalf("total mana = %d, want 2 (only the produced type is mirrored)", got)
	}
}

// TestTriggerLandProducedManaMirrorsColorlessType proves the doubler mirrors the
// exact type produced: a land tapped for colorless {C} yields the land's {C}
// plus one additional {C}, never colored mana.
func TestTriggerLandProducedManaMirrorsColorlessType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, triggerLandProducedManaSource())
	wastes := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:          "Wastes",
		Types:         []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.C)},
	}})

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(wastes.ObjectID, 0, nil, 0)) {
		t.Fatal("activating wastes mana ability = false, want true")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 2 {
		t.Fatalf("colorless mana = %d, want 2 (1 from Wastes + 1 of the same type from the trigger)", got)
	}
}

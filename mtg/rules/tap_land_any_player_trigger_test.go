package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// anyPlayerTapLandManaDoubler is the generic "Whenever a player taps a land for
// mana, that player adds an additional {G}" tapped-for-mana trigger (High Tide,
// Bubbling Muck family with fixed mana): the subject's controller is
// unrestricted and the produced mana goes to the player who tapped, resolved
// through EventPlayerReference.
func anyPlayerTapLandManaDoubler() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Mana Flare",
		Types: []types.Card{types.Enchantment},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:                game.EventPermanentTapped,
					Controller:           game.TriggerControllerAny,
					RequireTappedForMana: true,
					SubjectSelection:     game.Selection{RequiredTypes: []types.Card{types.Land}},
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.AddMana{
					Amount:    game.Fixed(1),
					ManaColor: mana.G,
					Player:    opt.Val(game.EventPlayerReference()),
				},
			}}}.Ability(),
		}},
	}}
}

// TestAnyPlayerTapLandTriggerAddsManaToThatPlayer proves the generic any-player
// tapped-for-mana doubler routes the additional mana to the player who tapped
// the land. An opponent (Player2) taps a Forest while Player1 controls the
// doubler enchantment; the extra {G} lands in Player2's pool, not Player1's.
func TestAnyPlayerTapLandTriggerAddsManaToThatPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player2
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, anyPlayerTapLandManaDoubler())
	forest := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:          "Forest",
		Types:         []types.Card{types.Land},
		Subtypes:      []types.Sub{types.Forest},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.G)},
	}})

	if !engine.applyAction(g, game.Player2, action.ActivateAbility(forest.ObjectID, 0, nil, 0)) {
		t.Fatal("activating forest mana ability = false, want true")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player2].ManaPool.Amount(mana.G); got != 2 {
		t.Fatalf("Player2 green mana = %d, want 2 (1 from Forest + 1 from trigger)", got)
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.G); got != 0 {
		t.Fatalf("Player1 green mana = %d, want 0 (mana goes to the tapping player)", got)
	}
}

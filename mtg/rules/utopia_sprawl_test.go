package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func chosenColorTapForManaSource(subtype types.Sub) *game.CardDef {
	recipient := opt.Val(game.ObjectControllerReference(game.EventPermanentReference()))
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Utopia Sprawl",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
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
				Primitive: game.AddMana{
					Amount:          game.Fixed(1),
					EntryChoiceFrom: game.EntryColorChoiceKey,
					Player:          recipient,
				},
			}}}.Ability(),
		}},
	}}
}

// TestUtopiaSprawlAddsChosenColorWhenLandTappedForMana proves the "As this Aura
// enters, choose a color." entry choice combined with "Whenever enchanted Forest
// is tapped for mana, its controller adds an additional one mana of the chosen
// color.": tapping the Forest for mana yields the land's {G} plus one mana of the
// entry-time chosen color (here red).
func TestUtopiaSprawlAddsChosenColorWhenLandTappedForMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	aura := addCombatPermanent(g, game.Player1, chosenColorTapForManaSource(types.Forest))
	aura.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryColorChoiceKey: {Kind: game.ResolutionChoiceMana, Color: mana.R},
	}
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

	pool := g.Players[game.Player1].ManaPool
	if got := pool.Amount(mana.G); got != 1 {
		t.Fatalf("green mana = %d, want 1 (from Forest)", got)
	}
	if got := pool.Amount(mana.R); got != 1 {
		t.Fatalf("red mana = %d, want 1 (chosen color added by trigger)", got)
	}
}

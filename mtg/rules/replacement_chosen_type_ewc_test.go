package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestChosenTypeGroupEntersWithCountersHonorsEntryChoice proves Metallic Mimic's
// "Each other creature you control of the chosen type enters with an additional
// +1/+1 counter on it." reads the creature type the source chose as it entered:
// a later creature of the chosen type gains the extra counter while a creature
// of a different type does not.
func TestChosenTypeGroupEntersWithCountersHonorsEntryChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	subtypes := types.SubtypesForType(types.Creature)
	if len(subtypes) < 2 {
		t.Fatal("need at least two creature subtypes to choose between")
	}
	const pick = 3
	chosen := subtypes[pick]
	var other types.Sub
	for _, s := range subtypes {
		if s != chosen {
			other = s
			break
		}
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{pick}}}}

	mimicDef := &game.CardDef{CardFace: game.CardFace{
		Name:  "Metallic Mimic",
		Types: []types.Card{types.Artifact, types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntryTypeChoiceReplacement("As this creature enters, choose a creature type."),
			game.EntersWithCountersGroupReplacement(
				"Each other creature you control of the chosen type enters with an additional +1/+1 counter on it.",
				&game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					SubtypeChoice: game.SubtypeChoiceSourceEntry,
					Controller:    game.ControllerYou,
					ExcludeSource: true,
				},
				game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1},
			),
		},
	}}
	mimicID := addCardToHand(g, game.Player1, mimicDef)
	mimicCard, ok := g.GetCardInstance(mimicID)
	if !ok {
		t.Fatal("mimic card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(mimicID)
	if _, ok := createCardPermanentWithChoices(engine, g, mimicCard, game.Player1, zone.Hand, agents, &TurnLog{}); !ok {
		t.Fatal("createCardPermanentWithChoices(mimic) = false, want true")
	}

	matching := enterCreatureWithSubtype(t, g, game.Player1, "Chosen Kin", chosen)
	if got := matching.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters on chosen-type creature = %d, want 1", got)
	}
	mismatch := enterCreatureWithSubtype(t, g, game.Player1, "Off Type", other)
	if got := mismatch.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("+1/+1 counters on off-type creature = %d, want 0", got)
	}
}

func enterCreatureWithSubtype(t *testing.T, g *game.Game, controller game.PlayerID, name string, sub types.Sub) *game.Permanent {
	t.Helper()
	def := &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{sub},
	}}
	cardID := addCardToHand(g, controller, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatalf("card instance for %q not found", name)
	}
	g.Players[controller].Hand.Remove(cardID)
	permanent, ok := createCardPermanent(g, card, controller, zone.Hand)
	if !ok {
		t.Fatalf("createCardPermanent(%q) = false, want true", name)
	}
	return permanent
}

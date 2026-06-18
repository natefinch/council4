package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestEntryColorChoiceRecordedOnEntry verifies that "As this permanent enters,
// choose a color." prompts the controller as the permanent enters and records
// the chosen color on the permanent under EntryColorChoiceKey (CR 614.12).
func TestEntryColorChoiceRecordedOnEntry(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// The single-option prompt for a color lists W,U,B,R,G; select index 2 (B).
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{2}}}}

	def := &game.CardDef{CardFace: game.CardFace{
		Name:                 "Sol Grail",
		Types:                []types.Card{types.Artifact},
		ReplacementAbilities: []game.ReplacementAbility{game.EntryColorChoiceReplacement("As this artifact enters, choose a color.")},
		ManaAbilities:        []game.ManaAbility{game.TapChosenColorManaAbility("{T}: Add one mana of the chosen color.")},
	}}
	cardID := addCardToHand(g, game.Player1, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)

	permanent, ok := createCardPermanentWithChoices(engine, g, card, game.Player1, zone.Hand, agents, &TurnLog{})
	if !ok {
		t.Fatal("createCardPermanentWithChoices() = false, want true")
	}
	if permanent.Tapped {
		t.Fatal("standalone choose-a-color permanent must not enter tapped")
	}
	result, ok := permanent.EntryChoices[game.EntryColorChoiceKey]
	if !ok {
		t.Fatalf("entry color choice not recorded: %+v", permanent.EntryChoices)
	}
	if result.Color != mana.B {
		t.Fatalf("recorded entry color = %v, want %v", result.Color, mana.B)
	}

	// The mana ability must produce one mana of the color chosen on entry.
	want := action.ActivateAbility(permanent.ObjectID, 0, nil, 0)
	if !engine.applyAction(g, game.Player1, want) {
		t.Fatal("applyAction(chosen-color mana ability) = false, want true")
	}
	if !permanent.Tapped {
		t.Fatal("mana ability did not tap the permanent")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.B); got != 1 {
		t.Fatalf("chosen-color mana = %d, want 1 black", got)
	}
	if total := g.Players[game.Player1].ManaPool.Total(); total != 1 {
		t.Fatalf("mana pool total = %d, want exactly 1", total)
	}
}

// TestEntersTappedColorChoiceTapsAndRecords verifies the combined "This land
// enters tapped. As it enters, choose a color." both taps the permanent and
// records the chosen color on entry.
func TestEntersTappedColorChoiceTapsAndRecords(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Select index 4 (G) from the color prompt.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{4}}}}

	def := &game.CardDef{CardFace: game.CardFace{
		Name:                 "Uncharted Haven",
		Types:                []types.Card{types.Land},
		ReplacementAbilities: []game.ReplacementAbility{game.EntersTappedColorChoiceReplacement("This land enters tapped. As it enters, choose a color.")},
		ManaAbilities:        []game.ManaAbility{game.TapChosenColorManaAbility("{T}: Add one mana of the chosen color.")},
	}}
	cardID := addCardToHand(g, game.Player1, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)

	permanent, ok := createCardPermanentWithChoices(engine, g, card, game.Player1, zone.Hand, agents, &TurnLog{})
	if !ok {
		t.Fatal("createCardPermanentWithChoices() = false, want true")
	}
	if !permanent.Tapped {
		t.Fatal("combined enters-tapped color-choice permanent did not enter tapped")
	}
	result, ok := permanent.EntryChoices[game.EntryColorChoiceKey]
	if !ok {
		t.Fatalf("entry color choice not recorded: %+v", permanent.EntryChoices)
	}
	if result.Color != mana.G {
		t.Fatalf("recorded entry color = %v, want %v", result.Color, mana.G)
	}
}

// TestEntersTappedColorChoiceExcludingExcludesForbiddenColorAndComposites covers
// the Gate/Thriving land cycle: "This land enters tapped. As it enters, choose a
// color other than white." excludes white from the prompt, and the composite
// "{T}: Add {W} or one mana of the chosen color." offers the fixed white or the
// entry-chosen color.
func TestEntersTappedColorChoiceExcludingExcludesForbiddenColorAndComposites(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// With white excluded, the prompt lists [U,B,R,G]; index 3 selects green.
	// Were white still offered the list would be [W,U,B,R,G] and index 3 would be
	// red, so asserting green proves the forbidden color was removed.
	entryAgents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{3}}}}

	def := &game.CardDef{CardFace: game.CardFace{
		Name:  "Thriving Heath",
		Types: []types.Card{types.Land},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedColorChoiceExcludingReplacement("This land enters tapped. As it enters, choose a color other than white.", mana.W),
		},
		ManaAbilities: []game.ManaAbility{
			game.TapFixedOrChosenColorManaAbility("{T}: Add {W} or one mana of the chosen color.", mana.W),
		},
	}}
	cardID := addCardToHand(g, game.Player1, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)

	permanent, ok := createCardPermanentWithChoices(engine, g, card, game.Player1, zone.Hand, entryAgents, &TurnLog{})
	if !ok {
		t.Fatal("createCardPermanentWithChoices() = false, want true")
	}
	if !permanent.Tapped {
		t.Fatal("Gate/Thriving land did not enter tapped")
	}
	result, ok := permanent.EntryChoices[game.EntryColorChoiceKey]
	if !ok {
		t.Fatalf("entry color choice not recorded: %+v", permanent.EntryChoices)
	}
	if result.Color != mana.G {
		t.Fatalf("recorded entry color = %v, want green (proves white excluded)", result.Color)
	}

	// The composite mana ability offers the fixed white (index 0) or the chosen
	// green (index 1). Untap first since the land entered tapped, then pick the
	// chosen color.
	permanent.Tapped = false
	manaAgents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	act := action.ActivateAbility(permanent.ObjectID, 0, nil, 0)
	if !engine.applyActionWithChoices(g, game.Player1, act, manaAgents, &TurnLog{}) {
		t.Fatal("applyActionWithChoices(composite mana ability, chosen color) = false, want true")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.G); got != 1 {
		t.Fatalf("chosen-color mana = %d, want 1 green", got)
	}
	if total := g.Players[game.Player1].ManaPool.Total(); total != 1 {
		t.Fatalf("mana pool total = %d, want exactly 1", total)
	}

	// Activating again and picking index 0 produces the fixed white alternative.
	permanent.Tapped = false
	g.Players[game.Player1].ManaPool.Empty()
	fixedAgents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}
	if !engine.applyActionWithChoices(g, game.Player1, act, fixedAgents, &TurnLog{}) {
		t.Fatal("applyActionWithChoices(composite mana ability, fixed color) = false, want true")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.W); got != 1 {
		t.Fatalf("fixed-color mana = %d, want 1 white", got)
	}
}

// TestEntryTypeChoiceRecordedOnEntry verifies that "As this permanent enters,
// choose a creature type." prompts the controller as the permanent enters and
// records the chosen creature type on the permanent under EntryTypeChoiceKey
// (CR 614.12). #554 groundwork: later abilities referencing "the chosen type"
// read this stored subtype.
func TestEntryTypeChoiceRecordedOnEntry(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// The creature-type prompt enumerates the creature subtypes in lexical order.
	subtypes := types.SubtypesForType(types.Creature)
	if len(subtypes) == 0 {
		t.Fatal("no creature subtypes available to choose from")
	}
	const pick = 3
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{pick}}}}

	def := &game.CardDef{CardFace: game.CardFace{
		Name:                 "Test Banner",
		Types:                []types.Card{types.Artifact},
		ReplacementAbilities: []game.ReplacementAbility{game.EntryTypeChoiceReplacement("As this artifact enters, choose a creature type.")},
	}}
	cardID := addCardToHand(g, game.Player1, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)

	permanent, ok := createCardPermanentWithChoices(engine, g, card, game.Player1, zone.Hand, agents, &TurnLog{})
	if !ok {
		t.Fatal("createCardPermanentWithChoices() = false, want true")
	}
	result, ok := permanent.EntryChoices[game.EntryTypeChoiceKey]
	if !ok {
		t.Fatalf("entry type choice not recorded: %+v", permanent.EntryChoices)
	}
	if result.Kind != game.ResolutionChoiceSubtype {
		t.Fatalf("recorded choice kind = %v, want subtype", result.Kind)
	}
	if result.Subtype != subtypes[pick] {
		t.Fatalf("recorded entry type = %q, want %q", result.Subtype, subtypes[pick])
	}
	if !types.KnownSubtypeForType(types.Creature, result.Subtype) {
		t.Fatalf("recorded entry type %q is not a known creature subtype", result.Subtype)
	}
}

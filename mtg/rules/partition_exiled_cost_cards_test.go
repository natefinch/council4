package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// coinOfFateAbility mirrors the activated ability generated for Coin of Fate:
// "{3}{W}, {T}, Exile two creature cards from your graveyard, Sacrifice this
// artifact: An opponent chooses one of the exiled cards. You put that card on the
// bottom of your library and return the other to the battlefield tapped. You
// become the monarch."
func coinOfFateAbility() *game.ActivatedAbility {
	return &game.ActivatedAbility{
		ManaCost: opt.Val(cost.Mana{cost.O(3), cost.W}),
		AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalTap},
			{
				Kind:          cost.AdditionalExile,
				Text:          "Exile two creature cards from your graveyard",
				Amount:        2,
				Source:        zone.Graveyard,
				MatchCardType: true,
				CardType:      types.Creature,
			},
			{Kind: cost.AdditionalSacrificeSource, Text: "Sacrifice this artifact", Amount: 1},
		},
		ZoneOfFunction: zone.Battlefield,
		Content: game.Mode{
			Sequence: []game.Instruction{
				{Primitive: game.PartitionExiledCostCards{
					ChooserOpponent:       true,
					ChosenToLibraryBottom: true,
					OtherEntersTapped:     true,
				}},
				{Primitive: game.BecomeMonarch{Player: game.ControllerReference()}},
			},
		}.Ability(),
	}
}

func coinOfFateCreatureCard(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}
}

// TestCoinOfFatePartitionsExiledCreaturesOnOpponentChoice exercises the full Coin
// of Fate activated ability: the cost exiles two creature cards from the
// controller's graveyard and sacrifices the artifact; on resolution an opponent
// chooses one exiled card, which goes to the bottom of the controller's library,
// while the other returns to the battlefield tapped under the controller's
// control and the controller becomes the monarch.
func TestCoinOfFatePartitionsExiledCreaturesOnOpponentChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	artifact := &game.CardDef{CardFace: game.CardFace{
		Name:               "Coin of Fate",
		Types:              []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{*coinOfFateAbility()},
	}}
	source := addCombatPermanent(g, game.Player1, artifact)
	first := addCardToGraveyard(g, game.Player1, coinOfFateCreatureCard("Graveyard One"))
	second := addCardToGraveyard(g, game.Player1, coinOfFateCreatureCard("Graveyard Two"))
	// Seed the library so bottom placement is distinguishable from the top.
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Library One"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Library Two"}})
	g.Players[game.Player1].ManaPool.Add(mana.W, 4)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Coin of Fate ability was not legal with two graveyard creatures and mana")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(Coin of Fate ability) = false, want true")
	}

	player := g.Players[game.Player1]
	if player.Graveyard.Contains(first) || player.Graveyard.Contains(second) ||
		!player.Exile.Contains(first) || !player.Exile.Contains(second) {
		t.Fatal("cost did not exile both creature cards from the graveyard")
	}
	if !source.Tapped {
		t.Fatal("cost did not tap the source")
	}
	if !player.Graveyard.Contains(source.CardInstanceID) {
		t.Fatal("cost did not sacrifice the source to the graveyard")
	}

	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("activated ability was not put on the stack")
	}
	if len(obj.ExiledAsCostIDs) != 2 {
		t.Fatalf("ExiledAsCostIDs = %v, want the two exiled cards", obj.ExiledAsCostIDs)
	}
	// The opponent chooses the card at index 1 to send to the library bottom; the
	// card at index 0 must return to the battlefield tapped.
	chosen := obj.ExiledAsCostIDs[1]
	other := obj.ExiledAsCostIDs[0]
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if !player.Library.Contains(chosen) {
		t.Fatal("opponent-chosen card was not put into the controller's library")
	}
	if bottom, ok := player.Library.Bottom(); !ok || bottom != chosen {
		t.Fatal("opponent-chosen card was not put on the bottom of the library")
	}
	if player.Exile.Contains(chosen) || player.Exile.Contains(other) {
		t.Fatal("exiled cards remained in exile after resolution")
	}
	returned := findBattlefieldPermanentForCard(g, other)
	if returned == nil {
		t.Fatal("the other exiled card did not return to the battlefield")
	}
	if returned.Controller != game.Player1 {
		t.Fatalf("returned permanent controller = %v, want Player1", returned.Controller)
	}
	if !returned.Tapped {
		t.Fatal("the other exiled card did not return to the battlefield tapped")
	}
	if !player.IsMonarch {
		t.Fatal("controller did not become the monarch")
	}
}

func findBattlefieldPermanentForCard(g *game.Game, cardID id.ID) *game.Permanent {
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return permanent
		}
	}
	return nil
}

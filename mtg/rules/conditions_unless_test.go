package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func compileUnlessCard(t *testing.T, card cardgen.ScryfallCard) *game.CardDef {
	t.Helper()
	defs, diagnostics, err := cardgen.CompileCardDefs(&card)
	if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
		t.Fatalf("compile %s: defs=%d diagnostics=%#v err=%v", card.Name, len(defs), diagnostics, err)
	}
	return defs[0]
}

func TestUnlessControllerSelectionResolution(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Unless Selection Test", Layout: "normal", TypeLine: "Instant",
		OracleText: "Draw a card. You lose 2 life unless you control a Villain.",
	})
	for _, tt := range []struct {
		name         string
		subtype      types.Sub
		controller   game.PlayerID
		wantLifeLost int
	}{
		{name: "no Villain", wantLifeLost: 2},
		{name: "own Villain", subtype: types.Villain, controller: game.Player1},
		{name: "opponent Villain", subtype: types.Villain, controller: game.Player2, wantLifeLost: 2},
		{name: "own other subtype", subtype: types.Rogue, controller: game.Player1, wantLifeLost: 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			startLife := g.Players[game.Player1].Life
			addCardToLibrary(g, game.Player1, evidenceCard("Drawn", 1))
			if tt.subtype != "" {
				addCombatPermanent(g, tt.controller, &game.CardDef{CardFace: game.CardFace{
					Name: "State subject", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{tt.subtype},
				}})
			}
			obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if got := startLife - g.Players[game.Player1].Life; got != tt.wantLifeLost {
				t.Fatalf("life lost = %d, want %d", got, tt.wantLifeLost)
			}
			if g.Players[game.Player1].Hand.Size() != 1 {
				t.Fatal("preceding unconditional draw did not resolve")
			}
		})
	}
}

func TestUnlessGraveyardThresholdResolution(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Shoreline Looter", Layout: "normal", TypeLine: "Creature — Rat Rogue", ManaCost: "{1}{U}",
		Power: new("1"), Toughness: new("1"),
		OracleText: "This creature can't be blocked.\nThreshold — Whenever this creature deals combat damage to a player, draw a card. Then discard a card unless there are seven or more cards in your graveyard.",
	})
	if len(def.TriggeredAbilities) != 1 || def.TriggeredAbilities[0].Trigger.InterveningCondition.Exists {
		t.Fatalf("trigger = %#v, want an unconditional trigger with a resolving body gate", def.TriggeredAbilities)
	}
	for _, count := range []int{0, 6, 7, 8} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for range count {
				addCardToGraveyard(g, game.Player1, evidenceCard("Graveyard card", 1))
			}
			for range 8 {
				addCardToGraveyard(g, game.Player2, evidenceCard("Opponent card", 1))
			}
			addCardToLibrary(g, game.Player1, evidenceCard("Drawn", 1))
			source := addCombatPermanent(g, game.Player1, def)
			obj := &game.StackObject{Kind: game.StackTriggeredAbility, Controller: game.Player1, SourceID: source.ObjectID}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			wantHand, wantGraveyard := 1, count
			if count < 7 {
				wantHand, wantGraveyard = 0, count+1
			}
			player := g.Players[game.Player1]
			if player.Hand.Size() != wantHand || player.Graveyard.Size() != wantGraveyard || player.Library.Size() != 0 {
				t.Fatalf("hand=%d graveyard=%d library=%d, want %d/%d/0", player.Hand.Size(), player.Graveyard.Size(), player.Library.Size(), wantHand, wantGraveyard)
			}
		})
	}
}

func taintedIndulgenceDef(t *testing.T) *game.CardDef {
	t.Helper()
	return compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Tainted Indulgence", Layout: "normal", TypeLine: "Instant", ManaCost: "{U}{B}",
		OracleText: "Draw two cards. Then discard a card unless there are five or more mana values among cards in your graveyard.",
	})
}

func TestUnlessGraveyardManaValuesResolution(t *testing.T) {
	t.Parallel()
	def := taintedIndulgenceDef(t)
	for _, count := range []int{0, 4, 5, 6} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for mv := range count {
				addCardToGraveyard(g, game.Player1, evidenceCard("Distinct mana value", mv))
				addCardToGraveyard(g, game.Player1, evidenceCard("Duplicate mana value", mv))
			}
			for mv := range 6 {
				addCardToGraveyard(g, game.Player2, evidenceCard("Opponent mana value", mv))
			}
			for range 2 {
				addCardToLibrary(g, game.Player1, evidenceCard("Drawn", 1))
			}
			obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			wantHand, wantGraveyard := 2, count*2
			if count < 5 {
				wantHand, wantGraveyard = 1, count*2+1
			}
			player := g.Players[game.Player1]
			if player.Hand.Size() != wantHand || player.Graveyard.Size() != wantGraveyard || player.Library.Size() != 0 {
				t.Fatalf("hand=%d graveyard=%d library=%d, want %d/%d/0", player.Hand.Size(), player.Graveyard.Size(), player.Library.Size(), wantHand, wantGraveyard)
			}
		})
	}
}

func TestUnlessGraveyardStateCheckedAfterDraw(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for mv := range 4 {
		addCardToGraveyard(g, game.Player1, evidenceCard("Existing mana value", mv))
	}
	dredger := dredgeCardDef(1)
	dredger.ManaCost = opt.Val(cost.Mana{cost.O(4)})
	addCardToGraveyard(g, game.Player1, dredger)
	// The first draw mills a duplicate value and removes the only value-4
	// card by dredging it; the second draw is ordinary. Five values become four
	// before the discard's gate is checked.
	addCardToLibrary(g, game.Player1, evidenceCard("Milled duplicate", 0))
	addCardToLibrary(g, game.Player1, evidenceCard("Drawn", 1))
	if got := controllerGraveyardManaValueCount(g, game.Player1); got != 5 {
		t.Fatalf("initial distinct values = %d, want 5", got)
	}
	obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: dredgeChoiceAgent{}}
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, taintedIndulgenceDef(t).SpellAbility.Val, agents, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want 1: the discard must see post-draw state, not the initial five values", got)
	}
	if got := g.Players[game.Player1].Graveyard.Size(); got != 6 {
		t.Fatalf("graveyard size = %d, want 6 after dredge and discard", got)
	}
}

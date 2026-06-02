package rules

import (
	"strconv"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestValidateCommanderConfigsAcceptsLegalDeck(t *testing.T) {
	config := legalCommanderConfig()

	errs := validateCommanderConfig(game.Player1, config)

	if len(errs) != 0 {
		t.Fatalf("errors = %+v, want none", errs)
	}
}

func TestNewGameTracksCommanderIDs(t *testing.T) {
	commander := commanderDef("Tracked Commander", color.Green)
	configs := [game.NumPlayers]game.PlayerConfig{
		game.Player1: {Commander: commander},
	}

	g := game.NewGame(configs)

	commanderID := g.Players[game.Player1].CommanderInstanceID
	if commanderID == 0 || !g.CommanderIDs[commanderID] {
		t.Fatalf("commander id = %v tracked=%v, want commander tracked in Game.CommanderIDs", commanderID, g.CommanderIDs[commanderID])
	}
}

func TestCommanderPermanentZoneChangesUseCommandZoneReplacement(t *testing.T) {
	tests := []struct {
		name string
		move func(*game.Game, *game.Permanent)
	}{
		{name: "destroy", move: func(g *game.Game, permanent *game.Permanent) { destroyPermanent(g, permanent.ObjectID) }},
		{name: "exile", move: func(g *game.Game, permanent *game.Permanent) { movePermanentToZone(g, permanent, game.ZoneExile) }},
		{name: "bounce", move: func(g *game.Game, permanent *game.Permanent) { movePermanentToZone(g, permanent, game.ZoneHand) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			commander := addCommanderPermanent(g, game.Player1)

			tt.move(g, commander)

			if !g.Players[game.Player1].CommandZone.Contains(commander.CardInstanceID) {
				t.Fatalf("commander was not moved to command zone for %s", tt.name)
			}
			if g.Players[game.Player1].Graveyard.Contains(commander.CardInstanceID) ||
				g.Players[game.Player1].Exile.Contains(commander.CardInstanceID) ||
				g.Players[game.Player1].Hand.Contains(commander.CardInstanceID) {
				t.Fatal("commander also appeared in a replaced destination zone")
			}
			assertEvent(t, g.Events, game.EventZoneChanged, func(event game.GameEvent) bool {
				return event.CardID == commander.CardInstanceID && event.FromZone == game.ZoneBattlefield && event.ToZone == game.ZoneCommand
			})
			assertNoEvent(t, g.Events, game.EventPermanentDied, func(event game.GameEvent) bool {
				return event.CardID == commander.CardInstanceID
			})
		})
	}
}

func TestCommanderCardZoneChangesUseCommandZoneReplacement(t *testing.T) {
	t.Run("discard", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		commanderID := addCommanderCardToHand(g, game.Player1)

		if !discardCardFromHand(g, game.Player1, commanderID) {
			t.Fatal("discardCardFromHand() = false, want true")
		}
		if !g.Players[game.Player1].CommandZone.Contains(commanderID) || g.Players[game.Player1].Graveyard.Contains(commanderID) {
			t.Fatal("discarded commander did not use command-zone replacement")
		}
		assertEvent(t, g.Events, game.EventCardDiscarded, func(event game.GameEvent) bool {
			return event.CardID == commanderID && event.ToZone == game.ZoneCommand
		})
	})
	t.Run("mill", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		commanderID := addCommanderCardToLibrary(g, game.Player1)

		millCards(g, game.Player1, 1)

		if !g.Players[game.Player1].CommandZone.Contains(commanderID) || g.Players[game.Player1].Graveyard.Contains(commanderID) {
			t.Fatal("milled commander did not use command-zone replacement")
		}
	})
	t.Run("surveil", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		commanderID := addCommanderCardToLibrary(g, game.Player1)
		agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}

		engine.surveilCards(g, agents, &TurnLog{}, game.Player1, 1)

		if !g.Players[game.Player1].CommandZone.Contains(commanderID) || g.Players[game.Player1].Graveyard.Contains(commanderID) {
			t.Fatal("surveilled commander did not use command-zone replacement")
		}
	})
	t.Run("stack", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		commanderID := addCommanderCardToHand(g, game.Player1)
		card, ok := g.GetCardInstance(commanderID)
		if !ok {
			t.Fatal("commander card instance not found")
		}
		g.Players[game.Player1].Hand.Remove(commanderID)
		g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: commanderID, Controller: game.Player1})

		obj, ok := g.Stack.Pop()
		if !ok {
			t.Fatal("stack is empty")
		}
		if !moveStackCardToGraveyard(g, obj, card) {
			t.Fatal("moveStackCardToGraveyard() = false, want true")
		}
		if !g.Players[game.Player1].CommandZone.Contains(commanderID) || g.Players[game.Player1].Graveyard.Contains(commanderID) {
			t.Fatal("countered commander spell did not use command-zone replacement")
		}
	})
}

func TestLegalActionsIncludeCommanderCastFromCommandZone(t *testing.T) {
	g := newCommanderCastGame(greenCommanderWithCost())
	engine := NewEngine(nil)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1
	commanderID := g.Players[game.Player1].CommanderInstanceID

	actions := engine.legalActions(g, game.Player1)

	if !containsAction(actions, action.CastCommanderSpell(commanderID, nil, 0, nil)) {
		t.Fatalf("legal actions did not include command-zone commander cast: %+v", actions)
	}
}

func TestApplyCommanderCastPaysTaxAndIncrementsCastCount(t *testing.T) {
	g := newCommanderCastGame(greenCommanderWithCost())
	engine := NewEngine(nil)
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Island)
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	player := g.Players[game.Player1]
	player.CommanderCastCount = 1
	commanderID := player.CommanderInstanceID
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastCommanderSpell(commanderID, nil, 0, nil)) {
		t.Fatal("applyAction commander cast = false, want true")
	}
	if player.CommanderCastCount != 2 {
		t.Fatalf("commander cast count = %d, want 2", player.CommanderCastCount)
	}
	obj, ok := g.Stack.Peek()
	if player.CommandZone.Contains(commanderID) || !ok || obj.SourceID != commanderID {
		t.Fatal("commander was not moved from command zone to stack")
	}
	if !forest.Tapped {
		t.Fatal("commander cast did not pay colored mana")
	}
	assertEvent(t, g.Events, game.EventSpellCast, func(event game.GameEvent) bool {
		return event.CardID == commanderID && event.FromZone == game.ZoneCommand && event.ToZone == game.ZoneStack
	})
}

func TestFailedCommanderTaxCastDoesNotMutate(t *testing.T) {
	g := newCommanderCastGame(greenCommanderWithCost())
	engine := NewEngine(nil)
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	player := g.Players[game.Player1]
	player.CommanderCastCount = 1
	commanderID := player.CommanderInstanceID
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1
	act := action.CastCommanderSpell(commanderID, nil, 0, nil)

	if engine.canCastSpellFromZoneWithKicker(g, game.Player1, commanderID, game.ZoneCommand, nil, 0, nil, false) {
		t.Fatal("canCast commander with tax = true with insufficient mana, want false")
	}
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction commander cast = true with insufficient tax, want false")
	}
	if forest.Tapped || !player.CommandZone.Contains(commanderID) || player.CommanderCastCount != 1 || g.Stack.Size() != 0 {
		t.Fatal("failed commander cast mutated mana, command zone, cast count, or stack")
	}
}

func TestCommanderZeroToughnessReplacementDoesNotLogDeath(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	commander := addCommanderPermanent(g, game.Player1)
	zero := game.PT{Value: 0}
	card, ok := g.GetCardInstance(commander.CardInstanceID)
	if !ok {
		t.Fatal("commander card instance not found")
	}
	card.Def.Power = optPT(zero)
	card.Def.Toughness = optPT(zero)

	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if len(deaths) != 0 {
		t.Fatalf("deaths = %+v, want command-zone replacement to avoid death log", deaths)
	}
	if !g.Players[game.Player1].CommandZone.Contains(commander.CardInstanceID) {
		t.Fatal("zero-toughness commander did not move to command zone")
	}
}

func TestValidateCommanderConfigRejectsWrongDeckSize(t *testing.T) {
	config := legalCommanderConfig()
	config.Deck = config.Deck[:98]

	errs := validateCommanderConfig(game.Player1, config)

	assertCommanderLegalityError(t, errs, "deck has 98 cards")
}

func TestValidateCommanderConfigRejectsDuplicateNonbasicButAllowsBasic(t *testing.T) {
	config := legalCommanderConfig()
	config.Deck[1] = config.Deck[0]
	config.Deck[2] = basicLandDef(types.Forest)
	config.Deck[3] = basicLandDef(types.Forest)

	errs := validateCommanderConfig(game.Player1, config)

	assertCommanderLegalityError(t, errs, "duplicate nonbasic")
	if countCommanderLegalityErrors(errs, string(types.Forest)) != 0 {
		t.Fatalf("basic duplicate produced errors: %+v", errs)
	}
}

func TestValidateCommanderConfigRejectsInvalidCommander(t *testing.T) {
	tests := []struct {
		name      string
		commander *game.CardDef
	}{
		{name: "nonlegendary", commander: creatureDef("Bear", color.Green)},
		{name: "noncreature", commander: &game.CardDef{Name: "types.Legendary Artifact", Supertypes: []types.Super{types.Legendary}, Types: []types.Card{types.Artifact}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := legalCommanderConfig()
			config.Commander = tt.commander

			errs := validateCommanderConfig(game.Player1, config)

			assertCommanderLegalityError(t, errs, "legendary creature")
		})
	}
}

func TestValidateCommanderConfigRejectsColorIdentityViolation(t *testing.T) {
	config := legalCommanderConfig()
	config.Deck[0] = creatureDef("Off Color Card", color.Blue)

	errs := validateCommanderConfig(game.Player1, config)

	assertCommanderLegalityError(t, errs, "outside commander's color identity")
}

func TestValidateCommanderConfigRejectsCommanderInDeck(t *testing.T) {
	config := legalCommanderConfig()
	config.Deck[0] = config.Commander

	errs := validateCommanderConfig(game.Player1, config)

	assertCommanderLegalityError(t, errs, "also present in deck")
}

func legalCommanderConfig() game.PlayerConfig {
	deck := make([]*game.CardDef, commanderDeckCardCount)
	for i := range deck {
		deck[i] = creatureDef("Green Card "+strconv.Itoa(i), color.Green)
	}
	return game.PlayerConfig{
		Name:      "Player",
		Commander: commanderDef("Green Commander", color.Green),
		Deck:      deck,
	}
}

func commanderDef(name string, colors ...color.Color) *game.CardDef {
	card := creatureDef(name, colors...)
	card.Supertypes = append(card.Supertypes, types.Legendary)
	return card
}

func creatureDef(name string, colors ...color.Color) *game.CardDef {
	return &game.CardDef{
		Name:          name,
		Types:         []types.Card{types.Creature},
		ColorIdentity: mana.NewColorIdentity(colors...),
	}
}

func basicLandDef(name types.Sub) *game.CardDef {
	return &game.CardDef{
		Name:       string(name),
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{name},
	}
}

func addCommanderPermanent(g *game.Game, owner game.PlayerID) *game.Permanent {
	permanent := addCombatPermanent(g, owner, commanderDef("Battlefield Commander", color.Green))
	trackCommanderID(g, owner, permanent.CardInstanceID)
	return permanent
}

func addCommanderCardToHand(g *game.Game, owner game.PlayerID) id.ID {
	cardID := addCardToHand(g, owner, commanderDef("Zone Commander", color.Green))
	trackCommanderID(g, owner, cardID)
	return cardID
}

func addCommanderCardToLibrary(g *game.Game, owner game.PlayerID) id.ID {
	cardID := addCardToLibrary(g, owner, commanderDef("Library Commander", color.Green))
	trackCommanderID(g, owner, cardID)
	return cardID
}

func newCommanderCastGame(commander *game.CardDef) *game.Game {
	configs := [game.NumPlayers]game.PlayerConfig{
		game.Player1: {Commander: commander},
	}
	return game.NewGame(configs)
}

func greenCommanderWithCost() *game.CardDef {
	commander := commanderDef("Castable Commander", color.Green)
	cost := mana.Cost{mana.G}
	commander.ManaCost = optCost(cost)
	return commander
}

func trackCommanderID(g *game.Game, owner game.PlayerID, cardID id.ID) {
	g.Players[owner].CommanderInstanceID = cardID
	if g.CommanderIDs == nil {
		g.CommanderIDs = make(map[id.ID]bool)
	}
	g.CommanderIDs[cardID] = true
}

func assertCommanderLegalityError(t *testing.T, errs []CommanderLegalityError, want string) {
	t.Helper()
	if countCommanderLegalityErrors(errs, want) == 0 {
		t.Fatalf("errors = %+v, want one containing %q", errs, want)
	}
}

func countCommanderLegalityErrors(errs []CommanderLegalityError, want string) int {
	count := 0
	for _, err := range errs {
		if strings.Contains(err.Reason, want) {
			count++
		}
	}
	return count
}

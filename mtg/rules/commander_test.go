package rules

import (
	"strconv"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
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

func TestNewGameTracksMultipleCommanders(t *testing.T) {
	primaryDef := commanderDef("Primary Commander", color.Green)
	partner := commanderDef("Partner Commander", color.Blue)
	configs := [game.NumPlayers]game.PlayerConfig{
		game.Player1: {Commanders: []*game.CardDef{primaryDef, partner}},
	}

	g := game.NewGame(configs)
	player := g.Players[game.Player1]
	commanders := player.CommandZone.All()

	if len(commanders) != 2 {
		t.Fatalf("command zone has %d commanders, want 2", len(commanders))
	}
	primaryCard, ok := g.GetCardInstance(player.CommanderInstanceID)
	if !ok || primaryCard.Def != primaryDef || !player.CommandZone.Contains(player.CommanderInstanceID) {
		t.Fatalf("primary commander = %#v, want %q in command zone", primaryCard, primaryDef.Name)
	}
	for _, commanderID := range commanders {
		if !g.CommanderIDs[commanderID] {
			t.Fatalf("commander id %v is not tracked", commanderID)
		}
	}
}

func TestMultipleCommanderColorIdentityUsesUnion(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{
		game.Player1: {Commanders: []*game.CardDef{
			commanderDef("Green Commander", color.Green),
			commanderDef("Blue Commander", color.Blue),
		}},
	})

	if got := commanderColorIdentityCount(g, game.Player1); got != 2 {
		t.Fatalf("commander color identity count = %d, want 2", got)
	}
	colors := commanderColorIdentityMana(g, game.Player1)
	found := map[mana.Color]bool{}
	for _, c := range colors {
		found[c] = true
	}
	if len(colors) != 2 || !found[mana.G] || !found[mana.U] {
		t.Fatalf("commander identity mana = %v, want green and blue", colors)
	}
}

func TestCommanderPermanentZoneChangesUseCommandZoneReplacement(t *testing.T) {
	tests := []struct {
		name string
		move func(*game.Game, *game.Permanent)
	}{
		{name: "destroy", move: func(g *game.Game, permanent *game.Permanent) { destroyPermanent(g, permanent.ObjectID) }},
		{name: "exile", move: func(g *game.Game, permanent *game.Permanent) { movePermanentToZone(g, permanent, zone.Exile) }},
		{name: "bounce", move: func(g *game.Game, permanent *game.Permanent) { movePermanentToZone(g, permanent, zone.Hand) }},
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
			assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
				return event.CardID == commander.CardInstanceID && event.FromZone == zone.Battlefield && event.ToZone == zone.Command
			})
			assertNoEvent(t, g.Events, game.EventPermanentDied, func(event game.Event) bool {
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
		assertEvent(t, g.Events, game.EventCardDiscarded, func(event game.Event) bool {
			return event.CardID == commanderID && event.ToZone == zone.Command
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
	assertEvent(t, g.Events, game.EventSpellCast, func(event game.Event) bool {
		return event.CardID == commanderID && event.FromZone == zone.Command && event.ToZone == zone.Stack
	})
}

func TestPartnerCommanderCastsShareHistoryAndKeepSeparateTax(t *testing.T) {
	primary := commanderDef("Primary Commander", color.Green)
	primary.ManaCost = opt.Val(cost.Mana{})
	partnerDef := commanderDef("Partner Commander", color.Blue)
	partnerDef.ManaCost = opt.Val(cost.Mana{})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{
		game.Player1: {Commanders: []*game.CardDef{primary, partnerDef}},
	})
	engine := NewEngine(nil)
	player := g.Players[game.Player1]
	primaryID := player.CommanderInstanceID
	var partnerID id.ID
	for _, commanderID := range player.CommandZone.All() {
		if commanderID != primaryID {
			partnerID = commanderID
			break
		}
	}
	if partnerID == 0 {
		t.Fatal("partner commander was not created")
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	legal := engine.legalActions(g, game.Player1)
	if !containsAction(legal, action.CastCommanderSpell(primaryID, nil, 0, nil)) ||
		!containsAction(legal, action.CastCommanderSpell(partnerID, nil, 0, nil)) {
		t.Fatalf("legal actions do not include both commanders: %+v", legal)
	}
	if !engine.applyAction(g, game.Player1, action.CastCommanderSpell(partnerID, nil, 0, nil)) {
		t.Fatal("casting partner commander failed")
	}
	if got := player.CommanderTaxFor(primaryID); got != 0 {
		t.Fatalf("uncast primary commander tax = %d, want 0", got)
	}
	g.Stack = game.Stack{}
	if !engine.applyAction(g, game.Player1, action.CastCommanderSpell(primaryID, nil, 0, nil)) {
		t.Fatal("casting primary commander failed")
	}

	if player.CommanderCastCount != 2 {
		t.Fatalf("aggregate commander cast count = %d, want 2", player.CommanderCastCount)
	}
	if got := player.CommanderCastCountFor(primaryID); got != 1 {
		t.Fatalf("primary cast count = %d, want 1", got)
	}
	if got := player.CommanderCastCountFor(partnerID); got != 1 {
		t.Fatalf("partner cast count = %d, want 1", got)
	}
	if player.CommanderTaxFor(primaryID) != 2 || player.CommanderTaxFor(partnerID) != 2 {
		t.Fatalf("commander taxes = primary %d partner %d, want 2 each",
			player.CommanderTaxFor(primaryID), player.CommanderTaxFor(partnerID))
	}
}

func lifeForCommanderTaxCommander() *game.CardDef {
	commander := commanderDef("Liesa, Shroud of Dusk", color.White, color.Black)
	commander.ManaCost = opt.Val(cost.Mana{cost.G})
	commander.StaticAbilities = []game.StaticAbility{{
		ZoneOfFunction: zone.Command,
		RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectPayLifeForCommanderTax,
			AffectedPlayer: game.PlayerYou,
			AffectedSource: true,
		}},
	}}
	return commander
}

func TestApplyCommanderCastPaysTaxWithLife(t *testing.T) {
	g := newCommanderCastGame(lifeForCommanderTaxCommander())
	engine := NewEngine(nil)
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	player := g.Players[game.Player1]
	player.CommanderCastCount = 2
	startLife := player.Life
	commanderID := player.CommanderInstanceID
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastCommanderSpell(commanderID, nil, 0, nil)) {
		t.Fatal("applyAction commander cast = false, want true")
	}
	if player.CommanderCastCount != 3 {
		t.Fatalf("commander cast count = %d, want 3", player.CommanderCastCount)
	}
	if !forest.Tapped {
		t.Fatal("commander cast did not pay the {G} base cost with mana")
	}
	if got := startLife - player.Life; got != 4 {
		t.Fatalf("life paid for tax = %d, want 4 (two {2} instances at 2 life each)", got)
	}
	obj, ok := g.Stack.Peek()
	if player.CommandZone.Contains(commanderID) || !ok || obj.SourceID != commanderID {
		t.Fatal("commander was not moved from command zone to stack")
	}
}

func TestApplyCommanderCastPaysTaxWithManaWhenAvailable(t *testing.T) {
	g := newCommanderCastGame(lifeForCommanderTaxCommander())
	engine := NewEngine(nil)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Island)
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	player := g.Players[game.Player1]
	player.CommanderCastCount = 1
	startLife := player.Life
	commanderID := player.CommanderInstanceID
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastCommanderSpell(commanderID, nil, 0, nil)) {
		t.Fatal("applyAction commander cast = false, want true")
	}
	if got := startLife - player.Life; got != 0 {
		t.Fatalf("life paid = %d, want 0 (tax paid with mana when available)", got)
	}
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

	if engine.canCastSpellFromZoneWithKicker(g, game.Player1, commanderID, zone.Command, nil, 0, nil, false) {
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
	card.Def.Power = opt.Val(zero)
	card.Def.Toughness = opt.Val(zero)

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

func TestValidateCommanderConfigAcceptsTwoCommanders(t *testing.T) {
	partner := func(name string, c color.Color) *game.CardDef {
		card := commanderDef(name, c)
		card.StaticAbilities = []game.StaticAbility{game.PartnerStaticBody}
		return card
	}
	backgroundCommander := commanderDef("Background Chooser", color.White)
	backgroundCommander.StaticAbilities = []game.StaticAbility{game.ChooseABackgroundStaticBody}
	background := &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Background",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment},
			Subtypes:   []types.Sub{types.Background},
		},
		ColorIdentity: color.NewIdentity(color.Black),
	}
	tests := []struct {
		name       string
		commanders []*game.CardDef
	}{
		{name: "partner", commanders: []*game.CardDef{partner("First Partner", color.Green), partner("Second Partner", color.Blue)}},
		{name: "background", commanders: []*game.CardDef{backgroundCommander, background}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := game.PlayerConfig{Commanders: tt.commanders, Deck: make([]*game.CardDef, 98)}
			for i := range config.Deck {
				config.Deck[i] = creatureDef("Colorless Card " + strconv.Itoa(i))
			}

			if errs := validateCommanderConfig(game.Player1, config); len(errs) != 0 {
				t.Fatalf("errors = %+v, want none", errs)
			}
		})
	}
}

func TestValidateCommanderConfigChecksPartnerPairing(t *testing.T) {
	partnerWith := func(name, other string) *game.CardDef {
		card := commanderDef(name)
		card.StaticAbilities = []game.StaticAbility{game.PartnerWithStaticBody}
		card.OracleText = "Partner with " + other + " (When this creature enters, target player may search their library for a card named " + other + ".)"
		return card
	}
	restrictedPartner := func(name, quality string) *game.CardDef {
		card := commanderDef(name)
		card.StaticAbilities = []game.StaticAbility{game.PartnerStaticBody}
		card.OracleText = "Partner—" + quality
		return card
	}
	tests := []struct {
		name       string
		commanders []*game.CardDef
		wantError  bool
	}{
		{
			name:       "matching partner with",
			commanders: []*game.CardDef{partnerWith("Alpha", "Beta"), partnerWith("Beta", "Alpha")},
		},
		{
			name:       "unrelated partner with",
			commanders: []*game.CardDef{partnerWith("Alpha", "Beta"), partnerWith("Gamma", "Delta")},
			wantError:  true,
		},
		{
			name:       "matching restricted partner",
			commanders: []*game.CardDef{restrictedPartner("Alpha", "Survivors"), restrictedPartner("Beta", "Survivors")},
		},
		{
			name:       "mismatched restricted partner",
			commanders: []*game.CardDef{restrictedPartner("Alpha", "Survivors"), restrictedPartner("Beta", "Friends forever")},
			wantError:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := game.PlayerConfig{Commanders: tt.commanders, Deck: make([]*game.CardDef, 98)}
			for i := range config.Deck {
				config.Deck[i] = creatureDef("Colorless Card " + strconv.Itoa(i))
			}

			errs := validateCommanderConfig(game.Player1, config)
			if tt.wantError {
				assertCommanderLegalityError(t, errs, "partner")
			} else if len(errs) != 0 {
				t.Fatalf("errors = %+v, want none", errs)
			}
		})
	}
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
		{name: "noncreature", commander: &game.CardDef{CardFace: game.CardFace{Name: "types.Legendary Artifact", Supertypes: []types.Super{types.Legendary}, Types: []types.Card{types.Artifact}}}},
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
	deck := make([]*game.CardDef, commanderTotalCardCount-1)
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
	return &game.CardDef{CardFace: game.CardFace{Name: name,
		Types: []types.Card{types.Creature}}, ColorIdentity: color.NewIdentity(colors...),
	}
}

func basicLandDef(name types.Sub) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: string(name),
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{name}},
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
	manaCost := cost.Mana{cost.G}
	commander.ManaCost = opt.Val(manaCost)
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

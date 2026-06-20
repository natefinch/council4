package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestReplacementRegistrationSkipsETBReplacementEffects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{
		Name:                 "Tapped Bear",
		Types:                []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{game.EntersTappedReplacement("This creature enters tapped.")},
	}}

	permanent := addReplacementPermanent(t, g, game.Player1, def)
	if !permanent.Tapped {
		t.Fatal("ETB replacement did not tap entering permanent")
	}
	if len(g.ReplacementEffects) != 0 {
		t.Fatalf("registered replacement effects = %d, want 0", len(g.ReplacementEffects))
	}
}

func TestReplacementRegistrationSkipsSelfZoneReplacementEffects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, selfLibraryReplacementCardDef())
	other := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Other Creature",
		Types: []types.Card{types.Creature},
	}})

	if len(g.ReplacementEffects) != 0 {
		t.Fatalf("registered replacement effects = %d, want 0", len(g.ReplacementEffects))
	}
	if !movePermanentToZone(g, other, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if !g.Players[game.Player1].Graveyard.Contains(other.CardInstanceID) {
		t.Fatal("other permanent was not put into graveyard")
	}
	if g.Players[game.Player1].Library.Contains(other.CardInstanceID) {
		t.Fatal("self-zone replacement affected a different permanent")
	}
}

func addReplacementPermanent(t *testing.T, g *game.Game, controller game.PlayerID, def *game.CardDef) *game.Permanent {
	t.Helper()
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   def,
		Owner: controller,
	}
	permanent, ok := createCardPermanent(g, g.CardInstances[cardID], controller, zone.Hand)
	if !ok {
		t.Fatal("createCardPermanent() = false, want true")
	}
	return permanent
}

func TestGenericETBReplacementAppliesTappedAndCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.CreateReplacement{
		Replacement: &game.ReplacementEffect{
			Description:        "enter modified",
			MatchEvent:         game.EventPermanentEnteredBattlefield,
			MatchToZone:        true,
			ToZone:             zone.Battlefield,
			EntersTapped:       true,
			EntersWithCounters: []game.CounterPlacement{{Kind: counter.PlusOnePlusOne, Amount: 1}},
		},
	}, nil)
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Entering Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})},
	})
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)

	permanent, ok := createCardPermanent(g, card, game.Player1, zone.Hand)
	if !ok || !permanent.Tapped {
		t.Fatalf("permanent = %+v, want tapped by replacement", permanent)
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters = %d, want 1", got)
	}
}

func TestMultipleGenericReplacementsRecordOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	for _, replacement := range []game.ReplacementEffect{
		{
			Description:   "exile instead",
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      zone.Battlefield,
			MatchToZone:   true,
			ToZone:        zone.Graveyard,
			ReplaceToZone: zone.Exile,
		},
		{
			Description:   "hand instead",
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      zone.Battlefield,
			ReplaceToZone: zone.Hand,
		},
	} {
		resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.CreateReplacement{
			Replacement: &replacement,
		}, nil)
	}

	if !movePermanentToZone(g, target, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %+v, want one order decision", g.ReplacementDecisions)
	}
	decision := g.ReplacementDecisions[0]
	if decision.Player != game.Player1 || len(decision.Selected) != 2 || decision.Selected[0] != 0 || decision.Selected[1] != 1 {
		t.Fatalf("replacement decision = %+v, want deterministic Player1 order", decision)
	}
	if !g.Players[game.Player1].Hand.Contains(target.CardInstanceID) {
		t.Fatal("second replacement in fallback order should move card to hand")
	}
}

func TestPermanentSourceReplacementStopsAfterSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Replacement Source",
		Types: []types.Card{types.Enchantment}},
	})
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	resolveInstruction(engine, g, &game.StackObject{
		Kind:         game.StackActivatedAbility,
		Controller:   game.Player1,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
	}, game.CreateReplacement{
		Replacement: &game.ReplacementEffect{
			Description:   "exile instead",
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      zone.Battlefield,
			MatchToZone:   true,
			ToZone:        zone.Graveyard,
			ReplaceToZone: zone.Exile,
		},
	}, nil)

	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("source should leave battlefield")
	}
	if !movePermanentToZone(g, target, zone.Graveyard) {
		t.Fatal("target should move to graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("replacement from departed source should not apply")
	}
}

func TestGroupEntersTappedReplacementTapsOpponentPermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	authority := &game.CardDef{CardFace: game.CardFace{
		Name:  "Authority of the Consuls",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedGroupReplacement(
				"Creatures your opponents control enter tapped.",
				game.TriggerControllerOpponent,
				types.Creature,
			),
		},
	}}
	addReplacementPermanent(t, g, game.Player1, authority)
	if len(g.ReplacementEffects) != 1 {
		t.Fatalf("registered replacement effects = %d, want 1", len(g.ReplacementEffects))
	}

	creatureDef := func() *game.CardDef {
		return &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}}
	}
	opponentCreature := addReplacementPermanent(t, g, game.Player2, creatureDef())
	if !opponentCreature.Tapped {
		t.Fatal("opponent creature should enter tapped")
	}
	ownCreature := addReplacementPermanent(t, g, game.Player1, creatureDef())
	if ownCreature.Tapped {
		t.Fatal("controller's own creature should not enter tapped")
	}
	opponentLand := addReplacementPermanent(t, g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Wastes",
		Types: []types.Card{types.Land},
	}})
	if opponentLand.Tapped {
		t.Fatal("opponent land should not enter tapped under a creature-only filter")
	}
}

func TestSkipStepEffectSkipsNextDrawStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Would Draw"}})
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.SkipStep{
		Player: game.ControllerReference(),
		Step:   game.StepDraw,
	}, nil)

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("hand size = %d, want skipped draw step", got)
	}
	if g.Players[game.Player1].Library.Size() != 1 {
		t.Fatalf("library size = %d, want card not drawn", g.Players[game.Player1].Library.Size())
	}
}

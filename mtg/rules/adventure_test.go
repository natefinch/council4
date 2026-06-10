package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestAdventureAlternateFaceLegalFromHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, adventureCreatureCard())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)
	if !actionsContain(legal, action.CastSpellFace(cardID, game.FaceAlternate, nil, 0, nil)) {
		t.Fatalf("legal actions = %+v, want CastSpellFace(alternate)", legal)
	}
}

func TestAdventureAlternateFaceCastPutsOnStackWithCorrectFace(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, adventureCreatureCard())
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceAlternate, nil, 0, nil)) {
		t.Fatal("applyAction CastSpellFace(alternate) = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Face != game.FaceAlternate {
		t.Fatalf("stack object = %+v, want alternate face", obj)
	}
}

func TestAdventureAlternateFaceResolvesAndExilesCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, adventureCreatureCard())
	drawnID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceAlternate, nil, 0, nil)) {
		t.Fatal("applyAction CastSpellFace(alternate) = false, want true")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("adventure spell went to graveyard")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("adventure spell was not exiled")
	}
	if !g.AdventureCards[cardID] {
		t.Fatal("adventure spell was not tracked in AdventureCards")
	}
	if !g.Players[game.Player1].Hand.Contains(drawnID) {
		t.Fatal("adventure spell did not resolve its draw effect")
	}
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == cardID && event.Face == game.FaceAlternate && event.FromZone == zone.Stack && event.ToZone == zone.Exile
	})
	assertEvent(t, g.Events, game.EventSpellResolved, func(event game.Event) bool {
		return event.CardID == cardID
	})
}

func TestAdventureCreatureFaceLegalFromAdventureExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := resolveAdventureToExile(t, g, engine, adventureCreatureCard())

	legal := engine.legalActions(g, game.Player1)
	if !actionsContain(legal, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatalf("legal actions = %+v, want exile CastSpellFace(front)", legal)
	}
	if actionsContain(legal, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceAlternate, nil, 0, nil)) {
		t.Fatalf("legal actions = %+v, did not want exile CastSpellFace(alternate)", legal)
	}
}

func TestAdventureCreatureFaceResolvesFromExileToNewBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := resolveAdventureToExile(t, g, engine, adventureCreatureCard())

	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatal("applyAction CastSpellFaceFromZone(exile, front) = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Face != game.FaceFront || obj.SourceZone != zone.Exile {
		t.Fatalf("stack object = %+v, want front face from exile", obj)
	}
	if g.AdventureCards[cardID] {
		t.Fatal("AdventureCards entry was not cleared when creature was cast from exile")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("adventure creature remained in exile after being cast")
	}
	permanent := permanentForCard(g, cardID)
	if permanent == nil || permanent.Face != game.FaceFront || !permanentHasType(g, permanent, types.Creature) {
		t.Fatalf("permanent = %+v, want front-face creature on battlefield", permanent)
	}
}

func TestAdventureFizzledGoesToGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, targetedAdventureCard())
	target := addCreaturePermanent(g, game.Player2)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceAlternate, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)) {
		t.Fatal("applyAction CastSpellFace(alternate, target) = false, want true")
	}
	movePermanentToZone(g, target, zone.Graveyard)
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("fizzled adventure spell did not go to graveyard")
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("fizzled adventure spell was exiled")
	}
	if g.AdventureCards[cardID] {
		t.Fatal("fizzled adventure spell was tracked in AdventureCards")
	}
}

func TestAdventureCounteredGoesToGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, adventureCreatureCard())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceAlternate, nil, 0, nil)) {
		t.Fatal("applyAction CastSpellFace(alternate) = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("adventure spell was not put on the stack")
	}
	if !counterStackObject(g, obj.ID) {
		t.Fatal("counterStackObject(adventure) = false, want true")
	}

	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("countered adventure spell did not go to graveyard")
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("countered adventure spell was exiled")
	}
	if g.AdventureCards[cardID] {
		t.Fatal("countered adventure spell was tracked in AdventureCards")
	}
}

func TestAdventureCreatureFaceNotLegalFromOrdinaryExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, adventureCreatureCard())
	g.Players[game.Player1].Hand.Remove(cardID)
	g.Players[game.Player1].Exile.Add(cardID)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)
	if actionsContain(legal, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatalf("legal actions = %+v, did not want exile CastSpellFace(front)", legal)
	}
}

func TestAdventurePermissionClearsWhenCardLeavesExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := resolveAdventureToExile(t, g, engine, adventureCreatureCard())

	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Exile, zone.Graveyard) {
		t.Fatal("moving Adventure card from exile to graveyard failed")
	}
	if g.AdventureCards[cardID] {
		t.Fatal("Adventure permission remained after card left exile")
	}
	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Graveyard, zone.Exile) {
		t.Fatal("re-exiling Adventure card failed")
	}

	legal := engine.legalActions(g, game.Player1)
	if actionsContain(legal, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatalf("legal actions = %+v, did not want Adventure cast after ordinary re-exile", legal)
	}
}

func resolveAdventureToExile(t *testing.T, g *game.Game, engine *Engine, def *game.CardDef) game.ObjectID {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, def)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceAlternate, nil, 0, nil)) {
		t.Fatal("applyAction CastSpellFace(alternate) = false, want true")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("adventure spell did not move to exile")
	}
	if !g.AdventureCards[cardID] {
		t.Fatal("adventure spell was not tracked in AdventureCards")
	}
	return cardID
}

func adventureCreatureCard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Adventure Creature",
			Types:     []types.Card{types.Creature},
			ManaCost:  opt.Val(cost.Mana{cost.O(1), cost.G}),
			Colors:    []color.Color{color.Green},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name:     "Adventure Spell",
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Adventure},
			ManaCost: opt.Val(cost.Mana{cost.G}),
			Colors:   []color.Color{color.Green},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
				}},
			}.Ability()),
		}),
	}
}

func targetedAdventureCard() *game.CardDef {
	card := adventureCreatureCard()
	card.Alternate = opt.Val(game.CardFace{
		Name:     "Targeted Adventure",
		Types:    []types.Card{types.Sorcery},
		Subtypes: []types.Sub{types.Adventure},
		ManaCost: opt.Val(cost.Mana{cost.G}),
		Colors:   []color.Color{color.Green},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target creature",
				Allow:      game.TargetAllowPermanent,
				Predicate: game.TargetPredicate{
					PermanentTypes: []types.Card{types.Creature},
				},
			}},
			Sequence: []game.Instruction{{
				Primitive: game.Tap{Object: game.TargetPermanentReference(0)},
			}},
		}.Ability()),
	})
	return card
}

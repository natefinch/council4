package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestLegalActionsIncludeFaceDownCastAtSorcerySpeed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, morphCreature(cost.Mana{cost.G}))
	g.Players[game.Player1].ManaPool.Add(mana.G, 3)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	actions := engine.legalActions(g, game.Player1)

	for _, act := range actions {
		if payload, ok := act.CastFaceDownPayload(); ok && payload.CardID == cardID && payload.FaceDownKind == game.FaceDownMorph {
			return
		}
	}
	t.Fatalf("legal actions = %+v, want face-down cast action", actions)
}

func TestLegalActionsExcludeFaceDownCastOutsideSorcerySpeed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, morphCreature(cost.Mana{cost.G}))
	g.Players[game.Player1].ManaPool.Add(mana.G, 3)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	actions := engine.legalActions(g, game.Player1)

	for _, act := range actions {
		if payload, ok := act.CastFaceDownPayload(); ok && payload.CardID == cardID {
			t.Fatalf("legal actions = %+v, did not want face-down cast outside sorcery speed", actions)
		}
	}
}

func TestCastFaceDownResolvesAsTwoTwoCreatureWithHiddenIdentity(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, morphCreature(cost.Mana{cost.G}))
	g.Players[game.Player1].ManaPool.Add(mana.G, 3)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, actionBuild.castFaceDown(cardID, game.FaceFront, game.FaceDownMorph)) {
		t.Fatal("face-down cast action failed")
	}
	if g.Players[game.Player1].Hand.Contains(cardID) || g.Stack.Size() != 1 {
		t.Fatal("face-down card did not move hand -> stack")
	}
	obj, _ := g.Stack.Peek()
	if !obj.FaceDown || obj.FaceDownKind != game.FaceDownMorph || obj.FaceDownFace != game.FaceFront {
		t.Fatalf("stack object face-down state = %+v", obj)
	}
	assertEvent(t, g.Events, game.EventSpellCast, func(event game.Event) bool {
		return event.CardID == cardID && len(event.Colors) == 0
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if len(g.Battlefield) != 1 {
		t.Fatalf("battlefield size = %d, want 1", len(g.Battlefield))
	}
	permanent := g.Battlefield[0]
	if !permanent.FaceDown || permanent.FaceDownKind != game.FaceDownMorph || permanent.CardInstanceID != cardID {
		t.Fatalf("permanent face-down state = %+v", permanent)
	}
	if !permanentHasType(g, permanent, types.Creature) || effectivePower(g, permanent) != 2 {
		t.Fatalf("face-down effective characteristics typeCreature=%t power=%d", permanentHasType(g, permanent, types.Creature), effectivePower(g, permanent))
	}
	if permanentEffectiveName(g, permanent) != "" || len(permanentEffectiveAbilities(g, permanent)) != 0 {
		t.Fatalf("face-down visible name/abilities = %q/%d, want hidden/no abilities", permanentEffectiveName(g, permanent), len(permanentEffectiveAbilities(g, permanent)))
	}
}

func TestTurnFaceUpPaysMorphCostAndRevealsPrintedCharacteristics(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addFaceDownPermanent(g, game.Player1, morphCreature(cost.Mana{cost.G}), game.FaceDownMorph)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID)) {
		t.Fatal("turn face-up action failed")
	}

	if permanent.FaceDown {
		t.Fatal("permanent remained face-down")
	}
	if permanentEffectiveName(g, permanent) != "Mystery Bear" || effectivePower(g, permanent) != 3 {
		t.Fatalf("face-up characteristics name=%q power=%d, want Mystery Bear/3", permanentEffectiveName(g, permanent), effectivePower(g, permanent))
	}
	if !hasEvent(g, game.EventCardRevealed) || !hasEvent(g, game.EventPermanentTurnedFaceUp) {
		t.Fatalf("events = %+v, want reveal and turned-face-up events", g.Events)
	}
}

func TestDisguiseTurnFaceUpGrantsNoShieldAndFaceDownHasWard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addFaceDownPermanent(g, game.Player1, disguiseCreature(cost.Mana{cost.W}), game.FaceDownDisguise)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	g.Turn.PriorityPlayer = game.Player1

	abilities := permanentEffectiveAbilities(g, permanent)
	if len(abilities) != 1 {
		t.Fatalf("face-down disguise abilities = %+v, want ward ability", abilities)
	}
	staticBody, ok := abilities[0].(*game.StaticAbility)
	if !ok || !game.BodyHasKeyword(staticBody, game.Ward) {
		t.Fatalf("face-down disguise abilities = %+v, want ward ability", abilities)
	}

	if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID)) {
		t.Fatal("turn face-up action failed")
	}

	// Disguise turn-up grants no shield counter (CR 702.168d); the disguise
	// effect simply ends and the permanent regains its normal characteristics.
	if got := permanent.Counters.Get(counter.Shield); got != 0 {
		t.Fatalf("shield counters = %d, want 0", got)
	}
	if permanent.FaceDownKind != game.FaceDownNone {
		t.Fatalf("FaceDownKind = %v, want none after turn face up", permanent.FaceDownKind)
	}
}

func TestManifestTurnFaceUpRequiresCreatureCard(t *testing.T) {
	t.Run("creature can turn face up for mana cost", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		permanent := addFaceDownPermanent(g, game.Player1, manifestCreature(cost.Mana{cost.G}), game.FaceDownManifest)
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		g.Turn.PriorityPlayer = game.Player1

		if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID)) {
			t.Fatal("manifest creature turn face-up action failed")
		}

		if permanent.FaceDown {
			t.Fatal("manifest creature remained face-down")
		}
		if permanentEffectiveName(g, permanent) != "Manifest Bear" || effectivePower(g, permanent) != 3 {
			t.Fatalf("face-up manifest characteristics name=%q power=%d, want Manifest Bear/3", permanentEffectiveName(g, permanent), effectivePower(g, permanent))
		}
	})
	t.Run("noncreature cannot turn face up", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		permanent := addFaceDownPermanent(g, game.Player1, manifestNoncreature(cost.Mana{cost.G}), game.FaceDownManifest)
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		g.Turn.PriorityPlayer = game.Player1

		if engine.canTurnFaceUp(g, game.Player1, permanent.ObjectID) {
			t.Fatal("manifest noncreature was allowed to turn face up")
		}
		if engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID)) {
			t.Fatal("manifest noncreature turn face-up action succeeded")
		}
		if !permanent.FaceDown {
			t.Fatal("manifest noncreature stopped being face-down")
		}
	})
	t.Run("creature can turn face up for morph cost when mana cost is not payable", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		permanent := addFaceDownPermanent(g, game.Player1, manifestMorphCreature(cost.Mana{cost.W}, cost.Mana{cost.G}), game.FaceDownManifest)
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		g.Turn.PriorityPlayer = game.Player1

		if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID)) {
			t.Fatal("manifested morph creature did not turn face up for morph cost")
		}
		if permanent.FaceDown || permanentEffectiveName(g, permanent) != "Manifest Morph Bear" {
			t.Fatalf("manifested morph creature state name=%q faceDown=%t", permanentEffectiveName(g, permanent), permanent.FaceDown)
		}
	})
	t.Run("noncreature with morph can turn face up for morph cost", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		permanent := addFaceDownPermanent(g, game.Player1, manifestMorphNoncreature(cost.Mana{cost.G}), game.FaceDownManifest)
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		g.Turn.PriorityPlayer = game.Player1

		if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID)) {
			t.Fatal("manifested noncreature with morph did not turn face up for morph cost")
		}
		if permanent.FaceDown || permanentEffectiveName(g, permanent) != "Manifest Morph Land" {
			t.Fatalf("manifested morph noncreature state name=%q faceDown=%t", permanentEffectiveName(g, permanent), permanent.FaceDown)
		}
	})
	t.Run("creature chooses between mana cost and morph cost", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		permanent := addFaceDownPermanent(g, game.Player1, manifestMorphCreature(cost.Mana{cost.W}, cost.Mana{cost.G}), game.FaceDownManifest)
		g.Players[game.Player1].ManaPool.Add(mana.W, 1)
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		g.Turn.PriorityPlayer = game.Player1
		agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
		log := TurnLog{}

		if !engine.applyActionWithChoices(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID), agents, &log) {
			t.Fatal("manifested morph creature did not turn face up with chosen morph cost")
		}

		if permanent.FaceDown {
			t.Fatal("manifested morph creature remained face-down")
		}
		if got := g.Players[game.Player1].ManaPool.Amount(mana.W); got != 1 {
			t.Fatalf("white mana = %d, want selected morph route to leave white mana unspent", got)
		}
		if got := g.Players[game.Player1].ManaPool.Amount(mana.G); got != 0 {
			t.Fatalf("green mana = %d, want selected morph route to spend green mana", got)
		}
		if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoicePayment || log.Choices[0].Selected[0] != 1 {
			t.Fatalf("turn face-up choice log = %+v, want selected morph cost route", log.Choices)
		}
	})
}

func morphCreature(manaCost cost.Mana) *game.CardDef {
	pt := game.PT{Value: 3}
	return &game.CardDef{CardFace: game.CardFace{Name: "Mystery Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.MorphKeyword{Cost: manaCost}},
		}}},
	}
}

func manifestCreature(manaCost cost.Mana) *game.CardDef {
	pt := game.PT{Value: 3}
	return &game.CardDef{CardFace: game.CardFace{Name: "Manifest Bear",
		ManaCost:  opt.Val(manaCost),
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt)},
	}
}

func manifestNoncreature(manaCost cost.Mana) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Manifest Spell",
		ManaCost: opt.Val(manaCost),
		Types:    []types.Card{types.Instant}},
	}
}

func manifestMorphCreature(manaCost, morphCost cost.Mana) *game.CardDef {
	pt := game.PT{Value: 3}
	return &game.CardDef{CardFace: game.CardFace{Name: "Manifest Morph Bear",
		ManaCost:  opt.Val(manaCost),
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.MorphKeyword{Cost: morphCost}},
		}}},
	}
}

func manifestMorphNoncreature(morphCost cost.Mana) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Manifest Morph Land",
		Types: []types.Card{types.Land},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.MorphKeyword{Cost: morphCost}},
		}}},
	}
}

func disguiseCreature(manaCost cost.Mana) *game.CardDef {
	pt := game.PT{Value: 2}
	return &game.CardDef{CardFace: game.CardFace{Name: "Veiled Guard",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.DisguiseKeyword{Cost: manaCost}},
		}}},
	}
}

func addFaceDownPermanent(g *game.Game, controller game.PlayerID, def *game.CardDef, kind game.FaceDownKind) *game.Permanent {
	cardID := addCardToHand(g, controller, def)
	card, _ := g.GetCardInstance(cardID)
	permanent, ok := createCardPermanentFaceDown(g, card, controller, zone.Stack, game.FaceFront, kind, true)
	if !ok {
		panic("test failed to create face-down permanent")
	}
	g.Players[controller].Hand.Remove(cardID)
	return permanent
}

func hasEvent(g *game.Game, kind game.EventKind) bool {
	for _, event := range g.Events {
		if event.Kind == kind {
			return true
		}
	}
	return false
}

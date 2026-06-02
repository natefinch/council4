package rules

import (
	"testing"

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

func TestDisguiseTurnFaceUpAddsShieldAndFaceDownHasWard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addFaceDownPermanent(g, game.Player1, disguiseCreature(cost.Mana{cost.W}), game.FaceDownDisguise)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	g.Turn.PriorityPlayer = game.Player1

	abilities := permanentEffectiveAbilities(g, permanent)
	if len(abilities) != 1 || !abilityHasKeyword(&abilities[0], game.Ward) || !abilities[0].WardCost.Exists {
		t.Fatalf("face-down disguise abilities = %+v, want ward ability", abilities)
	}

	if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID)) {
		t.Fatal("turn face-up action failed")
	}

	if got := permanent.Counters.Get(counter.Shield); got != 1 {
		t.Fatalf("shield counters = %d, want 1", got)
	}
	if permanent.FaceDownKind != game.FaceDownNone {
		t.Fatalf("FaceDownKind = %v, want none after turn face up", permanent.FaceDownKind)
	}
}

func morphCreature(manaCost cost.Mana) *game.CardDef {
	pt := game.PT{Value: 3}
	return &game.CardDef{
		Name:      "Mystery Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		Abilities: []game.AbilityDef{{
			Kind:      game.StaticAbility,
			Keywords:  []game.Keyword{game.Morph},
			MorphCost: opt.Val(manaCost),
		}},
	}
}

func disguiseCreature(manaCost cost.Mana) *game.CardDef {
	pt := game.PT{Value: 2}
	return &game.CardDef{
		Name:      "Veiled Guard",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		Abilities: []game.AbilityDef{{
			Kind:         game.StaticAbility,
			Keywords:     []game.Keyword{game.Disguise},
			DisguiseCost: opt.Val(manaCost),
		}},
	}
}

func addFaceDownPermanent(g *game.Game, controller game.PlayerID, def *game.CardDef, kind game.FaceDownKind) *game.Permanent {
	cardID := addCardToHand(g, controller, def)
	card, _ := g.GetCardInstance(cardID)
	permanent, ok := createCardPermanentFaceDown(g, card, controller, game.ZoneStack, game.FaceFront, kind)
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

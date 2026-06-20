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

func TestVandalblastNormalAndOverloadCasting(t *testing.T) {
	t.Run("normal costs one red and targets one opposing artifact", func(t *testing.T) {
		g, engine, spellID := overloadTestGame()
		target := addOverloadArtifact(g, game.Player2, false)
		g.Players[game.Player1].ManaPool.Add(mana.R, 1)

		act := action.CastSpell(spellID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
		if !containsAction(engine.legalActions(g, game.Player1), act) {
			t.Fatal("normal Vandalblast action is not legal")
		}
		if !engine.applyAction(g, game.Player1, act) {
			t.Fatal("normal Vandalblast cast failed")
		}
		obj, ok := g.Stack.Peek()
		if !ok || obj.Overloaded || len(obj.Targets) != 1 {
			t.Fatalf("normal stack object = %#v", obj)
		}
		if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
			t.Fatalf("mana remaining = %d, want 0", got)
		}
	})

	t.Run("overload costs five mana and has no targets", func(t *testing.T) {
		g, engine, spellID := overloadTestGame()
		hexproof := addOverloadArtifact(g, game.Player2, true)
		g.Players[game.Player1].ManaPool.Add(mana.R, 4)
		overloaded := action.CastOverloadedSpellFaceFromZone(spellID, zone.Hand, game.FaceFront, nil)
		if containsAction(engine.legalActions(g, game.Player1), overloaded) {
			t.Fatal("overload action is legal with only four mana")
		}
		g.Players[game.Player1].ManaPool.Add(mana.R, 1)
		actions := engine.legalActions(g, game.Player1)
		if !containsAction(actions, overloaded) {
			t.Fatal("overload action is not legal with five mana")
		}
		normal := action.CastSpell(spellID, []game.Target{game.PermanentTarget(hexproof.ObjectID)}, 0, nil)
		if containsAction(actions, normal) {
			t.Fatal("normal cast can target an opposing hexproof artifact")
		}
		if !engine.applyAction(g, game.Player1, overloaded) {
			t.Fatal("overloaded Vandalblast cast failed")
		}
		obj, ok := g.Stack.Peek()
		if !ok || !obj.Overloaded || len(obj.Targets) != 0 {
			t.Fatalf("overloaded stack object = %#v", obj)
		}
		if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
			t.Fatalf("mana remaining = %d, want 0", got)
		}
	})
}

func TestVandalblastOverloadDestroysEachOpposingArtifactSimultaneously(t *testing.T) {
	g, engine, spellID := overloadTestGame()
	own := addOverloadArtifact(g, game.Player1, false)
	first := addOverloadArtifact(g, game.Player2, true)
	second := addOverloadArtifact(g, game.Player3, false)
	g.Players[game.Player1].ManaPool.Add(mana.R, 5)

	act := action.CastOverloadedSpellFaceFromZone(spellID, zone.Hand, game.FaceFront, nil)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("overloaded Vandalblast cast failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, own.ObjectID); !ok {
		t.Fatal("controller's own artifact was destroyed")
	}
	for _, permanent := range []*game.Permanent{first, second} {
		if _, ok := permanentByObjectID(g, permanent.ObjectID); ok {
			t.Fatalf("opposing artifact %v survived overload", permanent.ObjectID)
		}
		if !g.Players[permanent.Owner].Graveyard.Contains(permanent.CardInstanceID) {
			t.Fatalf("opposing artifact %v did not enter its graveyard", permanent.ObjectID)
		}
	}

	batches := make(map[id.ID]bool)
	for _, event := range g.Events {
		if event.Kind != game.EventPermanentDied ||
			(event.PermanentID != first.ObjectID && event.PermanentID != second.ObjectID) {
			continue
		}
		if event.SimultaneousID == 0 {
			t.Fatalf("death event has no simultaneous ID: %#v", event)
		}
		batches[event.SimultaneousID] = true
	}
	if len(batches) != 1 {
		t.Fatalf("opposing artifacts died in %d batches, want 1", len(batches))
	}
}

func TestVandalblastOverloadDoesNotReplaceForcedFlashback(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	def := vandalblastDefinition()
	def.AlternativeCosts = []cost.Alternative{{
		Label:    "Flashback",
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
	}}
	spellID := addCardToGraveyard(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.R, 5)
	engine := NewEngine(nil)

	if engine.canCastOverloadedSpellFaceFromZone(
		g,
		game.Player1,
		spellID,
		zone.Graveyard,
		game.FaceFront,
		nil,
	) {
		t.Fatal("overload replaced the mandatory flashback alternative cost")
	}
}

func TestOverloadCombinesWithKicker(t *testing.T) {
	g, engine, spellID := overloadTestGameWithDefinition(func(def *game.CardDef) {
		def.StaticAbilities = []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.R}}},
		}}
	})
	g.Players[game.Player1].ManaPool.Add(mana.R, 6)
	act := action.CastOverloadedSpellFaceFromZoneWithOptions(
		spellID,
		zone.Hand,
		game.FaceFront,
		0,
		nil,
		true,
	)
	if !engine.canCastOverloadedSpellFaceFromZoneWithOptions(
		g,
		game.Player1,
		spellID,
		zone.Hand,
		game.FaceFront,
		0,
		nil,
		true,
	) {
		t.Fatal("kicked overload is not payable")
	}
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("kicked overload action is not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("kicked overload cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Overloaded || !obj.KickerPaid {
		t.Fatalf("stack object = %#v, want overloaded and kicked", obj)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana remaining = %d, want 0", got)
	}
}

func TestOverloadKickerPreservesPhyrexianPaymentChoice(t *testing.T) {
	g, engine, spellID := overloadTestGameWithDefinition(func(def *game.CardDef) {
		def.StaticAbilities = []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{
				Cost: cost.Mana{cost.PhyrexianMana(mana.R)},
			}},
		}}
	})
	g.Players[game.Player1].ManaPool.Add(mana.R, 6)
	act := action.CastOverloadedSpellFaceFromZoneWithOptions(
		spellID,
		zone.Hand,
		game.FaceFront,
		0,
		nil,
		true,
	)
	if !engine.canCastOverloadedSpellFaceFromZoneWithOptions(
		g,
		game.Player1,
		spellID,
		zone.Hand,
		game.FaceFront,
		0,
		nil,
		true,
	) {
		t.Fatal("phyrexian-kicked overload is not payable")
	}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := &TurnLog{}
	if !engine.applyActionWithChoices(g, game.Player1, act, agents, log) {
		t.Fatal("phyrexian-kicked overload cast failed")
	}
	if got := g.Players[game.Player1].Life; got != 38 {
		t.Fatalf("life = %d, want 38", got)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 1 {
		t.Fatalf("mana remaining = %d, want 1 after paying kicker with life", got)
	}
	if len(log.Choices) != 1 ||
		log.Choices[0].Request.Kind != game.ChoicePayment ||
		log.Choices[0].Selected[0] != 1 {
		t.Fatalf("payment choice log = %+v, want phyrexian life choice", log.Choices)
	}
}

func TestOverloadCombinesWithAdditionalCostX(t *testing.T) {
	g, engine, spellID := overloadTestGameWithDefinition(func(def *game.CardDef) {
		def.Overload.Val.Cost = cost.Mana{cost.R}
		def.AdditionalCosts = []cost.Additional{{
			Kind:        cost.AdditionalExile,
			Text:        "exile X cards from your graveyard",
			AmountFromX: true,
			Source:      zone.Graveyard,
		}}
	})
	firstID := addCardToHand(g, game.Player1, greenCreature())
	g.Players[game.Player1].Hand.Remove(firstID)
	g.Players[game.Player1].Graveyard.Add(firstID)
	secondID := addCardToHand(g, game.Player1, greenCreature())
	g.Players[game.Player1].Hand.Remove(secondID)
	g.Players[game.Player1].Graveyard.Add(secondID)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)

	act := action.CastOverloadedSpellFaceFromZoneWithOptions(
		spellID,
		zone.Hand,
		game.FaceFront,
		2,
		nil,
		false,
	)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("overload action with additional-cost X is not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("overload cast with additional-cost X failed")
	}
	if g.Players[game.Player1].Graveyard.Contains(firstID) ||
		g.Players[game.Player1].Graveyard.Contains(secondID) {
		t.Fatal("overload additional cost did not exile the announced X cards")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Overloaded || obj.XValue != 2 || len(obj.AdditionalCostsPaid) != 1 {
		t.Fatalf("stack object = %#v, want overloaded X=2 with one additional cost", obj)
	}
}

func TestOverloadDoesNotUseAnotherAlternativeCost(t *testing.T) {
	g, engine, spellID := overloadTestGameWithDefinition(func(def *game.CardDef) {
		def.AlternativeCosts = []cost.Alternative{{
			Label:    "Cheap alternative",
			ManaCost: opt.Val(cost.Mana{cost.R}),
		}}
	})
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	act := action.CastOverloadedSpellFaceFromZone(spellID, zone.Hand, game.FaceFront, nil)
	if containsAction(engine.legalActions(g, game.Player1), act) ||
		engine.canCastOverloadedSpellFaceFromZone(g, game.Player1, spellID, zone.Hand, game.FaceFront, nil) {
		t.Fatal("overload cast used the card's incompatible alternative cost")
	}
	g.Players[game.Player1].ManaPool.Add(mana.R, 4)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("overload cast is not legal when its own cost can be paid")
	}
}

func TestOverloadSelectedBackFaceCanBeCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	back := vandalblastDefinition().CardFace
	def := &game.CardDef{
		CardFace: game.CardFace{
			Name:         "Front",
			ManaCost:     opt.Val(cost.Mana{cost.U}),
			Types:        []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{}.Ability()),
		},
		Layout: game.LayoutModalDFC,
		Back:   opt.Val(back),
	}
	spellID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.R, 5)
	engine := NewEngine(nil)
	act := action.CastOverloadedSpellFaceFromZone(spellID, zone.Hand, game.FaceBack, nil)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("back-face overload action is not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("back-face overload cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Face != game.FaceBack || !obj.Overloaded {
		t.Fatalf("stack object = %#v, want overloaded back face", obj)
	}
}

func TestOverloadCommanderCastIsEnumerated(t *testing.T) {
	def := vandalblastDefinition()
	g := newCommanderCastGame(def)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1
	g.Players[game.Player1].ManaPool.Add(mana.R, 5)
	engine := NewEngine(nil)
	spellID := g.Players[game.Player1].CommanderInstanceID
	act := action.CastOverloadedSpellFaceFromZone(spellID, zone.Command, game.FaceFront, nil)

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("command-zone overload action is not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("command-zone overload cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Overloaded || obj.SourceZone != zone.Command {
		t.Fatalf("stack object = %#v, want overloaded command-zone cast", obj)
	}
}

func TestOverloadSiblingGroupsResolveSimultaneously(t *testing.T) {
	selection := game.Selection{
		RequiredTypes: []types.Card{types.Artifact},
		Controller:    game.ControllerNotYou,
	}
	t.Run("tap", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		first := addOverloadArtifact(g, game.Player2, false)
		second := addOverloadArtifact(g, game.Player3, false)
		resolveInstruction(
			NewEngine(nil),
			g,
			&game.StackObject{Controller: game.Player1},
			game.Tap{Group: game.BattlefieldGroup(selection)},
			&TurnLog{},
		)
		if !first.Tapped || !second.Tapped {
			t.Fatal("group tap did not tap every opposing artifact")
		}
		assertSharedEventBatch(t, g.Events, game.EventPermanentTapped, first.ObjectID, second.ObjectID)
	})
	t.Run("bounce", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		first := addOverloadArtifact(g, game.Player2, false)
		second := addOverloadArtifact(g, game.Player3, false)
		resolveInstruction(
			NewEngine(nil),
			g,
			&game.StackObject{Controller: game.Player1},
			game.Bounce{Group: game.BattlefieldGroup(selection)},
			&TurnLog{},
		)
		if !g.Players[first.Owner].Hand.Contains(first.CardInstanceID) ||
			!g.Players[second.Owner].Hand.Contains(second.CardInstanceID) {
			t.Fatal("group bounce did not return every opposing artifact")
		}
		assertSharedEventBatch(t, g.Events, game.EventZoneChanged, first.ObjectID, second.ObjectID)
	})
}

func assertSharedEventBatch(
	t *testing.T,
	events []game.Event,
	kind game.EventKind,
	permanentIDs ...id.ID,
) {
	t.Helper()
	wanted := make(map[id.ID]bool, len(permanentIDs))
	for _, permanentID := range permanentIDs {
		wanted[permanentID] = true
	}
	var batch id.ID
	seen := 0
	for _, event := range events {
		if event.Kind != kind || !wanted[event.PermanentID] {
			continue
		}
		seen++
		if event.SimultaneousID == 0 {
			t.Fatalf("event has no simultaneous ID: %#v", event)
		}
		if batch == 0 {
			batch = event.SimultaneousID
		} else if event.SimultaneousID != batch {
			t.Fatalf("event simultaneous ID = %v, want %v", event.SimultaneousID, batch)
		}
	}
	if seen != len(permanentIDs) {
		t.Fatalf("saw %d matching events, want %d", seen, len(permanentIDs))
	}
}

func overloadTestGame() (*game.Game, *Engine, id.ID) {
	return overloadTestGameWithDefinition(func(*game.CardDef) {})
}

func overloadTestGameWithDefinition(modify func(*game.CardDef)) (*game.Game, *Engine, id.ID) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	def := vandalblastDefinition()
	modify(def)
	return g, NewEngine(nil), addCardToHand(g, game.Player1, def)
}

func vandalblastDefinition() *game.CardDef {
	selection := game.Selection{
		RequiredTypes: []types.Card{types.Artifact},
		Controller:    game.ControllerNotYou,
	}
	normal := game.Mode{
		Targets: []game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Allow:      game.TargetAllowPermanent,
			Selection:  opt.Val(selection),
		}},
		Sequence: []game.Instruction{{
			Primitive: game.Destroy{Object: game.TargetPermanentReference(0)},
		}},
	}.Ability()
	overloaded := game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.Destroy{Group: game.BattlefieldGroup(selection)},
		}},
	}.Ability()
	return &game.CardDef{CardFace: game.CardFace{
		Name:         "Vandalblast",
		ManaCost:     opt.Val(cost.Mana{cost.R}),
		Types:        []types.Card{types.Sorcery},
		SpellAbility: opt.Val(normal),
		Overload: opt.Val(game.OverloadAbility{
			Cost:         cost.Mana{cost.O(4), cost.R},
			SpellAbility: overloaded,
		}),
	}}
}

func addOverloadArtifact(g *game.Game, controller game.PlayerID, hexproof bool) *game.Permanent {
	face := game.CardFace{
		Name:  "Test Artifact",
		Types: []types.Card{types.Artifact},
	}
	if hexproof {
		face.StaticAbilities = []game.StaticAbility{game.HexproofStaticBody}
	}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: face})
}

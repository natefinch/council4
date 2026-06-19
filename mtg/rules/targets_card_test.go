package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestCardTargetedSpellTypeUnionMatchesEitherMember confirms the disjunctive
// RequiredTypesAny semantics the graveyard-return union lowering relies on: a
// card that is only an enchantment (never a creature) is a legal target for a
// "creature or enchantment card" spell. The historical lowering bug set a
// conjunctive RequiredTypes alongside the union, which would have excluded this
// pure enchantment.
func TestCardTargetedSpellTypeUnionMatchesEitherMember(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	enchantmentID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Enchantment",
		Types: []types.Card{types.Enchantment},
	}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Land",
		Types: []types.Card{types.Land},
	}})
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Recovery Spell",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      game.TargetAllowCard,
				TargetZone: zone.Graveyard,
				Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Enchantment}, Controller: game.ControllerYou}),
			}},
			Sequence: []game.Instruction{{Primitive: game.MoveCard{
				Card:        game.CardReference{Kind: game.CardReferenceTarget},
				FromZone:    zone.Graveyard,
				Destination: zone.Hand,
			}}},
		}.Ability()),
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	want := action.CastSpell(spellID, []game.Target{currentCardTarget(t, g, enchantmentID)}, 0, nil)
	if !containsAction(legal, want) {
		t.Fatalf("legal actions did not include graveyard enchantment target action %+v", want)
	}
}

func TestCardTargetedSpellCreatesActionsForMatchingGraveyardCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	instantID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Instant",
		Types: []types.Card{types.Instant},
	}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Land",
		Types: []types.Card{types.Land},
	}})
	battlefieldCreature := addCreaturePermanent(g, game.Player1)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Regrow Spell",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      game.TargetAllowCard,
				TargetZone: zone.Graveyard,
				Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}, Controller: game.ControllerYou}),
			}},
			Sequence: []game.Instruction{{Primitive: game.MoveCard{
				Card:        game.CardReference{Kind: game.CardReferenceTarget},
				FromZone:    zone.Graveyard,
				Destination: zone.Hand,
			}}},
		}.Ability()),
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	want := action.CastSpell(spellID, []game.Target{currentCardTarget(t, g, instantID)}, 0, nil)
	if !containsAction(legal, want) {
		t.Fatalf("legal actions did not include graveyard instant target action %+v", want)
	}
	for _, act := range legal {
		cast, ok := act.CastSpellPayload()
		if !ok || cast.CardID != spellID {
			continue
		}
		if len(cast.Targets) != 1 || cast.Targets[0] != currentCardTarget(t, g, instantID) {
			t.Fatalf("unexpected card target %+v; battlefield creature was %+v", cast.Targets, battlefieldCreature)
		}
	}
}

func TestCardTargetedSpellMatchesCardsWithCyclingInGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cyclingID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Cycling Card",
		ActivatedAbilities: []game.ActivatedAbility{
			game.CyclingActivatedAbility(cost.Mana{cost.W}),
		},
	}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Plain Card",
	}})
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Excavation",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      game.TargetAllowCard,
				TargetZone: zone.Graveyard,
				Selection: opt.Val(game.Selection{
					Keyword:    game.Cycling,
					Controller: game.ControllerYou,
				}),
			}},
			Sequence: []game.Instruction{{Primitive: game.MoveCard{
				Card:        game.CardReference{Kind: game.CardReferenceTarget},
				FromZone:    zone.Graveyard,
				Destination: zone.Hand,
			}}},
		}.Ability()),
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	want := action.CastSpell(spellID, []game.Target{currentCardTarget(t, g, cyclingID)}, 0, nil)
	if !containsAction(legal, want) {
		t.Fatalf("legal actions did not include cycling-card target action %+v", want)
	}
	for _, act := range legal {
		cast, ok := act.CastSpellPayload()
		if !ok || cast.CardID != spellID {
			continue
		}
		if len(cast.Targets) != 1 || cast.Targets[0] != currentCardTarget(t, g, cyclingID) {
			t.Fatalf("unexpected cycling card target %+v", cast.Targets)
		}
	}
}

func TestIndexedCardTargetReferencesMoveMultipleTargetCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	firstID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First Cycling"}})
	secondID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second Cycling"}})
	sourceID := addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.MoveCard{
			Card:        game.CardReference{Kind: game.CardReferenceTarget},
			FromZone:    zone.Graveyard,
			Destination: zone.Hand,
		}},
		{Primitive: game.MoveCard{
			Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1},
			FromZone:    zone.Graveyard,
			Destination: zone.Hand,
		}},
	}, []game.Target{currentCardTarget(t, g, firstID), currentCardTarget(t, g, secondID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 0,
		MaxTargets: 2,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
	}}

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(firstID) || !g.Players[game.Player1].Hand.Contains(secondID) {
		t.Fatalf("hand = %+v, want both target cards moved", g.Players[game.Player1].Hand.All())
	}
}

func currentCardTarget(t *testing.T, g *game.Game, cardID id.ID) game.Target {
	t.Helper()
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatalf("card %v not found", cardID)
	}
	return game.CardTargetWithZoneVersion(cardID, card.ZoneVersion)
}

func TestCardTargetThatLeavesZoneBeforeResolutionCountersSpellByRules(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targetID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Instant",
		Types: []types.Card{types.Instant},
	}})
	sourceID := addEffectSpellToStack(g, game.Player1, game.MoveCard{
		Card:        game.CardReference{Kind: game.CardReferenceTarget},
		FromZone:    zone.Graveyard,
		Destination: zone.Hand,
	}, []game.Target{game.CardTarget(targetID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Instant}, Controller: game.ControllerYou}),
	}}
	g.Players[game.Player1].Graveyard.Remove(targetID)
	g.Players[game.Player1].Exile.Add(targetID)
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if len(log.Resolves) != 1 || log.Resolves[0].Result != "countered by rules" {
		t.Fatalf("resolve log = %+v, want countered by rules", log.Resolves)
	}
	if !g.Players[game.Player1].Exile.Contains(targetID) {
		t.Fatal("target card left exile")
	}
}

func TestCardTargetThatLeavesAndReturnsBeforeResolutionCountersSpellByRules(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targetID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Instant",
		Types: []types.Card{types.Instant},
	}})
	sourceID := addEffectSpellToStack(g, game.Player1, game.MoveCard{
		Card:        game.CardReference{Kind: game.CardReferenceTarget},
		FromZone:    zone.Graveyard,
		Destination: zone.Hand,
	}, []game.Target{currentCardTarget(t, g, targetID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Instant}, Controller: game.ControllerYou}),
	}}
	if !moveCardBetweenZones(g, game.Player1, targetID, zone.Graveyard, zone.Exile) {
		t.Fatal("failed to move target to exile")
	}
	if !moveCardBetweenZones(g, game.Player1, targetID, zone.Exile, zone.Graveyard) {
		t.Fatal("failed to return target to graveyard")
	}
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if len(log.Resolves) != 1 || log.Resolves[0].Result != "countered by rules" {
		t.Fatalf("resolve log = %+v, want countered by rules", log.Resolves)
	}
	if !g.Players[game.Player1].Graveyard.Contains(targetID) {
		t.Fatal("target card left graveyard")
	}
}

func TestMoveCardCanPutTargetOnBottomOfLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targetID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Card",
		Types: []types.Card{types.Instant},
	}})
	topID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top Card"}})
	sourceID := addEffectSpellToStack(g, game.Player1, game.MoveCard{
		Card:              game.CardReference{Kind: game.CardReferenceTarget},
		FromZone:          zone.Graveyard,
		Destination:       zone.Library,
		DestinationBottom: true,
	}, []game.Target{currentCardTarget(t, g, targetID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
	}}

	engine.resolveTopOfStack(g, &TurnLog{})

	if top, ok := g.Players[game.Player1].Library.Top(); !ok || top != topID {
		t.Fatalf("library top = %v, %v; want existing top %v", top, ok, topID)
	}
	if bottom, ok := g.Players[game.Player1].Library.Bottom(); !ok || bottom != targetID {
		t.Fatalf("library bottom = %v, %v; want target %v", bottom, ok, targetID)
	}
}

func TestMoveCardCanExileTargetedGraveyardCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targetID := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Graveyard Creature",
		Types: []types.Card{types.Creature},
	}})
	sourceID := addEffectSpellToStack(g, game.Player1, game.MoveCard{
		Card:        game.CardReference{Kind: game.CardReferenceTarget},
		FromZone:    zone.Graveyard,
		Destination: zone.Exile,
	}, []game.Target{currentCardTarget(t, g, targetID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection:  opt.Val(game.Selection{Controller: game.ControllerOpponent}),
	}}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player2].Graveyard.Contains(targetID) {
		t.Fatal("target card remained in graveyard")
	}
	if !g.Players[game.Player2].Exile.Contains(targetID) {
		t.Fatal("target card did not move to its owner's exile")
	}
}

func TestPutOnBattlefieldCanUseTargetedGraveyardCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targetID := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Opponent Graveyard Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	sourceID := addEffectSpellToStack(g, game.Player1, game.PutOnBattlefield{
		Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
		Recipient: opt.Val(game.ControllerReference()),
	}, []game.Target{currentCardTarget(t, g, targetID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	}}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player2].Graveyard.Contains(targetID) {
		t.Fatal("target card remained in graveyard")
	}
	permanent := permanentByCardID(g, targetID)
	if permanent == nil {
		t.Fatal("target card was not put onto the battlefield")
	}
	if permanent.Controller != game.Player1 {
		t.Fatalf("permanent controller = %v, want Player1", permanent.Controller)
	}
}

func TestPutOnBattlefieldEntryOptionsAreAtomic(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Recursive Construct",
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		SourceID:     sourceID,
		SourceCardID: sourceID,
		Controller:   game.Player1,
		InlineActivated: &game.ActivatedAbility{Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.PutOnBattlefield{
			Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
			EntryTapped:   true,
			EntryCounters: []game.CounterPlacement{{Kind: counter.PlusOnePlusOne, Amount: 2}},
		}}}}.Ability()},
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	permanent := permanentByCardID(g, sourceID)
	if permanent == nil {
		t.Fatal("returned card not on battlefield")
	}
	if !permanent.Tapped {
		t.Fatal("returned permanent did not enter tapped")
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("+1/+1 counters = %d, want 2", got)
	}
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentTapped || event.Kind == game.EventCountersAdded {
			t.Fatalf("entry option emitted follow-up event: %+v", event)
		}
	}
}

func TestSourceCardReferenceRequiresSameGraveyardIncarnation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Recursive Construct",
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	source, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	g.Stack.Push(&game.StackObject{
		ID:                g.IDGen.Next(),
		Kind:              game.StackActivatedAbility,
		SourceID:          sourceID,
		SourceCardID:      sourceID,
		SourceZone:        zone.Graveyard,
		SourceZoneVersion: source.ZoneVersion,
		Controller:        game.Player1,
		InlineActivated: &game.ActivatedAbility{Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.PutOnBattlefield{
			Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
		}}}}.Ability()},
	})
	if !moveCardBetweenZones(g, game.Player1, sourceID, zone.Graveyard, zone.Exile) {
		t.Fatal("failed to move source to exile")
	}
	if !moveCardBetweenZones(g, game.Player1, sourceID, zone.Exile, zone.Graveyard) {
		t.Fatal("failed to return source to graveyard")
	}
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if permanentByCardID(g, sourceID) != nil {
		t.Fatal("stale source incarnation returned to battlefield")
	}
	if !g.Players[game.Player1].Graveyard.Contains(sourceID) {
		t.Fatal("source card left graveyard")
	}
}

// TestOrderedSequenceDestroyThenGraveyardReturnResolvesInOrder proves the
// runtime resolves a two-instruction ordered sequence in order: a destroy of
// permanent target slot 0 followed by a graveyard-to-hand move that reads card
// target slot 0 (the second overall target, but the first card target, since the
// runtime numbers card references among card targets only). This is the
// resolution behavior the cardgen sequence target rebaser enables for
// "destroy ... return target card from your graveyard" bodies.
func TestOrderedSequenceDestroyThenGraveyardReturnResolvesInOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	destroyTarget := addCreaturePermanent(g, game.Player2)
	returnCardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Buried Bear",
		Types: []types.Card{types.Creature},
	}})
	sourceID := addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}},
		{Primitive: game.MoveCard{
			Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 0},
			FromZone:    zone.Graveyard,
			Destination: zone.Hand,
		}},
	}, []game.Target{
		game.PermanentTarget(destroyTarget.ObjectID),
		currentCardTarget(t, g, returnCardID),
	})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent},
		{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowCard, TargetZone: zone.Graveyard},
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, destroyTarget.ObjectID); ok {
		t.Fatal("first instruction did not destroy the target permanent")
	}
	if !g.Players[game.Player1].Hand.Contains(returnCardID) {
		t.Fatalf("hand = %+v, want second-slot graveyard card returned", g.Players[game.Player1].Hand.All())
	}
	if g.Players[game.Player1].Graveyard.Contains(returnCardID) {
		t.Fatal("returned card remained in graveyard")
	}
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const gilraenBlinkKey = game.LinkedKey("immediate-blink-1")

type gilraenChoiceAgent struct {
	accept bool
}

func (gilraenChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a gilraenChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoiceMay:
		if a.accept {
			return []int{1}
		}
		return []int{0}
	case game.ChoiceTarget:
		if len(request.Options) > 0 {
			return []int{len(request.Options) - 1}
		}
	default:
	}
	return request.DefaultSelection
}

func gilraenBlinkInstructions(entryCounters []game.CounterPlacement) []game.Instruction {
	return []game.Instruction{
		{
			Primitive: game.Exile{
				Object:         game.TargetPermanentReference(0),
				ExileLinkedKey: gilraenBlinkKey,
			},
		},
		{
			Primitive:     game.PutOnBattlefield{Source: game.LinkedBattlefieldSource(gilraenBlinkKey)},
			Optional:      true,
			PublishResult: game.ResultKey("if-you-do"),
		},
		{
			Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
				Timing:       game.DelayedAtBeginningOfNextEndStep,
				CapturedCard: opt.Val(game.LinkedObjectReference(string(gilraenBlinkKey))),
				Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.PutOnBattlefield{
					Source:            game.CardBattlefieldSource(game.CapturedCardReference()),
					EntryCounters:     entryCounters,
					LinkedReturnZones: []zone.Type{zone.Exile},
				}}}}.Ability(),
			}},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       game.ResultKey("if-you-do"),
				Succeeded: game.TriFalse,
			}),
		},
	}
}

func resolveGilraenBlink(
	g *game.Game,
	engine *Engine,
	source *game.Permanent,
	target *game.Permanent,
	acceptImmediate bool,
	entryCounters []game.CounterPlacement,
) {
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   source.Controller,
		Targets:      []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	var agents [game.NumPlayers]PlayerAgent
	agents[source.Controller] = gilraenChoiceAgent{accept: acceptImmediate}
	log := &TurnLog{}
	instructions := gilraenBlinkInstructions(entryCounters)
	for i := range instructions {
		engine.resolveInstructionWithChoices(g, obj, &instructions[i], agents, log)
	}
}

func gilraenCounters() []game.CounterPlacement {
	return []game.CounterPlacement{
		{Kind: counter.Vigilance, Amount: 1},
		{Kind: counter.Lifelink, Amount: 1},
	}
}

func gilraenTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Gilraen, Dúnedain Protector",
		Types:      []types.Card{types.Creature},
		Supertypes: []types.Super{types.Legendary},
		ActivatedAbilities: []game.ActivatedAbility{{
			ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
			AdditionalCosts: cost.Tap,
			ZoneOfFunction:  zone.Battlefield,
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Creature},
						Controller:       game.ControllerYou,
						ExcludeSource:    true,
					}),
				}},
				Sequence: gilraenBlinkInstructions(gilraenCounters()),
			}.Ability(),
		}},
	}}
}

func TestGilraenCardActivationCostsAndTargetLegality(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	gilraen := addCombatPermanent(g, game.Player1, gilraenTestDef())
	ownCreature := addCombatCreaturePermanent(g, game.Player1)
	opponentCreature := addCombatCreaturePermanent(g, game.Player2)
	ownLand := addBasicLandPermanent(g, game.Player1, types.Plains)
	secondLand := addBasicLandPermanent(g, game.Player1, types.Plains)

	activate := func(target *game.Permanent) action.Action {
		return action.ActivateAbility(
			gilraen.ObjectID,
			0,
			[]game.Target{game.PermanentTarget(target.ObjectID)},
			0,
		)
	}
	legal := engine.legalActions(g, game.Player1)
	if !containsAction(legal, activate(ownCreature)) {
		t.Fatal("Gilraen could not target another creature its controller controls")
	}
	for name, target := range map[string]*game.Permanent{
		"itself":                gilraen,
		"opponent creature":     opponentCreature,
		"noncreature permanent": ownLand,
	} {
		if containsAction(legal, activate(target)) {
			t.Fatalf("Gilraen activation illegally targeted %s", name)
		}
	}
	var chosen action.Action
	found := false
	for _, candidate := range legal {
		if containsAction([]action.Action{candidate}, activate(ownCreature)) {
			chosen = candidate
			found = true
			break
		}
	}
	if !found || !engine.applyAction(g, game.Player1, chosen) {
		t.Fatal("Gilraen activation failed")
	}
	if !gilraen.Tapped {
		t.Fatal("Gilraen was not tapped to pay its activation cost")
	}
	if !ownLand.Tapped || !secondLand.Tapped {
		t.Fatal("two lands were not tapped to pay {2}")
	}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = gilraenChoiceAgent{accept: true}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if permanentByCardID(g, ownCreature.CardInstanceID) == nil {
		t.Fatal("Gilraen's accepted immediate return did not return the target")
	}
}

func TestGilraenImmediateChoiceReturnsNewObjectWithoutCounters(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanent(g, game.Player1)
	target := addCombatPermanent(g, game.Player2, reanimationVanillaCreature("Borrowed Creature", 2, 2))
	target.Controller = game.Player1

	resolveGilraenBlink(g, engine, source, target, true, gilraenCounters())

	returned := permanentByCardID(g, target.CardInstanceID)
	if returned == nil || returned.ObjectID == target.ObjectID {
		t.Fatalf("returned permanent = %+v, want same card as a new object", returned)
	}
	if returned.Controller != game.Player2 {
		t.Fatalf("returned controller = %v, want owner Player2", returned.Controller)
	}
	if returned.Counters.Get(counter.Vigilance) != 0 || returned.Counters.Get(counter.Lifelink) != 0 {
		t.Fatalf("immediate return counters = %#v, want none", returned.Counters)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers = %d, want none after accepting immediate return", len(g.DelayedTriggers))
	}
}

func TestGilraenDeclineReturnsAtEndStepWithKeywordCounters(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanent(g, game.Player1)
	target := addCombatPermanent(g, game.Player2, reanimationVanillaCreature("Borrowed Creature", 2, 2))
	target.Controller = game.Player1

	resolveGilraenBlink(g, engine, source, target, false, gilraenCounters())
	if !g.Players[game.Player2].Exile.Contains(target.CardInstanceID) || len(g.DelayedTriggers) != 1 {
		t.Fatalf("after decline: exile=%v delayed=%d, want card exiled and one trigger", g.Players[game.Player2].Exile.Contains(target.CardInstanceID), len(g.DelayedTriggers))
	}
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	returned := permanentByCardID(g, target.CardInstanceID)
	if returned == nil || returned.ObjectID == target.ObjectID {
		t.Fatalf("returned permanent = %+v, want same card as a new object", returned)
	}
	if returned.Controller != game.Player2 {
		t.Fatalf("returned controller = %v, want owner Player2", returned.Controller)
	}
	if returned.Counters.Get(counter.Vigilance) != 1 || returned.Counters.Get(counter.Lifelink) != 1 {
		t.Fatalf("returned counters = %#v, want vigilance and lifelink", returned.Counters)
	}
	if !hasKeyword(g, returned, game.Vigilance) || !hasKeyword(g, returned, game.Lifelink) {
		t.Fatal("keyword counters did not grant vigilance and lifelink")
	}
}

func TestGilraenDelayedCaptureSurvivesSourceLeavingAndMultipleActivations(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanent(g, game.Player1)
	first := addCombatPermanent(g, game.Player1, reanimationVanillaCreature("First", 2, 2))
	second := addCombatPermanent(g, game.Player1, reanimationVanillaCreature("Second", 3, 3))

	resolveGilraenBlink(g, engine, source, first, false, gilraenCounters())
	resolveGilraenBlink(g, engine, source, second, false, gilraenCounters())
	if len(g.DelayedTriggers) != 2 {
		t.Fatalf("delayed triggers = %d, want one per activation", len(g.DelayedTriggers))
	}
	captured := map[id.ID]bool{}
	for _, delayed := range g.DelayedTriggers {
		captured[delayed.CapturedCardID] = true
	}
	if !captured[first.CardInstanceID] || !captured[second.CardInstanceID] {
		t.Fatalf("captured cards = %#v, want %v and %v", captured, first.CardInstanceID, second.CardInstanceID)
	}
	movePermanentToZone(g, source, zone.Graveyard)
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	for _, original := range []*game.Permanent{first, second} {
		returned := permanentByCardID(g, original.CardInstanceID)
		if returned == nil {
			t.Fatalf("card %v did not return after source left; exile=%v graveyard=%v battlefield=%#v", original.CardInstanceID, g.Players[original.Owner].Exile.Contains(original.CardInstanceID), g.Players[original.Owner].Graveyard.Contains(original.CardInstanceID), g.Battlefield)
		}
		if returned.Counters.Get(counter.Vigilance) != 1 || returned.Counters.Get(counter.Lifelink) != 1 {
			t.Fatalf("card %v counters = %#v, want vigilance and lifelink", original.CardInstanceID, returned.Counters)
		}
	}
}

func TestGilraenDelayedReturnRequiresCardToRemainInExile(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanent(g, game.Player1)
	target := addCombatPermanent(g, game.Player1, reanimationVanillaCreature("Wandering Creature", 2, 2))

	resolveGilraenBlink(g, engine, source, target, false, gilraenCounters())
	if !moveCardBetweenZones(g, game.Player1, target.CardInstanceID, zone.Exile, zone.Library) {
		t.Fatal("failed to move captured card out of exile")
	}
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if permanentByCardID(g, target.CardInstanceID) != nil {
		t.Fatal("captured card returned after leaving exile")
	}
	if !g.Players[game.Player1].Library.Contains(target.CardInstanceID) {
		t.Fatal("captured card did not remain in its new zone")
	}
}

func TestGilraenDelayedReturnDoesNotFollowCardBackIntoExile(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanent(g, game.Player1)
	target := addCombatPermanent(g, game.Player1, reanimationVanillaCreature("Wandering Creature", 2, 2))

	resolveGilraenBlink(g, engine, source, target, false, gilraenCounters())
	if !moveCardBetweenZones(g, game.Player1, target.CardInstanceID, zone.Exile, zone.Library) {
		t.Fatal("failed to move captured card out of exile")
	}
	if !moveCardBetweenZones(g, game.Player1, target.CardInstanceID, zone.Library, zone.Exile) {
		t.Fatal("failed to move captured card back into exile")
	}
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if permanentByCardID(g, target.CardInstanceID) != nil {
		t.Fatal("captured delayed effect followed a new incarnation back into exile")
	}
	if !g.Players[game.Player1].Exile.Contains(target.CardInstanceID) {
		t.Fatal("newly exiled incarnation did not remain in exile")
	}
}

func TestGilraenCommanderReplacementAndTokenDoNotReturn(t *testing.T) {
	t.Parallel()
	t.Run("commander", func(t *testing.T) {
		t.Parallel()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addCombatCreaturePermanent(g, game.Player1)
		commander := addCommanderPermanent(g, game.Player1)

		resolveGilraenBlink(g, engine, source, commander, false, gilraenCounters())
		engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

		if !g.Players[game.Player1].CommandZone.Contains(commander.CardInstanceID) {
			t.Fatal("commander did not remain in the command zone")
		}
		if permanentByCardID(g, commander.CardInstanceID) != nil {
			t.Fatal("commander redirected from exile was returned")
		}
	})
	t.Run("token", func(t *testing.T) {
		t.Parallel()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addCombatCreaturePermanent(g, game.Player1)
		token := addTokenCreaturePermanent(g, game.Player1, "Blink Token")

		resolveGilraenBlink(g, engine, source, token, false, gilraenCounters())
		engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

		if _, ok := permanentByObjectID(g, token.ObjectID); ok {
			t.Fatal("exiled token remained on the battlefield")
		}
		if len(g.DelayedTriggers) != 0 {
			t.Fatalf("delayed triggers = %d, want none for vanished token", len(g.DelayedTriggers))
		}
	})
}

func TestCapturedDelayedReturnReattachesAura(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Blink Source",
		Types: []types.Card{types.Artifact},
	}})
	host := addCombatCreaturePermanent(g, game.Player1)
	aura := addAuraPermanent(g, game.Player1)
	if !attachPermanent(g, aura, host) {
		t.Fatal("failed to establish initial Aura attachment")
	}

	resolveGilraenBlink(g, engine, source, aura, false, nil)
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = gilraenChoiceAgent{}
	engine.runEndingPhase(g, agents)

	returned := permanentByCardID(g, aura.CardInstanceID)
	if returned == nil || !returned.AttachedTo.Exists || returned.AttachedTo.Val != host.ObjectID {
		t.Fatalf("returned Aura = %+v, want attached to host %v", returned, host.ObjectID)
	}
}

func TestCapturedDelayedReturnHonorsExplicitController(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanent(g, game.Player1)
	target := addCombatPermanent(g, game.Player2, reanimationVanillaCreature("Stolen Creature", 2, 2))
	obj := linkedSourceObject(source)
	obj.Targets = []game.Target{game.PermanentTarget(target.ObjectID)}

	resolveInstruction(engine, g, obj, game.Exile{
		Object:         game.TargetPermanentReference(0),
		ExileLinkedKey: gilraenBlinkKey,
	}, nil)
	resolveInstruction(engine, g, obj, game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing:       game.DelayedAtBeginningOfNextEndStep,
		CapturedCard: opt.Val(game.LinkedObjectReference(string(gilraenBlinkKey))),
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.PutOnBattlefield{
			Source:            game.CardBattlefieldSource(game.CapturedCardReference()),
			Recipient:         opt.Val(game.ControllerReference()),
			LinkedReturnZones: []zone.Type{zone.Exile},
		}}}}.Ability(),
	}}, nil)
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	returned := permanentByCardID(g, target.CardInstanceID)
	if returned == nil || returned.Controller != game.Player1 || returned.Owner != game.Player2 {
		t.Fatalf("returned permanent = %+v, want Player1 controller and Player2 owner", returned)
	}
}

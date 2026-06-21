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

func TestDiscardEffectDiscardsDeterministicHandCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	bottom := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Bottom"}})
	top := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	addEffectSpellToStack(g, game.Player1, game.Discard{Amount: game.Fixed(1), Player: game.TargetPlayerReference(0)}, []game.Target{game.PlayerTarget(game.Player2)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player2].Hand.Contains(top) || !g.Players[game.Player2].Graveyard.Contains(top) {
		t.Fatal("discard effect did not discard deterministic top hand card")
	}
	if !g.Players[game.Player2].Hand.Contains(bottom) {
		t.Fatal("discard effect discarded more cards than requested")
	}
}

func TestSearchRevealAndInvestigateKeywordActions(t *testing.T) {
	t.Run("search library to hand with reveal", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		creature := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Creature", Types: []types.Card{types.Creature}}})
		_ = addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Instant", Types: []types.Card{types.Instant}}})
		addEffectSpellToStack(g, game.Player1, game.Search{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
			Spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
				CardType:    opt.Val(types.Creature),
				Reveal:      true,
			},
		}, nil)

		engine.resolveTopOfStack(g, &TurnLog{})

		if !g.Players[game.Player1].Hand.Contains(creature) || g.Players[game.Player1].Library.Contains(creature) {
			t.Fatal("search effect did not move matching card library -> hand")
		}
		if !hasEvent(g, game.EventCardRevealed) {
			t.Fatalf("events = %+v, want reveal event for searched card", g.Events)
		}
	})

	t.Run("search can require a basic land", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		basic := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
			Supertypes: []types.Super{types.Basic},
			Types:      []types.Card{types.Land}},
		})
		nonbasic := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Nonbasic Land",
			Types: []types.Card{types.Land}},
		})
		addEffectSpellToStack(g, game.Player1, game.Search{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
			Spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
				CardType:    opt.Val(types.Land),
				Supertype:   opt.Val(types.Basic),
				Reveal:      true,
			},
		}, nil)

		engine.resolveTopOfStack(g, &TurnLog{})

		if !g.Players[game.Player1].Hand.Contains(basic) || g.Players[game.Player1].Library.Contains(basic) {
			t.Fatal("search effect did not move matching basic land library -> hand")
		}
		if !g.Players[game.Player1].Library.Contains(nonbasic) {
			t.Fatal("search effect moved nonbasic land despite basic filter")
		}
		if !hasEvent(g, game.EventCardRevealed) {
			t.Fatalf("events = %+v, want reveal event for searched basic land", g.Events)
		}
	})

	t.Run("search without supertype filter still matches nonbasic lands", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		nonbasic := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Nonbasic Land",
			Types: []types.Card{types.Land}},
		})
		addEffectSpellToStack(g, game.Player1, game.Search{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
			Spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
				CardType:    opt.Val(types.Land),
			},
		}, nil)

		engine.resolveTopOfStack(g, &TurnLog{})

		if !g.Players[game.Player1].Hand.Contains(nonbasic) || g.Players[game.Player1].Library.Contains(nonbasic) {
			t.Fatal("search effect did not move nonbasic land without a supertype filter")
		}
	})

	t.Run("search can put subtype-matching land onto battlefield tapped", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		forest := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Forest}},
		})
		_ = addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Wastes",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Desert}},
		})
		addEffectSpellToStack(g, game.Player1, game.Search{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
			Spec: game.SearchSpec{
				SourceZone:   zone.Library,
				Destination:  zone.Battlefield,
				CardType:     opt.Val(types.Land),
				SubtypesAny:  []types.Sub{types.Forest},
				EntersTapped: true,
			},
		}, nil)

		engine.resolveTopOfStack(g, &TurnLog{})

		if g.Players[game.Player1].Library.Contains(forest) {
			t.Fatal("search effect left matching land in library")
		}
		permanent := permanentByCardID(g, forest)
		if permanent == nil {
			t.Fatal("search effect did not put matching land onto battlefield")
		}
		if !permanent.Tapped {
			t.Fatal("searched land entered untapped, want tapped")
		}
	})

	t.Run("reveal top library card", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		cardID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
		addEffectSpellToStack(g, game.Player1, game.Reveal{Amount: game.Fixed(1), Player: game.ControllerReference()}, nil)

		engine.resolveTopOfStack(g, &TurnLog{})

		if !g.Players[game.Player1].Library.Contains(cardID) {
			t.Fatal("reveal effect moved the card")
		}
		if !hasEvent(g, game.EventCardRevealed) {
			t.Fatalf("events = %+v, want reveal event", g.Events)
		}
	})

	t.Run("investigate creates clue token", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
		addEffectSpellToStack(g, game.Player1, game.Investigate{Amount: game.Fixed(2)}, nil)

		engine.resolveTopOfStack(g, &TurnLog{})

		if len(g.Battlefield) != 2 {
			t.Fatalf("battlefield size = %d, want 2 clues", len(g.Battlefield))
		}
		clue := g.Battlefield[0]
		if !clue.Token || clue.TokenDef == nil || clue.TokenDef.Name != "Clue Token" || !clue.TokenDef.HasSubtype(types.Clue) {
			t.Fatalf("clue token = %+v def=%+v", clue, clue.TokenDef)
		}
		if clue.TokenDef.AbilityCount() != 1 || len(clue.TokenDef.ActivatedAbilities) != 1 {
			t.Fatalf("clue abilities = count=%d activated=%d, want activated draw ability", clue.TokenDef.AbilityCount(), len(clue.TokenDef.ActivatedAbilities))
		}
		g.Players[game.Player1].ManaPool.Add(mana.C, 2)
		if !engine.applyAction(g, game.Player1, actionBuild.activateAbility(clue.ObjectID, 0, nil, 0)) {
			t.Fatal("clue activation failed")
		}
		if _, ok := permanentByObjectID(g, clue.ObjectID); ok {
			t.Fatal("clue activation did not sacrifice its source")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.Players[game.Player1].Hand.Contains(drawn) {
			t.Fatal("clue activation did not draw a card")
		}
	})
}

func TestStartEnginesAndSpeedIncreasesOncePerTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.TurnNumber = 1

	if !startEngines(g, game.Player1) {
		t.Fatal("startEngines failed")
	}
	if got := g.Players[game.Player1].Speed; got != 1 {
		t.Fatalf("speed = %d, want 1", got)
	}
	loseLife(g, game.Player2, 1)
	loseLife(g, game.Player3, 1)
	if got := g.Players[game.Player1].Speed; got != 2 {
		t.Fatalf("speed = %d, want one increase to 2 this turn", got)
	}
	g.Turn.TurnNumber = 2
	loseLife(g, game.Player2, 1)
	if got := g.Players[game.Player1].Speed; got != 3 {
		t.Fatalf("speed = %d, want 3 after next-turn opponent life loss", got)
	}
}

func TestMonstrosityEffectAddsCountersOnlyOnce(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Monster",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})},
	})
	obj := &game.StackObject{
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}

	resolveInstruction(engine, g, obj, game.Monstrosity{Amount: game.Fixed(5), Object: game.SourcePermanentReference()}, &TurnLog{})
	resolveInstruction(engine, g, obj, game.Monstrosity{Amount: game.Fixed(5), Object: game.SourcePermanentReference()}, &TurnLog{})

	if !source.Monstrous {
		t.Fatal("source did not become monstrous")
	}
	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 5 {
		t.Fatalf("+1/+1 counters = %d, want 5 after repeated monstrosity resolutions", got)
	}
}

func TestSetClassLevelEffectAndClassInitialLevel(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Class",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Class}},
	})
	card := g.CardInstances[cardID]
	source, ok := createCardPermanent(g, card, game.Player1, zone.Hand)
	if !ok {
		t.Fatal("createCardPermanent failed")
	}
	if got := source.ClassLevel; got != 1 {
		t.Fatalf("initial class level = %d, want 1", got)
	}
	obj := &game.StackObject{
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}

	resolveInstruction(engine, g, obj, game.SetClassLevel{Amount: game.Fixed(2), Object: game.SourcePermanentReference()}, &TurnLog{})
	resolveInstruction(engine, g, obj, game.SetClassLevel{Amount: game.Fixed(1), Object: game.SourcePermanentReference()}, &TurnLog{})

	if got := source.ClassLevel; got != 2 {
		t.Fatalf("class level = %d, want upgraded and not downgraded level 2", got)
	}
}

func TestRuleEffectCantBeBlockedBindsAffectedObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	otherAttacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind:             game.RuleEffectCantBeBlocked,
		AffectedObjectID: attacker.ObjectID,
	})

	if canBlockAttacker(g, blocker, attacker) {
		t.Fatal("blocker could block creature affected by can't-be-blocked rule effect")
	}
	if !canBlockAttacker(g, blocker, otherAttacker) {
		t.Fatal("can't-be-blocked rule effect affected the wrong attacker")
	}
}

func TestProliferateAddsOneChosenCounterKind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	permanent.Counters.Add(counter.PlusOnePlusOne, 1)
	permanent.Counters.Add(counter.Charge, 1)
	g.Players[game.Player2].PoisonCounters = 1
	addEffectSpellToStack(g, game.Player1, game.Proliferate{}, nil)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}, {0}}},
	}
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("+1/+1 counters = %d, want chosen counter incremented", got)
	}
	if got := permanent.Counters.Get(counter.Charge); got != 1 {
		t.Fatalf("charge counters = %d, want unchosen counter unchanged", got)
	}
	if got := g.Players[game.Player2].PoisonCounters; got != 2 {
		t.Fatalf("poison counters = %d, want proliferated player counter", got)
	}
	var playerCounterEvent game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventCountersAdded && event.Player == game.Player2 {
			playerCounterEvent = event
		}
	}
	if playerCounterEvent.CounterKind != counter.Poison ||
		playerCounterEvent.PreviousCounterAmount != 1 ||
		playerCounterEvent.Amount != 1 {
		t.Fatalf("player counter event = %+v", playerCounterEvent)
	}
	if len(log.Choices) != 2 || log.Choices[0].Request.Kind != game.ChoiceProliferate {
		t.Fatalf("choices = %+v, want proliferate choices", log.Choices)
	}
}

func TestProliferateTwiceRepeatsAction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	permanent.Counters.Add(counter.PlusOnePlusOne, 1)
	addEffectSpellToStack(g, game.Player1, game.Proliferate{Amount: game.Fixed(2)}, nil)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}, {0}}},
	}
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("+1/+1 counters = %d, want proliferate twice", got)
	}
	if len(log.Choices) != 2 {
		t.Fatalf("choices = %d, want one proliferate choice per repetition", len(log.Choices))
	}
}

func TestGoadEffectExpiresOnGoadingPlayersNextTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addEffectSpellToStack(g, game.Player1, game.Goad{Object: game.TargetPermanentReference(0)}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if !wasGoadedBy(target, game.Player1) {
		t.Fatal("target was not goaded")
	}
	g.Turn.TurnNumber = 5
	g.Turn.ActivePlayer = game.Player1
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if wasGoadedBy(target, game.Player1) {
		t.Fatal("goad did not expire on goading player's next turn")
	}
}

func TestScryAndSurveilUseChoiceAgent(t *testing.T) {
	t.Run("scry bottom", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		bottom := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bottom"}})
		top := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
		addEffectSpellToStack(g, game.Player1, game.Scry{Amount: game.Fixed(1), Player: game.ControllerReference()}, nil)
		log := TurnLog{}
		agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}

		engine.resolveTopOfStackWithChoices(g, agents, &log)

		if got := g.Players[game.Player1].Library.All(); len(got) != 2 || got[0] != bottom || got[1] != top {
			t.Fatalf("library after scry = %+v, want chosen card on bottom", got)
		}
		if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceScry || log.Choices[0].UsedFallback {
			t.Fatalf("choices = %+v, want non-fallback scry choice", log.Choices)
		}
		assertEvent(t, g.Events, game.EventScry, func(event game.Event) bool {
			return event.Player == game.Player1 && event.Amount == 1
		})
	})
	t.Run("surveil graveyard", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		top := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
		addEffectSpellToStack(g, game.Player1, game.Surveil{Amount: game.Fixed(1), Player: game.ControllerReference()}, nil)
		log := TurnLog{}
		agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}

		engine.resolveTopOfStackWithChoices(g, agents, &log)

		if g.Players[game.Player1].Library.Contains(top) || !g.Players[game.Player1].Graveyard.Contains(top) {
			t.Fatal("surveil choice did not move card to graveyard")
		}
		if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceSurveil || log.Choices[0].UsedFallback {
			t.Fatalf("choices = %+v, want non-fallback surveil choice", log.Choices)
		}
		assertEvent(t, g.Events, game.EventSurveil, func(event game.Event) bool {
			return event.Player == game.Player1 && event.Amount == 1
		})
	})
}

func TestExploreMovesRevealedLandToHand(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	land := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Forest",
		Types: []types.Card{types.Land},
	}})
	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}

	resolveInstruction(engine, g, obj, game.Explore{Creature: game.SourcePermanentReference()}, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(land) || g.Players[game.Player1].Library.Contains(land) {
		t.Fatal("explore did not move revealed land to hand")
	}
	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("+1/+1 counters = %d, want none for revealed land", got)
	}
	assertEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
		return event.CardID == land && event.FromZone == zone.Library
	})
}

func TestExplorePutsCounterAndMayMoveNonlandToGraveyard(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	nonland := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bear Cub",
		Types: []types.Card{types.Creature},
	}})
	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{
		Primitive: game.Explore{Creature: game.SourcePermanentReference()},
	}, agents, &TurnLog{})

	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters = %d, want one for revealed nonland", got)
	}
	if g.Players[game.Player1].Library.Contains(nonland) || !g.Players[game.Player1].Graveyard.Contains(nonland) {
		t.Fatal("explore choice did not move nonland to graveyard")
	}
	assertEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
		return event.CardID == nonland && event.FromZone == zone.Library
	})
}

func TestExploreAllowsNoncreaturePermanent(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Map",
		Types: []types.Card{types.Artifact},
	}})
	nonland := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Spell",
		Types: []types.Card{types.Instant},
	}})
	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}

	resolveInstruction(engine, g, obj, game.Explore{Creature: game.SourcePermanentReference()}, &TurnLog{})

	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters = %d, want noncreature permanent to explore", got)
	}
	if !g.Players[game.Player1].Library.Contains(nonland) {
		t.Fatal("default explore choice should leave nonland on top")
	}
}

func TestManifestForReferencedTargetControllerUsesThatPlayersLibrary(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	cardID := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Hidden Bear",
		Types:     []types.Card{types.Creature},
		ManaCost:  opt.Val(cost.Mana{cost.G}),
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	}})
	addEffectSpellToStack(g, game.Player1, game.Manifest{
		Player: game.ObjectControllerReference(game.TargetPermanentReference(0)),
	}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player2].Library.Contains(cardID) {
		t.Fatal("referenced controller's library card was not manifested")
	}
	var manifested *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			manifested = permanent
			break
		}
	}
	if manifested == nil {
		t.Fatal("manifested permanent not found on battlefield")
	}
	if manifested.Controller != game.Player2 {
		t.Fatalf("manifested controller = %v, want Player2 (the target's controller)", manifested.Controller)
	}
	if !manifested.FaceDown || manifested.FaceDownKind != game.FaceDownManifest {
		t.Fatalf("manifest face-down state = %+v", manifested)
	}
}

func TestManifestPutsTopLibraryCardOntoBattlefieldFaceDown(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Hidden Bear",
		Types:     []types.Card{types.Creature},
		ManaCost:  opt.Val(cost.Mana{cost.G}),
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	}})
	addEffectSpellToStack(g, game.Player1, game.Manifest{}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Library.Contains(cardID) {
		t.Fatal("manifested card remained in library")
	}
	var manifested *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			manifested = permanent
			break
		}
	}
	if manifested == nil {
		t.Fatal("manifested permanent not found on battlefield")
	}
	if !manifested.FaceDown || manifested.FaceDownKind != game.FaceDownManifest || manifested.FaceDownFace != game.FaceFront {
		t.Fatalf("manifest face-down state = %+v", manifested)
	}
	if !permanentHasType(g, manifested, types.Creature) || effectivePower(g, manifested) != 2 {
		t.Fatalf("manifest effective characteristics typeCreature=%t power=%d", permanentHasType(g, manifested, types.Creature), effectivePower(g, manifested))
	}
	if permanentEffectiveName(g, manifested) != "" || len(permanentEffectiveAbilities(g, manifested)) != 0 {
		t.Fatalf("manifest visible name/abilities = %q/%d, want hidden/no abilities", permanentEffectiveName(g, manifested), len(permanentEffectiveAbilities(g, manifested)))
	}
	assertEvent(t, g.Events, game.EventPermanentEnteredBattlefield, func(event game.Event) bool {
		return event.CardID == cardID &&
			event.PermanentID == manifested.ObjectID &&
			event.FromZone == zone.Library &&
			event.ToZone == zone.Battlefield
	})
}

func TestManifestAppliesGlobalETBReplacementEffects(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID:           g.IDGen.Next(),
		Controller:   game.Player1,
		Description:  "Creatures enter tapped with a +1/+1 counter.",
		MatchEvent:   game.EventPermanentEnteredBattlefield,
		MatchToZone:  true,
		ToZone:       zone.Battlefield,
		EntersTapped: true,
		EntersWithCounters: []game.CounterPlacement{{
			Kind:   counter.PlusOnePlusOne,
			Amount: 1,
		}},
	})
	cardID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Hidden Spell",
		Types: []types.Card{types.Instant},
	}})
	addEffectSpellToStack(g, game.Player1, game.Manifest{}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	var manifested *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			manifested = permanent
			break
		}
	}
	if manifested == nil {
		t.Fatal("manifested permanent not found on battlefield")
	}
	if !manifested.Tapped {
		t.Fatal("global ETB replacement did not tap manifested permanent")
	}
	if got := manifested.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters = %d, want global ETB replacement counter", got)
	}
}

func TestManifestDreadChoosesOneCardAndPutsOtherIntoGraveyard(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	bottomOfTwo := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Card",
		Types: []types.Card{types.Instant},
	}})
	topOfTwo := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Manifested Card",
		Types:     []types.Card{types.Creature},
		ManaCost:  opt.Val(cost.Mana{cost.G}),
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	}})
	addEffectSpellToStack(g, game.Player1, game.Manifest{Dread: true}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if !g.Players[game.Player1].Graveyard.Contains(topOfTwo) {
		t.Fatal("unchosen top card was not put into graveyard")
	}
	var manifested *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == bottomOfTwo {
			manifested = permanent
			break
		}
	}
	if manifested == nil {
		t.Fatal("chosen manifest dread card not found on battlefield")
	}
	if !manifested.FaceDown || manifested.FaceDownKind != game.FaceDownManifest {
		t.Fatalf("manifest dread face-down state = %+v", manifested)
	}
	if g.Players[game.Player1].Library.Contains(topOfTwo) || g.Players[game.Player1].Library.Contains(bottomOfTwo) {
		t.Fatal("manifest dread cards remained in library")
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceManifest || log.Choices[0].Selected[0] != 1 {
		t.Fatalf("manifest dread choice log = %+v, want selected card index 1", log.Choices)
	}
}

func TestManifestDreadWithOneCardManifestsWithoutChoice(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Only Card",
		Types: []types.Card{types.Instant},
	}})
	addEffectSpellToStack(g, game.Player1, game.Manifest{Dread: true}, nil)
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	if len(log.Choices) != 0 {
		t.Fatalf("choices = %+v, want no choice for one-card manifest dread", log.Choices)
	}
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID && permanent.FaceDown && permanent.FaceDownKind == game.FaceDownManifest {
			return
		}
	}
	t.Fatal("single-card manifest dread did not manifest the only card")
}

func TestEvolveAddsCounterWhenGreaterCreatureEnters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	evolver := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Evolving Creature",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 1}),
		Toughness:       opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{game.EvolveStaticBody},
	}})
	entering := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		PermanentID: entering.ObjectID,
	})

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("evolve trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := evolver.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("evolve counters = %d, want 1 after greater creature entered", got)
	}
}

func TestEvolveDoesNotTriggerWhenSmallerCreatureEnters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	evolver := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Evolving Creature",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 2}),
		Toughness:       opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{game.EvolveStaticBody},
	}})
	entering := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		PermanentID: entering.ObjectID,
	})

	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("evolve trigger fired for an equal-stats creature")
	}
	if got := evolver.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("evolve counters = %d, want 0 for equal-stats creature", got)
	}
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestChosenCreatureTypeRuleEffectAddsOneQualifyingTrigger(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	doubler := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Trigger Doubler",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Golem},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind: game.RuleEffectAdditionalTriggerForChosenCreatureType,
			}},
		}},
	}})
	doubler.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryTypeChoiceKey: {Kind: game.ResolutionChoiceSubtype, Subtype: types.Elf},
	}
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventLifeGained,
		Player: game.TriggerPlayerYou,
	}, nil, nil)
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("trigger source card not found")
	}
	card.Def.Subtypes = []types.Sub{types.Elf}

	emitEvent(g, game.Event{Kind: game.EventLifeGained, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("triggered ability was not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want original plus one additional trigger", got)
	}
}

func TestChosenCreatureTypeRuleEffectRespectsMaxTriggersPerTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	doubler := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Trigger Doubler",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind: game.RuleEffectAdditionalTriggerForChosenCreatureType,
			}},
		}},
	}})
	doubler.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryTypeChoiceKey: {Kind: game.ResolutionChoiceSubtype, Subtype: types.Elf},
	}
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventLifeGained,
		Player: game.TriggerPlayerYou,
	}, nil, nil)
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("trigger source card not found")
	}
	card.Def.Subtypes = []types.Sub{types.Elf}
	card.Def.TriggeredAbilities[0].MaxTriggersPerTurn = 1

	emitEvent(g, game.Event{Kind: game.EventLifeGained, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("triggered ability was not put on the stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want max-once limit to suppress the additional trigger", got)
	}
}

func TestChosenCreatureTypeRuleEffectIsCapturedAtDrawEventTime(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	doubler := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Trigger Doubler",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind: game.RuleEffectAdditionalTriggerForChosenCreatureType,
			}},
		}},
	}})
	doubler.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryTypeChoiceKey: {Kind: game.ResolutionChoiceSubtype, Subtype: types.Elf},
	}
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDrawn,
		Player: game.TriggerPlayerYou,
	}, nil, nil)
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("trigger source card not found")
	}
	card.Def.Subtypes = []types.Sub{types.Elf}

	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player1})
	source.Controller = game.Player2
	doubler.EntryChoices[game.EntryTypeChoiceKey] = game.ResolutionChoiceResult{
		Kind:    game.ResolutionChoiceSubtype,
		Subtype: types.Goblin,
	}
	for i, permanent := range g.Battlefield {
		if permanent.ObjectID == doubler.ObjectID {
			g.Battlefield = append(g.Battlefield[:i], g.Battlefield[i+1:]...)
			break
		}
	}

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("captured draw trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want event-time doubler to preserve the additional trigger", got)
	}
	for _, object := range g.Stack.Objects() {
		if object.Controller != game.Player1 || object.SourceID != source.ObjectID {
			t.Fatalf("trigger identity = controller %v source %v, want event-time Player1 source %v", object.Controller, object.SourceID, source.ObjectID)
		}
	}
}

func TestChosenCreatureTypeRuleEffectUsesTriggerSourceLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	doubler := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Trigger Doubler",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind: game.RuleEffectAdditionalTriggerForChosenCreatureType,
			}},
		}},
	}})
	doubler.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryTypeChoiceKey: {Kind: game.ResolutionChoiceSubtype, Subtype: types.Elf},
	}
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventPermanentDied,
		Source: game.TriggerSourceSelf,
	}, nil, nil)
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("trigger source card not found")
	}
	card.Def.Subtypes = []types.Sub{types.Elf}

	if _, ok := destroyPermanent(g, source.ObjectID); !ok {
		t.Fatal("destroyPermanent failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("dies trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want source LKI to qualify the additional trigger", got)
	}
}

func TestChosenCreatureTypeRuleEffectUsesSimultaneouslyDepartedDoublerLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	doubler := addChosenCreatureTypeDoubler(g, game.Player1, types.Elf)
	source := addChosenTypeTriggerSource(g, game.Player1, types.Elf, &game.TriggerPattern{
		Event:  game.EventPermanentDied,
		Source: game.TriggerSourceSelf,
	}, nil)
	batch := g.IDGen.Next()

	if _, ok := destroyPermanentInBatch(g, source.ObjectID, batch, false); !ok {
		t.Fatal("destroying trigger source failed")
	}
	if _, ok := destroyPermanentInBatch(g, doubler.ObjectID, batch, false); !ok {
		t.Fatal("destroying doubler failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("dies trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want simultaneously departed doubler to add a trigger", got)
	}
}

func TestChosenCreatureTypeRuleEffectAddsOncePerDoubler(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addChosenCreatureTypeDoubler(g, game.Player1, types.Elf)
	addChosenCreatureTypeDoubler(g, game.Player1, types.Elf)
	addChosenTypeTriggerSource(g, game.Player1, types.Elf, &game.TriggerPattern{
		Event:  game.EventLifeGained,
		Player: game.TriggerPlayerYou,
	}, nil)

	emitEvent(g, game.Event{Kind: game.EventLifeGained, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("triggered ability was not put on the stack")
	}
	if got := g.Stack.Size(); got != 3 {
		t.Fatalf("stack size = %d, want original plus one occurrence per doubler", got)
	}
}

func TestChosenCreatureTypeRuleEffectExclusions(t *testing.T) {
	tests := map[string]struct {
		sourceController game.PlayerID
		sourceTypes      []types.Card
		sourceSubtype    types.Sub
	}{
		"wrong subtype": {
			sourceController: game.Player1,
			sourceTypes:      []types.Card{types.Creature},
			sourceSubtype:    types.Goblin,
		},
		"not a creature": {
			sourceController: game.Player1,
			sourceTypes:      []types.Card{types.Enchantment},
			sourceSubtype:    types.Elf,
		},
		"not controlled by doubler controller": {
			sourceController: game.Player2,
			sourceTypes:      []types.Card{types.Creature},
			sourceSubtype:    types.Elf,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addChosenCreatureTypeDoubler(g, game.Player1, types.Elf)
			source := addChosenTypeTriggerSource(g, tc.sourceController, tc.sourceSubtype, &game.TriggerPattern{
				Event: game.EventLifeGained,
			}, nil)
			card, ok := g.GetCardInstance(source.CardInstanceID)
			if !ok {
				t.Fatal("trigger source card not found")
			}
			card.Def.Types = tc.sourceTypes

			emitEvent(g, game.Event{Kind: game.EventLifeGained, Player: game.Player3})
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("original triggered ability was not put on the stack")
			}
			if got := g.Stack.Size(); got != 1 {
				t.Fatalf("stack size = %d, want only the original trigger", got)
			}
		})
	}
}

func TestChosenCreatureTypeRuleEffectDoesNotDoubleItsOwnTrigger(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := chosenCreatureTypeDoublerDef()
	def.TriggeredAbilities = []game.TriggeredAbility{{
		Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
			Event:  game.EventLifeGained,
			Player: game.TriggerPlayerYou,
		}},
	}}
	source := addCombatPermanent(g, game.Player1, def)
	source.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryTypeChoiceKey: {Kind: game.ResolutionChoiceSubtype, Subtype: types.Elf},
	}

	emitEvent(g, game.Event{Kind: game.EventLifeGained, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("original triggered ability was not put on the stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want the doubler's own trigger excluded", got)
	}
}

func TestChosenCreatureTypeRuleEffectCoalescesBeforeDoubling(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addChosenCreatureTypeDoubler(g, game.Player1, types.Elf)
	addChosenTypeTriggerSource(g, game.Player1, types.Elf, &game.TriggerPattern{
		Event:                 game.EventPermanentDied,
		Controller:            game.TriggerControllerYou,
		RequirePermanentTypes: []types.Card{types.Creature},
		OneOrMore:             true,
	}, nil)
	first := addCombatCreaturePermanent(g, game.Player1)
	second := addCombatCreaturePermanent(g, game.Player1)
	batch := g.IDGen.Next()
	emitEvent(g, game.Event{Kind: game.EventPermanentDied, Controller: game.Player1, PermanentID: first.ObjectID, CardID: first.CardInstanceID, SimultaneousID: batch})
	emitEvent(g, game.Event{Kind: game.EventPermanentDied, Controller: game.Player1, PermanentID: second.ObjectID, CardID: second.CardInstanceID, SimultaneousID: batch})

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("one-or-more trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want one coalesced original doubled once", got)
	}
}

func TestChosenCreatureTypeRuleEffectPreservesAPNAPOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	addChosenCreatureTypeDoubler(g, game.Player1, types.Elf)
	addChosenTypeTriggerSource(g, game.Player1, types.Elf, &game.TriggerPattern{Event: game.EventLifeGained}, nil)
	addChosenTypeTriggerSource(g, game.Player2, types.Goblin, &game.TriggerPattern{Event: game.EventLifeGained}, nil)

	emitEvent(g, game.Event{Kind: game.EventLifeGained, Player: game.Player3})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("triggered abilities were not put on the stack")
	}
	objects := g.Stack.Objects()
	if len(objects) != 3 {
		t.Fatalf("stack size = %d, want three", len(objects))
	}
	for i := range 2 {
		if objects[i].Controller != game.Player1 {
			t.Fatalf("stack[%d] controller = %v, want active Player1", i, objects[i].Controller)
		}
	}
	if objects[2].Controller != game.Player2 {
		t.Fatalf("top controller = %v, want nonactive Player2", objects[2].Controller)
	}
}

func TestChosenCreatureTypeRuleEffectPreparesTargetsIndependently(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addChosenCreatureTypeDoubler(g, game.Player1, types.Elf)
	firstTarget := addCombatCreaturePermanent(g, game.Player2)
	secondTarget := addCombatCreaturePermanent(g, game.Player3)
	addChosenTypeTriggerSource(g, game.Player1, types.Elf, &game.TriggerPattern{
		Event:  game.EventLifeGained,
		Player: game.TriggerPlayerYou,
	}, []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}})
	agent := &choiceOnlyAgent{choices: [][]int{{0, 1}, {1}, {2}}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent

	emitEvent(g, game.Event{Kind: game.EventLifeGained, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, nil) {
		t.Fatal("triggered abilities were not put on the stack")
	}
	objects := g.Stack.Objects()
	if len(objects) != 2 || len(objects[0].Targets) != 1 || len(objects[1].Targets) != 1 {
		t.Fatalf("stack objects = %#v, want two independently targeted triggers", objects)
	}
	got := map[game.ObjectID]bool{
		objects[0].Targets[0].PermanentID: true,
		objects[1].Targets[0].PermanentID: true,
	}
	if !got[firstTarget.ObjectID] || !got[secondTarget.ObjectID] {
		t.Fatalf("trigger targets = %#v, want %v and %v", got, firstTarget.ObjectID, secondTarget.ObjectID)
	}
}

func TestChosenCreatureTypeRuleEffectCapturesDoublerStateAtEventTime(t *testing.T) {
	mutations := map[string]func(g *game.Game, doubler *game.Permanent){
		"doubler leaves the battlefield": func(g *game.Game, doubler *game.Permanent) {
			removeChosenTypeDoublerFromBattlefield(g, doubler.ObjectID)
		},
		"doubler changes controller": func(_ *game.Game, doubler *game.Permanent) {
			doubler.Controller = game.Player2
		},
		"doubler changes chosen type": func(_ *game.Game, doubler *game.Permanent) {
			doubler.EntryChoices[game.EntryTypeChoiceKey] = game.ResolutionChoiceResult{
				Kind:    game.ResolutionChoiceSubtype,
				Subtype: types.Goblin,
			}
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			doubler := addChosenCreatureTypeDoubler(g, game.Player1, types.Elf)
			addChosenTypeTriggerSource(g, game.Player1, types.Elf, &game.TriggerPattern{
				Event:  game.EventLifeGained,
				Player: game.TriggerPlayerYou,
			}, nil)

			// The multiplier count must be captured when the ordinary trigger's
			// event is emitted. Each later change would make a resolution-time
			// recomputation drop the doubling, so the additional trigger proves
			// the event-time doubler state was used.
			emitEvent(g, game.Event{Kind: game.EventLifeGained, Player: game.Player1})
			mutate(g, doubler)

			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("triggered ability was not put on the stack")
			}
			if got := g.Stack.Size(); got != 2 {
				t.Fatalf("stack size = %d, want event-time doubler to preserve the additional trigger", got)
			}
		})
	}
}

func TestChosenCreatureTypeRuleEffectExcludesNonOrdinaryTriggers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addChosenCreatureTypeDoubler(g, game.Player1, types.Elf)
	source := addChosenTypeTriggerSource(g, game.Player1, types.Elf, &game.TriggerPattern{
		Event: game.EventLifeGained,
	}, nil)
	event := game.Event{
		Kind:   game.EventLifeGained,
		Player: game.Player1,
		ChosenTypeTriggerDoublers: &game.ChosenTypeTriggerDoublerSnapshot{
			Doublers: captureChosenTypeTriggerDoublers(g),
		},
	}

	ordinary := pendingTriggeredAbility{controller: game.Player1, sourceID: source.ObjectID, event: event, hasEvent: true, ordinaryTrigger: true}
	if got := capturedChosenCreatureTypeAdditionalTriggerCount(g, &ordinary); got != 1 {
		t.Fatalf("ordinary additional triggers = %d, want the qualifying source doubled once", got)
	}

	chapter := pendingTriggeredAbility{controller: game.Player1, sourceID: source.ObjectID, event: event, hasEvent: true, sagaChapter: true}
	madness := pendingTriggeredAbility{controller: game.Player1, sourceID: source.ObjectID, event: event, hasEvent: true}
	stateOrDelayed := pendingTriggeredAbility{controller: game.Player1, sourceID: source.ObjectID}
	excluded := map[string]pendingTriggeredAbility{
		"saga chapter":                      chapter,
		"madness or other synthetic event":  madness,
		"state or delayed without an event": stateOrDelayed,
	}
	for name, trigger := range excluded {
		t.Run(name, func(t *testing.T) {
			candidate := trigger
			if got := capturedChosenCreatureTypeAdditionalTriggerCount(g, &candidate); got != 0 {
				t.Fatalf("additional triggers = %d, want non-ordinary trigger left undoubled", got)
			}
		})
	}
}

func TestChosenCreatureTypeRuleEffectExcludesSagaChapterTriggers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addChosenCreatureTypeDoubler(g, game.Player1, types.Elf)
	saga := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:             "Creature Saga",
		Types:            []types.Card{types.Enchantment, types.Creature},
		Subtypes:         []types.Sub{types.Saga, types.Elf},
		ChapterAbilities: []game.ChapterAbility{sagaDrawChapter(1)},
	}})

	if !addCountersToPermanent(g, saga, counter.Lore, 1) {
		t.Fatal("adding a lore counter failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("saga chapter was not put on the stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want the saga chapter left undoubled even on a chosen-type creature", got)
	}
}

func removeChosenTypeDoublerFromBattlefield(g *game.Game, objectID game.ObjectID) {
	for i, permanent := range g.Battlefield {
		if permanent.ObjectID == objectID {
			g.Battlefield = append(g.Battlefield[:i], g.Battlefield[i+1:]...)
			return
		}
	}
}

func addChosenCreatureTypeDoubler(g *game.Game, controller game.PlayerID, subtype types.Sub) *game.Permanent {
	permanent := addCombatPermanent(g, controller, chosenCreatureTypeDoublerDef())
	permanent.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryTypeChoiceKey: {Kind: game.ResolutionChoiceSubtype, Subtype: subtype},
	}
	return permanent
}

func chosenCreatureTypeDoublerDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Trigger Doubler",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Golem},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind: game.RuleEffectAdditionalTriggerForChosenCreatureType,
			}},
		}},
	}}
}

func addChosenTypeTriggerSource(
	g *game.Game,
	controller game.PlayerID,
	subtype types.Sub,
	pattern *game.TriggerPattern,
	targets []game.TargetSpec,
) *game.Permanent {
	source := addTriggeredPermanent(g, controller, pattern, nil, targets)
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		panic("trigger source card not found")
	}
	card.Def.Subtypes = []types.Sub{subtype}
	return source
}

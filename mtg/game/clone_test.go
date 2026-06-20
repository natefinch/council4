package game

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
)

// buildRichGame constructs a non-trivial game with permanents on the
// battlefield, cards in zones, an object on the stack, continuous effects,
// counters, combat, and various tracking maps populated.
func buildRichGame(t *testing.T) *Game {
	t.Helper()

	def := &CardDef{CardFace: CardFace{Name: "Test Creature"}}
	var configs [NumPlayers]PlayerConfig
	for p := range configs {
		configs[p].Name = "Player"
		configs[p].Commander = &CardDef{CardFace: CardFace{Name: "Commander"}}
		for range 6 {
			configs[p].Deck = append(configs[p].Deck, def)
		}
	}
	g := NewGameWithRand(configs, rand.New(rand.NewPCG(1, 2)))

	// Draw a couple of cards so hands are non-empty.
	for range 2 {
		top, ok := g.Players[Player1].Library.Top()
		if !ok {
			t.Fatal("library empty during setup")
		}
		g.Players[Player1].Library.Remove(top)
		g.Players[Player1].Hand.Add(top)
	}

	// Put a permanent on the battlefield with counters, attachments, goad, and
	// entry choices.
	perm := &Permanent{
		ObjectID:   g.IDGen.Next(),
		Owner:      Player1,
		Controller: Player1,
		Tapped:     true,
		Counters:   counter.NewSet(),
		Goaded:     map[PlayerID]GoadStatus{Player2: {CreatedTurn: 1, ExpiresFor: Player2}},
	}
	perm.Counters.Add(counter.PlusOnePlusOne, 2)
	perm.Attachments = append(perm.Attachments, g.IDGen.Next())
	perm.EntryChoices = map[ChoiceKey]ResolutionChoiceResult{
		EntryColorChoiceKey: {Kind: ResolutionChoiceMana},
	}
	g.Battlefield = append(g.Battlefield, perm)

	// Push a stack object with maps and slices populated.
	so := &StackObject{
		ID:              g.IDGen.Next(),
		Kind:            StackSpell,
		Controller:      Player1,
		Targets:         []Target{PlayerTarget(Player2)},
		TargetCounts:    []int{1},
		ChosenModes:     []int{0},
		ResolvedAmounts: map[string]int{"damage": 3},
		ResolutionChoices: map[string]ResolutionChoiceResult{
			"color": {Kind: ResolutionChoiceMana},
		},
	}
	g.Stack.Push(so)

	// Add a continuous effect with inner slices.
	g.ContinuousEffects = append(g.ContinuousEffects, ContinuousEffect{
		ID:         g.IDGen.Next(),
		Controller: Player1,
		Layer:      LayerPowerToughnessModify,
		DependsOn:  []id.ID{1, 2},
		AddTypes:   nil,
		PowerDelta: 1,
	})

	// Combat with a blocker order map of slices.
	g.Combat = &CombatState{
		Attackers:        []AttackDeclaration{{Attacker: perm.ObjectID, Target: AttackTarget{Player: Player2}}},
		BlockedAttackers: map[id.ID]bool{perm.ObjectID: true},
		BlockerOrder:     map[id.ID][]id.ID{perm.ObjectID: {10, 11}},
		DamageAssignment: map[id.ID]int{perm.ObjectID: 2},
	}

	// Day/night pointer.
	dn := Night
	g.DayNight = &dn

	// Nested tracking maps.
	g.SkippedSteps[Player1] = map[Step]int{StepDraw: 1}
	g.LinkedObjects[LinkedObjectKey{SourceID: 1, LinkID: "exile"}] = []LinkedObjectRef{{ObjectID: 5, CardID: 6}}
	g.Players[Player1].CommanderDamage[g.Players[Player2].CommanderInstanceID] = 4
	g.TurnOrder.Eliminate(Player4)

	// Record an event.
	g.AppendEvent(Event{Kind: EventCardDrawn, Player: Player1, Amount: 1})

	return g
}

func TestCloneStructurallyEqualImmediately(t *testing.T) {
	g := buildRichGame(t)
	c := g.Clone()

	if c.Players[Player1].Life != g.Players[Player1].Life {
		t.Fatalf("clone life = %d, want %d", c.Players[Player1].Life, g.Players[Player1].Life)
	}
	if len(c.Battlefield) != len(g.Battlefield) {
		t.Fatalf("clone battlefield len = %d, want %d", len(c.Battlefield), len(g.Battlefield))
	}
	if got, want := c.Battlefield[0].Counters.Get(counter.PlusOnePlusOne), g.Battlefield[0].Counters.Get(counter.PlusOnePlusOne); got != want {
		t.Fatalf("clone counters = %d, want %d", got, want)
	}
	if c.Stack.Size() != g.Stack.Size() {
		t.Fatalf("clone stack size = %d, want %d", c.Stack.Size(), g.Stack.Size())
	}
	if len(c.ContinuousEffects) != len(g.ContinuousEffects) {
		t.Fatalf("clone continuous effects = %d, want %d", len(c.ContinuousEffects), len(g.ContinuousEffects))
	}
	if len(c.CardInstances) != len(g.CardInstances) {
		t.Fatalf("clone card instances = %d, want %d", len(c.CardInstances), len(g.CardInstances))
	}
	if c.IDGen.Current() != g.IDGen.Current() {
		t.Fatalf("clone IDGen = %d, want %d", c.IDGen.Current(), g.IDGen.Current())
	}
	if c.Players[Player1].Hand.Size() != g.Players[Player1].Hand.Size() {
		t.Fatalf("clone hand size = %d, want %d", c.Players[Player1].Hand.Size(), g.Players[Player1].Hand.Size())
	}
	if *c.DayNight != *g.DayNight {
		t.Fatalf("clone day/night = %v, want %v", *c.DayNight, *g.DayNight)
	}
	if len(c.Events) != len(g.Events) {
		t.Fatalf("clone events = %d, want %d", len(c.Events), len(g.Events))
	}

	// CardDef pointers must be shared, not deep-copied.
	for cardID, ci := range g.CardInstances {
		if c.CardInstances[cardID].Def != ci.Def {
			t.Fatalf("CardDef for %v was deep-copied; want shared pointer", cardID)
		}
	}
}

func TestCloneCopiesAttackTaxRuleEffectsIndependently(t *testing.T) {
	g := NewGame([NumPlayers]PlayerConfig{})
	g.RuleEffects = []RuleEffect{{
		Kind:             RuleEffectAttackTax,
		AffectedPlayer:   PlayerYou,
		AttackTaxGeneric: 2,
	}}

	clone := g.Clone()
	if len(clone.RuleEffects) != 1 || clone.RuleEffects[0].AttackTaxGeneric != 2 {
		t.Fatalf("clone rule effects = %#v, want copied attack tax", clone.RuleEffects)
	}
	clone.RuleEffects[0].AttackTaxGeneric = 3
	if g.RuleEffects[0].AttackTaxGeneric != 2 {
		t.Fatalf("original attack tax = %d after clone mutation, want 2", g.RuleEffects[0].AttackTaxGeneric)
	}
}

func TestClonePreservesChosenTypeTriggerFields(t *testing.T) {
	g := NewGame([NumPlayers]PlayerConfig{})
	g.ContinuousEffects = []ContinuousEffect{{
		Layer:                     LayerType,
		AffectedSource:            true,
		AddSubtypeFromEntryChoice: EntryTypeChoiceKey,
	}}
	g.RuleEffects = []RuleEffect{{
		Kind: RuleEffectAdditionalTriggerForChosenCreatureType,
	}}
	g.Events = []Event{{
		Kind: EventCardDrawn,
		TriggeredAbilities: []EventTriggeredAbility{{
			AdditionalTriggers:        2,
			TriggerMultiplierCaptured: true,
		}},
	}}
	g.LastKnownInformation[1] = ObjectSnapshot{
		EntryChoices: map[ChoiceKey]ResolutionChoiceResult{
			EntryTypeChoiceKey: {Kind: ResolutionChoiceSubtype},
		},
		RuleEffectKinds: []RuleEffectKind{RuleEffectAdditionalTriggerForChosenCreatureType},
	}

	clone := g.Clone()
	if clone.ContinuousEffects[0].AddSubtypeFromEntryChoice != EntryTypeChoiceKey {
		t.Fatalf("clone choice key = %q, want %q", clone.ContinuousEffects[0].AddSubtypeFromEntryChoice, EntryTypeChoiceKey)
	}
	if clone.RuleEffects[0].Kind != RuleEffectAdditionalTriggerForChosenCreatureType {
		t.Fatalf("clone rule effect kind = %v", clone.RuleEffects[0].Kind)
	}
	if got := clone.Events[0].TriggeredAbilities[0]; got.AdditionalTriggers != 2 || !got.TriggerMultiplierCaptured {
		t.Fatalf("clone captured multiplier = %#v", got)
	}
	clone.Events[0].TriggeredAbilities[0].AdditionalTriggers = 1
	if g.Events[0].TriggeredAbilities[0].AdditionalTriggers != 2 {
		t.Fatal("mutating cloned captured trigger changed the original")
	}
	snapshot := clone.LastKnownInformation[1]
	delete(snapshot.EntryChoices, EntryTypeChoiceKey)
	snapshot.RuleEffectKinds[0] = RuleEffectCantAttack
	if len(g.LastKnownInformation[1].EntryChoices) != 1 ||
		g.LastKnownInformation[1].RuleEffectKinds[0] != RuleEffectAdditionalTriggerForChosenCreatureType {
		t.Fatal("mutating cloned last-known multiplier state changed the original")
	}
}

func TestCloneSeparatesCapturedAndLocalTargetControllerLKI(t *testing.T) {
	g := NewGame([NumPlayers]PlayerConfig{})
	g.Stack.Push(&StackObject{
		TargetControllerLKI:         map[int]PlayerID{0: Player3},
		TargetManaValueLKI:          map[int]int{0: 5},
		CapturedTargetControllerLKI: map[int]PlayerID{0: Player2},
		CapturedTargetManaValueLKI:  map[int]int{0: 4},
	})
	g.DelayedTriggers = append(g.DelayedTriggers, DelayedTrigger{
		CapturedTargetControllerLKI: map[int]PlayerID{0: Player2},
		CapturedTargetManaValueLKI:  map[int]int{0: 4},
	})

	clone := g.Clone()
	obj, ok := clone.Stack.Peek()
	if !ok {
		t.Fatal("cloned stack object missing")
	}
	obj.TargetControllerLKI[0] = Player4
	obj.TargetManaValueLKI[0] = 9
	obj.CapturedTargetControllerLKI[0] = Player4
	obj.CapturedTargetManaValueLKI[0] = 9
	clone.DelayedTriggers[0].CapturedTargetControllerLKI[0] = Player4
	clone.DelayedTriggers[0].CapturedTargetManaValueLKI[0] = 9

	original, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("original stack object missing")
	}
	if original.TargetControllerLKI[0] != Player3 {
		t.Fatalf("original local target LKI = %v, want Player3", original.TargetControllerLKI[0])
	}
	if original.TargetManaValueLKI[0] != 5 {
		t.Fatalf("original local target mana value LKI = %v, want 5", original.TargetManaValueLKI[0])
	}
	if original.CapturedTargetControllerLKI[0] != Player2 {
		t.Fatalf("original captured target LKI = %v, want Player2", original.CapturedTargetControllerLKI[0])
	}
	if original.CapturedTargetManaValueLKI[0] != 4 {
		t.Fatalf("original captured target mana value LKI = %v, want 4", original.CapturedTargetManaValueLKI[0])
	}
	if g.DelayedTriggers[0].CapturedTargetControllerLKI[0] != Player2 {
		t.Fatalf("original delayed captured target LKI = %v, want Player2", g.DelayedTriggers[0].CapturedTargetControllerLKI[0])
	}
	if g.DelayedTriggers[0].CapturedTargetManaValueLKI[0] != 4 {
		t.Fatalf("original delayed captured target mana value LKI = %v, want 4", g.DelayedTriggers[0].CapturedTargetManaValueLKI[0])
	}
}

func TestCloneMutatingCloneDoesNotAffectOriginal(t *testing.T) {
	g := buildRichGame(t)
	c := g.Clone()

	origLife := g.Players[Player1].Life
	origBattlefieldLen := len(g.Battlefield)
	origCounters := g.Battlefield[0].Counters.Get(counter.PlusOnePlusOne)
	origStackSize := g.Stack.Size()
	origEvents := len(g.Events)
	origIDGen := g.IDGen.Current()
	origHand := g.Players[Player1].Hand.Size()
	origLibrary := g.Players[Player1].Library.Size()
	origCmdDmg := g.Players[Player1].CommanderDamage[g.Players[Player2].CommanderInstanceID]
	origGoad := len(g.Battlefield[0].Goaded)
	origBlockerOrder := len(g.Combat.BlockerOrder[g.Battlefield[0].ObjectID])

	// Mutate the clone extensively.
	c.Players[Player1].Life = 1
	c.Battlefield[0].Counters.Add(counter.PlusOnePlusOne, 5)
	c.Battlefield[0].Goaded[Player3] = GoadStatus{}
	c.Battlefield = append(c.Battlefield, &Permanent{ObjectID: c.IDGen.Next()})
	c.Stack.Pop()
	c.ContinuousEffects[0].DependsOn[0] = 999
	c.AppendEvent(Event{Kind: EventLifeLost, Player: Player1, Amount: 5})
	c.IDGen.Next()
	c.Players[Player1].CommanderDamage[g.Players[Player2].CommanderInstanceID] = 21
	c.Combat.BlockerOrder[g.Battlefield[0].ObjectID] = append(c.Combat.BlockerOrder[g.Battlefield[0].ObjectID], 99)
	c.SkippedSteps[Player1][StepDraw] = 99
	*c.DayNight = Day
	// Draw a card on the clone (move zones).
	if top, ok := c.Players[Player1].Library.Top(); ok {
		c.Players[Player1].Library.Remove(top)
		c.Players[Player1].Hand.Add(top)
	}

	// The original must be untouched.
	if g.Players[Player1].Life != origLife {
		t.Errorf("original life changed: %d != %d", g.Players[Player1].Life, origLife)
	}
	if len(g.Battlefield) != origBattlefieldLen {
		t.Errorf("original battlefield changed: %d != %d", len(g.Battlefield), origBattlefieldLen)
	}
	if g.Battlefield[0].Counters.Get(counter.PlusOnePlusOne) != origCounters {
		t.Errorf("original counters changed: %d != %d", g.Battlefield[0].Counters.Get(counter.PlusOnePlusOne), origCounters)
	}
	if len(g.Battlefield[0].Goaded) != origGoad {
		t.Errorf("original goad map changed: %d != %d", len(g.Battlefield[0].Goaded), origGoad)
	}
	if g.Stack.Size() != origStackSize {
		t.Errorf("original stack changed: %d != %d", g.Stack.Size(), origStackSize)
	}
	if g.ContinuousEffects[0].DependsOn[0] == 999 {
		t.Error("original continuous effect inner slice shared with clone")
	}
	if len(g.Events) != origEvents {
		t.Errorf("original events changed: %d != %d", len(g.Events), origEvents)
	}
	if g.IDGen.Current() != origIDGen {
		t.Errorf("original IDGen changed: %d != %d", g.IDGen.Current(), origIDGen)
	}
	if g.Players[Player1].Hand.Size() != origHand {
		t.Errorf("original hand changed: %d != %d", g.Players[Player1].Hand.Size(), origHand)
	}
	if g.Players[Player1].Library.Size() != origLibrary {
		t.Errorf("original library changed: %d != %d", g.Players[Player1].Library.Size(), origLibrary)
	}
	if g.Players[Player1].CommanderDamage[g.Players[Player2].CommanderInstanceID] != origCmdDmg {
		t.Errorf("original commander damage changed: got %d, want %d", g.Players[Player1].CommanderDamage[g.Players[Player2].CommanderInstanceID], origCmdDmg)
	}
	if len(g.Combat.BlockerOrder[g.Battlefield[0].ObjectID]) != origBlockerOrder {
		t.Errorf("original blocker order changed: %d != %d", len(g.Combat.BlockerOrder[g.Battlefield[0].ObjectID]), origBlockerOrder)
	}
	if g.SkippedSteps[Player1][StepDraw] != 1 {
		t.Errorf("original skipped steps changed: %d != 1", g.SkippedSteps[Player1][StepDraw])
	}
	if *g.DayNight != Night {
		t.Errorf("original day/night changed: %v != Night", *g.DayNight)
	}
}

func TestCloneMutatingOriginalDoesNotAffectClone(t *testing.T) {
	g := buildRichGame(t)
	c := g.Clone()

	cloneLife := c.Players[Player1].Life
	cloneBattlefieldLen := len(c.Battlefield)
	cloneCounters := c.Battlefield[0].Counters.Get(counter.PlusOnePlusOne)
	cloneStackSize := c.Stack.Size()
	cloneEvents := len(c.Events)

	// Mutate the original.
	g.Players[Player1].Life = 99
	g.Battlefield[0].Counters.Add(counter.PlusOnePlusOne, 7)
	g.Battlefield = g.Battlefield[:0]
	g.Stack.Pop()
	g.AppendEvent(Event{Kind: EventLifeGained, Player: Player1, Amount: 2})
	*g.DayNight = Day

	if c.Players[Player1].Life != cloneLife {
		t.Errorf("clone life changed: %d != %d", c.Players[Player1].Life, cloneLife)
	}
	if len(c.Battlefield) != cloneBattlefieldLen {
		t.Errorf("clone battlefield changed: %d != %d", len(c.Battlefield), cloneBattlefieldLen)
	}
	if c.Battlefield[0].Counters.Get(counter.PlusOnePlusOne) != cloneCounters {
		t.Errorf("clone counters changed: %d != %d", c.Battlefield[0].Counters.Get(counter.PlusOnePlusOne), cloneCounters)
	}
	if c.Stack.Size() != cloneStackSize {
		t.Errorf("clone stack changed: %d != %d", c.Stack.Size(), cloneStackSize)
	}
	if len(c.Events) != cloneEvents {
		t.Errorf("clone events changed: %d != %d", len(c.Events), cloneEvents)
	}
	if *c.DayNight != Night {
		t.Errorf("clone day/night changed: %v != Night", *c.DayNight)
	}
}

func TestCloneRNGIsIndependentAndNonNil(t *testing.T) {
	g := buildRichGame(t)
	c := g.Clone()

	if c.RNG == nil {
		t.Fatal("clone RNG is nil")
	}
	if c.RNG == g.RNG {
		t.Fatal("clone shares the original RNG pointer")
	}
	// Draining the clone RNG must not change the original's stream.
	before := g.RNG.Uint64()
	for range 10 {
		c.RNG.Uint64()
	}
	after := g.RNG.Uint64()
	if before == after {
		// Two sequential draws being equal is astronomically unlikely; this is a
		// sanity check that the original stream still advances on its own.
		t.Fatal("original RNG stream did not advance independently")
	}
}

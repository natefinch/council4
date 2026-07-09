package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// findAddCounter returns the first AddCounter primitive in an ability's
// resolution sequence.
func findAddCounter(t *testing.T, content game.AbilityContent) game.AddCounter {
	t.Helper()
	for _, mode := range content.Modes {
		for i := range mode.Sequence {
			if add, ok := mode.Sequence[i].Primitive.(game.AddCounter); ok {
				return add
			}
		}
	}
	t.Fatalf("no AddCounter in %#v", content)
	return game.AddCounter{}
}

// TestLowerWillowduskDynamicCounterTarget proves Willowdusk, Essence Seer's
// "{1}, {T}: Choose another target creature. Put a number of +1/+1 counters on
// it equal to the amount of life you gained this turn or the amount of life you
// lost this turn, whichever is greater." lowers to an activated ability that
// places +1/+1 counters on the chosen target, counted by a DynamicAmountMaxOf
// combinator over the controller's life gained and lost this turn.
func TestLowerWillowduskDynamicCounterTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Willowdusk, Essence Seer",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Dryad Druid",
		OracleText: "{1}, {T}: Choose another target creature. Put a number of +1/+1 counters on it equal to the amount of life you gained this turn or the amount of life you lost this turn, whichever is greater. Activate only as a sorcery.",
		Power:      new("0"),
		Toughness:  new("3"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	add := findAddCounter(t, face.ActivatedAbilities[0].Content)
	if add.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counter kind = %v, want PlusOnePlusOne", add.CounterKind)
	}
	if add.Object != game.TargetPermanentReference(0) {
		t.Fatalf("object = %#v, want TargetPermanentReference(0)", add.Object)
	}
	if add.Group.Valid() {
		t.Fatalf("group = %#v, want none for a single target", add.Group)
	}
	dynamic := add.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountMaxOf {
		t.Fatalf("amount = %#v, want DynamicAmountMaxOf", add.Amount)
	}
	wantOperands := []game.DynamicAmountKind{
		game.DynamicAmountLifeGainedThisTurn,
		game.DynamicAmountLifeLostThisTurn,
	}
	got := make([]game.DynamicAmountKind, 0, len(dynamic.Val.Operands))
	for _, operand := range dynamic.Val.Operands {
		got = append(got, operand.Kind)
	}
	if !slices.Equal(got, wantOperands) {
		t.Fatalf("operands = %v, want %v", got, wantOperands)
	}
}

// TestLowerAerithDynamicCounterGroup proves Aerith Gainsborough's "When Aerith
// Gainsborough dies, put X +1/+1 counters on each legendary creature you
// control, where X is the number of +1/+1 counters on Aerith Gainsborough."
// lowers to a dies-triggered ability that places +1/+1 counters on every
// legendary creature the controller controls, counted by a
// DynamicAmountObjectCounters reading the source's +1/+1 counters.
func TestLowerAerithDynamicCounterGroup(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Aerith Gainsborough",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Hero",
		OracleText: "Lifelink\nWhenever you gain life, put a +1/+1 counter on Aerith Gainsborough.\nWhen Aerith Gainsborough dies, put X +1/+1 counters on each legendary creature you control, where X is the number of +1/+1 counters on Aerith Gainsborough.",
		Power:      new("2"),
		Toughness:  new("4"),
	})
	var groupAdd game.AddCounter
	var found bool
	for _, ability := range face.TriggeredAbilities {
		add := findAddCounter(t, ability.Content)
		if add.Group.Valid() {
			groupAdd = add
			found = true
		}
	}
	if !found {
		t.Fatal("no group AddCounter among triggered abilities")
	}
	if groupAdd.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counter kind = %v, want PlusOnePlusOne", groupAdd.CounterKind)
	}
	selection := groupAdd.Group.Selection()
	if !slices.Equal(selection.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("required types = %v, want [Creature]", selection.RequiredTypes)
	}
	if !slices.Equal(selection.Supertypes, []types.Super{types.Legendary}) {
		t.Fatalf("supertypes = %v, want [Legendary]", selection.Supertypes)
	}
	if selection.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", selection.Controller)
	}
	dynamic := groupAdd.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountObjectCounters {
		t.Fatalf("amount = %#v, want DynamicAmountObjectCounters", groupAdd.Amount)
	}
	if dynamic.Val.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counted kind = %v, want PlusOnePlusOne", dynamic.Val.CounterKind)
	}
}

func TestLowerDynamicCounterAmountObjectReferences(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		card          ScryfallCard
		wantObject    game.ObjectReference
		wantAmount    game.DynamicAmountKind
		wantAmountObj game.ObjectReference
		wantCountKind counter.Kind
	}{
		{
			name: "target power",
			card: ScryfallCard{
				Name:       "Soul's Might",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: "Put X +1/+1 counters on target creature, where X is that creature's power.",
			},
			wantObject:    game.TargetPermanentReference(0),
			wantAmount:    game.DynamicAmountObjectPower,
			wantAmountObj: game.TargetPermanentReference(0),
		},
		{
			name: "trigger body target power",
			card: ScryfallCard{
				Name:       "Thickest in the Thicket",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: "When this enchantment enters, put X +1/+1 counters on target creature, where X is that creature's power.",
			},
			wantObject:    game.TargetPermanentReference(0),
			wantAmount:    game.DynamicAmountObjectPower,
			wantAmountObj: game.TargetPermanentReference(0),
		},
		{
			name: "target mana value",
			card: ScryfallCard{
				Name:       "Living Armor",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: "{T}, Sacrifice this artifact: Put X +0/+1 counters on target creature, where X is that creature's mana value.",
			},
			wantObject:    game.TargetPermanentReference(0),
			wantAmount:    game.DynamicAmountObjectManaValue,
			wantAmountObj: game.TargetPermanentReference(0),
		},
		{
			name: "source counter count",
			card: ScryfallCard{
				Name:       "Servant of the Scale",
				Layout:     "normal",
				TypeLine:   "Creature — Soldier",
				OracleText: "This creature enters with a +1/+1 counter on it.\nWhen this creature dies, put X +1/+1 counters on target creature you control, where X is the number of +1/+1 counters on this creature.",
				Power:      new("0"),
				Toughness:  new("0"),
			},
			wantObject:    game.TargetPermanentReference(0),
			wantAmount:    game.DynamicAmountObjectCounters,
			wantAmountObj: game.SourcePermanentReference(),
			wantCountKind: counter.PlusOnePlusOne,
		},
		{
			name: "event permanent power",
			card: ScryfallCard{
				Name:       "Hamletback Goliath",
				Layout:     "normal",
				TypeLine:   "Creature — Giant Warrior",
				OracleText: "Whenever another creature enters, you may put X +1/+1 counters on this creature, where X is that creature's power.",
				Power:      new("6"),
				Toughness:  new("6"),
			},
			wantObject:    game.SourcePermanentReference(),
			wantAmount:    game.DynamicAmountObjectPower,
			wantAmountObj: game.EventPermanentReference(),
		},
		{
			name: "target creature dies event power",
			card: ScryfallCard{
				Name:       "Death's Presence",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: "Whenever a creature you control dies, put X +1/+1 counters on target creature you control, where X is the power of the creature that died.",
			},
			wantObject:    game.TargetPermanentReference(0),
			wantAmount:    game.DynamicAmountObjectPower,
			wantAmountObj: game.EventPermanentReference(),
		},
		{
			name: "target spell mana value",
			card: ScryfallCard{
				Name:       "Draining Whelk",
				Layout:     "normal",
				TypeLine:   "Creature — Illusion",
				OracleText: "Flash\nFlying\nWhen this creature enters, counter target spell. Put X +1/+1 counters on this creature, where X is that spell's mana value.",
				Power:      new("1"),
				Toughness:  new("1"),
			},
			wantObject:    game.SourcePermanentReference(),
			wantAmount:    game.DynamicAmountObjectManaValue,
			wantAmountObj: game.TargetStackObjectReference(0),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &test.card)
			add := firstFaceAddCounter(t, face)
			if add.Object != test.wantObject {
				t.Fatalf("object = %#v, want %#v", add.Object, test.wantObject)
			}
			dynamic := add.Amount.DynamicAmount()
			if !dynamic.Exists || dynamic.Val.Kind != test.wantAmount {
				t.Fatalf("amount = %#v, want kind %v", add.Amount, test.wantAmount)
			}
			if dynamic.Val.Object != test.wantAmountObj {
				t.Fatalf("amount object = %#v, want %#v", dynamic.Val.Object, test.wantAmountObj)
			}
			if dynamic.Val.CounterKind != test.wantCountKind {
				t.Fatalf("counted counter = %v, want %v", dynamic.Val.CounterKind, test.wantCountKind)
			}
		})
	}
}

func TestLowerGroupCounterTriggeringLifeAmount(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Nykthos Paragon",
		Layout:     "normal",
		TypeLine:   "Enchantment Creature — Human Soldier",
		OracleText: "Whenever you gain life, you may put that many +1/+1 counters on each creature you control. Do this only once each turn.",
		Power:      new("4"),
		Toughness:  new("6"),
	})
	add := firstFaceAddCounter(t, face)
	if !add.Group.Valid() {
		t.Fatalf("group = %#v, want battlefield group", add.Group)
	}
	dynamic := add.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountEventLifeChange {
		t.Fatalf("amount = %#v, want DynamicAmountEventLifeChange", add.Amount)
	}
}

func TestLowerTargetCounterDynamicForEachAmount(t *testing.T) {
	t.Parallel()
	mode := spellCounterMode(t, "Put a charge counter on target artifact for each land you control.")
	add, ok := mode.Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddCounter", mode.Sequence[0].Primitive)
	}
	if add.Object != game.TargetPermanentReference(0) || add.CounterKind != counter.Charge {
		t.Fatalf("counter placement = %#v, want charge counters on target 0", add)
	}
	dynamic := add.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountCountSelector {
		t.Fatalf("amount = %#v, want DynamicAmountCountSelector", add.Amount)
	}
	selection := dynamic.Val.Group.Selection()
	if selection.Controller != game.ControllerYou ||
		!slices.Equal(selection.RequiredTypes, []types.Card{types.Land}) {
		t.Fatalf("counted group = %#v, want lands you control", selection)
	}
}

func TestLowerOrderedCounterAmountTargetRemap(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Ordered Dynamic Counter",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Tap target artifact. Put X +1/+1 counters on target creature, where X is that creature's power.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 2 {
		t.Fatalf("targets = %#v, want two", mode.Targets)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", mode.Sequence)
	}
	add, ok := mode.Sequence[1].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("sequence[1] = %T, want game.AddCounter", mode.Sequence[1].Primitive)
	}
	if add.Object != game.TargetPermanentReference(1) {
		t.Fatalf("counter object = %#v, want target 1", add.Object)
	}
	dynamic := add.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Object != game.TargetPermanentReference(1) {
		t.Fatalf("counter amount = %#v, want dynamic target 1", add.Amount)
	}
}

func firstFaceAddCounter(t *testing.T, face loweredFaceAbilities) game.AddCounter {
	t.Helper()
	if face.SpellAbility.Exists {
		if add, ok := firstAddCounter(face.SpellAbility.Val); ok {
			return add
		}
	}
	for _, ability := range face.ActivatedAbilities {
		if add, ok := firstAddCounter(ability.Content); ok {
			return add
		}
	}
	for _, ability := range face.TriggeredAbilities {
		if add, ok := firstAddCounter(ability.Content); ok {
			return add
		}
	}
	t.Fatal("no AddCounter primitive")
	return game.AddCounter{}
}

func firstAddCounter(content game.AbilityContent) (game.AddCounter, bool) {
	for _, mode := range content.Modes {
		for _, instruction := range mode.Sequence {
			if add, ok := instruction.Primitive.(game.AddCounter); ok {
				return add, true
			}
		}
	}
	return game.AddCounter{}, false
}

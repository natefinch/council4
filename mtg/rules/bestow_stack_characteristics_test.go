package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestBestowSpellTypeTransforms exercises the pure game-layer transforms that are
// the single source of truth for a bestowed spell's stack characteristics
// (CR 702.103b): the Creature type is dropped and the Aura subtype is added,
// while other characteristics and ordering are preserved and inputs are not
// mutated.
func TestBestowSpellTypeTransforms(t *testing.T) {
	printedTypes := []types.Card{types.Enchantment, types.Creature}
	gotTypes := game.BestowSpellTypes(printedTypes)
	if slices.Contains(gotTypes, types.Creature) {
		t.Fatalf("BestowSpellTypes kept Creature: %v", gotTypes)
	}
	if !slices.Contains(gotTypes, types.Enchantment) {
		t.Fatalf("BestowSpellTypes dropped Enchantment: %v", gotTypes)
	}
	if len(printedTypes) != 2 || printedTypes[1] != types.Creature {
		t.Fatalf("BestowSpellTypes mutated its input: %v", printedTypes)
	}

	printedSubtypes := []types.Sub{types.Satyr}
	gotSubtypes := game.BestowSpellSubtypes(printedSubtypes)
	if !slices.Contains(gotSubtypes, types.Aura) {
		t.Fatalf("BestowSpellSubtypes did not add Aura: %v", gotSubtypes)
	}
	if !slices.Contains(gotSubtypes, types.Satyr) {
		t.Fatalf("BestowSpellSubtypes dropped a printed subtype: %v", gotSubtypes)
	}
	if len(printedSubtypes) != 1 {
		t.Fatalf("BestowSpellSubtypes mutated its input: %v", printedSubtypes)
	}
	// Adding Aura is idempotent so a card that already prints Aura isn't doubled.
	twice := game.BestowSpellSubtypes([]types.Sub{types.Aura})
	auraCount := 0
	for _, s := range twice {
		if s == types.Aura {
			auraCount++
		}
	}
	if auraCount != 1 {
		t.Fatalf("BestowSpellSubtypes duplicated Aura: %v", twice)
	}
}

// castBestowTester casts the bestowTestCreature bestowed onto a creature and
// returns the stack object and the spell's card instance ID before resolution.
func castBestowTester(t *testing.T, g *game.Game, engine *Engine) (*game.StackObject, game.Target) {
	t.Helper()
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	spellID := addCardToHand(g, game.Player1, bestowTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	targets := []game.Target{game.PermanentTarget(target.ObjectID)}
	if !engine.applyAction(g, game.Player1, action.CastBestowSpell(spellID, targets, 0, nil)) {
		t.Fatal("bestowed cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Bestowed {
		t.Fatalf("stack object = %#v, want Bestowed spell", obj)
	}
	return obj, game.PermanentTarget(target.ObjectID)
}

// TestBestowedSpellOnStackIsAuraNotCreature proves the reviewer's core finding is
// fixed: while on the stack a bestowed spell's effective characteristics are an
// Aura enchantment spell, not a creature spell (CR 702.103b), across the shared
// stack-characteristic readers.
func TestBestowedSpellOnStackIsAuraNotCreature(t *testing.T) {
	g, engine := setupBestowMain(t)
	obj, _ := castBestowTester(t, g, engine)

	chars, ok := stackObjectSourceChars(g, obj)
	if !ok {
		t.Fatal("stackObjectSourceChars failed for bestowed spell")
	}
	if slices.Contains(chars.types, types.Creature) {
		t.Fatalf("bestowed spell chars still list Creature: %v", chars.types)
	}
	if !slices.Contains(chars.types, types.Enchantment) {
		t.Fatalf("bestowed spell chars dropped Enchantment: %v", chars.types)
	}
	if !slices.Contains(chars.subtypes, types.Aura) {
		t.Fatalf("bestowed spell chars missing Aura subtype: %v", chars.subtypes)
	}

	// The shared def-fallback reader (used for StackObjectSourceTypes predicates,
	// which key on SourceCardID) also reports the transformed types. Exercise it
	// with an object whose source card is set, as an ability or battlefield-cast
	// source would be.
	fromCard := &game.StackObject{Kind: game.StackSpell, SourceCardID: obj.SourceID, Face: obj.Face, Bestowed: true}
	if stackObjectSourceHasTypes(g, fromCard, []types.Card{types.Creature}) {
		t.Fatal("bestowed spell reported as a creature by stackObjectSourceHasTypes")
	}
	if !stackObjectSourceHasTypes(g, fromCard, []types.Card{types.Enchantment}) {
		t.Fatal("bestowed spell not reported as an enchantment by stackObjectSourceHasTypes")
	}
	notBestowed := &game.StackObject{Kind: game.StackSpell, SourceCardID: obj.SourceID, Face: obj.Face}
	if !stackObjectSourceHasTypes(g, notBestowed, []types.Card{types.Creature}) {
		t.Fatal("non-bestowed spell not reported as a creature by stackObjectSourceHasTypes")
	}

	// "Counter target creature spell" must NOT accept a bestowed spell, while a
	// predicate matching enchantment spells (an Aura is an enchantment) must.
	creatureSpellSpec := &game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowStackObject,
		Predicate: game.TargetPredicate{
			SpellCardTypes:   []types.Card{types.Creature},
			StackObjectKinds: []game.StackObjectKind{game.StackSpell},
		},
	}
	if stackObjectTargetMatchesSpec(g, game.Player2, 0, creatureSpellSpec, obj.ID) {
		t.Fatal("bestowed spell was a legal target for 'target creature spell'")
	}
	enchantmentSpellSpec := &game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowStackObject,
		Predicate: game.TargetPredicate{
			SpellCardTypes:   []types.Card{types.Enchantment},
			StackObjectKinds: []game.StackObjectKind{game.StackSpell},
		},
	}
	if !stackObjectTargetMatchesSpec(g, game.Player2, 0, enchantmentSpellSpec, obj.ID) {
		t.Fatal("bestowed spell was not a legal target for 'target enchantment spell'")
	}
}

// TestBestowedCastEventReportsTransformedTypes proves the EventSpellCast payload
// carries transformed characteristics, so a "whenever you cast a creature spell"
// trigger does not fire on a bestowed cast while an enchantment-spell trigger
// does (they key on the event's CardTypes/CardSubtypes).
func TestBestowedCastEventReportsTransformedTypes(t *testing.T) {
	g, engine := setupBestowMain(t)
	obj, _ := castBestowTester(t, g, engine)

	assertEvent(t, g.Events, game.EventSpellCast, func(event game.Event) bool {
		return event.StackObjectID == obj.ID &&
			slices.Contains(event.CardTypes, types.Enchantment) &&
			slices.Contains(event.CardSubtypes, types.Aura)
	})
	// A creature-spell cast trigger keys on Creature being present in CardTypes;
	// the bestowed cast event must not carry it.
	assertNoEvent(t, g.Events, game.EventSpellCast, func(event game.Event) bool {
		return event.StackObjectID == obj.ID && slices.Contains(event.CardTypes, types.Creature)
	})
}

// TestNormalCastEventRetainsCreatureType proves the transform is scoped to
// bestowed casts: a normal cast of the same card reports Enchantment Creature
// with no Aura subtype in both the cast event and the shared chars reader.
func TestNormalCastEventRetainsCreatureType(t *testing.T) {
	g, engine := setupBestowMain(t)
	spellID := addCardToHand(g, game.Player1, bestowTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("normal cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Bestowed {
		t.Fatalf("stack object = %#v, want non-bestowed creature spell", obj)
	}

	chars, ok := stackObjectSourceChars(g, obj)
	if !ok {
		t.Fatal("stackObjectSourceChars failed for normal cast")
	}
	if !slices.Contains(chars.types, types.Creature) {
		t.Fatalf("normal cast chars dropped Creature: %v", chars.types)
	}
	if slices.Contains(chars.subtypes, types.Aura) {
		t.Fatalf("normal cast chars gained Aura: %v", chars.subtypes)
	}
	assertEvent(t, g.Events, game.EventSpellCast, func(event game.Event) bool {
		return event.StackObjectID == obj.ID &&
			slices.Contains(event.CardTypes, types.Creature) &&
			slices.Contains(event.CardTypes, types.Enchantment) &&
			!slices.Contains(event.CardSubtypes, types.Aura)
	})
}

// TestBestowedSpellCreatureSelectionForCost proves creature-spell cost/selection
// readers (mana-spend riders) treat a bestowed spell as a noncreature spell:
// a "creature spell" rider does not qualify and a "noncreature spell" rider does,
// while a normally cast enchantment creature is the reverse.
func TestBestowedSpellCreatureSelectionForCost(t *testing.T) {
	g, _ := setupBestowMain(t)
	def := bestowTestCreature()
	spellID := addCardToHand(g, game.Player1, def)

	bestowedSpell := &game.StackObject{Kind: game.StackSpell, SourceID: spellID, Controller: game.Player1, Bestowed: true}
	normalSpell := &game.StackObject{Kind: game.StackSpell, SourceID: spellID, Controller: game.Player1}

	creatureRider := game.ManaRiderInstance{Controller: game.Player1, Rider: game.ManaSpendRider{Condition: game.ManaSpendCastCreatureSpell}}
	noncreatureRider := game.ManaRiderInstance{Controller: game.Player1, Rider: game.ManaSpendRider{Condition: game.ManaSpendCastNoncreatureSpell}}

	if manaSpendConditionSatisfied(g, creatureRider, bestowedSpell, def) {
		t.Fatal("bestowed spell qualified for a creature-spell mana-spend rider")
	}
	if !manaSpendConditionSatisfied(g, noncreatureRider, bestowedSpell, def) {
		t.Fatal("bestowed spell did not qualify for a noncreature-spell mana-spend rider")
	}
	if !manaSpendConditionSatisfied(g, creatureRider, normalSpell, def) {
		t.Fatal("normally cast enchantment creature did not qualify for a creature-spell rider")
	}
	if manaSpendConditionSatisfied(g, noncreatureRider, normalSpell, def) {
		t.Fatal("normally cast enchantment creature qualified for a noncreature-spell rider")
	}
}

// TestStackObjectCardTypeHelpersScopeTransform documents that the reusable
// helpers only transform bestowed spells, leaving abilities and non-bestowed
// spells untouched.
func TestStackObjectCardTypeHelpersScopeTransform(t *testing.T) {
	def := &game.CardDef{CardFace: game.CardFace{
		Types:    []types.Card{types.Enchantment, types.Creature},
		Subtypes: []types.Sub{types.Satyr},
	}}
	bestowed := &game.StackObject{Kind: game.StackSpell, Bestowed: true}
	if got := stackObjectCardTypes(bestowed, def); slices.Contains(got, types.Creature) {
		t.Fatalf("bestowed helper kept Creature: %v", got)
	}
	if got := stackObjectCardSubtypes(bestowed, def); !slices.Contains(got, types.Aura) {
		t.Fatalf("bestowed helper missing Aura: %v", got)
	}
	// A triggered ability object carrying the same source def is not a spell and
	// must not be transformed.
	ability := &game.StackObject{Kind: game.StackTriggeredAbility, Bestowed: true}
	if got := stackObjectCardTypes(ability, def); !slices.Contains(got, types.Creature) {
		t.Fatalf("non-spell object was transformed: %v", got)
	}
	if got := stackObjectCardSubtypes(nil, def); slices.Contains(got, types.Aura) {
		t.Fatalf("nil object was transformed: %v", got)
	}
}

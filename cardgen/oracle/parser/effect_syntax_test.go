package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestParseTemporaryKeywordSubjectExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		{"This creature gains flying until end of turn.", true},
		{"This creature gains trample and haste until end of turn.", true},
		{"Target creature gains flying until end of turn.", true},
		// Unknown keyword stays fail-closed.
		{"This creature gains banding until end of turn.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].Exact != test.exact {
				t.Fatalf("effect Exact = %v, want %v", effects[0].Exact, test.exact)
			}
		})
	}
}

func TestParseLifeLostThisWayAmountExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source    string
		dynamic   bool
		exactGain bool
	}{
		// "equal to the life lost this way" is recognized as a dynamic amount and
		// the gain clause reconstructs exactly.
		{"Each opponent loses 1 life. You gain life equal to the life lost this way.", true, true},
		{"Each opponent loses X life. You gain life equal to the life lost this way.", true, true},
		// A bare fixed life gain stays exact without the dynamic amount (regression
		// guard).
		{"Each opponent loses 1 life. You gain 2 life.", false, true},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			var gain *EffectSyntax
			for si := range document.Abilities[0].Sentences {
				sentence := &document.Abilities[0].Sentences[si]
				for ei := range sentence.Effects {
					if sentence.Effects[ei].Kind == EffectGain {
						gain = &sentence.Effects[ei]
					}
				}
			}
			if gain == nil {
				t.Fatalf("no gain effect parsed from %q", test.source)
			}
			gotDynamic := gain.Amount.DynamicKind == EffectDynamicAmountLifeLostThisWay
			if gotDynamic != test.dynamic {
				t.Fatalf("gain dynamic kind = %v, want LifeLostThisWay=%v", gain.Amount.DynamicKind, test.dynamic)
			}
			if gain.Exact != test.exactGain {
				t.Fatalf("gain Exact = %v, want %v", gain.Exact, test.exactGain)
			}
		})
	}
}

// TestParseSacrificedCreatureAmount verifies that "the sacrificed creature's
// power/toughness/mana value" parses to the matching back-reference dynamic
// kind and reconstructs exactly, including the apostrophe-split possessive
// tokenization.
func TestParseSacrificedCreatureAmount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		kind   EffectDynamicAmountKind
	}{
		{"Sacrifice a creature: Target player mills cards equal to the sacrificed creature's power.", EffectDynamicAmountSacrificedPower},
		{"Sacrifice a creature: Target player mills cards equal to the sacrificed creature's toughness.", EffectDynamicAmountSacrificedToughness},
		{"Sacrifice a creature: Target player mills cards equal to the sacrificed creature's mana value.", EffectDynamicAmountSacrificedManaValue},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{})
			var mill *EffectSyntax
			for si := range document.Abilities[0].Sentences {
				sentence := &document.Abilities[0].Sentences[si]
				for ei := range sentence.Effects {
					if sentence.Effects[ei].Kind == EffectMill {
						mill = &sentence.Effects[ei]
					}
				}
			}
			if mill == nil {
				t.Fatalf("no mill effect parsed from %q", test.source)
			}
			if mill.Amount.DynamicKind != test.kind {
				t.Fatalf("mill dynamic kind = %v, want %v", mill.Amount.DynamicKind, test.kind)
			}
			if !mill.Exact {
				t.Fatal("mill Exact = false, want true")
			}
		})
	}
}

func TestParseGreatestCharacteristicDrawAmount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		kind   EffectDynamicAmountKind
	}{
		{"Draw cards equal to the greatest power among creatures you control.", EffectDynamicAmountGreatestPower},
		{"Draw cards equal to the greatest toughness among creatures you control.", EffectDynamicAmountGreatestToughness},
		{"Draw cards equal to the greatest mana value among permanents you control.", EffectDynamicAmountGreatestManaValue},
		{"Draw cards equal to the greatest power among Mutants you control.", EffectDynamicAmountGreatestPower},
		// Fixed draw stays non-dynamic (regression guard).
		{"Draw two cards.", EffectDynamicAmountNone},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if got := effects[0].Amount.DynamicKind; got != test.kind {
				t.Fatalf("draw dynamic kind = %v, want %v", got, test.kind)
			}
			if test.kind != EffectDynamicAmountNone && effects[0].Amount.Selection == nil {
				t.Fatalf("draw amount missing group selection: %#v", effects[0].Amount)
			}
		})
	}
}

// TestParseDynamicAmountCardsInHandPlusOffset covers the "the number of cards
// in your hand plus N" draw amount that carries a fixed offset on top of the
// dynamic count (Sea Gate Restoration).
func TestParseDynamicAmountCardsInHandPlusOffset(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		kind   EffectDynamicAmountKind
		offset int
	}{
		{"Draw cards equal to the number of cards in your hand plus one.", EffectDynamicAmountCount, 1},
		{"Draw cards equal to the number of cards in your hand plus two.", EffectDynamicAmountCount, 2},
		{"Draw cards equal to the number of cards in your hand plus 3.", EffectDynamicAmountCount, 3},
		// Without an offset rider the dynamic count carries no offset.
		{"Draw cards equal to the number of cards in your hand.", EffectDynamicAmountCount, 0},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if got := effects[0].Amount.DynamicKind; got != test.kind {
				t.Fatalf("draw dynamic kind = %v, want %v", got, test.kind)
			}
			if got := effects[0].Amount.Addend; got != test.offset {
				t.Fatalf("draw amount offset = %d, want %d", got, test.offset)
			}
		})
	}
}

// TestParseColorCountSelfBuffAmount covers the "for each color among <group>"
// and "the number of colors among <group>" dynamic amounts that scale a
// continuous P/T self-buff (Faeburrow Elder).
func TestParseColorCountSelfBuffAmount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		kind   EffectDynamicAmountKind
	}{
		{"This creature gets +1/+1 for each color among permanents you control.", EffectDynamicAmountColorCount},
		{"This creature gets +1/+1 for each color among creatures you control.", EffectDynamicAmountColorCount},
		{"Target creature gets +X/+X until end of turn, where X is the number of colors among permanents you control.", EffectDynamicAmountColorCount},
		// "for each creature" stays a plain count (regression guard).
		{"This creature gets +1/+1 for each creature you control.", EffectDynamicAmountCount},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if got := effects[0].Amount.DynamicKind; got != test.kind {
				t.Fatalf("amount dynamic kind = %v, want %v", got, test.kind)
			}
			if effects[0].Amount.Selection == nil {
				t.Fatalf("amount missing group selection: %#v", effects[0].Amount)
			}
		})
	}
}

// TestParseSharedCreatureTypeAnthemAmount covers the "for each other creature
// ... that shares a creature type with it" dynamic amount that scales the Coat
// of Arms tribal anthem, including the "you control" scope variant and the
// "other attacking creature" combat-scoped variant (Shared Animosity).
func TestParseSharedCreatureTypeAnthemAmount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source    string
		kind      EffectDynamicAmountKind
		attacking bool
	}{
		{"Each creature gets +1/+1 for each other creature on the battlefield that shares a creature type with it.", EffectDynamicAmountSharedCreatureTypeCount, false},
		{"Each creature gets +1/+1 for each other creature on the battlefield that shares at least one creature type with it.", EffectDynamicAmountSharedCreatureTypeCount, false},
		{"Each creature you control gets +1/+1 for each other creature you control that shares a creature type with it.", EffectDynamicAmountSharedCreatureTypeCount, false},
		{"Whenever a creature you control attacks, it gets +1/+0 until end of turn for each other attacking creature that shares a creature type with it.", EffectDynamicAmountSharedCreatureTypeCount, true},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if got := effects[0].Amount.DynamicKind; got != test.kind {
				t.Fatalf("amount dynamic kind = %v, want %v", got, test.kind)
			}
			if effects[0].Amount.Selection == nil {
				t.Fatalf("amount missing group selection: %#v", effects[0].Amount)
			}
			if got := effects[0].Amount.Selection.Attacking; got != test.attacking {
				t.Fatalf("amount group attacking = %v, want %v", got, test.attacking)
			}
		})
	}
}

func TestParseCastAsThoughFlashEffect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		flash  bool
	}{
		// The exact controller-scoped, turn-scoped timing permission is recognized
		// as a single non-optional effect (Borne Upon a Wind, Emergence Zone).
		{"You may cast spells this turn as though they had flash.", true},
		{"Cast spells this turn as though they had flash.", true},
		// Variant wordings fail closed and flow through the generic effect parser.
		{"You may cast a spell this turn as though it had flash.", false},
		{"You may cast spells as though they had flash.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			gotFlash := len(effects) == 1 && effects[0].Kind == EffectCastAsThoughFlash
			if gotFlash != test.flash {
				t.Fatalf("recognized flash permission = %v, want %v (effects=%#v)", gotFlash, test.flash, effects)
			}
			if gotFlash && effects[0].Optional {
				t.Fatal("flash permission should be unconditional, got Optional=true")
			}
		})
	}
}

func TestParseNoMaximumHandSizeForRestOfGameEffect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		noMax  bool
	}{
		// The exact controller-scoped rest-of-game wording is recognized as a
		// single non-optional continuous effect (Sea Gate Restoration,
		// Praetor's Counsel, Wisdom of Ages).
		{"You have no maximum hand size for the rest of the game.", true},
		// Variant wordings fail closed and flow through the generic effect parser.
		{"You have no maximum hand size.", false},
		{"Target player has no maximum hand size for the rest of the game.", false},
		{"You have no maximum hand size until end of turn.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			gotNoMax := len(effects) == 1 && effects[0].Kind == EffectNoMaximumHandSize
			if gotNoMax != test.noMax {
				t.Fatalf("recognized no-maximum-hand-size = %v, want %v (effects=%#v)", gotNoMax, test.noMax, effects)
			}
			if gotNoMax && effects[0].Optional {
				t.Fatal("no-maximum-hand-size effect should be unconditional, got Optional=true")
			}
		})
	}
}

func TestParseTriggeringLifeChangeAmount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source  string
		dynamic bool
	}{
		// "that much life" names the triggering life-change quantity in the
		// life-drain mirror family (Sanguine Bond, Exquisite Blood).
		{"Whenever you gain life, target opponent loses that much life.", true},
		{"Whenever you gain life, each opponent loses that much life.", true},
		{"Whenever an opponent loses life, you gain that much life.", true},
		// A fixed life loss is not a triggering-amount reference.
		{"Target opponent loses 2 life.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{})
			var found bool
			for _, ability := range document.Abilities {
				for _, sentence := range ability.Sentences {
					for _, effect := range sentence.Effects {
						if effect.Amount.DynamicKind == EffectDynamicAmountTriggeringLifeChange {
							found = true
						}
					}
				}
			}
			if found != test.dynamic {
				t.Fatalf("recognized triggering-life-change amount = %v, want %v", found, test.dynamic)
			}
		})
	}
}

// TestParseCantCastSpellsEffect proves the one-shot, turn-scoped player cast
// prohibition is recognized for the opponents and all-players scopes (Silence,
// Mandate of Peace) and the optional spell-type filter ("noncreature spells",
// Ranger-Captain of Eos; "creature spells"), while targeted, referenced, and
// unfiltered-duration wordings fail closed and flow through the generic effect
// parser.
func TestParseCantCastSpellsEffect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source       string
		recognized   bool
		allPlayers   bool
		requiredType []CardType
		excludedType []CardType
	}{
		{"Your opponents can't cast spells this turn.", true, false, nil, nil},
		{"Each opponent can't cast spells this turn.", true, false, nil, nil},
		{"Players can't cast spells this turn.", true, true, nil, nil},
		{"Your opponents can't cast noncreature spells this turn.", true, false, nil, []CardType{CardTypeCreature}},
		{"Players can't cast noncreature spells this turn.", true, true, nil, []CardType{CardTypeCreature}},
		{"Your opponents can't cast creature spells this turn.", true, false, []CardType{CardTypeCreature}, nil},
		// Variant wordings fail closed.
		{"Target player can't cast spells this turn.", false, false, nil, nil},
		{"Your opponents can't cast bogus spells this turn.", false, false, nil, nil},
		{"Your opponents can't cast spells.", false, false, nil, nil},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			got := len(effects) == 1 && effects[0].Kind == EffectCantCastSpells
			if got != test.recognized {
				t.Fatalf("recognized = %v, want %v (effects=%#v)", got, test.recognized, effects)
			}
			if !got {
				return
			}
			if effects[0].CantCastSpellsAllPlayers != test.allPlayers {
				t.Fatalf("allPlayers = %v, want %v", effects[0].CantCastSpellsAllPlayers, test.allPlayers)
			}
			if !cardTypeSlicesEqual(effects[0].CantCastSpellsRequiredTypes, test.requiredType) {
				t.Fatalf("requiredTypes = %v, want %v", effects[0].CantCastSpellsRequiredTypes, test.requiredType)
			}
			if !cardTypeSlicesEqual(effects[0].CantCastSpellsExcludedTypes, test.excludedType) {
				t.Fatalf("excludedTypes = %v, want %v", effects[0].CantCastSpellsExcludedTypes, test.excludedType)
			}
		})
	}
}

func cardTypeSlicesEqual(a, b []CardType) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestParseGroupMustAttackEffect proves the one-shot, turn-scoped forced-attack
// effect is recognized for the you/opponents/all creature group scopes (Bident
// of Thassa) and records the affected group in StaticSubject, while filtered,
// each-combat, and non-creature wordings fail closed and flow through the
// generic effect parser.
func TestParseGroupMustAttackEffect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source     string
		recognized bool
		subject    EffectStaticSubjectKind
	}{
		{"Creatures your opponents control attack this turn if able.", true, EffectStaticSubjectOpponentControlledCreatures},
		{"Creatures you control attack this turn if able.", true, EffectStaticSubjectControlledCreatures},
		{"All creatures attack this turn if able.", true, EffectStaticSubjectAllCreatures},
		// Variant wordings fail closed.
		{"Creatures your opponents control attack each combat if able.", false, EffectStaticSubjectNone},
		{"Goblins you control attack this turn if able.", false, EffectStaticSubjectNone},
		{"Creatures you control attack this turn.", false, EffectStaticSubjectNone},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			got := len(effects) == 1 && effects[0].Kind == EffectMustAttack
			if got != test.recognized {
				t.Fatalf("recognized = %v, want %v (effects=%#v)", got, test.recognized, effects)
			}
			if got && effects[0].StaticSubject.Kind != test.subject {
				t.Fatalf("subject = %v, want %v", effects[0].StaticSubject.Kind, test.subject)
			}
		})
	}
}

// TestParseSpellsCantBeCounteredEffect proves the controller-scoped, turn-scoped
// uncounterable buff is recognized for the next-spell (Mistrise Village) and
// all-spells (Domri, Anarch of Bolas) wordings, while targeted, typed, and
// otherwise-modified wordings fail closed and flow through the generic effect
// parser.
func TestParseSpellsCantBeCounteredEffect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source     string
		recognized bool
		nextOnly   bool
	}{
		{"The next spell you cast this turn can't be countered.", true, true},
		{"Spells you cast this turn can't be countered.", true, false},
		{"The next spell you cast this turn cannot be countered.", true, true},
		{"Spells you control can't be countered this turn.", true, false},
		{"Spells you control cannot be countered this turn.", true, false},
		// Variant wordings fail closed.
		{"The next creature spell you cast this turn can't be countered.", false, false},
		{"Target spell can't be countered.", false, false},
		{"Spells you cast can't be countered.", false, false},
		{"Spells you control can't be countered.", false, false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			got := len(effects) == 1 && effects[0].Kind == EffectSpellsCantBeCountered
			if got != test.recognized {
				t.Fatalf("recognized = %v, want %v (effects=%#v)", got, test.recognized, effects)
			}
			if got && effects[0].SpellsCantBeCounteredNextOnly != test.nextOnly {
				t.Fatalf("nextOnly = %v, want %v", effects[0].SpellsCantBeCounteredNextOnly, test.nextOnly)
			}
		})
	}
}

func TestParsePreventCombatDamageEffect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source     string
		recognized bool
		to         bool
		by         bool
		global     bool
	}{
		{"Prevent all combat damage that would be dealt to and dealt by that creature this turn.", true, true, true, false},
		{"Prevent all combat damage that would be dealt to and dealt by it this turn.", true, true, true, false},
		{"Prevent all combat damage that would be dealt to and dealt by this creature this turn.", true, true, true, false},
		{"Prevent all combat damage that would be dealt by that creature this turn.", true, false, true, false},
		{"Prevent all combat damage that would be dealt to that creature this turn.", true, true, false, false},
		{"Prevent all combat damage that would be dealt this turn.", true, false, false, true},
		// Variant wordings fail closed.
		{"Prevent all combat damage that would be dealt to and dealt by that creature.", false, false, false, false},
		{"Prevent all damage that would be dealt to that creature this turn.", false, false, false, false},
		{"Prevent all combat damage that would be dealt to you this turn.", false, false, false, false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			got := len(effects) == 1 && effects[0].Kind == EffectPreventDamage
			if got != test.recognized {
				t.Fatalf("recognized = %v, want %v (effects=%#v)", got, test.recognized, effects)
			}
			if got {
				if effects[0].PreventDamageTo != test.to {
					t.Fatalf("PreventDamageTo = %v, want %v", effects[0].PreventDamageTo, test.to)
				}
				if effects[0].PreventDamageBy != test.by {
					t.Fatalf("PreventDamageBy = %v, want %v", effects[0].PreventDamageBy, test.by)
				}
				if effects[0].PreventDamageGlobal != test.global {
					t.Fatalf("PreventDamageGlobal = %v, want %v", effects[0].PreventDamageGlobal, test.global)
				}
			}
		})
	}
}

func TestParseDevotionDrawAmount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		kind   EffectDynamicAmountKind
		colors []Color
	}{
		{"Draw cards equal to your devotion to black.", EffectDynamicAmountDevotion, []Color{ColorBlack}},
		{"Draw cards equal to your devotion to white and black.", EffectDynamicAmountDevotion, []Color{ColorWhite, ColorBlack}},
		{"Draw cards equal to your devotion to green.", EffectDynamicAmountDevotion, []Color{ColorGreen}},
		// Fixed draw stays non-dynamic (regression guard).
		{"Draw two cards.", EffectDynamicAmountNone, nil},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if got := effects[0].Amount.DynamicKind; got != test.kind {
				t.Fatalf("draw dynamic kind = %v, want %v", got, test.kind)
			}
			if !slices.Equal(effects[0].Amount.Colors, test.colors) {
				t.Fatalf("draw amount colors = %v, want %v", effects[0].Amount.Colors, test.colors)
			}
		})
	}
}

func TestParseDestroyedThisWayAmount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source     string
		dynamic    bool
		multiplier int
	}{
		// "for each <noun> destroyed this way" is the mass-destroy payoff amount;
		// the gain/draw clause stays a single effect and captures the per-permanent
		// multiplier.
		{"You gain 1 life for each creature destroyed this way.", true, 1},
		{"You gain 2 life for each permanent destroyed this way.", true, 2},
		{"Draw a card for each permanent destroyed this way.", true, 1},
		// A bare fixed life gain stays non-dynamic (regression guard).
		{"You gain 1 life.", false, 0},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			gotDynamic := effects[0].Amount.DynamicKind == EffectDynamicAmountDestroyedThisWay
			if gotDynamic != test.dynamic {
				t.Fatalf("amount dynamic kind = %v, want DestroyedThisWay=%v", effects[0].Amount.DynamicKind, test.dynamic)
			}
			if test.dynamic {
				if effects[0].Amount.DynamicForm != EffectDynamicAmountFormForEach {
					t.Fatalf("amount form = %v, want ForEach", effects[0].Amount.DynamicForm)
				}
				if effects[0].Amount.Multiplier != test.multiplier {
					t.Fatalf("amount multiplier = %d, want %d", effects[0].Amount.Multiplier, test.multiplier)
				}
				if !effects[0].Exact {
					t.Fatal("effect Exact = false, want true so the sequence lowerer can attach the payoff")
				}
			}
		})
	}
}

func TestParseSpellsCastThisTurnAmount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source  string
		dynamic bool
	}{
		// "for each spell you've cast this turn" is the storm-counter amount and
		// the gain clause stays a single effect (no phantom cast effect).
		{"You gain 1 life for each spell you've cast this turn.", true},
		{"You gain 1 life for each spell you have cast this turn.", true},
		// A bare fixed life gain stays non-dynamic (regression guard).
		{"You gain 1 life.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			gotDynamic := effects[0].Amount.DynamicKind == EffectDynamicAmountSpellsCastThisTurn
			if gotDynamic != test.dynamic {
				t.Fatalf("gain dynamic kind = %v, want SpellsCastThisTurn=%v", effects[0].Amount.DynamicKind, test.dynamic)
			}
		})
	}
}

func TestParseLifeChangedThisTurnAmount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		want   EffectDynamicAmountKind
	}{
		// "equal to the life you've lost this turn" (Children of Korlis).
		{"You gain life equal to the life you've lost this turn.", EffectDynamicAmountLifeLostThisTurn},
		{"You gain life equal to the life you have lost this turn.", EffectDynamicAmountLifeLostThisTurn},
		// "the amount of life you gained this turn" (the where-X life-tracking family).
		{"You lose life equal to the amount of life you gained this turn.", EffectDynamicAmountLifeGainedThisTurn},
		{"You gain life equal to the amount of life you lost this turn.", EffectDynamicAmountLifeLostThisTurn},
		{"You lose life equal to the life you gained this turn.", EffectDynamicAmountLifeGainedThisTurn},
		// Fail closed: a bare fixed life change stays non-dynamic.
		{"You gain 1 life.", EffectDynamicAmountNone},
		// Fail closed: unrecognized trailing wording is not the per-turn amount.
		{"You gain life equal to the life you've lost this game.", EffectDynamicAmountNone},
		{"You gain life equal to the life you've spent this turn.", EffectDynamicAmountNone},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if got := effects[0].Amount.DynamicKind; got != test.want {
				t.Fatalf("amount dynamic kind = %v, want %v", got, test.want)
			}
		})
	}
}

func TestParseCreateTokenDynamicCountExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		// Trailing "for each" iterator (the leading form was already exact).
		{"Create a 1/1 green Elf Warrior creature token for each Elf you control.", true},
		// Trailing "for each <card type> you control" (Avenger of Zendikar). The
		// counted permanent's card type must not fold into the token type line.
		{"Create a 0/1 green Plant creature token for each land you control.", true},
		{"Create a 1/1 white Soldier creature token for each artifact you control.", true},
		{"Create a 1/1 white Soldier creature token for each creature you control.", true},
		// "a number of ... equal to" dynamic count, including a keyword rider.
		{"Create a number of 1/1 white Soldier creature tokens equal to the number of opponents you have.", true},
		{"Create a number of 3/3 green Tyranid Warrior creature tokens with trample equal to the number of opponents you have.", true},
		// "where X is" dynamic count.
		{"Create X 1/1 white Soldier creature tokens, where X is the number of creatures you control.", true},
		// Bare variable X (count supplied by the spell's {X}).
		{"Create X 1/1 red Goblin creature tokens.", true},
		// Fixed counts remain exact (regression guard).
		{"Create a 1/1 white Soldier creature token.", true},
		{"Create two 1/1 white Soldier creature tokens.", true},
		// A quoted granted ability is not part of the spec, so it stays fail-closed.
		{`Create X 1/1 red Goblin creature tokens with "{T}: Add {R}."`, false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].Exact != test.exact {
				t.Fatalf("effect Exact = %v, want %v", effects[0].Exact, test.exact)
			}
		})
	}
}

func TestParseSacrificeChoiceFilterExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source     string
		exact      bool
		nonToken   bool
		typesUnion int
	}{
		// Plain filters reconstruct (regression guard).
		{"Each player sacrifices a creature of their choice.", true, false, 0},
		// "nontoken" qualifier reconstructs and records the flag (Accursed Marauder).
		{"Each player sacrifices a nontoken creature of their choice.", true, true, 0},
		// "creature or planeswalker" card-type union reconstructs (Plaguecrafter).
		{"Each player sacrifices a creature or planeswalker of their choice.", true, false, 2},
		// "token" qualifier reconstructs.
		{"Each player sacrifices a token creature of their choice.", true, false, 0},
		// An unmodeled qualifier stays fail-closed.
		{"Each player sacrifices a tapped creature of their choice.", false, false, 0},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].Exact != test.exact {
				t.Fatalf("effect Exact = %v, want %v", effects[0].Exact, test.exact)
			}
			if effects[0].Selection.NonToken != test.nonToken {
				t.Fatalf("selection NonToken = %v, want %v", effects[0].Selection.NonToken, test.nonToken)
			}
			if test.typesUnion != 0 && len(effects[0].Selection.RequiredTypesAny) != test.typesUnion {
				t.Fatalf("RequiredTypesAny = %#v, want %d entries", effects[0].Selection.RequiredTypesAny, test.typesUnion)
			}
		})
	}
}

// TestParseMassEffectLeadingControllerActorExactness verifies that a mass
// effect with a leading "you" controller actor ("You untap all lands you
// control.") reconstructs canonically, including when it appears as a sequence
// clause in an equipment combat-damage trigger (Sword of Feast and Famine).
func TestParseMassEffectLeadingControllerActorExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		// Bare mass untap (regression guard).
		{"Untap all lands you control.", true},
		// Leading "you" controller actor reconstructs.
		{"You untap all lands you control.", true},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].Exact != test.exact {
				t.Fatalf("effect Exact = %v, want %v", effects[0].Exact, test.exact)
			}
		})
	}
}

func TestParseDualTargetBounceExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		// Two independent single permanent targets joined by "and" and the plural
		// possessive destination reconstruct canonically.
		{"Return target permanent you control and target permanent you don't control to their owners' hands.", true},
		{"Return target creature you control and target creature you don't control to their owners' hands.", true},
		{"Return target creature and target land to their owners' hands.", true},
		// A plural target slot is the multi-slot bounce, not the dual form; the
		// dual recognizer requires two cardinality-one targets and stays closed.
		{"Return target creature and two target lands to their owners' hands.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].Exact != test.exact {
				t.Fatalf("effect Exact = %v, want %v", effects[0].Exact, test.exact)
			}
		})
	}
}

func TestParseCreateTokenMultiKeywordExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		// Two keywords joined by "and".
		{"Create a 2/1 black Spider creature token with menace and reach.", true},
		{"Create four 1/1 green Insect creature tokens with flying and deathtouch.", true},
		{"Create a 1/1 blue and red Insect creature token with flying and haste.", true},
		// Three keywords in an Oxford-comma series.
		{"Create a 4/4 white Angel creature token with flying, vigilance, and indestructible.", true},
		// Single keyword stays exact (regression guard).
		{"Create a 4/4 red Dragon creature token with flying.", true},
		// Landwalk evasion keywords on a created token reconstruct exactly.
		{"Create a 1/1 blue Squid creature token with islandwalk.", true},
		{"Create two 2/2 black Horror creature tokens with swampwalk.", true},
		{"Create a 3/3 green Beast creature token with forestwalk and trample.", true},
		// A keyword the token model does not grant fails the whole rider closed.
		{"Create a 3/3 green Beast creature token with trample and devour 2.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].Exact != test.exact {
				t.Fatalf("effect Exact = %v, want %v", effects[0].Exact, test.exact)
			}
		})
	}
}

// TestParseCreateNamedCreatureTokenExactness verifies that a creature token with
// an explicit Oracle name ("... creature token[s] named <Name>") reconstructs
// exactly and captures the name verbatim, while a name followed by a quoted
// granted-ability rider ("... named X with \"...\"") fails closed.
func TestParseCreateNamedCreatureTokenExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
		name   string
	}{
		{"Create a 3/1 red Beast creature token named Carnivore.", true, "Carnivore"},
		{"Create four 3/3 blue Serpent creature tokens named Koma's Coil.", true, "Koma's Coil"},
		{"Create a 1/1 colorless Sliver artifact creature token named Metallic Sliver.", true, "Metallic Sliver"},
		{"Create a 0/1 blue Plant Wall creature token with defender named Kelp.", true, "Kelp"},
		{"Create a 0/1 red Kobold creature token named Kobolds of Kher Keep.", true, "Kobolds of Kher Keep"},
		// A quoted granted-ability rider after the name fails closed and captures
		// no name.
		{`Create a 2/2 blue Demon creature token named Blue Horror with "Whenever you cast an instant or sorcery spell, this token deals 1 damage to any target."`, false, ""},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].Exact != test.exact {
				t.Fatalf("effect Exact = %v, want %v", effects[0].Exact, test.exact)
			}
			if effects[0].TokenName != test.name {
				t.Fatalf("effect TokenName = %q, want %q", effects[0].TokenName, test.name)
			}
		})
	}
}

// TestParseCreateNamedTokenChoiceExactness verifies that a "Create a X token or
// a Y token." choice between two predefined artifact tokens reconstructs exactly
// and sets the TokenChoice flag, while non-predefined or single-token forms do
// not.
func TestParseCreateNamedTokenChoiceExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
		choice bool
	}{
		{"Create a Food token or a Treasure token.", true, true},
		{"Create a Treasure token or a Clue token.", true, true},
		// "your choice of" two-way and Oxford-comma N-way list forms.
		{"Create your choice of a Blood token or a Food token.", true, true},
		{"Create your choice of a Clue token, a Food token, or a Treasure token.", true, true},
		{"Create your choice of a Blood token, a Clue token, a Food token, or a Treasure token.", true, true},
		// Single named token (no choice) stays exact but is not a choice.
		{"Create a Treasure token.", true, false},
		// A non-predefined alternative fails closed.
		{"Create an Incubator token or a Treasure token.", false, true},
		{"Create your choice of an Incubator token, a Food token, or a Treasure token.", false, true},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].Exact != test.exact {
				t.Fatalf("effect Exact = %v, want %v", effects[0].Exact, test.exact)
			}
			if effects[0].TokenChoice != test.choice {
				t.Fatalf("effect TokenChoice = %v, want %v", effects[0].TokenChoice, test.choice)
			}
		})
	}
}

func TestParseRegenerationRider(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		source   string
		prevent  bool
		riders   int
		excluded bool // the destroy effect should still be exact
	}{
		{
			name:     "single target it",
			source:   "Destroy target creature. It can't be regenerated.",
			prevent:  true,
			riders:   1,
			excluded: true,
		},
		{
			name:     "mass they",
			source:   "Destroy all creatures. They can't be regenerated.",
			prevent:  true,
			riders:   1,
			excluded: true,
		},
		{
			// Non-pronoun subject forms stay fail-closed to avoid phantom
			// targets, so the destroy is not credited.
			name:     "that creature subject not credited",
			source:   "Destroy target creature. That creature can't be regenerated.",
			prevent:  false,
			riders:   0,
			excluded: false,
		},
		{
			name:     "a creature destroyed this way",
			source:   "Destroy target creature. A creature destroyed this way can't be regenerated.",
			prevent:  true,
			riders:   1,
			excluded: true,
		},
		{
			name:     "mass destroyed this way",
			source:   "Destroy all creatures. A creature destroyed this way can't be regenerated.",
			prevent:  true,
			riders:   1,
			excluded: true,
		},
		{
			name:     "plural destroyed this way",
			source:   "Destroy all creatures. Creatures destroyed this way can't be regenerated.",
			prevent:  true,
			riders:   1,
			excluded: true,
		},
		{
			// "Dealt damage this way" belongs to a damage effect, not a
			// destroy, so it is not credited as a destroy rider.
			name:     "dealt damage this way not a destroy rider",
			source:   "Destroy target creature. A creature dealt damage this way can't be regenerated.",
			prevent:  false,
			riders:   0,
			excluded: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			ability := document.Abilities[0]
			var destroy *EffectSyntax
			riders := 0
			for i := range ability.Sentences {
				if ability.Sentences[i].RegenerationRider {
					riders++
				}
				for j := range ability.Sentences[i].Effects {
					if ability.Sentences[i].Effects[j].Kind == EffectDestroy {
						destroy = &ability.Sentences[i].Effects[j]
					}
				}
			}
			if destroy == nil {
				t.Fatal("no destroy effect parsed")
			}
			if destroy.PreventRegeneration != test.prevent {
				t.Fatalf("PreventRegeneration = %v, want %v", destroy.PreventRegeneration, test.prevent)
			}
			if riders != test.riders {
				t.Fatalf("rider sentences = %d, want %d", riders, test.riders)
			}
			if destroy.Exact != test.excluded {
				t.Fatalf("destroy Exact = %v, want %v", destroy.Exact, test.excluded)
			}
			if test.prevent {
				if got := len(ability.SemanticReferences); got != 0 {
					t.Fatalf("semantic references = %d, want 0 (rider pronoun excluded)", got)
				}
			}
		})
	}
}

// TestParseRegenerationRiderWithSiblingEffect verifies the "It can't be
// regenerated." rider folds onto a lone destroy even when a recognized sibling
// effect (token creation, life gain) accompanies it, and stays fail-closed when
// more than one destroy effect makes the lone-destroy fold ambiguous.
func TestParseRegenerationRiderWithSiblingEffect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		prevent bool
		riders  int
	}{
		{
			name:    "destroy then controller token",
			source:  "Destroy target creature. It can't be regenerated. Its controller creates a 3/3 green Ape creature token.",
			prevent: true,
			riders:  1,
		},
		{
			name:    "destroy then controller life",
			source:  "Destroy target artifact. It can't be regenerated. That artifact's controller gains life equal to its mana value.",
			prevent: true,
			riders:  1,
		},
		{
			// Two destroy effects leave the rider pronoun ambiguous, so it stays
			// uncredited and the card fails closed downstream.
			name:    "two destroys not credited",
			source:  "Destroy target creature. Destroy target artifact. It can't be regenerated.",
			prevent: false,
			riders:  0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			ability := document.Abilities[0]
			var destroy *EffectSyntax
			riders := 0
			for i := range ability.Sentences {
				if ability.Sentences[i].RegenerationRider {
					riders++
				}
				for j := range ability.Sentences[i].Effects {
					if ability.Sentences[i].Effects[j].Kind == EffectDestroy {
						destroy = &ability.Sentences[i].Effects[j]
					}
				}
			}
			if destroy == nil {
				t.Fatal("no destroy effect parsed")
			}
			if destroy.PreventRegeneration != test.prevent {
				t.Fatalf("PreventRegeneration = %v, want %v", destroy.PreventRegeneration, test.prevent)
			}
			if riders != test.riders {
				t.Fatalf("rider sentences = %d, want %d", riders, test.riders)
			}
		})
	}
}

func TestParseOptionalControllerEffectExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source   string
		exact    bool
		optional bool
	}{
		// A controller "you may" resolving optional carries the optionality in
		// EffectSyntax.Optional and reconstructs the canonical verb clause
		// byte-exactly, so it stays exact for the life and token recognizers.
		{"You may gain 3 life.", true, true},
		{"You may lose 2 life.", true, true},
		{"You may create a 1/1 white Soldier creature token.", true, true},
		{"You may create a Treasure token.", true, true},
		// The mandatory forms remain exact (regression guard).
		{"Gain 3 life.", true, false},
		{"Create a 1/1 white Soldier creature token.", true, false},
		// A non-controller "may" cannot be modeled by a single controller-asked
		// optional instruction, so it must not become exact.
		{"Each opponent may gain 3 life.", false, true},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].Optional != test.optional {
				t.Errorf("effect Optional = %v, want %v", effects[0].Optional, test.optional)
			}
			if effects[0].Exact != test.exact {
				t.Errorf("effect Exact = %v, want %v", effects[0].Exact, test.exact)
			}
		})
	}
}

func TestParseCreateNamedTokenExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		{"Create a Treasure token.", true},
		{"Create a Food token.", true},
		{"Create a Clue token.", true},
		{"Create a Blood token.", true},
		{"Create a Gold token.", true},
		{"Create a Lander token.", true},
		{"Create a Mutagen token.", true},
		{"Create a Map token.", true},
		{"Create a Junk token.", true},
		{"Create a Powerstone token.", true},
		{"Create two Treasure tokens.", true},
		// Incubator's transform ability is not yet modeled, so it stays
		// fail-closed.
		{"Create an Incubator token.", false},
		// A "tapped" entry on a recognized named token is now representable.
		{"Create a tapped Treasure token.", true},
		{"Create a tapped Lander token.", true},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].Kind != EffectCreate {
				t.Fatalf("effect kind = %v, want EffectCreate", effects[0].Kind)
			}
			if effects[0].Exact != test.exact {
				t.Fatalf("effect Exact = %v, want %v", effects[0].Exact, test.exact)
			}
		})
	}
}

func TestParseCreateCreatureTokenTypeExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		// Colorless vanilla creature tokens, single and multiple.
		{"Create a 1/1 colorless Hero creature token.", true},
		{"Create four 1/1 colorless Hero creature tokens.", true},
		// Artifact- and enchantment-creature tokens, optionally keyworded.
		{"Create a 1/1 colorless Thopter artifact creature token with flying.", true},
		{"Create two 1/1 colorless Thopter artifact creature tokens with flying.", true},
		{"Create a 3/3 colorless Phyrexian Golem artifact creature token.", true},
		{"Create a 1/1 white Glimmer enchantment creature token.", true},
		{"Create a 1/3 green Spider enchantment creature token with reach.", true},
		// A "tapped" entry is now representable, alone and with a keyword.
		{"Create a tapped 2/2 black Zombie creature token.", true},
		{"Create two tapped 1/1 white Dog creature tokens.", true},
		{"Create three tapped 1/1 white Spirit creature tokens with flying.", true},
		// A trailing attacking entry clause is now representable: tapped and
		// attacking (singular and plural), attacking-only, and after a keyword.
		{"Create a 2/2 green Boar creature token that's tapped and attacking.", true},
		{"Create two 1/1 red Human creature tokens that are tapped and attacking.", true},
		{"Create a 1/1 white Soldier creature token that's attacking.", true},
		{"Create a 1/1 white Cat Soldier creature token with vigilance that's attacking.", true},
		// A "blocking" entry remains unrepresentable and stays fail-closed.
		{"Create a 2/2 green Boar creature token that's tapped and blocking.", false},
		// A quoted granted ability is not representable and stays fail-closed.
		{"Create a 1/1 black Rat creature token with \"This token can't block.\"", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].Kind != EffectCreate {
				t.Fatalf("effect kind = %v, want EffectCreate", effects[0].Kind)
			}
			if effects[0].Exact != test.exact {
				t.Fatalf("effect Exact = %v, want %v", effects[0].Exact, test.exact)
			}
		})
	}
}

func TestParseCreateTokenTargetRecipientExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		// A single targeted player is a representable recipient for fixed counts.
		{"Target opponent creates a 4/4 black Horror creature token.", true},
		{"Target player creates a 1/1 white Soldier creature token.", true},
		{"Target opponent creates two Treasure tokens.", true},
		// Player-group recipients are not a single player reference; stay fail-closed.
		{"Each opponent creates a 1/1 white Human creature token.", false},
		{"Each player creates a 1/1 green Cat creature token.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].Kind != EffectCreate {
				t.Fatalf("effect kind = %v, want EffectCreate", effects[0].Kind)
			}
			if effects[0].Exact != test.exact {
				t.Fatalf("effect Exact = %v, want %v", effects[0].Exact, test.exact)
			}
		})
	}
}

func TestParseManaValueTargetExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		{"Exile target permanent with mana value 4 or greater.", true},
		{"Exile target creature with mana value 3 or greater.", true},
		{"Exile target creature with mana value 3 or less.", true},
		{"Exile target permanent with mana value 1.", true},
		{"Destroy target artifact with mana value 2 or less.", true},
		{"Destroy target tapped creature with mana value 3 or greater.", true},
		// A two-color union ("black or red") reconstructs canonically as
		// "<color> or <color> <noun>" and is exact.
		{"Exile target black or red permanent.", true},
		// A multicolored qualifier is not representable and must stay fail-closed.
		{"Exile target multicolored permanent with mana value 3 or greater.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || len(effects[0].Targets) != 1 {
				t.Fatalf("effects = %#v, want one effect with one target", effects)
			}
			if effects[0].Targets[0].Exact != test.exact {
				t.Fatalf("target Exact = %v, want %v", effects[0].Targets[0].Exact, test.exact)
			}
		})
	}
}

func TestParseMultiPermanentUnionTargetExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		{"Destroy up to one target artifact or enchantment.", true},
		{"Destroy up to two target artifacts or enchantments.", true},
		{"Destroy two target artifacts or enchantments.", true},
		{"Destroy up to one target creature or planeswalker.", true},
		{"Exile up to one other target creature or artifact you control.", true},
		{"Destroy up to one target artifact or enchantment you control.", true},
		// "and/or" is an equivalent union conjunction the round-trip also renders.
		{"Destroy two target artifacts, creatures, and/or lands you control.", true},
		{"Destroy two target artifacts and/or enchantments.", true},
		// A trailing keyword qualifier on a union is not reconstructed (it would
		// apply to only one branch) and must stay fail-closed.
		{"Destroy up to one target artifact or creature with flying.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || len(effects[0].Targets) != 1 {
				t.Fatalf("effects = %#v, want one effect with one target", effects)
			}
			if effects[0].Targets[0].Exact != test.exact {
				t.Fatalf("target Exact = %v, want %v", effects[0].Targets[0].Exact, test.exact)
			}
		})
	}
}

func TestParseExcludedTypeManaValueTargetExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		{"Destroy target nonland permanent with mana value 3 or less.", true},
		{"Destroy target noncreature permanent with mana value 2 or less.", true},
		{"Destroy target nonland permanent with mana value 4 or greater.", true},
		{"Destroy target nonland permanent.", true},
		// Power and toughness exist only on creatures, so a power/toughness
		// qualifier on a non-creature noun must stay fail-closed.
		{"Destroy target nonland permanent with power 3 or less.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || len(effects[0].Targets) != 1 {
				t.Fatalf("effects = %#v, want one effect with one target", effects)
			}
			if effects[0].Targets[0].Exact != test.exact {
				t.Fatalf("target Exact = %v, want %v", effects[0].Targets[0].Exact, test.exact)
			}
		})
	}
}

func TestParseExcludedColorTypeTargetExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		{"Destroy target nonblack creature.", true},
		{"Destroy target nonwhite permanent.", true},
		{"Destroy target noncreature artifact.", true},
		{"Destroy target nonartifact creature.", true},
		{"Destroy target nonwhite creature you control.", true},
		{"Destroy target creature.", true},
		// Two excluded colors are not reconstructed and must stay fail-closed.
		{"Destroy target nonblack nonred creature.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || len(effects[0].Targets) != 1 {
				t.Fatalf("effects = %#v, want one effect with one target", effects)
			}
			if effects[0].Targets[0].Exact != test.exact {
				t.Fatalf("target Exact = %v, want %v", effects[0].Targets[0].Exact, test.exact)
			}
		})
	}
}

func TestParseExcludedSupertypeTargetExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		{"Destroy target nonbasic land.", true},
		{"Destroy target nonlegendary creature.", true},
		{"Destroy target nonsnow creature.", true},
		{"Destroy target nonbasic land you control.", true},
		// A supertype paired with an excluded supertype is not reconstructed and
		// must stay fail-closed.
		{"Destroy target basic nonsnow land.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || len(effects[0].Targets) != 1 {
				t.Fatalf("effects = %#v, want one effect with one target", effects)
			}
			if effects[0].Targets[0].Exact != test.exact {
				t.Fatalf("target Exact = %v, want %v", effects[0].Targets[0].Exact, test.exact)
			}
		})
	}
}

func TestParseColorSpellTargetExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		{"Counter target blue spell.", true},
		{"Counter target nonblue spell.", true},
		{"Counter target colorless spell.", true},
		{"Counter target multicolored spell.", true},
		// Monocolored spells have no canonical predicate yet and stay fail-closed.
		{"Counter target monocolored spell.", false},
		// A color combined with a card-type filter is not reconstructed.
		{"Counter target blue creature spell.", false},
		// Two colors are not reconstructed and must stay fail-closed.
		{"Counter target blue and white spell.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || len(effects[0].Targets) != 1 {
				t.Fatalf("effects = %#v, want one effect with one target", effects)
			}
			if effects[0].Targets[0].Exact != test.exact {
				t.Fatalf("target Exact = %v, want %v", effects[0].Targets[0].Exact, test.exact)
			}
		})
	}
}

func TestParseMultiTargetExcludedTypeExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		// A single excluded card type on a multi-target or optional permanent
		// reconstructs canonically as "non<type> <noun>(s)".
		{"Return up to two target nonland permanents to their owners' hands.", true},
		{"Return six target nonland permanents to their owners' hands.", true},
		{"Return up to one target nonland permanent to its owner's hand.", true},
		{"Destroy up to two target noncreature artifacts.", true},
		{"Destroy up to one other target noncreature permanent you control.", true},
		// A subtype qualifier on a multi-target permanent is not reconstructed and
		// must stay fail-closed.
		{"Return up to two target Goblin creatures to their owners' hands.", false},
		// Two excluded types are not reconstructed and must stay fail-closed.
		{"Destroy up to two target nonland noncreature permanents.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || len(effects[0].Targets) != 1 {
				t.Fatalf("effects = %#v, want one effect with one target", effects)
			}
			if effects[0].Targets[0].Exact != test.exact {
				t.Fatalf("target Exact = %v, want %v", effects[0].Targets[0].Exact, test.exact)
			}
		})
	}
}

func TestParseResolvingEffectKinds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		kind   EffectKind
	}{
		{"Add {G}.", EffectAddMana},
		{"Attach target Equipment to target creature.", EffectAttach},
		{"Cast that card.", EffectCast},
		{"Counter target spell.", EffectCounter},
		{"Create a token.", EffectCreate},
		{"Deal 2 damage to any target.", EffectDealDamage},
		{"Destroy target creature.", EffectDestroy},
		{"Discard a card.", EffectDiscard},
		{"Discover 3.", EffectDiscover},
		{"Amass Orcs 1.", EffectAmass},
		{"Amass Zombies 3.", EffectAmass},
		{"Amass 2.", EffectAmass},
		{"Double its power.", EffectDouble},
		{"Draw a card.", EffectDraw},
		{"This land enters tapped.", EffectEnterTapped},
		{"This creature enters prepared.", EffectEnterPrepared},
		{"Exile target creature.", EffectExile},
		{"Target creature fights target creature.", EffectFight},
		{"Gain 2 life.", EffectGain},
		{"Gain control of target creature.", EffectGainControl},
		{"Target creature has flying.", EffectGrantKeyword},
		{"Investigate.", EffectInvestigate},
		{"Target creature explores.", EffectExplore},
		{"Lose 2 life.", EffectLose},
		{"Manifest the top card of your library.", EffectManifest},
		{"Manifest dread.", EffectManifestDread},
		{"Its controller manifests the top card of their library.", EffectManifest},
		{"Its controller manifests dread.", EffectManifestDread},
		{"Look at the top two cards of your library.", EffectDig},
		{"Mill two cards.", EffectMill},
		{"Target creature gets +2/+2.", EffectModifyPT},
		{"Put a +1/+1 counter on target creature.", EffectPut},
		{"Proliferate.", EffectProliferate},
		{"Regenerate target creature.", EffectRegenerate},
		{"Return target creature to its owner's hand.", EffectReturn},
		{"Reveal that card.", EffectReveal},
		{"Sacrifice a creature.", EffectSacrifice},
		{"Scry 2.", EffectScry},
		{"Surveil 2.", EffectSurveil},
		{"Search your library for a card.", EffectSearch},
		{"Shuffle your library.", EffectShuffle},
		{"Tap target creature.", EffectTap},
		{"Untap target creature.", EffectUntap},
		{"Transform target creature.", EffectTransform},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) == 0 || effects[0].Kind != test.kind {
				t.Fatalf("effects = %#v, want first kind %v", effects, test.kind)
			}
		})
	}
}

func TestParseAmassSubtype(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source  string
		subtype types.Sub
		amount  int
	}{
		{"Amass Orcs 1.", types.Orc, 1},
		{"Amass Zombies 3.", types.Zombie, 3},
		{"Amass 2.", types.Zombie, 2},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want exactly one", effects)
			}
			if effects[0].Kind != EffectAmass {
				t.Fatalf("kind = %v, want EffectAmass", effects[0].Kind)
			}
			if effects[0].AmassSubtype != test.subtype {
				t.Fatalf("AmassSubtype = %q, want %q", effects[0].AmassSubtype, test.subtype)
			}
			if !effects[0].Amount.Known || effects[0].Amount.Value != test.amount {
				t.Fatalf("Amount = %#v, want known %d", effects[0].Amount, test.amount)
			}
		})
	}
}

func TestParseMassBounceEffectExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		{"Return all creatures to their owners' hands.", true},
		{"Return all permanents to their owners' hands.", true},
		{"Return all lands to their owners' hands.", true},
		{"Return all artifacts and enchantments to their owners' hands.", true},
		{"Return all nonblue creatures to their owners' hands.", true},
		{"Return all artifacts you control to their owner's hand.", true},
		{"Return all creatures you control to their owner's hand.", true},
		{"Return all permanents you control to their owners' hands.", true},
		// Combat-state filters route to the same group bounce (CR 614/701.x).
		{"Return all attacking creatures to their owner's hand.", true},
		{"Return all blocking creatures to their owners' hands.", true},
		// Choice- and filter-based groups the executable backend cannot express stay fail-closed.
		{"Return all permanents of the color of your choice to their owners' hands.", false},
		{"Return all creatures to their owners' hands except for Krakens.", false},
		// "Return a permanent you control" is a controlled-choice bounce (the
		// resolving controller chooses one permanent they control), now exact.
		{"Return a permanent you control to its owner's hand.", true},
		// "each" stays fail-closed; the compiler cannot distinguish it from "a".
		{"Return each creature to its owner's hand.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || effects[0].Exact != test.exact {
				t.Fatalf("effects = %#v, want one effect with Exact=%v", effects, test.exact)
			}
		})
	}
}

func TestParseSelfNameBounceEffectExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source   string
		cardName string
		exact    bool
	}{
		// The source permanent returning itself, named by the card's own name.
		{"Return Selenia to its owner's hand.", "Selenia, Dark Angel", true},
		{"Return Oboro to its owner's hand.", "Oboro, Palace in the Clouds", true},
		{"Return Wydwen to its owner's hand.", "Wydwen, the Biting Gale", true},
		// The "this <object>" self form stays exact alongside the name form.
		{"Return this creature to its owner's hand.", "Test Merfolk", true},
		// A different name fails the byte-exact round-trip and stays fail-closed.
		{"Return Bob to its owner's hand.", "Selenia, Dark Angel", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{CardName: test.cardName})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || effects[0].Exact != test.exact {
				t.Fatalf("effects = %#v, want one effect with Exact=%v", effects, test.exact)
			}
		})
	}
}

func TestParseMassTapUntapEffectExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		// Mass tap/untap reuse the bare mass-group phrasing the executable
		// backend already renders for destroy and exile.
		{"Tap all creatures your opponents control.", true},
		{"Tap all artifacts.", true},
		{"Tap all nonwhite creatures.", true},
		{"Tap all other creatures.", true},
		{"Untap all creatures you control.", true},
		{"Untap all lands you control.", true},
		{"Untap all permanents.", true},
		// Choice- and filter-based groups the backend cannot express stay
		// fail-closed, exactly as the mass bounce/destroy paths do.
		{"Tap all tapped Goblins.", false},
		{"Untap all artifacts with power 3 or less.", false},
		{"Tap all artifacts and enchantments you control.", true},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || effects[0].Exact != test.exact {
				t.Fatalf("effects = %#v, want one effect with Exact=%v", effects, test.exact)
			}
		})
	}
}

func TestParseControlledChoiceBounceExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		// Supported controlled-choice bounce forms (resolving controller chooses).
		{"Return a permanent you control to its owner's hand.", true},
		{"Return a creature you control to its owner's hand.", true},
		{"Return a land you control to its owner's hand.", true},
		{"Return an artifact you control to its owner's hand.", true},
		{"Return another permanent you control to its owner's hand.", true},
		{"Return another creature you control to its owner's hand.", true},
		{"Return a white creature you control to its owner's hand.", true},
		// Fail-closed: no controller restriction (not "you control").
		{"Return a permanent to its owner's hand.", false},
		// Fail-closed: opponent-controlled choice is not modeled here.
		{"Return a creature an opponent controls to its owner's hand.", false},
		// Fail-closed: excluded-type predicates the chooser cannot express.
		{"Return a nonland permanent you control to its owner's hand.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || effects[0].Exact != test.exact {
				t.Fatalf("effects = %#v, want one effect with Exact=%v", effects, test.exact)
			}
		})
	}
}

func TestParseResolvingEffectExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
	}{
		{"Scry 2.", true},
		{"Scry 2, then celebrate.", false},
		{"Surveil two.", true},
		{"Surveil two, then celebrate.", false},
		{"Investigate.", true},
		{"Investigate twice.", true},
		{"Investigate, then celebrate.", false},
		{"Proliferate.", true},
		{"Proliferate two times.", true},
		{"Proliferate, then celebrate.", false},
		{"Creatures you control get +2/+2 until end of turn.", true},
		{"Creatures you control get +2/+2 until end of turn, then celebrate.", false},
		{"This creature gets +2/+0 until end of turn.", true},
		{"This creature gets +1/+1 until end of turn, then celebrate.", false},
		{"Put a +1/+1 counter on this creature.", true},
		{"Put a +1/+1 counter on this creature, then celebrate.", false},
		{"Gain control of target creature.", true},
		{"Gain control of target creature until end of turn.", true},
		{"Gain control of target creature for as long as you control this creature.", true},
		{"Gain control of target creature until end of turn, then celebrate.", false},
		{"Sacrifice a creature.", true},
		{"You sacrifice a creature.", true},
		{"Sacrifice two permanents.", true},
		{"Each opponent sacrifices a creature.", true},
		{"Each other player sacrifices a creature.", true},
		{"Sacrifice a creature, then celebrate.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || effects[0].Exact != test.exact {
				t.Fatalf("effects = %#v, want one effect with Exact=%v", effects, test.exact)
			}
		})
	}
}

func TestParseCreateCopyOfTargetToken(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		copy   bool
		amount int
		tapped bool
	}{
		{"Create a token that's a copy of target creature you control.", true, 1, false},
		{"Create a token that's a copy of target artifact.", true, 1, false},
		{"Create two tokens that are copies of target creature.", true, 2, false},
		{"Create three tokens that are copies of target creature you control.", true, 3, false},
		{"Create a tapped token that's a copy of target creature.", true, 1, true},
		{"Create two tapped tokens that are copies of target creature.", true, 2, true},
		{"Create a 1/1 white Soldier creature token.", false, 0, false},
		{"Create a token that's a copy of target creature you control, then celebrate.", false, 0, false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].TokenCopyOfTarget != test.copy {
				t.Fatalf("TokenCopyOfTarget = %v, want %v", effects[0].TokenCopyOfTarget, test.copy)
			}
			if test.copy && !effects[0].Exact {
				t.Fatalf("copy token effect should be exact: %#v", effects[0])
			}
			if test.copy && effects[0].Amount.Value != test.amount {
				t.Fatalf("Amount.Value = %d, want %d", effects[0].Amount.Value, test.amount)
			}
			if test.copy && effects[0].TokenCopyEntersTapped != test.tapped {
				t.Fatalf("TokenCopyEntersTapped = %v, want %v", effects[0].TokenCopyEntersTapped, test.tapped)
			}
		})
	}
}

func TestParseCreateCopyTokenForEach(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source    string
		forEach   bool
		tokenOnly bool
	}{
		{"For each token you control, create a token that's a copy of that permanent.", true, true},
		{"For each creature you control, create a token that's a copy of that permanent.", true, false},
		{"For each artifact you control, create a token that's a copy of it.", true, false},
		{"Create a token that's a copy of target creature you control.", false, false},
		{"For each token an opponent controls, create a token that's a copy of that permanent.", false, false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].TokenCopyOfForEach != test.forEach {
				t.Fatalf("TokenCopyOfForEach = %v, want %v", effects[0].TokenCopyOfForEach, test.forEach)
			}
			if !test.forEach {
				return
			}
			group := effects[0].TokenCopyForEachGroup
			if group == nil {
				t.Fatal("TokenCopyForEachGroup = nil, want a group")
			}
			if group.Controller != SelectionControllerYou {
				t.Fatalf("group controller = %v, want you", group.Controller)
			}
			if group.TokenOnly != test.tokenOnly {
				t.Fatalf("group TokenOnly = %v, want %v", group.TokenOnly, test.tokenOnly)
			}
		})
	}
}

func TestParseCreateCopyOfReferenceToken(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		copy   bool
		tapped bool
	}{
		{"Create a token that's a copy of this creature.", true, false},
		{"Create a token that's a copy of this creature instead.", true, false},
		{"Create two tokens that are copies of this creature.", true, false},
		{"Create a tapped token that's a copy of this creature.", true, true},
		{"Create two tapped tokens that are copies of this creature.", true, true},
		{"Create a token that's a copy of target creature you control.", false, false},
		{"Create a 1/1 white Soldier creature token.", false, false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{CardName: "Scute Swarm", InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].TokenCopyOfReference != test.copy {
				t.Fatalf("TokenCopyOfReference = %v, want %v", effects[0].TokenCopyOfReference, test.copy)
			}
			if test.copy && effects[0].TokenCopyEntersTapped != test.tapped {
				t.Fatalf("TokenCopyEntersTapped = %v, want %v", effects[0].TokenCopyEntersTapped, test.tapped)
			}
			if test.copy && !effects[0].Exact {
				t.Fatalf("copy-of-reference token effect should be exact: %#v", effects[0])
			}
		})
	}
}

func TestParseCreateCopyTokenModifiers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		source        string
		attached      bool
		dropLegendary bool
		grantKeywords int
	}{
		{
			name:          "target except not legendary",
			source:        "Create a token that's a copy of target creature you control, except it isn't legendary.",
			dropLegendary: true,
		},
		{
			name:          "attached except not legendary gains keyword",
			source:        "Create a token that's a copy of equipped creature, except the token isn't legendary. That token gains haste.",
			attached:      true,
			dropLegendary: true,
			grantKeywords: 1,
		},
		{
			name:     "attached enchanted creature",
			source:   "Create a token that's a copy of enchanted creature.",
			attached: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{CardName: "Test Copy", InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			effect := effects[0]
			if !effect.Exact {
				t.Fatalf("copy token effect should be exact: %#v", effect)
			}
			if effect.TokenCopyOfAttached != test.attached {
				t.Fatalf("TokenCopyOfAttached = %v, want %v", effect.TokenCopyOfAttached, test.attached)
			}
			if effect.TokenCopyDropLegendary != test.dropLegendary {
				t.Fatalf("TokenCopyDropLegendary = %v, want %v", effect.TokenCopyDropLegendary, test.dropLegendary)
			}
			if len(effect.TokenCopyGrantKeywords) != test.grantKeywords {
				t.Fatalf("TokenCopyGrantKeywords = %#v, want %d", effect.TokenCopyGrantKeywords, test.grantKeywords)
			}
		})
	}
}

func TestParseGainControlSequenceExactness(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Untap target creature and gain control of it until end of turn. That creature gains haste until end of turn.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 2 || !effects[0].Exact || !effects[1].Exact {
		t.Fatalf("effects = %#v, want two exact effects", effects)
	}
}

func TestParseGainControlFollowOnExactness(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Gain control of target creature until end of turn. Untap that creature. It gains haste until end of turn.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, sentence := range document.Abilities[0].Sentences {
		for _, effect := range sentence.Effects {
			if !effect.Exact {
				t.Errorf("%v effect is not exact: %#v", effect.Kind, effect)
			}
		}
	}
}

func TestParseSupportedGainControlEffectsExact(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		context Context
	}{
		{
			name:   "kicked trigger",
			source: "When this creature enters, if it was kicked, gain control of target creature until end of turn. Untap that creature. It gains haste until end of turn.",
		},
		{
			name:    "optional source duration trigger",
			source:  "Whenever a land you control enters, you may gain control of target creature for as long as you control this creature.",
			context: Context{CardName: "Control Creature"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, test.context)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, sentence := range document.Abilities[0].Sentences {
				for _, effect := range sentence.Effects {
					if !effect.Exact {
						t.Errorf("%v effect is not exact: %#v", effect.Kind, effect)
					}
				}
			}
		})
	}
}

func TestParseRejectsNamedSourceDurationWithTrailingText(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"{T}: Gain control of target creature for as long as you control Merieke Ri Berit, then celebrate.",
		Context{CardName: "Merieke Ri Berit"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Exact {
		t.Fatalf("effects = %#v, want one inexact effect", effects)
	}
}

func TestParseComposedResolvingSyntax(t *testing.T) {
	t.Parallel()
	source := "Return up to two target cards with cycling from your graveyard to your hand, then draw a card."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	sentence := document.Abilities[0].Sentences[0]
	if len(sentence.Effects) != 2 || sentence.Effects[0].Kind != EffectReturn || sentence.Effects[1].Kind != EffectDraw {
		t.Fatalf("effects = %#v", sentence.Effects)
	}
	if sentence.Effects[0].FromZone != zone.Graveyard || sentence.Effects[0].ToZone != zone.Hand ||
		sentence.Effects[1].Amount.Value != 1 || !sentence.Effects[1].Amount.Known {
		t.Fatalf("typed effects = %#v", sentence.Effects)
	}
	if len(sentence.Targets) != 1 ||
		sentence.Targets[0].Cardinality != (TargetCardinalitySyntax{Min: 0, Max: 2}) ||
		sentence.Targets[0].Selection.Kind != SelectionCard ||
		sentence.Targets[0].Selection.Keyword != KeywordCycling {
		t.Fatalf("targets = %#v", sentence.Targets)
	}
}

func TestParseResolvingDurationDynamicAmountAndPayment(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Counter target spell unless its controller pays {2}{U}.\nTarget creature gets +2/+2 for each creature you control until end of turn.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	counterEffect := document.Abilities[0].Sentences[0].Effects[0]
	if counterEffect.Payment.Payer != EffectPaymentPayerTargetController ||
		!slices.Equal(counterEffect.Payment.ManaCost, cost.Mana{cost.O(2), cost.U}) {
		t.Fatalf("payment = %#v", counterEffect.Payment)
	}
	buff := document.Abilities[1].Sentences[0].Effects[0]
	if buff.Duration != EffectDurationUntilEndOfTurn ||
		buff.Amount.DynamicKind != EffectDynamicAmountCount ||
		buff.Amount.DynamicForm != EffectDynamicAmountFormForEach {
		t.Fatalf("buff = %#v", buff)
	}
}

func TestParseEventPlayerResolutionPayment(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever an opponent casts a spell, you may draw a card unless that player pays {1}.",
		Context{CardName: "Rhystic Study"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	effect := ability.Sentences[0].Effects[0]
	if !ability.Optional || !effect.Optional || !effect.Exact {
		t.Fatalf("optional/exact semantics = ability %v, effect optional %v exact %v", ability.Optional, effect.Optional, effect.Exact)
	}
	if effect.Payment.Payer != EffectPaymentPayerEventPlayer ||
		effect.Payment.Form != EffectPaymentFormUnless ||
		!slices.Equal(effect.Payment.ManaCost, cost.Mana{cost.O(1)}) {
		t.Fatalf("payment = %#v", effect.Payment)
	}
}

func TestParseEventPlayerMayPayFailureConsequence(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever an opponent draws a card, that player may pay {2}. If the player doesn't, you create a Treasure token.",
		Context{CardName: "Smothering Tithe"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if ability.Optional || len(ability.Sentences) != 2 ||
		ability.Sentences[0].PaymentPrelude == nil {
		t.Fatalf("payment sequence = %#v", ability)
	}
	effect := ability.Sentences[1].Effects[0]
	if effect.Payment.Form != EffectPaymentFormMayPayThenIfDoesNot ||
		effect.Payment.Payer != EffectPaymentPayerEventPlayer ||
		!slices.Equal(effect.Payment.ManaCost, cost.Mana{cost.O(2)}) ||
		effect.Optional || effect.Negated || !effect.Exact {
		t.Fatalf("consequence = %#v", effect)
	}
	if len(ability.ConditionClauses) != 1 ||
		ability.ConditionClauses[0].Predicate != ConditionPredicateEventPlayerDoesNotPay {
		t.Fatalf("conditions = %#v", ability.ConditionClauses)
	}
}

func TestParseControllerMayPaySuccessConsequence(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"At the beginning of your upkeep, you may pay {4}. If you do, untap this artifact.",
		Context{CardName: "Mana Vault"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if ability.Optional || len(ability.Sentences) != 2 ||
		ability.Sentences[0].PaymentPrelude == nil {
		t.Fatalf("payment sequence = %#v", ability)
	}
	effect := ability.Sentences[1].Effects[0]
	if effect.Payment.Form != EffectPaymentFormMayPayThenIfDo ||
		effect.Payment.Payer != EffectPaymentPayerController ||
		!slices.Equal(effect.Payment.ManaCost, cost.Mana{cost.O(4)}) ||
		effect.Optional || effect.Negated || !effect.Exact {
		t.Fatalf("consequence = %#v", effect)
	}
	if len(ability.ConditionClauses) != 1 ||
		ability.ConditionClauses[0].Predicate != ConditionPredicatePriorInstructionAccepted ||
		len(ability.ConditionBoundaries) != 1 ||
		ability.ConditionBoundaries[0].NodeID != effect.Payment.SuccessConditionNodeID {
		t.Fatalf("conditions = %#v", ability.ConditionClauses)
	}
}

func TestParseControllerMandatoryPayOrLoseGame(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"At the beginning of your next upkeep, pay {3}{U}{U}. If you don't, you lose the game.",
		Context{CardName: "Pact of Negation", InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if ability.Optional || len(ability.Sentences) != 2 ||
		ability.Sentences[0].PaymentPrelude == nil {
		t.Fatalf("payment sequence = %#v", ability)
	}
	effect := ability.Sentences[1].Effects[0]
	if effect.Kind != EffectLoseGame ||
		effect.Payment.Form != EffectPaymentFormMayPayThenIfDoesNot ||
		effect.Payment.Payer != EffectPaymentPayerController ||
		!slices.Equal(effect.Payment.ManaCost, cost.Mana{cost.O(3), cost.U, cost.U}) ||
		effect.Optional || effect.Negated || effect.HasUnrecognizedSibling || !effect.Exact {
		t.Fatalf("consequence = %#v", effect)
	}
}

func TestParseControllerMandatoryPayOrLoseGameFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		// "you may pay" is the optional form, not the mandatory pact wording.
		"At the beginning of your next upkeep, you may pay {3}. If you don't, you lose the game.",
		// A non-game-loss consequence must not adopt the pact payment fold.
		"At the beginning of your next upkeep, pay {3}. If you don't, sacrifice this creature.",
	} {
		t.Run(oracle, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(oracle, Context{CardName: "Unsafe Pact", InstantOrSorcery: true})
			for _, ability := range document.Abilities {
				for _, sentence := range ability.Sentences {
					for _, effect := range sentence.Effects {
						if effect.Kind == EffectLoseGame &&
							effect.Payment.Payer == EffectPaymentPayerController &&
							effect.Payment.Form == EffectPaymentFormMayPayThenIfDoesNot {
							t.Fatalf("unexpected mandatory pay-or-lose fold = %#v", effect)
						}
					}
				}
			}
		})
	}
}

func TestParseControllerMayPaySuccessConsequenceFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"At the beginning of your upkeep, you may pay 4 life. If you do, untap this artifact.",
		"At the beginning of your upkeep, you may sacrifice a creature. If you do, untap this artifact.",
		"At the beginning of your upkeep, you may pay {4}. If you don't, untap this artifact.",
		"At the beginning of your upkeep, a player may pay {4}. If they do, untap this artifact.",
	} {
		t.Run(oracle, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(oracle, Context{CardName: "Unsafe Vault"})
			for _, sentence := range document.Abilities[0].Sentences {
				if sentence.PaymentPrelude != nil {
					t.Fatalf("unexpected payment prelude = %#v", sentence.PaymentPrelude)
				}
				for _, effect := range sentence.Effects {
					if effect.Payment.Form == EffectPaymentFormMayPayThenIfDo {
						t.Fatalf("unexpected payment = %#v", effect.Payment)
					}
				}
			}
		})
	}
}

func TestParseEventPlayerMayPayFailureConsequenceRejectsOtherPaymentWording(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"Whenever an opponent draws a card, that player may pay 2 life. If the player doesn't, you create a Treasure token.",
		"Whenever an opponent draws a card, that player may sacrifice a creature. If the player doesn't, you create a Treasure token.",
		"Whenever an opponent draws a card, you may pay {2}. If you don't, you create a Treasure token.",
		"Whenever an opponent draws a card, target opponent may pay {2}. If the player doesn't, you create a Treasure token.",
	} {
		t.Run(oracle, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(oracle, Context{CardName: "Unsafe Tithe"})
			for _, sentence := range document.Abilities[0].Sentences {
				if sentence.PaymentPrelude != nil {
					t.Fatalf("unexpected payment prelude = %#v", sentence.PaymentPrelude)
				}
				for _, effect := range sentence.Effects {
					if effect.Payment.Form == EffectPaymentFormMayPayThenIfDoesNot {
						t.Fatalf("unexpected payment = %#v", effect.Payment)
					}
				}
			}
		})
	}
}

func TestParseEventPlayerResolutionPaymentRejectsOtherPaymentWording(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"Whenever an opponent casts a spell, you may draw a card unless that player pays 2 life.",
		"Whenever an opponent casts a spell, you may draw a card unless you pay {1}.",
		"Whenever an opponent casts a spell, you may draw a card unless that player's controller pays {1}.",
		"Whenever an opponent casts a spell, you may draw a card unless that player pays {1}, then discards a card.",
	} {
		t.Run(oracle, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(oracle, Context{CardName: "Tax Study"})
			effect := document.Abilities[0].Sentences[0].Effects[0]
			if effect.Payment.Payer != EffectPaymentPayerUnknown || len(effect.Payment.ManaCost) != 0 {
				t.Fatalf("payment = %#v, want unsupported", effect.Payment)
			}
		})
	}
}

func TestParseResolvingCreateForEachIterator(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"When this enchantment enters, for each Shrine you control, create a 1/1 red Monk creature token.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if effect.Kind != EffectCreate || !effect.Exact {
		t.Fatalf("effect = %#v", effect)
	}
	if effect.Amount.DynamicKind != EffectDynamicAmountCount ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
		effect.Amount.Multiplier != 1 {
		t.Fatalf("amount = %#v", effect.Amount)
	}
	if effect.Amount.Selection == nil ||
		len(effect.Amount.Selection.SubtypesAny) != 1 ||
		effect.Amount.Selection.SubtypesAny[0] != "Shrine" {
		t.Fatalf("for-each selection = %#v", effect.Amount.Selection)
	}
}

func TestParseControlledPermanentCounterPlacementReplacement(t *testing.T) {
	t.Parallel()
	doubling, diagnostics := Parse(
		"If an effect would put one or more counters on a permanent you control, it puts twice that many of those counters on that permanent instead.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("doubling diagnostics = %#v", diagnostics)
	}
	cond := doubling.Abilities[0].ConditionClauses[0]
	if cond.Predicate != ConditionPredicateCounterPlacementOnControlledPermanent || cond.Counter != ConditionCounterNone {
		t.Fatalf("doubling condition = %#v", cond)
	}

	additive, diagnostics := Parse(
		"If one or more +1/+1 counters would be put on a permanent you control, that many plus one +1/+1 counters are put on that permanent instead.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("additive diagnostics = %#v", diagnostics)
	}
	cond = additive.Abilities[0].ConditionClauses[0]
	if cond.Predicate != ConditionPredicateCounterPlacementOnControlledPermanent ||
		cond.Counter != ConditionCounterPlusOnePlusOne {
		t.Fatalf("additive condition = %#v", cond)
	}
}

func TestParseAdditiveCounterPlacementReplacement(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"If one or more +1/+1 counters would be put on a creature you control, that many plus one +1/+1 counters are put on it instead.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	replacement := effects[len(effects)-1].Replacement
	if replacement.Kind != EffectReplacementThatManyPlus || replacement.Amount != 1 {
		t.Fatalf("replacement = %#v", replacement)
	}
	if replacement.EachCounterKind {
		t.Fatalf("unexpected each-counter-kind modifier: %#v", replacement)
	}
}

func TestParseAnyCreatureCounterPlacementReplacement(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"If one or more +1/+1 counters would be put on a creature, twice that many +1/+1 counters are put on that creature instead.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	cond := document.Abilities[0].ConditionClauses[0]
	if cond.Predicate != ConditionPredicateCounterPlacementOnAnyCreature ||
		cond.Counter != ConditionCounterPlusOnePlusOne {
		t.Fatalf("condition = %#v", cond)
	}
}

func TestParsePlusAdditionalTokenReplacement(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"If you would create one or more Treasure tokens, instead create those tokens plus an additional Treasure token.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	cond := document.Abilities[0].ConditionClauses[0]
	if cond.Predicate != ConditionPredicateTokenCreationUnderController {
		t.Fatalf("condition = %#v", cond)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	replacement := effects[len(effects)-1].Replacement
	if replacement.Kind != EffectReplacementPlusAdditional || replacement.Amount != 1 {
		t.Fatalf("replacement = %#v", replacement)
	}
}

func TestParseDrawFromEmptyLibraryWinReplacement(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"If you would draw a card while your library has no cards in it, you win the game instead.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %d, want 1", len(effects))
	}
	if effects[0].Kind != EffectWinGame {
		t.Fatalf("effect kind = %v, want EffectWinGame", effects[0].Kind)
	}
	if effects[0].Replacement.Kind != EffectReplacementInstead {
		t.Fatalf("replacement kind = %v, want instead", effects[0].Replacement.Kind)
	}
	if !effects[0].Exact {
		t.Fatal("win-game replacement effect not marked exact")
	}
	clauses := document.Abilities[0].ConditionClauses
	found := false
	for _, clause := range clauses {
		if clause.Predicate == ConditionPredicateWouldDrawFromEmptyLibrary {
			found = true
		}
	}
	if !found {
		t.Fatalf("would-draw-from-empty condition not recognized: %#v", clauses)
	}
}

func TestParseDrawDoublingReplacement(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		text      string
		predicate ConditionPredicateKind
	}{
		{
			name:      "plain",
			text:      "If you would draw a card, draw two cards instead.",
			predicate: ConditionPredicateWouldDrawCard,
		},
		{
			name:      "draw-step exception",
			text:      "If you would draw a card except the first one you draw in each of your draw steps, draw two cards instead.",
			predicate: ConditionPredicateWouldDrawCardExceptFirstInDrawStep,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(tc.text, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %d, want 1", len(effects))
			}
			if effects[0].Kind != EffectDraw {
				t.Fatalf("effect kind = %v, want EffectDraw", effects[0].Kind)
			}
			if effects[0].Replacement.Kind != EffectReplacementInstead {
				t.Fatalf("replacement kind = %v, want instead", effects[0].Replacement.Kind)
			}
			if effects[0].Amount.Value != 2 {
				t.Fatalf("amount = %d, want 2", effects[0].Amount.Value)
			}
			found := false
			for _, clause := range document.Abilities[0].ConditionClauses {
				if clause.Predicate == tc.predicate {
					found = true
				}
			}
			if !found {
				t.Fatalf("condition %v not recognized: %#v", tc.predicate, document.Abilities[0].ConditionClauses)
			}
		})
	}
}

func TestParseDrawReplacementDig(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		text      string
		look      int
		take      int
		remainder DigRemainderKind
		singular  bool
	}{
		{
			name:      "graveyard rest",
			text:      "If you would draw a card, instead look at the top three cards of your library, then put one into your hand and the rest into your graveyard.",
			look:      3,
			take:      1,
			remainder: DigRemainderGraveyard,
		},
		{
			name:      "library bottom other with source",
			text:      "If you would draw a card, instead look at the top two cards of your library, then put one of them into your hand and the other on the bottom of your library.",
			look:      2,
			take:      1,
			remainder: DigRemainderLibraryBottom,
			singular:  true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(tc.text, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %d, want 1", len(effects))
			}
			effect := effects[0]
			if effect.Kind != EffectDig {
				t.Fatalf("effect kind = %v, want EffectDig", effect.Kind)
			}
			if effect.Replacement.Kind != EffectReplacementInstead {
				t.Fatalf("replacement kind = %v, want instead", effect.Replacement.Kind)
			}
			if effect.Amount.Value != tc.look {
				t.Fatalf("look = %d, want %d", effect.Amount.Value, tc.look)
			}
			if !effect.Dig.Put {
				t.Fatal("dig put flag not set")
			}
			if effect.Dig.Take != tc.take {
				t.Fatalf("take = %d, want %d", effect.Dig.Take, tc.take)
			}
			if effect.Dig.Remainder != tc.remainder {
				t.Fatalf("remainder = %v, want %v", effect.Dig.Remainder, tc.remainder)
			}
			if effect.Dig.Singular != tc.singular {
				t.Fatalf("singular = %v, want %v", effect.Dig.Singular, tc.singular)
			}
			found := false
			for _, clause := range document.Abilities[0].ConditionClauses {
				if clause.Predicate == ConditionPredicateWouldDrawCard {
					found = true
				}
			}
			if !found {
				t.Fatalf("would-draw condition not recognized: %#v", document.Abilities[0].ConditionClauses)
			}
		})
	}
}

func TestParseDrawReplacementDigRejectsEqualTake(t *testing.T) {
	t.Parallel()
	// A take count not fewer than the look count is not a dig; it must fail the
	// whole-sentence recognizer and fall back to general parsing.
	text := "If you would draw a card, instead look at the top two cards of your library, then put two into your hand and the rest into your graveyard."
	document, _ := Parse(text, Context{})
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) == 1 && effects[0].Kind == EffectDig && effects[0].Replacement.Kind == EffectReplacementInstead {
		t.Fatal("equal look/take wrongly recognized as a draw-replacement dig")
	}
}

func TestParseLifeGainReplacement(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		text   string
		kind   EffectReplacementKind
		amount int
	}{
		{
			name:   "double",
			text:   "If you would gain life, you gain twice that much life instead.",
			kind:   EffectReplacementTwiceThatMuch,
			amount: 0,
		},
		{
			name:   "plus",
			text:   "If you would gain life, you gain that much life plus 1 instead.",
			kind:   EffectReplacementThatMuchPlus,
			amount: 1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(tc.text, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %d, want 1", len(effects))
			}
			if effects[0].Kind != EffectGain {
				t.Fatalf("effect kind = %v, want EffectGain", effects[0].Kind)
			}
			if effects[0].Replacement.Kind != tc.kind {
				t.Fatalf("replacement kind = %v, want %v", effects[0].Replacement.Kind, tc.kind)
			}
			if effects[0].Replacement.Amount != tc.amount {
				t.Fatalf("replacement amount = %d, want %d", effects[0].Replacement.Amount, tc.amount)
			}
			found := false
			for _, clause := range document.Abilities[0].ConditionClauses {
				if clause.Predicate == ConditionPredicateControllerLifeGain {
					found = true
				}
			}
			if !found {
				t.Fatalf("life-gain condition not recognized: %#v", document.Abilities[0].ConditionClauses)
			}
		})
	}
}

func TestParseLifeLossReplacement(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		text      string
		kind      EffectReplacementKind
		amount    int
		predicate ConditionPredicateKind
	}{
		{
			name:      "opponent during your turn",
			text:      "If an opponent would lose life during your turn, they lose twice that much life instead.",
			kind:      EffectReplacementTwiceThatMuch,
			amount:    0,
			predicate: ConditionPredicateOpponentLifeLossDuringControllerTurn,
		},
		{
			name:      "opponent any time",
			text:      "If an opponent would lose life, they lose twice that much instead.",
			kind:      EffectReplacementTwiceThatMuch,
			amount:    0,
			predicate: ConditionPredicateOpponentLifeLoss,
		},
		{
			name:      "any player plus",
			text:      "If a player would lose life, they lose that much life plus 1 instead.",
			kind:      EffectReplacementThatMuchPlus,
			amount:    1,
			predicate: ConditionPredicateAnyPlayerLifeLoss,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(tc.text, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %d, want 1", len(effects))
			}
			if effects[0].Kind != EffectLose {
				t.Fatalf("effect kind = %v, want EffectLose", effects[0].Kind)
			}
			if effects[0].Replacement.Kind != tc.kind {
				t.Fatalf("replacement kind = %v, want %v", effects[0].Replacement.Kind, tc.kind)
			}
			if effects[0].Replacement.Amount != tc.amount {
				t.Fatalf("replacement amount = %d, want %d", effects[0].Replacement.Amount, tc.amount)
			}
			found := false
			for _, clause := range document.Abilities[0].ConditionClauses {
				if clause.Predicate == tc.predicate {
					found = true
				}
			}
			if !found {
				t.Fatalf("life-loss condition not recognized: %#v", document.Abilities[0].ConditionClauses)
			}
		})
	}
}

func TestParseOneOfEachTokenReplacement(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"If you would create a Clue, Food, or Treasure token, instead create one of each.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 2 {
		t.Fatalf("effects = %d, want 2", len(effects))
	}
	if got := effects[0].Selection.SubtypesAny; len(got) != 3 ||
		string(got[0]) != "Clue" || string(got[1]) != "Food" || string(got[2]) != "Treasure" {
		t.Fatalf("trigger subtypes = %#v, want Clue/Food/Treasure", got)
	}
	if got := effects[1].Replacement.Kind; got != EffectReplacementOneOfEach {
		t.Fatalf("replacement kind = %v, want one-of-each", got)
	}
}

func TestParseResolvingReplacementAndManaMeaning(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.\n"+
			"Add {G}, {W}, or {U}.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	replacement := document.Abilities[0].Sentences[0].Effects[1].Replacement
	if replacement.Kind != EffectReplacementTwiceThatMany || replacement.Span.Start == replacement.Span.End {
		t.Fatalf("replacement = %#v", replacement)
	}
	if got := document.Abilities[0].Sentences[0].Effects[0].Replacement.Kind; got != EffectReplacementNone {
		t.Fatalf("replaced event modifier = %v, want none", got)
	}
	manaSyntax := document.Abilities[1].Sentences[0].Effects[0].Mana
	if !manaSyntax.Choice || manaSyntax.AnyColor || !slices.Equal(manaSyntax.Symbols, []string{"{G}", "{W}", "{U}"}) {
		t.Fatalf("mana = %#v", manaSyntax)
	}

	nearMiss, _ := Parse(
		"If an effect would create one or more tokens under your control, it creates twice those tokens instead.\n"+
			"Add {G} and {W}.",
		Context{InstantOrSorcery: true},
	)
	if got := nearMiss.Abilities[0].Sentences[0].Effects[1].Replacement.Kind; got != EffectReplacementInstead {
		t.Fatalf("near-miss replacement = %v, want plain instead", got)
	}
	if got := nearMiss.Abilities[1].Sentences[0].Effects[0].Mana; len(got.Symbols) != 0 || got.AnyColor {
		t.Fatalf("near-miss mana = %#v, want unknown", got)
	}

	modified, _ := Parse(
		"If an effect would create one or more tokens under your control, it creates twice that many tapped tokens instead.",
		Context{InstantOrSorcery: true},
	)
	if got := modified.Abilities[0].Sentences[0].Effects[1].Replacement.Kind; got != EffectReplacementInstead {
		t.Fatalf("modified replacement = %v, want plain instead", got)
	}
	treasure, _ := Parse(
		"If an effect would create one or more tokens under your control, it creates twice that many Treasure tokens instead.",
		Context{InstantOrSorcery: true},
	)
	if got := treasure.Abilities[0].Sentences[0].Effects[1].Replacement.Kind; got != EffectReplacementInstead {
		t.Fatalf("Treasure replacement = %v, want plain instead", got)
	}
}

func TestParseResolvingEffectCompositionOwnership(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Gain control of target creature, then it gains haste until end of turn.\n"+
			"They discard a card, then draw a card.\n"+
			"Add {R}, then draw a card.\n"+
			"Put a charge counter on target artifact with mana value X.\n"+
			"Untap target creature and gain control of it until end of turn.\n"+
			"Tap target creature that entered this turn.\n"+
			"Tap up to X target creatures.\n"+
			"Tap target creature named Bob.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	control := document.Abilities[0].Sentences[0].Effects
	if len(control) != 2 ||
		control[0].Duration != EffectDurationNone ||
		control[1].Duration != EffectDurationUntilEndOfTurn ||
		len(control[0].Targets) != 1 || len(control[1].References) != 1 {
		t.Fatalf("control effects = %#v", control)
	}

	discardDraw := document.Abilities[1].Sentences[0].Effects
	if len(discardDraw) != 2 ||
		discardDraw[0].Context != EffectContextEventPlayer ||
		discardDraw[1].Context != EffectContextPriorSubject {
		t.Fatalf("discard/draw contexts = %#v", discardDraw)
	}

	manaDraw := document.Abilities[2].Sentences[0].Effects
	if len(manaDraw) != 2 || !slices.Equal(manaDraw[0].Mana.Symbols, []string{"{R}"}) {
		t.Fatalf("mana/draw effects = %#v", manaDraw)
	}

	counterEffect := document.Abilities[3].Sentences[0].Effects[0]
	if !counterEffect.Amount.Known || counterEffect.Amount.Value != 1 ||
		len(counterEffect.Targets) != 1 || counterEffect.Targets[0].Selection.Kind != SelectionUnknown {
		t.Fatalf("counter effect = %#v", counterEffect)
	}

	untapControl := document.Abilities[4].Sentences[0].Effects
	if len(untapControl) != 2 ||
		untapControl[0].Duration != EffectDurationNone ||
		untapControl[1].Duration != EffectDurationUntilEndOfTurn {
		t.Fatalf("untap/control durations = %#v", untapControl)
	}
	if target := document.Abilities[5].Sentences[0].Targets[0]; target.Selection.Kind != SelectionUnknown {
		t.Fatalf("relative-clause target = %#v, want unknown selection", target)
	}
	if target := document.Abilities[6].Sentences[0].Targets[0]; target.Cardinality != (TargetCardinalitySyntax{}) {
		t.Fatalf("variable target cardinality = %#v, want unknown", target.Cardinality)
	}
	if target := document.Abilities[7].Sentences[0].Targets[0]; target.Selection.Kind != SelectionUnknown {
		t.Fatalf("unrecognized target qualifier = %#v, want unknown selection", target)
	}
}

func TestParseResolvingSyntaxFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"The counter remains on it.",
		"It was cast this turn.",
		"Double strike is useful.",
		"{1}: Draw a card. Activate only any time you could cast a sorcery.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, sentence := range document.Abilities[0].Sentences {
				if len(sentence.Effects) != 0 && source != "{1}: Draw a card. Activate only any time you could cast a sorcery." {
					t.Fatalf("effects = %#v, want none", sentence.Effects)
				}
				if source == "{1}: Draw a card. Activate only any time you could cast a sorcery." &&
					len(sentence.Effects) > 0 && sentence.Effects[0].Kind != EffectDraw {
					t.Fatalf("activation restriction emitted effect: %#v", sentence.Effects)
				}
			}
		})
	}

	document, _ := Parse("Draw a card for each creatures you control.", Context{InstantOrSorcery: true})
	amount := document.Abilities[0].Sentences[0].Effects[0].Amount
	if amount.Known || amount.DynamicKind != EffectDynamicAmountNone {
		t.Fatalf("ambiguous amount = %#v, want unknown", amount)
	}

	document, _ = Parse("Draw a card, 5 mill.", Context{InstantOrSorcery: true})
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 2 || effects[1].Context != EffectContextUnknown {
		t.Fatalf("non-word subject effects = %#v, want unknown context", effects)
	}
}

func TestParseEntersColorChoiceSyntax(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source           string
		wantColorChoice  bool
		wantEntersTapped bool
		wantExclude      mana.Color
	}{
		{"As this artifact enters, choose a color.", true, false, ""},
		{"This land enters tapped. As it enters, choose a color.", true, true, ""},
		// Forbidden-color variants now record the excluded color.
		{"As this land enters, choose a color other than white.", true, false, mana.W},
		{"This land enters tapped. As it enters, choose a color other than green.", true, true, mana.G},
		// Non-color named choices stay fail-closed.
		{"As this enchantment enters, choose Khans or Dragons.", false, false, ""},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{})
			var gotColorChoice, gotEntersTapped bool
			var gotExclude mana.Color
			for _, ability := range document.Abilities {
				for _, sentence := range ability.Sentences {
					for _, effect := range sentence.Effects {
						if effect.EntersColorChoice {
							gotColorChoice = true
							gotExclude = effect.EntersColorChoiceExclude
						}
						if effect.EntersTappedSelf {
							gotEntersTapped = true
						}
					}
				}
			}
			if gotColorChoice != test.wantColorChoice {
				t.Fatalf("EntersColorChoice = %v, want %v", gotColorChoice, test.wantColorChoice)
			}
			if gotEntersTapped != test.wantEntersTapped {
				t.Fatalf("EntersTappedSelf = %v, want %v", gotEntersTapped, test.wantEntersTapped)
			}
			if gotExclude != test.wantExclude {
				t.Fatalf("EntersColorChoiceExclude = %q, want %q", gotExclude, test.wantExclude)
			}
		})
	}
}

func TestParseEntersTypeChoiceSyntax(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source         string
		wantTypeChoice bool
	}{
		{"As this creature enters, choose a creature type.", true},
		{"As this artifact enters, choose a creature type.", true},
		// A color choice is not a type choice.
		{"As this artifact enters, choose a color.", false},
		// Other named choices stay fail-closed.
		{"As this enchantment enters, choose Khans or Dragons.", false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{})
			var gotTypeChoice bool
			for _, ability := range document.Abilities {
				for _, sentence := range ability.Sentences {
					for _, effect := range sentence.Effects {
						if effect.EntersTypeChoice {
							gotTypeChoice = true
						}
					}
				}
			}
			if gotTypeChoice != test.wantTypeChoice {
				t.Fatalf("EntersTypeChoice = %v, want %v", gotTypeChoice, test.wantTypeChoice)
			}
		})
	}
}

func TestParseMultiTargetExileExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		exact  bool
		min    int
		max    int
	}{
		{"Exile up to one target permanent.", true, 0, 1},
		{"Exile up to one target creature.", true, 0, 1},
		{"Exile two target artifacts.", true, 2, 2},
		{"Exile two target permanents.", true, 2, 2},
		{"Exile up to two target creatures.", true, 0, 2},
		{"Exile up to two target artifacts.", true, 0, 2},
		{"Exile up to three target enchantments.", true, 0, 3},
		{"Exile up to two target creatures you control.", true, 0, 2},
		{"Exile two target creatures an opponent controls.", true, 2, 2},
		// Single-target wording keeps its existing exact cardinality.
		{"Exile target creature.", true, 1, 1},
		// Fail-closed: a graveyard zone is not a permanent target.
		{"Exile up to two target cards from a single graveyard.", false, 0, 2},
		// Fail-closed: subtype and tapped qualifiers are not reconstructed here.
		{"Exile up to two target Goblin creatures.", false, 0, 2},
		{"Exile two target tapped creatures.", false, 2, 2},
		// Fail-closed: the unbounded "any number of" shape has no cardinal word.
		{"Exile any number of target creatures.", false, 0, 99},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || len(effects[0].Targets) != 1 {
				t.Fatalf("effects = %#v, want one effect with one target", effects)
			}
			target := effects[0].Targets[0]
			if target.Cardinality.Min != test.min || target.Cardinality.Max != test.max {
				t.Fatalf("cardinality = {%d,%d}, want {%d,%d}", target.Cardinality.Min, target.Cardinality.Max, test.min, test.max)
			}
			if target.Exact != test.exact {
				t.Fatalf("target Exact = %v, want %v", target.Exact, test.exact)
			}
			// An exact target makes the whole exile effect byte-exact.
			if test.exact && !effects[0].Exact {
				t.Fatal("effect Exact = false, want true for an exact target")
			}
		})
	}
}

// TestParseCommanderIdentityManaSyntax covers the Command Tower / Arcane Signet
// wording "Add one mana of any color in your commander's color identity." The
// body is recognized as an exact add-mana effect with CommanderIdentity set and
// LegacyBodyExact true, while the shorter "any color" wording stays AnyColor.
func TestParseCommanderIdentityManaSyntax(t *testing.T) {
	t.Parallel()
	document, _ := Parse("{T}: Add one mana of any color in your commander's color identity.", Context{})
	var found bool
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			for _, effect := range sentence.Effects {
				if effect.Mana.CommanderIdentity {
					found = true
					if effect.Mana.AnyColor {
						t.Fatal("commander-identity mana must not also set AnyColor")
					}
					if !effect.Mana.LegacyBodyExact {
						t.Fatal("commander-identity mana body must be LegacyBodyExact")
					}
				}
			}
		}
	}
	if !found {
		t.Fatal("expected Mana.CommanderIdentity for commander color identity body")
	}

	plain, _ := Parse("{T}: Add one mana of any color.", Context{})
	for _, ability := range plain.Abilities {
		for _, sentence := range ability.Sentences {
			for _, effect := range sentence.Effects {
				if effect.Mana.CommanderIdentity {
					t.Fatal("plain any-color body must not set CommanderIdentity")
				}
			}
		}
	}
}

// TestParseLandsProduceManaSyntax covers the Exotic Orchard / Reflecting Pool /
// Fellwar Stone wording "Add one mana of any color/type that a land <scope>
// could produce." The body is recognized as an exact add-mana effect with
// LandsProduce set, the correct scope, the AnyType flag for "any type" wording,
// and LegacyBodyExact true. Variants outside the exact shape stay unrecognized.
func TestParseLandsProduceManaSyntax(t *testing.T) {
	t.Parallel()
	cases := []struct {
		text    string
		scope   ManaLandsProduceScope
		anyType bool
	}{
		{"{T}: Add one mana of any color that a land an opponent controls could produce.", ManaLandsProduceOpponent, false},
		{"{T}: Add one mana of any color that a land you control could produce.", ManaLandsProduceYou, false},
		{"{T}: Add one mana of any type that a land you control could produce.", ManaLandsProduceYou, true},
		{"{T}: Add one mana of any type that a land an opponent controls could produce.", ManaLandsProduceOpponent, true},
	}
	for _, tc := range cases {
		t.Run(tc.text, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(tc.text, Context{})
			var found bool
			for _, ability := range document.Abilities {
				for _, sentence := range ability.Sentences {
					for _, effect := range sentence.Effects {
						if !effect.Mana.LandsProduce {
							continue
						}
						found = true
						if effect.Mana.LandsProduceScope != tc.scope {
							t.Fatalf("scope = %d, want %d", effect.Mana.LandsProduceScope, tc.scope)
						}
						if effect.Mana.LandsProduceAnyType != tc.anyType {
							t.Fatalf("anyType = %v, want %v", effect.Mana.LandsProduceAnyType, tc.anyType)
						}
						if effect.Mana.AnyColor || effect.Mana.CommanderIdentity {
							t.Fatal("lands-produce mana must not set AnyColor or CommanderIdentity")
						}
						if !effect.Mana.LegacyBodyExact {
							t.Fatal("lands-produce mana body must be LegacyBodyExact")
						}
					}
				}
			}
			if !found {
				t.Fatalf("expected Mana.LandsProduce for %q", tc.text)
			}
		})
	}
}

// TestParseLandsProduceManaFailsClosed asserts the lands-produce recognition does
// not over-match related but unmodeled "could produce" wordings (basic-land,
// Gate, sacrificed-land, plural-quantity, and non-land variants), which must stay
// unrecognized so they fail closed downstream.
func TestParseLandsProduceManaFailsClosed(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"{T}: Add one mana of any color that a basic land you control could produce.",
		"{T}: Add one mana of any color that a Gate you control could produce.",
		"{T}: Add one mana of any type the sacrificed land could produce.",
		"{T}: Add one mana of any type that land could produce.",
		"{T}: Add two mana of any color that a land an opponent controls could produce.",
		"{T}: Add one mana of any color that a creature you control could produce.",
		"{T}: Add one mana of any color that a land a player controls could produce.",
	} {
		document, _ := Parse(text, Context{})
		for _, ability := range document.Abilities {
			for _, sentence := range ability.Sentences {
				for _, effect := range sentence.Effects {
					if effect.Mana.LandsProduce {
						t.Fatalf("unmodeled body wrongly recognized as lands-produce: %q", text)
					}
				}
			}
		}
	}
}

// TestParseTriggerLandProducedManaSyntax verifies the "add one mana of any type
// that land produced" mana-doubler body (Mirari's Wake, Zendikar Resurgent)
// parses to Mana.TriggerLandProducedType without setting the battlefield
// lands-produce or any-color flags.
func TestParseTriggerLandProducedManaSyntax(t *testing.T) {
	t.Parallel()
	document, _ := Parse("Whenever you tap a land for mana, add one mana of any type that land produced.", Context{})
	var found bool
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			for _, effect := range sentence.Effects {
				if !effect.Mana.TriggerLandProducedType {
					continue
				}
				found = true
				if effect.Mana.LandsProduce || effect.Mana.AnyColor || effect.Mana.CommanderIdentity {
					t.Fatal("trigger-land-produced mana must not set LandsProduce, AnyColor, or CommanderIdentity")
				}
			}
		}
	}
	if !found {
		t.Fatal("expected Mana.TriggerLandProducedType for the mana-doubler body")
	}
}

// TestParseTriggerLandProducedManaFailsClosed asserts the mana-doubler
// recognition does not over-match the related battlefield "that a land you
// control could produce" body, which must stay LandsProduce.
func TestParseTriggerLandProducedManaFailsClosed(t *testing.T) {
	t.Parallel()
	document, _ := Parse("{T}: Add one mana of any type that a land you control could produce.", Context{})
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			for _, effect := range sentence.Effects {
				if effect.Mana.TriggerLandProducedType {
					t.Fatal("battlefield lands-produce body wrongly recognized as trigger-land-produced")
				}
			}
		}
	}
}

func TestParseChosenColorManaSyntax(t *testing.T) {
	t.Parallel()
	document, _ := Parse("{T}: Add one mana of the chosen color.", Context{})
	var found bool
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			for _, effect := range sentence.Effects {
				if effect.Mana.ChosenColor {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatal("expected Mana.ChosenColor for \"Add one mana of the chosen color.\"")
	}
}

func TestParseFixedOrChosenColorManaSyntax(t *testing.T) {
	t.Parallel()
	// The Gate/Thriving cycle prints "{T}: Add {W} or one mana of the chosen
	// color." — a fixed color alternative to the entry-chosen color.
	document, _ := Parse("{T}: Add {W} or one mana of the chosen color.", Context{})
	var found bool
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			for _, effect := range sentence.Effects {
				if !effect.Mana.ChosenColor {
					continue
				}
				found = true
				if !effect.Mana.ChosenColorFixedKnown || effect.Mana.ChosenColorFixed != mana.W {
					t.Fatalf("fixed color = %q known=%v, want white known", effect.Mana.ChosenColorFixed, effect.Mana.ChosenColorFixedKnown)
				}
			}
		}
	}
	if !found {
		t.Fatal("expected Mana.ChosenColor for the composite fixed-or-chosen body")
	}
}

// TestParseDualRecipientGroupDamage covers the "deals N damage to each X and
// each Y" board-sweep wording. A recognized pair captures both recipient groups
// separately so lowering can damage each in Oracle order, and the effect is
// exact only when both halves and the fixed amount reconstruct byte-for-byte.
// Single recipients, multi-color filters, and leading-player compounds stay off
// the dual path and fail closed.
func TestParseDualRecipientGroupDamage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source   string
		cardName string
		wantPair []SelectionKind
		exact    bool
	}{
		{
			source:   "Famine deals 3 damage to each creature and each player.",
			cardName: "Famine",
			wantPair: []SelectionKind{SelectionCreature, SelectionPlayer},
			exact:    true,
		},
		{
			source:   "Star of Extinction deals 20 damage to each creature and each planeswalker.",
			cardName: "Star of Extinction",
			wantPair: []SelectionKind{SelectionCreature, SelectionPlaneswalker},
			exact:    true,
		},
		{
			source:   "Test Bolt deals 1 damage to each creature.",
			cardName: "Test Bolt",
			wantPair: nil,
			exact:    true,
		},
		{
			source:   "Test Bolt deals 1 damage to each white and blue creature.",
			cardName: "Test Bolt",
			wantPair: nil,
			exact:    false,
		},
		{
			source:   "Test Bolt deals 3 damage to you and each creature you control.",
			cardName: "Test Bolt",
			wantPair: nil,
			exact:    false,
		},
		{
			source:   "Test Bolt deals X damage to each creature and each player.",
			cardName: "Test Bolt",
			wantPair: []SelectionKind{SelectionCreature, SelectionPlayer},
			exact:    true,
		},
		{
			source:   "Target creature you control deals damage equal to its power to each other creature and each opponent.",
			cardName: "Test Ignition",
			wantPair: []SelectionKind{SelectionCreature, SelectionOpponent},
			exact:    true,
		},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true, CardName: test.cardName})
			effect := document.Abilities[0].Sentences[0].Effects[0]
			gotKinds := make([]SelectionKind, 0, len(effect.DamageRecipientPair))
			for _, half := range effect.DamageRecipientPair {
				gotKinds = append(gotKinds, half.Kind)
			}
			if !slices.Equal(gotKinds, test.wantPair) {
				t.Fatalf("recipient pair kinds = %#v, want %#v", gotKinds, test.wantPair)
			}
			if effect.Exact != test.exact {
				t.Fatalf("exact = %v, want %v", effect.Exact, test.exact)
			}
		})
	}
}

// TestParseLeadingConditionPossessionNotKeywordGrant verifies that a player
// possession verb ("you have", "an opponent has") inside a leading "As long as
// ..." condition clause is not misclassified as a keyword-grant effect. The
// possession verb belongs to the condition, so the sentence must expose only
// its real characteristic-changing effects.
func TestParseLeadingConditionPossessionNotKeywordGrant(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source    string
		wantKinds []EffectKind
	}{
		{
			source:    "As long as you have 30 or more life, this creature gets +5/+5.",
			wantKinds: []EffectKind{EffectModifyPT},
		},
		{
			source:    "As long as you have seven or more cards in hand, this creature gets +2/+1 and has first strike.",
			wantKinds: []EffectKind{EffectModifyPT},
		},
		{
			source:    "As long as you have no cards in hand, this creature has double strike.",
			wantKinds: []EffectKind{EffectGrantKeyword},
		},
		{
			source:    "As long as an opponent has 10 or less life, this creature gets +2/+1.",
			wantKinds: []EffectKind{EffectModifyPT},
		},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{})
			effects := document.Abilities[0].Sentences[0].Effects
			gotKinds := make([]EffectKind, 0, len(effects))
			for _, effect := range effects {
				gotKinds = append(gotKinds, effect.Kind)
			}
			if !slices.Equal(gotKinds, test.wantKinds) {
				t.Fatalf("effect kinds = %#v, want %#v", gotKinds, test.wantKinds)
			}
		})
	}
}

// TestParseObjectCharacteristicLifeAmountExactness verifies the life-rider amount
// forms "equal to its power" and "equal to its toughness" parse to their dynamic
// kinds and reconstruct exactly, while an unmodeled characteristic ("its mana
// value") does not produce a power/toughness dynamic kind (regression guard for
// the toughness sibling added alongside the existing power form).
func TestParseObjectCharacteristicLifeAmountExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source    string
		wantKind  EffectDynamicAmountKind
		exactGain bool
	}{
		{"Exile target creature. Its controller gains life equal to its power.", EffectDynamicAmountSourcePower, true},
		{"Exile target creature. Its controller gains life equal to its toughness.", EffectDynamicAmountSourceToughness, true},
		{"Destroy target creature. You gain life equal to its toughness.", EffectDynamicAmountSourceToughness, true},
		// "its mana value" is not a power/toughness characteristic for this shape.
		{"Exile target creature. Its controller gains life equal to its mana value.", EffectDynamicAmountSourcePower, false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			var gain *EffectSyntax
			for si := range document.Abilities[0].Sentences {
				sentence := &document.Abilities[0].Sentences[si]
				for ei := range sentence.Effects {
					if sentence.Effects[ei].Kind == EffectGain {
						gain = &sentence.Effects[ei]
					}
				}
			}
			if gain == nil {
				t.Fatalf("no gain effect parsed from %q", test.source)
			}
			isPowerOrToughness := gain.Amount.DynamicKind == EffectDynamicAmountSourcePower ||
				gain.Amount.DynamicKind == EffectDynamicAmountSourceToughness
			if test.exactGain {
				if gain.Amount.DynamicKind != test.wantKind {
					t.Fatalf("gain dynamic kind = %v, want %v", gain.Amount.DynamicKind, test.wantKind)
				}
				if (gain.Amount.ReferenceSpan == shared.Span{}) {
					t.Fatal("gain amount has no reference span, want the \"its\" referent span")
				}
			} else if isPowerOrToughness {
				t.Fatalf("gain dynamic kind = %v, want a non-power/toughness kind", gain.Amount.DynamicKind)
			}
		})
	}
}

// TestParseManaValueLifeAmountExactness verifies the mana-value life-rider amount
// forms "equal to its mana value" and "equal to that <object>'s mana value" parse
// to the mana-value dynamic kind, carry the referent span, and round-trip exactly.
// This is the Feed the Swarm sibling of the power/toughness characteristic riders.
func TestParseManaValueLifeAmountExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source   string
		wantKind EffectDynamicAmountKind
		exact    bool
	}{
		{"Destroy target permanent. You lose life equal to its mana value.", EffectDynamicAmountSourceManaValue, true},
		{"Destroy target artifact or enchantment. You gain life equal to that permanent's mana value.", EffectDynamicAmountSourceManaValue, true},
		{"Destroy target artifact. You gain life equal to its mana value.", EffectDynamicAmountSourceManaValue, true},
		{"Put target creature card from a graveyard onto the battlefield under your control. You lose life equal to that card's mana value.", EffectDynamicAmountSourceManaValue, true},
		// A bare numeric amount must not be mistaken for the mana-value characteristic.
		{"Destroy target permanent. You lose 2 life.", EffectDynamicAmountKind(""), false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			var life *EffectSyntax
			for si := range document.Abilities[0].Sentences {
				sentence := &document.Abilities[0].Sentences[si]
				for ei := range sentence.Effects {
					if sentence.Effects[ei].Kind == EffectGain || sentence.Effects[ei].Kind == EffectLose {
						life = &sentence.Effects[ei]
					}
				}
			}
			if life == nil {
				t.Fatalf("no gain/lose effect parsed from %q", test.source)
			}
			if test.exact {
				if life.Amount.DynamicKind != test.wantKind {
					t.Fatalf("amount dynamic kind = %v, want %v", life.Amount.DynamicKind, test.wantKind)
				}
				if (life.Amount.ReferenceSpan == shared.Span{}) {
					t.Fatal("mana-value amount has no reference span, want the referent span")
				}
			} else if life.Amount.DynamicKind == EffectDynamicAmountSourceManaValue {
				t.Fatalf("amount dynamic kind = %v, want a non-mana-value kind", life.Amount.DynamicKind)
			}
		})
	}
}

func TestParseGroupEntersTappedEffect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		source    string
		recognize bool
		scope     EntersTappedGroupControllerScope
		cardTypes []types.Card
	}{
		{
			name:      "opponent creatures",
			source:    "Creatures your opponents control enter tapped.",
			recognize: true,
			scope:     EntersTappedGroupControllerOpponents,
			cardTypes: []types.Card{types.Creature},
		},
		{
			name:      "opponent creatures battlefield wording",
			source:    "Creatures your opponents control enter the battlefield tapped.",
			recognize: true,
			scope:     EntersTappedGroupControllerOpponents,
			cardTypes: []types.Card{types.Creature},
		},
		{
			name:      "multi-type opponent",
			source:    "Artifacts, creatures, and lands your opponents control enter the battlefield tapped.",
			recognize: true,
			scope:     EntersTappedGroupControllerOpponents,
			cardTypes: []types.Card{types.Artifact, types.Creature, types.Land},
		},
		{
			name:      "all permanents",
			source:    "Permanents enter the battlefield tapped.",
			recognize: true,
			scope:     EntersTappedGroupControllerEach,
			cardTypes: nil,
		},
		{
			name:      "you control",
			source:    "Lands you control enter tapped.",
			recognize: true,
			scope:     EntersTappedGroupControllerYou,
			cardTypes: []types.Card{types.Land},
		},
		{
			// Self enters-tapped must keep flowing to the self path.
			name:      "self form not group",
			source:    "This land enters tapped.",
			recognize: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{})
			ability := document.Abilities[0]
			var group *EffectSyntax
			for i := range ability.Sentences {
				for j := range ability.Sentences[i].Effects {
					if ability.Sentences[i].Effects[j].EntersTappedGroup {
						group = &ability.Sentences[i].Effects[j]
					}
				}
			}
			if !test.recognize {
				if group != nil {
					t.Fatalf("source %q unexpectedly recognized as group enters-tapped", test.source)
				}
				return
			}
			if group == nil {
				t.Fatalf("source %q not recognized as group enters-tapped", test.source)
			}
			if ability.Kind != AbilityReplacement {
				t.Fatalf("ability kind = %v, want AbilityReplacement", ability.Kind)
			}
			if !group.Exact {
				t.Fatal("group enters-tapped effect is not exact")
			}
			if group.EntersTappedGroupScope != test.scope {
				t.Fatalf("scope = %v, want %v", group.EntersTappedGroupScope, test.scope)
			}
			if !slices.Equal(group.EntersTappedGroupTypes, test.cardTypes) {
				t.Fatalf("types = %v, want %v", group.EntersTappedGroupTypes, test.cardTypes)
			}
			if len(group.Targets) != 0 || len(group.References) != 0 {
				t.Fatalf("group effect has phantom targets/references: targets=%d references=%d", len(group.Targets), len(group.References))
			}
		})
	}
}

func TestParseGreatestDiscardedThisWayDrawAmount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		kind   EffectDynamicAmountKind
	}{
		{"Draw cards equal to the greatest number of cards a player discarded this way.", EffectDynamicAmountGreatestDiscardedThisWay},
		// Fixed draw stays non-dynamic (regression guard).
		{"Draw two cards.", EffectDynamicAmountNone},
		// A near-miss wording fails closed.
		{"Draw cards equal to the number of cards a player discarded this way.", EffectDynamicAmountNone},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if got := effects[0].Amount.DynamicKind; got != test.kind {
				t.Fatalf("draw dynamic kind = %v, want %v", got, test.kind)
			}
		})
	}
}

// TestParseControllerOptionalNonManaCostThenSearch verifies that a non-mana
// optional cost paired with a multi-effect search consequence ("you may
// sacrifice a land. If you do, search ...") folds into a controller payment: the
// payment sentence becomes a prelude carrying the parsed additional cost, the
// ability is no longer flagged optional, and the consequence's first effect
// carries the payment while its sibling effects remain intact for downstream
// search merging. A single-effect consequence is intentionally left for the
// ordered optional path and does not fold.
func TestParseControllerOptionalNonManaCostThenSearch(t *testing.T) {
	t.Parallel()
	const source = "When this creature enters, you may sacrifice a land. If you do, search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle."
	document, _ := Parse(source, Context{})
	ability := document.Abilities[0]
	if ability.Optional {
		t.Fatal("ability Optional = true, want false after the payment fold")
	}
	if len(ability.Sentences) != 2 {
		t.Fatalf("sentences = %d, want 2", len(ability.Sentences))
	}
	prelude := ability.Sentences[0].PaymentPrelude
	if prelude == nil {
		t.Fatal("payment sentence PaymentPrelude = nil, want the folded payment")
	}
	if prelude.Form != EffectPaymentFormMayPayThenIfDo ||
		prelude.Payer != EffectPaymentPayerController ||
		prelude.AdditionalCost == nil {
		t.Fatalf("prelude = %+v, want a controller additional-cost payment", prelude)
	}
	if len(prelude.AdditionalCost.Components) != 1 ||
		prelude.AdditionalCost.Components[0].Kind != CostComponentSacrifice {
		t.Fatalf("prelude cost = %+v, want one sacrifice component", prelude.AdditionalCost)
	}
	consequence := ability.Sentences[1].Effects
	if len(consequence) != 3 || consequence[0].Kind != EffectSearch {
		t.Fatalf("consequence effects = %#v, want search/put/shuffle", consequence)
	}
	if consequence[0].Payment.Form != EffectPaymentFormMayPayThenIfDo {
		t.Fatalf("search effect Payment.Form = %q, want the folded payment", consequence[0].Payment.Form)
	}
	if !consequence[0].Exact {
		t.Fatal("search effect Exact = false, want true after folding")
	}
}

// TestParseControllerOptionalNonManaCostSingleEffectDoesNotFold confirms a
// single-effect consequence with a non-mana optional cost is left unfolded
// (optional) so the ordered optional-effect path lowers it unchanged.
func TestParseControllerOptionalNonManaCostSingleEffectDoesNotFold(t *testing.T) {
	t.Parallel()
	const source = "When this creature enters, you may sacrifice a land. If you do, draw a card."
	document, _ := Parse(source, Context{})
	ability := document.Abilities[0]
	if !ability.Optional {
		t.Fatal("ability Optional = false, want true (single-effect consequence stays on the ordered path)")
	}
	if ability.Sentences[0].PaymentPrelude != nil {
		t.Fatal("single-effect consequence must not fold into a payment prelude")
	}
}

// TestParseDoublePowerToughnessObject verifies the power/toughness doubling
// object "double the power[ and toughness] of <group>" captures the doubled
// characteristics and resolves the affected group as a static subject.
func TestParseDoublePowerToughnessObject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source          string
		doublePower     bool
		doubleToughness bool
		subject         EffectStaticSubjectKind
	}{
		{
			"Double the power and toughness of each creature you control until end of turn.",
			true, true, EffectStaticSubjectControlledCreatures,
		},
		{
			"Double the power of each creature you control until end of turn.",
			true, false, EffectStaticSubjectControlledCreatures,
		},
		{
			"Double the toughness of creatures you control until end of turn.",
			false, true, EffectStaticSubjectControlledCreatures,
		},
		{
			"Double the power and toughness of each creature until end of turn.",
			true, true, EffectStaticSubjectAllCreatures,
		},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			effect := effects[0]
			if effect.Kind != EffectDouble {
				t.Fatalf("effect Kind = %v, want EffectDouble", effect.Kind)
			}
			if effect.DoublePower != test.doublePower || effect.DoubleToughness != test.doubleToughness {
				t.Fatalf("doublePower=%v doubleToughness=%v, want %v %v",
					effect.DoublePower, effect.DoubleToughness, test.doublePower, test.doubleToughness)
			}
			if effect.StaticSubject.Kind != test.subject {
				t.Fatalf("static subject = %v, want %v", effect.StaticSubject.Kind, test.subject)
			}
		})
	}
}

// TestParseDoubleCountersObject verifies the counter-doubling object "double the
// number of <kind> counters on <object>" captures the doubled counter kind, the
// target vs self scope, and the all-kinds form, and fails closed for an
// unrecognized counter noun.
func TestParseDoubleCountersObject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		source            string
		wantDouble        bool
		wantCounterPlus11 bool
		wantTarget        bool
		wantAllKinds      bool
	}{
		{
			"self this creature",
			"Double the number of +1/+1 counters on this creature.",
			true, true, false, false,
		},
		{
			"self it",
			"Double the number of +1/+1 counters on it.",
			true, true, false, false,
		},
		{
			"target creature single kind",
			"Double the number of +1/+1 counters on target creature.",
			true, true, true, false,
		},
		{
			"target each kind",
			"Double the number of each kind of counter on target artifact, creature, or land.",
			true, false, true, true,
		},
		{
			"self each kind",
			"Double the number of each kind of counter on this creature.",
			true, false, false, true,
		},
		{
			"unrecognized counter noun fails closed",
			"Double the number of cards on target creature.",
			false, false, false, false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			effect := effects[0]
			if effect.Kind != EffectDouble {
				t.Fatalf("effect Kind = %v, want EffectDouble", effect.Kind)
			}
			if effect.DoubleSourceCounters != test.wantDouble {
				t.Fatalf("DoubleSourceCounters = %v, want %v", effect.DoubleSourceCounters, test.wantDouble)
			}
			if test.wantCounterPlus11 && effect.DoubleSourceCounterKind != counter.PlusOnePlusOne {
				t.Fatalf("DoubleSourceCounterKind = %v, want PlusOnePlusOne", effect.DoubleSourceCounterKind)
			}
			if effect.DoubleCountersTarget != test.wantTarget {
				t.Fatalf("DoubleCountersTarget = %v, want %v", effect.DoubleCountersTarget, test.wantTarget)
			}
			if effect.DoubleCountersAllKinds != test.wantAllKinds {
				t.Fatalf("DoubleCountersAllKinds = %v, want %v", effect.DoubleCountersAllKinds, test.wantAllKinds)
			}
		})
	}
}

func TestParseLeadingInsteadSearchReplacement(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Sacrifice a land. Search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle. If you control a creature with power 4 or greater, instead search your library for up to three basic land cards, put them onto the battlefield tapped, then shuffle.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := document.Abilities[0].Sentences[2].Effects
	if len(effects) == 0 {
		t.Fatal("no effects parsed for the conditional search sentence")
	}
	search := effects[0]
	if search.Kind != EffectSearch {
		t.Fatalf("first effect kind = %v, want EffectSearch", search.Kind)
	}
	if search.Replacement.Kind != EffectReplacementInstead {
		t.Fatalf("replacement kind = %v, want instead", search.Replacement.Kind)
	}
	if got := searchUnsupportedDetail(&search); got != "" {
		t.Fatalf("instead search clause unsupported: %q", got)
	}
}

func TestParseRegenerateRecipientExactness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		source       string
		cardName     string
		wantExact    bool
		wantAttached bool
	}{
		{name: "self this creature", source: "Regenerate this creature.", wantExact: true},
		{name: "self this permanent", source: "Regenerate this permanent.", wantExact: true},
		{name: "self this token", source: "Regenerate this token.", wantExact: true},
		{name: "self card name", source: "Regenerate Thrun.", cardName: "Thrun, the Last Troll", wantExact: true},
		{name: "self with reminder", source: "Regenerate this creature. (The next time this creature would be destroyed this turn, instead remove all damage from it and tap it.)", wantExact: true},
		{name: "target creature", source: "Regenerate target creature.", wantExact: true},
		{name: "target restricted", source: "Regenerate target creature you control.", wantExact: true},
		{name: "enchanted creature", source: "Regenerate enchanted creature.", wantExact: true, wantAttached: true},
		{name: "equipped creature", source: "Regenerate equipped creature.", wantExact: true, wantAttached: true},
		// Fail-closed: mass and pronoun forms stay non-exact.
		{name: "mass each", source: "Regenerate each creature you control.", wantExact: false},
		{name: "pronoun it", source: "If this creature would be destroyed, regenerate it.", wantExact: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{CardName: test.cardName})
			var regen *EffectSyntax
			for ai := range document.Abilities {
				for si := range document.Abilities[ai].Sentences {
					for ei := range document.Abilities[ai].Sentences[si].Effects {
						effect := &document.Abilities[ai].Sentences[si].Effects[ei]
						if effect.Kind == EffectRegenerate {
							regen = effect
						}
					}
				}
			}
			if regen == nil {
				t.Fatalf("no EffectRegenerate parsed from %q", test.source)
			}
			if regen.Exact != test.wantExact {
				t.Fatalf("Exact = %v, want %v", regen.Exact, test.wantExact)
			}
			if regen.RegenerateAttached != test.wantAttached {
				t.Fatalf("RegenerateAttached = %v, want %v", regen.RegenerateAttached, test.wantAttached)
			}
		})
	}
}

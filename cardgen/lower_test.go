package cardgen

import (
	"fmt"
	"go/parser"
	"go/token"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func lowerSingleFace(t *testing.T, card *ScryfallCard) loweredFaceAbilities {
	t.Helper()
	faces, diagnostics := lowerExecutableFaces(card)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if len(faces) == 0 {
		t.Fatal("no faces lowered")
	}
	return faces[0]
}

func TestLowerKeywordAbilityStaticBodies(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Flying\nVigilance",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.StaticAbilities) != 2 {
		t.Fatalf("got %d static abilities, want 2", len(face.StaticAbilities))
	}
	if got := face.StaticAbilities[0].VarName; got != "game.FlyingStaticBody" {
		t.Fatalf("first static VarName = %q", got)
	}
	if got := face.StaticAbilities[1].VarName; got != "game.VigilanceStaticBody" {
		t.Fatalf("second static VarName = %q", got)
	}
}

func TestLowerKeywordAbilityWard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Ward {2}",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("got %d static abilities, want 1", len(face.StaticAbilities))
	}
	body := face.StaticAbilities[0].Body
	if len(body.KeywordAbilities) != 1 {
		t.Fatalf("got %d keyword abilities, want 1", len(body.KeywordAbilities))
	}
	ward, ok := body.KeywordAbilities[0].(game.WardKeyword)
	if !ok {
		t.Fatalf("keyword ability = %T, want game.WardKeyword", body.KeywordAbilities[0])
	}
	if len(ward.Cost) != 1 || ward.Cost[0].Kind != cost.GenericSymbol || ward.Cost[0].Generic != 2 {
		t.Fatalf("ward cost = %#v, want {2}", ward.Cost)
	}
}

func TestLowerCyclingAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Card",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Cycling {1}{U} ({1}{U}, Discard this card: Draw a card.)",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("got %d activated abilities, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if !ability.ManaCost.Exists {
		t.Fatal("cycling ability has no mana cost")
	}
	if len(ability.AdditionalCosts) != 1 || ability.AdditionalCosts[0].Kind != cost.AdditionalDiscard {
		t.Fatalf("additional costs = %#v, want one discard", ability.AdditionalCosts)
	}
	if len(ability.KeywordAbilities) != 1 {
		t.Fatalf("got %d keyword abilities, want 1", len(ability.KeywordAbilities))
	}
	if _, ok := ability.KeywordAbilities[0].(game.CyclingKeyword); !ok {
		t.Fatalf("keyword ability = %T, want game.CyclingKeyword", ability.KeywordAbilities[0])
	}
}

func TestLowerIssue210AdditionalCosts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		want       cost.AdditionalKind
		assert     func(t *testing.T, additional cost.Additional)
	}{
		{
			name:       "exert source",
			oracleText: "Exert this creature: Draw a card.",
			want:       cost.AdditionalExert,
		},
		{
			name:       "mill cards",
			oracleText: "Mill four cards: Draw a card.",
			want:       cost.AdditionalMill,
			assert: func(t *testing.T, additional cost.Additional) {
				t.Helper()
				if additional.Amount != 4 {
					t.Fatalf("amount = %d, want 4", additional.Amount)
				}
			},
		},
		{
			name:       "put counter on source",
			oracleText: "Put a verse counter on Test Bard: Draw a card.",
			want:       cost.AdditionalPutCounter,
			assert: func(t *testing.T, additional cost.Additional) {
				t.Helper()
				if additional.Amount != 1 || additional.CounterKind != counter.Verse {
					t.Fatalf("additional = %#v, want one verse counter", additional)
				}
			},
		},
		{
			name:       "put counters on source",
			oracleText: "Put two charge counters on Test Bard: Draw a card.",
			want:       cost.AdditionalPutCounter,
			assert: func(t *testing.T, additional cost.Additional) {
				t.Helper()
				if additional.Amount != 2 || additional.CounterKind != counter.Charge {
					t.Fatalf("additional = %#v, want two charge counters", additional)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bard",
				Layout:     "normal",
				TypeLine:   "Creature — Human Bard",
				OracleText: test.oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
			}
			ability := face.ActivatedAbilities[0]
			if len(ability.AdditionalCosts) != 1 {
				t.Fatalf("additional costs = %#v, want 1", ability.AdditionalCosts)
			}
			additional := ability.AdditionalCosts[0]
			if additional.Kind != test.want {
				t.Fatalf("additional kind = %v, want %v", additional.Kind, test.want)
			}
			if test.assert != nil {
				test.assert(t, additional)
			}
		})
	}
}

func TestLowerCollectEvidenceAdditionalCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Detective",
		Layout:     "normal",
		TypeLine:   "Creature — Human Detective",
		OracleText: "Collect evidence 4: Draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	additional := face.ActivatedAbilities[0].AdditionalCosts[0]
	if additional.Kind != cost.AdditionalCollectEvidence ||
		additional.Amount != 4 ||
		additional.Source != zone.Graveyard {
		t.Fatalf("additional = %#v, want collect evidence 4 from graveyard", additional)
	}
}

func TestLowerCollectEvidenceRejectsMalformedThresholds(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Collect evidence 0: Draw a card.",
		"Collect evidence two: Draw a card.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Malformed Detective",
				Layout:     "normal",
				TypeLine:   "Creature — Human Detective",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(diagnostics) == 0 {
				t.Fatal("expected collect-evidence diagnostic")
			}
		})
	}
}

func TestLowerCounterSpellTargets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		oracleText        string
		wantSpellTypes    []types.Card
		wantExcludedTypes []types.Card
	}{
		{
			name:       "any spell",
			oracleText: "Counter target spell.",
		},
		{
			name:           "creature spell",
			oracleText:     "Counter target creature spell.",
			wantSpellTypes: []types.Card{types.Creature},
		},
		{
			name:           "artifact spell",
			oracleText:     "Counter target artifact spell.",
			wantSpellTypes: []types.Card{types.Artifact},
		},
		{
			name:           "instant spell",
			oracleText:     "Counter target instant spell.",
			wantSpellTypes: []types.Card{types.Instant},
		},
		{
			name:           "sorcery spell",
			oracleText:     "Counter target sorcery spell.",
			wantSpellTypes: []types.Card{types.Sorcery},
		},
		{
			name:              "noncreature spell",
			oracleText:        "Counter target noncreature spell.",
			wantExcludedTypes: []types.Card{types.Creature},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
			})
			if !face.SpellAbility.Exists {
				t.Fatal("spell ability missing")
			}
			ability := face.SpellAbility.Val
			if len(ability.Modes) != 1 {
				t.Fatalf("modes = %d, want 1", len(ability.Modes))
			}
			mode := ability.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(mode.Targets))
			}
			target := mode.Targets[0]
			if target.Allow != game.TargetAllowStackObject {
				t.Fatalf("target allow = %v, want stack object", target.Allow)
			}
			if !slices.Equal(target.Predicate.SpellCardTypes, test.wantSpellTypes) {
				t.Fatalf("spell types = %+v, want %+v", target.Predicate.SpellCardTypes, test.wantSpellTypes)
			}
			if !slices.Equal(target.Predicate.ExcludedSpellCardTypes, test.wantExcludedTypes) {
				t.Fatalf("excluded spell types = %+v, want %+v", target.Predicate.ExcludedSpellCardTypes, test.wantExcludedTypes)
			}
			if len(mode.Sequence) != 1 {
				t.Fatalf("sequence = %d, want 1", len(mode.Sequence))
			}
			counter, ok := mode.Sequence[0].Primitive.(game.CounterObject)
			if !ok {
				t.Fatalf("primitive = %T, want game.CounterObject", mode.Sequence[0].Primitive)
			}
			if counter.Object.Kind() != game.ObjectReferenceTargetStackObject || counter.Object.TargetIndex() != 0 {
				t.Fatalf("counter object = %+v, want target stack object 0", counter.Object)
			}
		})
	}
}

func TestLowerCounterSpellWithDrawRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dismiss",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Counter target spell. Draw a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	if mode.Targets[0].Allow != game.TargetAllowStackObject {
		t.Fatalf("target allow = %v, want stack object", mode.Targets[0].Allow)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want counter plus draw", len(mode.Sequence))
	}
	if _, ok := mode.Sequence[0].Primitive.(game.CounterObject); !ok {
		t.Fatalf("first primitive = %T, want game.CounterObject", mode.Sequence[0].Primitive)
	}
	if _, ok := mode.Sequence[1].Primitive.(game.Draw); !ok {
		t.Fatalf("second primitive = %T, want game.Draw", mode.Sequence[1].Primitive)
	}
}

func TestLowerCounterSpellRejectsUnsupportedForms(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Counter target blue spell.",
		"Counter target artifact or enchantment spell.",
		"Counter target spell unless its controller pays {1}.",
		"Counter target activated or triggered ability.",
		"Counter target activated ability from an artifact source.",
		"Counter target spell or ability that targets a creature.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Unsupported Counter",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: oracleText,
			})
			if len(diagnostics) == 0 {
				t.Fatal("expected counter-spell diagnostic")
			}
		})
	}
}

func TestLowerNinjutsuAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Ninja",
		Layout:     "normal",
		TypeLine:   "Creature — Human Ninja",
		OracleText: "Ninjutsu {1}{U} ({1}{U}, Return an unblocked attacker you control to hand: Put this card onto the battlefield from your hand tapped and attacking.)",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("got %d activated abilities, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if !ability.ManaCost.Exists || !slices.Equal(ability.ManaCost.Val, cost.Mana{cost.O(1), cost.U}) {
		t.Fatalf("mana cost = %#v, want {1}{U}", ability.ManaCost)
	}
	if len(ability.AdditionalCosts) != 1 || ability.AdditionalCosts[0].Kind != cost.AdditionalReturnUnblockedAttacker {
		t.Fatalf("additional costs = %#v, want return unblocked attacker", ability.AdditionalCosts)
	}
	if len(ability.KeywordAbilities) != 1 {
		t.Fatalf("got %d keyword abilities, want 1", len(ability.KeywordAbilities))
	}
	if _, ok := ability.KeywordAbilities[0].(game.NinjutsuKeyword); !ok {
		t.Fatalf("keyword ability = %T, want game.NinjutsuKeyword", ability.KeywordAbilities[0])
	}
}

func TestLowerNinjutsuAbilityRejectsMalformedForms(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Ninjutsu",
		"Ninjutsu {1}{U} extra text",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Malformed Ninja",
				Layout:     "normal",
				TypeLine:   "Creature — Ninja",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(diagnostics) == 0 {
				t.Fatal("expected malformed Ninjutsu diagnostic")
			}
		})
	}
}

func TestLowerSelfCardGraveyardReturnToHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dragon",
		Layout:     "normal",
		TypeLine:   "Creature — Dragon",
		OracleText: "{3}{W}{W}: Return this card from your graveyard to your hand.",
		Power:      new("5"),
		Toughness:  new("5"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if ability.ZoneOfFunction != zone.Graveyard {
		t.Fatalf("zone of function = %v, want graveyard", ability.ZoneOfFunction)
	}
	sequence := ability.Content.Modes[0].Sequence
	move, ok := sequence[0].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MoveCard", sequence[0].Primitive)
	}
	if move.Card.Kind != game.CardReferenceSource || move.FromZone != zone.Graveyard || move.Destination != zone.Hand {
		t.Fatalf("move = %#v", move)
	}
}

func TestLowerSelfCardGraveyardReturnToBattlefieldTapped(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Zombie",
		Layout:     "normal",
		TypeLine:   "Creature — Zombie",
		OracleText: "{1}{B}, Discard two cards: Return this card from your graveyard to the battlefield tapped.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if face.ActivatedAbilities[0].ZoneOfFunction != zone.Graveyard {
		t.Fatalf("zone of function = %v, want graveyard", face.ActivatedAbilities[0].ZoneOfFunction)
	}
	sequence := face.ActivatedAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(sequence))
	}
	put, ok := sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("first primitive = %T, want game.PutOnBattlefield", sequence[0].Primitive)
	}
	cardRef, ok := put.Source.CardRef()
	if !ok || cardRef.Kind != game.CardReferenceSource {
		t.Fatalf("source = %#v", put.Source)
	}
	if !put.EntryTapped {
		t.Fatalf("put = %#v, want EntryTapped", put)
	}
}

func TestLowerSelfCardGraveyardReturnToBattlefieldWithCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Construct",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Construct",
		OracleText: "{3}{B}: Return this card from your graveyard to the battlefield tapped with two +1/+1 counters on it.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if face.ActivatedAbilities[0].ZoneOfFunction != zone.Graveyard {
		t.Fatalf("zone of function = %v, want graveyard", face.ActivatedAbilities[0].ZoneOfFunction)
	}
	sequence := face.ActivatedAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(sequence))
	}
	put, ok := sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("primitive = %T, want game.PutOnBattlefield", sequence[0].Primitive)
	}
	if !put.EntryTapped ||
		len(put.EntryCounters) != 1 ||
		put.EntryCounters[0].Kind != counter.PlusOnePlusOne ||
		put.EntryCounters[0].Amount != 2 {
		t.Fatalf("put = %#v", put)
	}
}

func TestLowerTargetedGraveyardReturnToHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sage",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target instant or sorcery card from your graveyard to your hand.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one", mode.Targets)
	}
	target := mode.Targets[0]
	if target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard ||
		!slices.Equal(target.Selection.Val.RequiredTypesAny, []types.Card{types.Instant, types.Sorcery}) {
		t.Fatalf("target = %#v", target)
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MoveCard", mode.Sequence[0].Primitive)
	}
	if move.Card.Kind != game.CardReferenceTarget || move.FromZone != zone.Graveyard || move.Destination != zone.Hand {
		t.Fatalf("move = %#v", move)
	}
}

func TestLowerTargetedGraveyardReturnCardsWithCyclingToHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Excavation",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return up to two target cards with cycling from your graveyard to your hand.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one variable target spec", mode.Targets)
	}
	target := mode.Targets[0]
	if target.MinTargets != 0 || target.MaxTargets != 2 ||
		target.Allow != game.TargetAllowCard ||
		target.TargetZone != zone.Graveyard ||
		target.Selection.Val.Keyword != game.Cycling ||
		target.Selection.Val.Controller != game.ControllerYou {
		t.Fatalf("target = %#v", target)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(mode.Sequence))
	}
	for i, instruction := range mode.Sequence {
		move, ok := instruction.Primitive.(game.MoveCard)
		if !ok {
			t.Fatalf("primitive %d = %T, want game.MoveCard", i, instruction.Primitive)
		}
		if move.Card.Kind != game.CardReferenceTarget || move.Card.TargetIndex != i ||
			move.FromZone != zone.Graveyard ||
			move.Destination != zone.Hand {
			t.Fatalf("move %d = %#v", i, move)
		}
	}
}

func TestLowerTargetedGraveyardReturnToLibrary(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Shaman",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put target card from your graveyard on the bottom of your library.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MoveCard", mode.Sequence[0].Primitive)
	}
	if mode.Targets[0].Allow != game.TargetAllowCard || move.Destination != zone.Library || !move.DestinationBottom {
		t.Fatalf("mode = %#v move = %#v", mode, move)
	}
}

func TestLowerTargetedGraveyardReturnToBattlefield(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bishop",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target creature card with mana value 3 or less from your graveyard to the battlefield tapped.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one", mode.Targets)
	}
	target := mode.Targets[0]
	if target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard {
		t.Fatalf("target = %#v", target)
	}
	selection := target.Selection.Val
	if !slices.Equal(selection.RequiredTypes, []types.Card{types.Creature}) ||
		selection.Controller != game.ControllerYou ||
		!selection.ManaValue.Exists ||
		selection.ManaValue.Val.Op != compare.LessOrEqual ||
		selection.ManaValue.Val.Value != 3 {
		t.Fatalf("selection = %#v", selection)
	}
	put, ok := mode.Sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("primitive = %T, want game.PutOnBattlefield", mode.Sequence[0].Primitive)
	}
	cardRef, ok := put.Source.CardRef()
	if !ok || cardRef.Kind != game.CardReferenceTarget || !put.EntryTapped {
		t.Fatalf("put = %#v", put)
	}
}

func TestLowerTargetedGraveyardPutOntoBattlefieldUnderYourControl(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Reanimator",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put target creature card from a graveyard onto the battlefield under your control.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	target := mode.Targets[0]
	if target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard {
		t.Fatalf("target = %#v", target)
	}
	if target.Selection.Val.Controller != game.ControllerAny {
		t.Fatalf("selection controller = %v, want any", target.Selection.Val.Controller)
	}
	put, ok := mode.Sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("primitive = %T, want game.PutOnBattlefield", mode.Sequence[0].Primitive)
	}
	if !put.Recipient.Exists || put.Recipient.Val != game.ControllerReference() {
		t.Fatalf("recipient = %#v, want controller", put.Recipient)
	}
}

func TestLowerTargetedGraveyardVehicleReturnToBattlefield(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Pilot",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target Vehicle card from your graveyard to the battlefield.",
	})
	selection := face.SpellAbility.Val.Modes[0].Targets[0].Selection.Val
	if !slices.Equal(selection.SubtypesAny, []types.Sub{types.Vehicle}) ||
		selection.Controller != game.ControllerYou {
		t.Fatalf("selection = %#v", selection)
	}
}

func TestLowerDynamicDamageCountsCardsWithCyclingInGraveyard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Zenith Flare",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Zenith Flare deals X damage to any target and you gain X life, where X is the number of cards with a cycling ability in your graveyard.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %T, want game.Damage", mode.Sequence[0].Primitive)
	}
	dynamic := damage.Amount.DynamicAmount()
	if !dynamic.Exists ||
		dynamic.Val.Kind != game.DynamicAmountCountCardsInZone ||
		dynamic.Val.Player == nil ||
		*dynamic.Val.Player != game.ControllerReference() ||
		dynamic.Val.CardZone != zone.Graveyard ||
		dynamic.Val.Selection == nil ||
		dynamic.Val.Selection.Keyword != game.Cycling {
		t.Fatalf("dynamic amount = %#v", dynamic)
	}
	gain, ok := mode.Sequence[1].Primitive.(game.GainLife)
	if !ok {
		t.Fatalf("primitive = %T, want game.GainLife", mode.Sequence[1].Primitive)
	}
	if gain.Player != game.ControllerReference() || !reflect.DeepEqual(gain.Amount, damage.Amount) {
		t.Fatalf("gain = %#v, damage amount = %#v", gain, damage.Amount)
	}
}

func TestLowerStaticPTCountsCardsWithCyclingInGraveyard(t *testing.T) {
	t.Parallel()
	power := "0"
	toughness := "4"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Vile Manifestation",
		Layout:     "normal",
		TypeLine:   "Creature — Horror",
		OracleText: "Vile Manifestation gets +1/+0 for each card with cycling in your graveyard.",
		Power:      &power,
		Toughness:  &toughness,
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %#v, want one", face.StaticAbilities)
	}
	continuous := face.StaticAbilities[0].Body.ContinuousEffects[0]
	if !continuous.PowerDeltaDynamic.Exists ||
		continuous.PowerDeltaDynamic.Val.Kind != game.DynamicAmountCountCardsInZone ||
		continuous.PowerDeltaDynamic.Val.Selection == nil ||
		continuous.PowerDeltaDynamic.Val.Selection.Keyword != game.Cycling ||
		continuous.PowerDeltaDynamic.Val.CardZone != zone.Graveyard ||
		continuous.ToughnessDeltaDynamic.Exists {
		t.Fatalf("continuous effect = %#v", continuous)
	}
}

func TestGenerateExecutableCardSourceTargetedGraveyardReturnRendersCardTargetConstraints(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Shaman",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put target instant or sorcery card from your graveyard on the bottom of your library.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}

	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "test_shaman.go", source, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
	for _, want := range []string{
		"Allow:",
		"game.TargetAllowCard",
		"TargetZone:",
		"zone.Graveyard",
		"Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}, Controller: game.ControllerYou})",
		"Card:",
		"game.CardReference{Kind: game.CardReferenceTarget}",
		"Destination:",
		"zone.Library",
		"DestinationBottom:",
		"true",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateExecutableCardSourceWithCyclingTargetsRenderIndexedCardReferences(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Excavation",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return up to two target cards with cycling from your graveyard to your hand.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "test_excavation.go", source, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
	for _, want := range []string{
		"MinTargets: 0",
		"MaxTargets: 2",
		"Keyword: game.Cycling",
		"game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1}",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateExecutableCardSourceTargetedGraveyardReanimationRendersPutOnBattlefield(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Reanimator",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put target Vehicle card from a graveyard onto the battlefield under your control.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "test_reanimator.go", source, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
	for _, want := range []string{
		"Allow:",
		"game.TargetAllowCard",
		"TargetZone:",
		"zone.Graveyard",
		`SubtypesAny: []types.Sub{types.Sub("Vehicle")}`,
		"game.PutOnBattlefield",
		"game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget})",
		"Recipient: opt.Val(game.ControllerReference())",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateExecutableCardSourceSelfGraveyardReturnUsesEntryOptions(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Construct",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Construct",
		OracleText: "{3}{B}: Return this card from your graveyard to the battlefield tapped with two +1/+1 counters on it.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "test_construct.go", source, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
	for _, want := range []string{
		"game.PutOnBattlefield",
		"EntryTapped:",
		"true",
		"EntryCounters: []game.CounterPlacement{game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2}}",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	for _, unwanted := range []string{"game.Tap{", "game.AddCounter{"} {
		if strings.Contains(source, unwanted) {
			t.Fatalf("generated source contains follow-up primitive %q:\n%s", unwanted, source)
		}
	}
}

func TestLowerMutateAbilityAndTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mutator",
		Layout:     "normal",
		TypeLine:   "Creature — Beast",
		OracleText: "Mutate {1}{G}\nWhenever this creature mutates, draw a card.",
		Power:      new("3"),
		Toughness:  new("3"),
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want one Mutate ability", len(face.StaticAbilities))
	}
	mutateCost, ok := game.StaticBodyMutateCost(&face.StaticAbilities[0].Body)
	if !ok || !slices.Equal(mutateCost, cost.Mana{cost.O(1), cost.G}) {
		t.Fatalf("Mutate cost = %#v, want {1}{G}", mutateCost)
	}
	if len(face.TriggeredAbilities) != 1 ||
		face.TriggeredAbilities[0].Trigger.Type != game.TriggerWhenever ||
		face.TriggeredAbilities[0].Trigger.Pattern.Event != game.EventPermanentMutated ||
		face.TriggeredAbilities[0].Trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("Mutate trigger = %#v", face.TriggeredAbilities)
	}
}

func TestLowerMutateAbilityRejectsMalformedForms(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Mutate",
		"Mutate {1}{G} extra text",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Malformed Mutator",
				Layout:     "normal",
				TypeLine:   "Creature — Beast",
				OracleText: oracleText,
				Power:      new("3"),
				Toughness:  new("3"),
			})
			if len(diagnostics) == 0 {
				t.Fatal("expected malformed Mutate diagnostic")
			}
		})
	}
}

func TestLowerActivatedNonManaCosts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		check      func(*testing.T, []cost.Additional)
	}{
		{
			name:       "sacrifice source",
			oracleText: "Sacrifice this artifact: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 || costs[0].Kind != cost.AdditionalSacrificeSource {
					t.Fatalf("additional costs = %#v, want source sacrifice", costs)
				}
			},
		},
		{
			name:       "typed sacrifice after mana and tap",
			oracleText: "{2}, {T}, Sacrifice a creature: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 2 ||
					costs[0].Kind != cost.AdditionalTap ||
					costs[1].Kind != cost.AdditionalSacrifice ||
					!costs[1].MatchPermanentType ||
					costs[1].PermanentType != types.Creature {
					t.Fatalf("additional costs = %#v, want tap and creature sacrifice", costs)
				}
			},
		},
		{
			name:       "typed discard",
			oracleText: "Discard two creature cards: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalDiscard ||
					costs[0].Amount != 2 ||
					!costs[0].MatchCardType ||
					costs[0].CardType != types.Creature ||
					costs[0].Source != zone.Hand {
					t.Fatalf("additional costs = %#v, want two creature cards discarded", costs)
				}
			},
		},
		{
			name:       "pay life",
			oracleText: "Pay 2 life: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 || costs[0].Kind != cost.AdditionalPayLife || costs[0].Amount != 2 {
					t.Fatalf("additional costs = %#v, want 2 life", costs)
				}
			},
		},
		{
			name:       "exile source",
			oracleText: "Exile this artifact: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalExileSource ||
					costs[0].Source != zone.Battlefield {
					t.Fatalf("additional costs = %#v, want battlefield source exile", costs)
				}
			},
		},
		{
			name:       "exile graveyard card",
			oracleText: "Exile a card from your graveyard: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalExile ||
					costs[0].Amount != 1 ||
					costs[0].Source != zone.Graveyard ||
					costs[0].MatchCardType {
					t.Fatalf("additional costs = %#v, want one graveyard card exile", costs)
				}
			},
		},
		{
			name:       "exile typed graveyard card",
			oracleText: "Exile a creature card from your graveyard: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalExile ||
					costs[0].Amount != 1 ||
					costs[0].Source != zone.Graveyard ||
					!costs[0].MatchCardType ||
					costs[0].CardType != types.Creature {
					t.Fatalf("additional costs = %#v, want one graveyard creature card exile", costs)
				}
			},
		},
		{
			name:       "exile two graveyard cards",
			oracleText: "Exile two cards from your graveyard: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalExile ||
					costs[0].Amount != 2 ||
					costs[0].Source != zone.Graveyard {
					t.Fatalf("additional costs = %#v, want two graveyard card exiles", costs)
				}
			},
		},
		{
			name:       "untap source",
			oracleText: "{Q}: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalUntap ||
					costs[0].Text != "{Q}" {
					t.Fatalf("additional costs = %#v, want untap source", costs)
				}
			},
		},
		{
			name:       "remove source counter",
			oracleText: "Remove a +1/+1 counter from this artifact: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalRemoveCounter ||
					costs[0].Amount != 1 ||
					costs[0].CounterKind != counter.PlusOnePlusOne {
					t.Fatalf("additional costs = %#v, want source +1/+1 counter removal", costs)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: test.oracleText,
			})
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
			}
			test.check(t, face.ActivatedAbilities[0].AdditionalCosts)
		})
	}
}

func TestLowerActivatedTapPermanentsCosts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		check      func(*testing.T, cost.Additional)
	}{
		{
			name:       "tap two artifacts",
			oracleText: "Tap two untapped artifacts you control: Draw a card.",
			check: func(t *testing.T, additional cost.Additional) {
				t.Helper()
				if additional.Kind != cost.AdditionalTapPermanents ||
					additional.Amount != 2 ||
					!additional.MatchPermanentType ||
					additional.PermanentType != types.Artifact {
					t.Fatalf("additional cost = %#v, want tap two artifacts", additional)
				}
			},
		},
		{
			name:       "tap subtype permanent",
			oracleText: "Tap an untapped Merfolk you control: Draw a card.",
			check: func(t *testing.T, additional cost.Additional) {
				t.Helper()
				if additional.Kind != cost.AdditionalTapPermanents ||
					additional.Amount != 1 ||
					additional.MatchPermanentType ||
					additional.SubtypesAny[0] != types.Merfolk ||
					additional.SubtypesAny[1] != "" {
					t.Fatalf("additional cost = %#v, want tap one Merfolk", additional)
				}
			},
		},
		{
			name:       "tap elves",
			oracleText: "Tap two untapped Elves you control: Draw a card.",
			check: func(t *testing.T, additional cost.Additional) {
				t.Helper()
				if additional.Kind != cost.AdditionalTapPermanents ||
					additional.Amount != 2 ||
					additional.SubtypesAny[0] != types.Elf ||
					additional.SubtypesAny[1] != "" {
					t.Fatalf("additional cost = %#v, want tap two Elves", additional)
				}
			},
		},
		{
			name:       "tap dwarves",
			oracleText: "Tap two untapped Dwarves you control: Draw a card.",
			check: func(t *testing.T, additional cost.Additional) {
				t.Helper()
				if additional.Kind != cost.AdditionalTapPermanents ||
					additional.Amount != 2 ||
					additional.SubtypesAny[0] != types.Dwarf ||
					additional.SubtypesAny[1] != "" {
					t.Fatalf("additional cost = %#v, want tap two Dwarves", additional)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: test.oracleText,
			})
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
			}
			costs := face.ActivatedAbilities[0].AdditionalCosts
			if len(costs) != 1 {
				t.Fatalf("additional costs = %#v, want one", costs)
			}
			test.check(t, costs[0])
		})
	}
}

func TestLowerActivatedRemoveCounterCosts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wantAmount int
		wantKind   counter.Kind
	}{
		{
			name:       "plural storage counters",
			oracleText: "Remove two storage counters from this land: Draw a card.",
			wantAmount: 2,
			wantKind:   counter.Charge,
		},
		{
			name:       "number-word fuse counters",
			oracleText: "Remove five fuse counters from this enchantment: Draw a card.",
			wantAmount: 5,
			wantKind:   counter.Charge,
		},
		{
			name:       "verse counter",
			oracleText: "Remove a verse counter from this artifact: Draw a card.",
			wantAmount: 1,
			wantKind:   counter.Verse,
		},
		{
			name:       "time counters from it",
			oracleText: "Remove 3 time counters from it: Draw a card.",
			wantAmount: 3,
			wantKind:   counter.Time,
		},
		{
			name:       "oil counter",
			oracleText: "Remove an oil counter from this artifact: Draw a card.",
			wantAmount: 1,
			wantKind:   counter.Oil,
		},
		{
			name:       "blood counters",
			oracleText: "Remove two blood counters from this artifact: Draw a card.",
			wantAmount: 2,
			wantKind:   counter.Blood,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Artifact Enchantment Land",
				OracleText: test.oracleText,
			})
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
			}
			costs := face.ActivatedAbilities[0].AdditionalCosts
			if len(costs) != 1 ||
				costs[0].Kind != cost.AdditionalRemoveCounter ||
				costs[0].Amount != test.wantAmount ||
				costs[0].CounterKind != test.wantKind {
				t.Fatalf("additional costs = %#v, want amount %d kind %v", costs, test.wantAmount, test.wantKind)
			}
		})
	}
}

func TestLowerActivatedAbilityRejectsVariableRemoveCounterCosts(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Remove X storage counters from this land: Add {G}.",
		"Remove any number of storage counters from this land: Add {G}.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: oracleText,
			})
			if len(faces) != 1 || len(faces[0].ActivatedAbilities) != 0 {
				t.Fatalf("faces = %#v, want face with no partially lowered ability", faces)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported diagnostic")
			}
		})
	}
}

func TestLowerActivatedEnergyCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Engine",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Pay {E}{E}: Draw a card.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	costs := face.ActivatedAbilities[0].AdditionalCosts
	if len(costs) != 1 ||
		costs[0].Kind != cost.AdditionalEnergy ||
		costs[0].Amount != 2 {
		t.Fatalf("additional costs = %#v, want two-energy cost", costs)
	}
}

func TestLowerActivatedRevealCosts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		oracleText      string
		wantAmount      int
		wantAmountFromX bool
		wantColor       color.Color
	}{
		{
			name:       "fixed cards sharing color",
			oracleText: "{1}, {T}, Reveal two cards from your hand that share a color: Draw a card.",
			wantAmount: 2,
		},
		{
			name:            "variable blue cards",
			oracleText:      "{2}, Reveal X blue cards from your hand, Sacrifice this creature: Draw a card.",
			wantAmountFromX: true,
			wantColor:       color.Blue,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Creature",
				OracleText: test.oracleText,
			})
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
			}
			costs := face.ActivatedAbilities[0].AdditionalCosts
			var got cost.Additional
			for _, additional := range costs {
				if additional.Kind == cost.AdditionalReveal {
					got = additional
					break
				}
			}
			if got.Kind != cost.AdditionalReveal || got.Source != zone.Hand {
				t.Fatalf("additional costs = %#v, want reveal from hand", costs)
			}
			if got.Amount != test.wantAmount {
				t.Fatalf("Amount = %d, want %d", got.Amount, test.wantAmount)
			}
			if got.AmountFromX != test.wantAmountFromX {
				t.Fatalf("AmountFromX = %v, want %v", got.AmountFromX, test.wantAmountFromX)
			}
			if test.wantColor != "" {
				if !got.MatchCardColor || got.CardColor != test.wantColor {
					t.Fatalf("card color = %v/%v, want %v", got.MatchCardColor, got.CardColor, test.wantColor)
				}
			}
		})
	}
}

func TestLowerActivatedAbilityRejectsUnsupportedRevealCosts(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Reveal the player you chose: Draw a card.",
		"Reveal this card from your hand: Draw a card.",
		"Reveal a toy you own: Draw a card.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: oracleText,
			})
			if len(faces) != 1 || len(faces[0].ActivatedAbilities) != 0 {
				t.Fatalf("faces = %#v, want face with no partially lowered ability", faces)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported diagnostic")
			}
		})
	}
}

func TestLowerActivatedReturnToHandCosts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		oracleText        string
		wantAmount        int
		wantType          types.Card
		wantSubtype       types.Sub
		wantRequireTapped bool
		wantSupertype     types.Super
	}{
		{
			name:        "plural land subtype",
			oracleText:  "Return two Islands you control to their owner's hand: Draw a card.",
			wantAmount:  2,
			wantSubtype: types.Island,
		},
		{
			name:              "tapped creature",
			oracleText:        "Return a tapped creature you control to its owner's hand: Draw a card.",
			wantAmount:        1,
			wantType:          types.Creature,
			wantRequireTapped: true,
		},
		{
			name:          "snow lands",
			oracleText:    "Return three snow lands you control to their owner's hand: Draw a card.",
			wantAmount:    3,
			wantType:      types.Land,
			wantSupertype: types.Snow,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: test.oracleText,
			})
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
			}
			costs := face.ActivatedAbilities[0].AdditionalCosts
			if len(costs) != 1 || costs[0].Kind != cost.AdditionalReturnToHand || costs[0].Amount != test.wantAmount {
				t.Fatalf("additional costs = %#v, want return-to-hand amount %d", costs, test.wantAmount)
			}
			if costs[0].RequireTapped != test.wantRequireTapped {
				t.Fatalf("RequireTapped = %v, want %v", costs[0].RequireTapped, test.wantRequireTapped)
			}
			if test.wantType != "" && (!costs[0].MatchPermanentType || costs[0].PermanentType != test.wantType) {
				t.Fatalf("permanent type = %v/%v, want %v", costs[0].MatchPermanentType, costs[0].PermanentType, test.wantType)
			}
			if test.wantSubtype != "" && costs[0].SubtypesAny != (cost.SubtypeSet{test.wantSubtype}) {
				t.Fatalf("subtypes = %#v, want %v", costs[0].SubtypesAny, test.wantSubtype)
			}
			if costs[0].RequireSupertype != test.wantSupertype {
				t.Fatalf("RequireSupertype = %v, want %v", costs[0].RequireSupertype, test.wantSupertype)
			}
		})
	}
}

func TestLowerActivatedAbilityRejectsUnsupportedReturnToHandCosts(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Return target creature to its owner's hand: Draw a card.",
		"Return a creature an opponent controls to its owner's hand: Draw a card.",
		"Return a card from your graveyard to its owner's hand: Draw a card.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: oracleText,
			})
			if len(faces) != 1 || len(faces[0].ActivatedAbilities) != 0 {
				t.Fatalf("faces = %#v, want face with no partially lowered ability", faces)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported diagnostic")
			}
		})
	}
}

func TestLowerActivatedAbilityRejectsVariableTapPermanentsCost(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Engine",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Tap X untapped Soldiers you control: Draw a card.",
	})
	if len(faces) != 1 || len(faces[0].ActivatedAbilities) != 0 {
		t.Fatalf("faces = %#v, want face with no partially lowered ability", faces)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported diagnostic")
	}
}

func TestLowerActivatedAbilityRejectsAmbiguousExileCost(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Engine",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Exile a card: Draw a card.",
	})
	if len(faces) != 1 || len(faces[0].ActivatedAbilities) != 0 {
		t.Fatalf("faces = %#v, want face with no partially lowered ability", faces)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported diagnostic")
	}
}

func TestLowerActivatedAbilityRejectsCounterRemovalFromTarget(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Engine",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Remove a +1/+1 counter from target creature: Draw a card.",
	})
	if len(faces) != 1 || len(faces[0].ActivatedAbilities) != 0 {
		t.Fatalf("faces = %#v, want face with no partially lowered ability", faces)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported diagnostic")
	}
}

func TestLowerActivatedAbilityTiming(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		want       game.TimingRestriction
	}{
		{"sorcery", "{1}: Draw a card. Activate only as a sorcery.", game.SorceryOnly},
		{"once per turn", "{1}: Draw a card. Activate only once each turn.", game.OncePerTurn},
		{"combat", "{1}: Draw a card. Activate only during combat.", game.DuringCombat},
		{"upkeep", "{1}: Draw a card. Activate only during your upkeep.", game.DuringUpkeep},
		{
			"sorcery once per turn",
			"{1}: Draw a card. Activate only as a sorcery. Activate only once each turn.",
			game.SorceryOncePerTurn,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: test.oracleText,
			})
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
			}
			if got := face.ActivatedAbilities[0].Timing; got != test.want {
				t.Fatalf("timing = %v, want %v", got, test.want)
			}
		})
	}
}

func TestLowerManaAbilityTiming(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Engine",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{T}: Add {G}. Activate only during combat.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %d, want 1", len(face.ManaAbilities))
	}
	if got := face.ManaAbilities[0].Timing; got != game.DuringCombat {
		t.Fatalf("timing = %v, want %v", got, game.DuringCombat)
	}
}

func TestLowerUntapManaAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Engine",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{Q}: Add {G}.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %d, want 1", len(face.ManaAbilities))
	}
	costs := face.ManaAbilities[0].AdditionalCosts
	if len(costs) != 1 || costs[0].Kind != cost.AdditionalUntap {
		t.Fatalf("additional costs = %#v, want untap source", costs)
	}
}

func TestLowerEquipAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Equipment",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equip {2}",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("got %d activated abilities, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	equipCost, ok := game.ActivatedBodyEquipCost(&ability)
	if !ok || len(equipCost) != 1 || equipCost[0] != cost.O(2) {
		t.Fatalf("equip cost = %#v, %v; want {2}", equipCost, ok)
	}
}

func TestLowerEnchantCreatureAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("got %d static abilities, want 1", len(face.StaticAbilities))
	}
	target, ok := game.StaticBodyEnchantTarget(&face.StaticAbilities[0].Body)
	if !ok ||
		target.MinTargets != 1 ||
		target.MaxTargets != 1 ||
		target.Allow != game.TargetAllowPermanent ||
		len(target.Predicate.PermanentTypes) != 1 ||
		target.Predicate.PermanentTypes[0] != types.Creature {
		t.Fatalf("enchant target = %+v, %v; want one creature", target, ok)
	}
}

func TestLowerProtectionFromColorAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Protection from red",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("got %d static abilities, want 1", len(face.StaticAbilities))
	}
	protected := game.StaticBodyProtectionColors(&face.StaticAbilities[0].Body)
	if len(protected) != 1 || protected[0] != color.Red {
		t.Fatalf("protection colors = %v, want red", protected)
	}
}

func TestLowerProtectionFromColorWithSimpleKeyword(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Protection from red, haste",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.StaticAbilities) != 2 {
		t.Fatalf("got %d static abilities, want 2", len(face.StaticAbilities))
	}
	protected := game.StaticBodyProtectionColors(&face.StaticAbilities[0].Body)
	if len(protected) != 1 || protected[0] != color.Red {
		t.Fatalf("protection colors = %v, want red", protected)
	}
	if face.StaticAbilities[1].VarName != "game.HasteStaticBody" {
		t.Fatalf("second ability = %+v, want haste", face.StaticAbilities[1])
	}
}

func TestLowerProtectionFromMultipleColors(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Protection from black and from red",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	protected := game.StaticBodyProtectionColors(&face.StaticAbilities[0].Body)
	if len(protected) != 2 || protected[0] != color.Black || protected[1] != color.Red {
		t.Fatalf("protection colors = %v, want black and red", protected)
	}
}

func TestLowerEnchantedCreaturePTBuffAlongsideEnchant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\nEnchanted creature gets +2/+2.",
	})
	if len(face.StaticAbilities) != 2 {
		t.Fatalf("got %d static abilities, want 2", len(face.StaticAbilities))
	}
	body := face.StaticAbilities[1].Body
	if len(body.ContinuousEffects) != 1 {
		t.Fatalf("got %d continuous effects, want 1", len(body.ContinuousEffects))
	}
	ce := body.ContinuousEffects[0]
	if ce.Layer != game.LayerPowerToughnessModify {
		t.Fatalf("layer = %v, want LayerPowerToughnessModify", ce.Layer)
	}
	if ce.Group.Domain() != game.GroupDomainAttachedObject {
		t.Fatalf("group domain = %v, want GroupDomainAttachedObject", ce.Group.Domain())
	}
	if ce.PowerDelta != 2 || ce.ToughnessDelta != 2 {
		t.Fatalf("PT delta = %d/%d, want 2/2", ce.PowerDelta, ce.ToughnessDelta)
	}
}

func TestLowerEquippedCreaturePTBuff(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Equipment",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equipped creature gets +2/+0.\nEquip {2}",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("got %d static abilities, want 1", len(face.StaticAbilities))
	}
	body := face.StaticAbilities[0].Body
	if len(body.ContinuousEffects) != 1 {
		t.Fatalf("got %d continuous effects, want 1", len(body.ContinuousEffects))
	}
	ce := body.ContinuousEffects[0]
	if ce.Layer != game.LayerPowerToughnessModify {
		t.Fatalf("layer = %v, want LayerPowerToughnessModify", ce.Layer)
	}
	if ce.Group.Domain() != game.GroupDomainAttachedObject {
		t.Fatalf("group domain = %v, want GroupDomainAttachedObject", ce.Group.Domain())
	}
	if ce.PowerDelta != 2 || ce.ToughnessDelta != 0 {
		t.Fatalf("PT delta = %d/%d, want 2/0", ce.PowerDelta, ce.ToughnessDelta)
	}
}

func TestLowerCreaturesYouControlPTBuff(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Anthem",
		Layout:     "normal",
		TypeLine:   "Creature — Soldier",
		OracleText: "Creatures you control get +1/+1.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	body := face.StaticAbilities[0].Body
	ce := body.ContinuousEffects[0]
	if ce.Group.Domain() != game.GroupDomainObjectControlled {
		t.Fatalf("group domain = %v, want GroupDomainObjectControlled", ce.Group.Domain())
	}
	selection := ce.Group.Selection()
	if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %+v, want creature requirement", selection)
	}
	if _, excluded := ce.Group.Exclusion(); excluded {
		t.Fatal("group exclusion unexpectedly set")
	}
	if ce.PowerDelta != 1 || ce.ToughnessDelta != 1 {
		t.Fatalf("PT delta = %d/%d, want 1/1", ce.PowerDelta, ce.ToughnessDelta)
	}
}

func TestLowerOtherCreaturesYouControlPTBuff(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Lord",
		Layout:     "normal",
		TypeLine:   "Creature — Soldier",
		OracleText: "Other creatures you control get +1/+0.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	body := face.StaticAbilities[0].Body
	ce := body.ContinuousEffects[0]
	if ce.Group.Domain() != game.GroupDomainObjectControlled {
		t.Fatalf("group domain = %v, want GroupDomainObjectControlled", ce.Group.Domain())
	}
	if _, excluded := ce.Group.Exclusion(); !excluded {
		t.Fatal("group exclusion missing")
	}
	if ce.PowerDelta != 1 || ce.ToughnessDelta != 0 {
		t.Fatalf("PT delta = %d/%d, want 1/0", ce.PowerDelta, ce.ToughnessDelta)
	}
}

func TestLowerTapManaAbilityFixedColor(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {G}.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	mode := face.ManaAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("got %d instructions, want 1", len(mode.Sequence))
	}
	addMana, ok := mode.Sequence[0].Primitive.(game.AddMana)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddMana", mode.Sequence[0].Primitive)
	}
	if addMana.ManaColor != mana.G {
		t.Fatalf("mana color = %q, want mana.G", addMana.ManaColor)
	}
}

func TestLowerTapManaAbilityChoice(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {R} or {G}.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	mode := face.ManaAbilities[0].Content.Modes[0]
	choose, ok := mode.Sequence[0].Primitive.(game.Choose)
	if !ok {
		t.Fatalf("primitive = %T, want game.Choose", mode.Sequence[0].Primitive)
	}
	if choose.Choice.Kind != game.ResolutionChoiceMana {
		t.Fatalf("choice kind = %v, want ResolutionChoiceMana", choose.Choice.Kind)
	}
	if len(choose.Choice.Colors) != 2 {
		t.Fatalf("choice colors = %#v, want two colors", choose.Choice.Colors)
	}
}

// TestLowerManaAbilityMultiSymbolOutput verifies that "{T}: Add {G}{W}." is
// lowered to a mana ability with two sequential AddMana instructions, one for
// each mana symbol. This is the single-tap / two-color-output shape shared by
// dual-color tap lands (e.g. Sungrass Prairie).
func TestLowerManaAbilityMultiSymbolOutput(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {G}{W}.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	ab := face.ManaAbilities[0]
	if ab.ManaCost.Exists {
		t.Fatalf("ManaCost = %v, want none", ab.ManaCost)
	}
	if len(ab.AdditionalCosts) != 1 || ab.AdditionalCosts[0].Kind != cost.AdditionalTap {
		t.Fatalf("AdditionalCosts = %#v, want [tap]", ab.AdditionalCosts)
	}
	mode := ab.Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(mode.Sequence))
	}
	first, ok := mode.Sequence[0].Primitive.(game.AddMana)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.AddMana", mode.Sequence[0].Primitive)
	}
	second, ok := mode.Sequence[1].Primitive.(game.AddMana)
	if !ok {
		t.Fatalf("sequence[1] = %T, want game.AddMana", mode.Sequence[1].Primitive)
	}
	if first.ManaColor != mana.G {
		t.Fatalf("first mana color = %q, want G", first.ManaColor)
	}
	if second.ManaColor != mana.W {
		t.Fatalf("second mana color = %q, want W", second.ManaColor)
	}
}

// TestLowerManaAbilityManaCostAndTap verifies that "{1}, {T}: Add {G}{W}." is
// lowered to a mana ability with ManaCost {1} and AdditionalCosts [tap], plus
// two sequential AddMana instructions. This is the Signet / mana-cost-tap-dual
// shape (e.g. Selesnya Signet, Sungrass Prairie variant).
func TestLowerManaAbilityManaCostAndTap(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Signet",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{1}, {T}: Add {G}{W}.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	ab := face.ManaAbilities[0]
	if !ab.ManaCost.Exists {
		t.Fatal("ManaCost missing, want {1}")
	}
	if len(ab.ManaCost.Val) != 1 {
		t.Fatalf("ManaCost symbols = %d, want 1", len(ab.ManaCost.Val))
	}
	if len(ab.AdditionalCosts) != 1 || ab.AdditionalCosts[0].Kind != cost.AdditionalTap {
		t.Fatalf("AdditionalCosts = %#v, want [tap]", ab.AdditionalCosts)
	}
	mode := ab.Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(mode.Sequence))
	}
	first, ok := mode.Sequence[0].Primitive.(game.AddMana)
	if !ok || first.ManaColor != mana.G {
		t.Fatalf("first AddMana = %v, want G", mode.Sequence[0].Primitive)
	}
	second, ok := mode.Sequence[1].Primitive.(game.AddMana)
	if !ok || second.ManaColor != mana.W {
		t.Fatalf("second AddMana = %v, want W", mode.Sequence[1].Primitive)
	}
}

// TestLowerManaAbilityTapPayLife verifies that "{T}, Pay 1 life: Add {U} or {R}."
// is lowered with a tap additional cost, a pay-life additional cost, and a
// two-color mana choice. This is the pain-land / filter-land shape.
func TestLowerManaAbilityTapPayLife(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Pain Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}, Pay 1 life: Add {U} or {R}.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	ab := face.ManaAbilities[0]
	if ab.ManaCost.Exists {
		t.Fatalf("ManaCost = %v, want none", ab.ManaCost)
	}
	if len(ab.AdditionalCosts) != 2 {
		t.Fatalf("AdditionalCosts count = %d, want 2", len(ab.AdditionalCosts))
	}
	if ab.AdditionalCosts[0].Kind != cost.AdditionalTap {
		t.Fatalf("AdditionalCosts[0].Kind = %v, want AdditionalTap", ab.AdditionalCosts[0].Kind)
	}
	if ab.AdditionalCosts[1].Kind != cost.AdditionalPayLife || ab.AdditionalCosts[1].Amount != 1 {
		t.Fatalf("AdditionalCosts[1] = %#v, want AdditionalPayLife amount=1", ab.AdditionalCosts[1])
	}
	mode := ab.Content.Modes[0]
	choose, ok := mode.Sequence[0].Primitive.(game.Choose)
	if !ok || choose.Choice.Kind != game.ResolutionChoiceMana || len(choose.Choice.Colors) != 2 {
		t.Fatalf("sequence[0] = %v, want mana choice of 2 colors", mode.Sequence[0].Primitive)
	}
}

// TestLowerManaAbilitySacrificeSelf verifies that "Sacrifice this creature: Add {C}."
// is lowered with an AdditionalSacrificeSource cost and a fixed colorless mana output.
func TestLowerManaAbilitySacrificeSelf(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Scion",
		Layout:     "normal",
		TypeLine:   "Creature — Eldrazi Scion",
		OracleText: "Sacrifice this creature: Add {C}.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	ab := face.ManaAbilities[0]
	if ab.ManaCost.Exists {
		t.Fatalf("ManaCost = %v, want none", ab.ManaCost)
	}
	if len(ab.AdditionalCosts) != 1 || ab.AdditionalCosts[0].Kind != cost.AdditionalSacrificeSource {
		t.Fatalf("AdditionalCosts = %#v, want [sacrifice source]", ab.AdditionalCosts)
	}
	mode := ab.Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(mode.Sequence))
	}
	addMana, ok := mode.Sequence[0].Primitive.(game.AddMana)
	if !ok || addMana.ManaColor != mana.C {
		t.Fatalf("sequence[0] = %v, want AddMana{C}", mode.Sequence[0].Primitive)
	}
}

// TestLowerManaAbilityPureManaAnyColor verifies that "{G}: Add one mana of any
// color." is lowered with a mana cost {G}, no additional costs, and a five-color
// choice output. This is the Orochi Leafcaller / Nomadic Elf shape.
func TestLowerManaAbilityPureManaAnyColor(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Leafcaller",
		Layout:     "normal",
		TypeLine:   "Creature — Snake Shaman",
		OracleText: "{G}: Add one mana of any color.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	ab := face.ManaAbilities[0]
	if !ab.ManaCost.Exists {
		t.Fatal("ManaCost missing, want {G}")
	}
	if len(ab.AdditionalCosts) != 0 {
		t.Fatalf("AdditionalCosts = %#v, want empty", ab.AdditionalCosts)
	}
	mode := ab.Content.Modes[0]
	choose, ok := mode.Sequence[0].Primitive.(game.Choose)
	if !ok || choose.Choice.Kind != game.ResolutionChoiceMana || len(choose.Choice.Colors) != 5 {
		t.Fatalf("sequence[0] = %v, want any-color mana choice", mode.Sequence[0].Primitive)
	}
}

// TestLowerManaAbilityPureManaFixed verifies that "{R}: Add {B}." is lowered
// with a mana cost {R}, no additional costs, and a single AddMana{B} instruction.
// This is the Agent of Stromgald / mana-conversion shape.
func TestLowerManaAbilityPureManaFixed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Agent",
		Layout:     "normal",
		TypeLine:   "Creature — Human Cleric",
		OracleText: "{R}: Add {B}.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	ab := face.ManaAbilities[0]
	if !ab.ManaCost.Exists {
		t.Fatal("ManaCost missing, want {R}")
	}
	if len(ab.AdditionalCosts) != 0 {
		t.Fatalf("AdditionalCosts = %#v, want empty", ab.AdditionalCosts)
	}
	mode := ab.Content.Modes[0]
	addMana, ok := mode.Sequence[0].Primitive.(game.AddMana)
	if !ok || addMana.ManaColor != mana.B {
		t.Fatalf("sequence[0] = %v, want AddMana{B}", mode.Sequence[0].Primitive)
	}
}

// TestLowerManaAbilityDiscardCost verifies that "Discard a card: Add {B}." is
// lowered with an AdditionalDiscard cost and a single AddMana{B} output.
// This is the Skirge Familiar family shape (mana ability with discard cost).
func TestLowerManaAbilityDiscardCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Skirge",
		Layout:     "normal",
		TypeLine:   "Creature — Imp",
		OracleText: "Discard a card: Add {B}.",
		Power:      new("3"),
		Toughness:  new("1"),
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	ab := face.ManaAbilities[0]
	if ab.ManaCost.Exists {
		t.Fatalf("ManaCost = %v, want none", ab.ManaCost)
	}
	if len(ab.AdditionalCosts) != 1 || ab.AdditionalCosts[0].Kind != cost.AdditionalDiscard {
		t.Fatalf("AdditionalCosts = %#v, want [discard]", ab.AdditionalCosts)
	}
	if ab.AdditionalCosts[0].Amount != 1 {
		t.Fatalf("discard amount = %d, want 1", ab.AdditionalCosts[0].Amount)
	}
	mode := ab.Content.Modes[0]
	addMana, ok := mode.Sequence[0].Primitive.(game.AddMana)
	if !ok || addMana.ManaColor != mana.B {
		t.Fatalf("sequence[0] = %v, want AddMana{B}", mode.Sequence[0].Primitive)
	}
}

// TestLowerManaAbilityTypedSacrifice verifies that "Sacrifice a creature: Add {C}{C}."
// is lowered with an AdditionalSacrifice cost targeting creatures and a two-instruction
// colorless mana output. This is the Ashnod's Altar shape.
func TestLowerManaAbilityTypedSacrifice(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Altar",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Sacrifice a creature: Add {C}{C}.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	ab := face.ManaAbilities[0]
	if ab.ManaCost.Exists {
		t.Fatalf("ManaCost = %v, want none", ab.ManaCost)
	}
	if len(ab.AdditionalCosts) != 1 {
		t.Fatalf("AdditionalCosts count = %d, want 1", len(ab.AdditionalCosts))
	}
	sacCost := ab.AdditionalCosts[0]
	if sacCost.Kind != cost.AdditionalSacrifice || sacCost.Amount != 1 ||
		!sacCost.MatchPermanentType || sacCost.PermanentType != types.Creature {
		t.Fatalf("AdditionalCosts[0] = %#v, want sacrifice-a-creature", sacCost)
	}
	mode := ab.Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(mode.Sequence))
	}
	for i, instr := range mode.Sequence {
		addMana, ok := instr.Primitive.(game.AddMana)
		if !ok || addMana.ManaColor != mana.C {
			t.Fatalf("sequence[%d] = %v, want AddMana{C}", i, instr.Primitive)
		}
	}
}

// TestLowerManaAbilityRejectsComplexBody verifies that mana abilities with body
// patterns outside the three supported shapes (fixed, choice, any-color) are
// rejected. "Three mana in any combination" requires Amount > 1 with a
// repeated-choice mechanism that is not yet supported.
func TestLowerManaAbilityRejectsComplexBody(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Goblin",
		Layout:     "normal",
		TypeLine:   "Creature — Goblin",
		OracleText: "{T}, Sacrifice a Forest: Add three mana in any combination of {R} and/or {G}.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported diagnostic for complex mana body")
	}
}

func TestLowerEntersTappedReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "This land enters tapped.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	if !face.ReplacementAbilities[0].Replacement.EntersTapped {
		t.Fatal("replacement is not an enters-tapped replacement")
	}
}

func TestLowerTokenCreationReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Anointed Procession",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if replacement.MatchEvent != game.EventTokenCreated ||
		replacement.ControllerFilter != game.TriggerControllerYou ||
		replacement.TokenMultiplier != 2 ||
		replacement.Duration != game.DurationPermanent {
		t.Fatalf("replacement = %+v, want token creation doubler", replacement)
	}
}

func TestLowerDamageReplacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		oracleText   string
		multiplier   int
		addend       int
		sourceColors []color.Color
	}{
		{
			name:         "red additive damage",
			oracleText:   "If another red source you control would deal damage to a permanent or player, it deals that much damage plus 1 to that permanent or player instead.",
			addend:       1,
			sourceColors: []color.Color{color.Red},
		},
		{
			name:       "double damage",
			oracleText: "If a source you control would deal damage to a permanent or player, it deals double that damage to that permanent or player instead.",
			multiplier: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Damage Replacer",
				Layout:     "normal",
				TypeLine:   "Creature",
				OracleText: test.oracleText,
				Power:      new("4"),
				Toughness:  new("5"),
			})
			if len(face.ReplacementAbilities) != 1 {
				t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
			}
			replacement := face.ReplacementAbilities[0].Replacement
			if replacement.MatchEvent != game.EventDamageDealt ||
				replacement.ControllerFilter != game.TriggerControllerYou ||
				replacement.DamageMultiplier != test.multiplier ||
				replacement.DamageAddend != test.addend ||
				!slices.Equal(replacement.DamageSourceColors, test.sourceColors) ||
				replacement.DamageExcludeSource != (test.name == "red additive damage") ||
				replacement.Duration != game.DurationPermanent {
				t.Fatalf("replacement = %+v, want damage replacement", replacement)
			}
		})
	}
}

func TestLowerCounterPlacementReplacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		oracleText       string
		matchCounterKind bool
		counterKind      counter.Kind
	}{
		{
			name:             "specific plus one counters",
			oracleText:       "If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on that creature instead.",
			matchCounterKind: true,
			counterKind:      counter.PlusOnePlusOne,
		},
		{
			name:       "any counters",
			oracleText: "If you would put one or more counters on a permanent or player, put twice that many of each of those kinds of counters on that permanent or player instead.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Counter Doubler",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: test.oracleText,
			})
			if len(face.ReplacementAbilities) != 1 {
				t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
			}
			replacement := face.ReplacementAbilities[0].Replacement
			if replacement.MatchEvent != game.EventCountersAdded ||
				replacement.ControllerFilter != game.TriggerControllerYou ||
				replacement.CounterMultiplier != 2 ||
				replacement.MatchCounterKind != test.matchCounterKind ||
				replacement.CounterKindFilter != test.counterKind ||
				replacement.Duration != game.DurationPermanent {
				t.Fatalf("replacement = %+v, want counter placement doubler", replacement)
			}
		})
	}
}

func TestGenerateTokenCreationReplacementSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Parallel Lives",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.",
	}, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.TokenCreationReplacement",
		"game.TriggerControllerYou",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "generated.go", source, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestGenerateDamageReplacementSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Embermaw Hellion",
		Layout:     "normal",
		TypeLine:   "Creature — Hellion",
		OracleText: "If another red source you control would deal damage to a permanent or player, it deals that much damage plus 1 to that permanent or player instead.",
		Power:      new("4"),
		Toughness:  new("5"),
	}, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.DamageReplacementExcludingSource",
		"color.Red",
		"game.TriggerControllerYou",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "generated.go", source, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestGenerateCounterPlacementReplacementSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Branching Evolution",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on that creature instead.",
	}, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.CounterPlacementReplacement",
		"counter.PlusOnePlusOne",
		"game.TriggerControllerYou",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "generated.go", source, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestLowerEntersWithCountersReplacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
		kind       counter.Kind
		amount     int
	}{
		{
			name:       "plus one counters",
			typeLine:   "Creature — Shapeshifter",
			oracleText: "This creature enters with three +1/+1 counters on it.",
			kind:       counter.PlusOnePlusOne,
			amount:     3,
		},
		{
			name:       "shield counter",
			typeLine:   "Creature — Human Knight",
			oracleText: "This creature enters with a shield counter on it.",
			kind:       counter.Shield,
			amount:     1,
		},
		{
			name:       "charge counters",
			typeLine:   "Artifact",
			oracleText: "This artifact enters with two charge counters on it.",
			kind:       counter.Charge,
			amount:     2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Permanent",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			})
			if len(face.ReplacementAbilities) != 1 {
				t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
			}
			replacement := face.ReplacementAbilities[0].Replacement
			if replacement.EntersTapped {
				t.Fatal("replacement unexpectedly enters tapped")
			}
			if len(replacement.EntersWithCounters) != 1 {
				t.Fatalf("counter placements = %#v, want one", replacement.EntersWithCounters)
			}
			placement := replacement.EntersWithCounters[0]
			if placement.Kind != test.kind || placement.Amount != test.amount {
				t.Fatalf("placement = %#v, want %v x%d", placement, test.kind, test.amount)
			}
		})
	}
}

func TestGenerateEntersWithCountersReplacementSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Creature",
		Layout:     "normal",
		TypeLine:   "Creature — Shapeshifter",
		OracleText: "This creature enters with three +1/+1 counters on it.",
		Power:      new("0"),
		Toughness:  new("0"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		`game.EntersWithCountersReplacement("This creature enters with three +1/+1 counters on it."`,
		"game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 3}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "generated.go", source, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestLowerEntersWithCountersRejectsUnsupportedForms(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"conditional": "If a creature died this turn, this creature enters with a +1/+1 counter on it.",
		"dynamic":     "This creature enters with X +1/+1 counters on it.",
	}
	for name, oracleText := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Creature",
				Layout:     "normal",
				TypeLine:   "Creature",
				OracleText: oracleText,
				Power:      new("1"),
				Toughness:  new("1"),
			})
			if len(diagnostics) == 0 {
				t.Fatal("expected diagnostic")
			}
			if diagnostics[0].Summary != "unsupported enters-with-counters replacement" {
				t.Fatalf("summary = %q, want unsupported enters-with-counters replacement", diagnostics[0].Summary)
			}
		})
	}
}

func TestLowerSelfZoneDestinationReplacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		cardName      string
		typeLine      string
		oracleText    string
		matchFromZone bool
		fromZone      zone.Type
		replaceToZone zone.Type
	}{
		{
			name:          "from anywhere into library",
			cardName:      "Darksteel Colossus",
			typeLine:      "Artifact Creature — Golem",
			oracleText:    "If Darksteel Colossus would be put into a graveyard from anywhere, reveal Darksteel Colossus and shuffle it into its owner's library instead.",
			replaceToZone: zone.Library,
		},
		{
			name:          "dies into exile",
			cardName:      "Test Phoenix",
			typeLine:      "Creature — Phoenix",
			oracleText:    "If this creature would die, exile it instead.",
			matchFromZone: true,
			fromZone:      zone.Battlefield,
			replaceToZone: zone.Exile,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       test.cardName,
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Power:      new("11"),
				Toughness:  new("11"),
			})
			if len(face.ReplacementAbilities) != 1 {
				t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
			}
			replacement := face.ReplacementAbilities[0].Replacement
			if replacement.MatchEvent != game.EventZoneChanged ||
				replacement.MatchFromZone != test.matchFromZone ||
				replacement.FromZone != test.fromZone ||
				!replacement.MatchToZone ||
				replacement.ToZone != zone.Graveyard ||
				replacement.ReplaceToZone != test.replaceToZone ||
				replacement.ShuffleIntoLibrary != (test.replaceToZone == zone.Library) ||
				replacement.RevealSource != (test.replaceToZone == zone.Library) {
				t.Fatalf("replacement = %+v, want self zone-destination replacement", replacement)
			}
		})
	}
}

func TestGenerateSelfZoneDestinationReplacementSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Darksteel Colossus",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Golem",
		OracleText: "If Darksteel Colossus would be put into a graveyard from anywhere, reveal Darksteel Colossus and shuffle it into its owner's library instead.",
		Power:      new("11"),
		Toughness:  new("11"),
	}, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EventZoneChanged",
		"MatchToZone:",
		"ToZone:",
		"zone.Graveyard",
		"ReplaceToZone:",
		"zone.Library",
		"ShuffleIntoLibrary:",
		"RevealSource:",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "generated.go", source, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestGenerateEquippedCreaturePTBuff(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Equipment",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equipped creature gets +2/+0.\nEquip {2}",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if !strings.Contains(source, "LayerPowerToughnessModify") {
		t.Fatalf("source does not contain static PT effect:\n%s", source)
	}
	if !strings.Contains(source, "AttachedObjectGroup") {
		t.Fatalf("source does not contain AttachedObjectGroup:\n%s", source)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "generated.go", source, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestGenerateEquippedCreaturePTBuffWithKeywords(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Equipment",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equipped creature gets +2/+2 and has trample and lifelink.\nEquip {3}",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.LayerPowerToughnessModify",
		"game.LayerAbility",
		"AddKeywords: []game.Keyword",
		"game.Trample",
		"game.Lifelink",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateControlledCreaturesPTBuffWithKeyword(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Anthem",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Creatures you control get +1/+1 and have vigilance.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if !strings.Contains(source, "game.Vigilance") {
		t.Fatalf("source missing vigilance:\n%s", source)
	}
}

func TestLowerStandaloneStaticKeywordGrants(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		domain     game.GroupReferenceDomain
		excluded   bool
		subtypes   []types.Sub
		keywords   []game.Keyword
	}{
		"controlled creatures": {
			oracleText: "Creatures you control have haste and vigilance.",
			domain:     game.GroupDomainObjectControlled,
			keywords:   []game.Keyword{game.Haste, game.Vigilance},
		},
		"other controlled creatures": {
			oracleText: "Other creatures you control have flying.",
			domain:     game.GroupDomainObjectControlled,
			excluded:   true,
			keywords:   []game.Keyword{game.Flying},
		},
		"controlled artifacts": {
			oracleText: "Artifacts you control have indestructible.",
			domain:     game.GroupDomainObjectControlled,
			keywords:   []game.Keyword{game.Indestructible},
		},
		"equipped creature": {
			oracleText: "Equipped creature has shroud and wither.",
			domain:     game.GroupDomainAttachedObject,
			keywords:   []game.Keyword{game.Shroud, game.Wither},
		},
		"controlled subtype": {
			oracleText: "Zombies you control have flying.",
			domain:     game.GroupDomainObjectControlled,
			subtypes:   []types.Sub{types.Zombie},
			keywords:   []game.Keyword{game.Flying},
		},
		"other controlled subtype": {
			oracleText: "Other Dinosaurs you control have haste.",
			domain:     game.GroupDomainObjectControlled,
			excluded:   true,
			subtypes:   []types.Sub{types.Dinosaur},
			keywords:   []game.Keyword{game.Haste},
		},
		"irregular plural subtype": {
			oracleText: "Elves you control have vigilance.",
			domain:     game.GroupDomainObjectControlled,
			subtypes:   []types.Sub{types.Elf},
			keywords:   []game.Keyword{game.Vigilance},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Grant",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: test.oracleText,
			})
			if len(face.StaticAbilities) != 1 {
				t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
			}
			effects := face.StaticAbilities[0].Body.ContinuousEffects
			if len(effects) != 1 {
				t.Fatalf("continuous effects = %#v, want 1", effects)
			}
			effect := effects[0]
			if effect.Layer != game.LayerAbility || effect.Group.Domain() != test.domain {
				t.Fatalf("continuous effect = %#v", effect)
			}
			if _, excluded := effect.Group.Exclusion(); excluded != test.excluded {
				t.Fatalf("group exclusion = %v, want %v", excluded, test.excluded)
			}
			if got := effect.Group.Selection().SubtypesAny; !slices.Equal(got, test.subtypes) {
				t.Fatalf("subtypes = %v, want %v", got, test.subtypes)
			}
			if !slices.Equal(effect.AddKeywords, test.keywords) {
				t.Fatalf("keywords = %v, want %v", effect.AddKeywords, test.keywords)
			}
		})
	}
}

func TestRejectUnknownSubtypeStaticKeywordGrant(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Grant",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Splorps you control have haste.",
	})
	if len(diagnostics) == 0 || diagnostics[0].Summary != "unsupported static ability" {
		t.Fatalf("diagnostics = %#v, want unsupported static ability", diagnostics)
	}
}

func TestRejectMalformedStandaloneStaticKeywordGrants(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Creatures you control have flying or haste.",
		"Creatures you control have and flying.",
		"Creatures you control have flying and.",
		"Creatures you control have flying haste.",
		"Creatures you control have infect.",
	} {
		_, diagnostics := lowerExecutableFaces(&ScryfallCard{
			Name:       "Test Grant",
			Layout:     "normal",
			TypeLine:   "Enchantment",
			OracleText: oracleText,
		})
		if len(diagnostics) == 0 {
			t.Fatalf("%q lowered without diagnostics", oracleText)
		}
	}
}

func TestLowerSourceConditionalKeywordGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Climber",
		Layout:     "normal",
		TypeLine:   "Creature — Ape",
		OracleText: "As long as you control a Mountain, this creature has menace and vigilance.",
		Power:      new("3"),
		Toughness:  new("3"),
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	ability := face.StaticAbilities[0].Body
	if !ability.Condition.Exists {
		t.Fatal("static ability has no condition")
	}
	condition := ability.Condition.Val
	if condition.Text != "As long as you control a Mountain" ||
		!slices.Equal(condition.ControllerControls.SubtypesAny, []types.Sub{types.Mountain}) {
		t.Fatalf("condition = %+v", condition)
	}
	if len(ability.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %#v", ability.ContinuousEffects)
	}
	effect := ability.ContinuousEffects[0]
	if effect.Layer != game.LayerAbility ||
		!effect.AffectedSource ||
		!slices.Equal(effect.AddKeywords, []game.Keyword{game.Menace, game.Vigilance}) {
		t.Fatalf("continuous effect = %+v", effect)
	}
}

func TestLowerPostfixSourceConditionalKeywordGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Healer",
		Layout:     "normal",
		TypeLine:   "Creature — Cleric",
		OracleText: "This creature has lifelink as long as you control another Cleric.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ability := face.StaticAbilities[0].Body
	condition := ability.Condition.Val
	if !condition.ControllerControls.ExcludeSource ||
		!slices.Equal(condition.ControllerControls.SubtypesAny, []types.Sub{types.Cleric}) {
		t.Fatalf("condition = %+v", condition)
	}
	effect := ability.ContinuousEffects[0]
	if !effect.AffectedSource || !slices.Equal(effect.AddKeywords, []game.Keyword{game.Lifelink}) {
		t.Fatalf("continuous effect = %+v", effect)
	}
}

func TestLowerPostfixLandSubtypeConditionalKeywordGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sergeant",
		Layout:     "normal",
		TypeLine:   "Creature — Soldier",
		OracleText: "This creature has double strike as long as you control a Gate.",
		Power:      new("3"),
		Toughness:  new("3"),
	})
	condition := face.StaticAbilities[0].Body.Condition.Val
	if !slices.Equal(condition.ControllerControls.SubtypesAny, []types.Sub{types.Gate}) {
		t.Fatalf("condition = %+v", condition)
	}
}

func TestLowerColorQualifiedSourceConditionalKeywordGrants(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText     string
		types          []types.Card
		colors         []color.Color
		excludedColors []color.Color
	}{
		"one color": {
			oracleText: "This creature has haste as long as you control a red creature.",
			types:      []types.Card{types.Creature},
			colors:     []color.Color{color.Red},
		},
		"either color": {
			oracleText: "This creature has lifelink as long as you control a white or black permanent.",
			colors:     []color.Color{color.White, color.Black},
		},
		"colorless": {
			oracleText: "This creature has haste as long as you control another colorless creature.",
			types:      []types.Card{types.Creature},
			excludedColors: []color.Color{
				color.White,
				color.Blue,
				color.Black,
				color.Red,
				color.Green,
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Creature",
				Layout:     "normal",
				TypeLine:   "Creature — Human",
				OracleText: test.oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			filter := face.StaticAbilities[0].Body.Condition.Val.ControllerControls
			if !slices.Equal(filter.Types, test.types) ||
				!slices.Equal(filter.ColorsAny, test.colors) ||
				!slices.Equal(filter.ExcludedColors, test.excludedColors) {
				t.Fatalf("filter = %+v", filter)
			}
		})
	}
}

func TestGenerateSourceConditionalKeywordGrant(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Flier",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "As long as you control an artifact, this creature has flying.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		`Condition: opt.Val(game.Condition{`,
		`Text: "As long as you control an artifact"`,
		`Types: []types.Card{types.Artifact}`,
		`AffectedSource: true`,
		`game.Flying`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateSourceConditionalProtectionGrant verifies Finding 4: a conditional
// self-grant of a parameterized Protection keyword is lowered using AddAbilities
// (not AddKeywords), analogous to the non-conditional grant path.
func TestGenerateSourceConditionalProtectionGrant(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wantSnip   string
	}{
		{
			name:       "protection from color conditional",
			oracleText: "As long as you control an artifact, this creature has protection from black.",
			wantSnip:   "game.ProtectionFromColorsStaticAbility(color.Black)",
		},
		{
			name:       "protection from each color conditional postfix",
			oracleText: "This creature has protection from each color as long as you control three or more artifacts.",
			wantSnip:   "game.ProtectionFromEachColorStaticAbility()",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Champion",
				Layout:     "normal",
				TypeLine:   "Artifact Creature — Soldier",
				OracleText: tc.oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("unexpected diagnostics: %#v", diagnostics)
			}
			for _, want := range []string{
				"AffectedSource: true",
				"AddAbilities:",
				tc.wantSnip,
			} {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}

// TestLowerSourceConditionalProtectionKeywordGrant verifies that
// lowerSourceConditionalKeywordGrant produces a ContinuousEffect with
// AddAbilities (not AddKeywords) for parameterized Protection.
func TestLowerSourceConditionalProtectionKeywordGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Champion",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Soldier",
		OracleText: "Metalcraft — As long as you control three or more artifacts, this creature has protection from all colors.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	ability := face.StaticAbilities[0].Body
	if !ability.Condition.Exists {
		t.Fatal("static ability has no condition")
	}
	if len(ability.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %#v", ability.ContinuousEffects)
	}
	effect := ability.ContinuousEffects[0]
	if effect.Layer != game.LayerAbility {
		t.Fatalf("effect layer = %v, want LayerAbility", effect.Layer)
	}
	if !effect.AffectedSource {
		t.Fatal("effect.AffectedSource should be true")
	}
	if len(effect.AddKeywords) != 0 {
		t.Fatalf("effect.AddKeywords = %v, want empty (should use AddAbilities for Protection)", effect.AddKeywords)
	}
	if len(effect.AddAbilities) != 1 {
		t.Fatalf("effect.AddAbilities len = %d, want 1", len(effect.AddAbilities))
	}
}

func TestRejectUnsupportedSourceConditionalKeywordGrant(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Attacker",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "As long as it's attacking, this creature has flying.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" {
		t.Fatalf("unexpected source:\n%s", source)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported conditional keyword diagnostic")
	}
}

func TestRejectStaticPTBuffWithUnsupportedKeywordText(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Equipped creature gets +2/+2 and has trample or lifelink.\nEquip {3}",
		"Equipped creature gets +2/+2 and has and trample.\nEquip {3}",
		"Equipped creature gets +2/+2 and has trample and.\nEquip {3}",
		"Equipped creature gets +2/+2 and has flying lifelink.\nEquip {3}",
	} {
		source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Test Equipment",
			Layout:     "normal",
			TypeLine:   "Artifact — Equipment",
			OracleText: oracleText,
		}, "t")
		if err != nil {
			t.Fatal(err)
		}
		if source != "" {
			t.Fatalf("unexpected source for %q:\n%s", oracleText, source)
		}
		if len(diagnostics) == 0 {
			t.Fatalf("expected unsupported diagnostic for %q", oracleText)
		}
	}
}

func TestRejectResolvingPTBuffAsStatic(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Equipment",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equipped creature gets +2/+0 until end of turn.\nEquip {2}",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected rejection of resolving P/T effect, got none")
	}
}

func TestRejectVariablePTBuff(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Equipment",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equipped creature gets +1/+0 for each Equipment attached to it.\nEquip {2}",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected rejection of variable-amount P/T buff, got none")
	}
}

func TestGenerateExtendedStaticPTBuffSubjects(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		want       string
	}{
		"walls": {
			oracleText: "Each Wall you control gets +0/+2.",
			want:       `SubtypesAny: []types.Sub{types.Sub("Wall")}`,
		},
		"artifacts": {
			oracleText: "Artifacts you control get +1/+1.",
			want:       "RequiredTypes: []types.Card{types.Artifact}",
		},
		"tokens": {
			oracleText: "Tokens you control get +1/+1.",
			want:       "TokenOnly: true",
		},
		"opponents' creatures": {
			oracleText: "Creatures your opponents control get -1/-0.",
			want:       "Controller: game.ControllerOpponent",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Anthem",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: test.oracleText,
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if !strings.Contains(source, test.want) {
				t.Fatalf("source missing %q:\n%s", test.want, source)
			}
		})
	}
}

func TestLowerConditionalEntersTappedReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Vista",
		Layout:     "normal",
		TypeLine:   "Land — Forest Plains",
		OracleText: "This land enters tapped unless you control two or more basic lands.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	repl := face.ReplacementAbilities[0]
	if !repl.Replacement.EntersTapped {
		t.Fatal("replacement is not an enters-tapped replacement")
	}
	if !repl.Replacement.Condition.Exists {
		t.Fatal("conditional replacement has no condition")
	}
	cond := repl.Replacement.Condition.Val
	if !cond.Negate {
		t.Fatal("condition should be negated (unless)")
	}
	filter := cond.ControllerControls
	if len(filter.Types) != 1 || filter.Types[0] != types.Land {
		t.Fatalf("filter types = %#v, want [types.Land]", filter.Types)
	}
	if len(filter.Supertypes) != 1 || filter.Supertypes[0] != types.Basic {
		t.Fatalf("filter supertypes = %#v, want [types.Basic]", filter.Supertypes)
	}
	if filter.MinCount != 2 {
		t.Fatalf("filter MinCount = %d, want 2", filter.MinCount)
	}
}

func TestLowerCommonConditionalEntersTappedReplacements(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		oracleText    string
		negate        bool
		minCount      int
		excludeSource bool
		subtypes      []types.Sub
	}{
		{
			name:          "two or more other lands",
			oracleText:    "This land enters tapped unless you control two or more other lands.",
			negate:        true,
			minCount:      2,
			excludeSource: true,
		},
		{
			name:          "two or fewer other lands",
			oracleText:    "This land enters tapped unless you control two or fewer other lands.",
			minCount:      3,
			excludeSource: true,
		},
		{
			name:       "basic land subtype pair",
			oracleText: "This land enters tapped unless you control a Plains or an Island.",
			subtypes:   []types.Sub{types.Plains, types.Island},
			negate:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Land",
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: test.oracleText,
			})
			condition := face.ReplacementAbilities[0].Replacement.Condition.Val
			filter := condition.ControllerControls
			if condition.Negate != test.negate ||
				filter.MinCount != test.minCount ||
				filter.ExcludeSource != test.excludeSource ||
				!slices.Equal(filter.SubtypesAny, test.subtypes) {
				t.Fatalf("condition = %+v, want negate=%v min=%d exclude=%v subtypes=%v",
					condition, test.negate, test.minCount, test.excludeSource, test.subtypes)
			}
		})
	}
}

func TestLowerLifeAndOpponentConditionalEntersTappedReplacements(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		condition string
		assert    func(*testing.T, game.Condition)
	}{
		{
			name:      "controller life",
			condition: "unless you have 10 or more life",
			assert: func(t *testing.T, condition game.Condition) {
				if condition.ControllerLifeAtLeast != 10 {
					t.Fatalf("ControllerLifeAtLeast = %d, want 10", condition.ControllerLifeAtLeast)
				}
			},
		},
		{
			name:      "any player life",
			condition: "unless a player has 13 or less life",
			assert: func(t *testing.T, condition game.Condition) {
				if condition.AnyPlayerLifeAtMost != 13 {
					t.Fatalf("AnyPlayerLifeAtMost = %d, want 13", condition.AnyPlayerLifeAtMost)
				}
			},
		},
		{
			name:      "opponent count",
			condition: "unless you have two or more opponents",
			assert: func(t *testing.T, condition game.Condition) {
				if condition.OpponentCountAtLeast != 2 {
					t.Fatalf("OpponentCountAtLeast = %d, want 2", condition.OpponentCountAtLeast)
				}
			},
		},
		{
			name:      "one opponent land count",
			condition: "unless an opponent controls two or more lands",
			assert: func(t *testing.T, condition game.Condition) {
				if !condition.AnyOpponentControls.Exists ||
					condition.AnyOpponentControls.Val.MinCount != 2 {
					t.Fatalf("AnyOpponentControls = %+v, want two lands", condition.AnyOpponentControls)
				}
			},
		},
		{
			name:      "collective opponent land count",
			condition: "unless your opponents control eight or more lands",
			assert: func(t *testing.T, condition game.Condition) {
				if !condition.OpponentsControl.Exists ||
					condition.OpponentsControl.Val.MinCount != 8 {
					t.Fatalf("OpponentsControl = %+v, want eight lands", condition.OpponentsControl)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Land",
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: "This land enters tapped " + test.condition + ".",
			})
			condition := face.ReplacementAbilities[0].Replacement.Condition.Val
			if !condition.Negate {
				t.Fatal("unless condition was not negated")
			}
			test.assert(t, condition)
		})
	}
}

func TestLowerOptionalEntryPayments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		assert     func(*testing.T, game.ResolutionPayment)
	}{
		{
			name:       "pay life",
			oracleText: "As this land enters, you may pay 2 life. If you don't, it enters tapped.",
			assert: func(t *testing.T, payment game.ResolutionPayment) {
				if len(payment.AdditionalCosts) != 1 ||
					payment.AdditionalCosts[0].Kind != cost.AdditionalPayLife ||
					payment.AdditionalCosts[0].Amount != 2 {
					t.Fatalf("payment = %+v, want pay 2 life", payment)
				}
			},
		},
		{
			name:       "reveal land subtype",
			oracleText: "As this land enters, you may reveal a Mountain or Forest card from your hand. If you don't, this land enters tapped.",
			assert: func(t *testing.T, payment game.ResolutionPayment) {
				if len(payment.AdditionalCosts) != 1 {
					t.Fatalf("payment = %+v, want one reveal cost", payment)
				}
				additional := payment.AdditionalCosts[0]
				if additional.Kind != cost.AdditionalReveal ||
					additional.Source != zone.Hand ||
					additional.SubtypesAny != (cost.SubtypeSet{types.Mountain, types.Forest}) {
					t.Fatalf("additional cost = %+v, want Mountain-or-Forest reveal from hand", additional)
				}
			},
		},
		{
			name:       "reveal creature subtype",
			oracleText: "As this land enters, you may reveal a Giant card from your hand. If you don't, this land enters tapped.",
			assert: func(t *testing.T, payment game.ResolutionPayment) {
				if len(payment.AdditionalCosts) != 1 ||
					payment.AdditionalCosts[0].SubtypesAny != (cost.SubtypeSet{types.Giant}) {
					t.Fatalf("payment = %+v, want Giant reveal", payment)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Land",
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: test.oracleText,
			})
			if len(face.ReplacementAbilities) != 1 ||
				!face.ReplacementAbilities[0].UnlessPaid.Exists {
				t.Fatalf("replacement abilities = %+v, want one paid replacement", face.ReplacementAbilities)
			}
			test.assert(t, face.ReplacementAbilities[0].UnlessPaid.Val)
		})
	}
}

func TestLowerReminderManaAbilitySingleColor(t *testing.T) {
	t.Parallel()
	// Basic lands express their mana ability as a parenthesized reminder.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Forest",
		Layout:     "normal",
		TypeLine:   "Basic Land — Forest",
		OracleText: "({T}: Add {G}.)",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	mode := face.ManaAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("got %d instructions, want 1", len(mode.Sequence))
	}
	addMana, ok := mode.Sequence[0].Primitive.(game.AddMana)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddMana", mode.Sequence[0].Primitive)
	}
	if addMana.ManaColor != mana.G {
		t.Fatalf("mana color = %q, want mana.G", addMana.ManaColor)
	}
}

func TestLowerReminderManaAbilityChoice(t *testing.T) {
	t.Parallel()
	// Dual lands express their mana ability as a parenthesized reminder.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dual",
		Layout:     "normal",
		TypeLine:   "Land — Mountain Forest",
		OracleText: "({T}: Add {R} or {G}.)",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	mode := face.ManaAbilities[0].Content.Modes[0]
	choose, ok := mode.Sequence[0].Primitive.(game.Choose)
	if !ok {
		t.Fatalf("primitive = %T, want game.Choose", mode.Sequence[0].Primitive)
	}
	if choose.Choice.Kind != game.ResolutionChoiceMana {
		t.Fatalf("choice kind = %v, want ResolutionChoiceMana", choose.Choice.Kind)
	}
	if len(choose.Choice.Colors) != 2 {
		t.Fatalf("choice colors = %#v, want two colors", choose.Choice.Colors)
	}
}

func TestLowerNonManaHybridReminderDoesNotBlockCard(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Hybrid",
		Layout:     "normal",
		TypeLine:   "Creature — Soldier",
		OracleText: "({R/W} can be paid with either {R} or {W}.)\nFirst strike",
		Power:      new("1"),
		Toughness:  new("1"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	if source == "" {
		t.Fatal("expected generated source")
	}
}

func TestLowerNonManaReminderDoesNotBlockCard(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Card",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "(This creature can block as though it had flying.)\nFlying",
		Power:      new("1"),
		Toughness:  new("1"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	if source == "" {
		t.Fatal("expected generated source")
	}
}

func TestLowerAbilityWordDoesNotBlockSupportedKeyword(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Threshold",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "Threshold — Flying",
		Power:      new("1"),
		Toughness:  new("1"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}

	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	if source == "" {
		t.Fatal("expected generated source")
	}
}

func TestLowerAbilityWordConditions(t *testing.T) {
	tests := []struct {
		name       string
		cardName   string
		typeLine   string
		oracleText string
		wants      []string
	}{
		{"threshold static", "Threshold Bear", "Creature — Bear", "Threshold — This creature gets +2/+2 as long as there are seven or more cards in your graveyard.", []string{"ControllerGraveyardCardCountAtLeast: 7"}},
		{"delirium static", "Delirium Bear", "Creature — Bear", "Delirium — This creature gets +1/+1 and has menace as long as there are four or more card types among cards in your graveyard.", []string{"ControllerGraveyardCardTypeCountAtLeast: 4", "AffectedSource: true"}},
		{"domain static", "Domain Bear", "Creature — Bear", "Domain — This creature gets +1/+1 for each basic land type among lands you control.", []string{"PowerDeltaDynamic: opt.Val(game.DynamicAmount{", "ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{", "game.DynamicAmountControllerBasicLandTypeCount"}},
		{"domain spell", "Tribal Flames", "Sorcery", "Domain — Tribal Flames deals X damage to any target, where X is the number of basic land types among lands you control.", []string{"game.DynamicAmountControllerBasicLandTypeCount"}},
		{"metalcraft trigger", "Metalcraft Bear", "Creature — Bear", "Metalcraft — When this creature enters, if you control three or more artifacts, draw a card.", []string{"InterveningCondition: opt.Val(game.Condition{", "MinCount:"}},
		{"hellbent activation", "Hellbent Bear", "Creature — Bear", "Hellbent — {1}: Draw a card. Activate only if you have no cards in hand.", []string{"ActivationCondition: opt.Val(game.Condition{", "ControllerHandEmpty: true"}},
		{"ferocious activation", "Ferocious Bear", "Creature — Bear", "Ferocious — {1}: Draw a card. Activate only if you control a creature with power 4 or greater.", []string{"ActivationCondition: opt.Val(game.Condition{", "Value: 4"}},
		{"coven trigger", "Coven Bear", "Creature — Bear", "Coven — At the beginning of combat on your turn, if you control three or more creatures with different powers, draw a card.", []string{"InterveningCondition: opt.Val(game.Condition{", "ControllerCreaturePowerDiversityAtLeast: 3"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			card := &ScryfallCard{
				Name:       test.cardName,
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			}
			if strings.HasPrefix(test.typeLine, "Creature") {
				card.Power = new("2")
				card.Toughness = new("2")
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if source == "" {
				t.Fatal("expected generated source")
			}
			for _, want := range test.wants {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}

func TestLowerAbilityWordConditionsFailClosed(t *testing.T) {
	tests := []string{
		"Threshold — This creature gets +2/+2 as long as there are six or more cards in your graveyard.",
		"Delirium — This creature gets +2/+2 as long as there are three or more card types among cards in your graveyard.",
		"Metalcraft — This creature gets +2/+2 as long as you control two or more artifacts.",
		"Hellbent — {1}: Draw a card. Activate only if you have one or fewer cards in hand.",
		"Ferocious — {1}: Draw a card. Activate only if you control a creature with power 3 or greater.",
		"Coven — At the beginning of combat on your turn, if you control three or more creatures with the same power, draw a card.",
	}
	for _, oracleText := range tests {
		t.Run(oracleText, func(t *testing.T) {
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Fail Closed Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" || len(diagnostics) == 0 {
				t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
			}
		})
	}
}

func TestLowerAbilityWordSurfacesActualUnsupportedKeyword(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Threshold",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "Threshold — Protection from everything",
		Power:      new("1"),
		Toughness:  new("1"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported keyword diagnostic")
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.Summary == "unsupported ability word" {
			t.Fatalf("diagnostics = %#v, want actual unsupported keyword diagnostic", diagnostics)
		}
	}
}

func TestLowerUnknownEmDashHeaderRemainsUnsupported(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Ticketed",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "{TK}{TK} — Menace",
		Power:      new("1"),
		Toughness:  new("1"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unknown em-dash header to remain unsupported")
	}
}

func TestLowerEnterTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}

	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.Pattern.Event != game.EventPermanentEnteredBattlefield {
		t.Fatalf("event = %v, want EventPermanentEnteredBattlefield", trigger.Pattern.Event)
	}
	if trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("source = %v, want TriggerSourceSelf", trigger.Pattern.Source)
	}
}

func TestLowerCombatEventTriggers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		text    string
		want    game.TriggerPattern
		wantTyp game.TriggerType
	}{
		{
			name: "attacks",
			text: "Whenever this creature attacks, draw a card.",
			want: game.TriggerPattern{
				Event:  game.EventAttackerDeclared,
				Source: game.TriggerSourceSelf,
			},
			wantTyp: game.TriggerWhenever,
		},
		{
			name: "blocks",
			text: "Whenever this creature blocks, draw a card.",
			want: game.TriggerPattern{
				Event:  game.EventBlockerDeclared,
				Source: game.TriggerSourceSelf,
			},
			wantTyp: game.TriggerWhenever,
		},
		{
			name: "becomes blocked",
			text: "Whenever this creature becomes blocked, draw a card.",
			want: game.TriggerPattern{
				Event:  game.EventAttackerBecameBlocked,
				Source: game.TriggerSourceSelf,
			},
			wantTyp: game.TriggerWhenever,
		},
		{
			name: "combat damage to player",
			text: "Whenever this creature deals combat damage to a player, draw a card.",
			want: game.TriggerPattern{
				Event:               game.EventDamageDealt,
				Source:              game.TriggerSourceSelf,
				Subject:             game.TriggerSubjectDamageSource,
				DamageRecipient:     game.DamageRecipientPlayer,
				RequireCombatDamage: true,
			},
			wantTyp: game.TriggerWhenever,
		},
		{
			name: "combat damage to creature",
			text: "Whenever this creature deals combat damage to a creature, draw a card.",
			want: game.TriggerPattern{
				Event:                game.EventDamageDealt,
				Source:               game.TriggerSourceSelf,
				Subject:              game.TriggerSubjectDamageSource,
				DamageRecipient:      game.DamageRecipientPermanent,
				DamageRecipientTypes: []types.Card{types.Creature},
				RequireCombatDamage:  true,
			},
			wantTyp: game.TriggerWhenever,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Fighter",
				Layout:     "normal",
				TypeLine:   "Creature — Human Warrior",
				OracleText: tc.text,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			trigger := face.TriggeredAbilities[0].Trigger
			if trigger.Type != tc.wantTyp {
				t.Fatalf("trigger type = %v, want %v", trigger.Type, tc.wantTyp)
			}
			if !reflect.DeepEqual(trigger.Pattern, tc.want) {
				t.Fatalf("pattern = %+v, want %+v", trigger.Pattern, tc.want)
			}
		})
	}
}

func TestLowerCombatEventTriggersFailClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Whenever this creature attacks alone, draw a card.",
		"Whenever this creature attacks and isn't blocked, draw a card.",
		"Whenever this creature attacks a player, draw a card.",
		"Whenever this creature attacks or blocks, draw a card.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Fighter",
				Layout:     "normal",
				TypeLine:   "Creature — Human Warrior",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatal("unsupported combat trigger unexpectedly lowered")
			}
		})
	}
}

func TestLowerDamageSourceTriggers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		text string
		want game.TriggerPattern
	}{
		{
			name: "self damage",
			text: "Whenever this creature deals damage, draw a card.",
			want: game.TriggerPattern{
				Event:   game.EventDamageDealt,
				Source:  game.TriggerSourceSelf,
				Subject: game.TriggerSubjectDamageSource,
			},
		},
		{
			name: "self damage to player",
			text: "Whenever this creature deals damage to a player, draw a card.",
			want: game.TriggerPattern{
				Event:           game.EventDamageDealt,
				Source:          game.TriggerSourceSelf,
				Subject:         game.TriggerSubjectDamageSource,
				DamageRecipient: game.DamageRecipientPlayer,
			},
		},
		{
			name: "self damage to opponent",
			text: "Whenever this creature deals damage to an opponent, draw a card.",
			want: game.TriggerPattern{
				Event:           game.EventDamageDealt,
				Source:          game.TriggerSourceSelf,
				Subject:         game.TriggerSubjectDamageSource,
				Player:          game.TriggerPlayerOpponent,
				DamageRecipient: game.DamageRecipientPlayer,
			},
		},
		{
			name: "self damage to creature",
			text: "Whenever this creature deals damage to a creature, draw a card.",
			want: game.TriggerPattern{
				Event:                game.EventDamageDealt,
				Source:               game.TriggerSourceSelf,
				Subject:              game.TriggerSubjectDamageSource,
				DamageRecipient:      game.DamageRecipientPermanent,
				DamageRecipientTypes: []types.Card{types.Creature},
			},
		},
		{
			name: "self combat damage",
			text: "Whenever this creature deals combat damage, draw a card.",
			want: game.TriggerPattern{
				Event:               game.EventDamageDealt,
				Source:              game.TriggerSourceSelf,
				Subject:             game.TriggerSubjectDamageSource,
				RequireCombatDamage: true,
			},
		},
		{
			name: "self combat damage to opponent",
			text: "Whenever this creature deals combat damage to an opponent, draw a card.",
			want: game.TriggerPattern{
				Event:               game.EventDamageDealt,
				Source:              game.TriggerSourceSelf,
				Subject:             game.TriggerSubjectDamageSource,
				Player:              game.TriggerPlayerOpponent,
				DamageRecipient:     game.DamageRecipientPlayer,
				RequireCombatDamage: true,
			},
		},
		{
			name: "equipped creature combat damage to player",
			text: "Whenever equipped creature deals combat damage to a player, draw a card.",
			want: game.TriggerPattern{
				Event:               game.EventDamageDealt,
				Source:              game.TriggerSourceAttachedPermanent,
				Subject:             game.TriggerSubjectDamageSource,
				DamageRecipient:     game.DamageRecipientPlayer,
				RequireCombatDamage: true,
			},
		},
		{
			name: "enchanted creature damage",
			text: "Whenever enchanted creature deals damage, draw a card.",
			want: game.TriggerPattern{
				Event:   game.EventDamageDealt,
				Source:  game.TriggerSourceAttachedPermanent,
				Subject: game.TriggerSubjectDamageSource,
			},
		},
		{
			name: "enchanted creature damage to opponent",
			text: "Whenever enchanted creature deals damage to an opponent, draw a card.",
			want: game.TriggerPattern{
				Event:           game.EventDamageDealt,
				Source:          game.TriggerSourceAttachedPermanent,
				Subject:         game.TriggerSubjectDamageSource,
				Player:          game.TriggerPlayerOpponent,
				DamageRecipient: game.DamageRecipientPlayer,
			},
		},
		{
			name: "equipped creature damage to creature",
			text: "Whenever equipped creature deals damage to a creature, draw a card.",
			want: game.TriggerPattern{
				Event:                game.EventDamageDealt,
				Source:               game.TriggerSourceAttachedPermanent,
				Subject:              game.TriggerSubjectDamageSource,
				DamageRecipient:      game.DamageRecipientPermanent,
				DamageRecipientTypes: []types.Card{types.Creature},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Fighter",
				Layout:     "normal",
				TypeLine:   "Creature — Human Warrior",
				OracleText: tc.text,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			trigger := face.TriggeredAbilities[0].Trigger
			if trigger.Type != game.TriggerWhenever {
				t.Fatalf("trigger type = %v, want TriggerWhenever", trigger.Type)
			}
			if !reflect.DeepEqual(trigger.Pattern, tc.want) {
				t.Fatalf("pattern = %+v, want %+v", trigger.Pattern, tc.want)
			}
		})
	}
}

func TestLowerDamageSourceTriggersFailClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Whenever this creature deals damage to a player or planeswalker, draw a card.",
		"Whenever one or more creatures you control deal damage to a player, draw a card.",
		"Whenever a creature you control deals combat damage to a player, draw a card.",
		"Whenever this creature deals combat damage to defending player, draw a card.",
		"Whenever equipped creature or this Equipment deals damage, draw a card.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Fighter",
				Layout:     "normal",
				TypeLine:   "Creature — Human Warrior",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatal("unsupported damage-source trigger unexpectedly lowered")
			}
		})
	}
}

func TestLowerLifeDamageReceivedTriggers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		text string
		want game.TriggerPattern
	}{
		{
			name: "you gain life",
			text: "Whenever you gain life, draw a card.",
			want: game.TriggerPattern{
				Event:  game.EventLifeGained,
				Player: game.TriggerPlayerYou,
			},
		},
		{
			name: "you lose life",
			text: "Whenever you lose life, draw a card.",
			want: game.TriggerPattern{
				Event:  game.EventLifeLost,
				Player: game.TriggerPlayerYou,
			},
		},
		{
			name: "opponent gains life",
			text: "Whenever an opponent gains life, draw a card.",
			want: game.TriggerPattern{
				Event:  game.EventLifeGained,
				Player: game.TriggerPlayerOpponent,
			},
		},
		{
			name: "opponent loses life",
			text: "Whenever an opponent loses life, you gain 1 life.",
			want: game.TriggerPattern{
				Event:  game.EventLifeLost,
				Player: game.TriggerPlayerOpponent,
			},
		},
		{
			name: "self dealt damage",
			text: "Whenever this creature is dealt damage, draw a card.",
			want: game.TriggerPattern{
				Event:           game.EventDamageDealt,
				Source:          game.TriggerSourceSelf,
				Subject:         game.TriggerSubjectPermanent,
				DamageRecipient: game.DamageRecipientPermanent,
			},
		},
		{
			name: "enchanted creature dealt damage",
			text: "Whenever enchanted creature is dealt damage, draw a card.",
			want: game.TriggerPattern{
				Event:           game.EventDamageDealt,
				Source:          game.TriggerSourceAttachedPermanent,
				DamageRecipient: game.DamageRecipientPermanent,
			},
		},
		{
			name: "equipped creature dealt damage",
			text: "Whenever equipped creature is dealt damage, draw a card.",
			want: game.TriggerPattern{
				Event:           game.EventDamageDealt,
				Source:          game.TriggerSourceAttachedPermanent,
				DamageRecipient: game.DamageRecipientPermanent,
			},
		},
		{
			name: "you are dealt damage",
			text: "Whenever you're dealt damage, draw a card.",
			want: game.TriggerPattern{
				Event:           game.EventDamageDealt,
				Player:          game.TriggerPlayerYou,
				DamageRecipient: game.DamageRecipientPlayer,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Cleric",
				Layout:     "normal",
				TypeLine:   "Creature — Human Cleric",
				OracleText: tc.text,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			trigger := face.TriggeredAbilities[0].Trigger
			if trigger.Type != game.TriggerWhenever {
				t.Fatalf("trigger type = %v, want TriggerWhenever", trigger.Type)
			}
			if !reflect.DeepEqual(trigger.Pattern, tc.want) {
				t.Fatalf("pattern = %+v, want %+v", trigger.Pattern, tc.want)
			}
		})
	}
}

func TestLowerLifeDamageReceivedTriggersFailClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Whenever you gain or lose life, draw a card.",
		"Whenever you gain life for the first time each turn, draw a card.",
		"Whenever this creature is dealt combat damage, draw a card.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Cleric",
				Layout:     "normal",
				TypeLine:   "Creature — Human Cleric",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatal("unsupported life/damage trigger unexpectedly lowered")
			}
		})
	}
}

func TestLowerKickedEnterTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Kicker",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Creature — Wizard",
		OracleText: "Kicker {1}{U}\nWhen this creature enters, if it was kicked, draw two cards.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.InterveningIf != "if it was kicked" ||
		!trigger.InterveningIfEventPermanentWasKicked {
		t.Fatalf("trigger = %+v, want kicked intervening-if", trigger)
	}
	draw, ok := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount != game.Fixed(2) {
		t.Fatalf("primitive = %+v, want draw two", face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerWasCastEnterTriggers(t *testing.T) {
	t.Parallel()
	for _, condition := range []string{"if it was cast", "if you cast it"} {
		t.Run(condition, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Construct",
				Layout:     "normal",
				TypeLine:   "Artifact Creature — Construct",
				OracleText: "When this creature enters, " + condition + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			trigger := face.TriggeredAbilities[0].Trigger
			if trigger.InterveningIf != condition || !trigger.InterveningIfEventPermanentWasCast {
				t.Fatalf("trigger = %+v, want was-cast intervening-if", trigger)
			}
		})
	}
}

func TestLowerAttackedThisTurnEnterTriggerFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Warrior",
		Layout:     "normal",
		TypeLine:   "Creature — Warrior",
		OracleText: "When this creature enters, if this creature attacked this turn, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("attacked-this-turn self-enter condition unexpectedly lowered")
	}
}

func TestLowerControlsPermanentEnterTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Artificer",
		Layout:     "normal",
		TypeLine:   "Creature — Artificer",
		OracleText: "When this creature enters, if you control an artifact, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.InterveningIf != "if you control an artifact" ||
		!trigger.InterveningCondition.Exists {
		t.Fatalf("trigger = %+v, want controls-artifact intervening-if", trigger)
	}
	selection := trigger.InterveningCondition.Val.ControlsMatching
	if !selection.Exists ||
		!slices.Equal(selection.Val.Selection.RequiredTypes, []types.Card{types.Artifact}) {
		t.Fatalf("condition = %+v, want controls an artifact", trigger.InterveningCondition.Val)
	}
}

func TestLowerEnterTriggerRejectsUnsupportedInterveningWording(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Handler",
		Layout:     "normal",
		TypeLine:   "Creature — Elf",
		OracleText: "When this creature enters, if you control an Elf, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("unsupported subtype condition unexpectedly lowered")
	}
}

func TestLowerSagaChapterAbilities(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Saga",
		Layout:     "saga",
		TypeLine:   "Enchantment — Saga",
		OracleText: "I — Draw a card.\nII, III — Draw two cards.",
	})
	if len(face.ChapterAbilities) != 2 {
		t.Fatalf("got %d chapter abilities, want 2", len(face.ChapterAbilities))
	}
	if !slices.Equal(face.ChapterAbilities[0].Chapters, []int{1}) ||
		!slices.Equal(face.ChapterAbilities[1].Chapters, []int{2, 3}) {
		t.Fatalf("chapter numbers = %v, %v", face.ChapterAbilities[0].Chapters, face.ChapterAbilities[1].Chapters)
	}
	draw, ok := face.ChapterAbilities[1].Content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("primitive = %T, want game.Draw", face.ChapterAbilities[1].Content.Modes[0].Sequence[0].Primitive)
	}
	if got := draw.Amount; got != game.Fixed(2) {
		t.Fatalf("draw amount = %#v, want 2", got)
	}
}

func TestLowerReadAheadSaga(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Read Ahead Saga",
		Layout:     "saga",
		TypeLine:   "Enchantment — Saga",
		OracleText: "Read ahead (Choose a chapter and start with that many lore counters. Add one after your draw step. Skipped chapters don't trigger.)\nI — Draw a card.\nII — Draw a card.",
	})
	if len(face.StaticAbilities) != 1 || !game.BodyHasKeyword(face.StaticAbilities[0].Body, game.ReadAhead) {
		t.Fatalf("static abilities = %#v, want ReadAheadStaticBody", face.StaticAbilities)
	}
	if len(face.ChapterAbilities) != 2 {
		t.Fatalf("chapter abilities = %#v, want two", face.ChapterAbilities)
	}
}

func TestLowerReadAheadRejectsNoncanonicalReminder(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Malformed Read Ahead Saga",
		Layout:     "saga",
		TypeLine:   "Enchantment — Saga",
		OracleText: "Read ahead (Choose whichever chapter you want.)\nI — Draw a card.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("noncanonical Read ahead reminder unexpectedly lowered")
	}
}

func TestLowerReadAheadRejectsMismatchedSacrificeChapter(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Mismatched Read Ahead Saga",
		Layout:     "saga",
		TypeLine:   "Enchantment — Saga",
		OracleText: "Read ahead (Choose a chapter and start with that many lore counters. Add one after your draw step. Skipped chapters don't trigger. Sacrifice after IV.)\nI — Draw a card.\nII — Draw a card.\nIII — Draw a card.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("mismatched Read ahead sacrifice chapter unexpectedly lowered")
	}
}

func TestLowerChapterShapedTextRequiresSagaSubtype(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Not a Saga",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "I — Draw a card.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected non-Saga chapter-shaped text to be rejected")
	}
}

func TestOrdinarySagaReminder(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"(As this Saga enters and after your draw step, add a lore counter.)",
		"(As this Saga enters and after your draw step, add a lore counter. Sacrifice after I.)",
		"(As this Saga enters and after your draw step add a lore counter. Sacrifice after III.)",
	} {
		if !isOrdinarySagaReminder(text) {
			t.Errorf("isOrdinarySagaReminder(%q) = false", text)
		}
	}
	for _, text := range []string{
		"Read ahead (Choose a chapter and start with that many lore counters.)",
		"(As this Saga enters and after your draw step, add a lore counter. Sacrifice after VII.)",
		"(As this Saga enters, add a lore counter.)",
	} {
		if isOrdinarySagaReminder(text) {
			t.Errorf("isOrdinarySagaReminder(%q) = true", text)
		}
	}
}

func TestLowerSagaChapterConsumesInlineReminderText(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Saga",
		Layout:     "saga",
		TypeLine:   "Enchantment — Saga",
		OracleText: "I — Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)",
	})
	if len(face.ChapterAbilities) != 1 {
		t.Fatalf("got %d chapter abilities, want 1", len(face.ChapterAbilities))
	}
}

func TestLowerDiesTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, draw two cards.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	if face.TriggeredAbilities[0].Trigger.Pattern.Event != game.EventPermanentDied {
		t.Fatalf("event = %v, want EventPermanentDied", face.TriggeredAbilities[0].Trigger.Pattern.Event)
	}
}

func TestLowerDiesTriggerHadNoPlusPlusCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Undying Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, if it had no +1/+1 counters, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.InterveningIf != "if it had no +1/+1 counters" ||
		!trigger.InterveningIfEventPermanentHadNoCounterKind.Exists ||
		trigger.InterveningIfEventPermanentHadNoCounterKind.Val != counter.PlusOnePlusOne {
		t.Fatalf("trigger = %+v, want no +1/+1 counters intervening-if", trigger)
	}
}

func TestLowerDiesTriggerHadNoMinusMinusCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Persist Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, if it had no -1/-1 counters on it, it deals 3 damage to any target.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ability := face.TriggeredAbilities[0]
	trigger := ability.Trigger
	if trigger.InterveningIf != "if it had no -1/-1 counters on it" ||
		!trigger.InterveningIfEventPermanentHadNoCounterKind.Exists ||
		trigger.InterveningIfEventPermanentHadNoCounterKind.Val != counter.MinusOneMinusOne {
		t.Fatalf("trigger = %+v, want no -1/-1 counters intervening-if", trigger)
	}
	damage, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.Damage)
	if !ok || !damage.DamageSource.Exists ||
		damage.DamageSource.Val != game.EventPermanentReference() {
		t.Fatalf("primitive = %+v, want damage from event permanent", ability.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerDiesTriggerOptional(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, you may draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ability := face.TriggeredAbilities[0]
	if !ability.Optional {
		t.Fatal("dies trigger is not optional")
	}
	if _, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.Draw); !ok {
		t.Fatalf("primitive = %T, want game.Draw", ability.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerDiesTriggerRejectsAmbiguousCounterAbsence(t *testing.T) {
	t.Parallel()
	for _, condition := range []string{
		"if it had no counters on it",
		"if it had no charge counters on it",
		"if it didn't have a +1/+1 counter on it",
	} {
		t.Run(condition, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "When this creature dies, " + condition + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("ambiguous or unsupported condition %q unexpectedly lowered", condition)
			}
		})
	}
}

func TestLowerDiesTriggerReturnsEventCardToOwnersHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, return it to its owner's hand.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	primitive := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive
	move, ok := primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MoveCard", primitive)
	}
	if move.Card.Kind != game.CardReferenceEvent ||
		move.FromZone != zone.Graveyard ||
		move.Destination != zone.Hand {
		t.Fatalf("move = %+v, want event card from graveyard to hand", move)
	}
}

func TestLowerDiesTriggerGrantsAdventureCastFromGraveyard(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:   "Test Dreadknight // Test Whispers",
		Layout: "adventure",
		CardFaces: []ScryfallCardFace{
			{
				Name:       "Test Dreadknight",
				ManaCost:   "{1}{G}",
				TypeLine:   "Creature — Human Knight",
				OracleText: "When Test Dreadknight dies, you may cast it from your graveyard as an Adventure until the end of your next turn.",
				Power:      new("2"),
				Toughness:  new("1"),
			},
			{
				Name:       "Test Whispers",
				ManaCost:   "{1}{B}",
				TypeLine:   "Sorcery — Adventure",
				OracleText: "Draw a card.",
			},
		},
	})
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	ability := faces[0].TriggeredAbilities[0]
	if !ability.Optional {
		t.Fatal("cast-permission dies trigger is not optional")
	}
	primitive := ability.Content.Modes[0].Sequence[0].Primitive
	permission, ok := primitive.(game.GrantCastPermission)
	if !ok {
		t.Fatalf("primitive = %T, want game.GrantCastPermission", primitive)
	}
	if permission.Card.Kind != game.CardReferenceEvent ||
		permission.FromZone != zone.Graveyard ||
		permission.Face != game.FaceAlternate ||
		permission.Duration != game.DurationUntilEndOfYourNextTurn {
		t.Fatalf("permission = %+v, want event Adventure cast through next turn", permission)
	}
}

func TestLowerDiesTriggerRejectsAmbiguousEventCardReference(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"When this creature dies, return it to the battlefield.",
		"When this creature dies, cast it.",
		"When this creature dies, you may cast it from your graveyard.",
		"When this creature dies, return it to its owner's hand or the battlefield.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: text,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("ambiguous event-card reference unexpectedly lowered: %q", text)
			}
		})
	}
}

func TestLowerDiesTriggerRejectsEnterOnlyInterveningConditions(t *testing.T) {
	t.Parallel()
	for _, condition := range []string{
		"if it was kicked",
		"if it was cast",
		"if you cast it",
		"if this creature attacked this turn",
		"if you control an artifact",
	} {
		t.Run(condition, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "When this creature dies, " + condition + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("self-dies trigger unexpectedly lowered with %q", condition)
			}
		})
	}
}

func TestLowerSelfDiesDamageTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Devil",
		Layout:     "normal",
		TypeLine:   "Creature — Devil",
		OracleText: "When this creature dies, it deals 3 damage to any target.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("got %d targets, want 1", len(mode.Targets))
	}
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok ||
		damage.Amount.Value() != 3 ||
		!damage.DamageSource.Exists ||
		damage.DamageSource.Val != game.EventPermanentReference() {
		t.Fatalf("primitive = %+v, want damage from event permanent", mode.Sequence[0].Primitive)
	}
}

func TestLowerManaParameterizedKeywords(t *testing.T) {
	t.Parallel()

	kicker := lowerKeywordForTest(t, "Kicker {1}{G}", game.Kicker)
	kickerKeyword, ok := kicker.(game.KickerKeyword)
	if !ok || kickerKeyword.Cost.String() != "{1}{G}" {
		t.Fatalf("Kicker keyword = %#v, want {1}{G}", kicker)
	}

	madness := lowerKeywordForTest(t, "Madness {2}{B}", game.Madness)
	madnessKeyword, ok := madness.(game.MadnessKeyword)
	if !ok || madnessKeyword.Cost.String() != "{2}{B}" {
		t.Fatalf("Madness keyword = %#v, want {2}{B}", madness)
	}

	morph := lowerKeywordForTest(t, "Morph {3}{U}", game.Morph)
	morphKeyword, ok := morph.(game.MorphKeyword)
	if !ok || morphKeyword.Cost.String() != "{3}{U}" {
		t.Fatalf("Morph keyword = %#v, want {3}{U}", morph)
	}

	disguise := lowerKeywordForTest(t, "Disguise {4}{W}", game.Disguise)
	disguiseKeyword, ok := disguise.(game.DisguiseKeyword)
	if !ok || disguiseKeyword.Cost.String() != "{4}{W}" {
		t.Fatalf("Disguise keyword = %#v, want {4}{W}", disguise)
	}
}

func TestLowerToxicKeyword(t *testing.T) {
	t.Parallel()
	keyword := lowerKeywordForTest(t, "Toxic 2", game.Toxic)
	toxic, ok := keyword.(game.ToxicKeyword)
	if !ok || toxic.Amount != 2 {
		t.Fatalf("Toxic keyword = %#v, want amount 2", keyword)
	}
}

func TestLowerParameterizedKeywordRejectsVariableCost(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Variable Morph",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: "Morph {X}{U}",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "v")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 || diagnostics[0].Summary != "unsupported parameterized keyword" {
		t.Fatalf("diagnostics = %#v, want unsupported parameterized keyword", diagnostics)
	}
}

func lowerKeywordForTest(t *testing.T, oracleText string, kind game.Keyword) game.KeywordAbility {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Parameterized Creature",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: oracleText,
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	keyword, ok := game.BodyKeywordAbility(face.StaticAbilities[0].Body, kind)
	if !ok {
		t.Fatalf("%v keyword not found in %#v", kind, face.StaticAbilities[0].Body)
	}
	return keyword
}

func TestLowerSpellDamage(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bolt",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Test Bolt deals 3 damage to any target.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("got %d targets, want 1", len(mode.Targets))
	}
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %T, want game.Damage", mode.Sequence[0].Primitive)
	}
	if damage.Amount.Value() != 3 {
		t.Fatalf("damage amount = %d, want 3", damage.Amount.Value())
	}
}

func TestLowerSpellDamageQualifiedTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bolt",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Test Bolt deals 3 damage to target attacking or blocking creature.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if got := mode.Targets[0].Predicate.CombatState; got != game.CombatStateAttackingOrBlocking {
		t.Fatalf("combat state = %v, want attacking or blocking", got)
	}
}

func TestLowerSpellXAmounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		cardName   string
		oracleText string
		quantity   func(game.AbilityContent) game.Quantity
	}{
		{
			name:       "damage",
			cardName:   "Test Blaze",
			oracleText: "Test Blaze deals X damage to any target.",
			quantity: func(content game.AbilityContent) game.Quantity {
				primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.Damage)
				if !ok {
					return game.Fixed(0)
				}

				return primitive.Amount
			},
		},
		{
			name:       "draw",
			cardName:   "Test Insight",
			oracleText: "Draw X cards.",
			quantity: func(content game.AbilityContent) game.Quantity {
				primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.Draw)
				if !ok {
					return game.Fixed(0)
				}
				return primitive.Amount
			},
		},
		{
			name:       "life",
			cardName:   "Test Life",
			oracleText: "You gain X life.",
			quantity: func(content game.AbilityContent) game.Quantity {
				primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.GainLife)
				if !ok {
					return game.Fixed(0)
				}
				return primitive.Amount
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       test.cardName,
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			})
			dynamic := test.quantity(face.SpellAbility.Val).DynamicAmount()
			if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountX {
				t.Fatalf("dynamic amount = %+v, want X", dynamic)
			}
		})
	}
}

func TestLowerDynamicEffectAmounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		quantity   func(game.AbilityContent) game.Quantity
		kind       game.DynamicAmountKind
		multiplier int
		cardType   types.Card
		controller game.ControllerRelation
	}{
		{"controlled creatures damage", "Test Swarm deals damage equal to the number of creatures you control to any target.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.Damage)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountCountSelector, 1, types.Creature, game.ControllerYou},
		{"twice battlefield lands damage", "Test Swarm deals damage equal to twice the number of lands on the battlefield to any target.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.Damage)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountCountSelector, 2, types.Land, game.ControllerAny},
		{"life for opponents", "You gain 2 life for each opponent you have.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.GainLife)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountOpponentCount, 2, "", game.ControllerAny},
		{"controller life", "You gain life equal to your life total.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.GainLife)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountControllerLife, 1, "", game.ControllerAny},
		{"draw for controlled lands", "Draw a card for each land you control.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.Draw)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountCountSelector, 1, types.Land, game.ControllerYou},
		{"power for opponents", "Target creature gets +1/+0 for each opponent you have until end of turn.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.ModifyPT)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.PowerDelta
		}, game.DynamicAmountOpponentCount, 1, "", game.ControllerAny},
		{"power after duration", "Target creature gets +1/+0 until end of turn for each opponent you have.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.ModifyPT)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.PowerDelta
		}, game.DynamicAmountOpponentCount, 1, "", game.ControllerAny},
		{"counters for controlled lands", "Put X +1/+1 counters on target creature, where X is the number of lands you control.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.AddCounter)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountCountSelector, 1, types.Land, game.ControllerYou},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Swarm",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			})
			dynamic := test.quantity(face.SpellAbility.Val).DynamicAmount()
			if !dynamic.Exists ||
				dynamic.Val.Kind != test.kind ||
				dynamic.Val.Multiplier != test.multiplier {
				t.Fatalf("dynamic amount = %+v", dynamic)
			}
			if test.cardType != "" {
				selection := dynamic.Val.Group.Selection()
				if len(selection.RequiredTypes) != 1 ||
					selection.RequiredTypes[0] != test.cardType ||
					selection.Controller != test.controller {
					t.Fatalf("selection = %+v", selection)
				}
			}
		})
	}
}

func TestLowerSpellDestroyQualifiedTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Destroy",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Destroy target tapped creature an opponent controls.",
	})
	target := face.SpellAbility.Val.Modes[0].Targets[0]
	if target.Predicate.Tapped != game.TriTrue ||
		target.Predicate.Controller != game.ControllerOpponent {
		t.Fatalf("predicate = %+v, want tapped creature an opponent controls", target.Predicate)
	}
}

func TestLowerMassDestroyAndExile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		selection  game.Selection
		exile      bool
	}{
		{
			name:       "land",
			oracleText: "Destroy all lands.",
			selection:  game.Selection{RequiredTypes: []types.Card{types.Land}},
		},
		{
			name:       "nonland permanent",
			oracleText: "Destroy all nonland permanents.",
			selection:  game.Selection{ExcludedTypes: []types.Card{types.Land}},
		},
		{
			name:       "not controlled by you",
			oracleText: "Destroy all creatures you don't control.",
			selection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Controller:    game.ControllerOpponent,
			},
		},
		{
			name:       "excluded color",
			oracleText: "Destroy all nonwhite creatures.",
			selection: game.Selection{
				RequiredTypes:  []types.Card{types.Creature},
				ExcludedColors: []color.Color{color.White},
			},
		},
		{
			name:       "keyword",
			oracleText: "Destroy all creatures with flying.",
			selection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Keyword:       game.Flying,
			},
		},
		{
			name:       "mana value",
			oracleText: "Destroy all creatures with mana value 3 or less.",
			selection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				ManaValue: opt.Val(compare.Int{
					Op:    compare.LessOrEqual,
					Value: 3,
				}),
			},
		},
		{
			name:       "toughness",
			oracleText: "Destroy all creatures with toughness 4 or greater.",
			selection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Toughness: opt.Val(compare.Int{
					Op:    compare.GreaterOrEqual,
					Value: 4,
				}),
			},
		},
		{
			name:       "other",
			oracleText: "Destroy all other creatures.",
			selection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				ExcludeSource: true,
			},
		},
		{
			name:       "exile",
			oracleText: "Exile all creatures.",
			selection:  game.Selection{RequiredTypes: []types.Card{types.Creature}},
			exile:      true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Mass Effect",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			})
			primitive := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive
			var group game.GroupReference
			switch primitive := primitive.(type) {
			case game.Destroy:
				if test.exile {
					t.Fatalf("primitive = %T, want game.Exile", primitive)
				}
				group = primitive.Group
			case game.Exile:
				if !test.exile {
					t.Fatalf("primitive = %T, want game.Destroy", primitive)
				}
				group = primitive.Group
			default:
				t.Fatalf("primitive = %T, want mass destroy or exile", primitive)
			}
			if group.Domain() != game.GroupDomainBattlefield {
				t.Fatalf("group domain = %v, want battlefield", group.Domain())
			}
			if selection := group.Selection(); !reflect.DeepEqual(selection, test.selection) {
				t.Fatalf("selection = %#v, want %#v", selection, test.selection)
			}
		})
	}
}

func TestParseMassGroupQualifier(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phrase    string
		selection game.Selection
	}{
		{"artifacts, creatures, and enchantments", game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment}}},
		{"tapped creatures", game.Selection{RequiredTypes: []types.Card{types.Creature}, Tapped: game.TriTrue}},
		{"red planeswalkers", game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, ColorsAny: []color.Color{color.Red}}},
		{"nonartifact creatures", game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedTypes: []types.Card{types.Artifact}}},
		{"creatures your opponents control", game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}},
		{"creatures with power equal to 2", game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.Equal, Value: 2})}},
	}
	for _, test := range tests {
		t.Run(test.phrase, func(t *testing.T) {
			t.Parallel()
			selection, ok := parseMassGroupQualifier(test.phrase)
			if !ok {
				t.Fatalf("parseMassGroupQualifier(%q) = false", test.phrase)
			}
			if !reflect.DeepEqual(selection, test.selection) {
				t.Fatalf("selection = %#v, want %#v", selection, test.selection)
			}
		})
	}
	for _, phrase := range []string{
		"creature",
		"all creatures",
		"token creatures",
		"white creatures and lands",
		"creatures with hexproof",
		"creatures with flying you control",
		"untapped creatures",
		"other tapped creatures",
		"nonland",
		"creatures with mana value X or less",
		"creatures with power 3 or more",
		"creatures with flying and reach",
		"creatures controlled by you",
		"creatures except Dragons",
		"nonland cards",
		"white artifacts and creatures",
	} {
		t.Run("reject "+phrase, func(t *testing.T) {
			t.Parallel()
			if selection, ok := parseMassGroupQualifier(phrase); ok {
				t.Fatalf("parseMassGroupQualifier(%q) = %#v, true; want rejection", phrase, selection)
			}
		})
	}
}

func TestLowerSpellReturnQualifiedTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Return",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Return target creature you control to its owner's hand.",
	})
	target := face.SpellAbility.Val.Modes[0].Targets[0]
	if target.Predicate.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", target.Predicate.Controller)
	}
}

func TestLowerSpellModifyPTQualifiedTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Growth",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target untapped creature you control gets +2/+2 until end of turn.",
	})
	target := face.SpellAbility.Val.Modes[0].Targets[0]
	if target.Predicate.Tapped != game.TriFalse ||
		target.Predicate.Controller != game.ControllerYou {
		t.Fatalf("predicate = %+v, want untapped creature you control", target.Predicate)
	}
}

func TestLowerOrderedSpellEffects(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target artifact. Draw a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want one target and two instructions", mode)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.Destroy); !ok {
		t.Fatalf("first primitive = %T, want game.Destroy", mode.Sequence[0].Primitive)
	}
	draw, ok := mode.Sequence[1].Primitive.(game.Draw)
	if !ok || draw.Amount.Value() != 1 {
		t.Fatalf("second primitive = %+v, want draw one", mode.Sequence[1].Primitive)
	}
}

func TestLowerOrderedSpellEffectsWithMultipleTargets(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target artifact. Tap target creature.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 2 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want two targets and two instructions", mode)
	}
	destroy, ok := mode.Sequence[0].Primitive.(game.Destroy)
	if !ok || destroy.Object.TargetIndex() != 0 {
		t.Fatalf("first primitive = %+v, want target 0 destroy", mode.Sequence[0].Primitive)
	}
	tap, ok := mode.Sequence[1].Primitive.(game.Tap)
	if !ok || tap.Object.TargetIndex() != 1 {
		t.Fatalf("second primitive = %+v, want target 1 tap", mode.Sequence[1].Primitive)
	}
}

func TestLowerOrderedSpellEffectsRebasesEveryTargetClause(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target artifact. Tap target creature. Target player mills three cards.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 3 || len(mode.Sequence) != 3 {
		t.Fatalf("mode = %+v, want three targets and three instructions", mode)
	}
	destroy, destroyOK := mode.Sequence[0].Primitive.(game.Destroy)
	tap, tapOK := mode.Sequence[1].Primitive.(game.Tap)
	mill, millOK := mode.Sequence[2].Primitive.(game.Mill)
	if !destroyOK || !tapOK || !millOK {
		t.Fatalf(
			"primitives = %T, %T, %T; want game.Destroy, game.Tap, game.Mill",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
			mode.Sequence[2].Primitive,
		)
	}
	if destroy.Object.TargetIndex() != 0 ||
		tap.Object.TargetIndex() != 1 ||
		mill.Player.TargetIndex() != 2 {
		t.Fatalf(
			"target indices = %d, %d, %d; want 0, 1, 2",
			destroy.Object.TargetIndex(),
			tap.Object.TargetIndex(),
			mill.Player.TargetIndex(),
		)
	}
}

func TestLowerThenJoinedSpellSequence(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		typeLine    string
		oracleText  string
		checkFirst  func(*testing.T, game.Instruction)
		checkSecond func(*testing.T, game.Instruction)
	}{
		{
			name:       "draw then discard spell",
			typeLine:   "Sorcery",
			oracleText: "Draw two cards, then discard a card.",
			checkFirst: func(t *testing.T, inst game.Instruction) {
				draw, ok := inst.Primitive.(game.Draw)
				if !ok || draw.Amount.Value() != 2 || draw.Player != game.ControllerReference() {
					t.Fatalf("first = %+v, want controller draws 2", inst.Primitive)
				}
			},
			checkSecond: func(t *testing.T, inst game.Instruction) {
				discard, ok := inst.Primitive.(game.Discard)
				if !ok || discard.Amount.Value() != 1 || discard.Player != game.ControllerReference() {
					t.Fatalf("second = %+v, want controller discards 1", inst.Primitive)
				}
			},
		},
		{
			name:       "scry then draw spell",
			typeLine:   "Sorcery",
			oracleText: "Scry 2, then draw a card.",
			checkFirst: func(t *testing.T, inst game.Instruction) {
				scry, ok := inst.Primitive.(game.Scry)
				if !ok || scry.Amount.Value() != 2 || scry.Player != game.ControllerReference() {
					t.Fatalf("first = %+v, want controller scries 2", inst.Primitive)
				}
			},
			checkSecond: func(t *testing.T, inst game.Instruction) {
				draw, ok := inst.Primitive.(game.Draw)
				if !ok || draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
					t.Fatalf("second = %+v, want controller draws 1", inst.Primitive)
				}
			},
		},
		{
			name:       "discard then draw spell",
			typeLine:   "Sorcery",
			oracleText: "Discard a card, then draw a card.",
			checkFirst: func(t *testing.T, inst game.Instruction) {
				discard, ok := inst.Primitive.(game.Discard)
				if !ok || discard.Amount.Value() != 1 || discard.Player != game.ControllerReference() {
					t.Fatalf("first = %+v, want controller discards 1", inst.Primitive)
				}
			},
			checkSecond: func(t *testing.T, inst game.Instruction) {
				draw, ok := inst.Primitive.(game.Draw)
				if !ok || draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
					t.Fatalf("second = %+v, want controller draws 1", inst.Primitive)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Spell",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			})
			if !face.SpellAbility.Exists {
				t.Fatal("spell ability not lowered")
			}
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 0 || len(mode.Sequence) != 2 {
				t.Fatalf("mode = %+v, want no targets and two instructions", mode)
			}
			test.checkFirst(t, mode.Sequence[0])
			test.checkSecond(t, mode.Sequence[1])
		})
	}
}

func TestLowerThenJoinedEnterTriggerSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Looting Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, draw a card, then discard a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want no targets and two instructions", mode)
	}
	draw, drawOK := mode.Sequence[0].Primitive.(game.Draw)
	discard, discardOK := mode.Sequence[1].Primitive.(game.Discard)
	if !drawOK || !discardOK {
		t.Fatalf(
			"primitives = %T, %T; want game.Draw, game.Discard",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
		)
	}
	if draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
		t.Fatalf("draw = %+v, want controller draws 1", draw)
	}
	if discard.Amount.Value() != 1 || discard.Player != game.ControllerReference() {
		t.Fatalf("discard = %+v, want controller discards 1", discard)
	}
}

func TestLowerThenJoinedSharedTargetSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mill",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target player mills three cards, then draws a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want one target and two instructions", mode)
	}
	mill, millOK := mode.Sequence[0].Primitive.(game.Mill)
	draw, drawOK := mode.Sequence[1].Primitive.(game.Draw)
	if !millOK || !drawOK {
		t.Fatalf(
			"primitives = %T, %T; want game.Mill, game.Draw",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
		)
	}
	if mill.Amount.Value() != 3 || mill.Player.TargetIndex() != 0 {
		t.Fatalf("mill = %+v, want target player mills 3", mill)
	}
	if draw.Amount.Value() != 1 || draw.Player.TargetIndex() != 0 {
		t.Fatalf("draw = %+v, want target player draws 1", draw)
	}
}

// TestLowerThenJoinedThreeEffectSequence is a regression for a bug where
// 3-effect then-joined chains would assign the wrong clause start for
// effects after the first in the group, causing middle clauses to
// incorrectly include previous effects' tokens.
func TestLowerThenJoinedThreeEffectSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Chain",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Draw a card, then discard a card, then proliferate.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 3 {
		t.Fatalf("mode = %+v, want no targets and three instructions", mode)
	}
	draw, drawOK := mode.Sequence[0].Primitive.(game.Draw)
	discard, discardOK := mode.Sequence[1].Primitive.(game.Discard)
	_, prolifOK := mode.Sequence[2].Primitive.(game.Proliferate)
	if !drawOK || !discardOK || !prolifOK {
		t.Fatalf(
			"primitives = %T, %T, %T; want game.Draw, game.Discard, game.Proliferate",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
			mode.Sequence[2].Primitive,
		)
	}
	if draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
		t.Fatalf("draw = %+v, want controller draws 1", draw)
	}
	if discard.Amount.Value() != 1 || discard.Player != game.ControllerReference() {
		t.Fatalf("discard = %+v, want controller discards 1", discard)
	}
}

// TestLowerThenJoinedNonTargetFinalClause is a regression for the case where
// a then-joined sentence is followed by a separate sentence: the final
// clause of the then-group must be bounded to its own sentence and must not
// spill into subsequent-sentence tokens.
func TestLowerThenJoinedNonTargetFinalClause(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Multi",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Draw a card, then discard a card. You gain 3 life.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 3 {
		t.Fatalf("mode = %+v, want no targets and three instructions", mode)
	}
	draw, drawOK := mode.Sequence[0].Primitive.(game.Draw)
	discard, discardOK := mode.Sequence[1].Primitive.(game.Discard)
	gain, gainOK := mode.Sequence[2].Primitive.(game.GainLife)
	if !drawOK || !discardOK || !gainOK {
		t.Fatalf(
			"primitives = %T, %T, %T; want game.Draw, game.Discard, game.GainLife",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
			mode.Sequence[2].Primitive,
		)
	}
	if draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
		t.Fatalf("draw = %+v, want controller draws 1", draw)
	}
	if discard.Amount.Value() != 1 || discard.Player != game.ControllerReference() {
		t.Fatalf("discard = %+v, want controller discards 1", discard)
	}
	if gain.Amount.Value() != 3 || gain.Player != game.ControllerReference() {
		t.Fatalf("gain = %+v, want controller gains 3", gain)
	}
}

// TestLowerThenJoinedSharedTargetNoExtraSpec is a regression for the target
// deduplication requirement: a shared-subject then-joined sequence
// (e.g. "Target player mills N, then draws M") must produce exactly one
// game.TargetSpec, and both instructions must reference TargetIndex 0.
func TestLowerThenJoinedSharedTargetNoExtraSpec(t *testing.T) {
	t.Parallel()
	// Verify that compound-mill produces exactly one target spec and both
	// instructions reference the same target player at index 0.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Shared Target Test",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target player mills three cards, then draws a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want exactly 1 (no duplicate target spec)", len(mode.Targets))
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want 2", len(mode.Sequence))
	}
	mill, millOK := mode.Sequence[0].Primitive.(game.Mill)
	draw, drawOK := mode.Sequence[1].Primitive.(game.Draw)
	if !millOK || !drawOK {
		t.Fatalf("primitives = %T, %T, want game.Mill, game.Draw",
			mode.Sequence[0].Primitive, mode.Sequence[1].Primitive)
	}
	if mill.Player.TargetIndex() != 0 {
		t.Fatalf("mill.Player target index = %d, want 0", mill.Player.TargetIndex())
	}
	if draw.Player.TargetIndex() != 0 {
		t.Fatalf("draw.Player target index = %d, want 0 (reusing existing target)", draw.Player.TargetIndex())
	}
}

// TestLowerThenJoinedDestroyThenProliferate verifies that a then-joined pair
// where the second effect does not use the shared target (proliferate has no
// target) correctly discards the spurious shared target via the fallback
// path, producing one target spec for destroy and a standalone proliferate.
func TestLowerThenJoinedDestroyThenProliferate(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Spread",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target creature, then proliferate.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1 (destroy target only, no duplicate)", len(mode.Targets))
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want 2", len(mode.Sequence))
	}
	destroy, destroyOK := mode.Sequence[0].Primitive.(game.Destroy)
	_, prolifOK := mode.Sequence[1].Primitive.(game.Proliferate)
	if !destroyOK || !prolifOK {
		t.Fatalf("primitives = %T, %T, want game.Destroy, game.Proliferate",
			mode.Sequence[0].Primitive, mode.Sequence[1].Primitive)
	}
	if destroy.Object.TargetIndex() != 0 {
		t.Fatalf("destroy.Object target index = %d, want 0", destroy.Object.TargetIndex())
	}
}

func TestLowerThenJoinedActivatedAbilitySequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tome",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{2}, {T}: Draw a card, then discard a card.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("got %d activated abilities, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want no targets and two instructions", mode)
	}
	draw, drawOK := mode.Sequence[0].Primitive.(game.Draw)
	discard, discardOK := mode.Sequence[1].Primitive.(game.Discard)
	if !drawOK || !discardOK {
		t.Fatalf(
			"primitives = %T, %T; want game.Draw, game.Discard",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
		)
	}
	if draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
		t.Fatalf("draw = %+v, want controller draws 1", draw)
	}
	if discard.Amount.Value() != 1 || discard.Player != game.ControllerReference() {
		t.Fatalf("discard = %+v, want controller discards 1", discard)
	}
}

func TestLowerThenJoinedLoyaltyAbilitySequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: "+1: Scry 1, then draw a card.",
		Loyalty:    func() *string { s := "3"; return &s }(),
	})
	if len(face.LoyaltyAbilities) != 1 {
		t.Fatalf("got %d loyalty abilities, want 1", len(face.LoyaltyAbilities))
	}
	mode := face.LoyaltyAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want no targets and two instructions", mode)
	}
	scry, scryOK := mode.Sequence[0].Primitive.(game.Scry)
	draw, drawOK := mode.Sequence[1].Primitive.(game.Draw)
	if !scryOK || !drawOK {
		t.Fatalf(
			"primitives = %T, %T; want game.Scry, game.Draw",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
		)
	}
	if scry.Amount.Value() != 1 || scry.Player != game.ControllerReference() {
		t.Fatalf("scry = %+v, want controller scries 1", scry)
	}
	if draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
		t.Fatalf("draw = %+v, want controller draws 1", draw)
	}
}

func TestLowerThenJoinedSagaChapterSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Saga",
		Layout:     "saga",
		TypeLine:   "Enchantment — Saga",
		OracleText: "I, II — Scry 2, then draw a card.\nIII — Draw two cards.",
	})
	if len(face.ChapterAbilities) != 2 {
		t.Fatalf("got %d chapter abilities, want 2", len(face.ChapterAbilities))
	}
	mode := face.ChapterAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 2 {
		t.Fatalf("chapter I/II mode = %+v, want no targets and two instructions", mode)
	}
	scry, scryOK := mode.Sequence[0].Primitive.(game.Scry)
	draw, drawOK := mode.Sequence[1].Primitive.(game.Draw)
	if !scryOK || !drawOK {
		t.Fatalf(
			"primitives = %T, %T; want game.Scry, game.Draw",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
		)
	}
	if scry.Amount.Value() != 2 || scry.Player != game.ControllerReference() {
		t.Fatalf("scry = %+v, want controller scries 2", scry)
	}
	if draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
		t.Fatalf("draw = %+v, want controller draws 1", draw)
	}
}

// TestCompoundMillOracleIR documents the oracle compiler IR for the
// shared-subject then-joined pattern ("Target player mills three cards, then
// draws a card.") and proves that compound mill is achievable within the scope
// of issue #131 without additional effect kinds.
//
// Hypothesis verified: the oracle compiler emits exactly one CompiledTarget
// ("target player") for the sentence; it does NOT create a second implicit
// target for the "draws" clause. The second effect's subject is implied, not
// independently emitted. lowerOrderedEffectSequence resolves this through the
// shared-target deduplication path: abilityForEffect uses the sentence Span for
// both effects (finding the one target for each), allOracleTargetSpansClaimed
// recognises the second claim as a duplicate, and rebaseTargetedSequence with
// offset 0 correctly produces TargetPlayerReference(0) for both instructions
// without adding a duplicate game.TargetSpec.
func TestCompoundMillOracleIR(t *testing.T) {
	t.Parallel()
	const text = "Target player mills three cards, then draws a card."
	compilation, diags := oracle.Compile(text, oracle.ParseContext{CardName: "Test Mill"})
	if len(diags) > 0 {
		t.Fatalf("compile diagnostics: %v", diags)
	}
	if len(compilation.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(compilation.Abilities))
	}
	ab := compilation.Abilities[0]

	// Two effects with the same sentence Span — the root condition that
	// requires the then-join split.
	if len(ab.Effects) != 2 {
		t.Fatalf("IR effects = %d, want 2 (mills + draws)", len(ab.Effects))
	}
	if ab.Effects[0].Kind != oracle.EffectMill {
		t.Fatalf("effect[0].Kind = %v, want EffectMill", ab.Effects[0].Kind)
	}
	if ab.Effects[1].Kind != oracle.EffectDraw {
		t.Fatalf("effect[1].Kind = %v, want EffectDraw", ab.Effects[1].Kind)
	}
	if ab.Effects[0].Span != ab.Effects[1].Span {
		t.Fatalf("effect spans differ: %+v vs %+v; want same sentence span",
			ab.Effects[0].Span, ab.Effects[1].Span)
	}

	// Verb spans are at distinct offsets, enabling the split to locate each
	// clause boundary precisely.
	if ab.Effects[0].VerbSpan == ab.Effects[1].VerbSpan {
		t.Fatal("verb spans equal; want mills ≠ draws")
	}

	// Exactly one target ("target player") in the IR. The compiler does not
	// emit a separate target for the implied "draws" subject.
	if len(ab.Targets) != 1 {
		t.Fatalf("IR targets = %d, want 1 (shared; not duplicated for draws clause)", len(ab.Targets))
	}
	if ab.Targets[0].Selector.Kind != oracle.SelectorPlayer {
		t.Fatalf("target selector = %v, want SelectorPlayer", ab.Targets[0].Selector.Kind)
	}

	// End-to-end: compound mill lowers successfully with no diagnostics.
	card := &ScryfallCard{
		Name: "Test Mill", Layout: "normal", TypeLine: "Sorcery", OracleText: text,
	}
	_, execDiags, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(execDiags) != 0 {
		t.Fatalf("executable diagnostics: %v", execDiags)
	}
	fmt.Printf("compound mill IR: effects=%d same-span=%v verb-spans-distinct=%v targets=%d\n",
		len(ab.Effects),
		ab.Effects[0].Span == ab.Effects[1].Span,
		ab.Effects[0].VerbSpan != ab.Effects[1].VerbSpan,
		len(ab.Targets),
	)
}


func TestLowerSurveilSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Surveil",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Surveil 2. (Look at the top two cards of your library, then put any number of them into your graveyard and the rest on top of your library in any order.)",
	})
	mode := face.SpellAbility.Val.Modes[0]
	surveil, ok := mode.Sequence[0].Primitive.(game.Surveil)
	if !ok ||
		surveil.Amount.Value() != 2 ||
		surveil.Player != game.ControllerReference() {
		t.Fatalf("primitive = %+v, want controller surveils two", mode.Sequence[0].Primitive)
	}
}

func TestLowerInvestigateSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Investigate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Investigate.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	investigate, ok := mode.Sequence[0].Primitive.(game.Investigate)
	if !ok || investigate.Amount.Value() != 1 {
		t.Fatalf("primitive = %+v, want investigate once", mode.Sequence[0].Primitive)
	}
}

func TestLowerInvestigateTwiceSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Investigate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Investigate twice.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	investigate, ok := mode.Sequence[0].Primitive.(game.Investigate)
	if !ok || investigate.Amount.Value() != 2 {
		t.Fatalf("primitive = %+v, want investigate twice", mode.Sequence[0].Primitive)
	}
}

func TestLowerProliferateSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Proliferate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Proliferate.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if _, ok := mode.Sequence[0].Primitive.(game.Proliferate); !ok {
		t.Fatalf("primitive = %T, want game.Proliferate", mode.Sequence[0].Primitive)
	}
}

func TestLowerProliferateTwiceSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Proliferate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Proliferate twice.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	proliferate, ok := mode.Sequence[0].Primitive.(game.Proliferate)
	if !ok || proliferate.Amount.Value() != 2 {
		t.Fatalf("primitive = %+v, want proliferate twice", mode.Sequence[0].Primitive)
	}
}

func TestLowerExploreSourcePermanentTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Scout",
		Layout:     "normal",
		TypeLine:   "Creature — Merfolk Scout",
		OracleText: "When this creature enters, it explores.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	explore, ok := mode.Sequence[0].Primitive.(game.Explore)
	if !ok || explore.Creature.Kind() != game.ObjectReferenceSourcePermanent {
		t.Fatalf("primitive = %+v, want source permanent explores", mode.Sequence[0].Primitive)
	}
}

func TestLowerExploreRejectsUnsupportedTargets(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Explore",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target creature explores.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported explore diagnostic")
	}
}

func TestLowerManifestSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Manifest",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Manifest the top card of your library.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	manifest, ok := mode.Sequence[0].Primitive.(game.Manifest)
	if !ok {
		t.Fatalf("primitive = %T, want game.Manifest", mode.Sequence[0].Primitive)
	}
	if manifest.Dread {
		t.Fatal("basic manifest lowered with Dread=true")
	}
}

func TestLowerManifestDreadSpell(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
	}{
		{
			name:   "shorthand",
			oracle: "Manifest Dread.",
		},
		{
			name:   "long form",
			oracle: "Look at the top two cards of your library. Put one of them onto the battlefield face down as a 2/2 creature. Put the other into your graveyard.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Manifest Dread",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracle,
			})
			mode := face.SpellAbility.Val.Modes[0]
			manifest, ok := mode.Sequence[0].Primitive.(game.Manifest)
			if !ok {
				t.Fatalf("primitive = %T, want game.Manifest", mode.Sequence[0].Primitive)
			}
			if !manifest.Dread {
				t.Fatal("manifest dread lowered with Dread=false")
			}
		})
	}
}

func TestLowerManifestRejectsUnsupportedPatterns(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Manifest",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Manifest a card from your hand.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported manifest diagnostic")
	}
}

func TestLowerInterveningTriggerUtilityKeywordBodies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		text      string
		primitive any
	}{
		{
			name:      "scry",
			text:      "When this creature enters, if you control an artifact, scry 2.",
			primitive: game.Scry{Amount: game.Fixed(2), Player: game.ControllerReference()},
		},
		{
			name:      "investigate",
			text:      "When this creature enters, if you control an artifact, investigate.",
			primitive: game.Investigate{Amount: game.Fixed(1)},
		},
		{
			name:      "proliferate",
			text:      "When this creature enters, if you control an artifact, proliferate.",
			primitive: game.Proliferate{Amount: game.Fixed(1)},
		},
		{
			name:      "explore",
			text:      "When this creature enters, if you control an artifact, it explores.",
			primitive: game.Explore{Creature: game.SourcePermanentReference()},
		},
		{
			name:      "manifest",
			text:      "When this creature enters, if you control an artifact, manifest the top card of your library.",
			primitive: game.Manifest{},
		},
		{
			name:      "mill",
			text:      "When this creature enters, if you control an artifact, mill two cards.",
			primitive: game.Mill{Amount: game.Fixed(2), Player: game.ControllerReference()},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Utility",
				Layout:     "normal",
				TypeLine:   "Creature — Human Wizard",
				OracleText: tc.text,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			got := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive
			if !reflect.DeepEqual(got, tc.primitive) {
				t.Fatalf("primitive = %+v, want %+v", got, tc.primitive)
			}
		})
	}
}

func TestLowerVariableMillSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mill",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Mill X cards, where X is the number of creatures you control.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	mill, ok := mode.Sequence[0].Primitive.(game.Mill)
	if !ok {
		t.Fatalf("primitive = %T, want game.Mill", mode.Sequence[0].Primitive)
	}
	dynamic := mill.Amount.DynamicAmount()
	if !dynamic.Exists {
		t.Fatalf("mill amount = %+v, want dynamic controlled creature count", mill.Amount)
	}
	selection := dynamic.Val.Group.Selection()
	if dynamic.Val.Kind != game.DynamicAmountCountSelector ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature ||
		selection.Controller != game.ControllerYou {
		t.Fatalf("mill amount = %+v, want dynamic controlled creature count", mill.Amount)
	}
}

func TestLowerFixedCounterSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Counter",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Put two +1/+1 counters on target creature.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].Predicate.PermanentTypes[0] != types.Creature {
		t.Fatalf("targets = %+v, want one creature target", mode.Targets)
	}
	add, ok := mode.Sequence[0].Primitive.(game.AddCounter)
	if !ok ||
		add.Amount.Value() != 2 ||
		add.CounterKind != counter.PlusOnePlusOne ||
		add.Object != game.TargetPermanentReference(0) {
		t.Fatalf("primitive = %+v, want two +1/+1 counters on target 0", mode.Sequence[0].Primitive)
	}
}

func TestLowerNamedCounterPlacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text string
		kind counter.Kind
	}{
		{"Put a charge counter on target artifact.", counter.Charge},
		{"Put two shield counters on target creature you control.", counter.Shield},
		{"Put a first strike counter on target creature.", counter.FirstStrike},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.text,
			})
			add, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.AddCounter)
			if !ok || add.CounterKind != test.kind {
				t.Fatalf("primitive = %+v", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
			}
		})
	}
}

func TestLowerKeywordNamedCounterPlacementAbilityShapes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		card    *ScryfallCard
		content func(loweredFaceAbilities) (game.AbilityContent, bool)
		kind    counter.Kind
	}{
		{
			name: "activated",
			card: &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: "{T}: Put a flying counter on target creature.",
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.ActivatedAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.ActivatedAbilities[0].Content, true
			},
			kind: counter.Flying,
		},
		{
			name: "loyalty",
			card: &ScryfallCard{
				Name:       "Test Walker",
				Layout:     "normal",
				TypeLine:   "Legendary Planeswalker — Test",
				OracleText: "+1: Put a lifelink counter on target creature.",
				Loyalty:    func() *string { loyalty := "3"; return &loyalty }(),
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.LoyaltyAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.LoyaltyAbilities[0].Content, true
			},
			kind: counter.Lifelink,
		},
		{
			name: "triggered",
			card: &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Creature — Test",
				OracleText: "When this creature enters, put a first strike counter on target creature.",
				Power:      new("2"),
				Toughness:  new("2"),
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.TriggeredAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.TriggeredAbilities[0].Content, true
			},
			kind: counter.FirstStrike,
		},
		{
			name: "phase triggered",
			card: &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: "At the beginning of your upkeep, put a flying counter on target creature.",
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.TriggeredAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.TriggeredAbilities[0].Content, true
			},
			kind: counter.Flying,
		},
		{
			name: "non-self enter triggered",
			card: &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: "Whenever another creature enters, put a lifelink counter on target creature.",
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.TriggeredAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.TriggeredAbilities[0].Content, true
			},
			kind: counter.Lifelink,
		},
		{
			name: "ordered effects",
			card: &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: "Put a flying counter on target creature. Draw a card.",
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if !face.SpellAbility.Exists {
					return game.AbilityContent{}, false
				}
				return face.SpellAbility.Val, true
			},
			kind: counter.Flying,
		},
		{
			name: "Saga chapter",
			card: &ScryfallCard{
				Name:       "Test Saga",
				Layout:     "saga",
				TypeLine:   "Enchantment — Saga",
				OracleText: "I — Put a lifelink counter on target creature.",
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.ChapterAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.ChapterAbilities[0].Content, true
			},
			kind: counter.Lifelink,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, test.card)
			content, ok := test.content(face)
			if !ok ||
				len(content.Modes) != 1 ||
				len(content.Modes[0].Sequence) == 0 {
				t.Fatalf("face = %+v, want lowered counter placement", face)
			}
			add, ok := content.Modes[0].Sequence[0].Primitive.(game.AddCounter)
			if !ok || add.CounterKind != test.kind {
				t.Fatalf("primitive = %+v, want %s counter placement", content.Modes[0].Sequence[0].Primitive, test.kind)
			}
		})
	}
}

func TestLowerPlayerCounterPlacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text string
		kind counter.Kind
	}{
		{"Put an energy counter on target player.", counter.Energy},
		{"Put two experience counters on target player.", counter.Experience},
		{"Put three poison counters on target opponent.", counter.Poison},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.text,
			})
			mode := face.SpellAbility.Val.Modes[0]
			add, ok := mode.Sequence[0].Primitive.(game.AddPlayerCounter)
			if !ok ||
				add.CounterKind != test.kind ||
				add.Player != game.TargetPlayerReference(0) ||
				mode.Targets[0].Allow != game.TargetAllowPlayer {
				t.Fatalf("mode = %+v", mode)
			}
		})
	}
}

func TestLowerEveryRecognizedCounterKindOnItsValidTarget(t *testing.T) {
	t.Parallel()
	for kind := counter.PlusOnePlusOne; kind <= counter.Experience; kind++ {
		if kind == counter.Stun || kind == counter.Finality {
			continue
		}
		if !oracle.CounterKindPlacementSupported(kind) {
			t.Fatalf("%s unexpectedly excluded from named placement", kind)
		}
		name := kind.String()
		article := "a"
		if strings.ContainsRune("aeiou", rune(name[0])) {
			article = "an"
		}
		target := "target permanent"
		if kind.PlayerOnly() {
			target = "target player"
		}
		oracleText := fmt.Sprintf("Put %s %s counter on %s.", article, name, target)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: oracleText,
			})
			primitive := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive
			if kind.PlayerOnly() {
				add, ok := primitive.(game.AddPlayerCounter)
				if !ok || add.CounterKind != kind {
					t.Fatalf("primitive = %+v", primitive)
				}
				return
			}
			add, ok := primitive.(game.AddCounter)
			if !ok || add.CounterKind != kind {
				t.Fatalf("primitive = %+v", primitive)
			}
		})
	}
}

func TestLowerCounterPlacementRejectsMissingRuntimeMechanics(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		kind counter.Kind
	}{
		{"stun", counter.Stun},
		{"finality", counter.Finality},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ability := oracle.CompiledAbility{
				Text: "Put a " + test.name + " counter on target creature.",
				Targets: []oracle.CompiledTarget{{
					Text:        "target creature",
					Cardinality: oracle.TargetCardinality{Min: 1, Max: 1},
					Selector:    oracle.CompiledSelector{Kind: oracle.SelectorCreature},
				}},
				Effects: []oracle.CompiledEffect{{
					Kind:             oracle.EffectPut,
					Amount:           oracle.CompiledAmount{Value: 1, Known: true},
					CounterKind:      test.kind,
					CounterKindKnown: true,
				}},
			}
			if _, diagnostic := lowerCounterPlacementSpell(ability); diagnostic == nil {
				t.Fatal("lowering accepted counter kind without runtime mechanics")
			}
		})
	}
}

func TestLowerDynamicNamedCounterPlacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text string
		kind game.DynamicAmountKind
	}{
		{"Put X charge counters on target artifact.", game.DynamicAmountX},
		{"Put X poison counters on target player, where X is the number of lands you control.", game.DynamicAmountCountSelector},
		{"Put X energy counters on target player, where X is Test Counter's power.", game.DynamicAmountObjectPower},
	}
	for _, test := range tests {
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test Counter",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: test.text,
		})
		primitive := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive
		amount, ok := counterPlacementAmount(primitive)
		if !ok {
			t.Fatalf("%q primitive = %T", test.text, primitive)
		}
		dynamic := amount.DynamicAmount()
		if !dynamic.Exists || dynamic.Val.Kind != test.kind {
			t.Fatalf("%q amount = %+v", test.text, dynamic)
		}
		if test.kind == game.DynamicAmountObjectPower &&
			dynamic.Val.Object != game.SourcePermanentReference() {
			t.Fatalf("%q source reference = %+v", test.text, dynamic.Val.Object)
		}
	}

}

func counterPlacementAmount(primitive game.Primitive) (game.Quantity, bool) {
	switch primitive.Kind() {
	case game.PrimitiveAddCounter:
		add, ok := primitive.(game.AddCounter)
		return add.Amount, ok
	case game.PrimitiveAddPlayerCounter:
		add, ok := primitive.(game.AddPlayerCounter)
		return add.Amount, ok
	default:
		return game.Quantity{}, false
	}
}

func TestRebaseAddPlayerCounterTargetReference(t *testing.T) {
	t.Parallel()
	primitive, ok := rebaseTargetedPrimitive(game.AddPlayerCounter{
		Amount:      game.Fixed(1),
		Player:      game.TargetPlayerReference(0),
		CounterKind: counter.Poison,
	}, 2)
	if !ok {
		t.Fatal("AddPlayerCounter target was not rebased")
	}
	add, ok := primitive.(game.AddPlayerCounter)
	if !ok || add.Player != game.TargetPlayerReference(2) {
		t.Fatalf("rebased primitive = %+v", primitive)
	}
}

func TestLowerCounterPlacementRejectsUnsupportedForms(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Put a quest counter on target permanent.",
		"Put an energy counter on target creature.",
		"Put a charge counter on target player.",
		"Put a charge counter on any target.",
		"Put a +1/+1 counter on each creature you control.",
		"Put a charge and time counter on target artifact.",
		"Put 0 charge counters on target artifact.",
		"Put -1 charge counters on target artifact.",
		"Put a charge counter on target artifact for each land you control.",
	} {
		_, diagnostics := lowerExecutableFaces(&ScryfallCard{
			Name:       "Test Counter",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: oracleText,
		})
		if len(diagnostics) == 0 {
			t.Fatalf("%q lowered without diagnostics", oracleText)
		}
	}
}

func TestLowerRegenerateSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Regenerate",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Regenerate target creature.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	regenerate, ok := mode.Sequence[0].Primitive.(game.Regenerate)
	if !ok || regenerate.Object != game.TargetPermanentReference(0) {
		t.Fatalf("primitive = %+v, want regenerate target permanent", mode.Sequence[0].Primitive)
	}
}

func TestLowerFightSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Fight",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target creature you control fights target creature you don't control.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 2 {
		t.Fatalf("targets = %+v, want two creatures", mode.Targets)
	}
	fight, ok := mode.Sequence[0].Primitive.(game.Fight)
	if !ok ||
		fight.Object != game.TargetPermanentReference(0) ||
		fight.RelatedObject != game.TargetPermanentReference(1) {
		t.Fatalf("primitive = %+v, want targets 0 and 1 fight", mode.Sequence[0].Primitive)
	}
}

func TestLowerLoyaltyAbilityPositiveCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: "+1: Draw a card.",
		Loyalty:    func() *string { s := "3"; return &s }(),
	})
	if len(face.LoyaltyAbilities) != 1 {
		t.Fatalf("got %d loyalty abilities, want 1", len(face.LoyaltyAbilities))
	}
	la := face.LoyaltyAbilities[0]
	if la.LoyaltyCost != 1 {
		t.Fatalf("LoyaltyCost = %d, want 1", la.LoyaltyCost)
	}
	if la.Content.IsModal() || len(la.Content.Modes) != 1 {
		t.Fatalf("content = %+v, want single non-modal mode", la.Content)
	}
	draw, ok := la.Content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
		t.Fatalf("primitive = %+v, want controller draws one", la.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerLoyaltyAbilityNegativeCost(t *testing.T) {
	t.Parallel()
	loyaltyText := "\u22122: Target player mills three cards."
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: loyaltyText,
		Loyalty:    func() *string { s := "4"; return &s }(),
	})
	if len(face.LoyaltyAbilities) != 1 {
		t.Fatalf("got %d loyalty abilities, want 1", len(face.LoyaltyAbilities))
	}
	la := face.LoyaltyAbilities[0]
	if la.LoyaltyCost != -2 {
		t.Fatalf("LoyaltyCost = %d, want -2", la.LoyaltyCost)
	}
	mill, ok := la.Content.Modes[0].Sequence[0].Primitive.(game.Mill)
	if !ok || mill.Amount.Value() != 3 {
		t.Fatalf("primitive = %+v, want mills three", la.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerLoyaltyAbilityZeroCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: "0: Scry 2.",
		Loyalty:    func() *string { s := "3"; return &s }(),
	})
	if len(face.LoyaltyAbilities) != 1 {
		t.Fatalf("got %d loyalty abilities, want 1", len(face.LoyaltyAbilities))
	}
	la := face.LoyaltyAbilities[0]
	if la.LoyaltyCost != 0 {
		t.Fatalf("LoyaltyCost = %d, want 0", la.LoyaltyCost)
	}
	scry, ok := la.Content.Modes[0].Sequence[0].Primitive.(game.Scry)
	if !ok || scry.Amount.Value() != 2 {
		t.Fatalf("primitive = %+v, want scry two", la.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerLoyaltyAbilityMultiple(t *testing.T) {
	t.Parallel()
	oracleText := "+1: Draw a card.\n\u22122: You lose 3 life."
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: oracleText,
		Loyalty:    func() *string { s := "3"; return &s }(),
	})
	if len(face.LoyaltyAbilities) != 2 {
		t.Fatalf("got %d loyalty abilities, want 2", len(face.LoyaltyAbilities))
	}
	if face.LoyaltyAbilities[0].LoyaltyCost != 1 {
		t.Fatalf("first LoyaltyCost = %d, want 1", face.LoyaltyAbilities[0].LoyaltyCost)
	}
	if face.LoyaltyAbilities[1].LoyaltyCost != -2 {
		t.Fatalf("second LoyaltyCost = %d, want -2", face.LoyaltyAbilities[1].LoyaltyCost)
	}
}

func TestLowerLoyaltyAbilityVariableCostRejected(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: "\u2212X: Target player mills X cards.",
		Loyalty:    func() *string { s := "3"; return &s }(),
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for variable loyalty cost, got none")
	}
}

func TestLowerModalChooseOneSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Charm",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Choose one \u2014\n\u2022 Draw a card.\n\u2022 You gain 3 life.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("no spell ability")
	}
	content := face.SpellAbility.Val
	if !content.IsModal() {
		t.Fatal("spell ability is not modal")
	}
	if len(content.Modes) != 2 {
		t.Fatalf("got %d modes, want 2", len(content.Modes))
	}
	if content.MinModes != 1 || content.MaxModes != 1 {
		t.Fatalf("MinModes=%d MaxModes=%d, want both 1", content.MinModes, content.MaxModes)
	}
	draw, ok := content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount.Value() != 1 {
		t.Fatalf("mode 0 primitive = %+v, want draw one", content.Modes[0].Sequence[0].Primitive)
	}
	gain, ok := content.Modes[1].Sequence[0].Primitive.(game.GainLife)
	if !ok || gain.Amount.Value() != 3 {
		t.Fatalf("mode 1 primitive = %+v, want gain 3 life", content.Modes[1].Sequence[0].Primitive)
	}
}

func TestLowerModalChooseOneWithTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Charm",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Choose one \u2014\n\u2022 Destroy target artifact.\n\u2022 Draw a card.",
	})
	content := face.SpellAbility.Val
	if !content.IsModal() || len(content.Modes) != 2 {
		t.Fatalf("content = %+v, want modal with 2 modes", content)
	}
	if len(content.Modes[0].Targets) != 1 {
		t.Fatalf("mode 0 targets = %+v, want one target", content.Modes[0].Targets)
	}
	if _, ok := content.Modes[0].Sequence[0].Primitive.(game.Destroy); !ok {
		t.Fatalf("mode 0 primitive = %T, want game.Destroy", content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerModalChooseTwoSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Command",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Choose two \u2014\n\u2022 Draw a card.\n\u2022 You gain 3 life.\n\u2022 Proliferate.",
	})
	content := face.SpellAbility.Val
	if content.MinModes != 2 || content.MaxModes != 2 {
		t.Fatalf("MinModes=%d MaxModes=%d, want both 2", content.MinModes, content.MaxModes)
	}
	if len(content.Modes) != 3 {
		t.Fatalf("got %d modes, want 3", len(content.Modes))
	}
}

func TestLowerModalChooseOneOrBoth(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Charm",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Choose one or both \u2014\n\u2022 Draw a card.\n\u2022 You gain 3 life.",
	})
	content := face.SpellAbility.Val
	if content.MinModes != 1 || content.MaxModes != 2 {
		t.Fatalf("MinModes=%d MaxModes=%d, want 1 and 2", content.MinModes, content.MaxModes)
	}
	if len(content.Modes) != 2 {
		t.Fatalf("got %d modes, want 2", len(content.Modes))
	}
}

func TestLowerModalChoiceCountExceedsModesRejected(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Command",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Choose three \u2014\n\u2022 Draw a card.\n\u2022 You gain 3 life.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics when choice count exceeds modes, got none")
	}
}

func TestLowerModalUnsupportedModeBodyRejected(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Charm",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Choose one \u2014\n\u2022 Draw a card.\n\u2022 Search your library for a card.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for unsupported mode body, got none")
	}
}

func TestLowerAtTriggerYourUpkeepDrawCard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "At the beginning of your upkeep, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Type != game.TriggerAt {
		t.Fatalf("trigger type = %v, want TriggerAt", ta.Trigger.Type)
	}
	if ta.Trigger.Pattern.Event != game.EventBeginningOfStep {
		t.Fatalf("event = %v, want EventBeginningOfStep", ta.Trigger.Pattern.Event)
	}
	if ta.Trigger.Pattern.Step != game.StepUpkeep {
		t.Fatalf("step = %v, want StepUpkeep", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("controller = %v, want TriggerControllerYou", ta.Trigger.Pattern.Controller)
	}
	draw, ok := ta.Content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount != game.Fixed(1) {
		t.Fatalf("primitive = %+v, want Draw{Amount: Fixed(1)}", ta.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerAtTriggerEachOpponentUpkeepDamage(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Pinger",
		Layout:     "normal",
		TypeLine:   "Creature — Goblin",
		OracleText: "At the beginning of each opponent's upkeep, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Type != game.TriggerAt {
		t.Fatalf("trigger type = %v, want TriggerAt", ta.Trigger.Type)
	}
	if ta.Trigger.Pattern.Event != game.EventBeginningOfStep {
		t.Fatalf("event = %v, want EventBeginningOfStep", ta.Trigger.Pattern.Event)
	}
	if ta.Trigger.Pattern.Step != game.StepUpkeep {
		t.Fatalf("step = %v, want StepUpkeep", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerOpponent {
		t.Fatalf("controller = %v, want TriggerControllerOpponent", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerEachUpkeepAny(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Watcher",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "At the beginning of each upkeep, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Controller != game.TriggerControllerAny {
		t.Fatalf("controller = %v, want TriggerControllerAny", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerYourEndStep(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mystic",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "At the beginning of your end step, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Step != game.StepEnd {
		t.Fatalf("step = %v, want StepEnd", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("controller = %v, want TriggerControllerYou", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerBeginningOfCombatYourTurn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Fighter",
		Layout:     "normal",
		TypeLine:   "Creature — Human Warrior",
		OracleText: "At the beginning of combat on your turn, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Step != game.StepBeginningOfCombat {
		t.Fatalf("step = %v, want StepBeginningOfCombat", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("controller = %v, want TriggerControllerYou", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerYourDrawStep(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Scholar",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "At the beginning of your draw step, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Step != game.StepDraw {
		t.Fatalf("step = %v, want StepDraw", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("controller = %v, want TriggerControllerYou", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerEachCombat(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Battler",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		OracleText: "At the beginning of each combat, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Step != game.StepBeginningOfCombat {
		t.Fatalf("step = %v, want StepBeginningOfCombat", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerAny {
		t.Fatalf("controller = %v, want TriggerControllerAny", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerOptional(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sage",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "At the beginning of your upkeep, you may draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if !ta.Optional {
		t.Fatal("expected Optional = true for 'you may' trigger")
	}
	if ta.Trigger.Pattern.Step != game.StepUpkeep {
		t.Fatalf("step = %v, want StepUpkeep", ta.Trigger.Pattern.Step)
	}
}

func TestLowerAtTriggerPrecombatMainPhaseFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Planeswalker",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "At the beginning of your precombat main phase, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostic for precombat main phase trigger, got none")
	}
	found := false
	for _, d := range diagnostics {
		if strings.Contains(d.Summary, "unsupported phase/step trigger phrase") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'unsupported phase/step trigger phrase' diagnostic, got: %v", diagnostics)
	}
}

func TestLowerAtTriggerInterveningIfFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "At the beginning of your upkeep, if you control a creature, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostic for intervening-if on at-trigger, got none")
	}
}

func TestLowerAtTriggerPhraseVariants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phrase     string
		step       game.Step
		controller game.TriggerControllerFilter
	}{
		{"each upkeep", game.StepUpkeep, game.TriggerControllerAny},
		{"each player's upkeep", game.StepUpkeep, game.TriggerControllerAny},
		{"each opponent's upkeep", game.StepUpkeep, game.TriggerControllerOpponent},
		{"each end step", game.StepEnd, game.TriggerControllerAny},
		{"each player's end step", game.StepEnd, game.TriggerControllerAny},
		{"each combat", game.StepBeginningOfCombat, game.TriggerControllerAny},
	}
	for _, tc := range tests {
		t.Run(tc.phrase, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   "Creature — Human",
				OracleText: "At the beginning of " + tc.phrase + ", draw a card.",
				Power:      new("1"),
				Toughness:  new("1"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			ta := face.TriggeredAbilities[0]
			if ta.Trigger.Pattern.Step != tc.step {
				t.Errorf("step = %v, want %v", ta.Trigger.Pattern.Step, tc.step)
			}
			if ta.Trigger.Pattern.Controller != tc.controller {
				t.Errorf("controller = %v, want %v", ta.Trigger.Pattern.Controller, tc.controller)
			}
		})
	}
}

func TestLowerCastTriggerAcceptsPlayerPhrases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phrase     string
		controller game.TriggerControllerFilter
	}{
		{"you cast", game.TriggerControllerYou},
		{"a player casts", game.TriggerControllerAny},
		{"an opponent casts", game.TriggerControllerOpponent},
	}
	for _, tc := range tests {
		t.Run(tc.phrase, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "Whenever " + tc.phrase + " a spell, draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			ta := face.TriggeredAbilities[0]
			if ta.Trigger.Pattern.Event != game.EventSpellCast {
				t.Errorf("event = %v, want EventSpellCast", ta.Trigger.Pattern.Event)
			}
			if ta.Trigger.Pattern.Controller != tc.controller {
				t.Errorf("controller = %v, want %v", ta.Trigger.Pattern.Controller, tc.controller)
			}
			if !ta.Trigger.Pattern.CardSelection.Empty() {
				t.Errorf("CardSelection = %+v, want empty for 'a spell'", ta.Trigger.Pattern.CardSelection)
			}
		})
	}
}

func TestLowerCastTriggerAcceptsSpellTypePhrases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phrase    string
		wantTypes []types.Card
		wantAny   []types.Card
		wantExcl  []types.Card
	}{
		{"a creature spell", []types.Card{types.Creature}, nil, nil},
		{"a noncreature spell", nil, nil, []types.Card{types.Creature}},
		{"an instant or sorcery spell", nil, []types.Card{types.Instant, types.Sorcery}, nil},
		{"an instant spell", []types.Card{types.Instant}, nil, nil},
		{"an instant", []types.Card{types.Instant}, nil, nil},
		{"a sorcery spell", []types.Card{types.Sorcery}, nil, nil},
		{"an artifact spell", []types.Card{types.Artifact}, nil, nil},
		{"an enchantment spell", []types.Card{types.Enchantment}, nil, nil},
		{"a land spell", []types.Card{types.Land}, nil, nil},
		{"a planeswalker spell", []types.Card{types.Planeswalker}, nil, nil},
		{"a noncreature, nonland spell", nil, nil, []types.Card{types.Creature, types.Land}},
	}
	for _, tc := range tests {
		t.Run(tc.phrase, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "Whenever you cast " + tc.phrase + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			sel := face.TriggeredAbilities[0].Trigger.Pattern.CardSelection
			if !slices.Equal(sel.RequiredTypes, tc.wantTypes) {
				t.Errorf("RequiredTypes = %v, want %v", sel.RequiredTypes, tc.wantTypes)
			}
			if !slices.Equal(sel.RequiredTypesAny, tc.wantAny) {
				t.Errorf("RequiredTypesAny = %v, want %v", sel.RequiredTypesAny, tc.wantAny)
			}
			if !slices.Equal(sel.ExcludedTypes, tc.wantExcl) {
				t.Errorf("ExcludedTypes = %v, want %v", sel.ExcludedTypes, tc.wantExcl)
			}
		})
	}
}

func TestLowerCastTriggerAcceptsColorPhrases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phrase    string
		wantColor color.Color
	}{
		{"a white spell", color.White},
		{"a blue spell", color.Blue},
		{"a black spell", color.Black},
		{"a red spell", color.Red},
		{"a green spell", color.Green},
	}
	for _, tc := range tests {
		t.Run(tc.phrase, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "Whenever you cast " + tc.phrase + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			sel := face.TriggeredAbilities[0].Trigger.Pattern.CardSelection
			if len(sel.ColorsAny) != 1 || sel.ColorsAny[0] != tc.wantColor {
				t.Errorf("ColorsAny = %v, want [%v]", sel.ColorsAny, tc.wantColor)
			}
		})
	}
}

func TestLowerCastTriggerAcceptsColorCardinalityPhrases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phrase           string
		wantColorless    bool
		wantMulticolored bool
	}{
		{"a colorless spell", true, false},
		{"a multicolored spell", false, true},
	}
	for _, tc := range tests {
		t.Run(tc.phrase, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "Whenever you cast " + tc.phrase + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			sel := face.TriggeredAbilities[0].Trigger.Pattern.CardSelection
			if sel.Colorless != tc.wantColorless {
				t.Errorf("Colorless = %v, want %v", sel.Colorless, tc.wantColorless)
			}
			if sel.Multicolored != tc.wantMulticolored {
				t.Errorf("Multicolored = %v, want %v", sel.Multicolored, tc.wantMulticolored)
			}
		})
	}
}

func TestLowerCastTriggerAcceptsManaValueKickedAndZonePhrases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		phrase string
		assert func(t *testing.T, pattern game.TriggerPattern)
	}{
		{
			name:   "mana value",
			phrase: "a spell with mana value 5 or greater",
			assert: func(t *testing.T, pattern game.TriggerPattern) {
				t.Helper()
				mv := pattern.CardSelection.ManaValue
				if !mv.Exists || mv.Val.Op != compare.GreaterOrEqual || mv.Val.Value != 5 {
					t.Fatalf("ManaValue = %+v, want >= 5", mv)
				}
			},
		},
		{
			name:   "kicked",
			phrase: "a kicked spell",
			assert: func(t *testing.T, pattern game.TriggerPattern) {
				t.Helper()
				if !pattern.RequireKickerPaid {
					t.Fatal("RequireKickerPaid = false, want true")
				}
			},
		},
		{
			name:   "graveyard",
			phrase: "a spell from your graveyard",
			assert: func(t *testing.T, pattern game.TriggerPattern) {
				t.Helper()
				if !pattern.MatchFromZone || pattern.FromZone != zone.Graveyard {
					t.Fatalf("from-zone filter = (%v, %v), want graveyard", pattern.MatchFromZone, pattern.FromZone)
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "Whenever you cast " + tc.phrase + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			tc.assert(t, face.TriggeredAbilities[0].Trigger.Pattern)
		})
	}
}

func TestLowerCastTriggerRejectsUnsupportedForms(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
	}{
		{"self-cast TriggerWhen", "When you cast this spell, draw a card."},
		{"general TriggerWhen", "When you cast a spell, draw a card."},
		{"unrecognized player", "Whenever each player casts a spell, draw a card."},
		{"spell copy", "Whenever you cast or copy an instant or sorcery spell, draw a card."},
		{"ordinal spell", "Whenever you cast your second spell each turn, draw a card."},
		{"subtype spell", "Whenever you cast a Spirit or Arcane spell, draw a card."},
		{"historic spell", "Whenever you cast a historic spell, draw a card."},
		{"unsupported mana value comparison", "Whenever you cast a spell with mana value less than 5, draw a card."},
		{"unsupported zone-filtered spell", "Whenever you cast a spell from your library, draw a card."},
		{"any player your graveyard", "Whenever a player casts a spell from your graveyard, draw a card."},
		{"opponent your graveyard", "Whenever an opponent casts a spell from your graveyard, draw a card."},
		{"intervening if", "Whenever you cast a spell, if you control an artifact, draw a card."},
		{"ability word", "Spellcraft — Whenever you cast a spell, draw a card."},
		{"unsupported body", "Whenever you cast a spell, counter target activated or triggered ability."},
		{"partially optional body", "Whenever you cast a spell, draw a card. You may gain 1 life."},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: tc.oracle,
				Power:      new("2"),
				Toughness:  new("2"),
			}
			faces, diagnostics := lowerExecutableFaces(card)
			if len(diagnostics) == 0 {
				t.Fatalf("expected unsupported diagnostic for %q", tc.oracle)
			}
			if len(faces) > 0 && len(faces[0].TriggeredAbilities) > 0 {
				t.Fatalf("unexpected triggered ability for unsupported form %q", tc.oracle)
			}
		})
	}
}

func TestLowerCastTriggerOptionalBody(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Whenever you cast a creature spell, you may draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Event != game.EventSpellCast {
		t.Errorf("event = %v, want EventSpellCast", ta.Trigger.Pattern.Event)
	}
	if !ta.Optional {
		t.Error("expected optional triggered ability")
	}
}

func TestLowerCyclingTriggers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		oracle      string
		wantEvent   game.EventKind
		excludeSelf bool
	}{
		{
			name:      "cycle a card",
			oracle:    "Whenever you cycle a card, draw a card.",
			wantEvent: game.EventCycled,
		},
		{
			name:        "cycle another card",
			oracle:      "Whenever you cycle another card, draw a card.",
			wantEvent:   game.EventCycled,
			excludeSelf: true,
		},
		{
			name:      "cycle or discard",
			oracle:    "Whenever you cycle or discard a card, draw a card.",
			wantEvent: game.EventCardDiscarded,
		},
		{
			name:        "cycle or discard another",
			oracle:      "Whenever you cycle or discard another card, draw a card.",
			wantEvent:   game.EventCardDiscarded,
			excludeSelf: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: tc.oracle,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			pattern := face.TriggeredAbilities[0].Trigger.Pattern
			if pattern.Event != tc.wantEvent {
				t.Errorf("event = %v, want %v", pattern.Event, tc.wantEvent)
			}
			if pattern.Player != game.TriggerPlayerYou {
				t.Errorf("player = %v, want TriggerPlayerYou", pattern.Player)
			}
			if pattern.ExcludeSelf != tc.excludeSelf {
				t.Errorf("ExcludeSelf = %v, want %v", pattern.ExcludeSelf, tc.excludeSelf)
			}
		})
	}
}

func TestLowerHandCyclingGrants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		oracle    string
		wantTypes []types.Card
		wantCost  cost.Mana
	}{
		{
			name:      "land cards",
			oracle:    "Each land card in your hand has cycling {R}.",
			wantTypes: []types.Card{types.Land},
			wantCost:  cost.Mana{cost.R},
		},
		{
			name:      "creature cards",
			oracle:    "Each creature card in your hand has cycling {1}{U}. ({1}{U}, Discard that card: Draw a card.)",
			wantTypes: []types.Card{types.Creature},
			wantCost:  cost.Mana{cost.O(1), cost.U},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Grant",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: tc.oracle,
			})
			if len(face.StaticAbilities) != 1 {
				t.Fatalf("got %d static abilities, want 1", len(face.StaticAbilities))
			}
			body := face.StaticAbilities[0].Body
			if len(body.RuleEffects) != 1 {
				t.Fatalf("rule effects = %+v, want one", body.RuleEffects)
			}
			effect := body.RuleEffects[0]
			if effect.Kind != game.RuleEffectGrantHandCardAbility {
				t.Fatalf("rule effect kind = %v, want RuleEffectGrantHandCardAbility", effect.Kind)
			}
			if effect.AffectedPlayer != game.PlayerYou {
				t.Fatalf("affected player = %v, want PlayerYou", effect.AffectedPlayer)
			}
			if !slices.Equal(effect.CardSelection.RequiredTypes, tc.wantTypes) {
				t.Fatalf("required types = %v, want %v", effect.CardSelection.RequiredTypes, tc.wantTypes)
			}
			gotCost, ok := game.ActivatedBodyCyclingCost(&effect.GrantedAbility)
			if !ok || !slices.Equal(gotCost, tc.wantCost) {
				t.Fatalf("cycling cost = %v, %v; want %v", gotCost, ok, tc.wantCost)
			}
		})
	}
}

func TestLowerCyclingCostModifiers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		oracle              string
		wantReduction       int
		wantSetCost         bool
		wantHandSize        int
		wantFirstCycleLimit bool
	}{
		{
			name:          "Fluctuator",
			oracle:        "Cycling abilities you activate cost up to {2} less to activate.",
			wantReduction: 2,
		},
		{
			name:         "New Perspectives",
			oracle:       "As long as you have seven or more cards in hand, you may pay {0} rather than pay cycling costs.",
			wantSetCost:  true,
			wantHandSize: 7,
		},
		{
			name:                "Gavi Nest Warden",
			oracle:              "You may pay {0} rather than pay the cycling cost of the first card you cycle each turn.",
			wantSetCost:         true,
			wantFirstCycleLimit: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       tc.name,
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: tc.oracle,
			})
			if len(face.StaticAbilities) != 1 {
				t.Fatalf("got %d static abilities, want 1", len(face.StaticAbilities))
			}
			body := face.StaticAbilities[0].Body
			if len(body.RuleEffects) != 1 {
				t.Fatalf("rule effects = %+v, want one", body.RuleEffects)
			}
			if body.Condition.Exists != (tc.wantHandSize > 0) {
				t.Fatalf("condition exists = %v, want %v", body.Condition.Exists, tc.wantHandSize > 0)
			}
			if tc.wantHandSize > 0 && body.Condition.Val.ControllerHandSizeAtLeast != tc.wantHandSize {
				t.Fatalf("hand-size condition = %d, want %d", body.Condition.Val.ControllerHandSizeAtLeast, tc.wantHandSize)
			}
			effect := body.RuleEffects[0]
			if effect.Kind != game.RuleEffectCostModifier {
				t.Fatalf("rule effect kind = %v, want RuleEffectCostModifier", effect.Kind)
			}
			if effect.AffectedPlayer != game.PlayerYou {
				t.Fatalf("affected player = %v, want PlayerYou", effect.AffectedPlayer)
			}
			modifier := effect.CostModifier
			if modifier.Kind != game.CostModifierAbility || modifier.AbilityKeyword != game.Cycling {
				t.Fatalf("modifier = %+v, want Cycling ability modifier", modifier)
			}
			if modifier.GenericReduction != tc.wantReduction {
				t.Fatalf("generic reduction = %d, want %d", modifier.GenericReduction, tc.wantReduction)
			}
			if modifier.SetManaCost.Exists != tc.wantSetCost {
				t.Fatalf("set mana cost exists = %v, want %v", modifier.SetManaCost.Exists, tc.wantSetCost)
			}
			if tc.wantSetCost && len(modifier.SetManaCost.Val) != 0 {
				t.Fatalf("set mana cost = %v, want zero", modifier.SetManaCost.Val)
			}
			if modifier.FirstCycleEachTurn != tc.wantFirstCycleLimit {
				t.Fatalf("first-cycle limit = %v, want %v", modifier.FirstCycleEachTurn, tc.wantFirstCycleLimit)
			}
		})
	}
}

func TestLowerHandCyclingGrantRejectsHistoric(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Jo Grant",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human",
		OracleText: "Each historic card in your hand has cycling {2}{W}. ({2}{W}, Discard that card: Draw a card.)",
		Power:      new("3"),
		Toughness:  new("4"),
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostic for unsupported historic hand cycling grant")
	}
	if !strings.Contains(diagnostics[0].Detail, "historic card predicates are not supported") {
		t.Fatalf("diagnostic = %#v, want historic predicate detail", diagnostics[0])
	}
}

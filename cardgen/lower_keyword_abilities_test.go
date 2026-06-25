package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLoweringUsesTypedCostFieldsTextBlind(t *testing.T) {
	t.Parallel()

	reveal, ok := lowerRevealCost(compiler.CostComponent{
		Text:             "Reveal misleading words",
		Object:           "garbage goblin from nowhere",
		AmountValue:      2,
		AmountKnown:      true,
		ObjectKind:       compiler.SelectorCard,
		ObjectColor:      color.Blue,
		ObjectColorKnown: true,
		SourceZone:       zone.Hand,
	})
	if !ok ||
		reveal.Kind != cost.AdditionalReveal ||
		reveal.Amount != 2 ||
		!reveal.MatchCardColor ||
		reveal.CardColor != color.Blue ||
		reveal.Source != zone.Hand {
		t.Fatalf("lowerRevealCost text-blind result = %#v, %v", reveal, ok)
	}

	putCounter, ok := lowerPutCounterCost("Actual Card", compiler.CostComponent{
		Text:             "Put nonsense",
		Object:           "not two blood counters on Actual Card",
		AmountValue:      2,
		AmountKnown:      true,
		CounterKind:      counter.Blood,
		CounterKindKnown: true,
		SourceSelf:       true,
	})
	if !ok ||
		putCounter.Kind != cost.AdditionalPutCounter ||
		putCounter.Amount != 2 ||
		putCounter.CounterKind != counter.Blood {
		t.Fatalf("lowerPutCounterCost text-blind result = %#v, %v", putCounter, ok)
	}

	discard, ok := lowerDiscardCost(compiler.CostComponent{
		Text:            "Discard fake permanent",
		Object:          "one permanent",
		AmountValue:     1,
		AmountKnown:     true,
		ObjectKind:      compiler.SelectorCard,
		ObjectType:      types.Creature,
		ObjectTypeKnown: true,
	})
	if !ok ||
		discard.Kind != cost.AdditionalDiscard ||
		discard.Amount != 1 ||
		!discard.MatchCardType ||
		discard.CardType != types.Creature {
		t.Fatalf("lowerDiscardCost text-blind result = %#v, %v", discard, ok)
	}
}

func TestLoweringUsesTypedProtectionKeywordTextBlind(t *testing.T) {
	t.Parallel()

	ability, ok := lowerStaticGrantedAbility([]compiler.CompiledKeyword{{
		Kind:            parser.KeywordProtection,
		Name:            "Protection",
		Parameter:       "from malformed text",
		ProtectionKnown: true,
		Protection: game.ProtectionKeyword{
			FromColors: []color.Color{color.Green},
		},
	}})
	if !ok {
		t.Fatal("lowerStaticGrantedAbility did not consume typed Protection")
	}
	protected := game.StaticBodyProtectionColors(&ability)
	if !slices.Equal(protected, []color.Color{color.Green}) {
		t.Fatalf("protection colors = %v, want green", protected)
	}
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

func TestLowerChangelingKeywordStaticBody(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Woodland Changeling",
		Layout:     "normal",
		TypeLine:   "Creature — Shapeshifter",
		OracleText: "Changeling (This card is every creature type.)",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("got %d static abilities, want 1", len(face.StaticAbilities))
	}
	if got := face.StaticAbilities[0].VarName; got != "game.ChangelingStaticBody" {
		t.Fatalf("changeling static VarName = %q", got)
	}
}

func TestLowerSemicolonSeparatedKeywordLine(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Ancient Spider",
		Layout:     "normal",
		TypeLine:   "Creature — Spider",
		OracleText: "First strike; reach",
		Power:      new("2"),
		Toughness:  new("4"),
	})
	if len(face.StaticAbilities) != 2 {
		t.Fatalf("got %d static abilities, want 2", len(face.StaticAbilities))
	}
	if got := face.StaticAbilities[0].VarName; got != "game.FirstStrikeStaticBody" {
		t.Fatalf("first static VarName = %q", got)
	}
	if got := face.StaticAbilities[1].VarName; got != "game.ReachStaticBody" {
		t.Fatalf("second static VarName = %q", got)
	}
}

func TestLowerHorsemanshipKeyword(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Shu General",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		OracleText: "Vigilance; horsemanship (This creature can't be blocked except by creatures with horsemanship.)",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.StaticAbilities) != 2 {
		t.Fatalf("got %d static abilities, want 2", len(face.StaticAbilities))
	}
	if got := face.StaticAbilities[0].VarName; got != "game.VigilanceStaticBody" {
		t.Fatalf("first static VarName = %q", got)
	}
	if got := face.StaticAbilities[1].VarName; got != "game.HorsemanshipStaticBody" {
		t.Fatalf("second static VarName = %q", got)
	}
}

func TestLowerSemicolonKeywordLineFailsClosedOnUnknownKeyword(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Partial Keyword Tester",
		Layout:     "normal",
		TypeLine:   "Creature — Insect",
		OracleText: "Flying; phasing (This permanent phases in or out before you untap during each of your untap steps.)",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected a fail-closed diagnostic for the unmodeled phasing keyword")
	}
	if got := diagnostics[0].Summary; got != "unsupported mixed keyword ability" {
		t.Fatalf("summary = %q, want unsupported mixed keyword ability", got)
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

func TestLowerParameterizedKeywordsUseTypedValuesTextBlind(t *testing.T) {
	t.Parallel()
	wardBody, ok, diagnostic := lowerParameterizedKeywordToStaticAbility(
		compiler.CompiledAbility{},
		compiler.CompiledKeyword{
			Kind:          parser.KeywordWard,
			Name:          "irrelevant",
			Parameter:     "not mana",
			ParameterKind: parser.KeywordParameterManaCost,
			ManaCost:      cost.Mana{cost.U},
		},
	)
	if !ok || diagnostic != nil || len(wardBody.KeywordAbilities) != 1 {
		t.Fatalf("typed Ward lowering = %+v, %v, %+v", wardBody, ok, diagnostic)
	}
	ward, ok := wardBody.KeywordAbilities[0].(game.WardKeyword)
	if !ok || !slices.Equal(ward.Cost, cost.Mana{cost.U}) {
		t.Fatalf("typed Ward keyword = %+v; want {U}", wardBody.KeywordAbilities[0])
	}

	toxicBody, ok := lowerParameterizedStaticKeyword(compiler.CompiledKeyword{
		Kind:          parser.KeywordToxic,
		Name:          "irrelevant",
		Parameter:     "not an integer",
		ParameterKind: parser.KeywordParameterInteger,
		Integer:       4,
	})
	if !ok || len(toxicBody.KeywordAbilities) != 1 {
		t.Fatalf("typed Toxic lowering = %+v, %v", toxicBody, ok)
	}
	toxic, ok := toxicBody.KeywordAbilities[0].(game.ToxicKeyword)
	if !ok || toxic.Amount != 4 {
		t.Fatalf("typed Toxic keyword = %+v; want amount 4", toxicBody.KeywordAbilities[0])
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

func TestLowerLandcyclingAbility(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		typeLine      string
		oracle        string
		wantCardType  bool
		wantSupertype bool
		wantSubtype   types.Sub
	}{
		{
			name:          "basic landcycling",
			typeLine:      "Land",
			oracle:        "Basic landcycling {1} ({1}, Discard this card: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.)",
			wantCardType:  true,
			wantSupertype: true,
		},
		{
			name:        "plainscycling",
			typeLine:    "Creature — Bird",
			oracle:      "Plainscycling {1}{W} ({1}{W}, Discard this card: Search your library for a Plains card, reveal it, put it into your hand, then shuffle.)",
			wantSubtype: types.Plains,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   tc.typeLine,
				OracleText: tc.oracle,
			}
			if strings.HasPrefix(tc.typeLine, "Creature") {
				card.Power = new("2")
				card.Toughness = new("2")
			}
			face := lowerSingleFace(t, card)
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("got %d activated abilities, want 1", len(face.ActivatedAbilities))
			}
			ability := face.ActivatedAbilities[0]
			if ability.ZoneOfFunction != zone.Hand {
				t.Errorf("zone = %v, want hand", ability.ZoneOfFunction)
			}
			if len(ability.AdditionalCosts) != 1 || ability.AdditionalCosts[0].Kind != cost.AdditionalDiscard {
				t.Fatalf("additional costs = %#v, want one discard", ability.AdditionalCosts)
			}
			if _, ok := ability.KeywordAbilities[0].(game.CyclingKeyword); !ok {
				t.Fatalf("keyword ability = %T, want game.CyclingKeyword", ability.KeywordAbilities[0])
			}
			search, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.Search)
			if !ok {
				t.Fatalf("primitive = %T, want game.Search", ability.Content.Modes[0].Sequence[0].Primitive)
			}
			if search.Spec.Destination != zone.Hand || !search.Spec.Reveal {
				t.Errorf("spec = %#v, want hand destination with reveal", search.Spec)
			}
			if tc.wantCardType && (len(search.Spec.Filter.RequiredTypes) == 0 || search.Spec.Filter.RequiredTypes[0] != types.Land) {
				t.Errorf("card type = %v, want land", search.Spec.Filter.RequiredTypes)
			}
			if tc.wantSupertype && (len(search.Spec.Filter.Supertypes) == 0 || search.Spec.Filter.Supertypes[0] != types.Basic) {
				t.Errorf("supertype = %v, want basic", search.Spec.Filter.Supertypes)
			}
			if tc.wantSubtype != "" && !slices.Contains(search.Spec.Filter.SubtypesAny, tc.wantSubtype) {
				t.Errorf("subtypes = %v, want %v", search.Spec.Filter.SubtypesAny, tc.wantSubtype)
			}
		})
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

package cardgen

import (
	goparser "go/parser"
	"go/token"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

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
	if _, err := goparser.ParseFile(token.NewFileSet(), "test_shaman.go", source, goparser.AllErrors); err != nil {
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
	if _, err := goparser.ParseFile(token.NewFileSet(), "test_excavation.go", source, goparser.AllErrors); err != nil {
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
	if _, err := goparser.ParseFile(token.NewFileSet(), "test_reanimator.go", source, goparser.AllErrors); err != nil {
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
	if _, err := goparser.ParseFile(token.NewFileSet(), "test_construct.go", source, goparser.AllErrors); err != nil {
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

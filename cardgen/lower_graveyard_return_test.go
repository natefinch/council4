package cardgen

import (
	goparser "go/parser"
	"go/token"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
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

func TestLowerTargetedGraveyardPutOnOwnersLibrary(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Revival",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Put target card from a graveyard on top of its owner's library.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one", mode.Targets)
	}
	target := mode.Targets[0]
	if target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard ||
		target.Selection.Val.Controller != game.ControllerAny {
		t.Fatalf("target = %#v", target)
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MoveCard", mode.Sequence[0].Primitive)
	}
	if move.Card.Kind != game.CardReferenceTarget || move.FromZone != zone.Graveyard ||
		move.Destination != zone.Library || move.DestinationBottom {
		t.Fatalf("move = %#v", move)
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

func TestLowerTargetedGraveyardReturnToBattlefieldWithCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Evil Reawakened",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target creature card from your graveyard to the battlefield with two additional +1/+1 counters on it.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one", mode.Targets)
	}
	put, ok := mode.Sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("primitive = %T, want game.PutOnBattlefield", mode.Sequence[0].Primitive)
	}
	want := []game.CounterPlacement{{Kind: counter.PlusOnePlusOne, Amount: 2}}
	if !reflect.DeepEqual(put.EntryCounters, want) {
		t.Fatalf("EntryCounters = %#v, want %#v", put.EntryCounters, want)
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

// TestLowerTargetedGraveyardReturnTypeUnionUsesOrSemantics guards the union
// lowering fix: a "creature or enchantment card" target must carry the two card
// types as a disjunctive RequiredTypesAny with no conjunctive RequiredTypes, so
// it matches enchantment cards as well as creature cards rather than collapsing
// to creatures only.
func TestLowerTargetedGraveyardReturnTypeUnionUsesOrSemantics(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Recovery",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target creature or enchantment card from your graveyard to your hand.",
	})
	selection := face.SpellAbility.Val.Modes[0].Targets[0].Selection.Val
	if len(selection.RequiredTypes) != 0 {
		t.Fatalf("RequiredTypes = %#v, want empty (union must not set a conjunctive type)", selection.RequiredTypes)
	}
	if !slices.Equal(selection.RequiredTypesAny, []types.Card{types.Creature, types.Enchantment}) ||
		selection.Controller != game.ControllerYou {
		t.Fatalf("selection = %#v", selection)
	}
}

func TestLowerTargetedGraveyardReturnPermanentCard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Findbroker",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target permanent card from your graveyard to your hand.",
	})
	selection := face.SpellAbility.Val.Modes[0].Targets[0].Selection.Val
	if len(selection.RequiredTypes) != 0 ||
		!slices.Equal(selection.RequiredTypesAny, []types.Card{
			types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle,
		}) {
		t.Fatalf("selection = %#v", selection)
	}
}

func TestLowerTargetedGraveyardReturnMulticoloredCard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Reborn Hope",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target multicolored card from your graveyard to your hand.",
	})
	selection := face.SpellAbility.Val.Modes[0].Targets[0].Selection.Val
	if !selection.Multicolored || selection.Colorless ||
		len(selection.RequiredTypes) != 0 || len(selection.RequiredTypesAny) != 0 ||
		selection.Controller != game.ControllerYou {
		t.Fatalf("selection = %#v", selection)
	}
}

func TestLowerTargetedGraveyardReturnColorCard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Revive",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target green card from your graveyard to your hand.",
	})
	selection := face.SpellAbility.Val.Modes[0].Targets[0].Selection.Val
	if !slices.Equal(selection.ColorsAny, []color.Color{color.Green}) ||
		selection.Multicolored || selection.Colorless ||
		selection.Controller != game.ControllerYou {
		t.Fatalf("selection = %#v", selection)
	}
}

// TestLowerTargetedGraveyardReturnMultiTarget confirms an "up to N" graveyard
// return lowers to one variable-count target spec and one MoveCard per slot, each
// referencing its own indexed target.
func TestLowerTargetedGraveyardReturnMultiTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Soul Salvage",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return up to two target creature cards from your graveyard to your hand.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one variable target spec", mode.Targets)
	}
	target := mode.Targets[0]
	if target.MinTargets != 0 || target.MaxTargets != 2 ||
		target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard ||
		!slices.Equal(target.Selection.Val.RequiredTypes, []types.Card{types.Creature}) {
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
			move.FromZone != zone.Graveyard || move.Destination != zone.Hand {
			t.Fatalf("move %d = %#v", i, move)
		}
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

// TestLowerGraveyardReturnThenCreateTokenSequence covers a targeted
// graveyard-return clause followed by a non-targeting create-token clause. The
// return primitive (game.MoveCard) must be admitted by the sequence target
// rebaser so its target-card reference survives sequencing.
func TestLowerGraveyardReturnThenCreateTokenSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Reclaimer",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target artifact or creature card from your graveyard to your hand. Create a 1/1 colorless Soldier artifact creature token.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].TargetZone != zone.Graveyard {
		t.Fatalf("targets = %#v, want one graveyard target", mode.Targets)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", mode.Sequence)
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("first primitive = %T, want game.MoveCard", mode.Sequence[0].Primitive)
	}
	if move.Card.Kind != game.CardReferenceTarget || move.Card.TargetIndex != 0 ||
		move.Destination != zone.Hand {
		t.Fatalf("move = %#v", move)
	}
	if _, ok := mode.Sequence[1].Primitive.(game.CreateToken); !ok {
		t.Fatalf("second primitive = %T, want game.CreateToken", mode.Sequence[1].Primitive)
	}
}

// TestLowerGraveyardReanimateToBattlefieldThenGainLife covers a reanimation
// (game.PutOnBattlefield) clause sequenced before a non-targeting life-gain
// clause, exercising the battlefield-source rebase path at offset zero.
func TestLowerGraveyardReanimateToBattlefieldThenGainLife(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Reviver",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target creature card from your graveyard to the battlefield. You gain 3 life.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", mode.Sequence)
	}
	put, ok := mode.Sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("first primitive = %T, want game.PutOnBattlefield", mode.Sequence[0].Primitive)
	}
	card, ok := put.Source.CardRef()
	if !ok || card.Kind != game.CardReferenceTarget || card.TargetIndex != 0 {
		t.Fatalf("put source = %#v", put.Source)
	}
	if _, ok := mode.Sequence[1].Primitive.(game.GainLife); !ok {
		t.Fatalf("second primitive = %T, want game.GainLife", mode.Sequence[1].Primitive)
	}
}

// TestLowerDestroyThenGraveyardReturnKeepsCardSlotZero covers a permanent-target
// clause preceding a targeted graveyard-return clause. Because the runtime numbers
// card-target references among card targets only, the lone card return must stay
// at card slot zero even though its target spec is the second one overall — a
// global-index rebase would wrongly push it to slot one.
func TestLowerDestroyThenGraveyardReturnKeepsCardSlotZero(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Salvage",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target creature. Return target creature card from your graveyard to your hand.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 2 {
		t.Fatalf("targets = %#v, want two", mode.Targets)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", mode.Sequence)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.Destroy); !ok {
		t.Fatalf("first primitive = %T, want game.Destroy", mode.Sequence[0].Primitive)
	}
	move, ok := mode.Sequence[1].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("second primitive = %T, want game.MoveCard", mode.Sequence[1].Primitive)
	}
	// The card reference is numbered among card targets only; the destroy clause
	// owns a permanent target, not a card target, so the lone card return stays at
	// card slot zero even though it is the second target spec overall.
	if move.Card.Kind != game.CardReferenceTarget || move.Card.TargetIndex != 0 {
		t.Fatalf("move card slot = %#v, want card slot zero", move.Card)
	}
}

// TestLowerTwoGraveyardReturnsAdvanceCardSlot covers two targeted graveyard-return
// clauses in one body. Both target specs allow cards, so the second return's card
// reference advances to card slot one.
func TestLowerTwoGraveyardReturnsAdvanceCardSlot(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Double Salvage",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target creature card from your graveyard to your hand. Return target land card from your graveyard to your hand.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 2 || len(mode.Sequence) != 2 {
		t.Fatalf("targets = %#v sequence = %#v, want two of each", mode.Targets, mode.Sequence)
	}
	first, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok || first.Card.TargetIndex != 0 {
		t.Fatalf("first move = %#v, want card slot zero", mode.Sequence[0].Primitive)
	}
	second, ok := mode.Sequence[1].Primitive.(game.MoveCard)
	if !ok || second.Card.TargetIndex != 1 {
		t.Fatalf("second move = %#v, want card slot one", mode.Sequence[1].Primitive)
	}
}

// TestLowerChosenCardGraveyardReturnToHand covers the non-target "Return a
// <filter> card from your graveyard to your hand" recursion wording, which is
// chosen at resolution rather than targeted and lowers to a
// game.ReturnFromGraveyard primitive carrying the card filter.
func TestLowerChosenCardGraveyardReturnToHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Recursion",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return a creature or planeswalker card from your graveyard to your hand.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %#v, want none", mode.Targets)
	}
	ret, ok := mode.Sequence[0].Primitive.(game.ReturnFromGraveyard)
	if !ok {
		t.Fatalf("primitive = %T, want game.ReturnFromGraveyard", mode.Sequence[0].Primitive)
	}
	if ret.Player.Kind() != game.PlayerReferenceController {
		t.Fatalf("player = %#v, want controller", ret.Player)
	}
	if ret.Amount.Value() != 1 {
		t.Fatalf("amount = %#v, want fixed one", ret.Amount)
	}
	if !slices.Equal(ret.Selection.RequiredTypesAny, []types.Card{types.Creature, types.Planeswalker}) {
		t.Fatalf("selection = %#v", ret.Selection)
	}
}

// TestLowerChosenPlainCardGraveyardReturnToHand covers the unrestricted "Return
// a card from your graveyard to your hand" form, which carries no type filter.
func TestLowerChosenPlainCardGraveyardReturnToHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Salvage",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return a card from your graveyard to your hand.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	ret, ok := mode.Sequence[0].Primitive.(game.ReturnFromGraveyard)
	if !ok {
		t.Fatalf("primitive = %T, want game.ReturnFromGraveyard", mode.Sequence[0].Primitive)
	}
	if len(ret.Selection.RequiredTypes) != 0 || len(ret.Selection.RequiredTypesAny) != 0 {
		t.Fatalf("selection should be unrestricted, got %#v", ret.Selection)
	}
}

// TestLowerMassGraveyardReturnToBattlefield covers the mass reanimation wording
// "Return all <filter> cards from your graveyard to the battlefield"
// (Brilliant Restoration, Replenish), which moves every matching graveyard card
// at once with no choice.
func TestLowerMassGraveyardReturnToBattlefield(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Restoration",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return all artifact and enchantment cards from your graveyard to the battlefield.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %#v, want none", mode.Targets)
	}
	mass, ok := mode.Sequence[0].Primitive.(game.MassReturnFromGraveyard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MassReturnFromGraveyard", mode.Sequence[0].Primitive)
	}
	if mass.Destination != zone.Battlefield || mass.EntryTapped {
		t.Fatalf("mass = %#v", mass)
	}
	if !slices.Equal(mass.Selection.RequiredTypesAny, []types.Card{types.Artifact, types.Enchantment}) ||
		mass.Selection.Controller != game.ControllerYou {
		t.Fatalf("selection = %#v", mass.Selection)
	}
}

// TestLowerMassGraveyardReturnToHand covers the same mass recursion to hand.
func TestLowerMassGraveyardReturnToHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Recall",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return all creature cards from your graveyard to your hand.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	mass, ok := mode.Sequence[0].Primitive.(game.MassReturnFromGraveyard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MassReturnFromGraveyard", mode.Sequence[0].Primitive)
	}
	if mass.Destination != zone.Hand ||
		!slices.Equal(mass.Selection.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("mass = %#v", mass)
	}
}

// TestLowerMassGraveyardReturnAllGraveyardsUnderYourControl covers the
// all-graveyards reanimation "Put all <filter> cards from all graveyards onto
// the battlefield under your control" (Rise of the Dark Realms), which scans
// every player's graveyard and enters each card under the resolving controller.
func TestLowerMassGraveyardReturnAllGraveyardsUnderYourControl(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dark Realms",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put all creature cards from all graveyards onto the battlefield under your control.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	mass, ok := mode.Sequence[0].Primitive.(game.MassReturnFromGraveyard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MassReturnFromGraveyard", mode.Sequence[0].Primitive)
	}
	if mass.Destination != zone.Battlefield ||
		mass.EntryTapped ||
		mass.SourceGroup.Kind != game.PlayerGroupReferenceAllPlayers ||
		mass.ControlledByOwner ||
		!slices.Equal(mass.Selection.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("mass = %#v", mass)
	}
}

// TestLowerMassGraveyardReturnAllGraveyardsUnderOwnersControl covers the
// owners'-control all-graveyards reanimation "Return all <filter> cards from all
// graveyards to the battlefield under their owners' control" (Open the Vaults),
// which enters each card under its own owner's control.
func TestLowerMassGraveyardReturnAllGraveyardsUnderOwnersControl(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Open Vaults",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return all artifact and enchantment cards from all graveyards to the battlefield under their owners' control.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	mass, ok := mode.Sequence[0].Primitive.(game.MassReturnFromGraveyard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MassReturnFromGraveyard", mode.Sequence[0].Primitive)
	}
	if mass.Destination != zone.Battlefield ||
		mass.SourceGroup.Kind != game.PlayerGroupReferenceAllPlayers ||
		!mass.ControlledByOwner ||
		!slices.Equal(mass.Selection.RequiredTypesAny, []types.Card{types.Artifact, types.Enchantment}) {
		t.Fatalf("mass = %#v", mass)
	}
}

// TestLowerMassGraveyardReturnAllGraveyardsTapped covers the tapped owners'-
// control all-graveyards reanimation (Planar Birth), confirming the entry-tapped
// rider survives and the entry word does not pollute the selector.
func TestLowerMassGraveyardReturnAllGraveyardsTapped(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Planar Birth",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return all basic land cards from all graveyards to the battlefield tapped under their owners' control.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	mass, ok := mode.Sequence[0].Primitive.(game.MassReturnFromGraveyard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MassReturnFromGraveyard", mode.Sequence[0].Primitive)
	}
	if mass.Destination != zone.Battlefield ||
		!mass.EntryTapped ||
		mass.SourceGroup.Kind != game.PlayerGroupReferenceAllPlayers ||
		!mass.ControlledByOwner ||
		!slices.Equal(mass.Selection.RequiredTypes, []types.Card{types.Land}) {
		t.Fatalf("mass = %#v", mass)
	}
}

// TestLowerTargetedGraveyardReturnPermanentCardWithManaValueToBattlefield guards
// the generic reanimation shape used by Sevinne's Reclamation: "Return target
// permanent card with mana value N or less from your graveyard to the
// battlefield." The target must carry the full permanent type union with a
// mana-value upper bound and lower to a PutOnBattlefield reading the target card.
func TestLowerTargetedGraveyardReturnPermanentCardWithManaValueToBattlefield(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Reclamation",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target permanent card with mana value 1 or less from your graveyard to the battlefield.",
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
	if len(selection.RequiredTypes) != 0 ||
		!slices.Equal(selection.RequiredTypesAny, []types.Card{
			types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle,
		}) ||
		selection.Controller != game.ControllerYou ||
		!selection.ManaValue.Exists ||
		selection.ManaValue.Val.Op != compare.LessOrEqual ||
		selection.ManaValue.Val.Value != 1 {
		t.Fatalf("selection = %#v", selection)
	}
	put, ok := mode.Sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("primitive = %T, want game.PutOnBattlefield", mode.Sequence[0].Primitive)
	}
	cardRef, ok := put.Source.CardRef()
	if !ok || cardRef.Kind != game.CardReferenceTarget {
		t.Fatalf("put = %#v", put)
	}
}

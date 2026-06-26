package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerTokenGrantedAbilityCreatingTokenFailsClosed verifies a token whose
// granted ability itself creates a token ("create ... token with \"When this
// token dies, create a Food token.\"", Wolf's Quarry) fails closed: the
// token-definition emitter does not synthesize nested token definitions, so such
// granted abilities must not lower to a token def referencing an unemitted token.
func TestLowerTokenGrantedAbilityCreatingTokenFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Nested Token",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		Colors:     []string{"G"},
		OracleText: "Create three 1/1 green Boar creature tokens with \"When this token dies, create a Food token.\"",
	}, "t")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected fail-closed for token whose granted ability creates a token, got supported")
	}
}

// TestLowerTokenWithGrantedDeathTrigger verifies that a token created "with" a
// quoted granted ability ("create a 1/1 ... token with \"When this token dies,
// you gain 1 life.\"", Beledros Witherbloom) lowers the inner quoted ability and
// attaches it to the synthesized token definition as a triggered ability.
func TestLowerTokenWithGrantedDeathTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Granted Token",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		Colors:     []string{"B", "G"},
		OracleText: "Create a 1/1 black and green Pest creature token with \"When this token dies, you gain 1 life.\"",
	})
	create := createTokenPrimitive(t, face)
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	if def.Name != "Pest" {
		t.Fatalf("token name = %q, want Pest", def.Name)
	}
	if !def.Power.Exists || def.Power.Val.Value != 1 ||
		!def.Toughness.Exists || def.Toughness.Val.Value != 1 {
		t.Fatalf("token PT = %+v/%+v, want 1/1", def.Power, def.Toughness)
	}
	if len(def.TriggeredAbilities) != 1 {
		t.Fatalf("granted triggered abilities = %d, want 1", len(def.TriggeredAbilities))
	}
	if got := def.TriggeredAbilities[0].Trigger.Pattern.Event; got != game.EventPermanentDied {
		t.Fatalf("granted trigger event = %v, want EventPermanentDied", got)
	}
}

// TestLowerTokenGrantedAbilityWithKeyword verifies the text-blind lowering
// composes a token that combines a printed keyword with a trailing quoted
// granted ability, emitting both the keyword's static body and the granted
// ability.
func TestLowerTokenGrantedAbilityWithKeyword(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Granted Keyword Token",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		Colors:     []string{"B"},
		OracleText: "Create a 1/1 black Zombie creature token with flying and \"When this token dies, you gain 1 life.\"",
	})
	create := createTokenPrimitive(t, face)
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	if len(def.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1 (flying)", len(def.StaticAbilities))
	}
	if def.StaticAbilities[0].Text != "Flying" {
		t.Fatalf("static ability text = %q, want Flying", def.StaticAbilities[0].Text)
	}
	if len(def.TriggeredAbilities) != 1 {
		t.Fatalf("granted triggered abilities = %d, want 1", len(def.TriggeredAbilities))
	}
	if got := def.TriggeredAbilities[0].Trigger.Pattern.Event; got != game.EventPermanentDied {
		t.Fatalf("granted trigger event = %v, want EventPermanentDied", got)
	}
}

// TestLowerTokenWithGrantedStaticCantBlock verifies a token created "with" a
// quoted static restriction ability ("create a 1/1 black Rat creature token
// with \"This token can't block.\"") lowers the inner static ability and
// attaches it to the synthesized token definition. The token's self subject
// ("This token") threads to the same can't-block rule effect the printed "This
// creature can't block." form produces.
func TestLowerTokenWithGrantedStaticCantBlock(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Granted Static Token",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		Colors:     []string{"B"},
		OracleText: "Create a 1/1 black Rat creature token with \"This token can't block.\"",
	})
	create := createTokenPrimitive(t, face)
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	if def.Name != "Rat" {
		t.Fatalf("token name = %q, want Rat", def.Name)
	}
	if len(def.TriggeredAbilities) != 0 || len(def.ActivatedAbilities) != 0 {
		t.Fatalf("granted token has triggered=%d activated=%d, want 0/0",
			len(def.TriggeredAbilities), len(def.ActivatedAbilities))
	}
	if len(def.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1 (can't block)", len(def.StaticAbilities))
	}
	if len(def.StaticAbilities[0].RuleEffects) != 1 {
		t.Fatalf("rule effects = %d, want 1", len(def.StaticAbilities[0].RuleEffects))
	}
	effect := def.StaticAbilities[0].RuleEffects[0]
	if effect.Kind != game.RuleEffectCantBlock || !effect.AffectedSource {
		t.Fatalf("rule effect = %+v, want RuleEffectCantBlock on source", effect)
	}
}

// TestLowerDynamicSizedTokenWithPayLifeCost verifies the full Tivash, Gloom
// Summoner shape: an end-step trigger that may pay X life (X = life gained this
// turn) to create an X/X token. The create token carries a dynamic power and
// toughness override resolved against the same life-gained amount, and the
// resolution payment carries a dynamic pay-life cost.
func TestLowerDynamicSizedTokenWithPayLifeCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dynamic Token",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Demon Cleric",
		ManaCost:   "{4}{B}",
		Colors:     []string{"B"},
		OracleText: "At the beginning of your end step, if you gained life this turn, you may pay X life, where X is the amount of life you gained this turn. If you do, create an X/X black Demon creature token with flying.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence = %d instructions, want 2 (payment, create)", len(sequence))
	}

	payment, ok := sequence[0].Primitive.(game.Pay)
	if !ok {
		t.Fatalf("first instruction = %T, want game.Pay", sequence[0].Primitive)
	}
	if len(payment.Payment.AdditionalCosts) != 1 {
		t.Fatalf("additional costs = %d, want 1", len(payment.Payment.AdditionalCosts))
	}
	if payment.Payment.AdditionalCosts[0].Kind != cost.AdditionalPayLife {
		t.Fatalf("cost kind = %v, want AdditionalPayLife", payment.Payment.AdditionalCosts[0].Kind)
	}
	if payment.Payment.AdditionalCosts[0].AmountDynamic != cost.AdditionalDynamicLifeGainedThisTurn {
		t.Fatalf("cost dynamic amount = %v, want AdditionalDynamicLifeGainedThisTurn", payment.Payment.AdditionalCosts[0].AmountDynamic)
	}

	create, ok := sequence[1].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("second instruction = %T, want game.CreateToken", sequence[1].Primitive)
	}
	if !create.Power.Exists || !create.Toughness.Exists {
		t.Fatalf("create P/T override = %+v/%+v, want both set", create.Power, create.Toughness)
	}
	if got := create.Power.Val.DynamicAmount(); !got.Exists || got.Val.Kind != game.DynamicAmountLifeGainedThisTurn {
		t.Fatalf("create power dynamic = %+v, want DynamicAmountLifeGainedThisTurn", got)
	}
	if got := create.Toughness.Val.DynamicAmount(); !got.Exists || got.Val.Kind != game.DynamicAmountLifeGainedThisTurn {
		t.Fatalf("create toughness dynamic = %+v, want DynamicAmountLifeGainedThisTurn", got)
	}
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	if def.Power.Exists || def.Toughness.Exists {
		t.Fatalf("variable-size token def should leave printed P/T unset, got %+v/%+v", def.Power, def.Toughness)
	}
	if len(def.Subtypes) != 1 || def.Subtypes[0] != types.Demon {
		t.Fatalf("token subtypes = %v, want [Demon]", def.Subtypes)
	}
}

package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// groupKeywordGrant lowers a spell whose only effect is a resolving keyword
// grant to a creature or permanent group until end of turn and returns the
// single keyword-layer continuous effect.
func groupKeywordGrant(t *testing.T, oracleText string) game.ContinuousEffect {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Group Keyword",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: oracleText,
	})
	mode := face.SpellAbility.Val.Modes[0]
	primitive, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	if primitive.Object.Exists || primitive.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("primitive = %+v, want unanchored group effect until end of turn", primitive)
	}
	if len(primitive.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %d, want 1", len(primitive.ContinuousEffects))
	}
	effect := primitive.ContinuousEffects[0]
	if effect.Layer != game.LayerAbility {
		t.Fatalf("layer = %v, want LayerAbility", effect.Layer)
	}
	return effect
}

func TestLowerGroupKeywordGrantControlledCreatures(t *testing.T) {
	t.Parallel()
	effect := groupKeywordGrant(t, "Creatures you control gain trample until end of turn.")
	if len(effect.AddKeywords) != 1 || effect.AddKeywords[0] != game.Trample {
		t.Fatalf("keywords = %v, want [Trample]", effect.AddKeywords)
	}
	selection := effect.Group.Selection()
	if effect.Group.Domain() != game.GroupDomainBattlefield ||
		selection.Controller != game.ControllerYou ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %+v, want creatures you control", selection)
	}
	if _, excludes := effect.Group.Exclusion(); excludes {
		t.Fatal("creatures you control must not exclude the source")
	}
}

func TestLowerGroupKeywordGrantControlledPermanents(t *testing.T) {
	t.Parallel()
	effect := groupKeywordGrant(t, "Permanents you control gain hexproof and indestructible until end of turn.")
	if len(effect.AddKeywords) != 2 ||
		effect.AddKeywords[0] != game.Hexproof ||
		effect.AddKeywords[1] != game.Indestructible {
		t.Fatalf("keywords = %v, want [Hexproof Indestructible]", effect.AddKeywords)
	}
	selection := effect.Group.Selection()
	if effect.Group.Domain() != game.GroupDomainBattlefield ||
		selection.Controller != game.ControllerYou ||
		len(selection.RequiredTypes) != 0 {
		t.Fatalf("selection = %+v, want permanents you control", selection)
	}
}

func TestLowerGroupKeywordGrantMultipleKeywords(t *testing.T) {
	t.Parallel()
	effect := groupKeywordGrant(t, "Creatures you control gain first strike and deathtouch until end of turn.")
	if len(effect.AddKeywords) != 2 ||
		effect.AddKeywords[0] != game.FirstStrike ||
		effect.AddKeywords[1] != game.Deathtouch {
		t.Fatalf("keywords = %v, want [FirstStrike Deathtouch]", effect.AddKeywords)
	}
}

func TestLowerGroupKeywordGrantAttackingCreatures(t *testing.T) {
	t.Parallel()
	effect := groupKeywordGrant(t, "Attacking creatures gain first strike until end of turn.")
	selection := effect.Group.Selection()
	if selection.CombatState != game.CombatStateAttacking ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %+v, want attacking creatures", selection)
	}
}

func TestLowerGroupKeywordGrantOtherControlledCreatures(t *testing.T) {
	t.Parallel()
	effect := groupKeywordGrant(t, "Other creatures you control gain vigilance until end of turn.")
	exclude, excludes := effect.Group.Exclusion()
	if !excludes || exclude != game.SourcePermanentReference() {
		t.Fatalf("exclusion = %v/%v, want source permanent excluded", exclude, excludes)
	}
}

// TestLowerGroupKeywordGrantColorFilteredRejected verifies a color-filtered
// group keyword grant fails closed: the executable backend does not model the
// color constraint on a one-shot affected group.
func TestLowerGroupKeywordGrantColorFilteredRejected(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Color Group Keyword",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "White creatures you control gain trample until end of turn.",
	})
	if len(diagnostics) != 1 || diagnostics[0].Summary != "unsupported temporary keyword spell" {
		t.Fatalf("diagnostics = %#v, want one unsupported temporary keyword spell", diagnostics)
	}
}

// TestLowerGroupKeywordGrantQuotedAbilityRejected verifies a granted quoted
// ability fails closed rather than dropping the ability text.
func TestLowerGroupKeywordGrantQuotedAbilityRejected(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Quoted Group Keyword",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Creatures you control gain \"When this creature dies, draw a card\" until end of turn.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("granted quoted ability must fail closed")
	}
}

// dynamicModifyPT lowers a single-target dynamic power/toughness pump and
// returns the ModifyPT primitive.
func dynamicModifyPT(t *testing.T, oracleText string) game.ModifyPT {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dynamic PT",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: oracleText,
	})
	mode := face.SpellAbility.Val.Modes[0]
	modify, ok := mode.Sequence[0].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("primitive = %T, want game.ModifyPT", mode.Sequence[0].Primitive)
	}
	if modify.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("duration = %v, want until end of turn", modify.Duration)
	}
	return modify
}

func TestLowerAsymmetricDynamicPTPowerOnly(t *testing.T) {
	t.Parallel()
	modify := dynamicModifyPT(t, "Target creature gets +X/+0 until end of turn, where X is the number of creature cards in your graveyard.")
	if !modify.PowerDelta.IsDynamic() {
		t.Fatalf("power delta = %+v, want dynamic", modify.PowerDelta)
	}
	if modify.ToughnessDelta.IsDynamic() || modify.ToughnessDelta.Value() != 0 {
		t.Fatalf("toughness delta = %+v, want fixed 0", modify.ToughnessDelta)
	}
}

func TestLowerAsymmetricDynamicPTMixedSign(t *testing.T) {
	t.Parallel()
	modify := dynamicModifyPT(t, "Target creature gets +X/-X until end of turn, where X is the number of cards in your hand.")
	power := modify.PowerDelta.DynamicAmount()
	if !power.Exists || power.Val.Multiplier != 1 {
		t.Fatalf("power delta = %+v, want dynamic multiplier +1", modify.PowerDelta)
	}
	toughness := modify.ToughnessDelta.DynamicAmount()
	if !toughness.Exists || toughness.Val.Multiplier != -1 {
		t.Fatalf("toughness delta = %+v, want dynamic multiplier -1", modify.ToughnessDelta)
	}
}

func TestLowerSymmetricDynamicPTStillLowers(t *testing.T) {
	t.Parallel()
	modify := dynamicModifyPT(t, "Target creature gets +X/+X until end of turn, where X is the number of creature cards in your graveyard.")
	if !modify.PowerDelta.IsDynamic() || !modify.ToughnessDelta.IsDynamic() {
		t.Fatalf("deltas = %+v/%+v, want both dynamic", modify.PowerDelta, modify.ToughnessDelta)
	}
}

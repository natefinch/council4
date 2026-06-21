package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
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

// TestLowerGroupKeywordGrantProtectionFromEachColor verifies a mixed grant of
// simple keywords plus protection from each color lowers the simple keywords to
// AddKeywords and the parameterized protection keyword to a granted static
// ability via AddAbilities. This is the Akroma's Will mode-2 shape.
func TestLowerGroupKeywordGrantProtectionFromEachColor(t *testing.T) {
	t.Parallel()
	effect := groupKeywordGrant(t,
		"Creatures you control gain lifelink, indestructible, and protection from each color until end of turn.")
	if len(effect.AddKeywords) != 2 ||
		effect.AddKeywords[0] != game.Lifelink ||
		effect.AddKeywords[1] != game.Indestructible {
		t.Fatalf("keywords = %v, want [Lifelink Indestructible]", effect.AddKeywords)
	}
	if len(effect.AddAbilities) != 1 {
		t.Fatalf("abilities = %d, want 1 granted protection ability", len(effect.AddAbilities))
	}
	static, ok := effect.AddAbilities[0].(*game.StaticAbility)
	if !ok {
		t.Fatalf("ability = %T, want *game.StaticAbility", effect.AddAbilities[0])
	}
	prot, ok := game.StaticBodyProtectionKeyword(static)
	if !ok || !prot.EachColor {
		t.Fatalf("protection = %+v ok=%v, want protection from each color", prot, ok)
	}
}

// TestLowerGroupKeywordGrantProtectionFromColor verifies that a group grant of
// protection from a single named color lowers the parameterized protection
// keyword to a granted static ability carrying that color.
func TestLowerGroupKeywordGrantProtectionFromColor(t *testing.T) {
	t.Parallel()
	effect := groupKeywordGrant(t,
		"Creatures you control gain protection from red until end of turn.")
	if len(effect.AddKeywords) != 0 {
		t.Fatalf("keywords = %v, want none", effect.AddKeywords)
	}
	if len(effect.AddAbilities) != 1 {
		t.Fatalf("abilities = %d, want 1 granted protection ability", len(effect.AddAbilities))
	}
	static, ok := effect.AddAbilities[0].(*game.StaticAbility)
	if !ok {
		t.Fatalf("ability = %T, want *game.StaticAbility", effect.AddAbilities[0])
	}
	prot, ok := game.StaticBodyProtectionKeyword(static)
	if !ok || len(prot.FromColors) != 1 || prot.FromColors[0] != color.Red {
		t.Fatalf("protection = %+v ok=%v, want protection from red", prot, ok)
	}
}

// TestLowerGroupKeywordGrantProtectionFromChosenColor verifies that a group
// grant of protection from a color chosen on resolution lowers to a granted
// static ability marked ChosenColor; the rules resolve the color when the
// effect is applied.
func TestLowerGroupKeywordGrantProtectionFromChosenColor(t *testing.T) {
	t.Parallel()
	effect := groupKeywordGrant(t,
		"Creatures you control gain protection from the color of your choice until end of turn.")
	if len(effect.AddAbilities) != 1 {
		t.Fatalf("abilities = %d, want 1 granted protection ability", len(effect.AddAbilities))
	}
	static, ok := effect.AddAbilities[0].(*game.StaticAbility)
	if !ok {
		t.Fatalf("ability = %T, want *game.StaticAbility", effect.AddAbilities[0])
	}
	prot, ok := game.StaticBodyProtectionKeyword(static)
	if !ok || !prot.ChosenColor {
		t.Fatalf("protection = %+v ok=%v, want protection from chosen color", prot, ok)
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

// massPump lowers a spell whose body is a mass power/toughness pump over a
// creature group (optionally combined with a keyword grant) and returns the
// ApplyContinuous primitive's continuous effects.
func massPump(t *testing.T, oracleText string) []game.ContinuousEffect {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mass Pump",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: oracleText,
	})
	mode := face.SpellAbility.Val.Modes[0]
	primitive, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	if primitive.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("duration = %v, want until end of turn", primitive.Duration)
	}
	return primitive.ContinuousEffects
}

func groupPTEffect(t *testing.T, effects []game.ContinuousEffect) game.ContinuousEffect {
	t.Helper()
	for i := range effects {
		if effects[i].Layer == game.LayerPowerToughnessModify {
			return effects[i]
		}
	}
	t.Fatalf("effects = %+v, want a power/toughness layer effect", effects)
	return game.ContinuousEffect{}
}

// TestLowerMassDynamicPumpCountStandalone covers a standalone mass dynamic pump
// counted over the controller's creatures ("Creatures you control get +X/+X …,
// where X is the number of creatures you control.").
func TestLowerMassDynamicPumpCountStandalone(t *testing.T) {
	t.Parallel()
	effects := massPump(t, "Creatures you control get +X/+X until end of turn, where X is the number of creatures you control.")
	if len(effects) != 1 {
		t.Fatalf("effects = %d, want 1 group pump", len(effects))
	}
	effect := effects[0]
	if effect.Layer != game.LayerPowerToughnessModify {
		t.Fatalf("layer = %v, want LayerPowerToughnessModify", effect.Layer)
	}
	if !effect.PowerDeltaDynamic.Exists || !effect.ToughnessDeltaDynamic.Exists {
		t.Fatalf("deltas = %+v/%+v, want both dynamic", effect.PowerDeltaDynamic, effect.ToughnessDeltaDynamic)
	}
	if effect.PowerDeltaDynamic.Val.Kind != game.DynamicAmountCountSelector {
		t.Fatalf("kind = %v, want DynamicAmountCountSelector", effect.PowerDeltaDynamic.Val.Kind)
	}
	selection := effect.Group.Selection()
	if selection.Controller != game.ControllerYou ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %+v, want creatures you control", selection)
	}
}

// TestLowerMassDynamicPumpKeywordFirst covers the Craterhoof Behemoth shape:
// "creatures you control gain trample and get +X/+X until end of turn, where X
// is the number of creatures you control." The keyword grant precedes the pump
// and both apply to the same group in one ApplyContinuous.
func TestLowerMassDynamicPumpKeywordFirst(t *testing.T) {
	t.Parallel()
	effects := massPump(t, "Creatures you control gain trample and get +X/+X until end of turn, where X is the number of creatures you control.")
	if len(effects) != 2 {
		t.Fatalf("effects = %d, want pump plus keyword grant", len(effects))
	}
	pump := groupPTEffect(t, effects)
	if !pump.PowerDeltaDynamic.Exists ||
		pump.PowerDeltaDynamic.Val.Kind != game.DynamicAmountCountSelector {
		t.Fatalf("pump = %+v, want dynamic count power delta", pump)
	}
	var keyword game.ContinuousEffect
	for i := range effects {
		if effects[i].Layer == game.LayerAbility {
			keyword = effects[i]
		}
	}
	if len(keyword.AddKeywords) != 1 || keyword.AddKeywords[0] != game.Trample {
		t.Fatalf("keywords = %v, want [Trample]", keyword.AddKeywords)
	}
}

// TestLowerMassDynamicPumpGreatestPower covers the Overwhelming Stampede shape,
// where X measures the greatest power among the controller's creatures.
func TestLowerMassDynamicPumpGreatestPower(t *testing.T) {
	t.Parallel()
	effects := massPump(t, "Creatures you control gain trample and get +X/+X until end of turn, where X is the greatest power among creatures you control.")
	pump := groupPTEffect(t, effects)
	if !pump.PowerDeltaDynamic.Exists ||
		pump.PowerDeltaDynamic.Val.Kind != game.DynamicAmountGreatestPowerInGroup {
		t.Fatalf("pump = %+v, want greatest-power dynamic delta", pump)
	}
}

// TestLowerMassDynamicPumpLeadingDuration covers the Overwhelming Stampede
// printed wording, where the duration leads the sentence ("Until end of turn,
// creatures you control gain trample and get +X/+X, where X is …") rather than
// trailing the effects.
func TestLowerMassDynamicPumpLeadingDuration(t *testing.T) {
	t.Parallel()
	effects := massPump(t, "Until end of turn, creatures you control gain trample and get +X/+X, where X is the greatest power among creatures you control.")
	if len(effects) != 2 {
		t.Fatalf("effects = %d, want pump plus keyword grant", len(effects))
	}
	pump := groupPTEffect(t, effects)
	if !pump.PowerDeltaDynamic.Exists ||
		pump.PowerDeltaDynamic.Val.Kind != game.DynamicAmountGreatestPowerInGroup {
		t.Fatalf("pump = %+v, want greatest-power dynamic delta", pump)
	}
	var keyword game.ContinuousEffect
	for i := range effects {
		if effects[i].Layer == game.LayerAbility {
			keyword = effects[i]
		}
	}
	if len(keyword.AddKeywords) != 1 || keyword.AddKeywords[0] != game.Trample {
		t.Fatalf("keywords = %v, want [Trample]", keyword.AddKeywords)
	}
}

// TestLowerMassDynamicPumpMultipleKeywords covers the End-Raze Forerunners
// shape, granting two keywords alongside the dynamic pump.
func TestLowerMassDynamicPumpMultipleKeywords(t *testing.T) {
	t.Parallel()
	effects := massPump(t, "Creatures you control gain trample and haste and get +X/+X until end of turn, where X is the number of creatures you control.")
	var keyword game.ContinuousEffect
	for i := range effects {
		if effects[i].Layer == game.LayerAbility {
			keyword = effects[i]
		}
	}
	if len(keyword.AddKeywords) != 2 ||
		keyword.AddKeywords[0] != game.Trample ||
		keyword.AddKeywords[1] != game.Haste {
		t.Fatalf("keywords = %v, want [Trample Haste]", keyword.AddKeywords)
	}
}

// TestLowerMassFixedPumpKeywordFirst confirms the keyword-first combined shape
// also lowers with a fixed pump amount.
func TestLowerMassFixedPumpKeywordFirst(t *testing.T) {
	t.Parallel()
	effects := massPump(t, "Creatures you control gain trample and get +1/+1 until end of turn.")
	pump := groupPTEffect(t, effects)
	if pump.PowerDeltaDynamic.Exists || pump.PowerDelta != 1 || pump.ToughnessDelta != 1 {
		t.Fatalf("pump = %+v, want fixed +1/+1", pump)
	}
}

// TestLowerMassDynamicPumpSourcePowerRejected confirms a "where X is its power"
// mass pump fails closed: the executable backend does not bind the source-power
// referent for a group.
func TestLowerMassDynamicPumpSourcePowerRejected(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Mass Source Power",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Creatures you control get +X/+X until end of turn, where X is its power.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("source-power mass pump must fail closed")
	}
}

package cardgen

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

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
		"controlled subtype with creatures noun": {
			oracleText: "Sliver creatures you control have flying.",
			domain:     game.GroupDomainObjectControlled,
			subtypes:   []types.Sub{types.Sliver},
			keywords:   []game.Keyword{game.Flying},
		},
		"other controlled subtype with creatures noun": {
			oracleText: "Other Minotaur creatures you control have deathtouch.",
			domain:     game.GroupDomainObjectControlled,
			excluded:   true,
			subtypes:   []types.Sub{types.Minotaur},
			keywords:   []game.Keyword{game.Deathtouch},
		},
		"irregular plural subtype": {
			oracleText: "Elves you control have vigilance.",
			domain:     game.GroupDomainObjectControlled,
			subtypes:   []types.Sub{types.Elf},
			keywords:   []game.Keyword{game.Vigilance},
		},
		"controlled creatures horsemanship": {
			oracleText: "Creatures you control have horsemanship.",
			domain:     game.GroupDomainObjectControlled,
			keywords:   []game.Keyword{game.Horsemanship},
		},
		"enchanted creature infect": {
			oracleText: "Enchanted creature has infect.",
			domain:     game.GroupDomainAttachedObject,
			keywords:   []game.Keyword{game.Infect},
		},
		"other controlled subtype exalted": {
			oracleText: "Other Vampires you control have exalted.",
			domain:     game.GroupDomainObjectControlled,
			excluded:   true,
			subtypes:   []types.Sub{types.Vampire},
			keywords:   []game.Keyword{game.Exalted},
		},
		"controlled creatures riot": {
			oracleText: "Creatures you control have riot.",
			domain:     game.GroupDomainObjectControlled,
			keywords:   []game.Keyword{game.Riot},
		},
		"controlled permanents": {
			oracleText: "Permanents you control have indestructible.",
			domain:     game.GroupDomainObjectControlled,
			keywords:   []game.Keyword{game.Indestructible},
		},
		"other controlled permanents": {
			oracleText: "Other permanents you control have indestructible.",
			domain:     game.GroupDomainObjectControlled,
			excluded:   true,
			keywords:   []game.Keyword{game.Indestructible},
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

func TestLowerStaticDeclarationBattlefieldSelectionControllerRelation(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Curse",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Creatures your opponents control get -1/-0.",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %#v, want one", face.StaticAbilities)
	}
	effects := face.StaticAbilities[0].Body.ContinuousEffects
	if len(effects) != 1 {
		t.Fatalf("continuous effects = %#v, want one", effects)
	}
	effect := effects[0]
	if effect.Layer != game.LayerPowerToughnessModify ||
		effect.Group.Domain() != game.GroupDomainBattlefield ||
		effect.Group.Selection().Controller != game.ControllerOpponent ||
		!slices.Equal(effect.Group.Selection().RequiredTypes, []types.Card{types.Creature}) ||
		effect.PowerDelta != -1 ||
		effect.ToughnessDelta != 0 {
		t.Fatalf("continuous effect = %#v", effect)
	}
}

func TestLowerStaticDeclarationGroupAnthems(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText  string
		domain      game.GroupReferenceDomain
		excluded    bool
		requireType []types.Card
		subtypes    []types.Sub
		combatState game.CombatStateFilter
		power       int
		toughness   int
		keywords    []game.Keyword
	}{
		"all creatures modify": {
			oracleText:  "All creatures get +1/+1.",
			domain:      game.GroupDomainBattlefield,
			requireType: []types.Card{types.Creature},
			power:       1,
			toughness:   1,
		},
		"all creatures keyword": {
			oracleText:  "All creatures have haste.",
			domain:      game.GroupDomainBattlefield,
			requireType: []types.Card{types.Creature},
			keywords:    []game.Keyword{game.Haste},
		},
		"all other creatures excluded": {
			oracleText:  "All other creatures get -1/-1.",
			domain:      game.GroupDomainBattlefield,
			excluded:    true,
			requireType: []types.Card{types.Creature},
			power:       -1,
			toughness:   -1,
		},
		"attacking creatures battlefield": {
			oracleText:  "Attacking creatures get -1/-0.",
			domain:      game.GroupDomainBattlefield,
			requireType: []types.Card{types.Creature},
			combatState: game.CombatStateAttacking,
			power:       -1,
			toughness:   0,
		},
		"blocking creatures battlefield": {
			oracleText:  "Blocking creatures get +0/+2.",
			domain:      game.GroupDomainBattlefield,
			requireType: []types.Card{types.Creature},
			combatState: game.CombatStateBlocking,
			power:       0,
			toughness:   2,
		},
		"all subtype creatures modify": {
			oracleText: "All Sliver creatures get +1/+1.",
			domain:     game.GroupDomainBattlefield,
			subtypes:   []types.Sub{types.Sliver},
			power:      1,
			toughness:  1,
		},
		"all subtype creatures keyword": {
			oracleText: "All Sliver creatures have flying.",
			domain:     game.GroupDomainBattlefield,
			subtypes:   []types.Sub{types.Sliver},
			keywords:   []game.Keyword{game.Flying},
		},
		"other subtype creatures excluded": {
			oracleText: "Other Soldier creatures get +1/+1.",
			domain:     game.GroupDomainBattlefield,
			excluded:   true,
			subtypes:   []types.Sub{types.Soldier},
			power:      1,
			toughness:  1,
		},
		"attacking creatures you control": {
			oracleText:  "Attacking creatures you control get +1/+0.",
			domain:      game.GroupDomainObjectControlled,
			requireType: []types.Card{types.Creature},
			combatState: game.CombatStateAttacking,
			power:       1,
			toughness:   0,
		},
		"attacking creatures you control keyword": {
			oracleText:  "Attacking creatures you control have double strike.",
			domain:      game.GroupDomainObjectControlled,
			requireType: []types.Card{types.Creature},
			combatState: game.CombatStateAttacking,
			keywords:    []game.Keyword{game.DoubleStrike},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Anthem",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: test.oracleText,
			})
			if len(face.StaticAbilities) != 1 {
				t.Fatalf("static abilities = %#v, want one", face.StaticAbilities)
			}
			effects := face.StaticAbilities[0].Body.ContinuousEffects
			if len(effects) != 1 {
				t.Fatalf("continuous effects = %#v, want one", effects)
			}
			effect := effects[0]
			selection := effect.Group.Selection()
			if effect.Group.Domain() != test.domain ||
				!slices.Equal(selection.RequiredTypes, test.requireType) ||
				!slices.Equal(selection.SubtypesAny, test.subtypes) ||
				selection.CombatState != test.combatState {
				t.Fatalf("continuous effect group = %#v", effect)
			}
			if _, excluded := effect.Group.Exclusion(); excluded != test.excluded {
				t.Fatalf("group exclusion = %v, want %v", excluded, test.excluded)
			}
			if len(test.keywords) != 0 {
				if effect.Layer != game.LayerAbility || !slices.Equal(effect.AddKeywords, test.keywords) {
					t.Fatalf("continuous effect keywords = %#v", effect)
				}
			} else if effect.Layer != game.LayerPowerToughnessModify ||
				effect.PowerDelta != test.power ||
				effect.ToughnessDelta != test.toughness {
				t.Fatalf("continuous effect modify = %#v", effect)
			}
		})
	}
}

// TestLowerStaticDeclarationMixedKeywordAndProtectionGrant covers a continuous
// "creatures you control have <simple keywords> and protection from <colors>"
// anthem whose grant list mixes ordinary keywords (flying, vigilance, ...) with
// an ability-backed keyword (protection from a color). The lowering must
// populate BOTH AddKeywords (the simple keywords) and AddAbilities (the
// protection static ability) on a single LayerAbility effect — Akroma's
// Memorial.
func TestLowerStaticDeclarationMixedKeywordAndProtectionGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Memorial",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Creatures you control have flying, first strike, vigilance, trample, haste, and protection from black and from red.",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %#v, want one", face.StaticAbilities)
	}
	effects := face.StaticAbilities[0].Body.ContinuousEffects
	if len(effects) != 1 {
		t.Fatalf("continuous effects = %#v, want one", effects)
	}
	effect := effects[0]
	if effect.Layer != game.LayerAbility {
		t.Fatalf("layer = %v, want LayerAbility", effect.Layer)
	}
	wantKeywords := []game.Keyword{game.Flying, game.FirstStrike, game.Vigilance, game.Trample, game.Haste}
	if !slices.Equal(effect.AddKeywords, wantKeywords) {
		t.Fatalf("AddKeywords = %#v, want %#v", effect.AddKeywords, wantKeywords)
	}
	if len(effect.AddAbilities) != 1 {
		t.Fatalf("AddAbilities = %#v, want one", effect.AddAbilities)
	}
	wantProtection := game.ProtectionFromColorsStaticAbility(color.Black, color.Red)
	if !reflect.DeepEqual(effect.AddAbilities[0], &wantProtection) {
		t.Fatalf("AddAbilities[0] = %#v, want %#v", effect.AddAbilities[0], &wantProtection)
	}
}

// group: "creatures you control of the chosen type" buffs, whose runtime
// Selection must carry SubtypeChoiceSourceEntry so only permanents matching
// the source's entry-time creature-type choice are affected (Patchwork Banner,
// Adaptive Automaton, Obelisk of Urd).
func TestLowerStaticDeclarationChosenTypeAnthems(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		excluded   bool
		power      int
		toughness  int
	}{
		"controlled chosen type": {
			oracleText: "As this artifact enters, choose a creature type.\nCreatures you control of the chosen type get +1/+1.",
			power:      1,
			toughness:  1,
		},
		"other controlled chosen type": {
			oracleText: "As this creature enters, choose a creature type.\nOther creatures you control of the chosen type get +2/+2.",
			excluded:   true,
			power:      2,
			toughness:  2,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Chosen Anthem",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: test.oracleText,
			})
			if len(face.StaticAbilities) != 1 {
				t.Fatalf("static abilities = %#v, want one", face.StaticAbilities)
			}
			effects := face.StaticAbilities[0].Body.ContinuousEffects
			if len(effects) != 1 {
				t.Fatalf("continuous effects = %#v, want one", effects)
			}
			effect := effects[0]
			selection := effect.Group.Selection()
			if effect.Group.Domain() != game.GroupDomainObjectControlled ||
				!slices.Equal(selection.RequiredTypes, []types.Card{types.Creature}) ||
				selection.SubtypeChoice != game.SubtypeChoiceSourceEntry {
				t.Fatalf("continuous effect group = %#v", effect)
			}
			if _, excluded := effect.Group.Exclusion(); excluded != test.excluded {
				t.Fatalf("group exclusion = %v, want %v", excluded, test.excluded)
			}
			if effect.Layer != game.LayerPowerToughnessModify ||
				effect.PowerDelta != test.power ||
				effect.ToughnessDelta != test.toughness {
				t.Fatalf("continuous effect modify = %#v", effect)
			}
		})
	}
}

// TestLowerStaticDeclarationKeywordFilterGroupAnthems covers the group subjects
// added for this slice: keyword-filtered creatures ("Creatures with flying"),
// keyword-excluded creatures ("Creatures without flying"), artifact creatures,
// and nontoken creatures. Each asserts the runtime Selection carries the right
// Keyword/ExcludedKeyword/NonToken/RequiredTypes for the matcher.
func TestLowerStaticDeclarationKeywordFilterGroupAnthems(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText      string
		domain          game.GroupReferenceDomain
		excluded        bool
		controller      game.ControllerRelation
		requireType     []types.Card
		keyword         game.Keyword
		excludedKeyword game.Keyword
		nonToken        bool
		power           int
		toughness       int
		keywords        []game.Keyword
	}{
		"creatures with flying battlefield": {
			oracleText:  "Creatures with flying get +1/+1.",
			domain:      game.GroupDomainBattlefield,
			requireType: []types.Card{types.Creature},
			keyword:     game.Flying,
			power:       1,
			toughness:   1,
		},
		"creatures without flying battlefield": {
			oracleText:      "Creatures without flying get -2/-0.",
			domain:          game.GroupDomainBattlefield,
			requireType:     []types.Card{types.Creature},
			excludedKeyword: game.Flying,
			power:           -2,
			toughness:       0,
		},
		"creatures you control with flying": {
			oracleText:  "Creatures you control with flying get +1/+1.",
			domain:      game.GroupDomainObjectControlled,
			requireType: []types.Card{types.Creature},
			keyword:     game.Flying,
			power:       1,
			toughness:   1,
		},
		"other creatures you control with flying": {
			oracleText:  "Other creatures you control with flying get +1/+1.",
			domain:      game.GroupDomainObjectControlled,
			excluded:    true,
			requireType: []types.Card{types.Creature},
			keyword:     game.Flying,
			power:       1,
			toughness:   1,
		},
		"creatures with flying opponents control": {
			oracleText:  "Creatures with flying your opponents control get -1/-0.",
			domain:      game.GroupDomainBattlefield,
			controller:  game.ControllerOpponent,
			requireType: []types.Card{types.Creature},
			keyword:     game.Flying,
			power:       -1,
			toughness:   0,
		},
		"artifact creatures you control": {
			oracleText:  "Artifact creatures you control get +1/+1.",
			domain:      game.GroupDomainObjectControlled,
			requireType: []types.Card{types.Artifact, types.Creature},
			power:       1,
			toughness:   1,
		},
		"nontoken creatures you control": {
			oracleText:  "Nontoken creatures you control get +1/+1.",
			domain:      game.GroupDomainObjectControlled,
			requireType: []types.Card{types.Creature},
			nonToken:    true,
			power:       1,
			toughness:   1,
		},
		"creatures you control with flying keyword grant": {
			oracleText:  "Creatures you control with flying have vigilance.",
			domain:      game.GroupDomainObjectControlled,
			requireType: []types.Card{types.Creature},
			keyword:     game.Flying,
			keywords:    []game.Keyword{game.Vigilance},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Keyword Anthem",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: test.oracleText,
			})
			if len(face.StaticAbilities) != 1 {
				t.Fatalf("static abilities = %#v, want one", face.StaticAbilities)
			}
			effects := face.StaticAbilities[0].Body.ContinuousEffects
			if len(effects) != 1 {
				t.Fatalf("continuous effects = %#v, want one", effects)
			}
			effect := effects[0]
			selection := effect.Group.Selection()
			if effect.Group.Domain() != test.domain ||
				!slices.Equal(selection.RequiredTypes, test.requireType) ||
				selection.Keyword != test.keyword ||
				selection.ExcludedKeyword != test.excludedKeyword ||
				selection.NonToken != test.nonToken ||
				selection.Controller != test.controller {
				t.Fatalf("continuous effect group = %#v", effect)
			}
			if _, excluded := effect.Group.Exclusion(); excluded != test.excluded {
				t.Fatalf("group exclusion = %v, want %v", excluded, test.excluded)
			}
			if len(test.keywords) != 0 {
				if effect.Layer != game.LayerAbility || !slices.Equal(effect.AddKeywords, test.keywords) {
					t.Fatalf("continuous effect keywords = %#v", effect)
				}
			} else if effect.Layer != game.LayerPowerToughnessModify ||
				effect.PowerDelta != test.power ||
				effect.ToughnessDelta != test.toughness {
				t.Fatalf("continuous effect modify = %#v", effect)
			}
		})
	}
}

// TestRejectUnsupportedGroupAnthemVariants keeps group subjects whose runtime
// selection the static-declaration backend cannot yet express fail-closed: an
// "is enchanted" state filter, a battlefield-wide color-exclusion filter, and an
// excluded-supertype ("nonlegendary") filter.
func TestRejectUnsupportedGroupAnthemVariants(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Creatures you control that are enchanted get +1/+1.",
		"Nonblack creatures get -1/-1.",
		"Nonlegendary creatures get +1/+1.",
	} {
		_, diagnostics := lowerExecutableFaces(&ScryfallCard{
			Name:       "Test Reject",
			Layout:     "normal",
			TypeLine:   "Enchantment",
			OracleText: oracleText,
		})
		if len(diagnostics) == 0 {
			t.Fatalf("%q lowered without diagnostics", oracleText)
		}
	}
}

func TestLowerStaticDeclarationSubtypeCreaturesAnthem(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		subtype    types.Sub
		excluded   bool
		power      int
		toughness  int
	}{
		"controlled subtype anthem": {
			oracleText: "Sliver creatures you control get +2/+0.",
			subtype:    types.Sliver,
			excluded:   false,
			power:      2,
			toughness:  0,
		},
		"other controlled subtype anthem": {
			oracleText: "Other Zombie creatures you control get +1/+1.",
			subtype:    types.Zombie,
			excluded:   true,
			power:      1,
			toughness:  1,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Lord",
				Layout:     "normal",
				TypeLine:   "Creature — Test",
				OracleText: test.oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.StaticAbilities) != 1 {
				t.Fatalf("static abilities = %#v, want one", face.StaticAbilities)
			}
			effects := face.StaticAbilities[0].Body.ContinuousEffects
			if len(effects) != 1 {
				t.Fatalf("continuous effects = %#v, want one", effects)
			}
			effect := effects[0]
			if effect.Layer != game.LayerPowerToughnessModify ||
				effect.Group.Domain() != game.GroupDomainObjectControlled ||
				!slices.Equal(effect.Group.Selection().SubtypesAny, []types.Sub{test.subtype}) ||
				effect.PowerDelta != test.power ||
				effect.ToughnessDelta != test.toughness {
				t.Fatalf("continuous effect = %#v", effect)
			}
			if _, excluded := effect.Group.Exclusion(); excluded != test.excluded {
				t.Fatalf("group exclusion = %v, want %v", excluded, test.excluded)
			}
		})
	}
}

// TestRejectSubtypeCreaturesAnthemUnsupportedVariants keeps the explicit
// "creatures" noun group recognition fail-closed for an unknown subtype
// qualifier, which is not representable and must diagnose.
func TestRejectSubtypeCreaturesAnthemUnsupportedVariants(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		// "Splorp" is not a known creature subtype.
		"Splorp creatures you control get +1/+1.",
		// "Monocolored" has no Selection color-filter representation.
		"Monocolored creatures you control get +1/+1.",
	} {
		_, diagnostics := lowerExecutableFaces(&ScryfallCard{
			Name:       "Test Reject",
			Layout:     "normal",
			TypeLine:   "Enchantment",
			OracleText: oracleText,
		})
		if len(diagnostics) == 0 {
			t.Fatalf("%q lowered without diagnostics", oracleText)
		}
	}
}

// TestLowerStaticColorCreaturesAnthem verifies a color-filtered creature group
// anthem lowers onto a controlled-permanent group whose Selection carries the
// matching color predicate. Because the runtime Selection already matches
// permanents by ColorsAny, Colorless, and Multicolored, asserting on the lowered
// Selection is sufficient.
func TestLowerStaticColorCreaturesAnthem(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText   string
		colorsAny    []color.Color
		colorless    bool
		multicolored bool
		excluded     bool
	}{
		"leading color": {
			oracleText: "Red creatures you control get +1/+1.",
			colorsAny:  []color.Color{color.Red},
		},
		"other color": {
			oracleText: "Other white creatures you control get +1/+1.",
			colorsAny:  []color.Color{color.White},
			excluded:   true,
		},
		"colorless qualifier": {
			oracleText: "Other colorless creatures you control get +0/+1.",
			colorless:  true,
			excluded:   true,
		},
		"multicolored qualifier": {
			oracleText:   "Multicolored creatures you control get +1/+0.",
			multicolored: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Color Lord",
				Layout:     "normal",
				TypeLine:   "Creature — Test",
				OracleText: test.oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.StaticAbilities) != 1 {
				t.Fatalf("static abilities = %#v, want one", face.StaticAbilities)
			}
			effects := face.StaticAbilities[0].Body.ContinuousEffects
			if len(effects) != 1 {
				t.Fatalf("continuous effects = %#v, want one", effects)
			}
			effect := effects[0]
			selection := effect.Group.Selection()
			if effect.Layer != game.LayerPowerToughnessModify ||
				effect.Group.Domain() != game.GroupDomainObjectControlled ||
				!slices.Equal(selection.RequiredTypes, []types.Card{types.Creature}) ||
				!slices.Equal(selection.ColorsAny, test.colorsAny) ||
				selection.Colorless != test.colorless ||
				selection.Multicolored != test.multicolored {
				t.Fatalf("continuous effect = %#v selection = %#v", effect, selection)
			}
			if _, excluded := effect.Group.Exclusion(); excluded != test.excluded {
				t.Fatalf("group exclusion = %v, want %v", excluded, test.excluded)
			}
		})
	}
}

// TestLowerStaticFilteredCreatureGroupAnthem verifies that the bounded
// non-color filtered group anthems lower onto the correct group domain and
// Selection: creature-token groups (token-only), legendary groups (Legendary
// supertype), and tapped/untapped groups (Selection.Tapped). The runtime
// Selection already matches permanents by these predicates, so asserting on the
// lowered Selection is sufficient.
func TestLowerStaticFilteredCreatureGroupAnthem(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText        string
		domain            game.GroupReferenceDomain
		supertypes        []types.Super
		excludedSupertype types.Super
		tapped            game.TriState
		tokenOnly         bool
		excluded          bool
	}{
		"controlled creature tokens": {
			oracleText: "Creature tokens you control get +1/+1.",
			domain:     game.GroupDomainObjectControlled,
			tokenOnly:  true,
		},
		"battlefield creature tokens": {
			oracleText: "Creature tokens get -1/-1.",
			domain:     game.GroupDomainBattlefield,
			tokenOnly:  true,
		},
		"controlled legendary creatures": {
			oracleText: "Legendary creatures you control get +1/+1.",
			domain:     game.GroupDomainObjectControlled,
			supertypes: []types.Super{types.Legendary},
		},
		"controlled nonlegendary creatures": {
			oracleText:        "Nonlegendary creatures you control get +1/+1.",
			domain:            game.GroupDomainObjectControlled,
			excludedSupertype: types.Legendary,
		},
		"controlled untapped creatures": {
			oracleText: "Untapped creatures you control get +1/+1.",
			domain:     game.GroupDomainObjectControlled,
			tapped:     game.TriFalse,
		},
		"other controlled tapped creatures": {
			oracleText: "Other tapped creatures you control get +1/+1.",
			domain:     game.GroupDomainObjectControlled,
			tapped:     game.TriTrue,
			excluded:   true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Filter Lord",
				Layout:     "normal",
				TypeLine:   "Creature — Test",
				OracleText: test.oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.StaticAbilities) != 1 {
				t.Fatalf("static abilities = %#v, want one", face.StaticAbilities)
			}
			effects := face.StaticAbilities[0].Body.ContinuousEffects
			if len(effects) != 1 {
				t.Fatalf("continuous effects = %#v, want one", effects)
			}
			effect := effects[0]
			selection := effect.Group.Selection()
			if effect.Layer != game.LayerPowerToughnessModify ||
				effect.Group.Domain() != test.domain ||
				!slices.Equal(selection.RequiredTypes, []types.Card{types.Creature}) ||
				!slices.Equal(selection.Supertypes, test.supertypes) ||
				selection.ExcludedSupertype != test.excludedSupertype ||
				selection.Tapped != test.tapped ||
				selection.TokenOnly != test.tokenOnly {
				t.Fatalf("continuous effect = %#v selection = %#v", effect, selection)
			}
			if _, excluded := effect.Group.Exclusion(); excluded != test.excluded {
				t.Fatalf("group exclusion = %v, want %v", excluded, test.excluded)
			}
		})
	}
}

func TestLowerStaticControlGrantAura(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Control Magic",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\nYou control enchanted creature.",
	})
	if len(face.StaticAbilities) != 2 {
		t.Fatalf("static abilities = %#v, want two (enchant + control grant)", face.StaticAbilities)
	}
	effects := face.StaticAbilities[1].Body.ContinuousEffects
	if len(effects) != 1 {
		t.Fatalf("continuous effects = %#v, want one", effects)
	}
	effect := effects[0]
	if effect.Layer != game.LayerControl ||
		effect.Group.Domain() != game.GroupDomainAttachedObject ||
		!effect.NewController.Exists ||
		effect.AffectedSource {
		t.Fatalf("continuous effect = %#v", effect)
	}
}

func TestLowerMixedStaticDeclarationsConsumeWholeParagraph(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Dragon's Rage Channeler",
		Layout:     "normal",
		TypeLine:   "Creature — Human Shaman",
		OracleText: "Delirium — As long as there are four or more card types among cards in your graveyard, Dragon's Rage Channeler gets +2/+2, has flying, and attacks each combat if able.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %#v, want one", face.StaticAbilities)
	}
	ability := face.StaticAbilities[0].Body
	if !ability.Condition.Exists ||
		len(ability.Condition.Val.Aggregates) != 1 ||
		ability.Condition.Val.Aggregates[0].Aggregate != game.AggregateControllerGraveyardCardTypeCount ||
		ability.Condition.Val.Aggregates[0].Value != 4 {
		t.Fatalf("condition = %#v", ability.Condition)
	}
	if len(ability.ContinuousEffects) != 2 {
		t.Fatalf("continuous effects = %#v, want two", ability.ContinuousEffects)
	}
	if ability.ContinuousEffects[0].Layer != game.LayerPowerToughnessModify ||
		!ability.ContinuousEffects[0].AffectedSource ||
		ability.ContinuousEffects[0].PowerDelta != 2 ||
		ability.ContinuousEffects[0].ToughnessDelta != 2 {
		t.Fatalf("power/toughness effect = %#v", ability.ContinuousEffects[0])
	}
	if ability.ContinuousEffects[1].Layer != game.LayerAbility ||
		!ability.ContinuousEffects[1].AffectedSource ||
		!slices.Equal(ability.ContinuousEffects[1].AddKeywords, []game.Keyword{game.Flying}) {
		t.Fatalf("keyword effect = %#v", ability.ContinuousEffects[1])
	}
	if len(ability.RuleEffects) != 1 ||
		ability.RuleEffects[0].Kind != game.RuleEffectMustAttack ||
		!ability.RuleEffects[0].AffectedSource {
		t.Fatalf("rule effects = %#v", ability.RuleEffects)
	}
}

func TestGenerateMixedStaticDeclarationsSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Dragon's Rage Channeler",
		Layout:     "normal",
		TypeLine:   "Creature — Human Shaman",
		OracleText: "Delirium — As long as there are four or more card types among cards in your graveyard, Dragon's Rage Channeler gets +2/+2, has flying, and attacks each combat if able.",
		Power:      new("1"),
		Toughness:  new("1"),
	}, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Aggregate: game.AggregateControllerGraveyardCardTypeCount, Op: compare.GreaterOrEqual, Value: 4",
		"game.LayerPowerToughnessModify",
		"game.LayerAbility",
		"game.Flying",
		"game.RuleEffectMustAttack",
		"AffectedSource: true",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestStaticDeclarationBlockersAreCapabilityAware(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		summary    string
	}{
		"duration": {
			oracleText: "Creatures you control get +1/+1 until end of turn.",
			summary:    "unsupported static declaration duration",
		},
		"condition": {
			oracleText: "As long as the moon is full, creatures you control get +1/+1.",
			summary:    "unsupported static declaration condition",
		},
		"group": {
			oracleText: "Creatures you control that are enchanted get +1/+1.",
			summary:    "unsupported static declaration group",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Enchantment",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: test.oracleText,
			})
			if len(diagnostics) != 1 || diagnostics[0].Summary != test.summary {
				t.Fatalf("diagnostics = %#v, want %q", diagnostics, test.summary)
			}
		})
	}
}

func TestLowerStaticDeclarationsRejectMalformedPayloads(t *testing.T) {
	t.Parallel()
	tests := map[string]compiler.StaticDeclaration{
		"missing payload": {
			Kind: compiler.StaticDeclarationContinuous,
		},
		"mismatched payload": {
			Kind: compiler.StaticDeclarationContinuous,
			Rule: &compiler.StaticRuleDeclaration{Kind: compiler.StaticRuleCantBlock},
		},
		"multiple payloads": {
			Kind:       compiler.StaticDeclarationContinuous,
			Continuous: &compiler.StaticContinuousDeclaration{},
			Rule:       &compiler.StaticRuleDeclaration{},
		},
	}
	for name, declaration := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, handled, diagnostic := lowerStaticDeclarations(compiler.CompiledAbility{
				Kind: compiler.AbilityStatic,
				Static: &compiler.CompiledStaticSemantics{
					Declarations: []compiler.StaticDeclaration{declaration},
				},
			}, &parser.Ability{})
			if !handled || diagnostic == nil || diagnostic.Summary != "unsupported static declaration operation" {
				t.Fatalf("handled = %v, diagnostic = %#v", handled, diagnostic)
			}
		})
	}
}

func TestLowerStaticRuleDeclarationsWithoutInspectingText(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		rule    compiler.StaticRuleKind
		domain  compiler.StaticRuleDomain
		zone    compiler.StaticZone
		want    game.RuleEffectKind
		varName string
	}{
		"cannot block": {
			rule:    compiler.StaticRuleCantBlock,
			domain:  compiler.StaticRuleDomainBlock,
			zone:    compiler.StaticZoneBattlefield,
			want:    game.RuleEffectCantBlock,
			varName: "game.CantBlockStaticBody",
		},
		"cannot be blocked": {
			rule:    compiler.StaticRuleCantBeBlocked,
			domain:  compiler.StaticRuleDomainBlock,
			zone:    compiler.StaticZoneBattlefield,
			want:    game.RuleEffectCantBeBlocked,
			varName: "game.CantBeBlockedStaticBody",
		},
		"must attack": {
			rule:    compiler.StaticRuleMustAttack,
			domain:  compiler.StaticRuleDomainAttack,
			zone:    compiler.StaticZoneBattlefield,
			want:    game.RuleEffectMustAttack,
			varName: "game.MustAttackStaticBody",
		},
		"cannot be countered": {
			rule:    compiler.StaticRuleCantBeCountered,
			domain:  compiler.StaticRuleDomainCountering,
			zone:    compiler.StaticZoneStack,
			want:    game.RuleEffectCantBeCountered,
			varName: "game.CantBeCounteredStaticBody",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			lowered, handled, diagnostic := lowerStaticDeclarations(compiler.CompiledAbility{
				Kind: compiler.AbilityStatic,
				Text: "not Oracle wording",
				Static: &compiler.CompiledStaticSemantics{
					Declarations: []compiler.StaticDeclaration{{
						Kind:  compiler.StaticDeclarationRule,
						Group: compiler.StaticGroupReference{Domain: compiler.StaticGroupSource},
						Rule: &compiler.StaticRuleDeclaration{
							Domain: test.domain,
							Kind:   test.rule,
							Zone:   test.zone,
						},
					}},
				},
			}, &parser.Ability{})
			if !handled || diagnostic != nil || len(lowered.staticAbilities) != 1 {
				t.Fatalf("handled = %v, diagnostic = %#v, lowered = %#v", handled, diagnostic, lowered)
			}
			ability := lowered.staticAbilities[0]
			if ability.VarName != test.varName ||
				len(ability.Body.RuleEffects) != 1 ||
				ability.Body.RuleEffects[0].Kind != test.want {
				t.Fatalf("static ability = %#v", ability)
			}
		})
	}
}

func TestConditionalStaticRuleDoesNotUseUnconditionalCanonicalBody(t *testing.T) {
	t.Parallel()
	declaration := compiler.StaticDeclaration{
		Kind:      compiler.StaticDeclarationRule,
		Condition: &compiler.CompiledCondition{},
		Rule:      &compiler.StaticRuleDeclaration{Kind: compiler.StaticRuleMustAttack},
	}
	if got := canonicalStaticDeclarationVarName(declaration); got != "" {
		t.Fatalf("canonical variable = %q, want none", got)
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
	if len(diagnostics) == 0 || diagnostics[0].Summary != "unsupported static declaration operation" {
		t.Fatalf("diagnostics = %#v, want unsupported static declaration operation", diagnostics)
	}
}

func TestRejectMalformedStandaloneStaticKeywordGrants(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Creatures you control have flying or haste.",
		"Creatures you control have and flying.",
		"Creatures you control have flying and.",
		"Creatures you control have flying haste.",
		"Creatures you control have persist.",
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
		!condition.ControlsMatching.Exists ||
		!slices.Equal(condition.ControlsMatching.Val.Selection.SubtypesAny, []types.Sub{types.Mountain}) {
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
	if !condition.ControlsMatching.Exists ||
		!condition.ControlsMatching.Val.Selection.ExcludeSource ||
		!slices.Equal(condition.ControlsMatching.Val.Selection.SubtypesAny, []types.Sub{types.Cleric}) {
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
	if !condition.ControlsMatching.Exists ||
		!slices.Equal(condition.ControlsMatching.Val.Selection.SubtypesAny, []types.Sub{types.Gate}) {
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
		colorless      bool
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
			colorless:  true,
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
			match := face.StaticAbilities[0].Body.Condition.Val.ControlsMatching
			if !match.Exists {
				t.Fatal("condition has no matching-selection count")
			}
			filter := match.Val.Selection
			if !slices.Equal(filter.RequiredTypes, test.types) ||
				!slices.Equal(filter.ColorsAny, test.colors) ||
				!slices.Equal(filter.ExcludedColors, test.excludedColors) ||
				filter.Colorless != test.colorless {
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

// TestLowerSourceConditionalProtectionKeywordGrant verifies that declaration
// lowering produces AddAbilities (not AddKeywords) for parameterized Protection.
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

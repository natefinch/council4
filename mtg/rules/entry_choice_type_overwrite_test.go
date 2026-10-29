package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3525 (part 3 of #3069), "are the chosen type"
// overwriting creature types from an entry choice. Conspiracy, pinned oracle
// 775d284d-0884-4ae3-a1b5-2b2a50a895b4 (printing
// 8bbdb402-0e22-4f83-987a-9d0268a82f10): "As this enchantment enters, choose a
// creature type. / Creatures you control are the chosen type. The same is true
// for creature spells you control and creature cards you own that aren't on the
// battlefield." Rulings from
// https://api.scryfall.com/cards/8bbdb402-0e22-4f83-987a-9d0268a82f10/rulings:
// 2013-09-20 "Affected spells, cards, and creatures lose all of their other
// creature types."; 2004-10-04 it can grant a creature type to animated
// permanents and does not replace creature types in card text; 2005-08-01
// choosing Wall confers no defender; 2017-09-29 replacement effects that modify
// creatures of a type as they enter apply after this effect (CR 614.12, an
// entering-permanent route this Engine unit does not change). Conspiracy is
// not admitted here; the scenarios use a test-local declaration of its
// pinned Oracle text that a later card unit registers. Scenario cards use their
// pinned Oracle text: Vanquisher's Banner efa6944e-b5e2-4584-ae37-f5b59a8ec75e
// (creatures of the chosen type get +1/+1), Reflections of Littjara
// 578a1846-8c1a-4013-b669-1d3f4ddbbaa3 (copy a spell of the chosen type),
// Changeling Outcast ab499df5-ed76-43ad-81b3-70afd0487ccd (Changeling), Goblin
// Matron (search for a Goblin card), Arcane Adaptation ("in addition to their
// other types"), Mistform Ultimus ("is every creature type"), Raging Goblin and
// Grizzly Bears. CR 2026-09-25 (official MagicCompRules 20260925.txt, sha256
// 8d860e45...070ca) read: 205.1a (setting a subtype replaces existing subtypes
// of that set only), 205.1b ("in addition" retains prior types), 604.3/604.3a
// and 702.73a (CDAs, including Changeling, function in every zone),
// 611.3a-b (not locked in; applies while the source is on the battlefield),
// 613.1d (layer 4), 613.3 (within layers 2-6, CDA effects apply first, then
// others by timestamp), 613.7/613.7a (timestamp order of static effects) and
// 613.8a (no dependency: neither effect changes what the other applies to).

// conspiracyTestDef declares Conspiracy's pinned Oracle text with the shared
// entry-choice creature-type overwrite.
func conspiracyTestDef() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:     "Conspiracy",
			ManaCost: opt.Val(cost.Mana{cost.O(3), cost.B, cost.B}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{{
				Text: "Creatures you control are the chosen type. The same is true for creature spells you control and creature cards you own that aren't on the battlefield.",
				ContinuousEffects: []game.ContinuousEffect{{
					Duration:                   game.DurationPermanent,
					Group:                      game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
					Layer:                      game.LayerType,
					SpellSelection:             &game.Selection{RequiredTypes: []types.Card{types.Creature}},
					OwnedCardSelection:         &game.Selection{RequiredTypes: []types.Card{types.Creature}},
					SetSubtypeFromSourceChoice: game.EntryTypeChoiceKey,
					SetSubtypeChoiceType:       types.Creature,
				}},
			}},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this enchantment enters, choose a creature type."),
			},
		},
	}
}

func addConspiracy(t *testing.T, g *game.Game, controller game.PlayerID, chosen types.Sub) *game.Permanent {
	t.Helper()
	def := conspiracyTestDef()
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("Conspiracy declaration issues = %+v", issues)
	}
	permanent := addCombatPermanent(g, controller, def)
	permanent.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryTypeChoiceKey: {Kind: game.ResolutionChoiceSubtype, Subtype: chosen},
	}
	return permanent
}

func creatureSubtypes(subtypes []types.Sub) []types.Sub {
	family := types.SubtypesForType(types.Creature)
	var creature []types.Sub
	for _, subtype := range subtypes {
		if slices.Contains(family, subtype) {
			creature = append(creature, subtype)
		}
	}
	return creature
}

func assertCreatureSubtypes(t *testing.T, what string, subtypes []types.Sub, want ...types.Sub) {
	t.Helper()
	got := creatureSubtypes(subtypes)
	if len(got) != len(want) {
		t.Fatalf("%s creature types = %v, want %v", what, got, want)
	}
	for _, subtype := range want {
		if !slices.Contains(got, subtype) {
			t.Fatalf("%s creature types = %v, want %v", what, got, want)
		}
	}
}

// TestConspiracyOverwritesCreatureTypesOnBattlefield is acceptance scenario 1:
// with Conspiracy (Elf) and Vanquisher's Banner (Goblin), Player1's Goblin
// Matron loses Goblin and does not get +1/+1, and Changeling Outcast is only an
// Elf (its Changeling CDA applies first, CR 613.3). An opponent's creature and
// Player1's noncreature artifact keep their types, and the result survives a
// clone.
func TestConspiracyOverwritesCreatureTypesOnBattlefield(t *testing.T) {
	for _, withConspiracy := range []bool{false, true} {
		g, _ := spellTypeScenario(t)
		banner := addChosenTypePermanent(t, g, game.Player1, "Vanquisher's Banner", types.Goblin)
		if withConspiracy {
			addConspiracy(t, g, game.Player1, types.Elf)
		}
		matron := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Goblin Matron"))
		outcast := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Changeling Outcast"))
		opponentMatron := addCombatPermanent(g, game.Player2, productionChoiceCard(t, "Goblin Matron"))

		for _, state := range []*game.Game{g, g.Clone()} {
			if !withConspiracy {
				assertCreatureSubtypes(t, "Goblin Matron", effectivePermanentValues(state, matron).subtypes, types.Goblin)
				if got := effectivePower(state, matron); got != 2 {
					t.Fatalf("Goblin Matron power under Banner = %d, want 2", got)
				}
				if !permanentHasSubtype(state, outcast, types.Goblin) || !permanentHasSubtype(state, outcast, types.Shapeshifter) {
					t.Fatal("Changeling Outcast is not every creature type")
				}
				continue
			}
			assertCreatureSubtypes(t, "Goblin Matron", effectivePermanentValues(state, matron).subtypes, types.Elf)
			if got := effectivePower(state, matron); got != 1 {
				t.Fatalf("Goblin Matron power = %d, want 1 (no longer a Goblin)", got)
			}
			assertCreatureSubtypes(t, "Changeling Outcast", effectivePermanentValues(state, outcast).subtypes, types.Elf)
			if got := effectivePower(state, outcast); got != 1 {
				t.Fatalf("Changeling Outcast power = %d, want 1", got)
			}
			assertCreatureSubtypes(t, "opponent's Goblin Matron", effectivePermanentValues(state, opponentMatron).subtypes, types.Goblin)
			if values := effectivePermanentValues(state, banner); len(values.subtypes) != 0 || slices.Contains(values.types, types.Creature) {
				t.Fatalf("Vanquisher's Banner types = %v %v, want an unchanged noncreature artifact", values.types, values.subtypes)
			}
		}
	}
}

// TestConspiracyKeepsNonCreatureSubtypes proves only creature types are
// replaced (CR 205.1a): an Equipment creature keeps Equipment.
func TestConspiracyKeepsNonCreatureSubtypes(t *testing.T) {
	g, _ := spellTypeScenario(t)
	addConspiracy(t, g, game.Player1, types.Elf)
	equipment := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Equipment Creature",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Equipment, types.Golem},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	subtypes := effectivePermanentValues(g, equipment).subtypes
	assertCreatureSubtypes(t, "Equipment creature", subtypes, types.Elf)
	if !slices.Contains(subtypes, types.Equipment) {
		t.Fatalf("subtypes = %v, want Equipment kept", subtypes)
	}
}

// TestConspiracyCreatureSpellsAreOnlyTheChosenType is acceptance scenario 2:
// with Conspiracy (Elf), Reflections of Littjara (Goblin) does not trigger for
// Raging Goblin; with Reflections set to Elf, casting Grizzly Bears triggers and
// the copy, also Player1's creature spell, is only an Elf.
func TestConspiracyCreatureSpellsAreOnlyTheChosenType(t *testing.T) {
	g, engine := spellTypeScenario(t)
	addConspiracy(t, g, game.Player1, types.Elf)
	addChosenTypePermanent(t, g, game.Player1, "Reflections of Littjara", types.Goblin)

	goblin := castSpellTypeTestCard(t, g, engine, game.Player1, "Raging Goblin", nil, mana.R)
	assertCreatureSubtypes(t, "Raging Goblin spell", spellCharacteristics(t, g, goblin).Subtypes, types.Elf)
	assertCreatureSubtypes(t, "Raging Goblin cast event", lastSpellCastEvent(t, g).CardSubtypes, types.Elf)
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Reflections of Littjara (Goblin) triggered for a spell that is only an Elf")
	}

	g, engine = spellTypeScenario(t)
	addConspiracy(t, g, game.Player1, types.Elf)
	addChosenTypePermanent(t, g, game.Player1, "Reflections of Littjara", types.Elf)
	bears := castSpellTypeTestCard(t, g, engine, game.Player1, "Grizzly Bears", nil, mana.G, mana.C)
	assertCreatureSubtypes(t, "Grizzly Bears spell", spellCharacteristics(t, g, bears).Subtypes, types.Elf)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Reflections of Littjara (Elf) did not trigger for Grizzly Bears")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	copyObj, ok := g.Stack.Peek()
	if !ok || !copyObj.Copy {
		t.Fatalf("top of stack = %+v, want the copy of Grizzly Bears", copyObj)
	}
	for _, state := range []*game.Game{g, g.Clone()} {
		top, _ := state.Stack.Peek()
		assertCreatureSubtypes(t, "Grizzly Bears copy", spellCharacteristics(t, state, top).Subtypes, types.Elf)
	}
}

// TestConspiracyOpponentSpellsKeepTheirTypes proves "creature spells you
// control" is scoped to Conspiracy's controller.
func TestConspiracyOpponentSpellsKeepTheirTypes(t *testing.T) {
	g, engine := spellTypeScenario(t)
	addConspiracy(t, g, game.Player1, types.Elf)
	setMainPhasePriority(g, game.Player2)
	spell := castSpellTypeTestCard(t, g, engine, game.Player2, "Raging Goblin", nil, mana.R)
	assertCreatureSubtypes(t, "Player2's Raging Goblin spell", spellCharacteristics(t, g, spell).Subtypes, types.Goblin, types.Berserker)
}

// TestConspiracyGoblinMatronCannotFindGoblinCreatureCard is acceptance
// scenario 3: under Conspiracy (Elf), Goblin Matron cannot find the Goblin
// creature card Raging Goblin or Changeling Outcast in Player1's library; under
// Conspiracy (Goblin) it finds Grizzly Bears. An opponent's Conspiracy does not
// change Player1's cards.
func TestConspiracyGoblinMatronCannotFindGoblinCreatureCard(t *testing.T) {
	for _, tc := range []struct {
		name       string
		controller game.PlayerID
		chosen     types.Sub
		card       string
		want       bool
	}{
		{name: "Elf hides Raging Goblin", controller: game.Player1, chosen: types.Elf, card: "Raging Goblin", want: false},
		{name: "Elf hides Changeling Outcast", controller: game.Player1, chosen: types.Elf, card: "Changeling Outcast", want: false},
		{name: "Goblin reveals Grizzly Bears", controller: game.Player1, chosen: types.Goblin, card: "Grizzly Bears", want: true},
		{name: "opponent's Elf", controller: game.Player2, chosen: types.Elf, card: "Raging Goblin", want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, engine := spellTypeScenario(t)
			addConspiracy(t, g, tc.controller, tc.chosen)
			cardID := addProductionCardToLibrary(t, g, game.Player1, tc.card)

			resolveGoblinMatronSearch(t, g, engine, tc.card)

			if got := g.Players[game.Player1].Hand.Contains(cardID); got != tc.want {
				t.Fatalf("%s found = %v, want %v", tc.card, got, tc.want)
			}
		})
	}
}

// TestConspiracyAndArcaneAdaptationApplyInTimestampOrder is acceptance
// scenario 4 (CR 613.7, no dependency per 613.8a): Conspiracy (Elf) then Arcane
// Adaptation (Goblin) makes creatures, creature spells and owned creature cards
// Elf Goblin; the reverse order makes them only Elf.
func TestConspiracyAndArcaneAdaptationApplyInTimestampOrder(t *testing.T) {
	for _, conspiracyFirst := range []bool{true, false} {
		g, engine := spellTypeScenario(t)
		if conspiracyFirst {
			addConspiracy(t, g, game.Player1, types.Elf)
			addChosenTypePermanent(t, g, game.Player1, "Arcane Adaptation", types.Goblin)
		} else {
			addChosenTypePermanent(t, g, game.Player1, "Arcane Adaptation", types.Goblin)
			addConspiracy(t, g, game.Player1, types.Elf)
		}
		want := []types.Sub{types.Elf}
		if conspiracyFirst {
			want = append(want, types.Goblin)
		}
		creature := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Grizzly Bears"))
		libraryCard := addProductionCardToLibrary(t, g, game.Player1, "Grizzly Bears")
		spell := castSpellTypeTestCard(t, g, engine, game.Player1, "Grizzly Bears", nil, mana.G, mana.C)

		for _, state := range []*game.Game{g, g.Clone()} {
			assertCreatureSubtypes(t, "creature", effectivePermanentValues(state, creature).subtypes, want...)
			top, _ := state.Stack.Peek()
			if top == nil || top.ID != spell.ID {
				t.Fatal("Grizzly Bears spell is not on the stack")
			}
			assertCreatureSubtypes(t, "spell", spellCharacteristics(t, state, top).Subtypes, want...)
			card, _ := state.GetCardInstance(libraryCard)
			assertCreatureSubtypes(t, "library card", offBattlefieldCardTypeLine(state, card).subtypes, want...)
		}
	}
}

// TestConspiracyOverwritesLaterCharacteristicDefiningAbility proves CR 613.3:
// Mistform Ultimus's "is every creature type" CDA applies before Conspiracy
// even though Mistform entered later, so it is only an Elf on the battlefield,
// on the stack and in the library.
func TestConspiracyOverwritesLaterCharacteristicDefiningAbility(t *testing.T) {
	g, engine := spellTypeScenario(t)
	addConspiracy(t, g, game.Player1, types.Elf)
	mistform := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Mistform Ultimus"))
	assertCreatureSubtypes(t, "Mistform Ultimus", effectivePermanentValues(g, mistform).subtypes, types.Elf)

	libraryCard := addProductionCardToLibrary(t, g, game.Player1, "Mistform Ultimus")
	card, _ := g.GetCardInstance(libraryCard)
	assertCreatureSubtypes(t, "Mistform Ultimus card", offBattlefieldCardTypeLine(g, card).subtypes, types.Elf)

	spell := castSpellTypeTestCard(t, g, engine, game.Player1, "Mistform Ultimus", nil, mana.U, mana.C, mana.C, mana.C)
	assertCreatureSubtypes(t, "Mistform Ultimus spell", spellCharacteristics(t, g, spell).Subtypes, types.Elf)
}

// TestConspiracyLeavingRestoresOriginalTypes is acceptance scenario 5 (CR
// 611.3a-b): once Conspiracy leaves the battlefield, the creature, the spell on
// the stack and the library card have their original creature types again.
func TestConspiracyLeavingRestoresOriginalTypes(t *testing.T) {
	g, engine := spellTypeScenario(t)
	conspiracy := addConspiracy(t, g, game.Player1, types.Elf)
	creature := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Goblin Matron"))
	libraryCard := addProductionCardToLibrary(t, g, game.Player1, "Raging Goblin")
	spell := castSpellTypeTestCard(t, g, engine, game.Player1, "Grizzly Bears", nil, mana.G, mana.C)
	card, _ := g.GetCardInstance(libraryCard)

	assertCreatureSubtypes(t, "creature", effectivePermanentValues(g, creature).subtypes, types.Elf)
	assertCreatureSubtypes(t, "spell", spellCharacteristics(t, g, spell).Subtypes, types.Elf)
	assertCreatureSubtypes(t, "library card", offBattlefieldCardTypeLine(g, card).subtypes, types.Elf)

	removePermanentForSpellTypeTest(t, g, engine, conspiracy)

	assertCreatureSubtypes(t, "creature", effectivePermanentValues(g, creature).subtypes, types.Goblin)
	assertCreatureSubtypes(t, "spell", spellCharacteristics(t, g, spell).Subtypes, types.Bear)
	assertCreatureSubtypes(t, "library card", offBattlefieldCardTypeLine(g, card).subtypes, types.Goblin, types.Berserker)
}

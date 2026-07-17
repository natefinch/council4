package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

func selectionHasSubtype(selection TriggerSelection, name string) bool {
	for _, sub := range selection.SubtypesAny {
		if string(sub) == name {
			return true
		}
	}
	return false
}

func selectionSpellHasSubtype(selection TriggerEventSpellSelection, name string) bool {
	for _, sub := range selection.SubtypesAny {
		if string(sub) == name {
			return true
		}
	}
	return false
}

type triggerEventClauseTest struct {
	name     string
	source   string
	cardName string
	check    func(*testing.T, *TriggerEventClause)
}

func TestTriggerEventClauses(t *testing.T) {
	t.Parallel()
	for _, test := range triggerEventClauseTests() {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			trigger := parseTriggerEventFromSource(t, test.source, test.cardName)
			if trigger == nil {
				t.Fatal("trigger = nil")
			}
			test.check(t, trigger)
		})
	}
}

func triggerEventClauseTests() []triggerEventClauseTest {
	tests := zoneChangeTriggerEventClauseTests()
	tests = append(tests, spellAndAbilityTriggerEventClauseTests()...)
	tests = append(tests, spellTypeDisjunctionTriggerEventClauseTests()...)
	tests = append(tests, actorOrdinalSpellTriggerEventClauseTests()...)
	tests = append(tests, chosenTypeSpellTriggerEventClauseTests()...)
	tests = append(tests, chosenColorSpellTriggerEventClauseTests()...)
	tests = append(tests, combatTriggerEventClauseTests()...)
	tests = append(tests, blockUnionTriggerEventClauseTests()...)
	tests = append(tests, enterAttackUnionTriggerEventClauseTests()...)
	tests = append(tests, chosenTypeZoneChangeTriggerEventClauseTests()...)
	tests = append(tests, damageAndCounterTriggerEventClauseTests()...)
	tests = append(tests, subjectQualifierTriggerEventClauseTests()...)
	tests = append(tests, stateAndOtherTriggerEventClauseTests()...)
	return tests
}

func subjectQualifierTriggerEventClauseTests() []triggerEventClauseTest {
	return []triggerEventClauseTest{
		{
			name:   "subject base power enters",
			source: "Whenever another creature you control with base power 1 enters, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				power := clause.Subject.Selection.Power
				if power.Comparison != TriggerSelectionComparisonEqual || power.Value != 1 {
					t.Fatalf("base power = %#v, want {Equal, 1}", power)
				}
			},
		},
		{
			name:   "subject any counter combat damage",
			source: "Whenever a creature you control with a counter on it deals combat damage to a player, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				selection := clause.DamageSource.Selection
				if !selection.MatchAnyCounter {
					t.Fatalf("selection = %#v, want MatchAnyCounter", selection)
				}
			},
		},
		{
			name:   "subject power above base combat damage",
			source: "Whenever one or more creatures you control each with power greater than its base power deals combat damage to a player, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				selection := clause.DamageSource.Selection
				if !selection.PowerAboveBase {
					t.Fatalf("selection = %#v, want PowerAboveBase (Kutzil, Malamet Exemplar)", selection)
				}
			},
		},
		{
			name:   "enchanted creature combat damage",
			source: "Whenever an enchanted creature you control deals combat damage to a player, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.DamageSource.Kind != TriggerEventSubjectSelection ||
					!clause.DamageSource.Selection.Enchanted ||
					clause.Controller != ControllerYou {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
	}
}

func chosenTypeSpellTriggerEventClauseTests() []triggerEventClauseTest {
	return []triggerEventClauseTest{
		{
			name:   "spell creature of the chosen type",
			source: "Whenever you cast a creature spell of the chosen type, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if len(clause.SpellSelection.Types) != 1 ||
					clause.SpellSelection.Types[0] != TriggerCardTypeCreature ||
					!clause.SpellSelection.SubtypeFromEntryChoice {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "spell of the chosen type",
			source: "Whenever you cast a spell of the chosen type, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.SpellSelection.SubtypeFromEntryChoice ||
					len(clause.SpellSelection.Types) != 0 {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
	}
}

func chosenColorSpellTriggerEventClauseTests() []triggerEventClauseTest {
	return []triggerEventClauseTest{
		{
			name:   "spell of the chosen color you",
			source: "Whenever you cast a spell of the chosen color, you gain 1 life.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.SpellSelection.ColorFromEntryChoice ||
					len(clause.SpellSelection.ColorsAny) != 0 ||
					clause.Actor.Kind != TriggerEventActorYou {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "spell of the chosen color player",
			source: "Whenever a player casts a spell of the chosen color, that player loses 1 life.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.SpellSelection.ColorFromEntryChoice ||
					clause.Actor.Kind != TriggerEventActorPlayer {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
	}
}

func chosenTypeZoneChangeTriggerEventClauseTests() []triggerEventClauseTest {
	return []triggerEventClauseTest{
		{
			name:   "zone chosen type dies",
			source: "Whenever a creature of the chosen type dies, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Subject.Kind != TriggerEventSubjectSelection ||
					clause.Zone.ToZone.Kind != TriggerEventZoneGraveyard ||
					!selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) ||
					!clause.Subject.Selection.SubtypeFromEntryChoice {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
	}
}

func zoneChangeTriggerEventClauseTests() []triggerEventClauseTest {
	return []triggerEventClauseTest{
		{
			name:   "zone self enters tapped",
			source: "When this creature enters tapped, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindZoneChange || clause.Subject.Kind != TriggerEventSubjectSelf || !clause.Zone.MatchToZone || clause.Zone.ToZone.Kind != TriggerEventZoneBattlefield || clause.Tapped.Kind != TriggerEventTappedStateTapped {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone selection enters battlefield",
			source: "Whenever a creature enters the battlefield, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Subject.Kind != TriggerEventSubjectSelection || !selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone bare token enters battlefield",
			source: "Whenever a token enters the battlefield, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Subject.Kind != TriggerEventSubjectSelection || !clause.Subject.Selection.TokenOnly {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone from graveyard",
			source: "Whenever a creature enters from your graveyard, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Player.Kind != TriggerPlayerSelectorYou || clause.Zone.FromZone.Kind != TriggerEventZoneGraveyard {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone multi-origin exile from library and or graveyard",
			source: "Whenever one or more cards are put into exile from your library and/or your graveyard, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.OneOrMore ||
					clause.Player.Kind != TriggerPlayerSelectorYou ||
					clause.ZoneChange.Kind != TriggerEventZoneChangeMoved ||
					clause.Zone.MatchFromZone ||
					clause.Zone.ExcludeFromZone ||
					!clause.Zone.MatchToZone ||
					clause.Zone.ToZone.Kind != TriggerEventZoneExile ||
					!zoneKindsEqual(clause.Zone.FromZones, TriggerEventZoneLibrary, TriggerEventZoneGraveyard) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone multi-origin exile from graveyard or library with or joiner",
			source: "Whenever one or more cards are put into exile from your graveyard or your library, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.OneOrMore ||
					clause.Player.Kind != TriggerPlayerSelectorYou ||
					!clause.Zone.MatchToZone ||
					clause.Zone.ToZone.Kind != TriggerEventZoneExile ||
					!zoneKindsEqual(clause.Zone.FromZones, TriggerEventZoneGraveyard, TriggerEventZoneLibrary) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:     "zone named self dies",
			source:   "Whenever Ravenous Baloth dies, draw a card.",
			cardName: "Ravenous Baloth",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Subject.Kind != TriggerEventSubjectSelf || clause.Zone.ToZone.Kind != TriggerEventZoneGraveyard {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone selection dies adds creature",
			source: "Whenever another permanent dies, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.ExcludeSelf || !selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone self or another selection enters",
			source: "Whenever this creature or another Ally you control enters, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.SelfOrAnother || clause.ExcludeSelf ||
					clause.Subject.Kind != TriggerEventSubjectSelection ||
					clause.Controller != ControllerYou ||
					!selectionHasSubtype(clause.Subject.Selection, "Ally") {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:     "zone named self or another selection dies",
			source:   "Whenever Omnath or another Elemental you control dies, draw a card.",
			cardName: "Omnath",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.SelfOrAnother || clause.ExcludeSelf ||
					clause.Subject.Kind != TriggerEventSubjectSelection ||
					clause.Controller != ControllerYou ||
					!selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) ||
					!selectionHasSubtype(clause.Subject.Selection, "Elemental") {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone self dies or another selection put into graveyard",
			source: "Whenever this creature dies or another artifact you control is put into a graveyard from the battlefield, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.SelfOrAnother || clause.ExcludeSelf ||
					clause.Subject.Kind != TriggerEventSubjectSelection ||
					clause.Controller != ControllerYou ||
					clause.ZoneChange.Kind != TriggerEventZoneChangeMoved ||
					clause.Zone.FromZone.Kind != TriggerEventZoneBattlefield ||
					clause.Zone.ToZone.Kind != TriggerEventZoneGraveyard ||
					!selectionHasType(clause.Subject.Selection, TriggerCardTypeArtifact) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone self dies or another selection dies two verbs",
			source: "Whenever this creature dies or another creature you control dies, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.SelfOrAnother || clause.ExcludeSelf ||
					clause.Subject.Kind != TriggerEventSubjectSelection ||
					clause.Controller != ControllerYou ||
					clause.Zone.FromZone.Kind != TriggerEventZoneBattlefield ||
					clause.Zone.ToZone.Kind != TriggerEventZoneGraveyard ||
					!selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone attached noncreature dies adds creature",
			source: "Whenever enchanted artifact dies, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Subject.Kind != TriggerEventSubjectAttached ||
					!selectionHasType(clause.Subject.Selection, TriggerCardTypeArtifact) ||
					!selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone leaves battlefield",
			source: "Whenever an artifact leaves the battlefield, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Zone.FromZone.Kind != TriggerEventZoneBattlefield || clause.Zone.MatchToZone {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone cards leave your graveyard",
			source: "Whenever one or more cards leave your graveyard, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.ZoneChange.Kind != TriggerEventZoneChangeMoved ||
					!clause.OneOrMore ||
					clause.Subject.Kind != TriggerEventSubjectSelection ||
					clause.Zone.FromZone.Kind != TriggerEventZoneGraveyard ||
					clause.Zone.MatchToZone ||
					clause.Player.Kind != TriggerPlayerSelectorYou {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone creature cards leave your graveyard",
			source: "Whenever one or more creature cards leave your graveyard, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.ZoneChange.Kind != TriggerEventZoneChangeMoved ||
					!clause.OneOrMore ||
					clause.Zone.FromZone.Kind != TriggerEventZoneGraveyard ||
					clause.Player.Kind != TriggerPlayerSelectorYou ||
					!selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone put into graveyard",
			source: "Whenever an artifact is put into a graveyard, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Zone.FromZone.Kind != TriggerEventZoneBattlefield || clause.Zone.ToZone.Kind != TriggerEventZoneGraveyard {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone put into opponent graveyard",
			source: "Whenever an artifact is put into an opponent's graveyard from the battlefield, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Zone.FromZone.Kind != TriggerEventZoneBattlefield ||
					clause.Zone.ToZone.Kind != TriggerEventZoneGraveyard ||
					clause.Player.Kind != TriggerPlayerSelectorOpponent {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone put into your graveyard from anywhere",
			source: "Whenever a creature card is put into your graveyard from anywhere, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.ZoneChange.Kind != TriggerEventZoneChangeMoved ||
					clause.Zone.MatchFromZone ||
					!clause.Zone.MatchToZone ||
					clause.Zone.ToZone.Kind != TriggerEventZoneGraveyard ||
					clause.Player.Kind != TriggerPlayerSelectorYou ||
					!selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone put into graveyard excluding battlefield",
			source: "Whenever a creature card is put into a graveyard from anywhere other than the battlefield, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.ZoneChange.Kind != TriggerEventZoneChangeMoved ||
					clause.Zone.MatchFromZone ||
					!clause.Zone.ExcludeFromZone ||
					clause.Zone.FromZone.Kind != TriggerEventZoneBattlefield ||
					!clause.Zone.MatchToZone ||
					clause.Zone.ToZone.Kind != TriggerEventZoneGraveyard ||
					!selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone exiled",
			source: "Whenever an artifact is exiled from the battlefield, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Zone.ToZone.Kind != TriggerEventZoneExile {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone returned to hand",
			source: "Whenever an artifact is returned to your hand, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Zone.ToZone.Kind != TriggerEventZoneHand || clause.Player.Kind != TriggerPlayerSelectorYou {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "zone plural returned to owners hands",
			source: "Whenever artifacts are returned to their owners' hands, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindZoneChange ||
					clause.Zone.ToZone.Kind != TriggerEventZoneHand ||
					!selectionHasType(clause.Subject.Selection, TriggerCardTypeArtifact) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
	}
}

func spellTypeDisjunctionTriggerEventClauseTests() []triggerEventClauseTest {
	return []triggerEventClauseTest{
		{
			name:   "spell card-type disjunction",
			source: "Whenever you cast an artifact or enchantment spell, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if len(clause.SpellSelection.TypesAny) != 2 ||
					clause.SpellSelection.TypesAny[0] != TriggerCardTypeArtifact ||
					clause.SpellSelection.TypesAny[1] != TriggerCardTypeEnchantment {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "spell three-way card-type disjunction",
			source: "Whenever you cast an artifact, creature, or enchantment spell, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if len(clause.SpellSelection.TypesAny) != 3 ||
					clause.SpellSelection.TypesAny[0] != TriggerCardTypeArtifact ||
					clause.SpellSelection.TypesAny[1] != TriggerCardTypeCreature ||
					clause.SpellSelection.TypesAny[2] != TriggerCardTypeEnchantment {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "spell subtype disjunction",
			source: "Whenever you cast an Aura or Equipment spell, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if len(clause.SpellSelection.SubtypesAny) != 2 ||
					!selectionSpellHasSubtype(clause.SpellSelection, "Aura") ||
					!selectionSpellHasSubtype(clause.SpellSelection, "Equipment") {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "spell three-way subtype disjunction",
			source: "Whenever you cast an Aura, Equipment, or Vehicle spell, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if len(clause.SpellSelection.SubtypesAny) != 3 ||
					!selectionSpellHasSubtype(clause.SpellSelection, "Aura") ||
					!selectionSpellHasSubtype(clause.SpellSelection, "Equipment") ||
					!selectionSpellHasSubtype(clause.SpellSelection, "Vehicle") {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "spell color disjunction",
			source: "Whenever an opponent casts a blue or black spell, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if len(clause.SpellSelection.ColorsAny) != 2 ||
					clause.SpellSelection.ColorsAny[0] != TriggerColorBlue ||
					clause.SpellSelection.ColorsAny[1] != TriggerColorBlack {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "spell three-way color disjunction",
			source: "Whenever you cast a blue, black, or red spell, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if len(clause.SpellSelection.ColorsAny) != 3 ||
					clause.SpellSelection.ColorsAny[0] != TriggerColorBlue ||
					clause.SpellSelection.ColorsAny[1] != TriggerColorBlack ||
					clause.SpellSelection.ColorsAny[2] != TriggerColorRed {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
	}
}

func spellAndAbilityTriggerEventClauseTests() []triggerEventClauseTest {
	return []triggerEventClauseTest{
		{
			name:   "spell any spell",
			source: "Whenever you cast a spell, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindSpellCast || clause.Actor.Kind != TriggerEventActorYou {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "self cast this spell",
			source: "When you cast this spell, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindSpellCast || clause.Actor.Kind != TriggerEventActorYou || !clause.SelfCast {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "spell cast or copy",
			source: "Whenever you cast or copy an instant or sorcery spell, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindSpellCast || clause.Actor.Kind != TriggerEventActorYou || !clause.MatchCopy {
					t.Fatalf("clause = %#v", clause)
				}
				if len(clause.SpellSelection.TypesAny) != 2 ||
					clause.SpellSelection.TypesAny[0] != TriggerCardTypeInstant ||
					clause.SpellSelection.TypesAny[1] != TriggerCardTypeSorcery {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "spell typed spell",
			source: "Whenever you cast a creature spell, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if len(clause.SpellSelection.Types) != 1 || clause.SpellSelection.Types[0] != TriggerCardTypeCreature {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "spell colored spell",
			source: "Whenever you cast a white spell, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if len(clause.SpellSelection.ColorsAny) != 1 || clause.SpellSelection.ColorsAny[0] != TriggerColorWhite {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "spell historic",
			source: "Whenever you cast a historic spell, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.SpellSelection.Historic {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "spell kicked",
			source: "Whenever you cast a kicked spell, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.SpellSelection.Kicker {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "spell from graveyard",
			source: "Whenever you cast a spell from your graveyard, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.SpellSelection.FromZone.Kind != TriggerEventZoneGraveyard {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "spell mana value",
			source: "Whenever you cast a spell with mana value 4 or greater, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.SpellSelection.MatchManaValue || clause.SpellSelection.ManaValueAtLeast != 4 {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "creature spell mana value",
			source: "Whenever you cast a creature spell with mana value 6 or greater, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.SpellSelection.MatchManaValue || clause.SpellSelection.ManaValueAtLeast != 6 {
					t.Fatalf("clause = %#v", clause)
				}
				if len(clause.SpellSelection.Types) != 1 || clause.SpellSelection.Types[0] != TriggerCardTypeCreature {
					t.Fatalf("types = %#v", clause.SpellSelection.Types)
				}
			},
		},
		{
			name:   "colorless spell mana value",
			source: "Whenever you cast a colorless spell with mana value 7 or greater, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.SpellSelection.MatchManaValue || clause.SpellSelection.ManaValueAtLeast != 7 {
					t.Fatalf("clause = %#v", clause)
				}
				if !clause.SpellSelection.Colorless {
					t.Fatalf("colorless = false, clause = %#v", clause)
				}
			},
		},
		{
			name:   "artifact spell mana value",
			source: "Whenever you cast an artifact spell with mana value 5 or greater, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.SpellSelection.MatchManaValue || clause.SpellSelection.ManaValueAtLeast != 5 {
					t.Fatalf("clause = %#v", clause)
				}
				if len(clause.SpellSelection.Types) != 1 || clause.SpellSelection.Types[0] != TriggerCardTypeArtifact {
					t.Fatalf("types = %#v", clause.SpellSelection.Types)
				}
			},
		},
		{
			name:   "creature spell mana value at most",
			source: "Whenever you cast a creature spell with mana value 3 or less, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.SpellSelection.MatchManaValue || clause.SpellSelection.ManaValueAtMost != 3 {
					t.Fatalf("clause = %#v", clause)
				}
				if clause.SpellSelection.ManaValueAtLeast != 0 {
					t.Fatalf("ManaValueAtLeast = %d, want 0", clause.SpellSelection.ManaValueAtLeast)
				}
				if len(clause.SpellSelection.Types) != 1 || clause.SpellSelection.Types[0] != TriggerCardTypeCreature {
					t.Fatalf("types = %#v", clause.SpellSelection.Types)
				}
			},
		},
		{
			name:   "spell mana value at most or fewer",
			source: "Whenever you cast a spell with mana value 5 or fewer, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.SpellSelection.MatchManaValue || clause.SpellSelection.ManaValueAtMost != 5 {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "noncreature spell mana value less than source power",
			source: "Whenever an opponent casts a noncreature spell with mana value less than this creature's power, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.SpellSelection.ManaValueLessThanSourcePower {
					t.Fatalf("ManaValueLessThanSourcePower = false, clause = %#v", clause)
				}
				if clause.SpellSelection.MatchManaValue ||
					clause.SpellSelection.ManaValueAtLeast != 0 ||
					clause.SpellSelection.ManaValueAtMost != 0 {
					t.Fatalf("unexpected fixed mana-value bound, clause = %#v", clause)
				}
				if clause.Actor.Kind != TriggerEventActorOpponent {
					t.Fatalf("actor = %#v, want opponent", clause.Actor)
				}
				if len(clause.SpellSelection.ExcludedTypes) != 1 ||
					clause.SpellSelection.ExcludedTypes[0] != TriggerCardTypeCreature {
					t.Fatalf("excluded types = %#v, want creature", clause.SpellSelection.ExcludedTypes)
				}
			},
		},
		{
			name:   "spell mana value less than source power",
			source: "Whenever you cast a spell with mana value less than this creature's power, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.SpellSelection.ManaValueLessThanSourcePower {
					t.Fatalf("ManaValueLessThanSourcePower = false, clause = %#v", clause)
				}
			},
		},
		{
			name:   "spell ordinal first",
			source: "Whenever you cast your first spell each turn, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindSpellCast || clause.Actor.Kind != TriggerEventActorYou {
					t.Fatalf("clause = %#v", clause)
				}
				if clause.SpellSelection.Ordinal != 1 {
					t.Fatalf("ordinal = %d, want 1", clause.SpellSelection.Ordinal)
				}
			},
		},
		{
			name:   "spell ordinal second",
			source: "Whenever you cast your second spell each turn, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.SpellSelection.Ordinal != 2 {
					t.Fatalf("ordinal = %d, want 2", clause.SpellSelection.Ordinal)
				}
			},
		},
		{
			name:   "spell creature subtype",
			source: "Whenever you cast an Elf spell, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindSpellCast || clause.Actor.Kind != TriggerEventActorYou {
					t.Fatalf("clause = %#v", clause)
				}
				if len(clause.SpellSelection.SubtypesAny) != 1 ||
					!selectionSpellHasSubtype(clause.SpellSelection, "Elf") {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "spell creature subtype player actor",
			source: "Whenever a player casts a Goblin spell, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindSpellCast || clause.Actor.Kind != TriggerEventActorPlayer {
					t.Fatalf("clause = %#v", clause)
				}
				if len(clause.SpellSelection.SubtypesAny) != 1 ||
					!selectionSpellHasSubtype(clause.SpellSelection, "Goblin") {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "ability activated simple",
			source: "Whenever you activate an ability that isn't a mana ability, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindAbilityActivated || clause.Actor.Kind != TriggerEventActorYou || !clause.ExcludeManaAbility {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "ability activated with source selection",
			source: "Whenever an opponent activates an ability of a creature that isn't a mana ability, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Actor.Kind != TriggerEventActorOpponent || !selectionHasType(clause.SourceSelection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
	}
}

func actorOrdinalSpellTriggerEventClauseTests() []triggerEventClauseTest {
	return []triggerEventClauseTest{
		{
			name:   "spell ordinal second player actor",
			source: "Whenever a player casts their second spell each turn, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindSpellCast || clause.Actor.Kind != TriggerEventActorPlayer {
					t.Fatalf("clause = %#v", clause)
				}
				if clause.SpellSelection.Ordinal != 2 {
					t.Fatalf("ordinal = %d, want 2", clause.SpellSelection.Ordinal)
				}
			},
		},
		{
			name:   "spell ordinal third opponent actor",
			source: "Whenever an opponent casts their third spell each turn, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Actor.Kind != TriggerEventActorOpponent {
					t.Fatalf("clause = %#v", clause)
				}
				if clause.SpellSelection.Ordinal != 3 {
					t.Fatalf("ordinal = %d, want 3", clause.SpellSelection.Ordinal)
				}
			},
		},
		{
			name:   "spell ordinal filtered opponent actor",
			source: "Whenever an opponent casts their second instant spell each turn, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Actor.Kind != TriggerEventActorOpponent {
					t.Fatalf("clause = %#v", clause)
				}
				if clause.SpellSelection.Ordinal != 2 ||
					len(clause.SpellSelection.Types) != 1 ||
					clause.SpellSelection.Types[0] != TriggerCardTypeInstant {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
	}
}

func enterAttackUnionTriggerEventClauseTests() []triggerEventClauseTest {
	return []triggerEventClauseTest{
		{
			name:   "enter or attack union self",
			source: "Whenever this creature enters or attacks, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindZoneChange ||
					clause.ZoneChange.Kind != TriggerEventZoneChangeEnteredBattlefield ||
					clause.UnionKind != TriggerEventKindAttack ||
					clause.Subject.Kind != TriggerEventSubjectSelf {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "enter or dies union self",
			source: "When this creature enters or dies, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindZoneChange ||
					clause.ZoneChange.Kind != TriggerEventZoneChangeEnteredBattlefield ||
					clause.UnionKind != TriggerEventKindDied ||
					clause.Subject.Kind != TriggerEventSubjectSelf {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "dies or enter union self mirror",
			source: "When this creature dies or enters, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindZoneChange ||
					clause.UnionKind != TriggerEventKindDied ||
					clause.Subject.Kind != TriggerEventSubjectSelf {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "attack or enter union self mirror",
			source: "Whenever this creature attacks or enters, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindZoneChange ||
					clause.UnionKind != TriggerEventKindAttack ||
					clause.Subject.Kind != TriggerEventSubjectSelf {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "enter or put into graveyard union self",
			source: "When this artifact enters or is put into a graveyard from the battlefield, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindZoneChange ||
					clause.ZoneChange.Kind != TriggerEventZoneChangeEnteredBattlefield ||
					clause.UnionKind != TriggerEventKindDied ||
					clause.Subject.Kind != TriggerEventSubjectSelf {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "enter or attack union selection",
			source: "Whenever a creature you control enters or attacks, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindZoneChange ||
					clause.UnionKind != TriggerEventKindAttack ||
					clause.Subject.Kind != TriggerEventSubjectSelection ||
					!selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) ||
					clause.Controller != ControllerYou {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "enter or attack union chosen type selection",
			source: "Whenever a creature you control of the chosen type enters or attacks, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindZoneChange ||
					clause.UnionKind != TriggerEventKindAttack ||
					clause.Subject.Kind != TriggerEventSubjectSelection ||
					!selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) ||
					!clause.Subject.Selection.SubtypeFromEntryChoice ||
					clause.Controller != ControllerYou {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "attack or became target of a spell union self",
			source: "Whenever this creature attacks or becomes the target of a spell, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindBecameTarget ||
					clause.UnionKind != TriggerEventKindAttack ||
					clause.Subject.Kind != TriggerEventSubjectSelf ||
					clause.StackObject.Kind != TriggerEventStackObjectSpell {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "attack or became target of a spell or ability union self",
			source: "Whenever this creature attacks or becomes the target of a spell or ability, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindBecameTarget ||
					clause.UnionKind != TriggerEventKindAttack ||
					clause.Subject.Kind != TriggerEventSubjectSelf ||
					clause.StackObject.Kind != TriggerEventStackObjectAny {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
	}
}

func combatTriggerEventClauseTests() []triggerEventClauseTest {
	return []triggerEventClauseTest{
		{
			name:   "attack enchanted player is attacked",
			source: "Whenever enchanted player is attacked, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindAttack ||
					!clause.EnchantedPlayerIsAttacked ||
					!clause.OneOrMore ||
					clause.Subject.Kind != TriggerEventSubjectUnknown ||
					clause.Actor.Kind != TriggerEventActorUnknown ||
					clause.AttackRecipient.Kind != TriggerEventAttackRecipientAny {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "attack player attacks",
			source: "Whenever you attack, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindAttack || clause.Actor.Kind != TriggerEventActorYou || !clause.OneOrMore {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "attack self attacks",
			source: "Whenever this creature attacks, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Subject.Kind != TriggerEventSubjectSelf {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:     "attack self attacks by legendary name with comma",
			source:   "Whenever Etali, Primal Storm attacks, draw a card.",
			cardName: "Etali, Primal Storm",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindAttack ||
					clause.Subject.Kind != TriggerEventSubjectSelf {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "attack self attacks alone",
			source: "Whenever this creature attacks alone, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindAttack ||
					clause.Subject.Kind != TriggerEventSubjectSelf ||
					!clause.AttackAlone {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "attack self attacks while saddled",
			source: "Whenever this creature attacks while saddled, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindAttack ||
					clause.Subject.Kind != TriggerEventSubjectSelf ||
					!clause.AttackWhileSaddled {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "attack selected attacks alone",
			source: "Whenever a creature you control attacks alone, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindAttack ||
					clause.Subject.Kind != TriggerEventSubjectSelection ||
					!clause.AttackAlone ||
					!selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "attack with two or more creatures",
			source: "Whenever you attack with two or more creatures, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindAttack ||
					clause.Actor.Kind != TriggerEventActorYou ||
					!clause.OneOrMore ||
					clause.AttackerCountAtLeast != 2 {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "another player attacks with two or more creatures",
			source: "Whenever another player attacks with two or more creatures, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindAttack ||
					clause.Actor.Kind != TriggerEventActorOpponent ||
					!clause.OneOrMore ||
					clause.AttackerCountAtLeast != 2 {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "attack one or more creatures",
			source: "Whenever one or more creatures attack, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.OneOrMore || clause.Subject.Kind != TriggerEventSubjectSelection {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "battalion self and at least two other creatures attack",
			source: "Whenever this creature and at least two other creatures attack, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindAttack ||
					clause.Subject.Kind != TriggerEventSubjectSelf ||
					clause.OneOrMore ||
					clause.ExcludeSelf ||
					clause.AttackerCountAtLeast != 3 {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "attack creature subtype",
			source: "Whenever a Goblin you control attacks, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindAttack ||
					len(clause.Subject.Selection.SubtypesAny) != 1 ||
					clause.Subject.Selection.SubtypesAny[0] != "Goblin" {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "attack attached artifact",
			source: "Whenever enchanted artifact attacks, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindAttack ||
					clause.Subject.Kind != TriggerEventSubjectAttached ||
					!selectionHasType(clause.Subject.Selection, TriggerCardTypeArtifact) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "block bare token",
			source: "Whenever a token you control blocks, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindBlock || !clause.Subject.Selection.TokenOnly {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "attack player recipient",
			source: "Whenever an opponent attacks you, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Actor.Kind != TriggerEventActorOpponent || clause.Player.Kind != TriggerPlayerSelectorYou || clause.AttackRecipient.Kind != TriggerEventAttackRecipientPlayer {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "attack permanent recipient",
			source: "Whenever this creature attacks a player or planeswalker, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.AttackRecipient.Kind != TriggerEventAttackRecipientPlayer|TriggerEventAttackRecipientPlaneswalker {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "block self",
			source: "Whenever this creature blocks, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindBlock || clause.Subject.Kind != TriggerEventSubjectSelf {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "block another related selection",
			source: "Whenever this creature blocks another creature, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !selectionHasType(clause.RelatedSelection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "became blocked",
			source: "Whenever this creature becomes blocked by a creature, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindBecameBlocked || clause.Subject.Kind != TriggerEventSubjectSelf || !selectionHasType(clause.RelatedSelection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "one or more controlled creatures fight",
			source: "Whenever one or more creatures you control fight, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindFight ||
					!clause.OneOrMore ||
					clause.Controller != ControllerYou ||
					!selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
	}
}

func blockUnionTriggerEventClauseTests() []triggerEventClauseTest {
	return []triggerEventClauseTest{
		{
			name:   "blocks or becomes blocked union self",
			source: "Whenever this creature blocks or becomes blocked, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindBlock ||
					clause.UnionKind != TriggerEventKindBecameBlocked ||
					clause.Subject.Kind != TriggerEventSubjectSelf {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "becomes blocked or blocks union self",
			source: "Whenever this creature becomes blocked or blocks, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindBlock ||
					clause.UnionKind != TriggerEventKindBecameBlocked ||
					clause.Subject.Kind != TriggerEventSubjectSelf {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "attacks or blocks union self",
			source: "Whenever this creature attacks or blocks, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindAttack ||
					clause.UnionKind != TriggerEventKindBlock ||
					clause.Subject.Kind != TriggerEventSubjectSelf {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "blocks or attacks union self mirror",
			source: "Whenever this creature blocks or attacks, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindAttack ||
					clause.UnionKind != TriggerEventKindBlock ||
					clause.Subject.Kind != TriggerEventSubjectSelf {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "blocks or becomes blocked by a creature union self",
			source: "Whenever this creature blocks or becomes blocked by a creature, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindBlock ||
					clause.UnionKind != TriggerEventKindBecameBlocked ||
					clause.Subject.Kind != TriggerEventSubjectSelf ||
					!selectionHasType(clause.RelatedSelection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "blocks or becomes blocked by a white creature union self",
			source: "Whenever this creature blocks or becomes blocked by a white creature, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindBlock ||
					clause.UnionKind != TriggerEventKindBecameBlocked ||
					clause.Subject.Kind != TriggerEventSubjectSelf ||
					!selectionHasType(clause.RelatedSelection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "fight or become blocked one or more controlled creatures",
			source: "Whenever one or more creatures you control fight or become blocked, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindFight ||
					clause.UnionKind != TriggerEventKindBecameBlocked ||
					!clause.OneOrMore ||
					clause.Controller != ControllerYou ||
					!selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "becomes blocked or fights singular mirror",
			source: "Whenever a creature you control becomes blocked or fights, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindFight ||
					clause.UnionKind != TriggerEventKindBecameBlocked ||
					clause.OneOrMore ||
					clause.Controller != ControllerYou ||
					!selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
	}
}

func damageAndCounterTriggerEventClauseTests() []triggerEventClauseTest {
	return []triggerEventClauseTest{
		{
			name:   "damage self deals damage",
			source: "Whenever this creature deals damage, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindDamageDealt || clause.DamageSource.Kind != TriggerEventSubjectSelf {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "damage when self deals damage",
			source: "When this creature deals damage, sacrifice it.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindDamageDealt || clause.DamageSource.Kind != TriggerEventSubjectSelf {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "damage when attached is dealt damage",
			source: "When enchanted creature is dealt damage, destroy it.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Subject.Kind != TriggerEventSubjectAttached ||
					clause.DamageRecipient.Kind != TriggerEventDamageRecipientPermanent {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "damage attached source",
			source: "Whenever enchanted creature deals damage, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.DamageSource.Kind != TriggerEventSubjectAttached || clause.DamageSource.AttachKind != TriggerEventAttachEnchanted {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "damage selection to player",
			source: "Whenever a creature deals damage to a player, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.DamageSource.Kind != TriggerEventSubjectSelection || clause.DamageRecipient.Kind != TriggerEventDamageRecipientPlayer {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "instant or sorcery damage source",
			source: "Whenever an instant or sorcery spell you control deals damage to an opponent, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.DamageSourceIsStackObject ||
					clause.Controller != ControllerYou ||
					len(clause.DamageSourceSpellSelection.TypesAny) != 2 {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "damage source deals to you",
			source: "Whenever a source deals damage to you, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.DamageSource.Kind != TriggerEventSubjectDamageSource || clause.Player.Kind != TriggerPlayerSelectorYou {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "damage self is dealt damage",
			source: "Whenever this creature is dealt damage, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Subject.Kind != TriggerEventSubjectSelf || clause.DamageRecipient.Kind != TriggerEventDamageRecipientPermanent {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "damage combat qualifier",
			source: "Whenever this creature deals combat damage to a player, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.CombatQualifier.Kind != TriggerEventCombatQualifierCombat {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "counter plus one",
			source: "Whenever a +1/+1 counter is put on this creature, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindCounterAdded || clause.Counter.Kind != counter.PlusOnePlusOne {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "counter minus one",
			source: "Whenever a -1/-1 counter is put on this permanent, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Counter.Kind != counter.MinusOneMinusOne {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "counter one or more",
			source: "Whenever one or more +1/+1 counters are put on this creature, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if !clause.OneOrMore {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "counter controller scoped",
			source: "Whenever one or more +1/+1 counters are put on another creature you control, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindCounterAdded ||
					clause.Subject.Kind != TriggerEventSubjectSelection ||
					clause.Controller != ControllerYou ||
					!clause.ExcludeSelf ||
					!clause.OneOrMore ||
					clause.Counter.Kind != counter.PlusOnePlusOne {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "counter singular controller scoped",
			source: "Whenever a +1/+1 counter is put on a creature you control, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindCounterAdded ||
					clause.Subject.Kind != TriggerEventSubjectSelection ||
					clause.Controller != ControllerYou ||
					clause.ExcludeSelf ||
					clause.OneOrMore {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "counter active self",
			source: "Whenever you put one or more +1/+1 counters on this creature, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindCounterAdded ||
					clause.Subject.Kind != TriggerEventSubjectSelf ||
					clause.CauseController != TriggerEventActorYou ||
					!clause.OneOrMore ||
					clause.Counter.Kind != counter.PlusOnePlusOne {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "counter active controller scoped",
			source: "Whenever you put one or more +1/+1 counters on a creature you control, you may draw that many cards.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindCounterAdded ||
					clause.Subject.Kind != TriggerEventSubjectSelection ||
					clause.Controller != ControllerYou ||
					clause.CauseController != TriggerEventActorYou ||
					!clause.OneOrMore ||
					clause.Counter.Kind != counter.PlusOnePlusOne {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "counter active singular other minus one",
			source: "Whenever you put a -1/-1 counter on a creature, create a token.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindCounterAdded ||
					clause.Subject.Kind != TriggerEventSubjectSelection ||
					clause.CauseController != TriggerEventActorYou ||
					clause.OneOrMore ||
					clause.Counter.Kind != counter.MinusOneMinusOne {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "counter any kind self first time each turn",
			source: "Whenever one or more counters are put on this creature for the first time each turn, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindCounterAdded ||
					clause.Subject.Kind != TriggerEventSubjectSelf ||
					!clause.OneOrMore ||
					!clause.FirstTimeEachTurn ||
					clause.Counter.Known {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "counter any kind self plain",
			source: "Whenever one or more counters are put on this creature, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindCounterAdded ||
					clause.Subject.Kind != TriggerEventSubjectSelf ||
					!clause.OneOrMore ||
					clause.FirstTimeEachTurn ||
					clause.Counter.Known {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "counter named kind self first time each turn",
			source: "Whenever one or more +1/+1 counters are put on this creature for the first time each turn, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindCounterAdded ||
					clause.Subject.Kind != TriggerEventSubjectSelf ||
					!clause.OneOrMore ||
					!clause.FirstTimeEachTurn ||
					!clause.Counter.Known ||
					clause.Counter.Kind != counter.PlusOnePlusOne {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
	}
}

func stateAndOtherTriggerEventClauseTests() []triggerEventClauseTest {
	return []triggerEventClauseTest{
		{
			name:   "state self tapped",
			source: "Whenever this creature becomes tapped, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindBecomesTapped || clause.Subject.Kind != TriggerEventSubjectSelf {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "state selection tapped",
			source: "Whenever a creature becomes tapped, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindBecomesTapped || !selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "tapped for mana attached",
			source: "Whenever enchanted land is tapped for mana, its controller adds an additional {G}.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindBecomesTapped ||
					clause.Subject.Kind != TriggerEventSubjectAttached ||
					!clause.TappedForMana {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "tapped for mana selection",
			source: "Whenever a Forest is tapped for mana, its controller adds an additional {G}.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindBecomesTapped ||
					clause.Subject.Kind != TriggerEventSubjectSelection ||
					!clause.TappedForMana {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "tapped for mana active voice",
			source: "Whenever you tap a Swamp for mana, add an additional {B}.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindBecomesTapped ||
					clause.Subject.Kind != TriggerEventSubjectSelection ||
					clause.Controller != ControllerYou ||
					!clause.TappedForMana {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "tapped for mana active voice self",
			source: "Whenever you tap this land for mana, target opponent creates a 1/1 colorless Spirit creature token.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindBecomesTapped ||
					clause.Subject.Kind != TriggerEventSubjectSelf ||
					!clause.TappedForMana {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "tapped for mana any player",
			source: "Whenever a player taps a land for mana, that player adds an additional {U}.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindManaProduced ||
					clause.Subject.Kind != TriggerEventSubjectSelection ||
					clause.Controller != ControllerAny ||
					!clause.TappedForMana {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "tapped for mana opponent",
			source: "Whenever an opponent taps a land for mana, tap all lands that player controls.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindManaProduced ||
					clause.Subject.Kind != TriggerEventSubjectSelection ||
					clause.Controller != ControllerOpponent ||
					!clause.TappedForMana {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "state face up",
			source: "When enchanted creature is turned face up, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindTurnedFaceUp || clause.Subject.Kind != TriggerEventSubjectAttached {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "sacrifice self",
			source: "Whenever you sacrifice this creature, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindSacrificed || clause.Actor.Kind != TriggerEventActorYou || clause.Subject.Kind != TriggerEventSubjectSelf {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "sacrifice selection",
			source: "Whenever you sacrifice a creature, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Subject.Kind != TriggerEventSubjectSelection || !selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "sacrifice player one or more",
			source: "Whenever a player sacrifices one or more creatures, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Actor.Kind != TriggerEventActorPlayer || !clause.OneOrMore {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "mutate",
			source: "Whenever this creature mutates, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindMutated || clause.Subject.Kind != TriggerEventSubjectSelf {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "became target",
			source: "Whenever this creature becomes the target of a spell you control, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindBecameTarget || clause.StackObject.Kind != TriggerEventStackObjectSpell || clause.CauseController != TriggerEventActorYou {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "token created",
			source: "Whenever you create a token, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindTokenCreated || clause.Actor.Kind != TriggerEventActorYou ||
					clause.Subject.Kind != TriggerEventSubjectSelection || !clause.Subject.Selection.TokenOnly ||
					clause.UnionKind != TriggerEventKindUnknown {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "token created one or more",
			source: "Whenever you create one or more tokens, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindTokenCreated || !clause.OneOrMore || !clause.Subject.Selection.TokenOnly {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "token create or sacrifice union",
			source: "Whenever you create or sacrifice a token, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindTokenCreated || clause.UnionKind != TriggerEventKindSacrificed ||
					clause.Actor.Kind != TriggerEventActorYou || !clause.Subject.Selection.TokenOnly {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "token sacrifice or create union",
			source: "Whenever you sacrifice or create a token, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindSacrificed || clause.UnionKind != TriggerEventKindTokenCreated ||
					!clause.Subject.Selection.TokenOnly {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
	}
}

func TestTriggerEventFailClosed(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		source   string
		cardName string
	}{
		{name: "wrong intro", source: "At you cast a spell, draw a card."},
		{name: "no matching suffix", source: "Whenever a creature enters somewhere, draw a card."},
		{name: "unknown spell type", source: "Whenever you cast a permanent spell, draw a card."},
		{name: "bare noninstant spell noun", source: "Whenever you cast an artifact, draw a card."},
		{name: "unrepresentable another ability source", source: "Whenever you activate an ability of another creature that isn't a mana ability, draw a card."},
		{name: "near miss attack alone", source: "Whenever you attack alone, draw a card."},
		{name: "near miss ability target", source: "Whenever this creature becomes the target of an ability, draw a card."},
		{name: "unknown counter type", source: "Whenever a frobnicate counter is put on this creature, draw a card."},
		{name: "unknown counter type one or more passive", source: "Whenever one or more frobnicate counters are put on this creature, draw a card."},
		{name: "copy without cast", source: "Whenever you copy an instant or sorcery spell, draw a card."},
		{name: "opponent cast or copy", source: "Whenever an opponent casts or copies an instant or sorcery spell, draw a card."},
		{name: "ordinal beyond supported word", source: "Whenever you cast your sixth spell each turn, draw a card."},
		{name: "ordinal opponent actor", source: "Whenever an opponent casts your second spell each turn, draw a card."},
		{name: "unknown spell subtype noun", source: "Whenever you cast a frobnicate spell, draw a card."},
		{name: "mixed color and subtype disjunction", source: "Whenever you cast a blue or Dragon spell, draw a card."},
		{name: "subtype with trailing qualifier", source: "Whenever you cast a Goblin spell you control, draw a card."},
		{name: "singular spell damage source without article", source: "Whenever instant or sorcery spell you control deals damage to an opponent, draw a card."},
		{name: "self or non-another selection enters", source: "Whenever this creature or a creature you control enters, draw a card."},
		{name: "mana value less than other characteristic", source: "Whenever an opponent casts a noncreature spell with mana value less than this creature's toughness, draw a card."},
		{name: "mana value less than fixed with source power", source: "Whenever an opponent casts a noncreature spell with mana value 3 or less than this creature's power, draw a card."},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if trigger := parseTriggerEventFromSource(t, test.source, test.cardName); trigger != nil {
				t.Fatalf("trigger = %#v, want nil", trigger)
			}
		})
	}
}

func parseTriggerEventFromSource(t *testing.T, source, cardName string) *TriggerEventClause {
	t.Helper()
	document, diagnostics := Parse(source, Context{CardName: cardName})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d", len(document.Abilities))
	}
	if document.Abilities[0].Trigger == nil {
		t.Fatal("trigger = nil")
	}
	return document.Abilities[0].Trigger.TriggerEvent
}

func TestParseBecomesMonstrousTriggerEvent(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"When this creature becomes monstrous, goad up to X target creatures your opponents control.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	clause := document.Abilities[0].Trigger.TriggerEvent
	if clause == nil ||
		clause.Kind != TriggerEventKindBecameMonstrous ||
		clause.Subject.Kind != TriggerEventSubjectSelf {
		t.Fatalf("trigger event = %#v", clause)
	}
	target := document.Abilities[0].Sentences[0].Targets[0]
	if target.Cardinality != (TargetCardinalitySyntax{Min: 0, Max: 99, MaxEventX: true}) ||
		target.Selection.Kind != SelectionCreature ||
		target.Selection.Controller != SelectionControllerOpponent ||
		!target.Exact {
		t.Fatalf("target = %#v", target)
	}
}

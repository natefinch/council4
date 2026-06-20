package parser

import "testing"

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
	tests = append(tests, combatTriggerEventClauseTests()...)
	tests = append(tests, damageAndCounterTriggerEventClauseTests()...)
	tests = append(tests, stateAndOtherTriggerEventClauseTests()...)
	return tests
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

func combatTriggerEventClauseTests() []triggerEventClauseTest {
	return []triggerEventClauseTest{
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
				if clause.Kind != TriggerEventKindCounterAdded || clause.Counter.Kind != TriggerEventCounterPlusOnePlusOne {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "counter minus one",
			source: "Whenever a -1/-1 counter is put on this permanent, draw a card.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Counter.Kind != TriggerEventCounterMinusOneMinusOne {
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
					clause.Counter.Kind != TriggerEventCounterPlusOnePlusOne {
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
		{name: "unknown counter type", source: "Whenever a shield counter is put on this creature, draw a card."},
		{name: "copy without cast", source: "Whenever you copy an instant or sorcery spell, draw a card."},
		{name: "opponent cast or copy", source: "Whenever an opponent casts or copies an instant or sorcery spell, draw a card."},
		{name: "ordinal beyond supported word", source: "Whenever you cast your sixth spell each turn, draw a card."},
		{name: "ordinal opponent actor", source: "Whenever an opponent casts your second spell each turn, draw a card."},
		{name: "unknown spell subtype noun", source: "Whenever you cast a frobnicate spell, draw a card."},
		{name: "subtype with trailing qualifier", source: "Whenever you cast a Goblin spell you control, draw a card."},
		{name: "singular spell damage source without article", source: "Whenever instant or sorcery spell you control deals damage to an opponent, draw a card."},
		{name: "self or non-another selection enters", source: "Whenever this creature or a creature you control enters, draw a card."},
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

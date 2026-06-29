package cardgen

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerEnterTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}

	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.Pattern.Event != game.EventPermanentEnteredBattlefield {
		t.Fatalf("event = %v, want EventPermanentEnteredBattlefield", trigger.Pattern.Event)
	}
	if trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("source = %v, want TriggerSourceSelf", trigger.Pattern.Source)
	}
}

func TestLowerCombatEventTriggers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		text    string
		want    game.TriggerPattern
		wantTyp game.TriggerType
	}{
		{
			name: "attacks",
			text: "Whenever this creature attacks, draw a card.",
			want: game.TriggerPattern{
				Event:  game.EventAttackerDeclared,
				Source: game.TriggerSourceSelf,
			},
			wantTyp: game.TriggerWhenever,
		},
		{
			name: "blocks",
			text: "Whenever this creature blocks, draw a card.",
			want: game.TriggerPattern{
				Event:  game.EventBlockerDeclared,
				Source: game.TriggerSourceSelf,
			},
			wantTyp: game.TriggerWhenever,
		},
		{
			name: "becomes blocked",
			text: "Whenever this creature becomes blocked, draw a card.",
			want: game.TriggerPattern{
				Event:  game.EventAttackerBecameBlocked,
				Source: game.TriggerSourceSelf,
			},
			wantTyp: game.TriggerWhenever,
		},
		{
			name: "attacks and isn't blocked",
			text: "Whenever this creature attacks and isn't blocked, draw a card.",
			want: game.TriggerPattern{
				Event:  game.EventAttackerBecameUnblocked,
				Source: game.TriggerSourceSelf,
			},
			wantTyp: game.TriggerWhenever,
		},
		{
			name: "blocks or becomes blocked union",
			text: "Whenever this creature blocks or becomes blocked, draw a card.",
			want: game.TriggerPattern{
				Event:      game.EventBlockerDeclared,
				UnionEvent: game.EventAttackerBecameBlocked,
				Source:     game.TriggerSourceSelf,
			},
			wantTyp: game.TriggerWhenever,
		},
		{
			name: "combat damage to player",
			text: "Whenever this creature deals combat damage to a player, draw a card.",
			want: game.TriggerPattern{
				Event:               game.EventDamageDealt,
				Source:              game.TriggerSourceSelf,
				Subject:             game.TriggerSubjectDamageSource,
				DamageRecipient:     game.DamageRecipientPlayer,
				RequireCombatDamage: true,
			},
			wantTyp: game.TriggerWhenever,
		},
		{
			name: "combat damage to creature",
			text: "Whenever this creature deals combat damage to a creature, draw a card.",
			want: game.TriggerPattern{
				Event:                game.EventDamageDealt,
				Source:               game.TriggerSourceSelf,
				Subject:              game.TriggerSubjectDamageSource,
				DamageRecipient:      game.DamageRecipientPermanent,
				DamageRecipientTypes: []types.Card{types.Creature},
				RequireCombatDamage:  true,
			},
			wantTyp: game.TriggerWhenever,
		},
		{
			name: "controlled deathtouch source combat damage to player",
			text: "Whenever a creature you control with deathtouch deals combat damage to a player, draw a card.",
			want: game.TriggerPattern{
				Event:               game.EventDamageDealt,
				Controller:          game.TriggerControllerYou,
				Subject:             game.TriggerSubjectDamageSource,
				DamageRecipient:     game.DamageRecipientPlayer,
				RequireCombatDamage: true,
				DamageSourceSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					Keyword:       game.Deathtouch,
				},
			},
			wantTyp: game.TriggerWhenever,
		},
		{
			name: "attacks alone",
			text: "Whenever this creature attacks alone, draw a card.",
			want: game.TriggerPattern{
				Event:       game.EventAttackerDeclared,
				Source:      game.TriggerSourceSelf,
				AttackAlone: true,
			},
			wantTyp: game.TriggerWhenever,
		},
		{
			name: "attack with two or more creatures",
			text: "Whenever you attack with two or more creatures, draw a card.",
			want: game.TriggerPattern{
				Event:                game.EventAttackerDeclared,
				Controller:           game.TriggerControllerYou,
				OneOrMore:            true,
				AttackerCountAtLeast: 2,
			},
			wantTyp: game.TriggerWhenever,
		},
		{
			name: "battalion self and at least two other creatures attack",
			text: "Whenever this creature and at least two other creatures attack, draw a card.",
			want: game.TriggerPattern{
				Event:                game.EventAttackerDeclared,
				Source:               game.TriggerSourceSelf,
				AttackerCountAtLeast: 3,
			},
			wantTyp: game.TriggerWhenever,
		},
		{
			name: "self or another type union combat damage to player",
			text: "Whenever this creature or another Tyranid you control deals combat damage to a player, draw a card.",
			want: game.TriggerPattern{
				Event:                       game.EventDamageDealt,
				Controller:                  game.TriggerControllerYou,
				Subject:                     game.TriggerSubjectDamageSource,
				RequireCombatDamage:         true,
				DamageRecipient:             game.DamageRecipientPlayer,
				DamageSourceSelection:       game.Selection{SubtypesAny: []types.Sub{types.Sub("Tyranid")}},
				DamageSourceSelectionOrSelf: true,
			},
			wantTyp: game.TriggerWhenever,
		},
		{
			name: "self or equipped creature combat damage to player",
			text: "Whenever this creature or equipped creature deals combat damage to a player, draw a card.",
			want: game.TriggerPattern{
				Event:                       game.EventDamageDealt,
				Source:                      game.TriggerSourceAttachedPermanent,
				Subject:                     game.TriggerSubjectDamageSource,
				RequireCombatDamage:         true,
				DamageRecipient:             game.DamageRecipientPlayer,
				DamageSourceSelection:       game.Selection{RequiredTypes: []types.Card{types.Creature}},
				DamageSourceSelectionOrSelf: true,
			},
			wantTyp: game.TriggerWhenever,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Fighter",
				Layout:     "normal",
				TypeLine:   "Creature — Human Warrior",
				OracleText: tc.text,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			trigger := face.TriggeredAbilities[0].Trigger
			if trigger.Type != tc.wantTyp {
				t.Fatalf("trigger type = %v, want %v", trigger.Type, tc.wantTyp)
			}
			if !reflect.DeepEqual(trigger.Pattern, tc.want) {
				t.Fatalf("pattern = %+v, want %+v", trigger.Pattern, tc.want)
			}
		})
	}
}

func TestLowerCombatEventTriggersFailClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Whenever this creature attacks the player with the most life, draw a card.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Fighter",
				Layout:     "normal",
				TypeLine:   "Creature — Human Warrior",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatal("unsupported combat trigger unexpectedly lowered")
			}
		})
	}
}

func TestLowerSaturatedCombatPhaseAndStepTriggerPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		text  string
		check func(*testing.T, game.TriggerPattern)
	}{
		{
			name: "one or more controlled creatures attack",
			text: "Whenever one or more creatures you control attack, draw a card.",
			check: func(t *testing.T, pattern game.TriggerPattern) {
				if pattern.Event != game.EventAttackerDeclared ||
					pattern.Controller != game.TriggerControllerYou ||
					!pattern.OneOrMore ||
					!reflect.DeepEqual(pattern.SubjectSelection.RequiredTypes, []types.Card{types.Creature}) {
					t.Fatalf("pattern = %+v", pattern)
				}
			},
		},
		{
			name: "attacker exact recipient",
			text: "Whenever a creature attacks you or a planeswalker you control, draw a card.",
			check: func(t *testing.T, pattern game.TriggerPattern) {
				if pattern.AttackRecipient != game.AttackRecipientPlayer|game.AttackRecipientPlaneswalker ||
					pattern.Player != game.TriggerPlayerYou ||
					pattern.AttackRecipientSelection.Controller != game.ControllerYou {
					t.Fatalf("pattern = %+v", pattern)
				}
			},
		},
		{
			name: "player attack per recipient batch",
			text: "Whenever you attack a player, draw a card.",
			check: func(t *testing.T, pattern game.TriggerPattern) {
				if !pattern.OneOrMore || !pattern.OneOrMorePerAttackTarget ||
					pattern.Controller != game.TriggerControllerYou ||
					pattern.AttackRecipient != game.AttackRecipientPlayer {
					t.Fatalf("pattern = %+v", pattern)
				}
			},
		},
		{
			name: "block related Selection",
			text: "Whenever this creature blocks a creature with flying, draw a card.",
			check: func(t *testing.T, pattern game.TriggerPattern) {
				if pattern.Event != game.EventBlockerDeclared ||
					pattern.RelatedSubjectSelection.Keyword != game.Flying {
					t.Fatalf("pattern = %+v", pattern)
				}
			},
		},
		{
			name: "selected combat damage source batch",
			text: "Whenever one or more creatures you control deal combat damage to a player, draw a card.",
			check: func(t *testing.T, pattern game.TriggerPattern) {
				if pattern.Subject != game.TriggerSubjectDamageSource ||
					pattern.Controller != game.TriggerControllerYou ||
					!pattern.OneOrMore ||
					!pattern.RequireCombatDamage ||
					!reflect.DeepEqual(pattern.DamageSourceSelection.RequiredTypes, []types.Card{types.Creature}) {
					t.Fatalf("pattern = %+v", pattern)
				}
			},
		},
		{
			name: "noncombat damage",
			text: "Whenever this creature deals noncombat damage to an opponent, draw a card.",
			check: func(t *testing.T, pattern game.TriggerPattern) {
				if !pattern.RequireNonCombatDamage ||
					pattern.Player != game.TriggerPlayerOpponent ||
					pattern.DamageRecipient != game.DamageRecipientPlayer {
					t.Fatalf("pattern = %+v", pattern)
				}
			},
		},
		{
			name: "player or planeswalker damage recipient",
			text: "Whenever this creature deals damage to a player or planeswalker, draw a card.",
			check: func(t *testing.T, pattern game.TriggerPattern) {
				if pattern.DamageRecipient != game.DamageRecipientPlayer|game.DamageRecipientPermanent ||
					!reflect.DeepEqual(pattern.DamageRecipientSelection.RequiredTypes, []types.Card{types.Planeswalker}) {
					t.Fatalf("pattern = %+v", pattern)
				}
			},
		},
		{
			name: "ability source damage recipient",
			text: "Whenever a source deals damage to this creature, draw a card.",
			check: func(t *testing.T, pattern game.TriggerPattern) {
				if !pattern.DamageRecipientIsSource || pattern.DamageRecipient != game.DamageRecipientPermanent {
					t.Fatalf("pattern = %+v", pattern)
				}
			},
		},
		{
			name: "all end steps",
			text: "At the beginning of the end step, draw a card.",
			check: func(t *testing.T, pattern game.TriggerPattern) {
				if pattern.Event != game.EventBeginningOfStep ||
					pattern.Step != game.StepEnd ||
					pattern.Controller != game.TriggerControllerAny {
					t.Fatalf("pattern = %+v", pattern)
				}
			},
		},
		{
			name: "attached permanent controller upkeep",
			text: "At the beginning of the upkeep of enchanted creature's controller, draw a card.",
			check: func(t *testing.T, pattern game.TriggerPattern) {
				if pattern.Event != game.EventBeginningOfStep ||
					pattern.Step != game.StepUpkeep ||
					!reflect.DeepEqual(pattern.StepPlayerSourceAttachedSelection.RequiredTypes, []types.Card{types.Creature}) {
					t.Fatalf("pattern = %+v", pattern)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Fighter",
				Layout:     "normal",
				TypeLine:   "Creature — Human Warrior",
				OracleText: test.text,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
			}
			test.check(t, face.TriggeredAbilities[0].Trigger.Pattern)
		})
	}
}

func TestCombatPhaseAndStepTriggerDiagnosticsNameMissingCapability(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		text   string
		detail string
	}{
		{
			name:   "missing boundary event",
			text:   "At the beginning of your declare attackers step, draw a card.",
			detail: "runtime does not emit a beginning-of-declare attackers step event",
		},
		{
			name:   "missing combat relation",
			text:   "Whenever this creature attacks the player with the most life, draw a card.",
			detail: "requires a missing runtime capability",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Fighter",
				Layout:     "normal",
				TypeLine:   "Creature — Human Warrior",
				OracleText: test.text,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 || !strings.Contains(diagnostics[0].Detail, test.detail) {
				t.Fatalf("diagnostics = %#v, want detail containing %q", diagnostics, test.detail)
			}
		})
	}
}

func TestActionTriggerDiagnosticsNameMissingCapability(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		text   string
		detail string
	}{
		{
			name:   "event union",
			text:   "Whenever you scry or surveil, draw a card.",
			detail: "missing event-or-subject-union semantic slot",
		},
		{
			name:   "target relation",
			text:   "Whenever a spell targets a player, draw a card.",
			detail: "missing target-subject, targeting-cause, or source relation slot",
		},
		{
			name:   "missing broad event",
			text:   "Whenever you investigate, draw a card.",
			detail: "does not emit an authoritative event for this game action",
		},
		{
			name:   "nonmana activation source relation",
			text:   "Whenever you activate an ability of a card in your graveyard that isn't a mana ability, draw a card.",
			detail: "missing source, activation-cost, or ability-provenance semantic slot",
		},
		{
			name:   "nonmana activation intervening condition",
			text:   "Whenever you activate an ability, if it isn't a mana ability, draw a card.",
			detail: "non-mana exclusion in an intervening condition requires a missing semantic condition slot",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Watcher",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: test.text,
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 || !strings.Contains(diagnostics[0].Detail, test.detail) {
				t.Fatalf("diagnostics = %#v, want detail containing %q", diagnostics, test.detail)
			}
		})
	}
}

func TestUnrestrictedActivatedAbilityTriggerFailsClosed(t *testing.T) {
	t.Parallel()
	for _, event := range []string{
		"you activate an ability",
		"a player activates an ability",
		"an opponent activates an ability of a creature",
	} {
		t.Run(event, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Activation Watcher",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: "Whenever " + event + ", draw a card.",
			}, "a")
			if err != nil {
				t.Fatal(err)
			}
			const detail = "runtime ability-activated event stream omits payment-time mana abilities"
			if len(diagnostics) == 0 || !strings.Contains(diagnostics[0].Detail, detail) {
				t.Fatalf("diagnostics = %#v, want detail containing %q", diagnostics, detail)
			}
		})
	}

	if _, ok := lowerTriggerPattern(&compiler.TriggerPattern{
		Event:  compiler.TriggerEventAbilityActivated,
		Player: compiler.TriggerPlayerYou,
	}); ok {
		t.Fatal("unrestricted semantic ability-activated pattern lowered")
	}
}

func TestLowerExpandedSemanticTriggerPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		cardName string
		typeLine string
		text     string
		want     game.TriggerPattern
	}{
		{
			name:     "controlled creature attacks",
			cardName: "Test Watcher",
			typeLine: "Creature — Human",
			text:     "Whenever a creature you control attacks, draw a card.",
			want: game.TriggerPattern{
				Event:      game.EventAttackerDeclared,
				Controller: game.TriggerControllerYou,
				SubjectSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name:     "equipped creature blocks",
			cardName: "Test Equipment",
			typeLine: "Artifact — Equipment",
			text:     "Whenever equipped creature blocks, draw a card.",
			want: game.TriggerPattern{
				Event:  game.EventBlockerDeclared,
				Source: game.TriggerSourceAttachedPermanent,
				SubjectSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name:     "another controlled artifact taps",
			cardName: "Test Watcher",
			typeLine: "Artifact",
			text:     "Whenever another artifact you control becomes tapped, draw a card.",
			want: game.TriggerPattern{
				Event:       game.EventPermanentTapped,
				Controller:  game.TriggerControllerYou,
				ExcludeSelf: true,
				SubjectSelection: game.Selection{
					RequiredTypes: []types.Card{types.Artifact},
				},
			},
		},
		{
			name:     "controlled creature untaps",
			cardName: "Test Watcher",
			typeLine: "Creature — Human",
			text:     "Whenever a creature you control becomes untapped, draw a card.",
			want: game.TriggerPattern{
				Event:      game.EventPermanentUntapped,
				Controller: game.TriggerControllerYou,
				SubjectSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name:     "opponent forest taps",
			cardName: "Test Lifetap",
			typeLine: "Enchantment",
			text:     "Whenever a Forest an opponent controls becomes tapped, you gain 1 life.",
			want: game.TriggerPattern{
				Event:      game.EventPermanentTapped,
				Controller: game.TriggerControllerOpponent,
				SubjectSelection: game.Selection{
					SubtypesAny: []types.Sub{types.Forest},
				},
			},
		},
		{
			name:     "self turns face up",
			cardName: "Test Morph",
			typeLine: "Creature — Human",
			text:     "When this creature is turned face up, draw a card.",
			want: game.TriggerPattern{
				Event:  game.EventPermanentTurnedFaceUp,
				Source: game.TriggerSourceSelf,
			},
		},
		{
			name:     "self becomes spell target",
			cardName: "Test Wardless",
			typeLine: "Creature — Human",
			text:     "Whenever this creature becomes the target of a spell, draw a card.",
			want: game.TriggerPattern{
				Event:                game.EventObjectBecameTarget,
				Source:               game.TriggerSourceSelf,
				MatchStackObjectKind: true,
				StackObjectKind:      game.StackSpell,
			},
		},
		{
			name:     "controlled creature targeted by opponent cause",
			cardName: "Test Sanctuary",
			typeLine: "Enchantment",
			text:     "Whenever a creature you control becomes the target of a spell or ability an opponent controls, draw a card.",
			want: game.TriggerPattern{
				Event:           game.EventObjectBecameTarget,
				Controller:      game.TriggerControllerYou,
				CauseController: game.TriggerControllerOpponent,
				SubjectSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name:     "opponent draw step",
			cardName: "Test Watcher",
			typeLine: "Creature — Human",
			text:     "At the beginning of each opponent's draw step, draw a card.",
			want: game.TriggerPattern{
				Event:      game.EventBeginningOfStep,
				Controller: game.TriggerControllerOpponent,
				Step:       game.StepDraw,
			},
		},
		{
			name:     "any player cycles",
			cardName: "Test Cycle Watcher",
			typeLine: "Enchantment",
			text:     "Whenever a player cycles a card, draw a card.",
			want: game.TriggerPattern{
				Event: game.EventCycled,
			},
		},
		{
			name:     "controller sacrifices clue",
			cardName: "Test Mole",
			typeLine: "Creature — Mole",
			text:     "Whenever you sacrifice a Clue, you gain 3 life.",
			want: game.TriggerPattern{
				Event:  game.EventPermanentSacrificed,
				Player: game.TriggerPlayerYou,
				SubjectSelection: game.Selection{
					SubtypesAny: []types.Sub{types.Clue},
				},
			},
		},
		{
			name:     "controller scries",
			cardName: "Test Scry Watcher",
			typeLine: "Creature — Elf",
			text:     "Whenever you scry, put a +1/+1 counter on target creature.",
			want: game.TriggerPattern{
				Event:  game.EventScry,
				Player: game.TriggerPlayerYou,
			},
		},
		{
			name:     "opponent activates nonmana creature or land ability",
			cardName: "Test Armasaur",
			typeLine: "Creature — Dinosaur",
			text:     "Whenever an opponent activates an ability of a creature or land that isn't a mana ability, draw a card.",
			want: game.TriggerPattern{
				Event:              game.EventAbilityActivated,
				Player:             game.TriggerPlayerOpponent,
				ExcludeManaAbility: true,
				SubjectSelection: game.Selection{
					RequiredTypesAny: []types.Card{types.Creature, types.Land},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       test.cardName,
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.text,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
			}
			if got := face.TriggeredAbilities[0].Trigger.Pattern; !reflect.DeepEqual(got, test.want) {
				t.Fatalf("pattern = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestLowerWheneverEquippedCreatureDiesRegression(t *testing.T) {
	t.Parallel()
	for _, card := range []ScryfallCard{
		{
			Name:       "Skullclamp",
			Layout:     "normal",
			TypeLine:   "Artifact — Equipment",
			OracleText: "Equipped creature gets +1/-1.\nWhenever equipped creature dies, draw two cards.\nEquip {1}",
		},
		{
			Name:       "Sylvok Lifestaff",
			Layout:     "normal",
			TypeLine:   "Artifact — Equipment",
			OracleText: "Equipped creature gets +1/+0.\nWhenever equipped creature dies, you gain 3 life.\nEquip {1}",
		},
	} {
		t.Run(card.Name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &card)
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
			}
			trigger := face.TriggeredAbilities[0].Trigger
			if trigger.Type != game.TriggerWhenever ||
				trigger.Pattern.Event != game.EventPermanentDied ||
				trigger.Pattern.Source != game.TriggerSourceAttachedPermanent ||
				!slices.Equal(trigger.Pattern.SubjectSelection.RequiredTypes, []types.Card{types.Creature}) {
				t.Fatalf("trigger = %#v", trigger)
			}
		})
	}
}

// TestGenerateExecutableCardSourceSwordCombatDamageTrigger verifies the "Sword
// of X and Y" equipment combat-damage trigger lowers: "that player" in a
// combat-damage-to-a-player trigger binds to the damaged player (EventPlayer),
// and "you untap all lands you control" lowers as a controller-actor mass untap.
func TestGenerateExecutableCardSourceSwordCombatDamageTrigger(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Test Sword",
		Layout:   "normal",
		TypeLine: "Artifact — Equipment",
		OracleText: "Equipped creature gets +2/+2 and has protection from black and from green.\n" +
			"Whenever equipped creature deals combat damage to a player, that player discards a card and you untap all lands you control.\n" +
			"Equip {2}",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.EventDamageDealt",
		"game.DamageRecipientPlayer",
		"game.Discard{",
		"Player: game.EventPlayerReference()",
		"game.Untap{",
		"game.ControllerYou",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateExecutableCardSourceExpandedSemanticTriggerPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		text string
		want []string
	}{
		{
			name: "attack Selection",
			text: "Whenever a creature you control attacks, draw a card.",
			want: []string{
				"game.EventAttackerDeclared",
				"game.TriggerControllerYou",
				"SubjectSelection: game.Selection{",
			},
		},
		{
			name: "spell became target",
			text: "Whenever this creature becomes the target of a spell, draw a card.",
			want: []string{
				"game.EventObjectBecameTarget",
				"game.TriggerSourceSelf",
				"MatchStackObjectKind: true",
				"game.StackSpell",
			},
		},
		{
			name: "opponent draw step",
			text: "At the beginning of each opponent's draw step, draw a card.",
			want: []string{
				"game.EventBeginningOfStep",
				"game.TriggerControllerOpponent",
				"game.StepDraw",
			},
		},
		{
			name: "end of combat step",
			text: "At the beginning of the end of combat, draw a card.",
			want: []string{
				"game.EventBeginningOfStep",
				"game.StepEndOfCombat",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Watcher",
				Layout:     "normal",
				TypeLine:   "Creature — Human",
				OracleText: test.text,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range test.want {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}

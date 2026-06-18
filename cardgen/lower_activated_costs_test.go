package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerMutateAbilityAndTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mutator",
		Layout:     "normal",
		TypeLine:   "Creature — Beast",
		OracleText: "Mutate {1}{G}\nWhenever this creature mutates, draw a card.",
		Power:      new("3"),
		Toughness:  new("3"),
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want one Mutate ability", len(face.StaticAbilities))
	}
	mutateCost, ok := game.StaticBodyMutateCost(&face.StaticAbilities[0].Body)
	if !ok || !slices.Equal(mutateCost, cost.Mana{cost.O(1), cost.G}) {
		t.Fatalf("Mutate cost = %#v, want {1}{G}", mutateCost)
	}
	if len(face.TriggeredAbilities) != 1 ||
		face.TriggeredAbilities[0].Trigger.Type != game.TriggerWhenever ||
		face.TriggeredAbilities[0].Trigger.Pattern.Event != game.EventPermanentMutated ||
		face.TriggeredAbilities[0].Trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("Mutate trigger = %#v", face.TriggeredAbilities)
	}
}

func TestLowerMutateAbilityRejectsMalformedForms(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Mutate",
		"Mutate {1}{G} extra text",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Malformed Mutator",
				Layout:     "normal",
				TypeLine:   "Creature — Beast",
				OracleText: oracleText,
				Power:      new("3"),
				Toughness:  new("3"),
			})
			if len(diagnostics) == 0 {
				t.Fatal("expected malformed Mutate diagnostic")
			}
		})
	}
}

func TestLowerActivatedNonManaCosts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		check      func(*testing.T, []cost.Additional)
	}{
		{
			name:       "sacrifice source",
			oracleText: "Sacrifice this artifact: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 || costs[0].Kind != cost.AdditionalSacrificeSource {
					t.Fatalf("additional costs = %#v, want source sacrifice", costs)
				}
			},
		},
		{
			name:       "typed sacrifice after mana and tap",
			oracleText: "{2}, {T}, Sacrifice a creature: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 2 ||
					costs[0].Kind != cost.AdditionalTap ||
					costs[1].Kind != cost.AdditionalSacrifice ||
					!costs[1].MatchPermanentType ||
					costs[1].PermanentType != types.Creature {
					t.Fatalf("additional costs = %#v, want tap and creature sacrifice", costs)
				}
			},
		},
		{
			name:       "typed discard",
			oracleText: "Discard two creature cards: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalDiscard ||
					costs[0].Amount != 2 ||
					!costs[0].MatchCardType ||
					costs[0].CardType != types.Creature ||
					costs[0].Source != zone.Hand {
					t.Fatalf("additional costs = %#v, want two creature cards discarded", costs)
				}
			},
		},
		{
			name:       "pay life",
			oracleText: "Pay 2 life: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 || costs[0].Kind != cost.AdditionalPayLife || costs[0].Amount != 2 {
					t.Fatalf("additional costs = %#v, want 2 life", costs)
				}
			},
		},
		{
			name:       "exile source",
			oracleText: "Exile this artifact: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalExileSource ||
					costs[0].Source != zone.Battlefield {
					t.Fatalf("additional costs = %#v, want battlefield source exile", costs)
				}
			},
		},
		{
			name:       "exile graveyard card",
			oracleText: "Exile a card from your graveyard: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalExile ||
					costs[0].Amount != 1 ||
					costs[0].Source != zone.Graveyard ||
					costs[0].MatchCardType {
					t.Fatalf("additional costs = %#v, want one graveyard card exile", costs)
				}
			},
		},
		{
			name:       "exile typed graveyard card",
			oracleText: "Exile a creature card from your graveyard: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalExile ||
					costs[0].Amount != 1 ||
					costs[0].Source != zone.Graveyard ||
					!costs[0].MatchCardType ||
					costs[0].CardType != types.Creature {
					t.Fatalf("additional costs = %#v, want one graveyard creature card exile", costs)
				}
			},
		},
		{
			name:       "exile two graveyard cards",
			oracleText: "Exile two cards from your graveyard: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalExile ||
					costs[0].Amount != 2 ||
					costs[0].Source != zone.Graveyard {
					t.Fatalf("additional costs = %#v, want two graveyard card exiles", costs)
				}
			},
		},
		{
			name:       "untap source",
			oracleText: "{Q}: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalUntap ||
					costs[0].Text != "{Q}" {
					t.Fatalf("additional costs = %#v, want untap source", costs)
				}
			},
		},
		{
			name:       "remove source counter",
			oracleText: "Remove a +1/+1 counter from this artifact: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalRemoveCounter ||
					costs[0].Amount != 1 ||
					costs[0].CounterKind != counter.PlusOnePlusOne {
					t.Fatalf("additional costs = %#v, want source +1/+1 counter removal", costs)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: test.oracleText,
			})
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
			}
			test.check(t, face.ActivatedAbilities[0].AdditionalCosts)
		})
	}
}

func TestLowerActivatedTapPermanentsCosts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		check      func(*testing.T, cost.Additional)
	}{
		{
			name:       "tap two artifacts",
			oracleText: "Tap two untapped artifacts you control: Draw a card.",
			check: func(t *testing.T, additional cost.Additional) {
				t.Helper()
				if additional.Kind != cost.AdditionalTapPermanents ||
					additional.Amount != 2 ||
					!additional.MatchPermanentType ||
					additional.PermanentType != types.Artifact {
					t.Fatalf("additional cost = %#v, want tap two artifacts", additional)
				}
			},
		},
		{
			name:       "tap subtype permanent",
			oracleText: "Tap an untapped Merfolk you control: Draw a card.",
			check: func(t *testing.T, additional cost.Additional) {
				t.Helper()
				if additional.Kind != cost.AdditionalTapPermanents ||
					additional.Amount != 1 ||
					additional.MatchPermanentType ||
					additional.SubtypesAny[0] != types.Merfolk ||
					additional.SubtypesAny[1] != "" {
					t.Fatalf("additional cost = %#v, want tap one Merfolk", additional)
				}
			},
		},
		{
			name:       "tap elves",
			oracleText: "Tap two untapped Elves you control: Draw a card.",
			check: func(t *testing.T, additional cost.Additional) {
				t.Helper()
				if additional.Kind != cost.AdditionalTapPermanents ||
					additional.Amount != 2 ||
					additional.SubtypesAny[0] != types.Elf ||
					additional.SubtypesAny[1] != "" {
					t.Fatalf("additional cost = %#v, want tap two Elves", additional)
				}
			},
		},
		{
			name:       "tap dwarves",
			oracleText: "Tap two untapped Dwarves you control: Draw a card.",
			check: func(t *testing.T, additional cost.Additional) {
				t.Helper()
				if additional.Kind != cost.AdditionalTapPermanents ||
					additional.Amount != 2 ||
					additional.SubtypesAny[0] != types.Dwarf ||
					additional.SubtypesAny[1] != "" {
					t.Fatalf("additional cost = %#v, want tap two Dwarves", additional)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: test.oracleText,
			})
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
			}
			costs := face.ActivatedAbilities[0].AdditionalCosts
			if len(costs) != 1 {
				t.Fatalf("additional costs = %#v, want one", costs)
			}
			test.check(t, costs[0])
		})
	}
}

func TestLowerActivatedRemoveCounterCosts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wantAmount int
		wantKind   counter.Kind
	}{
		{
			name:       "plural storage counters",
			oracleText: "Remove two storage counters from this land: Draw a card.",
			wantAmount: 2,
			wantKind:   counter.Charge,
		},
		{
			name:       "number-word fuse counters",
			oracleText: "Remove five fuse counters from this enchantment: Draw a card.",
			wantAmount: 5,
			wantKind:   counter.Charge,
		},
		{
			name:       "verse counter",
			oracleText: "Remove a verse counter from this artifact: Draw a card.",
			wantAmount: 1,
			wantKind:   counter.Verse,
		},
		{
			name:       "time counters from it",
			oracleText: "Remove 3 time counters from it: Draw a card.",
			wantAmount: 3,
			wantKind:   counter.Time,
		},
		{
			name:       "oil counter",
			oracleText: "Remove an oil counter from this artifact: Draw a card.",
			wantAmount: 1,
			wantKind:   counter.Oil,
		},
		{
			name:       "blood counters",
			oracleText: "Remove two blood counters from this artifact: Draw a card.",
			wantAmount: 2,
			wantKind:   counter.Blood,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Artifact Enchantment Land",
				OracleText: test.oracleText,
			})
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
			}
			costs := face.ActivatedAbilities[0].AdditionalCosts
			if len(costs) != 1 ||
				costs[0].Kind != cost.AdditionalRemoveCounter ||
				costs[0].Amount != test.wantAmount ||
				costs[0].CounterKind != test.wantKind {
				t.Fatalf("additional costs = %#v, want amount %d kind %v", costs, test.wantAmount, test.wantKind)
			}
		})
	}
}

func TestLowerActivatedAbilityRejectsVariableRemoveCounterCosts(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Remove X storage counters from this land: Add {G}.",
		"Remove any number of storage counters from this land: Add {G}.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: oracleText,
			})
			if len(faces) != 1 || len(faces[0].ActivatedAbilities) != 0 {
				t.Fatalf("faces = %#v, want face with no partially lowered ability", faces)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported diagnostic")
			}
		})
	}
}

func TestLowerActivatedEnergyCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Engine",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Pay {E}{E}: Draw a card.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	costs := face.ActivatedAbilities[0].AdditionalCosts
	if len(costs) != 1 ||
		costs[0].Kind != cost.AdditionalEnergy ||
		costs[0].Amount != 2 {
		t.Fatalf("additional costs = %#v, want two-energy cost", costs)
	}
}

func TestLowerActivatedRevealCosts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		oracleText      string
		wantAmount      int
		wantAmountFromX bool
		wantColor       color.Color
	}{
		{
			name:       "fixed cards sharing color",
			oracleText: "{1}, {T}, Reveal two cards from your hand that share a color: Draw a card.",
			wantAmount: 2,
		},
		{
			name:            "variable blue cards",
			oracleText:      "{2}, Reveal X blue cards from your hand, Sacrifice this creature: Draw a card.",
			wantAmountFromX: true,
			wantColor:       color.Blue,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Creature",
				OracleText: test.oracleText,
			})
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
			}
			costs := face.ActivatedAbilities[0].AdditionalCosts
			var got cost.Additional
			for _, additional := range costs {
				if additional.Kind == cost.AdditionalReveal {
					got = additional
					break
				}
			}
			if got.Kind != cost.AdditionalReveal || got.Source != zone.Hand {
				t.Fatalf("additional costs = %#v, want reveal from hand", costs)
			}
			if got.Amount != test.wantAmount {
				t.Fatalf("Amount = %d, want %d", got.Amount, test.wantAmount)
			}
			if got.AmountFromX != test.wantAmountFromX {
				t.Fatalf("AmountFromX = %v, want %v", got.AmountFromX, test.wantAmountFromX)
			}
			if test.wantColor != "" {
				if !got.MatchCardColor || got.CardColor != test.wantColor {
					t.Fatalf("card color = %v/%v, want %v", got.MatchCardColor, got.CardColor, test.wantColor)
				}
			}
		})
	}
}

func TestLowerActivatedAbilityRejectsUnsupportedRevealCosts(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Reveal the player you chose: Draw a card.",
		"Reveal this card from your hand: Draw a card.",
		"Reveal a toy you own: Draw a card.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: oracleText,
			})
			if len(faces) != 1 || len(faces[0].ActivatedAbilities) != 0 {
				t.Fatalf("faces = %#v, want face with no partially lowered ability", faces)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported diagnostic")
			}
		})
	}
}

func TestLowerActivatedReturnToHandCosts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		oracleText        string
		wantAmount        int
		wantType          types.Card
		wantSubtype       types.Sub
		wantRequireTapped bool
		wantSupertype     types.Super
	}{
		{
			name:        "plural land subtype",
			oracleText:  "Return two Islands you control to their owner's hand: Draw a card.",
			wantAmount:  2,
			wantSubtype: types.Island,
		},
		{
			name:              "tapped creature",
			oracleText:        "Return a tapped creature you control to its owner's hand: Draw a card.",
			wantAmount:        1,
			wantType:          types.Creature,
			wantRequireTapped: true,
		},
		{
			name:          "snow lands",
			oracleText:    "Return three snow lands you control to their owner's hand: Draw a card.",
			wantAmount:    3,
			wantType:      types.Land,
			wantSupertype: types.Snow,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: test.oracleText,
			})
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
			}
			costs := face.ActivatedAbilities[0].AdditionalCosts
			if len(costs) != 1 || costs[0].Kind != cost.AdditionalReturnToHand || costs[0].Amount != test.wantAmount {
				t.Fatalf("additional costs = %#v, want return-to-hand amount %d", costs, test.wantAmount)
			}
			if costs[0].RequireTapped != test.wantRequireTapped {
				t.Fatalf("RequireTapped = %v, want %v", costs[0].RequireTapped, test.wantRequireTapped)
			}
			if test.wantType != "" && (!costs[0].MatchPermanentType || costs[0].PermanentType != test.wantType) {
				t.Fatalf("permanent type = %v/%v, want %v", costs[0].MatchPermanentType, costs[0].PermanentType, test.wantType)
			}
			if test.wantSubtype != "" && costs[0].SubtypesAny != (cost.SubtypeSet{test.wantSubtype}) {
				t.Fatalf("subtypes = %#v, want %v", costs[0].SubtypesAny, test.wantSubtype)
			}
			if costs[0].RequireSupertype != test.wantSupertype {
				t.Fatalf("RequireSupertype = %v, want %v", costs[0].RequireSupertype, test.wantSupertype)
			}
		})
	}
}

func TestLowerActivatedAbilityRejectsUnsupportedReturnToHandCosts(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Return target creature to its owner's hand: Draw a card.",
		"Return a creature an opponent controls to its owner's hand: Draw a card.",
		"Return a card from your graveyard to its owner's hand: Draw a card.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: oracleText,
			})
			if len(faces) != 1 || len(faces[0].ActivatedAbilities) != 0 {
				t.Fatalf("faces = %#v, want face with no partially lowered ability", faces)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported diagnostic")
			}
		})
	}
}

func TestLowerActivatedAbilityRejectsVariableTapPermanentsCost(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Engine",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Tap X untapped Soldiers you control: Draw a card.",
	})
	if len(faces) != 1 || len(faces[0].ActivatedAbilities) != 0 {
		t.Fatalf("faces = %#v, want face with no partially lowered ability", faces)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported diagnostic")
	}
}

func TestLowerActivatedAbilityRejectsAmbiguousExileCost(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Engine",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Exile a card: Draw a card.",
	})
	if len(faces) != 1 || len(faces[0].ActivatedAbilities) != 0 {
		t.Fatalf("faces = %#v, want face with no partially lowered ability", faces)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported diagnostic")
	}
}

func TestLowerActivatedAbilityRejectsCounterRemovalFromTarget(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Engine",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Remove a +1/+1 counter from target creature: Draw a card.",
	})
	if len(faces) != 1 || len(faces[0].ActivatedAbilities) != 0 {
		t.Fatalf("faces = %#v, want face with no partially lowered ability", faces)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported diagnostic")
	}
}

func TestLowerActivatedSacrificeSubtypeAndAnotherCosts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
		check      func(*testing.T, cost.Additional)
	}{
		{
			name:       "sacrifice subtype",
			typeLine:   "Creature — Goblin",
			oracleText: "Sacrifice a Goblin: Draw a card.",
			check: func(t *testing.T, additional cost.Additional) {
				t.Helper()
				if additional.Kind != cost.AdditionalSacrifice ||
					additional.Amount != 1 ||
					additional.MatchPermanentType ||
					additional.SubtypesAny[0] != types.Goblin ||
					additional.SubtypesAny[1] != "" ||
					additional.ExcludeSource {
					t.Fatalf("additional cost = %#v, want sacrifice one Goblin", additional)
				}
			},
		},
		{
			name:       "sacrifice plural subtype",
			typeLine:   "Creature — Human",
			oracleText: "Sacrifice three Treasures: Draw a card.",
			check: func(t *testing.T, additional cost.Additional) {
				t.Helper()
				if additional.Kind != cost.AdditionalSacrifice ||
					additional.Amount != 3 ||
					additional.SubtypesAny[0] != types.Treasure {
					t.Fatalf("additional cost = %#v, want sacrifice three Treasures", additional)
				}
			},
		},
		{
			name:       "sacrifice another typed permanent",
			typeLine:   "Creature — Vampire",
			oracleText: "Sacrifice another creature: Draw a card.",
			check: func(t *testing.T, additional cost.Additional) {
				t.Helper()
				if additional.Kind != cost.AdditionalSacrifice ||
					additional.Amount != 1 ||
					!additional.MatchPermanentType ||
					additional.PermanentType != types.Creature ||
					!additional.ExcludeSource {
					t.Fatalf("additional cost = %#v, want sacrifice another creature", additional)
				}
			},
		},
		{
			name:       "sacrifice this subtype is source",
			typeLine:   "Enchantment — Aura",
			oracleText: "Sacrifice this Aura: Draw a card.",
			check: func(t *testing.T, additional cost.Additional) {
				t.Helper()
				if additional.Kind != cost.AdditionalSacrificeSource {
					t.Fatalf("additional cost = %#v, want sacrifice source", additional)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Sacrificer",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Power:      new("1"),
				Toughness:  new("1"),
			})
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
			}
			costs := face.ActivatedAbilities[0].AdditionalCosts
			if len(costs) != 1 {
				t.Fatalf("additional costs = %#v, want one", costs)
			}
			test.check(t, costs[0])
		})
	}
}

func TestLowerActivatedExileSelfFromGraveyard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Recurser",
		Layout:     "normal",
		TypeLine:   "Creature — Spirit",
		OracleText: "{1}{W}, Exile this card from your graveyard: Draw a card. Activate only as a sorcery.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if ability.ZoneOfFunction != zone.Graveyard {
		t.Fatalf("zone of function = %v, want graveyard", ability.ZoneOfFunction)
	}
	if ability.Timing != game.SorceryOnly {
		t.Fatalf("timing = %v, want sorcery only", ability.Timing)
	}
	costs := ability.AdditionalCosts
	if len(costs) != 1 ||
		costs[0].Kind != cost.AdditionalExileSource ||
		costs[0].Source != zone.Graveyard {
		t.Fatalf("additional costs = %#v, want graveyard source exile", costs)
	}
}

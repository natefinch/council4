package rules

import (
	"maps"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
)

// These tests pin the #1731 consolidation: the agent-facing choice presentation
// (candidateSacrificePermanents) and the payment planner enumerate eligible
// permanents through one shared selection engine
// (payment.CandidatePermanentsForCost), so the candidate set offered to the
// player and the set the planner accepts are identical for every permanent cost
// kind and every filter dimension.

func permanentIDSet(permanents []*game.Permanent) map[id.ID]bool {
	set := map[id.ID]bool{}
	for _, permanent := range permanents {
		set[permanent.ObjectID] = true
	}
	return set
}

func TestObjectCostChoiceMatchesPlannerEnumeration(t *testing.T) {
	tests := []struct {
		name     string
		addCost  cost.Additional
		defs     map[string]*game.CardDef
		tapped   []string
		eligible []string
	}{
		{
			name: "sacrifice artifact or creature union",
			addCost: cost.Additional{
				Kind:               cost.AdditionalSacrifice,
				Amount:             1,
				MatchPermanentType: true,
				PermanentType:      types.Artifact,
				PermanentTypeAlt:   types.Creature,
			},
			defs: map[string]*game.CardDef{
				"creature": filterTestCardDef("Creature", []types.Card{types.Creature}, nil, nil, nil),
				"artifact": filterTestCardDef("Artifact", []types.Card{types.Artifact}, nil, nil, nil),
				"land":     filterTestCardDef("Land", []types.Card{types.Land}, nil, nil, nil),
			},
			eligible: []string{"creature", "artifact"},
		},
		{
			name: "sacrifice black creature",
			addCost: cost.Additional{
				Kind:               cost.AdditionalSacrifice,
				Amount:             1,
				MatchPermanentType: true,
				PermanentType:      types.Creature,
				MatchCardColor:     true,
				CardColor:          color.Black,
			},
			defs: map[string]*game.CardDef{
				"black": filterTestCardDef("Black Creature", []types.Card{types.Creature}, nil, nil, []color.Color{color.Black}),
				"white": filterTestCardDef("White Creature", []types.Card{types.Creature}, nil, nil, []color.Color{color.White}),
			},
			eligible: []string{"black"},
		},
		{
			name: "return historic permanent",
			addCost: cost.Additional{
				Kind:          cost.AdditionalReturnToHand,
				Amount:        1,
				MatchHistoric: true,
			},
			defs: map[string]*game.CardDef{
				"artifact":  filterTestCardDef("Artifact", []types.Card{types.Artifact}, nil, nil, nil),
				"legendary": filterTestCardDef("Legendary", []types.Card{types.Creature}, []types.Super{types.Legendary}, nil, nil),
				"saga":      filterTestCardDef("Saga", []types.Card{types.Enchantment}, nil, []types.Sub{types.Saga}, nil),
				"vanilla":   filterTestCardDef("Vanilla", []types.Card{types.Creature}, nil, nil, nil),
			},
			eligible: []string{"artifact", "legendary", "saga"},
		},
		{
			name: "sacrifice legendary supertype",
			addCost: cost.Additional{
				Kind:             cost.AdditionalSacrifice,
				Amount:           1,
				RequireSupertype: types.Legendary,
			},
			defs: map[string]*game.CardDef{
				"legendary":    filterTestCardDef("Legendary", []types.Card{types.Creature}, []types.Super{types.Legendary}, nil, nil),
				"nonlegendary": filterTestCardDef("Nonlegendary", []types.Card{types.Creature}, nil, nil, nil),
			},
			eligible: []string{"legendary"},
		},
		{
			name: "tap goblin or orc subtype excludes tapped",
			addCost: cost.Additional{
				Kind:        cost.AdditionalTapPermanents,
				Amount:      1,
				SubtypesAny: cost.SubtypeSet{types.Goblin, types.Orc},
			},
			defs: map[string]*game.CardDef{
				"goblin": filterTestCardDef("Goblin", []types.Card{types.Creature}, nil, []types.Sub{types.Goblin}, nil),
				"orc":    filterTestCardDef("Orc", []types.Card{types.Creature}, nil, []types.Sub{types.Orc}, nil),
				"elf":    filterTestCardDef("Elf", []types.Card{types.Creature}, nil, []types.Sub{types.Elf}, nil),
			},
			tapped:   []string{"orc"},
			eligible: []string{"goblin"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			tappedKeys := map[string]bool{}
			for _, key := range test.tapped {
				tappedKeys[key] = true
			}
			ids := map[string]id.ID{}
			for key, def := range test.defs {
				permanent := addCombatPermanent(g, game.Player1, def)
				permanent.Tapped = tappedKeys[key]
				ids[key] = permanent.ObjectID
			}

			want := map[id.ID]bool{}
			for _, key := range test.eligible {
				want[ids[key]] = true
			}

			choice := permanentIDSet(candidateSacrificePermanents(g, game.Player1, test.addCost, nil))
			planner := permanentIDSet(payment.CandidatePermanentsForCost(&rulesPaymentState{g: g}, game.Player1, test.addCost, nil))

			if !maps.Equal(choice, planner) {
				t.Fatalf("choice set %v != planner set %v", choice, planner)
			}
			if !maps.Equal(choice, want) {
				t.Fatalf("eligible set %v, want %v", choice, want)
			}
		})
	}
}

// TestObjectCostExcludeSourceDropsSource pins that the shared engine honors an
// "another"-style cost: the source permanent is excluded while other eligible
// permanents remain. ExcludeSource is unused by the current corpus, so this
// exercises the engine's capability directly through the exported enumerator.
func TestObjectCostExcludeSourceDropsSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := filterTestCardDef("Creature", []types.Card{types.Creature}, nil, nil, nil)
	source := addCombatPermanent(g, game.Player1, def)
	other := addCombatPermanent(g, game.Player1, def)

	addCost := cost.Additional{
		Kind:               cost.AdditionalSacrifice,
		Amount:             1,
		MatchPermanentType: true,
		PermanentType:      types.Creature,
		ExcludeSource:      true,
	}

	got := permanentIDSet(payment.CandidatePermanentsForCost(&rulesPaymentState{g: g}, game.Player1, addCost, source))
	want := map[id.ID]bool{other.ObjectID: true}
	if !maps.Equal(got, want) {
		t.Fatalf("exclude-source candidates %v, want %v", got, want)
	}
}

// TestObjectCostReservationExcludesObject pins that a permanent reserved by an
// earlier cost in the same payment is not offered again, so two simultaneous
// costs cannot select the same object.
func TestObjectCostReservationExcludesObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := filterTestCardDef("Creature", []types.Card{types.Creature}, nil, nil, nil)
	first := addCombatPermanent(g, game.Player1, def)
	second := addCombatPermanent(g, game.Player1, def)

	addCost := cost.Additional{
		Kind:               cost.AdditionalSacrifice,
		Amount:             1,
		MatchPermanentType: true,
		PermanentType:      types.Creature,
	}

	got := permanentIDSet(payment.CandidatePermanentsForCost(&rulesPaymentState{g: g}, game.Player1, addCost, nil, first.ObjectID))
	want := map[id.ID]bool{second.ObjectID: true}
	if !maps.Equal(got, want) {
		t.Fatalf("reservation candidates %v, want %v", got, want)
	}
}

// TestObjectCostPhasedOutExcluded pins that a phased-out permanent is offered by
// neither the choice layer nor the planner, since a phased-out permanent is
// treated as though it does not exist (CR 702.26e).
func TestObjectCostPhasedOutExcluded(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := filterTestCardDef("Creature", []types.Card{types.Creature}, nil, nil, nil)
	active := addCombatPermanent(g, game.Player1, def)
	phased := addCombatPermanent(g, game.Player1, def)
	phased.PhasedOut = true

	addCost := cost.Additional{
		Kind:               cost.AdditionalSacrifice,
		Amount:             1,
		MatchPermanentType: true,
		PermanentType:      types.Creature,
	}

	choice := permanentIDSet(candidateSacrificePermanents(g, game.Player1, addCost, nil))
	planner := permanentIDSet(payment.CandidatePermanentsForCost(&rulesPaymentState{g: g}, game.Player1, addCost, nil))
	want := map[id.ID]bool{active.ObjectID: true}

	if !maps.Equal(choice, planner) {
		t.Fatalf("choice set %v != planner set %v", choice, planner)
	}
	if !maps.Equal(choice, want) {
		t.Fatalf("phased-out candidates %v, want %v", choice, want)
	}
}

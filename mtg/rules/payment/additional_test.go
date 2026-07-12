package payment

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestAdditionalCostSourceZone(t *testing.T) {
	tests := []struct {
		name   string
		source zone.Type
		want   zone.Type
	}{
		{name: "default is graveyard", source: zone.None, want: zone.Graveyard},
		{name: "explicit graveyard", source: zone.Graveyard, want: zone.Graveyard},
		{name: "hand", source: zone.Hand, want: zone.Hand},
		{name: "library", source: zone.Library, want: zone.Library},
		{name: "exile", source: zone.Exile, want: zone.Exile},
		{name: "command", source: zone.Command, want: zone.Command},
		{name: "unknown is unchanged", source: zone.Type(99), want: zone.Type(99)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := additionalCostSourceZone(test.source); got != test.want {
				t.Fatalf("additionalCostSourceZone(%d) = %v, want %v", test.source, got, test.want)
			}
		})
	}
}

func TestSelectionForAdditionalCost(t *testing.T) {
	tests := []struct {
		name string
		cost cost.Additional
		want game.Selection
	}{
		{
			name: "single permanent type",
			cost: cost.Additional{Kind: cost.AdditionalSacrifice, MatchPermanentType: true, PermanentType: types.Creature},
			want: game.Selection{RequiredTypesAny: []types.Card{types.Creature}},
		},
		{
			name: "permanent type union",
			cost: cost.Additional{Kind: cost.AdditionalSacrifice, MatchPermanentType: true, PermanentType: types.Artifact, PermanentTypeAlt: types.Creature},
			want: game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}},
		},
		{
			name: "card type",
			cost: cost.Additional{Kind: cost.AdditionalDiscard, MatchCardType: true, CardType: types.Creature},
			want: game.Selection{RequiredTypes: []types.Card{types.Creature}},
		},
		{
			name: "permanent color",
			cost: cost.Additional{Kind: cost.AdditionalSacrifice, MatchCardColor: true, CardColor: color.Black},
			want: game.Selection{ColorsAny: []color.Color{color.Black}},
		},
		{
			name: "supertype and subtype",
			cost: cost.Additional{Kind: cost.AdditionalSacrifice, RequireSupertype: types.Legendary, SubtypesAny: cost.SubtypeSet{types.Goblin, types.Orc}},
			want: game.Selection{Supertypes: []types.Super{types.Legendary}, SubtypesAny: []types.Sub{types.Goblin, types.Orc}},
		},
		{
			name: "exclude permanent type and token",
			cost: cost.Additional{Kind: cost.AdditionalSacrifice, ExcludePermanentType: types.Land, RequireToken: true},
			want: game.Selection{ExcludedTypes: []types.Card{types.Land}, TokenOnly: true},
		},
		{
			name: "exclude subtype",
			cost: cost.Additional{Kind: cost.AdditionalReturnToHand, MatchPermanentType: true, PermanentType: types.Land, ExcludeSubtype: types.Lair},
			want: game.Selection{RequiredTypesAny: []types.Card{types.Land}, ExcludedSubtype: types.Lair},
		},
		{
			name: "tapped",
			cost: cost.Additional{Kind: cost.AdditionalTapPermanents, RequireTapped: true},
			want: game.Selection{Tapped: game.TriTrue},
		},
		{
			name: "historic disjunction",
			cost: cost.Additional{Kind: cost.AdditionalExile, MatchHistoric: true},
			want: game.Selection{AnyOf: []game.Selection{
				{RequiredTypes: []types.Card{types.Artifact}},
				{Supertypes: []types.Super{types.Legendary}},
				{SubtypesAny: []types.Sub{types.Saga}},
			}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := SelectionForAdditionalCost(test.cost)
			if !ok {
				t.Fatalf("SelectionForAdditionalCost(%s) returned ok=false", test.name)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("SelectionForAdditionalCost(%s) = %#v, want %#v", test.name, got, test.want)
			}
		})
	}
}

func TestAdditionalCostMatchesAnyPermanentSubtype(t *testing.T) {
	additional := cost.Additional{
		Kind:        cost.AdditionalSacrifice,
		SubtypesAny: cost.SubtypeSet{types.Orc, types.Goblin},
	}
	goblin := &game.Permanent{ObjectID: 1}
	state := subtypeMatchState{subtypes: map[id.ID][]types.Sub{1: {types.Goblin}}}
	if !additionalCostMatchesPermanent(state, goblin, additional) {
		t.Fatal("Goblin did not match Orc-or-Goblin sacrifice cost")
	}
	bear := &game.Permanent{ObjectID: 2}
	if additionalCostMatchesPermanent(state, bear, additional) {
		t.Fatal("non-Orc, non-Goblin permanent matched Orc-or-Goblin sacrifice cost")
	}
}

type subtypeMatchState struct {
	fakePaymentState

	subtypes map[id.ID][]types.Sub
}

func (s subtypeMatchState) PermanentMatchesSelection(permanent *game.Permanent, sel game.Selection) bool {
	if len(sel.SubtypesAny) == 0 {
		return true
	}
	return slices.ContainsFunc(sel.SubtypesAny, func(sub types.Sub) bool {
		return slices.Contains(s.subtypes[permanent.ObjectID], sub)
	})
}

type colorMatchState struct {
	fakePaymentState

	colors map[id.ID][]color.Color
}

func (s colorMatchState) PermanentMatchesSelection(permanent *game.Permanent, sel game.Selection) bool {
	if len(sel.ColorsAny) == 0 {
		return true
	}
	return slices.ContainsFunc(sel.ColorsAny, func(c color.Color) bool {
		return slices.Contains(s.colors[permanent.ObjectID], c)
	})
}

func TestAdditionalCostMatchesPermanentColor(t *testing.T) {
	additional := cost.Additional{
		Kind:           cost.AdditionalSacrifice,
		MatchCardColor: true,
		CardColor:      color.Black,
	}
	state := colorMatchState{colors: map[id.ID][]color.Color{
		1: {color.Black},
		2: {color.White},
		3: {color.Black, color.Green},
	}}
	blackCreature := &game.Permanent{ObjectID: 1}
	if !additionalCostMatchesPermanent(state, blackCreature, additional) {
		t.Fatal("black creature did not match black-creature sacrifice cost")
	}
	whiteCreature := &game.Permanent{ObjectID: 2}
	if additionalCostMatchesPermanent(state, whiteCreature, additional) {
		t.Fatal("white creature matched black-creature sacrifice cost")
	}
	multicolor := &game.Permanent{ObjectID: 3}
	if !additionalCostMatchesPermanent(state, multicolor, additional) {
		t.Fatal("black-green creature did not match black-creature sacrifice cost")
	}
}

func TestPreferredSacrificePermanentsHonorsColor(t *testing.T) {
	white := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	black := &game.Permanent{ObjectID: 2, Controller: game.Player1}
	state := colorMatchState{
		fakePaymentState: fakePaymentState{battlefield: []*game.Permanent{white, black}},
		colors: map[id.ID][]color.Color{
			1: {color.White},
			2: {color.Black},
		},
	}
	additional := cost.Additional{
		Kind:           cost.AdditionalSacrifice,
		Amount:         1,
		MatchCardColor: true,
		CardColor:      color.Black,
	}

	chosen := preferredSacrificePermanents(state, game.Player1, additional, 1, nil, nil, nil)
	if len(chosen) != 1 || chosen[0].ObjectID != black.ObjectID {
		t.Fatalf("chosen = %#v, want only the black permanent", chosen)
	}
}

func TestPreferredReturnPermanentsRejectsInvalidPreference(t *testing.T) {
	permanent := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	state := fakePaymentState{battlefield: []*game.Permanent{permanent}}
	additional := cost.Additional{Kind: cost.AdditionalReturnToHand, Amount: 1}
	prefs := &Preferences{ReturnChoices: []id.ID{999}}

	if chosen := preferredReturnPermanents(state, game.Player1, additional, 1, nil, prefs, nil); chosen != nil {
		t.Fatalf("chosen = %#v, want invalid preference rejected", chosen)
	}
}

type fakePaymentState struct {
	battlefield []*game.Permanent
	powers      map[id.ID]int
}

func (fakePaymentState) Player(playerID game.PlayerID) (*game.Player, bool) {
	return &game.Player{ID: playerID, Life: 40}, true
}

func (fakePaymentState) CanPayLife(game.PlayerID) bool { return true }

func (fakePaymentState) PayLifeForManaColor(game.PlayerID, mana.Color) bool { return false }
func (fakePaymentState) ManaProductionMultiplier(game.PlayerID) int         { return 1 }
func (fakePaymentState) ActivateAbilitiesAsThoughHaste(game.PlayerID) bool  { return false }

func (fakePaymentState) ActivePlayer() game.PlayerID { return game.Player1 }

func (fakePaymentState) OpponentLostLifeThisTurn(game.PlayerID) bool { return false }

func (fakePaymentState) OpponentGainedLifeThisTurn(game.PlayerID) bool { return false }

func (fakePaymentState) AttackingCreatureCount() int { return 0 }

func (fakePaymentState) AdditionalDynamicAmountValue(game.PlayerID, cost.AdditionalDynamicAmount) int {
	return 0
}

func (s fakePaymentState) Battlefield() []*game.Permanent { return s.battlefield }

func (fakePaymentState) EffectiveController(p *game.Permanent) game.PlayerID {
	return p.Controller
}

func (fakePaymentState) PermanentCardDef(*game.Permanent) (*game.CardDef, bool) { return nil, false }
func (fakePaymentState) IsCommanderPermanent(*game.Permanent) bool              { return false }

func (s fakePaymentState) PermanentByObjectID(objectID id.ID) (*game.Permanent, bool) {
	for _, permanent := range s.battlefield {
		if permanent.ObjectID == objectID {
			return permanent, true
		}
	}
	return nil, false
}

func (fakePaymentState) CardInstance(id.ID) (*game.CardInstance, bool) { return nil, false }
func (s fakePaymentState) PermanentPower(p *game.Permanent) int        { return s.powers[p.ObjectID] }
func (fakePaymentState) CardFace(*game.CardInstance, game.FaceIndex) *game.CardDef {
	return nil
}
func (fakePaymentState) PermanentHasType(*game.Permanent, types.Card) bool       { return false }
func (fakePaymentState) PermanentHasSupertype(*game.Permanent, types.Super) bool { return false }
func (fakePaymentState) PermanentHasSubtype(*game.Permanent, types.Sub) bool     { return false }
func (fakePaymentState) PermanentEffectiveColors(*game.Permanent) []color.Color  { return nil }
func (fakePaymentState) PermanentMatchesSelection(*game.Permanent, game.Selection) bool {
	return true
}
func (fakePaymentState) CardMatchesSelection(*game.CardDef, game.Selection) bool { return true }
func (fakePaymentState) PermanentEffectiveAbilities(*game.Permanent) []game.Ability {
	return nil
}
func (fakePaymentState) ActivationConditionSatisfied(game.PlayerID, *game.Permanent, opt.V[game.Condition]) bool {
	return true
}
func (fakePaymentState) ManaAbilityTimingAllowed(game.PlayerID, *game.Permanent, int, game.TimingRestriction) bool {
	return true
}
func (fakePaymentState) CostModifiersForSpell(game.PlayerID, *game.CardDef, id.ID, zone.Type, []game.Target) []game.CostModifier {
	return nil
}
func (fakePaymentState) SetTapped(*game.Permanent, bool)                                   {}
func (fakePaymentState) SetTappedForMana(*game.Permanent)                                  {}
func (fakePaymentState) RecordManaAbilityUse(*game.Permanent, int, game.TimingRestriction) {}
func (fakePaymentState) AddCounters(game.PlayerID, *game.Permanent, counter.Kind, int) bool {
	return true
}
func (fakePaymentState) ExertPermanent(*game.Permanent) bool                    { return true }
func (fakePaymentState) MillCards(game.PlayerID, int)                           {}
func (fakePaymentState) RemoveCounters(*game.Permanent, counter.Kind, int) bool { return false }
func (fakePaymentState) LoseLife(game.PlayerID, int)                            {}
func (fakePaymentState) SetPlayerEnergyCounters(game.PlayerID, int) bool        { return true }
func (fakePaymentState) EmitZoneChange(game.Event)                              {}
func (fakePaymentState) EmitCardReveal(game.PlayerID, id.ID, id.ID, zone.Type)  {}
func (fakePaymentState) MovePermanentToZone(*game.Permanent, zone.Type) bool    { return true }
func (fakePaymentState) SacrificePermanent(*game.Permanent) bool                { return true }
func (fakePaymentState) DiscardFromHand(game.PlayerID, id.ID) bool              { return false }
func (fakePaymentState) DiscardAtRandom(game.PlayerID, int) bool                { return false }
func (fakePaymentState) MoveCard(game.PlayerID, id.ID, zone.Type, zone.Type) bool {
	return false
}

func TestChooseSacrificePermanentsExcludesSource(t *testing.T) {
	source := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	other := &game.Permanent{ObjectID: 2, Controller: game.Player1}
	state := fakePaymentState{battlefield: []*game.Permanent{source, other}}
	additional := cost.Additional{Kind: cost.AdditionalSacrifice, Amount: 1, ExcludeSource: true}

	chosen := chooseSacrificePermanents(state, game.Player1, additional, 1, nil, source)
	if len(chosen) != 1 || chosen[0].ObjectID != other.ObjectID {
		t.Fatalf("chosen = %#v, want only the non-source permanent", chosen)
	}

	soloState := fakePaymentState{battlefield: []*game.Permanent{source}}
	if chosen := chooseSacrificePermanents(soloState, game.Player1, additional, 1, nil, source); len(chosen) != 0 {
		t.Fatalf("chosen = %#v, want no permanent when only the source is present", chosen)
	}

	plain := cost.Additional{Kind: cost.AdditionalSacrifice, Amount: 1}
	if chosen := chooseSacrificePermanents(soloState, game.Player1, plain, 1, nil, source); len(chosen) != 1 {
		t.Fatalf("chosen = %#v, want the source eligible for a plain sacrifice", chosen)
	}
}

func TestPreferredSacrificePermanentsRejectsSourcePreferenceWhenExcluded(t *testing.T) {
	source := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	other := &game.Permanent{ObjectID: 2, Controller: game.Player1}
	state := fakePaymentState{battlefield: []*game.Permanent{source, other}}
	additional := cost.Additional{Kind: cost.AdditionalSacrifice, Amount: 1, ExcludeSource: true}

	prefs := &Preferences{SacrificeChoices: []id.ID{source.ObjectID}}
	if chosen := preferredSacrificePermanents(state, game.Player1, additional, 1, nil, prefs, source); chosen != nil {
		t.Fatalf("chosen = %#v, want rejected preference choosing the excluded source", chosen)
	}

	prefs = &Preferences{SacrificeChoices: []id.ID{other.ObjectID}}
	chosen := preferredSacrificePermanents(state, game.Player1, additional, 1, nil, prefs, source)
	if len(chosen) != 1 || chosen[0].ObjectID != other.ObjectID {
		t.Fatalf("chosen = %#v, want the non-source preference honored", chosen)
	}
}

func TestChooseTapPermanentsTotalPowerSelectsThreshold(t *testing.T) {
	source := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	mid := &game.Permanent{ObjectID: 2, Controller: game.Player1}
	small := &game.Permanent{ObjectID: 3, Controller: game.Player1}
	tapped := &game.Permanent{ObjectID: 4, Controller: game.Player1, Tapped: true}
	opp := &game.Permanent{ObjectID: 5, Controller: game.Player2}
	state := fakePaymentState{
		battlefield: []*game.Permanent{source, mid, small, tapped, opp},
		powers:      map[id.ID]int{1: 5, 2: 2, 3: 1, 4: 9, 5: 9},
	}
	additional := cost.Additional{
		Kind:              cost.AdditionalTapPermanents,
		ExcludeSource:     true,
		TotalPowerAtLeast: 3,
	}
	chosen := chooseTapPermanentsTotalPower(state, game.Player1, additional, nil, source)
	got := map[id.ID]bool{}
	total := 0
	for _, p := range chosen {
		got[p.ObjectID] = true
		total += state.powers[p.ObjectID]
	}
	if total < additional.TotalPowerAtLeast {
		t.Fatalf("total power = %d, want >= %d", total, additional.TotalPowerAtLeast)
	}
	if got[source.ObjectID] {
		t.Fatal("source must be excluded by ExcludeSource")
	}
	if got[tapped.ObjectID] {
		t.Fatal("tapped permanents must be excluded")
	}
	if got[opp.ObjectID] {
		t.Fatal("opponent's permanents must be excluded")
	}
}

func TestChooseTapPermanentsTotalPowerUnreachableReturnsNil(t *testing.T) {
	source := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	c2 := &game.Permanent{ObjectID: 2, Controller: game.Player1}
	state := fakePaymentState{
		battlefield: []*game.Permanent{source, c2},
		powers:      map[id.ID]int{1: 5, 2: 2},
	}
	additional := cost.Additional{
		Kind:              cost.AdditionalTapPermanents,
		ExcludeSource:     true,
		TotalPowerAtLeast: 10,
	}
	if chosen := chooseTapPermanentsTotalPower(state, game.Player1, additional, nil, source); chosen != nil {
		t.Fatalf("expected nil when threshold unreachable, got %v", chosen)
	}
}

func removeCounterAmongTotal(removals []counterRemoval) int {
	total := 0
	for _, removal := range removals {
		total += removal.amount
	}
	return total
}

func TestPlanRemoveCounterAmongGreedySpreadsAcrossPermanents(t *testing.T) {
	first := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	first.Counters.Add(counter.PlusOnePlusOne, 1)
	second := &game.Permanent{ObjectID: 2, Controller: game.Player1}
	second.Counters.Add(counter.PlusOnePlusOne, 3)
	state := fakePaymentState{battlefield: []*game.Permanent{first, second}}
	additional := cost.Additional{Kind: cost.AdditionalRemoveCounterAmong, Amount: 2, CounterKind: counter.PlusOnePlusOne}

	removals, ok := planRemoveCounterAmong(state, game.Player1, additional, 2, nil, nil)
	if !ok || removeCounterAmongTotal(removals) != 2 {
		t.Fatalf("removals = %#v ok = %t, want total 2", removals, ok)
	}
	for _, removal := range removals {
		if removal.kind != counter.PlusOnePlusOne {
			t.Fatalf("removal kind = %v, want +1/+1", removal.kind)
		}
	}
}

func TestPlanRemoveCounterAmongHonorsPreference(t *testing.T) {
	first := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	first.Counters.Add(counter.PlusOnePlusOne, 1)
	second := &game.Permanent{ObjectID: 2, Controller: game.Player1}
	second.Counters.Add(counter.PlusOnePlusOne, 3)
	state := fakePaymentState{battlefield: []*game.Permanent{first, second}}
	additional := cost.Additional{Kind: cost.AdditionalRemoveCounterAmong, Amount: 2, CounterKind: counter.PlusOnePlusOne}
	prefs := &Preferences{RemoveCounterChoices: []id.ID{2, 2}}

	removals, ok := planRemoveCounterAmong(state, game.Player1, additional, 2, nil, prefs)
	if !ok || len(removals) != 1 || removals[0].source != second || removals[0].amount != 2 {
		t.Fatalf("removals = %#v ok = %t, want both from permanent 2", removals, ok)
	}
	if len(prefs.RemoveCounterChoices) != 0 {
		t.Fatalf("remaining choices = %#v, want consumed", prefs.RemoveCounterChoices)
	}
}

func TestPlanRemoveCounterAmongFailsWhenInsufficient(t *testing.T) {
	only := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	only.Counters.Add(counter.PlusOnePlusOne, 1)
	state := fakePaymentState{battlefield: []*game.Permanent{only}}
	additional := cost.Additional{Kind: cost.AdditionalRemoveCounterAmong, Amount: 2, CounterKind: counter.PlusOnePlusOne}

	if removals, ok := planRemoveCounterAmong(state, game.Player1, additional, 2, nil, nil); ok {
		t.Fatalf("removals = %#v ok = true, want failure for insufficient counters", removals)
	}
}

func TestPlanRemoveCounterAmongReservesPlannedCounters(t *testing.T) {
	only := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	only.Counters.Add(counter.PlusOnePlusOne, 2)
	state := fakePaymentState{battlefield: []*game.Permanent{only}}
	additional := cost.Additional{Kind: cost.AdditionalRemoveCounterAmong, Amount: 2, CounterKind: counter.PlusOnePlusOne}
	planned := []counterRemoval{{source: only, kind: counter.PlusOnePlusOne, amount: 1}}

	if removals, ok := planRemoveCounterAmong(state, game.Player1, additional, 2, planned, nil); ok {
		t.Fatalf("removals = %#v ok = true, want failure once reserved counters are excluded", removals)
	}
}

func TestPlanRemoveCounterAmongInvalidPreferenceFallsBack(t *testing.T) {
	only := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	only.Counters.Add(counter.PlusOnePlusOne, 2)
	state := fakePaymentState{battlefield: []*game.Permanent{only}}
	additional := cost.Additional{Kind: cost.AdditionalRemoveCounterAmong, Amount: 1, CounterKind: counter.PlusOnePlusOne}

	// Default policy: a stale preference falls back to a deterministic legal
	// removal rather than rejecting the payment.
	prefs := &Preferences{RemoveCounterChoices: []id.ID{999}}
	removals, ok := planRemoveCounterAmong(state, game.Player1, additional, 1, nil, prefs)
	if !ok {
		t.Fatal("planRemoveCounterAmong with stale preference = !ok, want fallback to deterministic removal")
	}
	if len(removals) != 1 || removals[0].source != only || removals[0].amount != 1 {
		t.Fatalf("removals = %#v, want one counter removed from the only matching permanent", removals)
	}

	// Strict replay: a stale preference rejects the payment instead of
	// substituting a different choice.
	strict := &Preferences{RemoveCounterChoices: []id.ID{999}, StrictReplay: true}
	if removals, ok := planRemoveCounterAmong(state, game.Player1, additional, 1, nil, strict); ok {
		t.Fatalf("removals = %#v ok = true, want strict replay to reject invalid preference", removals)
	}
}

func TestPlanRemoveCounterAmongAnyKindSpreadsAcrossKinds(t *testing.T) {
	first := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	first.Counters.Add(counter.Vigilance, 1)
	second := &game.Permanent{ObjectID: 2, Controller: game.Player1}
	second.Counters.Add(counter.Charge, 1)
	second.Counters.Add(counter.PlusOnePlusOne, 1)
	state := fakePaymentState{battlefield: []*game.Permanent{first, second}}
	additional := cost.Additional{Kind: cost.AdditionalRemoveCounterAmong, Amount: 3, AnyCounterKind: true}

	removals, ok := planRemoveCounterAmong(state, game.Player1, additional, 3, nil, nil)
	if !ok || removeCounterAmongTotal(removals) != 3 {
		t.Fatalf("removals = %#v ok = %t, want total 3 across any kinds", removals, ok)
	}
}

func TestPlanRemoveCounterAmongAnyKindHonorsPreference(t *testing.T) {
	first := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	first.Counters.Add(counter.Vigilance, 1)
	second := &game.Permanent{ObjectID: 2, Controller: game.Player1}
	second.Counters.Add(counter.Charge, 2)
	state := fakePaymentState{battlefield: []*game.Permanent{first, second}}
	additional := cost.Additional{Kind: cost.AdditionalRemoveCounterAmong, Amount: 2, AnyCounterKind: true}
	prefs := &Preferences{RemoveCounterChoices: []id.ID{2, 2}}

	removals, ok := planRemoveCounterAmong(state, game.Player1, additional, 2, nil, prefs)
	if !ok || len(removals) != 1 || removals[0].source != second || removals[0].kind != counter.Charge || removals[0].amount != 2 {
		t.Fatalf("removals = %#v ok = %t, want both charge counters from permanent 2", removals, ok)
	}
}

func TestPlanRemoveCounterAmongAnyKindFailsWhenInsufficient(t *testing.T) {
	only := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	only.Counters.Add(counter.Charge, 1)
	state := fakePaymentState{battlefield: []*game.Permanent{only}}
	additional := cost.Additional{Kind: cost.AdditionalRemoveCounterAmong, Amount: 2, AnyCounterKind: true}

	if removals, ok := planRemoveCounterAmong(state, game.Player1, additional, 2, nil, nil); ok {
		t.Fatalf("removals = %#v ok = true, want failure for insufficient counters", removals)
	}
}

func TestPlanRemoveCounterFromSourceSingleKind(t *testing.T) {
	source := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	source.Counters.Add(counter.MinusOneMinusOne, 2)

	removals, ok := planRemoveCounterFromSource(source, 1, nil)
	if !ok || len(removals) != 1 ||
		removals[0].source != source ||
		removals[0].kind != counter.MinusOneMinusOne ||
		removals[0].amount != 1 {
		t.Fatalf("removals = %#v ok = %t, want one -1/-1 counter from the source", removals, ok)
	}
}

func TestPlanRemoveCounterFromSourceSpreadsAcrossKinds(t *testing.T) {
	source := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	source.Counters.Add(counter.MinusOneMinusOne, 1)
	source.Counters.Add(counter.Charge, 1)

	removals, ok := planRemoveCounterFromSource(source, 2, nil)
	if !ok || removeCounterAmongTotal(removals) != 2 {
		t.Fatalf("removals = %#v ok = %t, want two counters across the source's kinds", removals, ok)
	}
	for _, removal := range removals {
		if removal.source != source {
			t.Fatalf("removal source = %#v, want the cost source", removal.source)
		}
	}
}

func TestPlanRemoveCounterFromSourceFailsWhenInsufficient(t *testing.T) {
	source := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	source.Counters.Add(counter.MinusOneMinusOne, 1)

	if removals, ok := planRemoveCounterFromSource(source, 2, nil); ok {
		t.Fatalf("removals = %#v ok = true, want failure for insufficient counters", removals)
	}
}

func TestPlanRemoveCounterFromSourceReservesPlannedCounters(t *testing.T) {
	source := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	source.Counters.Add(counter.MinusOneMinusOne, 1)
	planned := []counterRemoval{{source: source, kind: counter.MinusOneMinusOne, amount: 1}}

	if removals, ok := planRemoveCounterFromSource(source, 1, planned); ok {
		t.Fatalf("removals = %#v ok = true, want failure once the only counter is reserved", removals)
	}
}

func TestChooseReturnPermanentsExcludesSource(t *testing.T) {
	source := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	other := &game.Permanent{ObjectID: 2, Controller: game.Player1}
	state := fakePaymentState{battlefield: []*game.Permanent{source, other}}
	additional := cost.Additional{Kind: cost.AdditionalReturnToHand, Amount: 1, ExcludeSource: true}

	chosen := chooseReturnPermanents(state, game.Player1, additional, 1, nil, source)
	if len(chosen) != 1 || chosen[0].ObjectID != other.ObjectID {
		t.Fatalf("chosen = %#v, want only the non-source permanent", chosen)
	}

	soloState := fakePaymentState{battlefield: []*game.Permanent{source}}
	if chosen := chooseReturnPermanents(soloState, game.Player1, additional, 1, nil, source); len(chosen) != 0 {
		t.Fatalf("chosen = %#v, want no permanent when only the source is present", chosen)
	}

	plain := cost.Additional{Kind: cost.AdditionalReturnToHand, Amount: 1}
	if chosen := chooseReturnPermanents(soloState, game.Player1, plain, 1, nil, source); len(chosen) != 1 {
		t.Fatalf("chosen = %#v, want the source eligible for a plain return", chosen)
	}
}

func TestPreferredReturnPermanentsRejectsSourcePreferenceWhenExcluded(t *testing.T) {
	source := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	other := &game.Permanent{ObjectID: 2, Controller: game.Player1}
	state := fakePaymentState{battlefield: []*game.Permanent{source, other}}
	additional := cost.Additional{Kind: cost.AdditionalReturnToHand, Amount: 1, ExcludeSource: true}

	prefs := &Preferences{ReturnChoices: []id.ID{source.ObjectID}}
	if chosen := preferredReturnPermanents(state, game.Player1, additional, 1, nil, prefs, source); chosen != nil {
		t.Fatalf("chosen = %#v, want rejected preference choosing the excluded source", chosen)
	}

	prefs = &Preferences{ReturnChoices: []id.ID{other.ObjectID}}
	chosen := preferredReturnPermanents(state, game.Player1, additional, 1, nil, prefs, source)
	if len(chosen) != 1 || chosen[0].ObjectID != other.ObjectID {
		t.Fatalf("chosen = %#v, want the non-source preference honored", chosen)
	}
}

func TestCostSourcePermanentByCardID(t *testing.T) {
	source := &game.Permanent{ObjectID: 1, Controller: game.Player1, CardInstanceID: 42}
	other := &game.Permanent{ObjectID: 2, Controller: game.Player1, CardInstanceID: 43}
	state := fakePaymentState{battlefield: []*game.Permanent{source, other}}

	if got := costSourcePermanentByCardID(state, 42); got != source {
		t.Fatalf("costSourcePermanentByCardID(42) = %#v, want the source permanent", got)
	}
	if got := costSourcePermanentByCardID(state, 99); got != nil {
		t.Fatalf("costSourcePermanentByCardID(99) = %#v, want nil when absent", got)
	}
	if got := costSourcePermanentByCardID(state, 0); got != nil {
		t.Fatalf("costSourcePermanentByCardID(0) = %#v, want nil for zero card ID", got)
	}
}

package payment

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
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

func TestAdditionalCostMatchesAnyCardSubtype(t *testing.T) {
	additional := cost.Additional{
		Kind:        cost.AdditionalReveal,
		SubtypesAny: cost.SubtypeSet{types.Forest, types.Mountain},
	}
	forest := &game.CardDef{CardFace: game.CardFace{
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Forest},
	}}
	if !additionalCostMatchesCard(forest, additional) {
		t.Fatal("Forest did not match Forest-or-Mountain reveal cost")
	}
	creature := &game.CardDef{CardFace: game.CardFace{
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Elf},
	}}
	if additionalCostMatchesCard(creature, additional) {
		t.Fatal("Elf matched Forest-or-Mountain reveal cost")
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

func (s subtypeMatchState) PermanentHasSubtype(permanent *game.Permanent, sub types.Sub) bool {
	return slices.Contains(s.subtypes[permanent.ObjectID], sub)
}

func TestPreferredReturnPermanentsRejectsInvalidPreference(t *testing.T) {
	permanent := &game.Permanent{ObjectID: 1, Controller: game.Player1}
	state := fakePaymentState{battlefield: []*game.Permanent{permanent}}
	additional := cost.Additional{Kind: cost.AdditionalReturnToHand, Amount: 1}
	prefs := &Preferences{ReturnChoices: []id.ID{999}}

	if chosen := preferredReturnPermanents(state, game.Player1, additional, 1, nil, prefs); chosen != nil {
		t.Fatalf("chosen = %#v, want invalid preference rejected", chosen)
	}
}

type fakePaymentState struct {
	battlefield []*game.Permanent
}

func (fakePaymentState) Player(playerID game.PlayerID) (*game.Player, bool) {
	return &game.Player{ID: playerID, Life: 40}, true
}

func (fakePaymentState) CanPayLife(game.PlayerID) bool { return true }

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
func (fakePaymentState) CardFace(*game.CardInstance, game.FaceIndex) *game.CardDef {
	return nil
}
func (fakePaymentState) PermanentHasType(*game.Permanent, types.Card) bool       { return false }
func (fakePaymentState) PermanentHasSupertype(*game.Permanent, types.Super) bool { return false }
func (fakePaymentState) PermanentHasSubtype(*game.Permanent, types.Sub) bool     { return false }
func (fakePaymentState) PermanentEffectiveColors(*game.Permanent) []color.Color  { return nil }
func (fakePaymentState) PermanentEffectiveAbilities(*game.Permanent) []game.Ability {
	return nil
}
func (fakePaymentState) ActivationConditionSatisfied(game.PlayerID, *game.Permanent, opt.V[game.Condition]) bool {
	return true
}
func (fakePaymentState) ManaAbilityTimingAllowed(game.PlayerID, *game.Permanent, int, game.TimingRestriction) bool {
	return true
}
func (fakePaymentState) CostModifiersForSpell(game.PlayerID, *game.CardDef, id.ID, zone.Type) []game.CostModifier {
	return nil
}
func (fakePaymentState) SetTapped(*game.Permanent, bool)                                   {}
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

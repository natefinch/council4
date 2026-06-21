package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// searchByNameAgent answers search choices by selecting the option whose label
// matches the wanted card name, or fails to find when wanted is empty.
type searchByNameAgent struct {
	wanted string
}

func (*searchByNameAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *searchByNameAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if a.wanted == "" {
		return []int{}
	}
	for _, option := range request.Options {
		if option.Label == a.wanted {
			return []int{option.Index}
		}
	}
	return nil
}

func TestSearchLibraryLetsPlayerChooseAmongMatchingCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	bear := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}})
	wolf := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Wolf", Types: []types.Card{types.Creature}}})
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:  zone.Library,
			Destination: zone.Hand,
			CardType:    opt.Val(types.Creature),
		},
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "Wolf"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(wolf) || g.Players[game.Player1].Library.Contains(wolf) {
		t.Fatal("search did not move the player-chosen matching card to hand")
	}
	if !g.Players[game.Player1].Library.Contains(bear) {
		t.Fatal("search moved an unchosen matching card out of the library")
	}
}

func TestSearchLibraryToGraveyardMovesChosenCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	bear := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}})
	wolf := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Wolf", Types: []types.Card{types.Creature}}})
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:  zone.Library,
			Destination: zone.Graveyard,
			CardType:    opt.Val(types.Creature),
		},
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "Wolf"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(wolf) || g.Players[game.Player1].Library.Contains(wolf) {
		t.Fatal("search-to-graveyard did not move the chosen card to the graveyard")
	}
	if !g.Players[game.Player1].Library.Contains(bear) {
		t.Fatal("search-to-graveyard moved an unchosen matching card out of the library")
	}
}

func TestSearchLibraryAllowsLegalFailToFind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	bear := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}})
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:  zone.Library,
			Destination: zone.Hand,
			CardType:    opt.Val(types.Creature),
		},
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: ""}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Library.Contains(bear) || g.Players[game.Player1].Hand.Contains(bear) {
		t.Fatal("search did not allow the player to legally fail to find a matching card")
	}
}

func TestLinkedSearchConditionalUntap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		startingLands int
		wanted        string
		wantFound     bool
		wantTapped    bool
	}{
		{name: "below threshold", startingLands: 2, wanted: "Forest", wantFound: true, wantTapped: true},
		{name: "reaches threshold after search", startingLands: 3, wanted: "Forest", wantFound: true, wantTapped: false},
		{name: "legal fail to find publishes no link", startingLands: 4, wanted: "", wantFound: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			for range test.startingLands - 1 {
				addBasicLandPermanent(g, game.Player1, types.Island)
			}
			unrelated := addBasicLandPermanent(g, game.Player1, types.Swamp)
			unrelated.Tapped = true
			forest := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name:       "Forest",
				Supertypes: []types.Super{types.Basic},
				Types:      []types.Card{types.Land},
				Subtypes:   []types.Sub{types.Forest},
			}})
			key := game.LinkedKey("searched-land")
			addInstructionSpellToStack(g, []game.Instruction{
				{Primitive: game.Search{
					Player: game.ControllerReference(),
					Spec: game.SearchSpec{
						SourceZone:   zone.Library,
						Destination:  zone.Battlefield,
						CardType:     opt.Val(types.Land),
						Supertype:    opt.Val(types.Basic),
						EntersTapped: true,
					},
					Amount:        game.Fixed(1),
					PublishLinked: key,
				}},
				{
					Primitive: game.Untap{Object: game.LinkedObjectReference(string(key))},
					Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
							MinCount:  4,
						}),
					})}),
				},
			})
			agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: test.wanted}}

			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

			found := permanentForCard(g, forest)
			if (found != nil) != test.wantFound {
				t.Fatalf("found permanent = %#v, wantFound=%v", found, test.wantFound)
			}
			if found != nil && found.Tapped != test.wantTapped {
				t.Fatalf("found land tapped=%v, want %v", found.Tapped, test.wantTapped)
			}
			if !unrelated.Tapped {
				t.Fatal("conditional untap affected an unrelated land")
			}
		})
	}
}

func TestLinkedSearchRepeatedActivationsReplacePriorResult(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		secondChoice string
		wantSecond   bool
	}{
		{name: "success then fail to find", secondChoice: "", wantSecond: false},
		{name: "success then success", secondChoice: "Island", wantSecond: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			forest := addCardToLibrary(g, game.Player1, repeatedSearchLandDef("Forest", types.Forest))
			island := addCardToLibrary(g, game.Player1, repeatedSearchLandDef("Island", types.Island))
			key := game.LinkedKey("searched-land")
			sourceCardID := g.IDGen.Next()

			resolveLinkedSearchActivation(
				engine,
				g,
				&game.StackObject{
					ID:           g.IDGen.Next(),
					Kind:         game.StackActivatedAbility,
					SourceID:     g.IDGen.Next(),
					SourceCardID: sourceCardID,
					Controller:   game.Player1,
				},
				key,
				"Forest",
			)
			first := permanentForCard(g, forest)
			if first == nil || first.Tapped {
				t.Fatalf("first searched land = %#v, want untapped permanent", first)
			}
			first.Tapped = true

			resolveLinkedSearchActivation(
				engine,
				g,
				&game.StackObject{
					ID:           g.IDGen.Next(),
					Kind:         game.StackActivatedAbility,
					SourceID:     g.IDGen.Next(),
					SourceCardID: sourceCardID,
					Controller:   game.Player1,
				},
				key,
				test.secondChoice,
			)

			if !first.Tapped {
				t.Fatal("second activation untapped the prior activation's linked land")
			}
			second := permanentForCard(g, island)
			if (second != nil) != test.wantSecond {
				t.Fatalf("second searched land = %#v, wantSecond=%v", second, test.wantSecond)
			}
			if second != nil && second.Tapped {
				t.Fatal("second successful activation did not untap the newest linked land")
			}
		})
	}
}

func resolveLinkedSearchActivation(
	engine *Engine,
	g *game.Game,
	obj *game.StackObject,
	key game.LinkedKey,
	wanted string,
) {
	search := game.Instruction{Primitive: game.Search{
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:   zone.Library,
			Destination:  zone.Battlefield,
			CardType:     opt.Val(types.Land),
			Supertype:    opt.Val(types.Basic),
			EntersTapped: true,
		},
		Amount:        game.Fixed(1),
		PublishLinked: key,
	}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: wanted}}
	engine.resolveInstructionWithChoices(g, obj, &search, agents, &TurnLog{})
	untap := game.Instruction{Primitive: game.Untap{
		Object: game.LinkedObjectReference(string(key)),
	}}
	engine.resolveInstructionWithChoices(g, obj, &untap, agents, &TurnLog{})
}

func repeatedSearchLandDef(name string, subtype types.Sub) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       name,
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{subtype},
	}}
}

// selectAllAgent answers every search choice by selecting all offered options,
// up to the choice's maximum.
type selectAllAgent struct{}

func (selectAllAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (selectAllAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	indices := make([]int, 0, len(request.Options))
	for i := range request.Options {
		if i >= request.MaxChoices {
			break
		}
		indices = append(indices, i)
	}
	return indices
}

func TestSearchLibraryUpToTwoBasicLandsEntersBattlefieldTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	basicLand := func(name string, sub types.Sub) *game.CardDef {
		return &game.CardDef{CardFace: game.CardFace{
			Name:       name,
			Supertypes: []types.Super{types.Basic},
			Types:      []types.Card{types.Land},
			Subtypes:   []types.Sub{sub},
		}}
	}
	forest := addCardToLibrary(g, game.Player1, basicLand("Forest", types.Forest))
	island := addCardToLibrary(g, game.Player1, basicLand("Island", types.Island))
	nonbasic := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Wastes Nonbasic",
		Types: []types.Card{types.Land},
	}})
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:   zone.Library,
			Destination:  zone.Battlefield,
			CardType:     opt.Val(types.Land),
			Supertype:    opt.Val(types.Basic),
			EntersTapped: true,
		},
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: selectAllAgent{}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	for _, want := range []id.ID{forest, island} {
		permanent := permanentForCard(g, want)
		if permanent == nil {
			t.Fatalf("basic land %v did not enter the battlefield", want)
		}
		if !permanent.Tapped {
			t.Fatalf("basic land %v entered untapped, want tapped", want)
		}
	}
	if g.Players[game.Player1].Library.Contains(forest) || g.Players[game.Player1].Library.Contains(island) {
		t.Fatal("found basic lands were not removed from the library")
	}
	if !g.Players[game.Player1].Library.Contains(nonbasic) {
		t.Fatal("nonbasic land matched a basic-only search filter")
	}
}

func TestSearchLibrarySubtypeUnionMatchesAnyNamedLand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	island := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       "Island",
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{types.Island},
	}})
	swamp := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       "Swamp",
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{types.Swamp},
	}})
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:  zone.Library,
			Destination: zone.Battlefield,
			SubtypesAny: []types.Sub{types.Forest, types.Island},
		},
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "Island"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if permanentForCard(g, island) == nil {
		t.Fatal("Island matching the subtype union did not enter the battlefield")
	}
	if !g.Players[game.Player1].Library.Contains(swamp) {
		t.Fatal("Swamp outside the subtype union was incorrectly findable")
	}
}

// TestSearchLibrarySubtypeWithCardTypeRequiresBoth verifies a tutor for a
// subtype paired with a card type ("a Myr creature card") matches only cards
// that are both that type and that subtype, exercising the combined
// CardType+SubtypesAny envelope newly reachable from generated cards.
func TestSearchLibrarySubtypeWithCardTypeRequiresBoth(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	myrCreature := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Myr Servitor",
		Types:    []types.Card{types.Artifact, types.Creature},
		Subtypes: []types.Sub{types.Myr},
	}})
	myrArtifact := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Myr Relic",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Myr},
	}})
	plainCreature := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Grizzly Bears",
		Types: []types.Card{types.Creature},
	}})
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:  zone.Library,
			Destination: zone.Battlefield,
			CardType:    opt.Val(types.Creature),
			SubtypesAny: []types.Sub{types.Myr},
		},
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "Myr Servitor"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if permanentForCard(g, myrCreature) == nil {
		t.Fatal("the Myr creature matching both type and subtype did not enter the battlefield")
	}
	if !g.Players[game.Player1].Library.Contains(myrArtifact) {
		t.Fatal("a non-creature Myr matched a Myr-creature search filter")
	}
	if !g.Players[game.Player1].Library.Contains(plainCreature) {
		t.Fatal("a non-Myr creature matched a Myr-creature search filter")
	}
}

// TestSearchLibraryPermanentAndManaValueFilter verifies the SearchSpec.Permanent
// and MaxManaValue filters newly reachable from generated cards. A "Rebel
// permanent card with mana value 5 or less" tutor must match only Rebel
// permanents at or below the mana-value bound, excluding non-permanents (an
// instant), over-cost permanents, and off-subtype permanents.
func TestSearchLibraryPermanentAndManaValueFilter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	match := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Loyal Rebel",
		ManaCost: opt.Val(cost.Mana{cost.O(3)}),
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Rebel},
	}})
	pricey := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Costly Rebel",
		ManaCost: opt.Val(cost.Mana{cost.O(7)}),
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Rebel},
	}})
	nonPermanent := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Rebel Rally",
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		Types:    []types.Card{types.Instant},
		Subtypes: []types.Sub{types.Rebel},
	}})
	offSubtype := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Goblin Raider",
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Goblin},
	}})
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:   zone.Library,
			Destination:  zone.Battlefield,
			Permanent:    true,
			SubtypesAny:  []types.Sub{types.Rebel},
			MaxManaValue: opt.Val(5),
		},
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: selectAllAgent{}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if permanentForCard(g, match) == nil {
		t.Fatal("the Rebel permanent within the mana-value bound did not enter the battlefield")
	}
	if !g.Players[game.Player1].Library.Contains(pricey) {
		t.Fatal("a Rebel permanent above the mana-value bound was incorrectly findable")
	}
	if !g.Players[game.Player1].Library.Contains(nonPermanent) {
		t.Fatal("a non-permanent Rebel card matched a permanent-only search filter")
	}
	if !g.Players[game.Player1].Library.Contains(offSubtype) {
		t.Fatal("an off-subtype permanent matched a Rebel-permanent search filter")
	}
}

// TestSearchLibraryManaValueFromXBound verifies the SearchSpec.MaxManaValueFromX
// filter ("a creature card with mana value X or less" onto the battlefield):
// the bound resolves from the resolving spell's chosen X, so only creatures at
// or below X are findable and the chosen card enters the battlefield.
func TestSearchLibraryManaValueFromXBound(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cheap := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Small Bear",
		ManaCost: opt.Val(cost.Mana{cost.O(2)}),
		Types:    []types.Card{types.Creature},
	}})
	pricey := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Giant Bear",
		ManaCost: opt.Val(cost.Mana{cost.O(5)}),
		Types:    []types.Card{types.Creature},
	}})
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:        zone.Library,
			Destination:       zone.Battlefield,
			CardType:          opt.Val(types.Creature),
			MaxManaValueFromX: true,
		},
	}, nil)
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("expected the search spell on the stack")
	}
	obj.XValue = 3
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "Small Bear"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if permanentForCard(g, cheap) == nil {
		t.Fatal("a creature within the X mana-value bound did not enter the battlefield")
	}
	if !g.Players[game.Player1].Library.Contains(pricey) {
		t.Fatal("a creature above the X mana-value bound was incorrectly findable")
	}
}

func TestSearchLibraryNilAgentFindsFirstMatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Creature", Types: []types.Card{types.Creature}}})
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:  zone.Library,
			Destination: zone.Hand,
			CardType:    opt.Val(types.Creature),
		},
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(creature) || g.Players[game.Player1].Library.Contains(creature) {
		t.Fatal("nil-agent search did not deterministically find the matching card")
	}
}

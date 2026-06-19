package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
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

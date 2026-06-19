package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestPermanentEntersTappedAndWithCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{Name: "Tapped Walker",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedReplacement("Tapped Walker enters tapped."),
			game.EntersWithCountersReplacement("Tapped Walker enters with two +1/+1 counters.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2}),
		}},
	}

	cardID := addCardToHand(g, game.Player1, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)

	permanent, ok := createCardPermanent(g, card, game.Player1, zone.Hand)

	if !ok || !permanent.Tapped {
		t.Fatalf("permanent = %+v, want enters tapped", permanent)
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("+1/+1 counters = %d, want 2", got)
	}
}

// TestEntersWithCountersIfReplacementMorbid covers the Morbid-style conditional
// enters-with-counters replacement ("This creature enters with two +1/+1
// counters on it if a creature died this turn."). The counters are placed only
// when a creature died earlier this turn, which requires the entering permanent
// to be supplied as the replacement condition's source.
func TestEntersWithCountersIfReplacementMorbid(t *testing.T) {
	morbidDef := func() *game.CardDef {
		return &game.CardDef{CardFace: game.CardFace{Name: "Festerhide Boar",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement(
					"Festerhide Boar enters with two +1/+1 counters on it if a creature died this turn.",
					&game.Condition{
						EventHistory: opt.Val(game.EventHistoryCondition{
							Pattern: game.TriggerPattern{
								Event:            game.EventPermanentDied,
								SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							},
							Window: game.EventHistoryCurrentTurn,
						}),
					},
					game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2},
				),
			}},
		}
	}

	enter := func(g *game.Game) *game.Permanent {
		cardID := addCardToHand(g, game.Player1, morbidDef())
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			t.Fatal("card instance not found")
		}
		g.Players[game.Player1].Hand.Remove(cardID)
		permanent, ok := createCardPermanent(g, card, game.Player1, zone.Hand)
		if !ok {
			t.Fatal("permanent not created")
		}
		return permanent
	}

	t.Run("no creature died", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		permanent := enter(g)
		if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 0 {
			t.Fatalf("+1/+1 counters = %d, want 0 when no creature died", got)
		}
	})

	t.Run("creature died this turn", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		emitCreatureDiedEvent(g)
		permanent := enter(g)
		if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 2 {
			t.Fatalf("+1/+1 counters = %d, want 2 when a creature died this turn", got)
		}
	})
}

// TestEntersTappedWithCountersReplacement covers the combined "This land enters
// tapped with N charge counters on it." replacement (the Vivid land cycle): the
// permanent enters both tapped and with the listed counters.
func TestEntersTappedWithCountersReplacement(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{Name: "Vivid Marsh",
		Types: []types.Card{types.Land},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedWithCountersReplacement(
				"Vivid Marsh enters tapped with two charge counters on it.",
				game.CounterPlacement{Kind: counter.Charge, Amount: 2},
			),
		}},
	}
	cardID := addCardToHand(g, game.Player1, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)
	permanent, ok := createCardPermanent(g, card, game.Player1, zone.Hand)
	if !ok || !permanent.Tapped {
		t.Fatalf("permanent = %+v, want enters tapped", permanent)
	}
	if got := permanent.Counters.Get(counter.Charge); got != 2 {
		t.Fatalf("charge counters = %d, want 2", got)
	}
}

func TestEntersTappedUnlessPaidPaysLifeByDefault(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player1)
	cardID := addCardToHand(g, game.Player1, payLifeETBModalLand())
	engine := NewEngine(nil)
	log := &TurnLog{}

	if !engine.applyPlayLandFaceWithChoices(g, game.Player1, cardID, game.FaceBack, [game.NumPlayers]PlayerAgent{}, log) {
		t.Fatal("applyPlayLandFaceWithChoices() = false")
	}

	permanent := g.Battlefield[len(g.Battlefield)-1]
	if permanent.Tapped {
		t.Fatalf("permanent = %+v, want untapped after paying life", permanent)
	}
	if got := g.Players[game.Player1].Life; got != 37 {
		t.Fatalf("life = %d, want 37", got)
	}
	if len(log.Choices) != 1 {
		t.Fatalf("choices = %+v, want one ETB payment choice", log.Choices)
	}
	choice := log.Choices[0]
	if choice.Request.Kind != game.ChoiceMay || choice.Request.Prompt != "Pay 3 life?" || len(choice.Selected) != 1 || choice.Selected[0] != 1 || !choice.UsedFallback {
		t.Fatalf("choice = %+v, want fallback yes for ETB payment", choice)
	}
}

func TestEntersTappedUnlessPaidDeclinedEntersTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player1)
	cardID := addCardToHand(g, game.Player1, payLifeETBModalLand())
	engine := NewEngine(nil)
	log := &TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	}

	if !engine.applyPlayLandFaceWithChoices(g, game.Player1, cardID, game.FaceBack, agents, log) {
		t.Fatal("applyPlayLandFaceWithChoices() = false")
	}

	permanent := g.Battlefield[len(g.Battlefield)-1]
	if !permanent.Tapped {
		t.Fatalf("permanent = %+v, want tapped after declining payment", permanent)
	}
	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("life = %d, want 40", got)
	}
	if len(log.Choices) != 1 || len(log.Choices[0].Selected) != 1 || log.Choices[0].Selected[0] != 0 || log.Choices[0].UsedFallback {
		t.Fatalf("choices = %+v, want explicit no", log.Choices)
	}
}

func TestEntersTappedUnlessPaidCannotPayEntersTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].Life = 2
	setSorcerySpeedTurn(g, game.Player1)
	cardID := addCardToHand(g, game.Player1, payLifeETBModalLand())
	engine := NewEngine(nil)
	log := &TurnLog{}

	if !engine.applyPlayLandFaceWithChoices(g, game.Player1, cardID, game.FaceBack, [game.NumPlayers]PlayerAgent{}, log) {
		t.Fatal("applyPlayLandFaceWithChoices() = false")
	}

	permanent := g.Battlefield[len(g.Battlefield)-1]
	if !permanent.Tapped {
		t.Fatalf("permanent = %+v, want tapped when payment is not payable", permanent)
	}
	if got := g.Players[game.Player1].Life; got != 2 {
		t.Fatalf("life = %d, want 2", got)
	}
	if len(log.Choices) != 0 {
		t.Fatalf("choices = %+v, want no prompt for unpayable ETB payment", log.Choices)
	}
}

func TestEntersTappedUnlessRevealMatchingCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player2)
	forestID := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:     "Forest",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Forest},
	}})
	cardID := addCardToHand(g, game.Player2, revealETBLand())
	engine := NewEngine(nil)

	if !engine.applyPlayLand(g, game.Player2, cardID) {
		t.Fatal("applyPlayLand() = false")
	}

	permanent := g.Battlefield[len(g.Battlefield)-1]
	if permanent.Tapped {
		t.Fatalf("permanent = %+v, want untapped after revealing Forest", permanent)
	}
	if !g.Players[game.Player2].Hand.Contains(forestID) {
		t.Fatal("revealed Forest left its owner's hand")
	}
	if !eventRevealedCardFromZone(g, game.Player2, cardID, forestID, zone.Hand) {
		t.Fatal("revealing Forest did not emit a reveal event")
	}
}

func eventRevealedCardFromZone(g *game.Game, player game.PlayerID, sourceID, cardID id.ID, from zone.Type) bool {
	for _, event := range g.Events {
		if event.Kind == game.EventCardRevealed &&
			event.Controller == player &&
			event.Player == player &&
			event.SourceID == sourceID &&
			event.CardID == cardID &&
			event.FromZone == from {
			return true
		}
	}
	return false
}

func TestEntersTappedUnlessRevealRejectsNonmatchingCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player1)
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Elvish Mystic",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elf},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	cardID := addCardToHand(g, game.Player1, revealETBLand())
	engine := NewEngine(nil)
	log := &TurnLog{}

	if !engine.applyPlayLandFaceWithChoices(g, game.Player1, cardID, game.FaceFront, [game.NumPlayers]PlayerAgent{}, log) {
		t.Fatal("applyPlayLandFaceWithChoices() = false")
	}

	permanent := g.Battlefield[len(g.Battlefield)-1]
	if !permanent.Tapped {
		t.Fatalf("permanent = %+v, want tapped without a matching card", permanent)
	}
	if len(log.Choices) != 0 {
		t.Fatalf("choices = %+v, want no prompt when reveal cost is unpayable", log.Choices)
	}
}

func revealETBLand() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Reveal Land",
		Types: []types.Card{types.Land},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedUnlessPaidReplacement(
				"As this land enters, you may reveal a Forest or Mountain card from your hand. If you don't, this land enters tapped.",
				game.ResolutionPayment{
					Prompt: "Reveal a matching card?",
					AdditionalCosts: []cost.Additional{{
						Kind:        cost.AdditionalReveal,
						SubtypesAny: cost.SubtypeSet{types.Forest, types.Mountain},
						Source:      zone.Hand,
					}},
				},
			),
		},
	}}
}

func TestGenericReplacementChangesZoneDestination(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.CreateReplacement{
		Replacement: &game.ReplacementEffect{
			Description:   "exile instead",
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      zone.Battlefield,
			MatchToZone:   true,
			ToZone:        zone.Graveyard,
			ReplaceToZone: zone.Exile,
		},
	}, nil)

	if !movePermanentToZone(g, target, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if g.Players[game.Player1].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("replacement did not redirect away from graveyard")
	}
	if !g.Players[game.Player1].Exile.Contains(target.CardInstanceID) {
		t.Fatal("replacement did not move card to exile")
	}
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.PermanentID == target.ObjectID && event.ToZone == zone.Exile
	})
}

func TestStaticSelfZoneReplacementMovesPermanentToLibrary(t *testing.T) {
	g := game.NewGameWithRand([game.NumPlayers]game.PlayerConfig{}, rand.New(rand.NewPCG(1, 2)))
	bottomID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Library Card"}})
	target := addCombatPermanent(g, game.Player1, selfLibraryReplacementCardDef())

	if !movePermanentToZone(g, target, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if g.Players[game.Player1].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("self replacement did not redirect away from graveyard")
	}
	if !g.Players[game.Player1].Library.Contains(target.CardInstanceID) {
		t.Fatal("self replacement did not move card to library")
	}
	if top, ok := g.Players[game.Player1].Library.Top(); !ok || top != bottomID {
		t.Fatalf("library top = %v, %v; want existing card on top after deterministic shuffle", top, ok)
	}
	assertEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
		return event.CardID == target.CardInstanceID &&
			event.PermanentID == target.ObjectID
	})
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == target.CardInstanceID &&
			event.PermanentID == target.ObjectID &&
			event.FromZone == zone.Battlefield &&
			event.ToZone == zone.Library
	})
}

func TestStaticSelfZoneReplacementDoesNotApplyFaceDownPermanentAbilities(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addFaceDownPermanent(g, game.Player1, selfLibraryReplacementCardDef(), game.FaceDownMorph)

	if !movePermanentToZone(g, target, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if g.Players[game.Player1].Library.Contains(target.CardInstanceID) {
		t.Fatal("face-down permanent used its hidden self zone replacement")
	}
	if !g.Players[game.Player1].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("face-down permanent did not move to graveyard")
	}
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == target.CardInstanceID &&
			event.PermanentID == target.ObjectID &&
			event.ToZone == zone.Graveyard
	})
}

func TestStaticSelfZoneReplacementAppliesWhenDiscardedFromHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToHand(g, game.Player1, selfLibraryReplacementCardDef())

	if !discardCardFromHand(g, game.Player1, cardID) {
		t.Fatal("discardCardFromHand() = false, want true")
	}
	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("self replacement did not redirect discarded card away from graveyard")
	}
	if !g.Players[game.Player1].Library.Contains(cardID) {
		t.Fatal("self replacement did not move discarded card to library")
	}
	assertEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
		return event.CardID == cardID && event.Player == game.Player1
	})
	assertEvent(t, g.Events, game.EventCardDiscarded, func(event game.Event) bool {
		return event.CardID == cardID &&
			event.FromZone == zone.Hand &&
			event.ToZone == zone.Library
	})
}

func TestStaticSelfZoneReplacementAppliesToGenericZoneMove(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToHand(g, game.Player1, selfLibraryReplacementCardDef())

	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Hand, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones() = false, want true")
	}
	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("self replacement did not redirect generic zone move away from graveyard")
	}
	if !g.Players[game.Player1].Library.Contains(cardID) {
		t.Fatal("self replacement did not move generic zone move to library")
	}
	assertEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
		return event.CardID == cardID && event.Player == game.Player1
	})
}

func selfLibraryReplacementCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Darksteel Colossus",
		Types: []types.Card{types.Artifact, types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{{
			Text: "If Darksteel Colossus would be put into a graveyard from anywhere, reveal Darksteel Colossus and shuffle it into its owner's library instead.",
			Replacement: game.ReplacementEffect{
				MatchEvent:         game.EventZoneChanged,
				MatchToZone:        true,
				ToZone:             zone.Graveyard,
				ReplaceToZone:      zone.Library,
				ShuffleIntoLibrary: true,
				RevealSource:       true,
				Duration:           game.DurationPermanent,
			},
		}},
	}}
}

func payLifeETBModalLand() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Front Spell // Pay Life Land",

		Types: []types.Card{types.Sorcery}}, Layout: game.LayoutModalDFC,

		Back: opt.Val(game.CardFace{
			Name:  "Pay Life Land",
			Types: []types.Card{types.Land},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedUnlessPaidReplacement("As this land enters, you may pay 3 life. If you don't, it enters tapped.", game.ResolutionPayment{
					Prompt: "Pay 3 life?",
					AdditionalCosts: []cost.Additional{
						{Kind: cost.AdditionalPayLife, Amount: 3, Text: "Pay 3 life"},
					},
				}),
			},
		}),
	}
}

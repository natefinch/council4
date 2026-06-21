package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestPermanentEntersWithXCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{Name: "Walking Ballista",
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersWithCountersReplacement("Walking Ballista enters with X +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, AmountFromX: true}),
		}},
	}

	cardID := addCardToHand(g, game.Player1, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)

	permanent, ok := createCardPermanentFaceWithOptions(NewEngine(nil), g, card, game.Player1, zone.Hand, game.FaceFront, nil, permanentCreationOptions{XValue: 3}, [game.NumPlayers]PlayerAgent{}, nil)
	if !ok {
		t.Fatal("permanent not created")
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("+1/+1 counters = %d, want 3", got)
	}
}

func TestPermanentEntersWithXCountersZeroWhenNotCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{Name: "Walking Ballista",
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersWithCountersReplacement("Walking Ballista enters with X +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, AmountFromX: true}),
		}},
	}

	cardID := addCardToHand(g, game.Player1, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)

	permanent, ok := createCardPermanent(g, card, game.Player1, zone.Hand)
	if !ok {
		t.Fatal("permanent not created")
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("+1/+1 counters = %d, want 0 when entering without a cast X", got)
	}
}

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

// TestBloodthirstReplacementEntersWithCountersWhenOpponentDamaged covers the
// Bloodthirst N keyword's conditional enters-with-counters replacement: the
// creature enters with N +1/+1 counters only when an opponent was dealt damage
// earlier this turn.
func TestBloodthirstReplacementEntersWithCountersWhenOpponentDamaged(t *testing.T) {
	bloodthirstDef := func() *game.CardDef {
		return &game.CardDef{CardFace: game.CardFace{Name: "Bloodthirst Brute",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.BloodthirstReplacement("Bloodthirst 2", 2),
			}},
		}
	}

	enter := func(g *game.Game) *game.Permanent {
		cardID := addCardToHand(g, game.Player1, bloodthirstDef())
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

	damagePlayer := func(g *game.Game, victim game.PlayerID) {
		emitEvent(g, game.Event{
			Kind:            game.EventDamageDealt,
			Controller:      game.Player1,
			Player:          victim,
			Amount:          3,
			DamageRecipient: game.DamageRecipientPlayer,
		})
	}

	t.Run("no damage dealt", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		permanent := enter(g)
		if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 0 {
			t.Fatalf("+1/+1 counters = %d, want 0 when no opponent was damaged", got)
		}
	})

	t.Run("opponent damaged this turn", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		damagePlayer(g, game.Player2)
		permanent := enter(g)
		if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 2 {
			t.Fatalf("+1/+1 counters = %d, want 2 when an opponent was damaged", got)
		}
	})

	t.Run("only controller damaged this turn", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		damagePlayer(g, game.Player1)
		permanent := enter(g)
		if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 0 {
			t.Fatalf("+1/+1 counters = %d, want 0 when only the controller was damaged", got)
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

func TestEntersTappedUnlessPaidMaterializesDynamicGenericCostFromEnteringSource(t *testing.T) {
	dynamic := game.DynamicAmount{
		Kind:        game.DynamicAmountObjectCounters,
		Object:      game.SourcePermanentReference(),
		CounterKind: counter.Loyalty,
	}
	def := etbManaPaymentCard(game.ResolutionPayment{
		DynamicGenericManaCost: opt.Val(&dynamic),
	})
	def.Types = []types.Card{types.Planeswalker}
	def.Loyalty = opt.Val(3)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	log := &TurnLog{}

	permanent := enterETBManaPaymentCard(t, g, def, log)

	if permanent.Tapped {
		t.Fatal("permanent entered tapped after paying source-counter cost")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 0 {
		t.Fatalf("colorless mana = %d, want 0 after paying {3}", got)
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Prompt != "Pay {3}?" {
		t.Fatalf("choices = %+v, want materialized {3} prompt", log.Choices)
	}
	payment := def.ReplacementAbilities[0].UnlessPaid.Val
	if payment.ManaCost.Exists || !payment.DynamicGenericManaCost.Exists {
		t.Fatalf("payment template = %+v, want unchanged dynamic generic cost", payment)
	}
}

func TestEntersTappedUnlessPaidMultiplierChargesEveryCopy(t *testing.T) {
	multiplier := game.DynamicAmount{
		Kind:        game.DynamicAmountObjectCounters,
		Object:      game.SourcePermanentReference(),
		CounterKind: counter.Loyalty,
	}
	def := etbManaPaymentCard(game.ResolutionPayment{
		ManaCost:           opt.Val(cost.Mana{cost.O(1)}),
		ManaCostMultiplier: opt.Val(&multiplier),
	})
	def.Types = []types.Card{types.Planeswalker}
	def.Loyalty = opt.Val(3)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	log := &TurnLog{}

	permanent := enterETBManaPaymentCard(t, g, def, log)

	if permanent.Tapped {
		t.Fatal("permanent entered tapped after paying all multiplied copies")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 0 {
		t.Fatalf("colorless mana = %d, want 0 after paying three copies of {1}", got)
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Prompt != "Pay {3}?" {
		t.Fatalf("choices = %+v, want multiplied {3} prompt", log.Choices)
	}
	payment := def.ReplacementAbilities[0].UnlessPaid.Val
	if !payment.ManaCostMultiplier.Exists || payment.ManaCost.Val.String() != "{1}" {
		t.Fatalf("payment template = %+v, want unchanged {1} multiplier template", payment)
	}
}

func TestEntersTappedUnlessPaidMultiplierChecksMaterializedAffordability(t *testing.T) {
	multiplier := game.DynamicAmount{
		Kind:        game.DynamicAmountObjectCounters,
		Object:      game.SourcePermanentReference(),
		CounterKind: counter.Loyalty,
	}
	def := etbManaPaymentCard(game.ResolutionPayment{
		ManaCost:           opt.Val(cost.Mana{cost.O(1)}),
		ManaCostMultiplier: opt.Val(&multiplier),
	})
	def.Types = []types.Card{types.Planeswalker}
	def.Loyalty = opt.Val(3)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	log := &TurnLog{}

	permanent := enterETBManaPaymentCard(t, g, def, log)

	if !permanent.Tapped {
		t.Fatal("permanent entered untapped despite insufficient mana for multiplied cost")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 2 {
		t.Fatalf("colorless mana = %d, want 2 after declining unpayable replacement cost", got)
	}
	if len(log.Choices) != 0 {
		t.Fatalf("choices = %+v, want no prompt for materialized unpayable cost", log.Choices)
	}
}

func TestEntersTappedUnlessPaidZeroMultiplierIsExplicitZeroCost(t *testing.T) {
	multiplier := game.DynamicAmount{
		Kind:        game.DynamicAmountObjectCounters,
		Object:      game.SourcePermanentReference(),
		CounterKind: counter.Age,
	}
	def := etbManaPaymentCard(game.ResolutionPayment{
		ManaCost:           opt.Val(cost.Mana{cost.O(2), cost.U}),
		ManaCostMultiplier: opt.Val(&multiplier),
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	log := &TurnLog{}

	permanent := enterETBManaPaymentCard(t, g, def, log)

	if permanent.Tapped {
		t.Fatal("permanent entered tapped after paying explicit zero cost")
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Prompt != "Pay {0}?" {
		t.Fatalf("choices = %+v, want explicit zero-cost prompt", log.Choices)
	}
}

func TestEntersTappedUnlessPaidFixedManaCostUnchanged(t *testing.T) {
	def := etbManaPaymentCard(game.ResolutionPayment{
		Prompt:   "Pay fixed ETB cost?",
		ManaCost: opt.Val(cost.Mana{cost.O(2)}),
		XValue:   4,
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	log := &TurnLog{}

	permanent := enterETBManaPaymentCard(t, g, def, log)

	if permanent.Tapped {
		t.Fatal("permanent entered tapped after paying fixed legacy cost")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 0 {
		t.Fatalf("colorless mana = %d, want 0 after paying fixed {2}", got)
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Prompt != "Pay fixed ETB cost?" {
		t.Fatalf("choices = %+v, want unchanged fixed-cost prompt", log.Choices)
	}
	payment := def.ReplacementAbilities[0].UnlessPaid.Val
	if payment.ManaCost.Val.String() != "{2}" || payment.XValue != 4 {
		t.Fatalf("payment template = %+v, want fixed cost and X preserved", payment)
	}
}

func TestEntersTappedUnlessPaidUnsupportedDynamicContextFailsClosed(t *testing.T) {
	multiplier := game.DynamicAmount{
		Kind:   game.DynamicAmountObjectPower,
		Object: game.SourcePermanentReference(),
	}
	def := etbManaPaymentCard(game.ResolutionPayment{
		ManaCost:           opt.Val(cost.Mana{cost.O(1)}),
		ManaCostMultiplier: opt.Val(&multiplier),
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	log := &TurnLog{}

	permanent := enterETBManaPaymentCard(t, g, def, log)

	if !permanent.Tapped {
		t.Fatal("permanent entered untapped for unsupported dynamic replacement context")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 1 {
		t.Fatalf("colorless mana = %d, want 1 after fail-closed replacement", got)
	}
	if len(log.Choices) != 0 {
		t.Fatalf("choices = %+v, want no prompt for unsupported dynamic context", log.Choices)
	}
}

func etbManaPaymentCard(payment game.ResolutionPayment) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Payment Permanent",
		Types: []types.Card{types.Artifact},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedUnlessPaidReplacement(
				"As Payment Permanent enters, you may pay a cost. If you don't, it enters tapped.",
				payment,
			),
		},
	}}
}

func enterETBManaPaymentCard(t *testing.T, g *game.Game, def *game.CardDef, log *TurnLog) *game.Permanent {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)
	permanent, ok := createCardPermanentWithChoices(NewEngine(nil), g, card, game.Player1, zone.Hand, [game.NumPlayers]PlayerAgent{}, log)
	if !ok {
		t.Fatal("permanent not created")
	}
	return permanent
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

func optionalDiscardElseGraveyardArtifact() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Mox Diamond",
		Types: []types.Card{types.Artifact},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersUnlessPaidElseZoneReplacement(
				"If this artifact would enter, you may discard a land card instead. If you do, put this artifact onto the battlefield. If you don't, put it into its owner's graveyard.",
				game.ResolutionPayment{
					Prompt: "Pay the alternative cost?",
					AdditionalCosts: []cost.Additional{{
						Kind:          cost.AdditionalDiscard,
						Amount:        1,
						Source:        zone.Hand,
						MatchCardType: true,
						CardType:      types.Land,
					}},
				},
				zone.Graveyard,
			),
		}},
	}
}

func TestOptionalEntryReplacementPaysCostAndEnters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	log := &TurnLog{}

	moxID := addCardToHand(g, game.Player1, optionalDiscardElseGraveyardArtifact())
	forestID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land},
	}})
	mox, ok := g.GetCardInstance(moxID)
	if !ok {
		t.Fatal("mox card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(moxID)

	permanent, ok := createCardPermanentFaceWithChoices(engine, g, mox, game.Player1, zone.Hand, game.FaceFront, [game.NumPlayers]PlayerAgent{}, log)
	if !ok || permanent == nil {
		t.Fatalf("createCardPermanentFaceWithChoices() = %v, %v, want permanent entered", permanent, ok)
	}
	if !g.Players[game.Player1].Graveyard.Contains(forestID) {
		t.Fatal("forest was not discarded to pay the alternative cost")
	}
	if g.Players[game.Player1].Graveyard.Contains(moxID) {
		t.Fatal("mox should be on the battlefield, not in the graveyard")
	}
}

func TestOptionalEntryReplacementUnpayableGoesToGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	log := &TurnLog{}

	moxID := addCardToHand(g, game.Player1, optionalDiscardElseGraveyardArtifact())
	mox, ok := g.GetCardInstance(moxID)
	if !ok {
		t.Fatal("mox card instance not found")
	}

	permanent, ok := createCardPermanentFaceWithChoices(engine, g, mox, game.Player1, zone.Hand, game.FaceFront, [game.NumPlayers]PlayerAgent{}, log)
	if ok || permanent != nil {
		t.Fatalf("createCardPermanentFaceWithChoices() = %v, %v, want declined entry", permanent, ok)
	}
	if !g.Players[game.Player1].Graveyard.Contains(moxID) {
		t.Fatal("mox should be in the graveyard when the cost cannot be paid")
	}
	for _, p := range g.Battlefield {
		if p.CardInstanceID == moxID {
			t.Fatal("mox should not have entered the battlefield")
		}
	}
}

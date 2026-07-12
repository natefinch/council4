package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// giftCardReturnSpell mirrors Peerless Recycling's generated shape: an unpromised
// clause returns one target card from the graveyard (card spec 0, gated
// GiftNotPromised) and a promised clause returns two target cards (card spec 1,
// gated GiftPromised). The two clauses' MoveCard instructions carry compile-time
// card-reference indices 0, 1 and 2. When the gift is promised, card spec 0 is
// inactive and drops out of the compacted target list, so resolving the promised
// instructions (indices 1 and 2) must remap down onto the two surviving card
// slots — the card-domain remap under test.
func giftCardReturnSpell() *game.CardDef {
	delivery := game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Draw{Amount: game.Fixed(1), Player: game.GiftRecipientReference()},
	}}}.Ability()
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Gift Card Return",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Instant},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.GiftKeyword{Delivery: delivery}},
		}},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{
				{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowCard, TargetZone: zone.Graveyard, Gate: game.TargetGateGiftNotPromised},
				{MinTargets: 2, MaxTargets: 2, Allow: game.TargetAllowCard, TargetZone: zone.Graveyard, Gate: game.TargetGateGiftPromised},
			},
			Sequence: []game.Instruction{
				moveTargetCardToHand(0, opt.Val(game.Condition{Negate: true, GiftPromised: true})),
				moveTargetCardToHand(1, opt.Val(game.Condition{GiftPromised: true})),
				moveTargetCardToHand(2, opt.Val(game.Condition{GiftPromised: true})),
			},
		}.Ability()),
	}}
}

// kickerCardReturnSpell mirrors Blood Beckoning's generated shape: the unkicked
// clause returns one target creature card (card spec 0, gated SpellNotKicked) and
// the kicked clause returns two (card spec 1, gated SpellKicked), with the
// MoveCard instructions carrying card-reference indices 0, 1 and 2. When kicked,
// card spec 0 is inactive and drops out, so the kicked instructions (1 and 2)
// must remap down onto the two surviving card slots.
func kickerCardReturnSpell() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Kicker Card Return",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Instant},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: greenCost().Val}},
		}},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{
				{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowCard, TargetZone: zone.Graveyard, Gate: game.TargetGateSpellNotKicked},
				{MinTargets: 2, MaxTargets: 2, Allow: game.TargetAllowCard, TargetZone: zone.Graveyard, Gate: game.TargetGateSpellKicked},
			},
			Sequence: []game.Instruction{
				moveTargetCardToHand(0, opt.Val(game.Condition{Negate: true, SpellWasKicked: true})),
				moveTargetCardToHand(1, opt.Val(game.Condition{SpellWasKicked: true})),
				moveTargetCardToHand(2, opt.Val(game.Condition{SpellWasKicked: true})),
			},
		}.Ability()),
	}}
}

// kickerAnotherCardReturnSpell mirrors Urborg Repossession's generated shape: an
// always-active clause returns one target creature card (card spec 0) and a
// kicked clause returns "another target permanent card" (card spec 1, gated
// SpellKicked, DistinctFromPriorTargets). The distinctness must be enforced
// across the card domain so the kicked cast cannot pick the same graveyard card
// twice.
func kickerAnotherCardReturnSpell() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Kicker Another Card Return",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Instant},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: greenCost().Val}},
		}},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{
				{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowCard, TargetZone: zone.Graveyard},
				{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowCard, TargetZone: zone.Graveyard, DistinctFromPriorTargets: true, Gate: game.TargetGateSpellKicked},
			},
			Sequence: []game.Instruction{
				moveTargetCardToHand(0, opt.V[game.Condition]{}),
				moveTargetCardToHand(1, opt.Val(game.Condition{SpellWasKicked: true})),
			},
		}.Ability()),
	}}
}

func moveTargetCardToHand(cardIndex int, condition opt.V[game.Condition]) game.Instruction {
	instr := game.Instruction{Primitive: game.MoveCard{
		Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: cardIndex},
		FromZone:    zone.Graveyard,
		Destination: zone.Hand,
	}}
	if condition.Exists {
		instr.Condition = opt.Val(game.EffectCondition{Condition: condition})
	}
	return instr
}

func registerSpellInstance(g *game.Game, controller game.PlayerID, def *game.CardDef) (id.ID, *game.CardInstance) {
	cardID := g.IDGen.Next()
	instance := &game.CardInstance{ID: cardID, Def: def, Owner: controller}
	g.CardInstances[cardID] = instance
	return cardID, instance
}

func gyCreatureNamed(g *game.Game, player game.PlayerID, name string) id.ID {
	return addCardToGraveyard(g, player, &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Creature},
	}})
}

func cardInHand(g *game.Game, player game.PlayerID, cardID id.ID) bool {
	return g.Players[player].Hand.Contains(cardID)
}

// TestGiftCardDomainRemapReturnsAllPromisedCards proves the card-reference remap
// (HIGH 1): on the promised branch the earlier GiftNotPromised card spec is
// dropped, and the two promised MoveCard instructions (card indices 1 and 2) must
// still return both chosen graveyard cards, not one plus an out-of-range miss.
func TestGiftCardDomainRemapReturnsAllPromisedCards(t *testing.T) {
	t.Run("promised returns both chosen cards", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		spellID, spell := registerSpellInstance(g, game.Player1, giftCardReturnSpell())
		first := gyCreatureNamed(g, game.Player1, "Graveyard One")
		second := gyCreatureNamed(g, game.Player1, "Graveyard Two")

		obj := &game.StackObject{
			ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spellID, Controller: game.Player1,
			GiftPromised: true,
			Targets:      []game.Target{currentCardTarget(t, g, first), currentCardTarget(t, g, second)},
		}
		log := &TurnLog{}
		engine.resolveSpellEffects(g, obj, spell, log)

		if !cardInHand(g, game.Player1, first) || !cardInHand(g, game.Player1, second) {
			t.Fatalf("promised return moved first=%v second=%v to hand, want both", cardInHand(g, game.Player1, first), cardInHand(g, game.Player1, second))
		}
	})

	t.Run("unpromised returns its single chosen card", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		spellID, spell := registerSpellInstance(g, game.Player1, giftCardReturnSpell())
		only := gyCreatureNamed(g, game.Player1, "Graveyard Only")
		bystander := gyCreatureNamed(g, game.Player1, "Graveyard Bystander")

		obj := &game.StackObject{
			ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spellID, Controller: game.Player1,
			GiftPromised: false,
			Targets:      []game.Target{currentCardTarget(t, g, only)},
		}
		log := &TurnLog{}
		engine.resolveSpellEffects(g, obj, spell, log)

		if !cardInHand(g, game.Player1, only) {
			t.Error("unpromised return did not move the single chosen card to hand")
		}
		if cardInHand(g, game.Player1, bystander) {
			t.Error("unpromised return moved an unchosen graveyard card to hand")
		}
	})
}

// TestKickerCardDomainRemapReturnsAllKickedCards proves the same card-reference
// remap for the kicker branch (Blood Beckoning's shape): kicked returns both
// chosen cards; unkicked returns exactly the one.
func TestKickerCardDomainRemapReturnsAllKickedCards(t *testing.T) {
	t.Run("kicked returns both chosen cards", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		spellID, spell := registerSpellInstance(g, game.Player1, kickerCardReturnSpell())
		first := gyCreatureNamed(g, game.Player1, "Graveyard One")
		second := gyCreatureNamed(g, game.Player1, "Graveyard Two")

		obj := &game.StackObject{
			ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spellID, Controller: game.Player1,
			KickerPaid: true,
			Targets:    []game.Target{currentCardTarget(t, g, first), currentCardTarget(t, g, second)},
		}
		log := &TurnLog{}
		engine.resolveSpellEffects(g, obj, spell, log)

		if !cardInHand(g, game.Player1, first) || !cardInHand(g, game.Player1, second) {
			t.Fatalf("kicked return moved first=%v second=%v to hand, want both", cardInHand(g, game.Player1, first), cardInHand(g, game.Player1, second))
		}
	})

	t.Run("unkicked returns its single chosen card", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		spellID, spell := registerSpellInstance(g, game.Player1, kickerCardReturnSpell())
		only := gyCreatureNamed(g, game.Player1, "Graveyard Only")

		obj := &game.StackObject{
			ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spellID, Controller: game.Player1,
			KickerPaid: false,
			Targets:    []game.Target{currentCardTarget(t, g, only)},
		}
		log := &TurnLog{}
		engine.resolveSpellEffects(g, obj, spell, log)

		if !cardInHand(g, game.Player1, only) {
			t.Error("unkicked return did not move the single chosen card to hand")
		}
	})
}

// TestKickerAnotherCardTargetDistinct proves card-domain "another target"
// distinctness (MEDIUM 2, Urborg Repossession's shape): kicked, the second card
// target must be distinct from the first, so the engine never offers a cast that
// picks the same graveyard card twice, and resolving a distinct pair returns
// both; unkicked needs only the always-active base target.
func TestKickerAnotherCardTargetDistinct(t *testing.T) {
	t.Run("kicked never duplicates the base card target", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		spellID := addCardToHand(g, game.Player1, kickerAnotherCardReturnSpell())
		gyCreatureNamed(g, game.Player1, "Graveyard One")
		gyCreatureNamed(g, game.Player1, "Graveyard Two")
		addBasicLandPermanent(g, game.Player1, types.Forest)
		addBasicLandPermanent(g, game.Player1, types.Forest)
		g.Turn.Phase = game.PhasePrecombatMain
		g.Turn.Step = game.StepNone
		g.Turn.PriorityPlayer = game.Player1

		var kicked, unkicked int
		for _, cast := range castActionsForCard(t, engine, g, game.Player1, spellID) {
			if !cast.KickerPaid {
				unkicked++
				if len(cast.Targets) != 1 {
					t.Errorf("unkicked cast has %d targets, want 1", len(cast.Targets))
				}
				continue
			}
			kicked++
			if len(cast.Targets) != 2 {
				t.Fatalf("kicked cast has %d targets, want 2", len(cast.Targets))
			}
			if cast.Targets[0] == cast.Targets[1] {
				t.Errorf("kicked cast duplicated card target %v; 'another target' must be distinct", cast.Targets[0])
			}
		}
		if unkicked == 0 {
			t.Error("no unkicked cast action")
		}
		if kicked == 0 {
			t.Error("no kicked cast action with two distinct graveyard cards")
		}
	})

	t.Run("kicked resolution returns both distinct cards", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		spellID, spell := registerSpellInstance(g, game.Player1, kickerAnotherCardReturnSpell())
		first := gyCreatureNamed(g, game.Player1, "Graveyard One")
		second := gyCreatureNamed(g, game.Player1, "Graveyard Two")

		obj := &game.StackObject{
			ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spellID, Controller: game.Player1,
			KickerPaid: true,
			Targets:    []game.Target{currentCardTarget(t, g, first), currentCardTarget(t, g, second)},
		}
		log := &TurnLog{}
		engine.resolveSpellEffects(g, obj, spell, log)

		if !cardInHand(g, game.Player1, first) || !cardInHand(g, game.Player1, second) {
			t.Fatalf("kicked return moved first=%v second=%v, want both", cardInHand(g, game.Player1, first), cardInHand(g, game.Player1, second))
		}
	})
}

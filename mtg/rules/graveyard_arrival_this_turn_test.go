package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine gap #3828 (Selection.PutIntoGraveyardThisTurn) is
// proved with test-local shapes of three blocked cards; none is registered here.
// Faith's Reward (printing fa19fe94-0302-4d08-9821-fa0db722048a) Oracle "Return
// to the battlefield all permanent cards in your graveyard that were put there
// from the battlefield this turn."; its Scryfall rulings returned four 2012-07-01
// rulings: discarded and countered permanent cards stay, as do cards put into
// another player's graveyard. Garna, the Bloodflame (printing
// 732797d3-59e1-4fff-bcf0-22c2606a7a5f) Oracle
// "When Garna enters, return to your hand all creature cards in your graveyard
// that were put there from anywhere this turn."; one 2018-12-07 ruling fixes the
// set on resolution. Night Nurse, Healer of Heroes (printing
// 3199907b-9f24-4554-942a-8ab5a7701717) Oracle
// "choose target permanent card in your graveyard that was put there from
// anywhere this turn. Return it to your hand."; no rulings. Second Sunrise Oracle
// "Each player returns to the battlefield all artifact, creature, enchantment,
// and land cards in their graveyard that were put there from the battlefield
// this turn." CR 2026-11-06 (MagicCompRules 20261106.txt): 400.7 (a card that
// changes zones is a new object), 404.1, 608.2n, 700.4 (dies), 701.8a
// (destroy), 701.21a (sacrifice), 704.5g (lethal damage SBA).

func gyArrivalCreature(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		ManaCost:  opt.Val(cost.Mana{cost.O(1)}),
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}
}

func gyArrivalEnchantment(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		Types:    []types.Card{types.Enchantment},
	}}
}

func gyArrivalInstant(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		Types:    []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability()),
	}}
}

// faithsRewardShape is a test-local copy of Faith's Reward's text.
func faithsRewardShape() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Faith's Reward",
		ManaCost: opt.Val(cost.Mana{cost.O(3), cost.W}),
		Types:    []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
			Primitive: game.MassReturnFromGraveyard{
				Player: game.ControllerReference(),
				Selection: game.Selection{
					RequirePermanentCard:     true,
					Controller:               game.ControllerYou,
					PutIntoGraveyardThisTurn: game.GraveyardArrivalFromBattlefield,
				},
				Destination: zone.Battlefield,
			},
		}}}.Ability()),
	}}
}

// nightNurseShape is a test-local instant carrying Night Nurse's graveyard
// target and return.
func nightNurseShape() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Night Nurse return",
		ManaCost: opt.Val(cost.Mana{cost.W}),
		Types:    []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target permanent card in your graveyard that was put there from anywhere this turn",
				Allow:      game.TargetAllowCard,
				TargetZone: zone.Graveyard,
				Selection: opt.Val(game.Selection{
					RequirePermanentCard:     true,
					Controller:               game.ControllerYou,
					PutIntoGraveyardThisTurn: game.GraveyardArrivalFromAnywhere,
				}),
			}},
			Sequence: []game.Instruction{{Primitive: game.MoveCard{
				Card:        game.CardReference{Kind: game.CardReferenceTarget},
				FromZone:    zone.Graveyard,
				Destination: zone.Hand,
			}}},
		}.Ability()),
	}}
}

func gyArrivalMatches(g *game.Game, cardID id.ID, origin game.GraveyardArrivalThisTurn) bool {
	return cardMatchesSelection(g, nil, g.CardInstances[cardID], game.Selection{PutIntoGraveyardThisTurn: origin})
}

func gyArrivalNextTurn(g *game.Game) {
	g.Turn.TurnNumber++
	markCurrentTurnEventStart(g)
}

// gyArrivalDestroy puts a creature onto the battlefield and destroys it, so it
// dies this turn.
func gyArrivalDestroy(t *testing.T, g *game.Game, player game.PlayerID, def *game.CardDef) id.ID {
	t.Helper()
	permanent := addCombatPermanent(g, player, def)
	if _, ok := destroyPermanent(g, permanent.ObjectID); !ok {
		t.Fatalf("%s was not destroyed", def.Name)
	}
	return permanent.CardInstanceID
}

func gyArrivalCastCreatureAndCounter(t *testing.T, g *game.Game, engine *Engine, def *game.CardDef) id.ID {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatalf("%s cast failed", def.Name)
	}
	spell, _ := g.Stack.Peek()
	addEffectSpellToStack(g, game.Player2, game.CounterObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(spell.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatalf("countered %s did not reach its owner's graveyard", def.Name)
	}
	return cardID
}

// TestGraveyardArrivalThisTurnRoutes verifies every route that puts a card into
// a graveyard matches the "from anywhere this turn" predicate, while only the
// battlefield departures (lethal-damage SBA, destroy, sacrifice) match the
// "from the battlefield this turn" predicate.
func TestGraveyardArrivalThisTurnRoutes(t *testing.T) {
	tests := []struct {
		name            string
		fromBattlefield bool
		put             func(t *testing.T, g *game.Game, engine *Engine) id.ID
	}{
		{"lethal damage SBA", true, func(t *testing.T, g *game.Game, engine *Engine) id.ID {
			permanent := addCombatPermanent(g, game.Player1, gyArrivalCreature("Damaged Bear"))
			permanent.MarkedDamage = 2
			engine.applyStateBasedActions(g)
			return permanent.CardInstanceID
		}},
		{"destroy", true, func(t *testing.T, g *game.Game, _ *Engine) id.ID {
			return gyArrivalDestroy(t, g, game.Player1, gyArrivalCreature("Destroyed Bear"))
		}},
		{"sacrifice", true, func(t *testing.T, g *game.Game, _ *Engine) id.ID {
			permanent := addCombatPermanent(g, game.Player1, gyArrivalEnchantment("Sacrificed Charm"))
			if !sacrificePermanent(g, permanent) {
				t.Fatal("sacrifice failed")
			}
			return permanent.CardInstanceID
		}},
		{"discard", false, func(t *testing.T, g *game.Game, _ *Engine) id.ID {
			cardID := addCardToHand(g, game.Player1, gyArrivalCreature("Discarded Bear"))
			if !discardCardFromHand(g, game.Player1, cardID) {
				t.Fatal("discard failed")
			}
			return cardID
		}},
		{"mill", false, func(t *testing.T, g *game.Game, _ *Engine) id.ID {
			cardID := addCardToLibrary(g, game.Player1, gyArrivalCreature("Milled Bear"))
			if milled := millCards(g, game.Player1, 1); len(milled) != 1 || milled[0] != cardID {
				t.Fatalf("milled = %v, want [%v]", milled, cardID)
			}
			return cardID
		}},
		{"resolved instant", false, func(t *testing.T, g *game.Game, engine *Engine) id.ID {
			cardID := addCardToHand(g, game.Player1, gyArrivalInstant("Resolved Instant"))
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			setMainPhasePriority(g, game.Player1)
			if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
				t.Fatal("instant cast failed")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			return cardID
		}},
		{"countered spell", false, func(t *testing.T, g *game.Game, engine *Engine) id.ID {
			return gyArrivalCastCreatureAndCounter(t, g, engine, gyArrivalCreature("Countered Bear"))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			cardID := tt.put(t, g, engine)
			if !g.Players[game.Player1].Graveyard.Contains(cardID) {
				t.Fatal("card did not reach its owner's graveyard")
			}
			if !gyArrivalMatches(g, cardID, game.GraveyardArrivalFromAnywhere) {
				t.Fatal("this-turn graveyard arrival did not match from anywhere")
			}
			if got := gyArrivalMatches(g, cardID, game.GraveyardArrivalFromBattlefield); got != tt.fromBattlefield {
				t.Fatalf("from-battlefield match = %v, want %v", got, tt.fromBattlefield)
			}
		})
	}
}

// TestFaithsRewardShapeReturnsOnlyThisTurnBattlefieldDeaths casts Faith's
// Reward's text after a wrath: the two creatures and the enchantment destroyed
// this turn return together, while a creature that died last turn, a discarded
// permanent card, a countered creature spell, and the opponent's dead creature
// all stay where they are.
func TestFaithsRewardShapeReturnsOnlyThisTurnBattlefieldDeaths(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	older := gyArrivalDestroy(t, g, game.Player1, gyArrivalCreature("Older Bear"))
	gyArrivalNextTurn(g)

	bearA := addCombatPermanent(g, game.Player1, gyArrivalCreature("Bear A")).CardInstanceID
	bearB := addCombatPermanent(g, game.Player1, gyArrivalCreature("Bear B")).CardInstanceID
	charm := addCombatPermanent(g, game.Player1, gyArrivalEnchantment("Charm")).CardInstanceID
	theirs := addCombatPermanent(g, game.Player2, gyArrivalCreature("Opponent Bear")).CardInstanceID
	addEffectSpellToStack(g, game.Player2, game.Destroy{Group: game.BattlefieldGroup(game.Selection{
		RequiredTypesAny: []types.Card{types.Creature, types.Enchantment},
	})}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})

	discarded := addCardToHand(g, game.Player1, gyArrivalEnchantment("Discarded Charm"))
	if !discardCardFromHand(g, game.Player1, discarded) {
		t.Fatal("discard failed")
	}
	countered := gyArrivalCastCreatureAndCounter(t, g, engine, gyArrivalCreature("Countered Bear"))

	faithsReward := addCardToHand(g, game.Player1, faithsRewardShape())
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(faithsReward, nil, 0, nil)) {
		t.Fatal("Faith's Reward cast failed")
	}
	eventStart := len(g.Events)
	engine.resolveTopOfStack(g, &TurnLog{})

	returned := map[id.ID]bool{bearA: true, bearB: true, charm: true}
	for cardID := range returned {
		if !onBattlefieldByCard(g, cardID) {
			t.Fatalf("%s did not return", g.CardInstances[cardID].Def.Name)
		}
	}
	for _, cardID := range []id.ID{older, discarded, countered} {
		if !g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatalf("%s left the graveyard", g.CardInstances[cardID].Def.Name)
		}
	}
	if !g.Players[game.Player2].Graveyard.Contains(theirs) {
		t.Fatal("the opponent's creature left their graveyard")
	}
	var simultaneous id.ID
	entered := 0
	for _, event := range g.Events[eventStart:] {
		if event.Kind != game.EventZoneChanged || event.ToZone != zone.Battlefield || !returned[event.CardID] {
			continue
		}
		entered++
		if event.SimultaneousID == 0 || (simultaneous != 0 && event.SimultaneousID != simultaneous) {
			t.Fatalf("returned cards did not enter simultaneously: %+v", event)
		}
		simultaneous = event.SimultaneousID
	}
	if entered != len(returned) {
		t.Fatalf("battlefield entries = %d, want %d", entered, len(returned))
	}
}

// TestSecondSunriseShapeReturnsEachPlayersThisTurnDeaths verifies the
// all-graveyards owner-controlled mass return ("Each player returns ... in their
// graveyard that were put there from the battlefield this turn") returns each
// player's this-turn battlefield deaths under their owner's control, but not a
// milled card.
func TestSecondSunriseShapeReturnsEachPlayersThisTurnDeaths(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mine := gyArrivalDestroy(t, g, game.Player1, gyArrivalCreature("My Bear"))
	theirs := gyArrivalDestroy(t, g, game.Player2, gyArrivalCreature("Their Bear"))
	milled := addCardToLibrary(g, game.Player2, gyArrivalCreature("Milled Bear"))
	millCards(g, game.Player2, 1)

	source := addCombatPermanent(g, game.Player1, gyArrivalEnchantment("Sunrise source"))
	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, triggeredObjFor(source), &game.Instruction{Primitive: game.MassReturnFromGraveyard{
		Player: game.ControllerReference(),
		Selection: game.Selection{
			RequiredTypesAny:         []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land},
			PutIntoGraveyardThisTurn: game.GraveyardArrivalFromBattlefield,
		},
		Destination:       zone.Battlefield,
		SourceGroup:       game.AllPlayersReference(),
		ControlledByOwner: true,
	}}, agents, &TurnLog{})

	if controller, ok := battlefieldControllerByCard(g, mine); !ok || controller != game.Player1 {
		t.Fatalf("my creature: on battlefield %v under %v", ok, controller)
	}
	if controller, ok := battlefieldControllerByCard(g, theirs); !ok || controller != game.Player2 {
		t.Fatalf("their creature: on battlefield %v under %v", ok, controller)
	}
	if !g.Players[game.Player2].Graveyard.Contains(milled) {
		t.Fatal("milled creature returned though it never left the battlefield")
	}
}

// TestGraveyardArrivalFollowsTheCurrentObject verifies CR 400.7: a creature
// that dies, returns to hand, and is then discarded is a new graveyard object
// that was not put there from the battlefield, though it was put there this turn.
func TestGraveyardArrivalFollowsTheCurrentObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := gyArrivalDestroy(t, g, game.Player1, gyArrivalCreature("Recurring Bear"))
	if !gyArrivalMatches(g, cardID, game.GraveyardArrivalFromBattlefield) {
		t.Fatal("dead creature did not match from the battlefield")
	}

	source := addCombatPermanent(g, game.Player1, gyArrivalEnchantment("Return source"))
	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, triggeredObjFor(source), returnCreatureFromGraveyardInstruction(), agents, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("creature card did not return to hand")
	}
	if gyArrivalMatches(g, cardID, game.GraveyardArrivalFromAnywhere) {
		t.Fatal("a card in hand matched a graveyard-arrival predicate")
	}

	if !discardCardFromHand(g, game.Player1, cardID) {
		t.Fatal("discard failed")
	}
	if gyArrivalMatches(g, cardID, game.GraveyardArrivalFromBattlefield) {
		t.Fatal("discarded card still matched its earlier battlefield death")
	}
	if !gyArrivalMatches(g, cardID, game.GraveyardArrivalFromAnywhere) {
		t.Fatal("discarded card did not match from anywhere")
	}
}

// TestGarnaShapeReturnsCreatureCardsPutThereFromAnywhereThisTurn resolves
// Garna's enters trigger text: milled, discarded, and dead creature cards
// return to hand, while a creature already in the graveyard at turn start, a
// milled noncreature card, and the opponent's dead creature stay.
func TestGarnaShapeReturnsCreatureCardsPutThereFromAnywhereThisTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	older := gyArrivalDestroy(t, g, game.Player1, gyArrivalCreature("Older Bear"))
	gyArrivalNextTurn(g)

	milledInstant := addCardToLibrary(g, game.Player1, gyArrivalInstant("Milled Instant"))
	milled := addCardToLibrary(g, game.Player1, gyArrivalCreature("Milled Bear"))
	if got := millCards(g, game.Player1, 2); len(got) != 2 {
		t.Fatalf("milled %d cards, want 2", len(got))
	}
	discarded := addCardToHand(g, game.Player1, gyArrivalCreature("Discarded Bear"))
	if !discardCardFromHand(g, game.Player1, discarded) {
		t.Fatal("discard failed")
	}
	died := gyArrivalDestroy(t, g, game.Player1, gyArrivalCreature("Dead Bear"))
	theirs := gyArrivalDestroy(t, g, game.Player2, gyArrivalCreature("Their Bear"))

	selection := game.Selection{
		RequiredTypes:            []types.Card{types.Creature},
		Controller:               game.ControllerYou,
		PutIntoGraveyardThisTurn: game.GraveyardArrivalFromAnywhere,
	}
	you := game.ControllerReference()
	garna := addCombatPermanent(g, game.Player1, gyArrivalCreature("Garna"))
	obj := triggeredObjFor(garna)
	if got := dynamicAmountValue(g, obj, game.Player1, game.DynamicAmount{
		Kind: game.DynamicAmountCountCardsInZone, Player: &you, CardZone: zone.Graveyard, Selection: &selection,
	}); got != 3 {
		t.Fatalf("counted %d matching graveyard cards, want 3", got)
	}

	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.MassReturnFromGraveyard{
		Player:      you,
		Selection:   selection,
		Destination: zone.Hand,
	}}, agents, &TurnLog{})

	for _, cardID := range []id.ID{milled, discarded, died} {
		if !g.Players[game.Player1].Hand.Contains(cardID) {
			t.Fatalf("%s did not return to hand", g.CardInstances[cardID].Def.Name)
		}
	}
	for _, cardID := range []id.ID{older, milledInstant} {
		if !g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatalf("%s left the graveyard", g.CardInstances[cardID].Def.Name)
		}
	}
	if !g.Players[game.Player2].Graveyard.Contains(theirs) {
		t.Fatal("the opponent's creature left their graveyard")
	}
}

// TestNightNurseShapeTargetsOnlyThisTurnGraveyardArrivals verifies the
// graveyard TargetSpec: a permanent card already in the graveyard before this
// turn is not a legal target, while one milled this turn is and returns to hand.
func TestNightNurseShapeTargetsOnlyThisTurnGraveyardArrivals(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	older := gyArrivalDestroy(t, g, game.Player1, gyArrivalCreature("Older Bear"))
	gyArrivalNextTurn(g)
	milled := addCardToLibrary(g, game.Player1, gyArrivalCreature("Milled Bear"))
	millCards(g, game.Player1, 1)

	nurse := addCardToHand(g, game.Player1, nightNurseShape())
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	setMainPhasePriority(g, game.Player1)
	if engine.applyAction(g, game.Player1, action.CastSpell(nurse, []game.Target{game.CardTarget(older)}, 0, nil)) {
		t.Fatal("targeted a card put into the graveyard on an earlier turn")
	}
	if !engine.applyAction(g, game.Player1, action.CastSpell(nurse, []game.Target{game.CardTarget(milled)}, 0, nil)) {
		t.Fatal("could not target a card milled this turn")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(milled) {
		t.Fatal("targeted card did not return to hand")
	}
	if !g.Players[game.Player1].Graveyard.Contains(older) {
		t.Fatal("older card left the graveyard")
	}
}

// TestGraveyardArrivalEndsAtTurnRolloverAndSurvivesClone verifies a cloned
// Game evaluates the predicate identically and that nothing matches on the
// following turn. A battlefield subject fails closed.
func TestGraveyardArrivalEndsAtTurnRolloverAndSurvivesClone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	died := gyArrivalDestroy(t, g, game.Player1, gyArrivalCreature("Dead Bear"))
	discarded := addCardToHand(g, game.Player1, gyArrivalCreature("Discarded Bear"))
	discardCardFromHand(g, game.Player1, discarded)
	survivor := addCombatPermanent(g, game.Player1, gyArrivalCreature("Survivor"))

	clone := g.Clone()
	for _, origin := range []game.GraveyardArrivalThisTurn{game.GraveyardArrivalFromBattlefield, game.GraveyardArrivalFromAnywhere} {
		for _, cardID := range []id.ID{died, discarded} {
			if gyArrivalMatches(g, cardID, origin) != gyArrivalMatches(clone, cardID, origin) {
				t.Fatalf("clone disagrees for %s origin %d", g.CardInstances[cardID].Def.Name, origin)
			}
		}
	}
	if !gyArrivalMatches(clone, died, game.GraveyardArrivalFromBattlefield) ||
		gyArrivalMatches(clone, discarded, game.GraveyardArrivalFromBattlefield) ||
		!gyArrivalMatches(clone, discarded, game.GraveyardArrivalFromAnywhere) {
		t.Fatal("clone lost this turn's graveyard arrivals")
	}

	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.Destroy{Group: game.BattlefieldGroup(game.Selection{
		PutIntoGraveyardThisTurn: game.GraveyardArrivalFromAnywhere,
	})}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if !onBattlefieldByCard(g, survivor.CardInstanceID) {
		t.Fatal("a battlefield permanent matched a graveyard-arrival predicate")
	}

	gyArrivalNextTurn(clone)
	for _, cardID := range []id.ID{died, discarded} {
		if gyArrivalMatches(clone, cardID, game.GraveyardArrivalFromAnywhere) {
			t.Fatalf("%s still matched on the next turn", clone.CardInstances[cardID].Def.Name)
		}
	}
}

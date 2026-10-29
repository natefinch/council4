package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/d"
	"github.com/natefinch/council4/mtg/cards/u"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine gap #2299 (life amount equal to an object's power or
// toughness in sequences). CR 2026-09-25 from the verified official copy:
// 107.1b (a negative result in a life or damage calculation means zero),
// 113.7a/603.10a (leaves-the-battlefield abilities look back in time; the
// ability uses the object's last known information), 608.2b (no information is
// determined from an illegal target) and 608.2h (characteristics are read once,
// when the effect is applied).
// Dying Wish GTC 64 b46e83d3-c66d-42fb-8435-b6c448db01ae: two WotC rulings
// dated 2013-01-24 at
// https://api.scryfall.com/cards/b46e83d3-c66d-42fb-8435-b6c448db01ae/rulings
// use the enchanted creature's last-known power and lose/gain nothing when it
// is negative.
// Ukkima, Stalking Shadow OTC 247 c66b452b-9c49-4c58-9966-6bafbfa34c36: three
// WotC rulings dated 2020-04-17 at
// https://api.scryfall.com/cards/c66b452b-9c49-4c58-9966-6bafbfa34c36/rulings
// use its last-known power; with negative power no damage is dealt and no life
// is gained.
// Predator's Rapport GTC 129 47324ab7-df78-4859-be0b-2eef5d4f8082: one ruling
// dated 2013-01-24 uses power and toughness as the spell resolves.
// Jaws of Defeat TDC 27 c495f3bf-86a4-4d2f-821f-c01515f084c9: one ruling dated
// 2025-04-04 subtracts the smaller of power and toughness from the larger.
// Raw responses: .cardwork/card_*.json and .cardwork/rulings_*.json.

func addPTCreature(g *game.Game, controller game.PlayerID, power, toughness int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Characteristic Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
	}})
}

func creatureTargetSpec() game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1, MaxTargets: 1, Constraint: "target creature", Allow: game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	}
}

func playerTargetSpec() game.TargetSpec {
	return game.TargetSpec{MinTargets: 1, MaxTargets: 1, Constraint: "target player", Allow: game.TargetAllowPlayer}
}

func addModeSpellToStack(g *game.Game, controller game.PlayerID, mode game.Mode, targets []game.Target) id.ID {
	sourceID := g.IDGen.Next()
	def := &game.CardDef{CardFace: game.CardFace{
		Name:         "Characteristic Life Spell",
		Types:        []types.Card{types.Sorcery},
		SpellAbility: opt.Val(mode.Ability()),
	}}
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		panic(issues)
	}
	g.CardInstances[sourceID] = &game.CardInstance{ID: sourceID, Def: def, Owner: controller}
	g.Stack.Push(&game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   sourceID,
		Controller: controller,
		Targets:    targets,
	})
	return sourceID
}

func objectCharacteristicAmount(kind game.DynamicAmountKind, ref game.ObjectReference) game.DynamicAmount {
	return game.DynamicAmount{Kind: kind, Multiplier: 1, Object: ref}
}

func assertLife(t *testing.T, g *game.Game, player game.PlayerID, want int) {
	t.Helper()
	if got := g.Players[player].Life; got != want {
		t.Fatalf("player %v life = %d, want %d", player, got, want)
	}
}

func dieWithDyingWish(t *testing.T, power, toughness, plusCounters, minusCounters int) (*game.Game, *Engine, [game.NumPlayers]PlayerAgent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addPTCreature(g, game.Player1, power, toughness)
	aura := addCombatPermanent(g, game.Player1, d.DyingWish())
	if !attachPermanent(g, aura, creature) {
		t.Fatal("Dying Wish did not attach")
	}
	creature.Counters.Add(counter.PlusOnePlusOne, plusCounters)
	creature.Counters.Add(counter.MinusOneMinusOne, minusCounters)
	if minusCounters == 0 {
		creature.MarkedDamage = toughness + plusCounters
	}
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("enchanted creature survived state-based actions")
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) || g.Stack.Size() != 1 {
		t.Fatalf("Dying Wish trigger not on stack (stack size %d)", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)
	return g, engine, agents
}

func TestDyingWishUsesEnchantedCreatureLastKnownPower(t *testing.T) {
	// A 1/2 with two +1/+1 counters dies as a 3/4; the Aura has no power.
	g, _, _ := dieWithDyingWish(t, 1, 2, 2, 0)
	assertLife(t, g, game.Player2, 37)
	assertLife(t, g, game.Player1, 43)
}

func TestDyingWishNegativeLastKnownPowerChangesNoLife(t *testing.T) {
	// A 1/3 with three -1/-1 counters dies as a -2/0 (CR 107.1b).
	g, _, _ := dieWithDyingWish(t, 1, 3, 0, 3)
	assertLife(t, g, game.Player2, 40)
	assertLife(t, g, game.Player1, 40)
}

func TestUkkimaLeavesBattlefieldUsesLastKnownPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	ukkima := addCombatPermanent(g, game.Player1, u.UkkimaStalkingShadow())
	ukkima.Counters.Add(counter.PlusOnePlusOne, 2)
	if !movePermanentToZone(g, ukkima, zone.Hand) {
		t.Fatal("Ukkima did not leave the battlefield")
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) || g.Stack.Size() != 1 {
		t.Fatalf("Ukkima trigger not on stack (stack size %d)", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)
	assertLife(t, g, game.Player2, 36)
	assertLife(t, g, game.Player1, 44)
}

func TestGainLifeEqualToTargetCreaturePowerReadsResolutionValue(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addPTCreature(g, game.Player2, 2, 2)
	addModeSpellToStack(g, game.Player1, game.Mode{
		Targets: []game.TargetSpec{creatureTargetSpec()},
		Sequence: []game.Instruction{{Primitive: game.GainLife{
			Player: game.ControllerReference(),
			Amount: game.Dynamic(objectCharacteristicAmount(game.DynamicAmountObjectPower, game.TargetPermanentReference(0))),
		}}},
	}, []game.Target{game.PermanentTarget(creature.ObjectID)})
	creature.Counters.Add(counter.PlusOnePlusOne, 3)
	engine.resolveTopOfStack(g, &TurnLog{})
	assertLife(t, g, game.Player1, 45)
}

func TestExileTargetThenLoseLifeEqualToItsLastKnownToughness(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addPTCreature(g, game.Player2, 1, 4)
	creature.Counters.Add(counter.PlusOnePlusOne, 1)
	addModeSpellToStack(g, game.Player1, game.Mode{
		Targets: []game.TargetSpec{creatureTargetSpec()},
		Sequence: []game.Instruction{
			{Primitive: game.MovePermanent{Object: game.TargetPermanentReference(0), Destination: zone.Exile}},
			{Primitive: game.LoseLife{
				Player: game.ControllerReference(),
				Amount: game.Dynamic(objectCharacteristicAmount(game.DynamicAmountObjectToughness, game.TargetPermanentReference(0))),
			}},
		},
	}, []game.Target{game.PermanentTarget(creature.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("target creature was not exiled")
	}
	assertLife(t, g, game.Player1, 35)
}

func TestIllegalCreatureTargetSuppliesNoLifeAmount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addPTCreature(g, game.Player2, 4, 4)
	addModeSpellToStack(g, game.Player1, game.Mode{
		Targets: []game.TargetSpec{playerTargetSpec(), creatureTargetSpec()},
		Sequence: []game.Instruction{{Primitive: game.LoseLife{
			Player: game.TargetPlayerReference(0),
			Amount: game.Dynamic(objectCharacteristicAmount(game.DynamicAmountObjectPower, game.TargetPermanentReference(1))),
		}}},
	}, []game.Target{game.PlayerTarget(game.Player2), game.PermanentTarget(creature.ObjectID)})
	if !movePermanentToZone(g, creature, zone.Graveyard) {
		t.Fatal("creature did not leave the battlefield")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Stack.Size() != 0 {
		t.Fatalf("spell with one legal target did not resolve (stack size %d)", g.Stack.Size())
	}
	assertLife(t, g, game.Player2, 40)
}

func TestGainLifeEqualToTargetPowerPlusToughness(t *testing.T) {
	tests := []struct {
		name                      string
		power, toughness, minuses int
		want                      int
	}{
		{name: "positive", power: 2, toughness: 3, want: 5},
		// A -2/3 gives 1: operands are summed before CR 107.1b applies.
		{name: "negative power", power: 1, toughness: 6, minuses: 3, want: 1},
		{name: "negative total", power: 0, toughness: 5, minuses: 3, want: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			creature := addPTCreature(g, game.Player1, test.power, test.toughness)
			creature.Counters.Add(counter.MinusOneMinusOne, test.minuses)
			target := game.TargetPermanentReference(0)
			addModeSpellToStack(g, game.Player1, game.Mode{
				Targets: []game.TargetSpec{creatureTargetSpec()},
				Sequence: []game.Instruction{{Primitive: game.GainLife{
					Player: game.ControllerReference(),
					Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountSumOf, Operands: []game.DynamicAmount{
						objectCharacteristicAmount(game.DynamicAmountObjectPower, target),
						objectCharacteristicAmount(game.DynamicAmountObjectToughness, target),
					}}),
				}}},
			}, []game.Target{game.PermanentTarget(creature.ObjectID)})
			engine.resolveTopOfStack(g, &TurnLog{})
			assertLife(t, g, game.Player1, 40+test.want)
		})
	}
}

func TestLoseLifeEqualToDifferenceBetweenTargetPowerAndToughness(t *testing.T) {
	tests := []struct {
		power, toughness, minuses, want int
	}{
		{power: 3, toughness: 5, want: 2},
		{power: 5, toughness: 3, want: 2},
		{power: 4, toughness: 4, want: 0},
		// A -2/3 creature: the negative power is used as-is.
		{power: 1, toughness: 6, minuses: 3, want: 5},
	}
	for _, test := range tests {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		creature := addPTCreature(g, game.Player1, test.power, test.toughness)
		creature.Counters.Add(counter.MinusOneMinusOne, test.minuses)
		target := game.TargetPermanentReference(1)
		addModeSpellToStack(g, game.Player1, game.Mode{
			Targets: []game.TargetSpec{playerTargetSpec(), creatureTargetSpec()},
			Sequence: []game.Instruction{{Primitive: game.LoseLife{
				Player: game.TargetPlayerReference(0),
				Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountDifferenceOf, Operands: []game.DynamicAmount{
					objectCharacteristicAmount(game.DynamicAmountObjectPower, target),
					objectCharacteristicAmount(game.DynamicAmountObjectToughness, target),
				}}),
			}}},
		}, []game.Target{game.PlayerTarget(game.Player2), game.PermanentTarget(creature.ObjectID)})
		engine.resolveTopOfStack(g, &TurnLog{})
		if got := g.Players[game.Player2].Life; got != 40-test.want {
			t.Fatalf("%d/%d with %d -1/-1 counters: Player2 life = %d, want %d", test.power, test.toughness, test.minuses, got, 40-test.want)
		}
	}
}

// A legacy target spec with no explicit Allow is read by the same kind
// inference at validation and resolution, so a "creature card in a graveyard"
// slot both validates and supplies the targeted card's characteristics.
func TestUnspecifiedCardTargetSuppliesLifeAmountCharacteristics(t *testing.T) {
	tests := []struct {
		name string
		kind game.DynamicAmountKind
		want int
	}{
		{"mana value", game.DynamicAmountObjectManaValue, 43},
		{"power", game.DynamicAmountObjectPower, 44},
		{"toughness", game.DynamicAmountObjectToughness, 45},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			cardID := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Graveyard Creature", Types: []types.Card{types.Creature},
				ManaCost:  opt.Val(cost.Mana{cost.O(3)}),
				Power:     opt.Val(game.PT{Value: 4}),
				Toughness: opt.Val(game.PT{Value: 5}),
			}})
			mode := game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1, MaxTargets: 1, Constraint: "target creature card in a graveyard",
					TargetZone: zone.Graveyard,
				}},
				Sequence: []game.Instruction{{Primitive: game.GainLife{
					Player: game.ControllerReference(),
					Amount: game.Dynamic(objectCharacteristicAmount(test.kind, game.TargetCardReference(0))),
				}}},
			}
			def := &game.CardDef{CardFace: game.CardFace{
				Name: "Card Target Amount", Types: []types.Card{types.Sorcery}, SpellAbility: opt.Val(mode.Ability()),
			}}
			if issues := game.ValidateCardDef(def); len(issues) != 0 {
				t.Fatalf("inferred card-target characteristic rejected: %v", issues)
			}
			addModeSpellToStack(g, game.Player1, mode, []game.Target{currentCardTarget(t, g, cardID)})
			engine.resolveTopOfStack(g, &TurnLog{})
			assertLife(t, g, game.Player1, test.want)
		})
	}
}

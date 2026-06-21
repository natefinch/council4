package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// scryRider is the Path of Ancestry spend rider: when the tagged mana is spent
// to cast a creature spell sharing a creature type with the commander, scry 1.
func scryRider() game.ManaSpendRider {
	return game.ManaSpendRider{
		Condition: game.ManaSpendCastCommanderCreatureType,
		Effect: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.Scry{Amount: game.Fixed(1), Player: game.ControllerReference()}},
		}},
	}
}

// elfCommanderGame returns a game where Player1's commander is an Elf on the
// battlefield, so a cast Elf creature spell shares a creature type with it.
func elfCommanderGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := commanderDef("Elf Commander", color.Green)
	def.Subtypes = []types.Sub{types.Elf}
	permanent := addCombatPermanent(g, game.Player1, def)
	trackCommanderID(g, game.Player1, permanent.CardInstanceID)
	return g
}

// addRiders appends n tagged rider units of the given color to the player's
// pool bookkeeping (the matching mana must be added to the pool separately).
func addRiders(g *game.Game, player game.PlayerID, c mana.Color, n int) {
	addUnitRiders(g, player, mana.Unit{Color: c}, n)
}

// addUnitRiders appends n tagged rider units of the exact mana unit (color and
// snow provenance) to the player's pool bookkeeping.
func addUnitRiders(g *game.Game, player game.PlayerID, unit mana.Unit, n int) {
	p := g.Players[player]
	for range n {
		p.ManaRiders = append(p.ManaRiders, game.ManaRiderInstance{
			Unit:       unit,
			Controller: player,
			Rider:      scryRider(),
		})
	}
}

func countTriggeredAbilities(g *game.Game) int {
	count := 0
	for _, obj := range g.Stack.Objects() {
		if obj.Kind == game.StackTriggeredAbility {
			count++
		}
	}
	return count
}

// firedSpendRiderCount reports how many mana-spend riders have fired and are
// queued to be put on the stack with the next triggered-ability pass (where they
// are ordered under APNAP). The firing path enqueues riders rather than pushing
// them directly, so firing-logic tests assert against this queue.
func firedSpendRiderCount(g *game.Game) int {
	return len(g.FiredManaSpendRiders)
}

func elfCreatureDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Elf Warrior",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Elf},
	}}
}

// TestManaSpendRiderQualifyingSpellScries covers the core provenance case: a
// rider's tagged mana spent to cast a creature spell sharing the commander's
// creature type fires its scry, and the consumed rider is removed.
func TestManaSpendRiderQualifyingSpellScries(t *testing.T) {
	t.Parallel()
	g := elfCommanderGame(t)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	addRiders(g, game.Player1, mana.G, 1)

	before := poolUnitsSnapshot(player)
	player.ManaPool.Spend(mana.G, 1)
	resolveSpellCastManaSpendRiders(g, game.Player1, before, map[mana.Unit]int{{Color: mana.G}: 1}, elfCreatureDef())

	if got := firedSpendRiderCount(g); got != 1 {
		t.Fatalf("scry triggers on stack = %d, want 1", got)
	}
	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders remaining = %d, want 0", len(player.ManaRiders))
	}
}

// TestManaSpendRiderOverproductionStillScries is a regression test for the
// missed-trigger case where a mid-payment source over-produces the rider's unit
// and leaves a leftover in the pool. Inferring spend from the gross pool delta
// would hide the tagged-mana spend; threading the exact per-unit pool spend keeps
// the rider firing. Here the pool held one tagged unit (before[G]=1) but two
// units of that unit were spent from the pool overall (spent[G]=2), so the single
// tagged unit is still recognized as consumed.
func TestManaSpendRiderOverproductionStillScries(t *testing.T) {
	t.Parallel()
	g := elfCommanderGame(t)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	addRiders(g, game.Player1, mana.G, 1)

	before := poolUnitsSnapshot(player)
	resolveSpellCastManaSpendRiders(g, game.Player1, before, map[mana.Unit]int{{Color: mana.G}: 2}, elfCreatureDef())

	if got := firedSpendRiderCount(g); got != 1 {
		t.Fatalf("scry triggers on stack = %d, want 1", got)
	}
	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders remaining = %d, want 0", len(player.ManaRiders))
	}
}

// TestManaSpendRiderMissedTriggerSameColorOverproduce is the reviewer's exact
// missed-trigger scenario: the pool holds one tagged unit, and during payment a
// source produces one more plain unit of the same color while one unit is spent,
// keeping the pool color total unchanged. A pool-total reconcile would see no
// change and miss the spend; the per-unit spend (spent[G]=1) recognizes the
// tagged unit was consumed and fires.
func TestManaSpendRiderMissedTriggerSameColorOverproduce(t *testing.T) {
	t.Parallel()
	g := elfCommanderGame(t)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	addRiders(g, game.Player1, mana.G, 1)

	// Snapshot before the payment sees one pre-existing tagged unit.
	before := poolUnitsSnapshot(player)
	// During payment: produce one more plain unit, then spend one unit. The pool
	// color total returns to one, masking the spend from a gross delta.
	player.ManaPool.Add(mana.G, 1)
	player.ManaPool.Spend(mana.G, 1)
	resolveSpellCastManaSpendRiders(g, game.Player1, before, map[mana.Unit]int{{Color: mana.G}: 1}, elfCreatureDef())

	if got := firedSpendRiderCount(g); got != 1 {
		t.Fatalf("scry triggers on stack = %d, want 1 (spend masked by overproduction)", got)
	}
	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders remaining = %d, want 0", len(player.ManaRiders))
	}
}

// TestManaSpendRiderNonCreatureSpellForcedSpendNoScry covers a non-qualifying
// spell: the tagged mana is the only mana available, so it is forcibly spent and
// the rider consumed, but no scry fires because the spell is not a creature
// spell.
func TestManaSpendRiderNonCreatureSpellForcedSpendNoScry(t *testing.T) {
	t.Parallel()
	g := elfCommanderGame(t)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	addRiders(g, game.Player1, mana.G, 1)

	instant := &game.CardDef{CardFace: game.CardFace{Name: "Shock", Types: []types.Card{types.Instant}}}
	before := poolUnitsSnapshot(player)
	player.ManaPool.Spend(mana.G, 1)
	resolveSpellCastManaSpendRiders(g, game.Player1, before, map[mana.Unit]int{{Color: mana.G}: 1}, instant)

	if got := firedSpendRiderCount(g); got != 0 {
		t.Fatalf("scry triggers on stack = %d, want 0", got)
	}
	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders remaining = %d, want 0 (forced spend)", len(player.ManaRiders))
	}
}

// TestManaSpendRiderNonCommanderTypeNoScry covers a creature spell that does not
// share a creature type with the commander: the rider does not fire, and since
// plain mana of the color is available it is preserved.
func TestManaSpendRiderNonCommanderTypeNoScry(t *testing.T) {
	t.Parallel()
	g := elfCommanderGame(t)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 2)
	addRiders(g, game.Player1, mana.G, 1)

	goblin := &game.CardDef{CardFace: game.CardFace{
		Name:     "Goblin Raider",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Goblin},
	}}
	before := poolUnitsSnapshot(player)
	player.ManaPool.Spend(mana.G, 1)
	resolveSpellCastManaSpendRiders(g, game.Player1, before, map[mana.Unit]int{{Color: mana.G}: 1}, goblin)

	if got := firedSpendRiderCount(g); got != 0 {
		t.Fatalf("scry triggers on stack = %d, want 0", got)
	}
	if len(player.ManaRiders) != 1 {
		t.Fatalf("riders remaining = %d, want 1 (plain mana spent first)", len(player.ManaRiders))
	}
}

// TestManaSpendRiderUnrelatedManaPreservesRider covers spending plain mana of
// the rider's color on an unrelated spell: the rider is preserved because the
// engine prefers to keep tagged mana for a later qualifying spell.
func TestManaSpendRiderUnrelatedManaPreservesRider(t *testing.T) {
	t.Parallel()
	g := elfCommanderGame(t)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 2)
	addRiders(g, game.Player1, mana.G, 1)

	instant := &game.CardDef{CardFace: game.CardFace{Name: "Giant Growth", Types: []types.Card{types.Instant}}}
	before := poolUnitsSnapshot(player)
	player.ManaPool.Spend(mana.G, 1)
	resolveSpellCastManaSpendRiders(g, game.Player1, before, map[mana.Unit]int{{Color: mana.G}: 1}, instant)

	if got := firedSpendRiderCount(g); got != 0 {
		t.Fatalf("scry triggers on stack = %d, want 0", got)
	}
	if len(player.ManaRiders) != 1 {
		t.Fatalf("riders remaining = %d, want 1 (rider preserved)", len(player.ManaRiders))
	}
}

// TestManaSpendRiderMixedTaggedPlainSameColor covers a pool holding both a tagged
// and a plain unit of one color. A qualifying spell consumes the tagged unit
// first (firing), while a non-qualifying spell consumes the plain unit first
// (preserving the rider).
func TestManaSpendRiderMixedTaggedPlainSameColor(t *testing.T) {
	t.Parallel()
	t.Run("qualifying consumes tagged first", func(t *testing.T) {
		t.Parallel()
		g := elfCommanderGame(t)
		player := g.Players[game.Player1]
		player.ManaPool.Add(mana.G, 2) // one tagged, one plain
		addRiders(g, game.Player1, mana.G, 1)

		before := poolUnitsSnapshot(player)
		player.ManaPool.Spend(mana.G, 1)
		resolveSpellCastManaSpendRiders(g, game.Player1, before, map[mana.Unit]int{{Color: mana.G}: 1}, elfCreatureDef())

		if got := firedSpendRiderCount(g); got != 1 {
			t.Fatalf("scry triggers on stack = %d, want 1 (tagged spent first)", got)
		}
		if len(player.ManaRiders) != 0 {
			t.Fatalf("riders remaining = %d, want 0", len(player.ManaRiders))
		}
	})
	t.Run("non-qualifying consumes plain first", func(t *testing.T) {
		t.Parallel()
		g := elfCommanderGame(t)
		player := g.Players[game.Player1]
		player.ManaPool.Add(mana.G, 2)
		addRiders(g, game.Player1, mana.G, 1)

		goblin := &game.CardDef{CardFace: game.CardFace{
			Name: "Goblin", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Goblin},
		}}
		before := poolUnitsSnapshot(player)
		player.ManaPool.Spend(mana.G, 1)
		resolveSpellCastManaSpendRiders(g, game.Player1, before, map[mana.Unit]int{{Color: mana.G}: 1}, goblin)

		if got := firedSpendRiderCount(g); got != 0 {
			t.Fatalf("scry triggers on stack = %d, want 0", got)
		}
		if len(player.ManaRiders) != 1 {
			t.Fatalf("riders remaining = %d, want 1 (plain spent first)", len(player.ManaRiders))
		}
	})
	t.Run("qualifying spending both fires once", func(t *testing.T) {
		t.Parallel()
		g := elfCommanderGame(t)
		player := g.Players[game.Player1]
		player.ManaPool.Add(mana.G, 2)
		addRiders(g, game.Player1, mana.G, 1)

		before := poolUnitsSnapshot(player)
		player.ManaPool.Spend(mana.G, 2)
		resolveSpellCastManaSpendRiders(g, game.Player1, before, map[mana.Unit]int{{Color: mana.G}: 2}, elfCreatureDef())

		if got := firedSpendRiderCount(g); got != 1 {
			t.Fatalf("scry triggers on stack = %d, want 1 (only one tagged unit)", got)
		}
		if len(player.ManaRiders) != 0 {
			t.Fatalf("riders remaining = %d, want 0", len(player.ManaRiders))
		}
	})
}

// TestManaSpendRiderSnowProvenanceDistinct proves provenance is tracked per exact
// unit, not per color: a snow-tagged unit is not consumed when a same-color but
// non-snow unit is spent.
func TestManaSpendRiderSnowProvenanceDistinct(t *testing.T) {
	t.Parallel()
	g := elfCommanderGame(t)
	player := g.Players[game.Player1]
	player.ManaPool.AddSnow(mana.G, 1)
	player.ManaPool.Add(mana.G, 1)
	addUnitRiders(g, game.Player1, mana.Unit{Color: mana.G, Snow: true}, 1)

	before := poolUnitsSnapshot(player)
	// Spend the non-snow unit on a qualifying spell.
	player.ManaPool.Spend(mana.G, 1)
	resolveSpellCastManaSpendRiders(g, game.Player1, before, map[mana.Unit]int{{Color: mana.G}: 1}, elfCreatureDef())

	if got := firedSpendRiderCount(g); got != 0 {
		t.Fatalf("scry triggers on stack = %d, want 0 (snow rider not spent)", got)
	}
	if len(player.ManaRiders) != 1 {
		t.Fatalf("riders remaining = %d, want 1 (snow rider preserved)", len(player.ManaRiders))
	}
}

// TestManaSpendRiderMultipleUnits covers multiple activations: two tagged units
// spent on one qualifying creature spell fire two scries and consume both
// riders, while a partial spend consumes and fires exactly one.
func TestManaSpendRiderMultipleUnits(t *testing.T) {
	t.Parallel()
	t.Run("all spent", func(t *testing.T) {
		t.Parallel()
		g := elfCommanderGame(t)
		player := g.Players[game.Player1]
		player.ManaPool.Add(mana.G, 2)
		addRiders(g, game.Player1, mana.G, 2)

		before := poolUnitsSnapshot(player)
		player.ManaPool.Spend(mana.G, 2)
		resolveSpellCastManaSpendRiders(g, game.Player1, before, map[mana.Unit]int{{Color: mana.G}: 2}, elfCreatureDef())

		if got := firedSpendRiderCount(g); got != 2 {
			t.Fatalf("scry triggers on stack = %d, want 2", got)
		}
		if len(player.ManaRiders) != 0 {
			t.Fatalf("riders remaining = %d, want 0", len(player.ManaRiders))
		}
	})
	t.Run("partial spend", func(t *testing.T) {
		t.Parallel()
		g := elfCommanderGame(t)
		player := g.Players[game.Player1]
		player.ManaPool.Add(mana.G, 2)
		addRiders(g, game.Player1, mana.G, 2)

		before := poolUnitsSnapshot(player)
		player.ManaPool.Spend(mana.G, 1)
		resolveSpellCastManaSpendRiders(g, game.Player1, before, map[mana.Unit]int{{Color: mana.G}: 1}, elfCreatureDef())

		if got := firedSpendRiderCount(g); got != 1 {
			t.Fatalf("scry triggers on stack = %d, want 1", got)
		}
		if len(player.ManaRiders) != 1 {
			t.Fatalf("riders remaining = %d, want 1", len(player.ManaRiders))
		}
	})
}

// TestManaSpendRiderEmptyPoolClearsRiders covers unused mana: emptying mana
// pools between steps discards tagged mana along with the riders, so leftover
// mana never fires a later scry.
func TestManaSpendRiderEmptyPoolClearsRiders(t *testing.T) {
	t.Parallel()
	g := elfCommanderGame(t)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	addRiders(g, game.Player1, mana.G, 1)

	emptyManaPools(g)

	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders remaining = %d, want 0 after empty", len(player.ManaRiders))
	}
}

// TestManaSpendRiderNonSpellPaymentDropsRider covers tagged mana spent on a
// non-spell payment (an activated ability, a ward or additional cost): the
// payment consumes the rider's unit and drops it without firing, because a
// non-spell payment never satisfies the rider condition.
func TestManaSpendRiderNonSpellPaymentDropsRider(t *testing.T) {
	t.Parallel()
	g := elfCommanderGame(t)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	addRiders(g, game.Player1, mana.G, 1)

	before := poolUnitsSnapshot(player)
	player.ManaPool.Spend(mana.G, 1)
	consumeManaSpendRidersForPayment(g, game.Player1, nil, before, map[mana.Unit]int{{Color: mana.G}: 1})

	if got := firedSpendRiderCount(g); got != 0 {
		t.Fatalf("scry triggers on stack = %d, want 0 (non-spell payment)", got)
	}
	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders remaining = %d, want 0 (consumed by non-spell payment)", len(player.ManaRiders))
	}
}

// TestManaSpendRiderNoFalseReattachAfterNonSpellSpend is the adversarial
// false-trigger regression: tagged mana spent on a non-spell payment is consumed
// immediately, so when plain mana of the same color is later produced and spent
// to cast a qualifying creature spell, the stale rider cannot reattach to it and
// no scry fires.
func TestManaSpendRiderNoFalseReattachAfterNonSpellSpend(t *testing.T) {
	t.Parallel()
	g := elfCommanderGame(t)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	addRiders(g, game.Player1, mana.G, 1)

	// The tagged green is spent on a non-spell payment (e.g. an ability).
	beforeAbility := poolUnitsSnapshot(player)
	player.ManaPool.Spend(mana.G, 1)
	consumeManaSpendRidersForPayment(g, game.Player1, nil, beforeAbility, map[mana.Unit]int{{Color: mana.G}: 1})
	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders after ability payment = %d, want 0", len(player.ManaRiders))
	}

	// Later, plain green is produced and spent to cast a qualifying creature.
	player.ManaPool.Add(mana.G, 1)
	beforeSpell := poolUnitsSnapshot(player)
	player.ManaPool.Spend(mana.G, 1)
	resolveSpellCastManaSpendRiders(g, game.Player1, beforeSpell, map[mana.Unit]int{{Color: mana.G}: 1}, elfCreatureDef())

	if got := firedSpendRiderCount(g); got != 0 {
		t.Fatalf("scry triggers on stack = %d, want 0 (no stale reattach)", got)
	}
}

// TestManaSpendRiderCommanderFaceDownFailsClosed covers MEDIUM #2: when the
// commander is a face-down battlefield permanent it has no creature subtypes, so
// the rider must fail closed and not fire even for a creature spell that shares
// the commander's printed subtype.
func TestManaSpendRiderCommanderFaceDownFailsClosed(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := commanderDef("Elf Commander", color.Green)
	def.Subtypes = []types.Sub{types.Elf}
	permanent := addCombatPermanent(g, game.Player1, def)
	permanent.FaceDown = true
	permanent.FaceDownKind = game.FaceDownMorph
	trackCommanderID(g, game.Player1, permanent.CardInstanceID)

	if spellSatisfiesCommanderCreatureTypeRider(g, game.Player1, elfCreatureDef()) {
		t.Fatal("face-down commander satisfied the rider on printed subtype")
	}
}

// TestManaSpendRiderCommanderOffBattlefieldUsesFrontFace covers MEDIUM #2: a
// commander outside the battlefield (e.g. in the command zone) has no continuous
// effects, so its printed front-face subtypes are its current characteristics
// and a creature spell sharing them satisfies the rider.
func TestManaSpendRiderCommanderOffBattlefieldUsesFrontFace(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := commanderDef("Elf Commander", color.Green)
	def.Subtypes = []types.Sub{types.Elf}
	cardID := addCardInstance(g, game.Player1, def)
	trackCommanderID(g, game.Player1, cardID)

	if !spellSatisfiesCommanderCreatureTypeRider(g, game.Player1, elfCreatureDef()) {
		t.Fatal("off-battlefield commander did not satisfy the rider on its front-face subtype")
	}
}

// TestSpellSatisfiesCommanderCreatureTypeRiderFailsClosed covers the
// qualification boundaries: a non-creature spell, a creature with no shared
// type, and a controller without a single modeled commander never satisfy the
// rider.
func TestSpellSatisfiesCommanderCreatureTypeRiderFailsClosed(t *testing.T) {
	t.Parallel()
	t.Run("non-creature spell", func(t *testing.T) {
		t.Parallel()
		g := elfCommanderGame(t)
		instant := &game.CardDef{CardFace: game.CardFace{Name: "Shock", Types: []types.Card{types.Instant}}}
		if spellSatisfiesCommanderCreatureTypeRider(g, game.Player1, instant) {
			t.Fatal("non-creature spell satisfied the rider")
		}
	})
	t.Run("no shared type", func(t *testing.T) {
		t.Parallel()
		g := elfCommanderGame(t)
		goblin := &game.CardDef{CardFace: game.CardFace{
			Name: "Goblin", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Goblin},
		}}
		if spellSatisfiesCommanderCreatureTypeRider(g, game.Player1, goblin) {
			t.Fatal("creature with no shared type satisfied the rider")
		}
	})
	t.Run("no modeled commander", func(t *testing.T) {
		t.Parallel()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		if spellSatisfiesCommanderCreatureTypeRider(g, game.Player1, elfCreatureDef()) {
			t.Fatal("missing commander satisfied the rider")
		}
	})
}

// TestManaSpendRiderEndToEndPathOfAncestry casts a real creature spell paying
// from a pool that holds Path of Ancestry's tagged mana, exercising the live
// cast path's snapshot and resolve hooks so the spend-linked scry fires.
func TestManaSpendRiderEndToEndPathOfAncestry(t *testing.T) {
	t.Parallel()
	g := elfCommanderGame(t)
	engine := NewEngine(nil)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	addRiders(g, game.Player1, mana.G, 1)

	spellDef := elfCreatureDef()
	spellDef.ManaCost = opt.Val(cost.Mana{cost.G})
	spellDef.Power = opt.Val(game.PT{Value: 1})
	spellDef.Toughness = opt.Val(game.PT{Value: 1})
	spellID := addCardToHand(g, game.Player1, spellDef)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast Elf creature) = false, want true")
	}
	// The rider fired into the pending queue during the cast; it reaches the
	// stack only through the normal triggered-ability pass (ordered under APNAP).
	if got := firedSpendRiderCount(g); got != 1 {
		t.Fatalf("fired spend riders queued = %d, want 1", got)
	}
	engine.putTriggeredAbilitiesOnStack(g)
	if got := countTriggeredAbilities(g); got != 1 {
		t.Fatalf("scry triggers on stack = %d, want 1", got)
	}
	if got := firedSpendRiderCount(g); got != 0 {
		t.Fatalf("fired spend riders queued after pass = %d, want 0 (drained)", got)
	}
	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders remaining = %d, want 0", len(player.ManaRiders))
	}
}

func TestChosenTypeManaSpendRiderMakesQualifyingSpellUncounterable(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	player.ManaRiders = append(player.ManaRiders, game.ManaRiderInstance{
		Unit:          mana.Unit{Color: mana.G},
		Controller:    game.Player1,
		ChosenSubtype: types.Elf,
		Rider: game.ManaSpendRider{
			Condition:         game.ManaSpendCastChosenCreatureType,
			Restriction:       game.ManaSpendRestrictedToCondition,
			SpellRuleEffect:   game.RuleEffectCantBeCountered,
			ChosenSubtypeFrom: game.EntryTypeChoiceKey,
		},
	})

	spellDef := elfCreatureDef()
	spellDef.ManaCost = opt.Val(cost.Mana{cost.G})
	spellDef.Power = opt.Val(game.PT{Value: 1})
	spellDef.Toughness = opt.Val(game.PT{Value: 1})
	spellID := addCardToHand(g, game.Player1, spellDef)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast chosen-type creature) = false, want true")
	}
	stackObjects := g.Stack.Objects()
	if len(stackObjects) != 1 || stackObjects[0].Kind != game.StackSpell {
		t.Fatalf("stack = %#v, want one spell", stackObjects)
	}
	if stackSpellCanBeCountered(g, stackObjects[0]) {
		t.Fatal("qualifying spell paid with tagged mana can be countered")
	}
	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders remaining = %d, want 0", len(player.ManaRiders))
	}
}

func TestChosenTypeManaCannotPayForNonqualifyingSpell(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	player.ManaRiders = append(player.ManaRiders, game.ManaRiderInstance{
		Unit:          mana.Unit{Color: mana.G},
		Controller:    game.Player1,
		ChosenSubtype: types.Elf,
		Rider: game.ManaSpendRider{
			Condition:         game.ManaSpendCastChosenCreatureType,
			Restriction:       game.ManaSpendRestrictedToCondition,
			SpellRuleEffect:   game.RuleEffectCantBeCountered,
			ChosenSubtypeFrom: game.EntryTypeChoiceKey,
		},
	})

	goblin := creatureSpellDef("Goblin", types.Goblin)
	goblin.ManaCost = opt.Val(cost.Mana{cost.G})
	goblin.Power = opt.Val(game.PT{Value: 1})
	goblin.Toughness = opt.Val(game.PT{Value: 1})
	spellID := addCardToHand(g, game.Player1, goblin)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast non-chosen creature with restricted mana succeeded")
	}
	if player.ManaPool.Amount(mana.G) != 1 || len(player.ManaRiders) != 1 {
		t.Fatalf("restricted mana changed after rejected cast: pool=%d riders=%d", player.ManaPool.Amount(mana.G), len(player.ManaRiders))
	}
}

func TestLegendaryManaSpendRiderMakesQualifyingSpellUncounterable(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	player.ManaRiders = append(player.ManaRiders, game.ManaRiderInstance{
		Unit:       mana.Unit{Color: mana.G},
		Controller: game.Player1,
		Rider: game.ManaSpendRider{
			Condition:       game.ManaSpendCastLegendarySpell,
			Restriction:     game.ManaSpendRestrictedToCondition,
			SpellRuleEffect: game.RuleEffectCantBeCountered,
		},
	})

	spellDef := elfCreatureDef()
	spellDef.Supertypes = []types.Super{types.Legendary}
	spellDef.ManaCost = opt.Val(cost.Mana{cost.G})
	spellDef.Power = opt.Val(game.PT{Value: 1})
	spellDef.Toughness = opt.Val(game.PT{Value: 1})
	spellID := addCardToHand(g, game.Player1, spellDef)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast legendary creature) = false, want true")
	}
	stackObjects := g.Stack.Objects()
	if len(stackObjects) != 1 || stackObjects[0].Kind != game.StackSpell {
		t.Fatalf("stack = %#v, want one spell", stackObjects)
	}
	if stackSpellCanBeCountered(g, stackObjects[0]) {
		t.Fatal("qualifying legendary spell paid with tagged mana can be countered")
	}
	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders remaining = %d, want 0", len(player.ManaRiders))
	}
}

func TestLegendaryManaCannotPayForNonlegendarySpell(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	player.ManaRiders = append(player.ManaRiders, game.ManaRiderInstance{
		Unit:       mana.Unit{Color: mana.G},
		Controller: game.Player1,
		Rider: game.ManaSpendRider{
			Condition:       game.ManaSpendCastLegendarySpell,
			Restriction:     game.ManaSpendRestrictedToCondition,
			SpellRuleEffect: game.RuleEffectCantBeCountered,
		},
	})

	goblin := creatureSpellDef("Goblin", types.Goblin)
	goblin.ManaCost = opt.Val(cost.Mana{cost.G})
	goblin.Power = opt.Val(game.PT{Value: 1})
	goblin.Toughness = opt.Val(game.PT{Value: 1})
	spellID := addCardToHand(g, game.Player1, goblin)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast nonlegendary creature with legendary-restricted mana succeeded")
	}
	if player.ManaPool.Amount(mana.G) != 1 || len(player.ManaRiders) != 1 {
		t.Fatalf("restricted mana changed after rejected cast: pool=%d riders=%d", player.ManaPool.Amount(mana.G), len(player.ManaRiders))
	}
}

// TestCreatureSpellHasteManaSpendRiderGrantsHaste covers the unrestricted Arena
// of Glory / Generator Servant bonus rider: a creature spell paid for with the
// tagged mana resolves with haste until end of turn, and the haste is cleared at
// cleanup.
func TestCreatureSpellHasteManaSpendRiderGrantsHaste(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.R, 1)
	player.ManaRiders = append(player.ManaRiders, game.ManaRiderInstance{
		Unit:       mana.Unit{Color: mana.R},
		Controller: game.Player1,
		Rider: game.ManaSpendRider{
			Condition:          game.ManaSpendCastCreatureSpell,
			SpellGainsKeywords: []game.Keyword{game.Haste},
		},
	})

	goblin := creatureSpellDef("Goblin", types.Goblin)
	goblin.ManaCost = opt.Val(cost.Mana{cost.R})
	goblin.Power = opt.Val(game.PT{Value: 1})
	goblin.Toughness = opt.Val(game.PT{Value: 1})
	spellID := addCardToHand(g, game.Player1, goblin)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast creature with haste-tagged mana) = false, want true")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent, ok := findPermanentByCardID(g, spellID)
	if !ok {
		t.Fatal("creature spell did not resolve to the battlefield")
	}
	if !hasKeyword(g, permanent, game.Haste) {
		t.Fatal("creature paid with haste-tagged mana did not gain haste")
	}
	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders remaining = %d, want 0", len(player.ManaRiders))
	}
	expireCleanupDurations(g)
	if hasKeyword(g, permanent, game.Haste) {
		t.Fatal("granted haste persisted past end of turn")
	}
}

// TestCreatureSpellHasteManaSpendRiderUntaggedManaNoHaste verifies the haste
// grant is gated on spending the tagged mana: a creature cast with ordinary mana
// resolves without haste even when the controller also holds a haste rider unit.
func TestCreatureSpellHasteManaSpendRiderUntaggedManaNoHaste(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.R, 1)

	goblin := creatureSpellDef("Goblin", types.Goblin)
	goblin.ManaCost = opt.Val(cost.Mana{cost.R})
	goblin.Power = opt.Val(game.PT{Value: 1})
	goblin.Toughness = opt.Val(game.PT{Value: 1})
	spellID := addCardToHand(g, game.Player1, goblin)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast creature with plain mana) = false, want true")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent, ok := findPermanentByCardID(g, spellID)
	if !ok {
		t.Fatal("creature spell did not resolve to the battlefield")
	}
	if hasKeyword(g, permanent, game.Haste) {
		t.Fatal("creature cast with untagged mana gained haste")
	}
}

func TestAddManaCapturesChosenSubtypeAndSourceIdentity(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	rider := game.ManaSpendRider{
		Condition:         game.ManaSpendCastChosenCreatureType,
		Restriction:       game.ManaSpendRestrictedToCondition,
		SpellRuleEffect:   game.RuleEffectCantBeCountered,
		ChosenSubtypeFrom: game.EntryTypeChoiceKey,
	}
	sourceCardID := addInstructionSpellToStack(g, []game.Instruction{{
		Primitive: game.AddMana{
			Amount:     game.Fixed(1),
			ManaColor:  mana.G,
			SpendRider: opt.Val(rider),
		},
	}})
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack is empty")
	}
	obj.SourceCardID = sourceCardID
	obj.ResolutionChoices = map[string]game.ResolutionChoiceResult{
		string(game.EntryTypeChoiceKey): {
			Kind:    game.ResolutionChoiceSubtype,
			Subtype: types.Elf,
		},
	}
	sourceObjectID := obj.SourceID

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	instances := g.Players[game.Player1].ManaRiders
	if len(instances) != 1 {
		t.Fatalf("mana riders = %#v, want one", instances)
	}
	if instances[0].ChosenSubtype != types.Elf ||
		instances[0].SourceID != sourceCardID ||
		instances[0].SourceObjectID != sourceObjectID {
		t.Fatalf("captured rider = %#v", instances[0])
	}
}

func TestChosenTypeManaFailsClosedWithoutEntryChoice(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addInstructionSpellToStack(g, []game.Instruction{{
		Primitive: game.AddMana{
			Amount:    game.Fixed(1),
			ManaColor: mana.G,
			SpendRider: opt.Val(game.ManaSpendRider{
				Condition:         game.ManaSpendCastChosenCreatureType,
				Restriction:       game.ManaSpendRestrictedToCondition,
				SpellRuleEffect:   game.RuleEffectCantBeCountered,
				ChosenSubtypeFrom: game.EntryTypeChoiceKey,
			}),
		},
	}})

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	player := g.Players[game.Player1]
	if got := player.ManaPool.Amount(mana.G); got != 0 {
		t.Fatalf("mana pool = %d, want 0 when the chosen subtype is unavailable", got)
	}
	if len(player.ManaRiders) != 0 {
		t.Fatalf("mana riders = %#v, want none", player.ManaRiders)
	}
}

func TestChosenTypeManaCannotPayNonspellCost(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	player.ManaRiders = append(player.ManaRiders, game.ManaRiderInstance{
		Unit:          mana.Unit{Color: mana.G},
		Controller:    game.Player1,
		ChosenSubtype: types.Elf,
		Rider: game.ManaSpendRider{
			Condition:         game.ManaSpendCastChosenCreatureType,
			Restriction:       game.ManaSpendRestrictedToCondition,
			SpellRuleEffect:   game.RuleEffectCantBeCountered,
			ChosenSubtypeFrom: game.EntryTypeChoiceKey,
		},
	})
	genericCost := cost.Mana{cost.G}
	if paymentOrch.payGenericCost(g, payment.GenericRequest{
		PlayerID: game.Player1,
		Cost:     &genericCost,
	}) {
		t.Fatal("restricted chosen-type mana paid a nonspell cost")
	}
	if player.ManaPool.Amount(mana.G) != 1 || len(player.ManaRiders) != 1 {
		t.Fatalf("restricted mana changed after rejected payment: pool=%d riders=%d", player.ManaPool.Amount(mana.G), len(player.ManaRiders))
	}
}

// TestChosenTypeCastOrActivateManaPaysCreatureSourceAbility verifies the
// cast-or-activate chosen-type restriction (Secluded Courtyard): its tagged mana
// may pay to activate an ability of a creature source of the chosen type and is
// consumed without firing.
func TestChosenTypeCastOrActivateManaPaysCreatureSourceAbility(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	player.ManaRiders = append(player.ManaRiders, game.ManaRiderInstance{
		Unit:          mana.Unit{Color: mana.G},
		Controller:    game.Player1,
		ChosenSubtype: types.Elf,
		Rider: game.ManaSpendRider{
			Condition:         game.ManaSpendCastOrActivateChosenCreatureType,
			Restriction:       game.ManaSpendRestrictedToCondition,
			ChosenSubtypeFrom: game.EntryTypeChoiceKey,
		},
	})

	elfSource := addPermanentForSBA(g, game.Player1, elfCreatureDef())
	manaCost := cost.Mana{cost.G}
	if _, ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID: game.Player1,
		Source:   elfSource,
		ManaCost: opt.Val(manaCost),
	}); !ok {
		t.Fatal("cast-or-activate mana did not pay an Elf source's ability cost")
	}
	if player.ManaPool.Amount(mana.G) != 0 || len(player.ManaRiders) != 0 {
		t.Fatalf("rider not consumed by qualifying activation: pool=%d riders=%d", player.ManaPool.Amount(mana.G), len(player.ManaRiders))
	}
	if firedSpendRiderCount(g) != 0 {
		t.Fatalf("effectless rider fired an ability: count=%d", firedSpendRiderCount(g))
	}
}

// TestChosenTypeCastOrActivateManaRejectsNonCreatureSourceAbility verifies the
// cast-or-activate chosen-type restriction does not admit its tagged mana to the
// activated-ability cost of a source that is not a creature of the chosen type.
func TestChosenTypeCastOrActivateManaRejectsNonCreatureSourceAbility(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	player.ManaRiders = append(player.ManaRiders, game.ManaRiderInstance{
		Unit:          mana.Unit{Color: mana.G},
		Controller:    game.Player1,
		ChosenSubtype: types.Elf,
		Rider: game.ManaSpendRider{
			Condition:         game.ManaSpendCastOrActivateChosenCreatureType,
			Restriction:       game.ManaSpendRestrictedToCondition,
			ChosenSubtypeFrom: game.EntryTypeChoiceKey,
		},
	})

	goblinSource := addPermanentForSBA(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Goblin Source",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Goblin},
	}})
	manaCost := cost.Mana{cost.G}
	if _, ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID: game.Player1,
		Source:   goblinSource,
		ManaCost: opt.Val(manaCost),
	}); ok {
		t.Fatal("cast-or-activate mana paid a non-matching source's ability cost")
	}
	if player.ManaPool.Amount(mana.G) != 1 || len(player.ManaRiders) != 1 {
		t.Fatalf("restricted mana changed after rejected payment: pool=%d riders=%d", player.ManaPool.Amount(mana.G), len(player.ManaRiders))
	}
}

func TestChosenTypeManaProvenanceDistinguishesSameColorUnits(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 2)
	rider := game.ManaSpendRider{
		Condition:         game.ManaSpendCastChosenCreatureType,
		Restriction:       game.ManaSpendRestrictedToCondition,
		SpellRuleEffect:   game.RuleEffectCantBeCountered,
		ChosenSubtypeFrom: game.EntryTypeChoiceKey,
	}
	player.ManaRiders = append(player.ManaRiders,
		game.ManaRiderInstance{
			Unit:          mana.Unit{Color: mana.G},
			Controller:    game.Player1,
			ChosenSubtype: types.Goblin,
			Rider:         rider,
		},
		game.ManaRiderInstance{
			Unit:          mana.Unit{Color: mana.G},
			Controller:    game.Player1,
			ChosenSubtype: types.Elf,
			Rider:         rider,
		},
	)

	spellDef := elfCreatureDef()
	spellDef.ManaCost = opt.Val(cost.Mana{cost.G})
	spellDef.Power = opt.Val(game.PT{Value: 1})
	spellDef.Toughness = opt.Val(game.PT{Value: 1})
	spellID := addCardToHand(g, game.Player1, spellDef)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast Elf creature) = false, want true")
	}
	stackObjects := g.Stack.Objects()
	if len(stackObjects) != 1 || stackSpellCanBeCountered(g, stackObjects[0]) {
		t.Fatalf("qualifying spell lacks uncounterable provenance: stack=%#v", stackObjects)
	}
	if len(player.ManaRiders) != 1 || player.ManaRiders[0].ChosenSubtype != types.Goblin {
		t.Fatalf("remaining provenance = %#v, want Goblin unit", player.ManaRiders)
	}
}

func TestChosenTypeManaPreferredOverSameColorUnrestrictedRider(t *testing.T) {
	t.Parallel()
	g := elfCommanderGame(t)
	engine := NewEngine(nil)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 2)
	player.ManaRiders = append(player.ManaRiders,
		game.ManaRiderInstance{
			Unit:       mana.Unit{Color: mana.G},
			Controller: game.Player1,
			Rider:      scryRider(),
		},
		game.ManaRiderInstance{
			Unit:          mana.Unit{Color: mana.G},
			Controller:    game.Player1,
			ChosenSubtype: types.Elf,
			Rider: game.ManaSpendRider{
				Condition:         game.ManaSpendCastChosenCreatureType,
				Restriction:       game.ManaSpendRestrictedToCondition,
				SpellRuleEffect:   game.RuleEffectCantBeCountered,
				ChosenSubtypeFrom: game.EntryTypeChoiceKey,
			},
		},
	)

	spellDef := elfCreatureDef()
	spellDef.ManaCost = opt.Val(cost.Mana{cost.G})
	spellDef.Power = opt.Val(game.PT{Value: 1})
	spellDef.Toughness = opt.Val(game.PT{Value: 1})
	spellID := addCardToHand(g, game.Player1, spellDef)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast Elf creature) = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok || stackSpellCanBeCountered(g, obj) {
		t.Fatalf("spell paid from mixed same-color provenance is counterable: stack=%#v", obj)
	}
	if len(player.ManaRiders) != 1 ||
		player.ManaRiders[0].Rider.Condition != game.ManaSpendCastCommanderCreatureType {
		t.Fatalf("remaining provenance = %#v, want unrestricted commander-type rider", player.ManaRiders)
	}
}

func TestManaSpendUncounterableEffectIsStackObjectScoped(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := elfCreatureDef()
	protected := &game.StackObject{
		ID:             g.IDGen.Next(),
		Kind:           game.StackSpell,
		Controller:     game.Player1,
		SourceTokenDef: def,
	}
	other := &game.StackObject{
		ID:             g.IDGen.Next(),
		Kind:           game.StackSpell,
		Controller:     game.Player1,
		SourceTokenDef: def,
	}
	g.Stack.Push(protected)
	g.Stack.Push(other)
	applyManaSpendSpellRuleEffect(protected, game.ManaRiderInstance{
		Controller: game.Player1,
		Rider: game.ManaSpendRider{
			SpellRuleEffect: game.RuleEffectCantBeCountered,
		},
	})

	if stackSpellCanBeCountered(g, protected) {
		t.Fatal("protected stack object can be countered")
	}
	if !stackSpellCanBeCountered(g, other) {
		t.Fatal("unrelated stack object inherited mana-spend protection")
	}
}

// TestManaSpendRiderEndToEndNonSpellAbilityNoFalseScry exercises the live
// payment path: tagged mana spent paying a generic (non-spell) mana cost is
// consumed without firing, proving the orchestrator processes riders on the
// ability payment path too.
func TestManaSpendRiderEndToEndNonSpellAbilityNoFalseScry(t *testing.T) {
	t.Parallel()
	g := elfCommanderGame(t)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	addRiders(g, game.Player1, mana.G, 1)

	genericCost := cost.Mana{cost.G}
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{
		PlayerID: game.Player1,
		Cost:     &genericCost,
	}) {
		t.Fatal("payGenericCost({G}) = false, want true")
	}
	if got := firedSpendRiderCount(g); got != 0 {
		t.Fatalf("scry triggers on stack = %d, want 0 (ability payment)", got)
	}
	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders remaining = %d, want 0 (consumed by ability payment)", len(player.ManaRiders))
	}
}

// commanderCardDef registers a commander card instance in g (off the
// battlefield) and tracks it for the player, returning its instance ID. It is
// used to place the commander on the stack or in another zone for current-
// characteristic tests without putting it on the battlefield.
func commanderCardDef(g *game.Game, player game.PlayerID, def *game.CardDef) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: player}
	trackCommanderID(g, player, cardID)
	return cardID
}

func creatureSpellDef(name string, subtype types.Sub) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{subtype},
	}}
}

func madnessCreature(name string, subtype types.Sub, manaCost cost.Mana) *game.CardDef {
	def := creatureSpellDef(name, subtype)
	def.Power = opt.Val(game.PT{Value: 1})
	def.Toughness = opt.Val(game.PT{Value: 1})
	def.StaticAbilities = []game.StaticAbility{{
		KeywordAbilities: []game.KeywordAbility{game.MadnessKeyword{Cost: manaCost}},
	}}
	return def
}

// TestManaSpendRiderManualActivationTagsMana covers MEDIUM #2: a fixed-output
// mana ability that carries a spend rider is excluded from the automatic payment
// path (which would add untagged mana) and instead stays a manual action whose
// activation tags the produced mana with its rider.
func TestManaSpendRiderManualActivationTagsMana(t *testing.T) {
	t.Parallel()
	g := elfCommanderGame(t)
	engine := NewEngine(nil)
	manaBody := game.ManaAbility{
		Text:            "{T}: Add {G}. (rider)",
		AdditionalCosts: cost.Tap,
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.G, SpendRider: opt.Val(scryRider())},
		}}}.Ability(),
	}
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:          "Ancestral Source",
		Types:         []types.Card{types.Artifact},
		ManaAbilities: []game.ManaAbility{manaBody},
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("rider-bearing mana ability was not exposed as a manual action")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(rider mana ability) = false, want true")
	}
	player := g.Players[game.Player1]
	if got := player.ManaPool.Amount(mana.G); got != 1 {
		t.Fatalf("pool G = %d, want 1 after manual activation", got)
	}
	if len(player.ManaRiders) != 1 {
		t.Fatalf("tagged riders = %d, want 1 (manual activation tags mana)", len(player.ManaRiders))
	}
}

// TestManaSpendRiderMadnessQualifyingCastScries covers MEDIUM #1: a madness cast
// is a spell cast, so tagged mana spent on its madness cost fires the rider when
// the madness spell shares the commander's creature type.
func TestManaSpendRiderMadnessQualifyingCastScries(t *testing.T) {
	t.Parallel()
	g := elfCommanderGame(t)
	engine := NewEngine(nil)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	addRiders(g, game.Player1, mana.G, 1)

	cardID := addCardToHand(g, game.Player1, madnessCreature("Madness Elf", types.Elf, cost.Mana{cost.G}))
	if !discardCardFromHand(g, game.Player1, cardID) {
		t.Fatal("discardCardFromHand() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("madness trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := firedSpendRiderCount(g); got != 1 {
		t.Fatalf("fired spend riders = %d, want 1 (qualifying madness cast)", got)
	}
	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders remaining = %d, want 0", len(player.ManaRiders))
	}
	if obj, ok := g.Stack.Peek(); !ok || obj.Kind != game.StackSpell || obj.SourceID != cardID {
		t.Fatalf("stack top = %+v, want madness creature spell", obj)
	}
}

func TestChosenTypeManaPaysQualifyingMadnessAndMakesItUncounterable(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	player.ManaRiders = append(player.ManaRiders, game.ManaRiderInstance{
		Unit:          mana.Unit{Color: mana.G},
		Controller:    game.Player1,
		ChosenSubtype: types.Elf,
		Rider: game.ManaSpendRider{
			Condition:         game.ManaSpendCastChosenCreatureType,
			Restriction:       game.ManaSpendRestrictedToCondition,
			SpellRuleEffect:   game.RuleEffectCantBeCountered,
			ChosenSubtypeFrom: game.EntryTypeChoiceKey,
		},
	})

	cardID := addCardToHand(g, game.Player1, madnessCreature("Madness Elf", types.Elf, cost.Mana{cost.G}))
	if !discardCardFromHand(g, game.Player1, cardID) {
		t.Fatal("discardCardFromHand() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("madness trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackSpell || obj.SourceID != cardID {
		t.Fatalf("stack top = %+v, want madness creature spell", obj)
	}
	if stackSpellCanBeCountered(g, obj) {
		t.Fatal("qualifying madness spell paid with chosen-type mana can be countered")
	}
	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders remaining = %d, want 0", len(player.ManaRiders))
	}
}

// TestManaSpendRiderMadnessNonqualifyingCastNoScry confirms a madness cast that
// does not share the commander's creature type consumes the tagged mana on the
// cast without firing the rider.
func TestManaSpendRiderMadnessNonqualifyingCastNoScry(t *testing.T) {
	t.Parallel()
	g := elfCommanderGame(t)
	engine := NewEngine(nil)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	addRiders(g, game.Player1, mana.G, 1)

	cardID := addCardToHand(g, game.Player1, madnessCreature("Madness Goblin", types.Goblin, cost.Mana{cost.G}))
	if !discardCardFromHand(g, game.Player1, cardID) {
		t.Fatal("discardCardFromHand() = false, want true")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := firedSpendRiderCount(g); got != 0 {
		t.Fatalf("fired spend riders = %d, want 0 (nonqualifying madness cast)", got)
	}
	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders remaining = %d, want 0 (tagged mana consumed by cast)", len(player.ManaRiders))
	}
}

// TestManaSpendRiderCommanderStackBackFace covers MEDIUM #2 (current
// characteristics): when the commander is on the stack cast as its back face,
// the rider condition uses the back face's creature types, not the printed front
// face.
func TestManaSpendRiderCommanderStackBackFace(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := commanderDef("DFC Commander", color.Green)
	def.Subtypes = []types.Sub{types.Elf}
	def.Back = opt.Val(game.CardFace{
		Name:     "Risen Commander",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Zombie},
	})
	commanderID := commanderCardDef(g, game.Player1, def)
	g.Stack.Push(&game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   commanderID,
		Face:       game.FaceBack,
		Controller: game.Player1,
	})

	if !spellSatisfiesCommanderCreatureTypeRider(g, game.Player1, creatureSpellDef("Zombie", types.Zombie)) {
		t.Fatal("back-face Zombie commander should match a Zombie creature spell")
	}
	if spellSatisfiesCommanderCreatureTypeRider(g, game.Player1, elfCreatureDef()) {
		t.Fatal("printed front-face Elf must not match while commander is cast as its back face")
	}
}

// TestManaSpendRiderCommanderMergedComponent covers MEDIUM #2 (current
// characteristics): when the commander is a card merged beneath another by
// Mutate, the rider condition uses the merged permanent's effective creature
// types (its chosen top card), not the commander's printed types.
func TestManaSpendRiderCommanderMergedComponent(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	commanderDefn := commanderDef("Elf Commander", color.Green)
	commanderDefn.Subtypes = []types.Sub{types.Elf}
	commanderID := commanderCardDef(g, game.Player1, commanderDefn)

	topCardID := g.IDGen.Next()
	g.CardInstances[topCardID] = &game.CardInstance{
		ID:    topCardID,
		Def:   creatureSpellDef("Goblin Top", types.Goblin),
		Owner: game.Player1,
	}
	g.Battlefield = append(g.Battlefield, &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: topCardID,
		Owner:          game.Player1,
		Controller:     game.Player1,
		MergedCards:    []game.MergedCard{{CardInstanceID: commanderID, Face: game.FaceFront}},
	})

	if !spellSatisfiesCommanderCreatureTypeRider(g, game.Player1, creatureSpellDef("Goblin", types.Goblin)) {
		t.Fatal("commander merged under a Goblin top card should match a Goblin creature spell")
	}
	if spellSatisfiesCommanderCreatureTypeRider(g, game.Player1, elfCreatureDef()) {
		t.Fatal("commander's printed Elf type must not match once merged under a Goblin top card")
	}
}

// TestManaSpendRiderFiredTriggersUseAPNAPOrder covers MEDIUM #4: fired riders are
// queued and placed on the stack through the normal triggered-ability pass, so
// they follow APNAP ordering (active player's first/bottom) rather than the order
// in which they fired.
func TestManaSpendRiderFiredTriggersUseAPNAPOrder(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	// Queue the nonactive player's rider first to prove the pass reorders by
	// APNAP rather than preserving fire order.
	g.FiredManaSpendRiders = []game.ManaRiderInstance{
		{Controller: game.Player2, SourceID: 101, SourceObjectID: 201, Rider: scryRider()},
		{Controller: game.Player1, SourceID: 102, SourceObjectID: 202, Rider: scryRider()},
	}

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("putTriggeredAbilitiesOnStack() = false, want fired riders placed")
	}
	objects := g.Stack.Objects()
	if len(objects) != 2 {
		t.Fatalf("stack objects = %d, want 2", len(objects))
	}
	if objects[0].Controller != game.Player1 || objects[1].Controller != game.Player2 {
		t.Fatalf("stack controllers bottom-to-top = %v/%v, want APNAP Player1/Player2",
			objects[0].Controller, objects[1].Controller)
	}
	if len(g.FiredManaSpendRiders) != 0 {
		t.Fatalf("fired rider queue = %d, want 0 (drained)", len(g.FiredManaSpendRiders))
	}
}

// TestManaSpendRiderFiredTriggersSameControllerOrder covers MEDIUM #4 same-
// controller ordering: when one player controls multiple fired riders they are
// ordered by that player through the normal trigger-order choice.
func TestManaSpendRiderFiredTriggersSameControllerOrder(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	g.FiredManaSpendRiders = []game.ManaRiderInstance{
		{Controller: game.Player1, SourceID: 301, SourceObjectID: 401, Rider: scryRider()},
		{Controller: game.Player1, SourceID: 302, SourceObjectID: 402, Rider: scryRider()},
	}
	// The controller orders its two triggers as [second, first].
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1, 0}}}}

	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("putTriggeredAbilitiesOnStackWithChoices() = false, want fired riders placed")
	}
	objects := g.Stack.Objects()
	if len(objects) != 2 {
		t.Fatalf("stack objects = %d, want 2", len(objects))
	}
	// Chosen order [second, first] is placed in that order, so the second fired
	// rider (source 402) is on the bottom and the first (401) on top.
	if objects[0].SourceID != 402 || objects[1].SourceID != 401 {
		t.Fatalf("stack sources bottom-to-top = %v/%v, want chosen 402/401",
			objects[0].SourceID, objects[1].SourceID)
	}
}

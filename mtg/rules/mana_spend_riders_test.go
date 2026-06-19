package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
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

	if got := countTriggeredAbilities(g); got != 1 {
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

	if got := countTriggeredAbilities(g); got != 1 {
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

	if got := countTriggeredAbilities(g); got != 1 {
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

	if got := countTriggeredAbilities(g); got != 0 {
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

	if got := countTriggeredAbilities(g); got != 0 {
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

	if got := countTriggeredAbilities(g); got != 0 {
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

		if got := countTriggeredAbilities(g); got != 1 {
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

		if got := countTriggeredAbilities(g); got != 0 {
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

		if got := countTriggeredAbilities(g); got != 1 {
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

	if got := countTriggeredAbilities(g); got != 0 {
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

		if got := countTriggeredAbilities(g); got != 2 {
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

		if got := countTriggeredAbilities(g); got != 1 {
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
	consumeManaSpendRidersForPayment(g, game.Player1, before, map[mana.Unit]int{{Color: mana.G}: 1})

	if got := countTriggeredAbilities(g); got != 0 {
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
	consumeManaSpendRidersForPayment(g, game.Player1, beforeAbility, map[mana.Unit]int{{Color: mana.G}: 1})
	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders after ability payment = %d, want 0", len(player.ManaRiders))
	}

	// Later, plain green is produced and spent to cast a qualifying creature.
	player.ManaPool.Add(mana.G, 1)
	beforeSpell := poolUnitsSnapshot(player)
	player.ManaPool.Spend(mana.G, 1)
	resolveSpellCastManaSpendRiders(g, game.Player1, beforeSpell, map[mana.Unit]int{{Color: mana.G}: 1}, elfCreatureDef())

	if got := countTriggeredAbilities(g); got != 0 {
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
	if got := countTriggeredAbilities(g); got != 1 {
		t.Fatalf("scry triggers on stack = %d, want 1", got)
	}
	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders remaining = %d, want 0", len(player.ManaRiders))
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
	if got := countTriggeredAbilities(g); got != 0 {
		t.Fatalf("scry triggers on stack = %d, want 0 (ability payment)", got)
	}
	if len(player.ManaRiders) != 0 {
		t.Fatalf("riders remaining = %d, want 0 (consumed by ability payment)", len(player.ManaRiders))
	}
}

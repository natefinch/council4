package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
)

// Source review: partial prevention and prevented-amount riders on one-shot
// "a source of your choice" shields (#3807). Comprehensive Rules effective
// 2025-11-14 (local verified copy .cardwork/cr.txt):
//   - 615.5: some prevention effects include an additional effect that may
//     refer to the amount prevented; the prevention happens when the original
//     event would have, and the rest happens immediately afterward.
//   - 603.12: "when [something happens] this way" creates a reflexive
//     triggered ability, checked right after it is created.
//   - 615.12: unpreventable damage still has prevention effects applied, but
//     no damage is prevented and shields aren't reduced.
//   - 609.7b: if for any reason a shield prevents no damage, it isn't used up
//     (so half of 1 damage, rounded down, leaves Dark Sphere's shield).
//   - 608.2h: an effect needing information from an object uses its current
//     information, or its last known information once it has left its zone
//     (so "that source's controller" is checked as the rider happens).
// Scryfall Oracle (pinned corpus 6da9ced0...):
//   - Reverse Damage eaaf7c30 "The next time a source of your choice would
//     deal damage to you this turn, prevent that damage. You gain life equal
//     to the damage prevented this way." Cho-Arrim Alchemist 1e1d4d81 and
//     Intervention Pact 9f9882f5 share that rider.
//   - Bone Mask ddfc7dac "... prevent that damage. Exile cards from the top of
//     your library equal to the damage prevented this way." Ruling: with too
//     few cards, exile all of them; the damage is still prevented.
//   - Honorable Passage 66805334 "The next time a source of your choice would
//     deal damage to any target this turn, prevent that damage. If damage from
//     a red source is prevented this way, Honorable Passage deals that much
//     damage to the source's controller."
//   - Shadowbane 9e9ec33b "... deal damage to you and/or creatures you control
//     this turn, prevent that damage. If damage from a black source is
//     prevented this way, you gain that much life."
//   - New Way Forward 429368e6 "... prevent that damage. When damage is
//     prevented this way, New Way Forward deals that much damage to that
//     source's controller and you draw that many cards."
//   - Dark Sphere 397f53f7 "{T}, Sacrifice Dark Sphere: The next time a source
//     of your choice would deal damage to you this turn, prevent half that
//     damage, rounded down." Ruling (2013-04-15): multiple Dark Sphere effects
//     apply one after another.

func riderShield(riders ...game.PreventedDamageRider) game.PreventDamage {
	prim := oneShotShield(nil, nil)
	prim.PreventedRiders = riders
	return prim
}

func lifeTotals(g *game.Game) (int, int) {
	return g.Players[game.Player1].Life, g.Players[game.Player2].Life
}

// TestReverseDamageGainsLifeAfterPreventingBolt plays Reverse Damage against a
// red Bolt on the stack: the Bolt's 3 damage is prevented and Player2 then
// gains 3 life (CR 615.5), so the prevention event precedes the life gain.
func TestReverseDamageGainsLifeAfterPreventingBolt(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	bolt := boltOnStack(g, game.Player1, game.PlayerTarget(game.Player2))
	reverse := addPreventionSourceSpell(g, game.Player2, "Reverse Damage")
	createOneShotShield(t, engine, g, reverse, riderShield(game.PreventedDamageRiderControllerGainsLife), sourceChoiceAgent(t, bolt, nil))

	engine.resolveTopOfStack(g, &TurnLog{})

	if p1, p2 := lifeTotals(g); p1 != 40 || p2 != 43 {
		t.Fatalf("life = %d/%d, want 40/43", p1, p2)
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("shields = %+v, want used up", g.PreventionShields)
	}
	prevented, gained := -1, -1
	for i, event := range g.Events {
		if event.Kind == game.EventDamagePrevented && event.Player == game.Player2 && event.Amount == 3 {
			prevented = i
		}
		if event.Kind == game.EventLifeGained && event.Player == game.Player2 && event.Amount == 3 {
			gained = i
		}
	}
	if prevented < 0 || gained < prevented {
		t.Fatalf("prevented event %d, gained event %d; want the gain after the prevention", prevented, gained)
	}
}

// TestChoArrimAlchemistGainsOnlyChosenCombatDamage plays Cho-Arrim Alchemist's
// shield against two attackers: only the chosen 4-power attacker's damage is
// prevented and only that amount is gained.
func TestChoArrimAlchemistGainsOnlyChosenCombatDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	chosen := addCombatCreaturePermanentWithPower(g, game.Player1, 4)
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	_, ability := protectionShieldSource(g, game.Player2)
	agent := sourceChoiceAgent(t, chosen.CardInstanceID, nil)
	createOneShotShield(t, engine, g, ability, riderShield(game.PreventedDamageRiderControllerGainsLife), agent)

	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: chosen.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: other.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}}
	agents := [game.NumPlayers]PlayerAgent{game.Player2: agent}
	engine.setReplacementChoiceContext(g, agents, &TurnLog{})
	defer g.ClearChoiceContext()
	engine.resolveCombatDamage(g, &TurnLog{})

	if _, p2 := lifeTotals(g); p2 != 42 {
		t.Fatalf("defender life = %d, want 40-2+4 = 42", p2)
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("shields = %+v, want used up", g.PreventionShields)
	}
}

// TestBoneMaskExilesAvailableLibraryCardsForPreventedDamage plays Bone Mask:
// 4 damage is prevented and its controller exiles the top cards of their
// library equal to that amount, exiling all 2 when the library is short.
func TestBoneMaskExilesAvailableLibraryCardsForPreventedDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addColoredCreaturePermanent(g, game.Player1, 4, color.Black)
	first := addLibraryCard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "First"}})
	second := addLibraryCard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})
	_, ability := protectionShieldSource(g, game.Player2)
	createOneShotShield(t, engine, g, ability, riderShield(game.PreventedDamageRiderControllerExilesLibraryTop),
		sourceChoiceAgent(t, source.CardInstanceID, nil))

	if dealt := dealDamageFrom(g, source, game.Player2, 4); dealt != 0 {
		t.Fatalf("dealt = %d, want 0", dealt)
	}
	player := g.Players[game.Player2]
	if player.Life != 40 || player.Library.Size() != 0 || !player.Exile.Contains(first) || !player.Exile.Contains(second) {
		t.Fatalf("life %d, library %v, exile %v; want 40 with both library cards exiled",
			player.Life, player.Library.All(), player.Exile.All())
	}
}

// TestHonorablePassageReflectsOnlyRedSourceDamage plays Honorable Passage
// targeting Player2: a red Bolt's prevented 3 damage is dealt back to its
// controller by Honorable Passage, while a non-red source's damage is
// prevented with nothing dealt back.
func TestHonorablePassageReflectsOnlyRedSourceDamage(t *testing.T) {
	passage := func() game.PreventDamage {
		return game.PreventDamage{
			AnyTarget: game.AnyTargetDamageRecipient(0), All: true, OneShot: true,
			PreventedRiders:   []game.PreventedDamageRider{game.PreventedDamageRiderDamageSourceController},
			RiderSourceColors: []color.Color{color.Red},
		}
	}
	t.Run("red", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		bolt := boltOnStack(g, game.Player1, game.PlayerTarget(game.Player2))
		spell := addPreventionSourceSpell(g, game.Player2, "Honorable Passage")
		spell.Targets = []game.Target{game.PlayerTarget(game.Player2)}
		createOneShotShield(t, engine, g, spell, passage(), sourceChoiceAgent(t, bolt, nil))

		engine.resolveTopOfStack(g, &TurnLog{})

		if p1, p2 := lifeTotals(g); p1 != 37 || p2 != 40 {
			t.Fatalf("life = %d/%d, want 37/40", p1, p2)
		}
		assertEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
			return event.Player == game.Player1 && event.SourceID == spell.SourceID && event.Amount == 3
		})
	})
	t.Run("not red", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addColoredCreaturePermanent(g, game.Player1, 3, color.Green)
		spell := addPreventionSourceSpell(g, game.Player2, "Honorable Passage")
		spell.Targets = []game.Target{game.PlayerTarget(game.Player2)}
		createOneShotShield(t, engine, g, spell, passage(), sourceChoiceAgent(t, source.CardInstanceID, nil))

		if dealt := dealDamageFrom(g, source, game.Player2, 3); dealt != 0 {
			t.Fatalf("dealt = %d, want 0", dealt)
		}
		if p1, p2 := lifeTotals(g); p1 != 40 || p2 != 40 || len(g.PreventionShields) != 0 {
			t.Fatalf("life = %d/%d, shields %d; want 40/40 with the shield used up", p1, p2, len(g.PreventionShields))
		}
	})
}

// TestShadowbaneGainsOnceForPlayerAndCreatureDamageFromBlackSource plays
// Shadowbane against a black source dealing 1 damage to each creature and each
// player: the damage to Player2 and their creature is prevented in the one
// event and Player2 gains the 2 prevented life once.
func TestShadowbaneGainsOnceForPlayerAndCreatureDamageFromBlackSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addColoredCreaturePermanent(g, game.Player1, 3, color.Black)
	shielded := addColoredCreaturePermanent(g, game.Player2, 3, color.White)
	_, ability := protectionShieldSource(g, game.Player2)
	prim := riderShield(game.PreventedDamageRiderControllerGainsLife)
	prim.AndControlledCreatures = true
	prim.RiderSourceColors = []color.Color{color.Black}
	createOneShotShield(t, engine, g, ability, prim, sourceChoiceAgent(t, source.CardInstanceID, nil))

	pestilence := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, SourceID: source.ObjectID, Controller: game.Player1}
	resolveInstruction(engine, g, pestilence, eachCreatureAndPlayer(game.Fixed(1)), nil)

	if p1, p2 := lifeTotals(g); p1 != 39 || p2 != 42 {
		t.Fatalf("life = %d/%d, want 39/42", p1, p2)
	}
	if shielded.MarkedDamage != 0 || source.MarkedDamage != 1 {
		t.Fatalf("marked damage = shielded %d, source %d; want 0 and 1", shielded.MarkedDamage, source.MarkedDamage)
	}
	gains := 0
	for _, event := range g.Events {
		if event.Kind == game.EventLifeGained && event.Player == game.Player2 {
			gains++
		}
	}
	if gains != 1 {
		t.Fatalf("life gain events = %d, want one aggregated gain", gains)
	}
}

// TestNewWayForwardReflexiveTriggerDamagesAndDraws plays New Way Forward: the
// chosen source's 5 damage is prevented, then a reflexive trigger (CR 603.12)
// goes on the stack and deals 5 damage to the source's controller from New
// Way Forward and draws its controller 5 cards. "That source's controller" is
// determined as the trigger resolves: after a control change in response it
// is the new controller, and after the source leaves the battlefield it is
// the source's last-known controller.
func TestNewWayForwardReflexiveTriggerDamagesAndDraws(t *testing.T) {
	for _, test := range []struct {
		name             string
		respond          func(g *game.Game, source *game.Permanent)
		wantP1, wantP3   int
		sourceController game.PlayerID
	}{
		{name: "same controller", respond: func(*game.Game, *game.Permanent) {}, wantP1: 35, wantP3: 40},
		{name: "control changed in response", respond: func(_ *game.Game, source *game.Permanent) {
			source.Controller = game.Player3
		}, wantP1: 40, wantP3: 35},
		{name: "control changed then source left", respond: func(g *game.Game, source *game.Permanent) {
			source.Controller = game.Player3
			sacrificePermanent(g, source)
		}, wantP1: 40, wantP3: 35},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addColoredCreaturePermanent(g, game.Player1, 5, color.Red)
			for range 5 {
				addLibraryCard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Library Card"}})
			}
			spell := addPreventionSourceSpell(g, game.Player2, "New Way Forward")
			prim := riderShield(game.PreventedDamageRiderDamageSourceController, game.PreventedDamageRiderControllerDraws)
			prim.RidersReflexive = true
			createOneShotShield(t, engine, g, spell, prim, sourceChoiceAgent(t, source.CardInstanceID, nil))

			if dealt := dealDamageFrom(g, source, game.Player2, 5); dealt != 0 {
				t.Fatalf("dealt = %d, want 0", dealt)
			}
			if p1, p2 := lifeTotals(g); p1 != 40 || p2 != 40 || len(g.PendingReflexiveTriggers) != 1 {
				t.Fatalf("life = %d/%d, pending reflexive = %d; want 40/40 with one pending trigger",
					p1, p2, len(g.PendingReflexiveTriggers))
			}
			test.respond(g, source)
			resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})

			p1, p2 := lifeTotals(g)
			if p3 := g.Players[game.Player3].Life; p1 != test.wantP1 || p2 != 40 || p3 != test.wantP3 {
				t.Fatalf("life = %d/%d/%d, want %d/40/%d", p1, p2, p3, test.wantP1, test.wantP3)
			}
			if hand := g.Players[game.Player2].Hand.Size(); hand != 5 {
				t.Fatalf("hand = %d, want 5 drawn", hand)
			}
			assertEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
				return event.SourceID == spell.SourceID && event.Amount == 5
			})
		})
	}
}

// TestDarkSpherePreventsHalfRoundedDown plays Dark Sphere: half of the chosen
// source's damage is prevented, rounded down, and two Dark Spheres apply one
// after the other. A shield whose half rounds down to nothing prevents no
// damage, so it is not used up and halves the source's next damage instead
// (CR 609.7b).
func TestDarkSpherePreventsHalfRoundedDown(t *testing.T) {
	sphere := func() game.PreventDamage {
		prim := oneShotShield(nil, nil)
		prim.PreventHalfRoundedDown = true
		return prim
	}
	for _, test := range []struct {
		name             string
		spheres          int
		damage           []int
		want             []int
		remainingShields int
	}{
		{name: "one sphere", spheres: 1, damage: []int{5, 5}, want: []int{3, 5}},
		{name: "two spheres", spheres: 2, damage: []int{5, 5}, want: []int{2, 5}},
		{name: "one damage keeps the shield", spheres: 1, damage: []int{1, 5, 5}, want: []int{1, 3, 5}},
		{name: "two spheres on one damage", spheres: 2, damage: []int{1}, want: []int{1}, remainingShields: 2},
		{name: "second sphere rounds to nothing", spheres: 2, damage: []int{2, 4}, want: []int{1, 2}},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addColoredCreaturePermanent(g, game.Player1, 5, color.Red)
			for range test.spheres {
				_, ability := protectionShieldSource(g, game.Player2)
				createOneShotShield(t, engine, g, ability, sphere(), sourceChoiceAgent(t, source.CardInstanceID, nil))
			}
			life := 40
			for i, damage := range test.damage {
				if dealt := dealDamageFrom(g, source, game.Player2, damage); dealt != test.want[i] {
					t.Fatalf("damage %d: dealt = %d, want %d", i, dealt, test.want[i])
				}
				life -= test.want[i]
				if _, p2 := lifeTotals(g); p2 != life {
					t.Fatalf("damage %d: life = %d, want %d", i, p2, life)
				}
			}
			if len(g.PreventionShields) != test.remainingShields {
				t.Fatalf("shields = %+v, want %d remaining", g.PreventionShields, test.remainingShields)
			}
		})
	}
}

// TestPreventionRiderSkipsUnpreventableDamage proves CR 615.12 for a rider
// shield: unpreventable damage is dealt in full, no damage was prevented so
// no life is gained, and the shield remains for the source's next damage.
func TestPreventionRiderSkipsUnpreventableDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addColoredCreaturePermanent(g, game.Player1, 3, color.Red)
	_, ability := protectionShieldSource(g, game.Player2)
	createOneShotShield(t, engine, g, ability, riderShield(game.PreventedDamageRiderControllerGainsLife),
		sourceChoiceAgent(t, source.CardInstanceID, nil))

	burn := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, SourceID: source.ObjectID, Controller: game.Player1,
		Targets: []game.Target{game.PlayerTarget(game.Player2)},
	}
	resolveInstruction(engine, g, burn, game.Damage{
		Amount: game.Fixed(3), Recipient: game.PlayerDamageRecipient(game.TargetPlayerReference(0)), CantBePrevented: true,
	}, nil)

	if _, p2 := lifeTotals(g); p2 != 37 || len(g.PreventionShields) != 1 {
		t.Fatalf("life = %d, shields %d; want 37 with the shield remaining", p2, len(g.PreventionShields))
	}
	if dealt := dealDamageFrom(g, source, game.Player2, 2); dealt != 0 {
		t.Fatalf("later dealt = %d, want 0", dealt)
	}
	if _, p2 := lifeTotals(g); p2 != 39 {
		t.Fatalf("life = %d, want 37+2 = 39 after the later prevention", p2)
	}
}

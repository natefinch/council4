package rules

import (
	"strconv"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func TestStunUntapSourceCostRemainsPaid(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalUntap, Text: "{Q}"}},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
		}}}.Ability(),
	}))
	source.Tapped = true
	source.Counters.Add(counter.Stun, 3)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	life := g.Players[game.Player1].Life
	for attempt := 1; attempt <= 4; attempt++ {
		if !containsAction(engine.legalActions(g, game.Player1), act) || !engine.applyAction(g, game.Player1, act) {
			t.Fatalf("untap-cost activation %d was not paid", attempt)
		}
		if source.Tapped != (attempt < 4) || source.Counters.Get(counter.Stun) != max(0, 3-attempt) {
			t.Fatalf("cost attempt %d = tapped %v, stun %d", attempt, source.Tapped, source.Counters.Get(counter.Stun))
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if g.Players[game.Player1].Life != life+attempt {
			t.Fatal("stun replacement prevented a paid ability from resolving")
		}
	}
	if containsAction(engine.legalActions(g, game.Player1), act) || engine.applyAction(g, game.Player1, act) {
		t.Fatal("untapped source paid an untap cost")
	}
	events := stunUntappedEvents(g)
	if len(events) != 1 || events[0].PermanentID != source.ObjectID {
		t.Fatalf("cost untap events = %+v, want only the fourth actual untap", events)
	}
}

func TestStunUntapCostEligibility(t *testing.T) {
	for _, tc := range []struct {
		name    string
		tapped  bool
		sick    bool
		haste   bool
		allowed bool
	}{
		{"untapped", false, false, false, false},
		{"sick", true, true, false, false},
		{"as-though-haste", true, true, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
				AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalUntap, Text: "{Q}"}},
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
				}}}.Ability(),
			}))
			source.Tapped, source.SummoningSick = tc.tapped, tc.sick
			source.Counters.Add(counter.Stun, 2)
			if tc.haste {
				activateAsThoughHasteStaticPermanent(g, game.Player1)
			}
			g.Turn.Phase = game.PhasePrecombatMain
			g.Turn.Step = game.StepNone
			act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
			if containsAction(engine.legalActions(g, game.Player1), act) != tc.allowed ||
				engine.applyAction(g, game.Player1, act) != tc.allowed {
				t.Fatalf("activation eligibility differs from %v", tc.allowed)
			}
			wantStun := 2
			if tc.allowed {
				wantStun--
			}
			if source.Tapped != tc.tapped || source.Counters.Get(counter.Stun) != wantStun || len(stunUntappedEvents(g)) != 0 {
				t.Fatal("cost eligibility changed counters or tapped state incorrectly")
			}
		})
	}
}

func TestStunUntapAutomaticManaCost(t *testing.T) {
	for _, stun := range []int{0, 1, 2} {
		t.Run(strconv.Itoa(stun), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			body := game.TapManaAbility(mana.G)
			body.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalUntap, Text: "{Q}"}}
			source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Untap Mana Source", Types: []types.Card{types.Artifact},
				ManaAbilities: []game.ManaAbility{body}, StaticAbilities: []game.StaticAbility{game.DoesntUntapStaticBody},
			}})
			source.Tapped = true
			source.Counters.Add(counter.Stun, stun)
			manaCost := cost.Mana{cost.G}
			if !payTestGenericCost(g, game.Player1, &manaCost) {
				t.Fatal("stun-replaced untap mana cost failed payment")
			}
			if source.Tapped != (stun > 0) || source.Counters.Get(counter.Stun) != max(0, stun-1) {
				t.Fatalf("mana cost state = tapped %v, stun %d", source.Tapped, source.Counters.Get(counter.Stun))
			}
			wantEvents := 0
			if stun == 0 {
				wantEvents = 1
			}
			if len(stunUntappedEvents(g)) != wantEvents {
				t.Fatal("mana cost emitted an untap event for a replaced action")
			}
			if g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatal("paid cost did not spend its produced mana")
			}
		})
	}
}

func TestStunUntapTapAndSacrificeCostsDoNotRemoveStun(t *testing.T) {
	for _, additional := range []cost.Additional{cost.T, {Kind: cost.AdditionalSacrificeSource}} {
		t.Run(strconv.Itoa(int(additional.Kind)), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
				AdditionalCosts: []cost.Additional{additional},
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
				}}}.Ability(),
			}))
			source.Counters.Add(counter.Stun, 2)
			g.Turn.Phase = game.PhasePrecombatMain
			g.Turn.Step = game.StepNone
			if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
				t.Fatal("tap/sacrifice cost was not paid")
			}
			if source.Counters.Get(counter.Stun) != 2 || len(stunUntappedEvents(g)) != 0 {
				t.Fatal("tap/sacrifice cost consumed stun or emitted an untap")
			}
		})
	}
}

func TestStunUntapCostWithCounterRemoval(t *testing.T) {
	for _, removal := range []cost.Additional{
		{Kind: cost.AdditionalRemoveCounter, CounterKind: counter.Stun, Amount: 1},
		{Kind: cost.AdditionalRemoveCounter, AnyCounterKind: true, Amount: 1},
		{Kind: cost.AdditionalRemoveCounterAmong, CounterKind: counter.Stun, Amount: 1},
		{Kind: cost.AdditionalRemoveCounterAmong, AnyCounterKind: true, Amount: 1},
	} {
		for _, stun := range []int{1, 2} {
			t.Run(strconv.Itoa(int(removal.Kind))+"/"+strconv.FormatBool(removal.AnyCounterKind)+"/"+strconv.Itoa(stun), func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := NewEngine(nil)
				source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
					AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalUntap, Text: "{Q}"}, removal},
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
					}}}.Ability(),
				}))
				source.Tapped = true
				source.Counters.Add(counter.Stun, stun)
				g.Turn.Phase = game.PhasePrecombatMain
				g.Turn.Step = game.StepNone
				act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
				if !containsAction(engine.legalActions(g, game.Player1), act) || !engine.applyAction(g, game.Player1, act) {
					t.Fatal("payable counter removal and untap cost was rejected")
				}
				if source.Tapped != (stun > 1) || source.Counters.Get(counter.Stun) != 0 {
					t.Fatalf("combined cost state = tapped %v, stun %d; want tapped %v, stun 0", source.Tapped, source.Counters.Get(counter.Stun), stun > 1)
				}
			})
		}
	}

}

func TestStunUntapManaPreservesReservedCounters(t *testing.T) {
	for _, generic := range []bool{false, true} {
		for _, anyKind := range []bool{false, true} {
			for _, tc := range []struct {
				name      string
				stun      int
				alternate bool
				tapped    bool
				paid      bool
			}{
				{"conflict", 1, false, true, false},
				{"enough-counters", 2, false, true, true},
				{"alternate-mana", 1, true, true, true},
				{"tap-mana", 1, false, false, true},
			} {
				t.Run(strconv.FormatBool(generic)+"/"+strconv.FormatBool(anyKind)+"/"+tc.name, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					body := game.TapManaAbility(mana.G)
					if tc.tapped {
						body.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalUntap, Text: "{Q}"}}
					}
					source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
						Name: "Shared Counter Mana Source", Types: []types.Card{types.Artifact},
						ManaAbilities: []game.ManaAbility{body},
					}})
					source.Tapped = tc.tapped
					source.Counters.Add(counter.Stun, tc.stun)
					var alternate *game.Permanent
					if tc.alternate {
						alternate = addBasicLandPermanent(g, game.Player1, types.Forest)
					}
					removal := cost.Additional{Kind: cost.AdditionalRemoveCounter, Amount: 1, CounterKind: counter.Stun, AnyCounterKind: anyKind}
					manaCost := cost.Mana{cost.G}
					start := len(g.Events)
					paid := false
					if generic {
						paid = paymentOrch.payGenericCost(g, payment.GenericRequest{
							PlayerID: game.Player1, Source: source, Cost: &manaCost, AdditionalCosts: []cost.Additional{removal},
						})
					} else {
						_, paid = paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
							PlayerID: game.Player1, Source: source, ManaCost: opt.Val(manaCost), AdditionalCosts: []cost.Additional{removal},
						})
					}
					if paid != tc.paid {
						t.Fatalf("payment = %v, want %v", paid, tc.paid)
					}
					if !paid {
						if source.Tapped != tc.tapped || source.Counters.Get(counter.Stun) != tc.stun ||
							len(g.Events) != start || g.Players[game.Player1].ManaPool.Total() != 0 {
							t.Fatal("unpayable combined plan mutated state before rejection")
						}
						return
					}
					if !source.Tapped || source.Counters.Get(counter.Stun) != 0 || len(stunUntappedEvents(g)) != 0 ||
						g.Players[game.Player1].ManaPool.Total() != 0 {
						t.Fatal("combined payment did not preserve the reserved removal or mana accounting")
					}
					if alternate != nil && !alternate.Tapped {
						t.Fatal("payment did not select the safe alternate mana source")
					}
				})
			}
		}
	}
}

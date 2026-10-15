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

type strikeHistoryPolicy struct {
	simultaneousCombatPolicy

	source    id.ID
	activated bool
}

func (p *strikeHistoryPolicy) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if obs.Turn.Step == game.StepFirstStrikeDamage && !p.activated {
		for _, candidate := range legal {
			if activated, ok := candidate.ActivateAbilityPayload(); ok && activated.SourceID == p.source {
				p.activated = true
				return candidate
			}
		}
	}

	return p.simultaneousCombatPolicy.ChooseAction(obs, legal)
}

func TestCombatStrikeHistoryKeywordChangesUseOriginalEligibility(t *testing.T) {
	for _, test := range []struct {
		name     string
		original game.Keyword
		add      game.Keyword
		remove   game.Keyword
		want     int
	}{
		{"ordinary gains first strike", game.KeywordNone, game.FirstStrike, game.KeywordNone, 3},
		{"ordinary gains double strike", game.KeywordNone, game.DoubleStrike, game.KeywordNone, 3},
		{"first strike gains double strike", game.FirstStrike, game.DoubleStrike, game.KeywordNone, 6},
		{"first strike loses keyword", game.FirstStrike, game.KeywordNone, game.FirstStrike, 3},
		{"double strike loses keyword", game.DoubleStrike, game.KeywordNone, game.DoubleStrike, 3},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3, test.original)
			other := addCombatCreaturePermanentWithPower(g, game.Player1, 0, game.FirstStrike)
			g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
				{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
				{Attacker: other.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			}}
			engine := NewEngine(nil)
			engine.resolveCombatDamagePass(g, firstStrikeCombatDamage, &TurnLog{})
			g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
				Layer: game.LayerAbility, AffectedObjectID: attacker.ObjectID,
				AddKeywords: []game.Keyword{test.add}, RemoveKeywords: []game.Keyword{test.remove},
			})
			engine.resolveCombatDamagePass(g, normalCombatDamage, &TurnLog{})
			if dealt := 40 - g.Players[game.Player2].Life; dealt != test.want {
				t.Fatalf("dealt %d, want %d", dealt, test.want)
			}
		})
	}
}

func TestCombatStrikeHistoryZeroDamageAndCloneIndependence(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 0, game.DoubleStrike)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}}
	engine := NewEngine(nil)
	engine.resolveCombatDamagePass(g, firstStrikeCombatDamage, &TurnLog{})
	branch := g.Clone()
	branch.ContinuousEffects = append(branch.ContinuousEffects,
		game.ContinuousEffect{Layer: game.LayerPowerToughnessSet, AffectedObjectID: attacker.ObjectID, SetPower: opt.Val(game.PT{Value: 3})},
		game.ContinuousEffect{Layer: game.LayerAbility, AffectedObjectID: attacker.ObjectID, RemoveKeywords: []game.Keyword{game.DoubleStrike}},
	)
	engine.resolveCombatDamagePass(branch, normalCombatDamage, &TurnLog{})
	if branch.Players[game.Player2].Life != 40 {
		t.Fatal("zero first-step damage erased the original double-strike eligibility")
	}
	branch.Combat.FirstStrikeEligible[attacker.ObjectID] = false
	if !g.Combat.FirstStrikeEligible[attacker.ObjectID] {
		t.Fatal("cloned eligibility history shared its map")
	}
}

func TestCombatStrikeHistoryBlockerLosesDoubleStrike(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 8)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 3, game.DoubleStrike)
	g.Combat = blockedCombat(attacker, blocker)
	engine := NewEngine(nil)
	engine.resolveCombatDamagePass(g, firstStrikeCombatDamage, &TurnLog{})
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		Layer: game.LayerAbility, AffectedObjectID: blocker.ObjectID, RemoveKeywords: []game.Keyword{game.DoubleStrike},
	})
	engine.resolveCombatDamagePass(g, normalCombatDamage, &TurnLog{})
	if attacker.MarkedDamage != 3 || blocker.MarkedDamage != 8 {
		t.Fatal("blocker's second-step assignment ignored its first-step history")
	}
}
func TestCombatStrikeHistoryLosingKeywordThroughInterveningPriority(t *testing.T) {
	for _, keyword := range []game.Keyword{game.FirstStrike, game.DoubleStrike} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		def := &game.CardDef{CardFace: game.CardFace{
			Name: "Strike history attacker", Types: []types.Card{types.Creature},
			Power: opt.Val(game.PT{Value: 3}), Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{{KeywordAbilities: game.SimpleKeywords(keyword)}},
			ActivatedAbilities: []game.ActivatedAbility{{Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.ApplyContinuous{
					Object: opt.Val(game.SourcePermanentReference()),
					ContinuousEffects: []game.ContinuousEffect{{
						Layer: game.LayerAbility, RemoveKeywords: []game.Keyword{keyword},
					}},
				},
			}}}.Ability()}},
		}}
		attacker := addCombatPermanent(g, game.Player1, def)
		g.Turn.ActivePlayer = game.Player1
		g.Turn.Phase = game.PhaseCombat
		g.Turn.Step = game.StepDeclareAttackers
		g.Combat = &game.CombatState{}
		probe := &strikeHistoryPolicy{
			simultaneousCombatPolicy: simultaneousCombatPolicy{attacker: attacker.ObjectID, defender: game.Player2},
			source:                   attacker.ObjectID,
		}
		policies := simPassPolicies()
		policies[game.Player1] = probe
		declare := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2},
		}}}
		resolved, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, declare, policies)
		if !ok || !probe.activated {
			t.Fatal("fixture did not resolve a real keyword-removal ability between damage steps")
		}
		if resolved.Players[game.Player2].Life != 37 {
			t.Fatalf("damage after losing %v = %d, want only the first assignment", keyword, 40-resolved.Players[game.Player2].Life)
		}
		probe.activated = false
		live := g.Clone()
		NewEngine(nil).runCombatPhase(live, policies, &TurnLog{})
		if live.Players[game.Player2].Life != 37 || g.Players[game.Player2].Life != 40 {
			t.Fatal("live strike history or Simulator isolation disagreed")
		}
	}
}

func TestCombatStrikeHistoryPreventionDoesNotEraseEligibility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.DoubleStrike)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}}
	g.PreventionShields = []game.PreventionShield{{
		ID: g.IDGen.Next(), Player: game.Player2, Amount: 3, CombatOnly: true,
	}}
	engine := NewEngine(nil)
	engine.resolveCombatDamagePass(g, firstStrikeCombatDamage, &TurnLog{})
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		Layer: game.LayerAbility, AffectedObjectID: attacker.ObjectID, RemoveKeywords: []game.Keyword{game.DoubleStrike},
	})
	engine.resolveCombatDamagePass(g, normalCombatDamage, &TurnLog{})
	if g.Players[game.Player2].Life != 40 || !g.Combat.FirstStrikeEligible[attacker.ObjectID] {
		t.Fatal("preventing all first-step damage erased its original eligibility")
	}
}

func TestCombatStrikeHistoryReentryUsesFreshObjectIdentity(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.FirstStrike)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}}
	engine := NewEngine(nil)
	engine.resolveCombatDamagePass(g, firstStrikeCombatDamage, &TurnLog{})
	if !movePermanentToZone(g, attacker, zone.Graveyard) {
		t.Fatal("could not remove the original combatant")
	}
	card, _ := g.GetCardInstance(attacker.CardInstanceID)
	g.Players[game.Player1].Graveyard.Remove(card.ID)
	returned, entered := createCardPermanent(g, card, game.Player1, zone.Graveyard)
	if !entered || returned.ObjectID == attacker.ObjectID {
		t.Fatal("reentry did not create a fresh object")
	}
	g.Combat.Attackers = append(g.Combat.Attackers, game.AttackDeclaration{
		Attacker: returned.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	})
	engine.resolveCombatDamagePass(g, normalCombatDamage, &TurnLog{})
	if g.Players[game.Player2].Life != 34 || g.Combat.FirstStrikeEligible[returned.ObjectID] {
		t.Fatal("new combatant inherited an earlier incarnation's first-step history")
	}
}

func TestCombatStrikeHistoryFreshCombatAndNoFirstStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	declare := []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}}
	g.Combat = &game.CombatState{Attackers: declare}
	engine := NewEngine(nil)
	engine.resolveCombatDamagePass(g, normalCombatDamage, &TurnLog{})
	if g.Players[game.Player2].Life != 37 || g.Combat.FirstStrikeDamageStarted {
		t.Fatal("ordinary combat acquired a fictitious first-strike step")
	}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		Layer: game.LayerAbility, AffectedObjectID: attacker.ObjectID, AddKeywords: []game.Keyword{game.FirstStrike},
	})
	g.Combat = &game.CombatState{Attackers: declare}
	engine.resolveCombatDamagePass(g, firstStrikeCombatDamage, &TurnLog{})
	engine.resolveCombatDamagePass(g, normalCombatDamage, &TurnLog{})
	if g.Players[game.Player2].Life != 34 || !g.Combat.FirstStrikeEligible[attacker.ObjectID] {
		t.Fatal("a fresh combat reused or lost the prior combat's participation state")
	}
}

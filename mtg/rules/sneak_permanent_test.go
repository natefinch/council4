package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine proving fixtures for Sneak on permanent spells (CR
// 702.190b), not the production registry definitions, from the pinned corpus
// SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Foot Ninjas: oracle 306652a8-9a9c-4b33-96d7-b59234d148e9, printing
// abb1ab9c-b067-4b75-8e5b-a893b7948df1. Shredder, Unrelenting: oracle
// 59d96fc0-6580-43c2-8c6a-dcf983644561, printing
// 88a48867-5c65-483c-92b0-f70c53ea2a9e. Leonardo, Leader in Blue: oracle
// 348c7e22-2179-41b8-a5d2-8dd488e7e66d, printing
// d6eaae35-d513-43d8-be2d-b97c15e25937. Turncoat Kunoichi: oracle
// 0f35cf85-6783-4d69-b5e6-a81f7c1f21e7, printing
// 3ae61336-a4ee-40cd-9a18-392d96b873a4. Splinter, Hamato Yoshi: oracle
// 41fbdaa2-7a7a-44d5-a8f4-77624350eedd, printing
// 9ea9d072-aa98-405e-a475-26f93cc37e53. Read the complete Scryfall rulings
// responses (all 2026-01-27): a sneaked permanent enters attacking the same
// player, planeswalker or battle as the returned creature; it was never
// declared as an attacker, so "whenever ~ attacks" does not trigger; Sneak is
// usable only in your declare blockers step; Turncoat Kunoichi exiles nothing
// if it was not sneaked and leaves before its trigger resolves. Verified
// official CR (effective 2026-11-06 copy) read: 702.190a-b, 506.3a-c/f and
// 506.4 (what can be attacking; a returned creature leaves combat), 508.4,
// 508.4a/c/d (put onto the battlefield attacking: never declared, a gone
// planeswalker means not attacking, restrictions don't apply, unblocked in the
// declare blockers step), 509.1h, 603.4 (intervening if), 610.3 (until leaves),
// 707.10 (a copy copies the paid alternative cost), 903.8 (commander tax).
const (
	footNinjasText       = "Sneak {3}{W/B} (You may cast this spell for {3}{W/B} if you also return an unblocked attacker you control to hand during the declare blockers step. It enters tapped and attacking.)\nWhen this creature enters, you gain 3 life."
	shredderText         = "Sneak {3}{B} (You may cast this spell for {3}{B} if you also return an unblocked attacker you control to hand during the declare blockers step. He enters tapped and attacking.)\nDeathtouch\nWhenever Shredder enters or attacks, another target creature you control gains deathtouch until end of turn."
	leonardoText         = "Sneak {3}{W}{W} (You may cast this spell for {3}{W}{W} if you also return an unblocked attacker you control to hand during the declare blockers step. He enters tapped and attacking.)\nWhen Leonardo enters, if his sneak cost was paid, creatures you control get +2/+0 until end of turn.\n{1}{W}: Leonardo gains first strike until end of turn."
	turncoatKunoichiText = "Sneak {2}{W}{B} (You may cast this spell for {2}{W}{B} if you also return an unblocked attacker you control to hand during the declare blockers step. It enters tapped and attacking.)\nWhen this creature enters, choose target creature an opponent controls. Exile that creature until this creature leaves the battlefield. If this creature's sneak cost was paid, instead exile the chosen creature."
	splinterText         = "Sneak {B} (You may cast this spell for {B} if you also return an unblocked attacker you control to hand during the declare blockers step. He enters tapped and attacking.)\nMenace (This creature can't be blocked except by two or more creatures.)\nOther Ninjas you control get +1/+1."
)

var (
	turncoatKunoichiProgram = game.ProgramReference{ID: "test/turncoat-kunoichi/trigger/0", Version: "1"}
	copyTargetSpellProgram  = game.ProgramReference{ID: "test/copy-target-spell/spell/0", Version: "1"}
)

func sneakCreatureFixture(name, text string, manaCost, sneak cost.Mana, power, toughness int, subtypes ...types.Sub) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: name, ManaCost: opt.Val(manaCost), OracleText: text,
		Types: []types.Card{types.Creature}, Subtypes: subtypes,
		Power: opt.Val(game.PT{Value: power}), Toughness: opt.Val(game.PT{Value: toughness}),
		AlternativeCosts: []cost.Alternative{cost.SneakAlternative(sneak)},
	}}
}

func footNinjasFixture() *game.CardDef {
	wb := cost.HybridMana(mana.W, mana.B)
	def := sneakCreatureFixture("Foot Ninjas", footNinjasText,
		cost.Mana{cost.O(4), wb, wb}, cost.Mana{cost.O(3), wb}, 5, 5, types.Human, types.Ninja)
	def.ColorIdentity = color.NewIdentity(color.White, color.Black)
	def.Colors = []color.Color{color.White, color.Black}
	def.TriggeredAbilities = []game.TriggeredAbility{{
		Text: "When this creature enters, you gain 3 life.",
		Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
			Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
		}},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(3), Player: game.ControllerReference()},
		}}}.Ability(),
	}}
	return def
}

func shredderFixture() *game.CardDef {
	def := sneakCreatureFixture("Shredder, Unrelenting", shredderText,
		cost.Mana{cost.O(4), cost.B}, cost.Mana{cost.O(3), cost.B}, 6, 4, types.Human, types.Ninja)
	def.Supertypes = []types.Super{types.Legendary}
	def.StaticAbilities = []game.StaticAbility{{KeywordAbilities: game.SimpleKeywords(game.Deathtouch)}}
	def.TriggeredAbilities = []game.TriggeredAbility{{
		Text: "Whenever Shredder enters or attacks, another target creature you control gains deathtouch until end of turn.",
		Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
			Event: game.EventPermanentEnteredBattlefield, UnionEvent: game.EventAttackerDeclared,
			Source: game.TriggerSourceSelf,
		}},
		Content: game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Constraint: "another target creature you control",
				Allow: game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{
					RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true,
				}),
			}},
			Sequence: []game.Instruction{{Primitive: game.ApplyContinuous{
				Object:            opt.Val(game.TargetPermanentReference(0)),
				ContinuousEffects: []game.ContinuousEffect{{Layer: game.LayerAbility, AddKeywords: []game.Keyword{game.Deathtouch}}},
				Duration:          game.DurationUntilEndOfTurn,
			}}},
		}.Ability(),
	}}
	return def
}

func leonardoLeaderInBlueFixture() *game.CardDef {
	def := sneakCreatureFixture("Leonardo, Leader in Blue", leonardoText,
		cost.Mana{cost.W}, cost.Mana{cost.O(3), cost.W, cost.W}, 2, 1, types.Mutant, types.Ninja, types.Turtle)
	def.Supertypes = []types.Super{types.Legendary}
	def.TriggeredAbilities = []game.TriggeredAbility{{
		Text: "When Leonardo enters, if his sneak cost was paid, creatures you control get +2/+0 until end of turn.",
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhen,
			Pattern: game.TriggerPattern{
				Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
			},
			InterveningIfEventPermanentWasSneaked: true,
		},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.ApplyContinuous{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerPowerToughnessModify, PowerDelta: 2,
				Group: game.PlayerControlledGroup(game.ControllerReference(), game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
			}},
			Duration: game.DurationUntilEndOfTurn,
		}}}}.Ability(),
	}}
	return def
}

func turncoatKunoichiFixture() *game.CardDef {
	def := sneakCreatureFixture("Turncoat Kunoichi", turncoatKunoichiText,
		cost.Mana{cost.O(2), cost.W}, cost.Mana{cost.O(2), cost.W, cost.B}, 3, 2, types.Mutant, types.Ninja, types.Fox)
	def.TriggeredAbilities = []game.TriggeredAbility{{
		Text: "When this creature enters, choose target creature an opponent controls. Exile that creature until this creature leaves the battlefield. If this creature's sneak cost was paid, instead exile the chosen creature.",
		Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
			Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
		}},
		Content: game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Constraint: "target creature an opponent controls",
				Allow: game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{
					RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent,
				}),
			}},
			Program: turncoatKunoichiProgram,
		}.Ability(),
	}}
	return def
}

func splinterFixture() *game.CardDef {
	def := sneakCreatureFixture("Splinter, Hamato Yoshi", splinterText,
		cost.Mana{cost.O(1), cost.B}, cost.Mana{cost.B}, 1, 3, types.Mutant, types.Ninja, types.Rat)
	def.Supertypes = []types.Super{types.Legendary}
	def.ColorIdentity = color.NewIdentity(color.Black)
	def.StaticAbilities = []game.StaticAbility{{KeywordAbilities: game.SimpleKeywords(game.Menace)}}
	return def
}

// copyTargetSpellFixture is "Copy target spell." so a creature spell can be
// copied on the stack (CR 707.10).
func copyTargetSpellFixture() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Spell Copier", ManaCost: opt.Val(cost.Mana{cost.O(1)}), Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Text: "Copy target spell.",
			Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject,
				Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}},
			}},
			Program: copyTargetSpellProgram,
		}.Ability()),
	}}
}

// kunoichiSource records what the Turncoat Kunoichi trigger read as it
// resolved.
type kunoichiSource struct {
	resolved                     bool
	sneakPaid, sneakPaidThisTurn bool
}

func sneakPermanentEngine(t *testing.T, kunoichi *kunoichiSource) *Engine {
	t.Helper()
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{
		{
			Reference: turncoatKunoichiProgram,
			Resolve: func(ctx cardprogram.Context) error {
				source := ctx.View().Source()
				if kunoichi != nil {
					*kunoichi = kunoichiSource{resolved: true, sneakPaid: source.SneakPaid, sneakPaidThisTurn: source.SneakPaidThisTurn}
				}
				_, err := ctx.Execute(game.MovePermanent{
					Object: game.TargetPermanentReference(0), Destination: zone.Exile,
					UntilSourceLeaves: !source.SneakPaid,
				})
				return err
			},
		},
		{
			Reference: copyTargetSpellProgram,
			Resolve: func(ctx cardprogram.Context) error {
				_, err := ctx.Copy(game.CopyStackObject{Object: game.TargetStackObjectReference(0)}, ctx.View().Source().Controller)
				return err
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewEngineWithPrograms(nil, catalog)
}

func newSneakPermanentFixture(t *testing.T, kunoichi *kunoichiSource) sneakFixture {
	t.Helper()
	f := newSneakFixture(t)
	f.engine = sneakPermanentEngine(t, kunoichi)
	return f
}

func attackDeclaredFor(g *game.Game, permanent *game.Permanent) bool {
	return slices.ContainsFunc(g.Events, func(event game.Event) bool {
		return event.Kind == game.EventAttackerDeclared && event.PermanentID == permanent.ObjectID
	})
}

func enterEventFor(g *game.Game, permanent *game.Permanent) (game.Event, bool) {
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentEnteredBattlefield && event.PermanentID == permanent.ObjectID {
			return event, true
		}
	}
	return game.Event{}, false
}

// TestFootNinjasSneaksInTappedAndAttacking casts Foot Ninjas for its sneak cost
// in the declare blockers step: it enters tapped and attacking the returned
// creature's defender without being declared, its enter trigger gains 3 life,
// and it deals its 5 combat damage unblocked (CR 702.190b, 508.4, 508.4d).
func TestFootNinjasSneaksInTappedAndAttacking(t *testing.T) {
	f := newSneakPermanentFixture(t, nil)
	g := f.game
	cardID := addCardToHand(g, game.Player1, footNinjasFixture())
	g.Players[game.Player1].ManaPool.Add(mana.B, 4)
	cast := action.CastSpell(cardID, nil, 0, nil)
	if !actionsContain(f.engine.legalActions(g, game.Player1), cast) {
		t.Fatal("legal actions omit Foot Ninjas' sneak cast in the declare blockers step")
	}
	agent := &sneakAgent{cost: "Sneak"}
	f.cast(t, agent, cast)
	f.assertAttackerReturned(t)
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana left = %d, want 0 after paying {3}{W/B}", got)
	}
	obj, _ := g.Stack.Peek()
	want := game.AttackTarget{Player: game.Player2}
	if !obj.Sneaked || !obj.SneakAttackTarget.Exists || obj.SneakAttackTarget.Val != want {
		t.Fatalf("stack object Sneaked = %v, SneakAttackTarget = %+v; want the returned creature's %+v", obj.Sneaked, obj.SneakAttackTarget, want)
	}
	if cloned, _ := g.Clone().Stack.Peek(); cloned.SneakAttackTarget != obj.SneakAttackTarget {
		t.Fatalf("cloned SneakAttackTarget = %+v, want %+v", cloned.SneakAttackTarget, obj.SneakAttackTarget)
	}

	lifeBefore := g.Players[game.Player1].Life
	f.resolve(agent)
	ninjas := permanentForCard(g, cardID)
	if ninjas == nil || !ninjas.Tapped || !attackingPlayer(g, ninjas, game.Player2) {
		t.Fatalf("Foot Ninjas = %+v is not tapped and attacking Player2; attackers %+v", ninjas, g.Combat.Attackers)
	}
	if attackDeclaredFor(g, ninjas) {
		t.Fatal("Foot Ninjas was declared as an attacker")
	}
	if event, ok := enterEventFor(g, ninjas); !ok || !event.EnterSneaked {
		t.Fatalf("enter event = %+v, want EnterSneaked", event)
	}
	if cloned := permanentForCard(g.Clone(), cardID); !cloned.SneakPaid || cloned.SneakPaidTurn != g.Turn.TurnNumber {
		t.Fatalf("cloned permanent sneak state = %v/%d", cloned.SneakPaid, cloned.SneakPaidTurn)
	}
	if got := g.Players[game.Player1].Life - lifeBefore; got != 3 {
		t.Fatalf("life gained = %d, want 3", got)
	}
	opponentLife := g.Players[game.Player2].Life
	f.engine.resolveCombatDamage(g, &TurnLog{})
	if got := opponentLife - g.Players[game.Player2].Life; got != 5 {
		t.Fatalf("combat damage to Player2 = %d, want 5 from the unblocked Foot Ninjas", got)
	}
}

// TestShredderSneakTriggersOnEnterNotAttack proves "whenever Shredder enters or
// attacks" triggers once for a sneak: he enters, but was never declared as an
// attacker (CR 508.4).
func TestShredderSneakTriggersOnEnterNotAttack(t *testing.T) {
	f := newSneakPermanentFixture(t, nil)
	g := f.game
	ally := addCombatPermanent(g, game.Player1, vanillaCreature("Ally", 1, 1))
	cardID := addCardToHand(g, game.Player1, shredderFixture())
	g.Players[game.Player1].ManaPool.Add(mana.B, 4)
	agent := &sneakAgent{cost: "Sneak"}
	f.cast(t, agent, action.CastSpell(cardID, nil, 0, nil))
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	f.engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	f.engine.putTriggeredAbilitiesOnStack(g)
	shredder := permanentForCard(g, cardID)
	if shredder == nil || !shredder.Tapped || !attackingPlayer(g, shredder, game.Player2) {
		t.Fatalf("Shredder = %+v is not tapped and attacking", shredder)
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("triggers on the stack = %d, want only the enter trigger", got)
	}
	f.resolve(agent)
	if !hasKeyword(g, ally, game.Deathtouch) {
		t.Fatal("the enter trigger did not give the other creature deathtouch")
	}
	if attackDeclaredFor(g, shredder) {
		t.Fatal("Shredder was declared as an attacker")
	}
}

// TestLeonardoLeaderInBlueChecksSneakCostPaid proves the intervening "if his
// sneak cost was paid" (CR 603.4) is true for a sneak cast and false for a
// normal main-phase cast.
func TestLeonardoLeaderInBlueChecksSneakCostPaid(t *testing.T) {
	for _, sneaked := range []bool{true, false} {
		name := "normal cast"
		if sneaked {
			name = "sneaked"
		}
		t.Run(name, func(t *testing.T) {
			f := newSneakPermanentFixture(t, nil)
			g := f.game
			ally := addCombatPermanent(g, game.Player1, vanillaCreature("Ally", 1, 1))
			cardID := addCardToHand(g, game.Player1, leonardoLeaderInBlueFixture())
			agent := &sneakAgent{cost: "Sneak"}
			if sneaked {
				g.Players[game.Player1].ManaPool.Add(mana.W, 5)
			} else {
				g.Combat = nil
				setMainPhasePriority(g, game.Player1)
				g.Players[game.Player1].ManaPool.Add(mana.W, 1)
				agent.cost = "Normal cost"
			}
			f.cast(t, agent, action.CastSpell(cardID, nil, 0, nil))
			f.resolve(agent)
			leonardo := permanentForCard(g, cardID)
			if leonardo == nil || leonardo.Tapped != sneaked || leonardo.SneakPaid != sneaked {
				t.Fatalf("Leonardo = %+v, want tapped and sneak-paid %v", leonardo, sneaked)
			}
			wantBonus := 0
			if sneaked {
				wantBonus = 2
			}
			if got := effectivePower(g, ally); got != 1+wantBonus {
				t.Fatalf("ally power = %d, want %d", got, 1+wantBonus)
			}
			if got := effectivePower(g, leonardo); got != 2+wantBonus {
				t.Fatalf("Leonardo power = %d, want %d", got, 2+wantBonus)
			}
		})
	}
}

// TestTurncoatKunoichiSneakExilesPermanently proves "if this creature's sneak
// cost was paid" reads the trigger source: sneaked, the exile is permanent;
// cast normally, the creature returns when the Kunoichi leaves (CR 610.3).
func TestTurncoatKunoichiSneakExilesPermanently(t *testing.T) {
	for _, sneaked := range []bool{true, false} {
		name := "normal cast"
		if sneaked {
			name = "sneaked"
		}
		t.Run(name, func(t *testing.T) {
			var read kunoichiSource
			f := newSneakPermanentFixture(t, &read)
			g := f.game
			victim := addCombatPermanent(g, game.Player2, vanillaCreature("Victim", 2, 2))
			cardID := addCardToHand(g, game.Player1, turncoatKunoichiFixture())
			agent := &sneakAgent{cost: "Sneak"}
			if sneaked {
				g.Players[game.Player1].ManaPool.Add(mana.W, 2)
				g.Players[game.Player1].ManaPool.Add(mana.B, 2)
			} else {
				g.Combat = nil
				setMainPhasePriority(g, game.Player1)
				g.Players[game.Player1].ManaPool.Add(mana.W, 3)
				agent.cost = "Normal cost"
			}
			f.cast(t, agent, action.CastSpell(cardID, nil, 0, nil))
			f.resolve(agent)
			if !read.resolved || read.sneakPaid != sneaked || read.sneakPaidThisTurn != sneaked {
				t.Fatalf("trigger read %+v, want sneak paid (this turn) %v", read, sneaked)
			}
			if permanentForCard(g, victim.CardInstanceID) != nil || !g.Players[game.Player2].Exile.Contains(victim.CardInstanceID) {
				t.Fatal("the opponent's creature was not exiled")
			}
			resolveOpponentDestroy(t, f.engine, g, game.Player2, permanentForCard(g, cardID))
			returned := permanentForCard(g, victim.CardInstanceID) != nil
			if returned == sneaked {
				t.Fatalf("victim returned = %v after the Kunoichi left, want %v", returned, !sneaked)
			}
		})
	}
}

// TestTurncoatKunoichiSneakedLeavesBeforeTrigger reads the sneak-paid state
// from the Kunoichi's last known information when it leaves before its
// trigger resolves: the chosen creature is still exiled for good.
func TestTurncoatKunoichiSneakedLeavesBeforeTrigger(t *testing.T) {
	var read kunoichiSource
	f := newSneakPermanentFixture(t, &read)
	g := f.game
	victim := addCombatPermanent(g, game.Player2, vanillaCreature("Victim", 2, 2))
	cardID := addCardToHand(g, game.Player1, turncoatKunoichiFixture())
	g.Players[game.Player1].ManaPool.Add(mana.W, 2)
	g.Players[game.Player1].ManaPool.Add(mana.B, 2)
	agent := &sneakAgent{cost: "Sneak"}
	f.cast(t, agent, action.CastSpell(cardID, nil, 0, nil))
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	f.engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	f.engine.putTriggeredAbilitiesOnStack(g)
	resolveOpponentDestroy(t, f.engine, g, game.Player2, permanentForCard(g, cardID))
	f.resolve(agent)
	if !read.sneakPaid || !g.Players[game.Player2].Exile.Contains(victim.CardInstanceID) {
		t.Fatalf("trigger read %+v; victim exiled = %v, want a permanent exile", read, g.Players[game.Player2].Exile.Contains(victim.CardInstanceID))
	}
}

// TestSplinterSneaksFromCommandZoneWithCommanderTax casts a commander from the
// command zone for its sneak cost plus commander tax (CR 903.8): Sneak {B} plus
// {2} for one earlier cast; it enters tapped and attacking.
func TestSplinterSneaksFromCommandZoneWithCommanderTax(t *testing.T) {
	f := newSneakPermanentFixture(t, nil)
	g := f.game
	cardID := addCommanderToCommandZone(g, game.Player1, splinterFixture())
	player := g.Players[game.Player1]
	player.RecordCommanderCast(cardID)
	player.ManaPool.Add(mana.B, 1)
	cast := action.CastCommanderSpell(cardID, nil, 0, nil)
	if actionsContain(f.engine.legalActions(g, game.Player1), cast) {
		t.Fatal("Splinter's sneak cast is legal without its {2} commander tax")
	}
	player.ManaPool.Add(mana.C, 2)
	if !actionsContain(f.engine.legalActions(g, game.Player1), cast) {
		t.Fatal("legal actions omit Splinter's sneak cast from the command zone")
	}
	agent := &sneakAgent{cost: "Sneak"}
	f.cast(t, agent, cast)
	f.assertAttackerReturned(t)
	if got := player.ManaPool.Total(); got != 0 || player.CommanderTaxFor(cardID) != 4 {
		t.Fatalf("mana left = %d, tax = %d; want 0 and 4", got, player.CommanderTaxFor(cardID))
	}
	f.resolve(agent)
	splinter := permanentForCard(g, cardID)
	if splinter == nil || !splinter.Tapped || !attackingPlayer(g, splinter, game.Player2) {
		t.Fatalf("Splinter = %+v is not tapped and attacking", splinter)
	}
}

// TestSneakedCreatureFollowsPlaneswalkerAttack proves the creature attacks the
// planeswalker the returned creature attacked. It enters tapped and not
// attacking when that planeswalker left before the spell resolved (CR 506.3c,
// 508.4a), and also when the returned creature was attacking nothing as it was
// returned because the planeswalker had been removed from combat, whether by
// leaving the battlefield or while staying on it (CR 506.4, 506.4c).
func TestSneakedCreatureFollowsPlaneswalkerAttack(t *testing.T) {
	cases := []struct {
		name       string
		beforeCast func(g *game.Game, walker *game.Permanent)
		afterCast  func(g *game.Game, walker *game.Permanent)
		attacking  bool
	}{
		{name: "planeswalker remains", attacking: true},
		{name: "planeswalker left after cast", afterCast: func(g *game.Game, walker *game.Permanent) {
			movePermanentToZone(g, walker, zone.Graveyard)
		}},
		{name: "planeswalker left before cast", beforeCast: func(g *game.Game, walker *game.Permanent) {
			movePermanentToZone(g, walker, zone.Graveyard)
		}},
		{name: "planeswalker removed from combat before cast", beforeCast: func(g *game.Game, walker *game.Permanent) {
			removePermanentFromCombat(g, walker.ObjectID)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newSneakPermanentFixture(t, nil)
			g := f.game
			walker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Test Planeswalker", Types: []types.Card{types.Planeswalker}, Loyalty: opt.Val(3),
			}})
			target := game.AttackTarget{Player: game.Player2, PlaneswalkerID: walker.ObjectID}
			g.Combat.Attackers[0].Target = target
			if tc.beforeCast != nil {
				tc.beforeCast(g, walker)
				if !g.Combat.Attackers[0].Target.NoTarget {
					t.Fatalf("returned attacker target = %+v, want attacking nothing", g.Combat.Attackers[0].Target)
				}
			}
			cardID := addCardToHand(g, game.Player1, footNinjasFixture())
			g.Players[game.Player1].ManaPool.Add(mana.W, 4)
			agent := &sneakAgent{cost: "Sneak"}
			f.cast(t, agent, action.CastSpell(cardID, nil, 0, nil))
			if tc.beforeCast == nil {
				if obj, _ := g.Stack.Peek(); obj.SneakAttackTarget.Val != target {
					t.Fatalf("SneakAttackTarget = %+v, want %+v", obj.SneakAttackTarget, target)
				}
			}
			if tc.afterCast != nil {
				tc.afterCast(g, walker)
			}
			f.resolve(agent)
			ninjas := permanentForCard(g, cardID)
			if ninjas == nil {
				t.Fatal("Foot Ninjas did not enter the battlefield")
			}
			attackingWalker := slices.ContainsFunc(g.Combat.Attackers, func(attack game.AttackDeclaration) bool {
				return attack.Attacker == ninjas.ObjectID && attack.Target == target
			})
			if !ninjas.Tapped || attackingWalker != tc.attacking || permanentIsAttacking(g, ninjas) != tc.attacking {
				t.Fatalf("Foot Ninjas tapped = %v, attacking the planeswalker = %v, attacking = %v; want tapped, attacking %v (attackers %+v)",
					ninjas.Tapped, attackingWalker, permanentIsAttacking(g, ninjas), tc.attacking, g.Combat.Attackers)
			}
		})
	}
}

// TestCopyOfSneakedCreatureSpellEntersAttacking proves a copy of a sneaked
// creature spell keeps the paid sneak cost (CR 707.10): its token enters
// tapped and attacking the same defender.
func TestCopyOfSneakedCreatureSpellEntersAttacking(t *testing.T) {
	f := newSneakPermanentFixture(t, nil)
	g := f.game
	cardID := addCardToHand(g, game.Player1, footNinjasFixture())
	copierID := addCardToHand(g, game.Player1, copyTargetSpellFixture())
	g.Players[game.Player1].ManaPool.Add(mana.W, 5)
	agent := &sneakAgent{cost: "Sneak"}
	f.cast(t, agent, action.CastSpell(cardID, nil, 0, nil))
	original, _ := g.Stack.Peek()
	agent.cost = "Normal cost"
	f.cast(t, agent, action.CastSpell(copierID, []game.Target{game.StackObjectTarget(original.ID)}, 0, nil))
	f.engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
	copied, _ := g.Stack.Peek()
	if !copied.Copy || !copied.Sneaked || copied.SneakAttackTarget != original.SneakAttackTarget {
		t.Fatalf("copy = %+v, want the original's paid sneak cost and attack target", copied)
	}
	f.resolve(agent)
	var ninjas []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token || permanent.CardInstanceID == cardID {
			ninjas = append(ninjas, permanent)
		}
	}
	if len(ninjas) != 2 {
		t.Fatalf("Foot Ninjas permanents = %d, want the card and its token copy", len(ninjas))
	}
	for _, permanent := range ninjas {
		if !permanent.Tapped || !permanent.SneakPaid || !attackingPlayer(g, permanent, game.Player2) {
			t.Fatalf("permanent %+v is not a tapped, sneak-paid attacker", permanent)
		}
	}
}

// TestSneakedNoncreaturePermanentEntersTappedNotAttacking proves a sneaked
// permanent that is not a creature enters tapped but never attacking (CR
// 506.3a, 702.190b).
func TestSneakedNoncreaturePermanentEntersTappedNotAttacking(t *testing.T) {
	f := newSneakPermanentFixture(t, nil)
	g := f.game
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Sneak Artifact", Types: []types.Card{types.Artifact}, ManaCost: opt.Val(cost.Mana{cost.O(3)}),
		AlternativeCosts: []cost.Alternative{cost.SneakAlternative(cost.Mana{cost.U})},
	}})
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	agent := &sneakAgent{cost: "Sneak"}
	f.cast(t, agent, action.CastSpell(cardID, nil, 0, nil))
	f.assertAttackerReturned(t)
	f.resolve(agent)
	artifact := permanentForCard(g, cardID)
	if artifact == nil || !artifact.Tapped || permanentIsAttacking(g, artifact) || len(g.Combat.Attackers) != 0 {
		t.Fatalf("artifact = %+v, attackers %+v; want tapped and not attacking", artifact, g.Combat.Attackers)
	}
}

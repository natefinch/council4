package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func numericTargetSpellEntries() []cardset.Entry {
	return []cardset.Entry{authored.DeathDenied, authored.WildestDreams, authored.RatsFeast, authored.ShatteredCrypt, authored.ReturnToTheRanks}
}

func TestProductionNumericTargetCardsExactCountsAndClones(t *testing.T) {
	for _, entry := range numericTargetSpellEntries() {
		for _, count := range []int{0, 21} {
			t.Run(fmt.Sprintf("%s/X=%d", entry.Name, count), func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				setMainPhasePriority(g, game.Player1)
				engine := removalTokenEngine(t, entry)
				owner := game.Player1
				if entry.OracleID == authored.RatsFeast.OracleID {
					owner = game.Player2
				}
				var targets []game.Target
				for range count {
					def := vanillaCreature("Numeric return", 2, 2)
					def.ManaCost = opt.Val(cost.Mana{cost.O(2)})
					card := addCardToGraveyard(g, owner, def)
					targets = append(targets, currentCardTarget(t, g, card))
				}
				spell := addCardToHand(g, game.Player1, entry.New())
				for _, c := range []mana.Color{mana.C, mana.B, mana.G, mana.W} {
					g.Players[game.Player1].ManaPool.Add(c, 50)
				}
				branch, accepted := engine.Simulator().Apply(g, game.Player1, action.CastSpell(spell, targets, count, nil), simPassPolicies())
				if !accepted || !g.Players[game.Player1].Hand.Contains(spell) {
					t.Fatal("named numeric declaration was rejected or shared its source")
				}
				branch = engine.Simulator().ResolvePriority(branch, simPassPolicies())
				var simultaneous game.ObjectID
				for _, target := range targets {
					arrived := branch.Players[owner].Hand.Contains(target.CardID)
					switch entry.OracleID {
					case authored.RatsFeast.OracleID:
						arrived = branch.Players[owner].Exile.Contains(target.CardID)
					case authored.ReturnToTheRanks.OracleID:
						arrived = permanentByCardID(branch, target.CardID) != nil
					default:
					}
					if !arrived || !g.Players[owner].Graveyard.Contains(target.CardID) {
						t.Fatal("an exact target was omitted or source clone mutated")
					}
					for _, event := range branch.Events {
						if event.Kind == game.EventZoneChanged && event.CardID == target.CardID {
							if count > 1 && (event.SimultaneousID == 0 || simultaneous != 0 && simultaneous != event.SimultaneousID) {
								t.Fatal("numeric target arrivals were not one simultaneous operation")
							}
							simultaneous = event.SimultaneousID
						}
					}
				}
				wantLife := 40
				if entry.OracleID == authored.ShatteredCrypt.OracleID {
					wantLife -= count
				}
				if branch.Players[game.Player1].Life != wantLife {
					t.Fatal("announced X did not control the complete resolving rider")
				}
				if entry.OracleID == authored.WildestDreams.OracleID {
					if !branch.Players[game.Player1].Exile.Contains(spell) {
						t.Fatal("Wildest Dreams omitted its resolving-spell exile")
					}
				} else if !branch.Players[game.Player1].Graveyard.Contains(spell) {
					t.Fatal("numeric spell did not finish in its normal destination")
				}
			})
		}
	}
}

func TestProductionNumericTargetCardsIllegalAndReplacementResults(t *testing.T) {
	for _, entry := range numericTargetSpellEntries() {
		t.Run(entry.Name, func(t *testing.T) {
			for _, allIllegal := range []bool{false, true} {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := removalTokenEngine(t, entry)
				first := addCardToGraveyard(g, game.Player1, vanillaCreature("First target", 2, 2))
				second := addCardToGraveyard(g, game.Player1, vanillaCreature("Second target", 2, 2))
				spell := addImplementationSpellToStack(g, game.Player1, entry.New(),
					[]game.Target{currentCardTarget(t, g, first), currentCardTarget(t, g, second)})
				obj, _ := g.Stack.Peek()
				obj.XValue, obj.TargetCounts = 2, []int{2}
				moveCardBetweenZones(g, game.Player1, first, zone.Graveyard, zone.Exile)
				if allIllegal {
					moveCardBetweenZones(g, game.Player1, second, zone.Graveyard, zone.Exile)
				} else if entry.OracleID == authored.DeathDenied.OracleID ||
					entry.OracleID == authored.WildestDreams.OracleID || entry.OracleID == authored.ShatteredCrypt.OracleID {
					g.ReplacementEffects = []game.ReplacementEffect{{
						ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Graveyard,
						MatchToZone: true, ToZone: zone.Hand, ReplaceToZone: zone.Exile,
					}}
				}
				engine.resolveTopOfStack(g, &TurnLog{})
				if allIllegal {
					if !g.Players[game.Player1].Graveyard.Contains(spell) || g.Players[game.Player1].Life != 40 {
						t.Fatal("all-illegal targets ran a rider or exiled Wildest Dreams")
					}
					continue
				}
				arrived := g.Players[game.Player1].Exile.Contains(second)
				if entry.OracleID == authored.ReturnToTheRanks.OracleID {
					arrived = permanentByCardID(g, second) != nil
				}
				if !arrived || !g.Players[game.Player1].Exile.Contains(first) {
					t.Fatal("partial-illegal targeting or actual replacement destination was lost")
				}
				if entry.OracleID == authored.ShatteredCrypt.OracleID && g.Players[game.Player1].Life != 38 {
					t.Fatal("partial-illegal targets changed the announced life-loss amount")
				}
			}
		})
	}
}

func TestProductionReturnToRanksConvokeAndRestrictions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	engine := removalTokenEngine(t, authored.ReturnToTheRanks)
	var targets []game.Target
	for range 21 {
		card := addCardToGraveyard(g, game.Player1, vanillaCreature("Small return", 2, 2))
		targets = append(targets, currentCardTarget(t, g, card))
	}
	for range 23 {
		convoke := vanillaCreature("Convoke payer", 1, 1)
		convoke.Colors = []color.Color{color.White}
		addCombatPermanent(g, game.Player1, convoke)
	}
	spell := addCardToHand(g, game.Player1, authored.ReturnToTheRanks.New())
	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, targets, 21, nil)) {
		t.Fatal("Return to the Ranks could not convoke its exact total cost")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	for _, target := range targets {
		if permanentByCardID(g, target.CardID) == nil {
			t.Fatal("Return to the Ranks omitted a convoke-funded arrival")
		}
	}
	wrong := vanillaCreature("Too expensive", 3, 3)
	wrong.ManaCost = opt.Val(cost.Mana{cost.O(3)})
	card := addCardToGraveyard(g, game.Player1, wrong)
	spell = addCardToHand(g, game.Player1, authored.ReturnToTheRanks.New())
	if engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{currentCardTarget(t, g, card)}, 1, nil)) {
		t.Fatal("Return to the Ranks ignored its mana-value target restriction")
	}
}

func TestProductionChampionOfStraySoulsCompleteAbilities(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	engine := removalTokenEngine(t, authored.ChampionOfStraySouls)
	source := addCombatPermanent(g, game.Player1, authored.ChampionOfStraySouls.New())
	source.SummoningSick = false
	var targets []game.Target
	for range 21 {
		card := addCardToGraveyard(g, game.Player1, vanillaCreature("Champion return", 2, 2))
		targets = append(targets, currentCardTarget(t, g, card))
	}
	for range 20 {
		addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 8)
	g.Players[game.Player1].ManaPool.Add(mana.B, 4)
	activate := action.ActivateAbility(source.ObjectID, 0, targets, 21)
	if engine.applyAction(g, game.Player1, activate) || source.Tapped || g.Players[game.Player1].ManaPool.Total() != 12 {
		t.Fatal("Champion sacrificed itself or partially paid an impossible X")
	}
	addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	if !engine.applyAction(g, game.Player1, activate) || !source.Tapped {
		t.Fatal("Champion could not pay the exact sacrifice/tap/mana costs")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	for _, target := range targets {
		if permanentByCardID(g, target.CardID) == nil {
			t.Fatal("Champion omitted a chosen return beyond the legacy width")
		}
	}
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("could not put Champion in its graveyard")
	}
	g.Players[game.Player1].ManaPool.Add(mana.B, 2)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.CardInstanceID, 1, nil, 0)) {
		_, ability, found := graveyardAbilitySource(g, game.Player1, source.CardInstanceID, 1)
		t.Fatalf("Champion graveyard ability rejected: found=%v zone=%v mana=%d C/%d B can=%v",
			found, game.BodyFunctionZone(&ability), g.Players[game.Player1].ManaPool.Amount(mana.C),
			g.Players[game.Player1].ManaPool.Amount(mana.B),
			canActivateGraveyardAbility(g, game.Player1, source.CardInstanceID, &ability, 1, nil, 0))
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	top, present := g.Players[game.Player1].Library.Top()
	if !present || top != source.CardInstanceID {
		t.Fatal("Champion did not put its exact graveyard incarnation on top")
	}
}

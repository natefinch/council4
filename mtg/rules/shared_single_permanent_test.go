package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompiledSinglePermanentColorGateReadsResolvingTarget(t *testing.T) {
	for _, tc := range []struct {
		name     string
		colors   []color.Color
		departed bool
		destroy  bool
	}{
		{"blue", []color.Color{color.Blue}, false, true},
		{"red after announcement", []color.Color{color.Red}, false, false},
		{"colorless after announcement", nil, false, false},
		{"unavailable blue target", []color.Color{color.Blue}, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			def := legacyCard(t, carddata.ScryfallCard{
				Name: "Color Gate Probe", Layout: "normal", TypeLine: "Instant",
				OracleText: "Destroy target creature if it's blue.",
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			creature := permanentDef("Target", nil, types.Creature)
			creature.Colors = []color.Color{color.Blue}
			target := addCombatPermanent(g, game.Player2, creature)
			decoy := addCombatPermanent(g, game.Player3, permanentDef("Blue Decoy", nil, types.Creature))
			g.CardInstances[decoy.CardInstanceID].Def.Colors = []color.Color{color.Blue}
			source := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].Hand.Remove(source)
			g.Stack.Push(&game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: source, Controller: game.Player1,
				Targets: []game.Target{game.PermanentTarget(target.ObjectID)}, TargetCounts: []int{1},
			})
			creature.Colors = tc.colors
			if tc.departed {
				g.Battlefield = slices.DeleteFunc(g.Battlefield, func(p *game.Permanent) bool {
					return p.ObjectID == target.ObjectID
				})
			}
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			if got := g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID); got != tc.destroy {
				t.Fatalf("destroyed=%v, want %v", got, tc.destroy)
			}
			if _, ok := permanentByObjectID(g, decoy.ObjectID); !ok {
				t.Fatal("resolving gate substituted another blue permanent for the target")
			}
		})
	}
}

func TestCompiledSinglePermanentQualifierCandidates(t *testing.T) {
	for _, tc := range []struct {
		name       string
		cardType   types.Card
		tapped     bool
		controller game.PlayerID
		owner      game.PlayerID
		want       bool
	}{
		{"opponent controls despite own ownership", types.Artifact, true, game.Player2, game.Player1, true},
		{"controller is not owner", types.Artifact, true, game.Player1, game.Player2, false},
		{"untapped", types.Artifact, false, game.Player2, game.Player2, false},
		{"land", types.Land, true, game.Player2, game.Player2, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			def := legacyCard(t, carddata.ScryfallCard{
				Name: "Qualifier Probe", Layout: "normal", TypeLine: "Instant",
				OracleText: "Destroy target tapped nonland permanent an opponent controls.",
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addCombatPermanent(g, tc.owner, permanentDef("Candidate", nil, tc.cardType))
			target.Controller, target.Tapped = tc.controller, tc.tapped
			source := addCardToHand(g, game.Player1, def)
			spec := &def.SpellAbility.Val.Modes[0].Targets[0]
			candidates := targetCandidatesForSpecChosenBy(g, game.Player1, game.Player1, def, source, game.Event{}, spec)
			if got := slices.Contains(candidates, game.PermanentTarget(target.ObjectID)); got != tc.want {
				t.Fatalf("candidate=%v, want %v", got, tc.want)
			}
			g.Players[game.Player1].Hand.Remove(source)
			g.Stack.Push(&game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: source, Controller: game.Player1,
				Targets: []game.Target{game.PermanentTarget(target.ObjectID)}, TargetCounts: []int{1},
			})
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			if got := g.Players[tc.owner].Graveyard.Contains(target.CardInstanceID); got != tc.want {
				t.Fatalf("destroyed into physical owner's graveyard=%v, want %v", got, tc.want)
			}
		})
	}

}

func TestCompiledSinglePermanentAttackDefenderNotEventPlayer(t *testing.T) {
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "The Spear of Bashenga", Layout: "normal", ManaCost: "{4}{W}", TypeLine: "Legendary Artifact — Equipment",
		OracleText: "When The Spear of Bashenga enters, if there is no monarch, you become the monarch.\n" +
			"Equipped creature gets +2/+2 and has vigilance.\n" +
			"Whenever equipped creature attacks the monarch, destroy target tapped nonland permanent that player controls.\n" +
			"Equip {2}",
	})
	g, spear, attacker := newSpearCombat(t)
	g.CardInstances[spear.CardInstanceID].Def = def
	victim := addCombatPermanent(g, game.Player3, permanentDef("Defender's Artifact", nil, types.Artifact))
	victim.Controller, victim.Tapped = game.Player2, true
	decoy := addCombatPermanent(g, game.Player2, permanentDef("Attacker's Artifact", nil, types.Artifact))
	decoy.Controller, decoy.Tapped = game.Player1, true
	if !setMonarch(g, game.Player2) {
		t.Fatal("could not set defending monarch")
	}
	engine := NewEngine(nil)
	attack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{{
		Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}))
	if !engine.applyDeclareAttackers(g, game.Player1, attack) {
		t.Fatal("could not declare actual monarch attack")
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}
	log := &TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
		t.Fatal("compiled attack-defender body did not trigger")
	}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	if !g.Players[game.Player3].Graveyard.Contains(victim.CardInstanceID) {
		t.Fatal("defender's permanent did not go to its physical owner's graveyard")
	}
	if _, ok := permanentByObjectID(g, decoy.ObjectID); !ok {
		t.Fatal("defender qualifier was replaced by event player or physical owner")
	}
}

func TestCompiledSinglePermanentEventPlayerNotOwner(t *testing.T) {
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "Event Player Destroy Probe", Layout: "normal", TypeLine: "Creature", Power: new("2"), Toughness: new("2"),
		OracleText: "Whenever this creature deals combat damage to a player, destroy target artifact or enchantment that player controls.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	victim := addCombatPermanent(g, game.Player3, permanentDef("Victim", nil, types.Artifact))
	victim.Controller = game.Player2
	decoy := addCombatPermanent(g, game.Player2, permanentDef("Decoy", nil, types.Enchantment))
	decoy.Controller = game.Player3
	engine := NewEngine(nil)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: source.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}}
	engine.resolveCombatDamage(g, &TurnLog{})
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}
	log := &TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
		t.Fatal("compiled event-player body did not trigger")
	}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	if !g.Players[game.Player3].Graveyard.Contains(victim.CardInstanceID) {
		t.Fatal("event player's permanent did not go to its physical owner's graveyard")
	}
	if _, ok := permanentByObjectID(g, decoy.ObjectID); !ok {
		t.Fatal("event-player qualifier was replaced by physical owner")
	}
}

func TestCompiledSinglePermanentOptionalDestroyChoice(t *testing.T) {
	for _, accept := range []bool{false, true} {
		t.Run(map[bool]string{false: "declined", true: "accepted"}[accept], func(t *testing.T) {
			def := legacyCard(t, carddata.ScryfallCard{
				Name: "Optional Destroy Probe", Layout: "normal", TypeLine: "Instant",
				OracleText: "You may destroy target creature.",
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
			source := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].Hand.Remove(source)
			g.Stack.Push(&game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: source, Controller: game.Player1,
				Targets: []game.Target{game.PermanentTarget(target.ObjectID)}, TargetCounts: []int{1},
			})
			choice := 0
			if accept {
				choice = 1
			}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{choice}}}}
			NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			if got := g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID); got != accept {
				t.Fatalf("destroyed=%v, want %v", got, accept)
			}
		})
	}
}

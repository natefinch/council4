package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine proving fixture for a granted sneak ability and the
// permission to cast creature spells from the graveyard using sneak (#3916),
// not the production registry definition, from the pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. Ninja
// Teen: oracle 2c5044ff-a303-4276-8e23-7caac30a275b, printing
// 0825a28f-f60b-4f80-83e3-cad6f9b266ce, {2}{B} Enchantment — Class: "Whenever
// a creature you control leaves the battlefield, each opponent loses 1 life.
// {1}{B}: Level 2 / Creatures you control get +1/+0 and have menace. {B}:
// Level 3 / Creature cards in your graveyard have sneak {3}{B}. / You may cast
// creature spells from your graveyard using their sneak abilities." Read the
// complete Scryfall rulings response (all 2026-01-27): a creature card in your
// graveyard with multiple sneak abilities may be cast using any of them; the
// permission does not change timing (only in your declare blockers step when
// you could cast an instant, still paying the sneak cost including returning
// an unblocked attacker); the sneaked creature enters tapped and attacking the
// returned creature's defender and was never declared as an attacker. Foot
// Ninjas (oracle 306652a8-9a9c-4b33-96d7-b59234d148e9) supplies the printed
// Sneak {3}{W/B}. Verified official CR (effective 2026-11-06 copy) read:
// 702.190a-b (sneak functions on the stack; its timing and return cost),
// 601.2a-b, 601.3 and 601.3e (a permission to cast from a zone, and an
// alternative cost granted only in that zone), 113.6d (abilities of cards in
// graveyards), 118.9a-b (one alternative cost), 707.10 (copies keep the paid
// alternative cost), 508.4 (entered attacking, never declared).
const ninjaTeenText = "(Gain the next level as a sorcery to add its ability.)\n" +
	"Whenever a creature you control leaves the battlefield, each opponent loses 1 life.\n" +
	"{1}{B}: Level 2\nCreatures you control get +1/+0 and have menace.\n" +
	"{B}: Level 3\nCreature cards in your graveyard have sneak {3}{B}.\n" +
	"You may cast creature spells from your graveyard using their sneak abilities."

// ninjaTeenFixture carries the level 1 trigger and both level 3 abilities;
// tests set the Class level directly.
func ninjaTeenFixture() *game.CardDef {
	levelThree := opt.Val(game.Condition{SourceClassLevelAtLeast: 3})
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Ninja Teen", ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B}), OracleText: ninjaTeenText,
		Types: []types.Card{types.Enchantment}, Subtypes: []types.Sub{types.Class},
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: "Whenever a creature you control leaves the battlefield, each opponent loses 1 life.",
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event: game.EventZoneChanged, Controller: game.TriggerControllerYou,
				SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
				MatchFromZone:    true, FromZone: zone.Battlefield,
			}},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.LoseLife{Amount: game.Fixed(1), PlayerGroup: game.OpponentsReference()},
			}}}.Ability(),
		}},
		StaticAbilities: []game.StaticAbility{
			{
				Text:      "Creature cards in your graveyard have sneak {3}{B}.",
				Condition: levelThree,
				RuleEffects: []game.RuleEffect{{
					Kind:           game.RuleEffectGrantGraveyardCardKeyword,
					AffectedPlayer: game.PlayerYou,
					CardSelection:  game.Selection{RequiredTypes: []types.Card{types.Creature}},
					GrantedKeyword: game.Sneak,
					GraveyardCastCost: game.GraveyardCastGrantCost{
						ManaCost: opt.Val(cost.Mana{cost.O(3), cost.B}),
					},
				}},
			},
			{
				Text:      "You may cast creature spells from your graveyard using their sneak abilities.",
				Condition: levelThree,
				RuleEffects: []game.RuleEffect{{
					Kind:           game.RuleEffectCastSpellsFromGraveyardWithSneak,
					AffectedPlayer: game.PlayerYou,
					SpellTypes:     []types.Card{types.Creature},
				}},
			},
		},
	}}
}

func graveyardBearsFixture() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Grizzly Bears", ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
		Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Bear},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
	}}
}

// newNinjaTeenFixture is the Sneak window with Ninja Teen at level under
// controller's control.
func newNinjaTeenFixture(t *testing.T, controller game.PlayerID, level int) sneakFixture {
	t.Helper()
	f := newSneakPermanentFixture(t, nil)
	addCombatPermanent(f.game, controller, ninjaTeenFixture()).ClassLevel = level
	return f
}

// TestNinjaTeenSneaksCreatureFromGraveyard casts a creature card from the
// graveyard for Ninja Teen's granted sneak {3}{B}: the normal cost is never
// offered, the cast returns the unblocked attacker (triggering Ninja Teen's
// leaves-the-battlefield ability), counts as cast, and the creature enters
// tapped and attacking without being declared (CR 601.3e, 702.190a-b, 508.4).
func TestNinjaTeenSneaksCreatureFromGraveyard(t *testing.T) {
	f := newNinjaTeenFixture(t, game.Player1, 3)
	g := f.game
	cardID := addCardToGraveyard(g, game.Player1, graveyardBearsFixture())
	g.Players[game.Player1].ManaPool.Add(mana.B, 4)
	g.Players[game.Player1].ManaPool.Add(mana.G, 2)
	watcher := watchCasts(g)
	cast := action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)
	if !actionsContain(f.engine.legalActions(g, game.Player1), cast) {
		t.Fatal("legal actions omit the graveyard sneak cast")
	}
	agent := &sneakAgent{cost: "Sneak"}
	f.cast(t, agent, cast)

	// The permission allows only sneak, so no payment choice is offered.
	if len(agent.offered) != 0 {
		t.Fatalf("offered payments = %v, want only the granted Sneak", agent.offered)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 2 {
		t.Fatalf("mana left = %d, want 2 after paying {3}{B}", got)
	}
	f.assertAttackerReturned(t)
	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("cast card is still in the graveyard")
	}
	obj, ok := g.Stack.Peek()
	want := game.AttackTarget{Player: game.Player2}
	if !ok || obj.SourceID != cardID || !obj.Sneaked || obj.SneakAttackTarget != opt.Val(want) {
		t.Fatalf("stack object = %+v, want a sneaked Grizzly Bears attacking %+v", obj, want)
	}
	if cloned, _ := g.Clone().Stack.Peek(); !cloned.Sneaked || cloned.SneakAttackTarget != obj.SneakAttackTarget {
		t.Fatalf("cloned stack object = %+v, want the sneak state", cloned)
	}
	if triggers := resolveCastWatcherTriggers(t, g, f.engine, watcher); triggers != 1 {
		t.Fatalf("cast triggers = %d, want 1 for a graveyard sneak cast", triggers)
	}

	opponentLife := g.Players[game.Player2].Life
	f.resolve(agent)
	if got := opponentLife - g.Players[game.Player2].Life; got != 1 {
		t.Fatalf("Player2 lost %d life, want 1 from the returned attacker leaving", got)
	}
	bears := permanentForCard(g, cardID)
	if bears == nil || !bears.Tapped || !bears.SneakPaid || !attackingPlayer(g, bears, game.Player2) {
		t.Fatalf("Grizzly Bears = %+v is not a tapped, sneak-paid attacker of Player2", bears)
	}
	if attackDeclaredFor(g, bears) {
		t.Fatal("Grizzly Bears was declared as an attacker")
	}
}

// TestNinjaTeenGraveyardSneakRequiresItsPermission proves the granted sneak is
// usable only from its controller's graveyard, at level 3, for creature
// spells, and only in the Sneak window (CR 702.190a, 601.3).
func TestNinjaTeenGraveyardSneakRequiresItsPermission(t *testing.T) {
	tests := []struct {
		name       string
		controller game.PlayerID
		level      int
		owner      game.PlayerID
		def        func() *game.CardDef
		inHand     bool
		mutate     func(sneakFixture)
	}{
		{name: "level 2", controller: game.Player1, level: 2, owner: game.Player1, def: graveyardBearsFixture},
		{name: "opponent controls Ninja Teen", controller: game.Player2, level: 3, owner: game.Player1, def: graveyardBearsFixture},
		{name: "opponent's graveyard", controller: game.Player1, level: 3, owner: game.Player2, def: graveyardBearsFixture},
		{name: "creature card in hand", controller: game.Player1, level: 3, owner: game.Player1, def: graveyardBearsFixture, inHand: true},
		{name: "noncreature card with printed sneak", controller: game.Player1, level: 3, owner: game.Player1, def: donatellosTechniqueFixture},
		{name: "main phase", controller: game.Player1, level: 3, owner: game.Player1, def: graveyardBearsFixture, mutate: func(f sneakFixture) {
			f.game.Combat = nil
			setMainPhasePriority(f.game, game.Player1)
		}},
		{name: "attacker was blocked", controller: game.Player1, level: 3, owner: game.Player1, def: graveyardBearsFixture, mutate: func(f sneakFixture) {
			f.game.Combat.BlockedAttackers = map[game.ObjectID]bool{f.attacker.ObjectID: true}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newNinjaTeenFixture(t, tt.controller, tt.level)
			g := f.game
			g.Players[game.Player1].ManaPool.Add(mana.B, 6)
			g.Players[game.Player1].ManaPool.Add(mana.G, 2)
			g.Players[game.Player1].ManaPool.Add(mana.U, 2)
			var cast action.Action
			if tt.inHand {
				cast = action.CastSpell(addCardToHand(g, tt.owner, tt.def()), nil, 0, nil)
			} else {
				cast = action.CastSpellFromZone(addCardToGraveyard(g, tt.owner, tt.def()), zone.Graveyard, nil, 0, nil)
			}
			if tt.mutate != nil {
				tt.mutate(f)
			}
			if actionsContain(f.engine.legalActions(g, game.Player1), cast) {
				t.Fatal("card was castable without Ninja Teen's graveyard sneak permission")
			}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: &sneakAgent{cost: "Sneak"}}
			if f.engine.applyActionWithChoices(g, game.Player1, cast, agents, &TurnLog{}) {
				t.Fatal("cast succeeded without Ninja Teen's graveyard sneak permission")
			}
		})
	}
}

// TestNinjaTeenOffersEachSneakAbilityFromGraveyard proves a creature card
// with a printed sneak may be cast from the graveyard using either its own
// sneak or the granted one (Ninja Teen ruling), each presented with the cost
// it pays, and both enter attacking.
func TestNinjaTeenOffersEachSneakAbilityFromGraveyard(t *testing.T) {
	for _, tt := range []struct {
		name  string
		label string
		pool  mana.Color
	}{{name: "printed sneak", label: "Sneak {3}{W/B}", pool: mana.W}, {name: "granted sneak", label: "Sneak {3}{B}", pool: mana.B}} {
		t.Run(tt.name, func(t *testing.T) {
			f := newNinjaTeenFixture(t, game.Player1, 3)
			g := f.game
			cardID := addCardToGraveyard(g, game.Player1, footNinjasFixture())
			g.Players[game.Player1].ManaPool.Add(mana.B, 3)
			g.Players[game.Player1].ManaPool.Add(tt.pool, 1)
			var offered []string
			agent := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
				if request.Kind != game.ChoicePayment {
					return request.DefaultSelection
				}
				chosen := -1
				for _, option := range request.Options {
					offered = append(offered, option.Label)
					if option.Label == tt.label {
						chosen = option.Index
					}
					if !option.ManaCost.Exists || "Sneak "+option.ManaCost.Val.String() != option.Label {
						t.Fatalf("option %+v does not carry the mana cost it names", option)
					}
				}
				if chosen < 0 {
					t.Fatalf("payment options = %v, want %q", offered, tt.label)
				}
				return []int{chosen}
			}}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
			if !f.engine.applyActionWithChoices(g, game.Player1, action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil), agents, &TurnLog{}) {
				t.Fatalf("graveyard sneak cast failed; offered %v", offered)
			}
			if len(offered) != 2 {
				t.Fatalf("offered = %v, want exactly the two Sneak abilities", offered)
			}
			if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
				t.Fatalf("mana left = %d, want 0 after paying the chosen sneak", got)
			}
			f.assertAttackerReturned(t)
			resolveStackWithTriggers(f.engine, g, agents)
			ninjas := permanentForCard(g, cardID)
			if ninjas == nil || !ninjas.Tapped || !ninjas.SneakPaid || !attackingPlayer(g, ninjas, game.Player2) {
				t.Fatalf("Foot Ninjas = %+v is not a tapped, sneak-paid attacker", ninjas)
			}
		})
	}
}

// TestSpellCostChoiceAcrossRoutesNamesEachCost proves the multi-route spell
// cost choice (CR 601.2b, 118.9a) qualifies equally named options by the cost
// each pays: with Ninja Teen's sneak permission and Leonardo's graveyard cast
// permission both reaching Foot Ninjas, every offered option is distinct and
// each sneak names its own mana cost, so the granted {3}{B} can be chosen.
func TestSpellCostChoiceAcrossRoutesNamesEachCost(t *testing.T) {
	f := newNinjaTeenFixture(t, game.Player1, 3)
	g := f.game
	addCombatPermanent(g, game.Player1, leonardoShape())
	cardID := addCardToGraveyard(g, game.Player1, footNinjasFixture())
	g.Players[game.Player1].ManaPool.Add(mana.B, 4)
	var offered []game.ChoiceOption
	agent := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
		if request.Kind != game.ChoicePayment || offered != nil {
			return request.DefaultSelection
		}
		offered = request.Options
		for _, option := range request.Options {
			if strings.HasPrefix(option.Label, "Sneak {3}{B}") {
				return []int{option.Index}
			}
		}
		return request.DefaultSelection
	}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	if !f.engine.applyActionWithChoices(g, game.Player1, action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatalf("graveyard sneak cast failed; offered %+v", offered)
	}
	labels := map[string]bool{}
	costs := map[string]bool{}
	for _, option := range offered {
		if labels[option.Label] {
			t.Fatalf("offered %+v repeats %q", offered, option.Label)
		}
		labels[option.Label] = true
		if !option.ManaCost.Exists {
			t.Fatalf("option %+v carries no mana cost", option)
		}
		if strings.HasPrefix(option.Label, "Sneak") {
			if !strings.HasPrefix(option.Label, "Sneak "+option.ManaCost.Val.String()) {
				t.Fatalf("option %+v does not name its mana cost", option)
			}
			costs[option.ManaCost.Val.String()] = true
		}
	}
	if !costs["{3}{W/B}"] || !costs["{3}{B}"] {
		t.Fatalf("offered %+v, want both the printed and the granted sneak", offered)
	}
	if obj, ok := g.Stack.Peek(); !ok || !obj.Sneaked || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("stack object = %+v, want Foot Ninjas sneaked for {3}{B}", obj)
	}
	f.assertAttackerReturned(t)
}

// TestNinjaTeenGraveyardSneakCombinesWithKicker pays the granted sneak with an
// optional kicker cost from the graveyard (CR 118.9d, 601.2f).
func TestNinjaTeenGraveyardSneakCombinesWithKicker(t *testing.T) {
	f := newNinjaTeenFixture(t, game.Player1, 3)
	g := f.game
	def := graveyardBearsFixture()
	def.StaticAbilities = []game.StaticAbility{{
		KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.G}}},
	}}
	cardID := addCardToGraveyard(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.B, 4)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	agent := &sneakAgent{cost: "Sneak"}
	f.cast(t, agent, action.CastKickedSpellFromZone(cardID, zone.Graveyard, nil, 0, nil))
	for _, labels := range agent.offered {
		for _, label := range labels {
			if !strings.Contains(label, "Sneak") {
				t.Fatalf("offered %v, want only Sneak", agent.offered)
			}
		}
	}
	obj, _ := g.Stack.Peek()
	if !obj.Sneaked || !obj.KickerPaid {
		t.Fatalf("Sneaked = %v, KickerPaid = %v; want both", obj.Sneaked, obj.KickerPaid)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana left = %d, want 0 after {3}{B} plus kicker {G}", got)
	}
	f.assertAttackerReturned(t)
}

// TestCopyOfGraveyardSneakedCreatureSpellEntersAttacking proves a copy of a
// creature spell cast with the granted sneak keeps the paid sneak cost (CR
// 707.10), so its token also enters tapped and attacking.
func TestCopyOfGraveyardSneakedCreatureSpellEntersAttacking(t *testing.T) {
	f := newNinjaTeenFixture(t, game.Player1, 3)
	g := f.game
	cardID := addCardToGraveyard(g, game.Player1, graveyardBearsFixture())
	copierID := addCardToHand(g, game.Player1, copyTargetSpellFixture())
	g.Players[game.Player1].ManaPool.Add(mana.B, 5)
	agent := &sneakAgent{cost: "Sneak"}
	f.cast(t, agent, action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil))
	original, _ := g.Stack.Peek()
	agent.cost = "Normal cost"
	f.cast(t, agent, action.CastSpell(copierID, []game.Target{game.StackObjectTarget(original.ID)}, 0, nil))
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	for {
		top, ok := g.Stack.Peek()
		if !ok || top.SourceID == copierID {
			break
		}
		f.engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	}
	f.engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	copied, _ := g.Stack.Peek()
	if !copied.Copy || !copied.Sneaked || copied.SneakAttackTarget != original.SneakAttackTarget {
		t.Fatalf("copy = %+v, want the original's paid sneak cost and attack target", copied)
	}
	f.resolve(agent)
	var bears []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token || permanent.CardInstanceID == cardID {
			bears = append(bears, permanent)
		}
	}
	if len(bears) != 2 {
		t.Fatalf("Grizzly Bears permanents = %d, want the card and its token copy", len(bears))
	}
	for _, permanent := range bears {
		if !permanent.Tapped || !permanent.SneakPaid || !attackingPlayer(g, permanent, game.Player2) {
			t.Fatalf("permanent %+v is not a tapped, sneak-paid attacker", permanent)
		}
	}
}

// TestNinjaTeenShapeValidates proves the granted sneak declaration validates
// and that a sneak grant must name its fixed mana cost and nothing else, and
// the sneak-cast permission must name the player it affects.
func TestNinjaTeenShapeValidates(t *testing.T) {
	if issues := game.ValidateCardDef(ninjaTeenFixture()); len(issues) != 0 {
		t.Fatalf("Ninja Teen fixture issues = %v", issues)
	}
	for _, tt := range []struct {
		name   string
		mutate func(def *game.CardDef)
	}{
		{name: "sneak grant without mana cost", mutate: func(def *game.CardDef) {
			def.StaticAbilities[0].RuleEffects[0].GraveyardCastCost = game.GraveyardCastGrantCost{}
		}},
		{name: "sneak grant using the card's mana cost", mutate: func(def *game.CardDef) {
			def.StaticAbilities[0].RuleEffects[0].GraveyardCastCost.UseCardManaCost = true
		}},
		{name: "non-sneak grant with a fixed mana cost", mutate: func(def *game.CardDef) {
			def.StaticAbilities[0].RuleEffects[0].GrantedKeyword = game.Flashback
		}},
		{name: "permission without affected player", mutate: func(def *game.CardDef) {
			def.StaticAbilities[1].RuleEffects[0].AffectedPlayer = game.PlayerAny
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			def := ninjaTeenFixture()
			tt.mutate(def)
			if issues := game.ValidateCardDef(def); len(issues) == 0 {
				t.Fatal("invalid sneak declaration validated")
			}
		})
	}
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// Fear of Falling Oracle ID 00ba96e0-42a7-432a-a500-5e607d75a358, printing
// 0e814e48-cd9d-428f-90e2-74d97cb9c8f1: {3}{U}{U}, 4/4, "Flying /
// Whenever this creature attacks, target creature defending player controls
// gets -2/-0 and loses flying until your next turn." Its Scryfall rulings URI
// https://api.scryfall.com/cards/0e814e48-cd9d-428f-90e2-74d97cb9c8f1/rulings
// returned the WotC ruling dated 2024-09-20 (retrieved 2026-11-10): “A
// creature that loses flying due to the effect of Fear of Falling's last
// ability can still gain flying from the effect of another spell or ability,
// such as Jump Scare.” Official CR effective 2026-11-06:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt.
// Read 115.1, 508.5, 603.1-3, 608.2b, 611.2a, 613.1f/g, 613.3/7,
// 613.4c and 702.9a-b: the attack event supplies its defending player, targets
// are chosen and rechecked, the -2/-0 and flying removal are separate layers
// with the stated duration, and a later timestamped effect can add flying.
const fearOfFallingScenarioOracleID = "00ba96e0-42a7-432a-a500-5e607d75a358"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// Short Circuit Oracle ID 5ea0e3a5-9c38-4da4-98a6-41c14be512d9, printing
// 74fe0fb5-e994-41b1-a73d-4e68f7afdf34: {1}{U}, "Flash / Enchant artifact or
// creature / As long as enchanted permanent is a creature, it gets -3/-0 and
// loses flying." Its Scryfall rulings URI
// https://api.scryfall.com/cards/74fe0fb5-e994-41b1-a73d-4e68f7afdf34/rulings
// returned an empty list on 2026-11-10. Official CR effective 2026-11-06:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt.
// Read 303.4a-b, 604.1-4, 613.1f/g and 613.4c, 702.5a and 702.8a: the Aura
// targets and may enchant artifacts or creatures, while its live static
// condition applies the keyword and power effects only to a creature.
const shortCircuitScenarioOracleID = "5ea0e3a5-9c38-4da4-98a6-41c14be512d9"

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// Canopy Dragon Oracle ID a9e59ee4-02a6-498c-82f0-f4f2ff334d65, printing
// 20c5e4d6-b716-4994-b226-b1eb799bec25: {4}{G}{G}, 4/4, "Trample /
// {1}{G}: This creature gains flying and loses trample until end of turn." Its
// Scryfall rulings URI
// https://api.scryfall.com/cards/20c5e4d6-b716-4994-b226-b1eb799bec25/rulings
// returned an empty list on 2026-11-10. Official CR effective 2026-11-06:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt.
// Read 602.1-2, 611.2a, 613.1f, 613.3/7 and 702.9a/702.19a: pay the
// activated cost, add flying and remove trample in layer 6, then expire both
// changes at end of turn.
const canopyDragonScenarioOracleID = "a9e59ee4-02a6-498c-82f0-f4f2ff334d65"

func keywordChangeCard(t *testing.T, oracleID, name string) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("default registry lookup for %s: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("default registry Oracle ID %s returned %q, want %q", oracleID, def.Name, name)
	}
	return def
}

type fearOfFallingTargetPolicy struct {
	simPassPolicy
	target  game.Target
	offered []game.Target
	chosen  bool
}

func (p *fearOfFallingTargetPolicy) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceTarget {
		selected := -1
		for i, option := range request.Options {
			for _, target := range option.Targets {
				p.offered = append(p.offered, target)
				if target == p.target {
					p.chosen = true
					selected = i
				}
			}
		}
		if selected >= 0 {
			return []int{selected}
		}
	}
	return request.DefaultSelection
}

func fearOfFallingFlyingGrantSpell() *game.CardDef {
	const text = "Target creature gains flying until end of turn."
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Test Flying Grant",
		ManaCost:   opt.Val(cost.Mana{cost.U}),
		Types:      []types.Card{types.Instant},
		Colors:     []color.Color{color.Blue},
		OracleText: text,
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target creature",
				Allow:      game.TargetAllowPermanent,
				Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			}},
			Sequence: []game.Instruction{{
				Primitive: game.ApplyContinuous{
					Object: opt.Val(game.TargetPermanentReference(0)),
					ContinuousEffects: []game.ContinuousEffect{{
						Layer:       game.LayerAbility,
						AddKeywords: []game.Keyword{game.Flying},
					}},
					Duration: game.DurationUntilEndOfTurn,
				},
			}},
		}.Ability()),
	}}
}

func TestFearOfFallingAttacksDefendingPlayerAndRestoresFlyingNextTurn(t *testing.T) {
	def := keywordChangeCard(t, fearOfFallingScenarioOracleID, "Fear of Falling")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, evasionTestCreature("Defender's flyer", game.FlyingStaticBody))
	otherDefenderCreature := addCombatPermanent(g, game.Player2, evasionTestCreature("Defender's other creature"))
	otherPlayerCreature := addCombatPermanent(g, game.Player3, evasionTestCreature("Other player's flyer", game.FlyingStaticBody))
	attackerControlledCreature := addCombatPermanent(g, game.Player1, evasionTestCreature("Attacker's flyer", game.FlyingStaticBody))
	fearID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Players[game.Player1].ManaPool.Add(mana.U, 2)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(fearID, nil, 0, nil)) {
		t.Fatal("could not cast Fear of Falling for {3}{U}{U}")
	}
	if g.Players[game.Player1].ManaPool.Amount(mana.C) != 0 ||
		g.Players[game.Player1].ManaPool.Amount(mana.U) != 0 {
		t.Fatal("Fear of Falling did not pay {3}{U}{U}")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	attacker, ok := findPermanentByCardID(g, fearID)
	if !ok || !hasKeyword(g, attacker, game.Flying) || !hasKeyword(g, target, game.Flying) {
		t.Fatal("Fear of Falling did not resolve as a flying creature alongside its flying target")
	}

	// Let the creature reach its controller's next turn so it can attack.
	createdTurn := g.Turn.TurnNumber
	engine.runEndingPhase(g, simPassPolicies())
	g.Turn.TurnNumber = createdTurn + 1
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Turn-start draw"}})
	engine.runBeginningPhase(g, simPassPolicies(), &TurnLog{})
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	declareAttack(t, g, engine, attacker, game.Player2)
	policy := &fearOfFallingTargetPolicy{target: game.PermanentTarget(target.ObjectID)}
	agents := simPassPolicies()
	agents[game.Player1] = policy
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Fear of Falling's attack trigger was not put on the stack")
	}
	allowedTargets := map[game.Target]bool{
		game.PermanentTarget(target.ObjectID):                true,
		game.PermanentTarget(otherDefenderCreature.ObjectID): true,
	}
	if !policy.chosen || len(policy.offered) != len(allowedTargets) {
		t.Fatalf("trigger offered targets %v and chose=%v, want both of the defending player's creatures",
			policy.offered, policy.chosen)
	}
	for _, offered := range policy.offered {
		if !allowedTargets[offered] {
			t.Fatalf("trigger offered non-defending player's creature %v", offered)
		}
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := effectivePower(g, target); got != 1 {
		t.Fatalf("defending player's target power = %d, want 1 after -2/-0", got)
	}
	if hasKeyword(g, target, game.Flying) {
		t.Fatal("Fear of Falling did not remove the defending player's target's flying")
	}
	for _, unaffected := range []*game.Permanent{otherPlayerCreature, attackerControlledCreature} {
		if got := effectivePower(g, unaffected); got != 3 || !hasKeyword(g, unaffected, game.Flying) {
			t.Fatalf("non-defending creature changed: power=%d flying=%v, want 3 and flying",
				got, hasKeyword(g, unaffected, game.Flying))
		}
	}

	grantID := addCardToHand(g, game.Player1, fearOfFallingFlyingGrantSpell())
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(
		grantID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
	)) {
		t.Fatal("could not cast the later flying-grant spell on the affected creature")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !hasKeyword(g, target, game.Flying) || effectivePower(g, target) != 1 {
		t.Fatal("later flying grant did not coexist with Fear of Falling's -2/-0 until end of turn")
	}

	createdTurn = g.Turn.TurnNumber
	engine.runEndingPhase(g, simPassPolicies())
	if hasKeyword(g, target, game.Flying) || effectivePower(g, target) != 1 {
		t.Fatal("Fear of Falling's effect did not persist after the temporary flying grant expired")
	}
	g.Turn.TurnNumber = createdTurn + 1
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Next-turn draw"}})
	engine.runBeginningPhase(g, simPassPolicies(), &TurnLog{})
	if !hasKeyword(g, target, game.Flying) || effectivePower(g, target) != 3 {
		t.Fatal("Fear of Falling's -2/-0 and flying removal did not expire on its controller's next turn")
	}
}

func TestShortCircuitEnchantsArtifactsAndCreaturesConditionally(t *testing.T) {
	def := keywordChangeCard(t, shortCircuitScenarioOracleID, "Short Circuit")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:            "Test noncreature Vehicle with flying",
		Types:           []types.Card{types.Artifact},
		Subtypes:        []types.Sub{types.Vehicle},
		Power:           opt.Val(game.PT{Value: 4}),
		Toughness:       opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
	}})
	creature := addCombatPermanent(g, game.Player2, evasionTestCreature("Flying creature", game.FlyingStaticBody))
	artifactAuraID := addCardToHand(g, game.Player1, def)
	creatureAuraID := addCardToHand(g, game.Player1, def)

	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.CastSpell(
		artifactAuraID, []game.Target{game.PermanentTarget(artifact.ObjectID)}, 0, nil,
	)) {
		t.Fatal("Flash did not allow Short Circuit to enchant an artifact during the opponent's main phase")
	}
	if g.Players[game.Player1].ManaPool.Amount(mana.U) != 0 ||
		g.Players[game.Player1].ManaPool.Amount(mana.C) != 0 {
		t.Fatal("Short Circuit did not pay {1}{U} for the artifact Aura")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	artifactAura, ok := findPermanentByCardID(g, artifactAuraID)
	if !ok || !artifactAura.AttachedTo.Exists || artifactAura.AttachedTo.Val != artifact.ObjectID {
		t.Fatal("Short Circuit did not attach to its legal artifact target")
	}
	if effectivePower(g, artifact) != 4 || !hasKeyword(g, artifact, game.Flying) {
		t.Fatal("Short Circuit affected the noncreature artifact")
	}

	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(
		creatureAuraID, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil,
	)) {
		t.Fatal("Short Circuit could not enchant a creature")
	}
	if g.Players[game.Player1].ManaPool.Amount(mana.U) != 0 ||
		g.Players[game.Player1].ManaPool.Amount(mana.C) != 0 {
		t.Fatal("Short Circuit did not pay {1}{U} for the creature Aura")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	creatureAura, ok := findPermanentByCardID(g, creatureAuraID)
	if !ok || !creatureAura.AttachedTo.Exists || creatureAura.AttachedTo.Val != creature.ObjectID {
		t.Fatal("Short Circuit did not attach to its legal creature target")
	}
	if got := effectivePower(g, creature); got != 0 {
		t.Fatalf("enchanted creature power = %d, want 0 after -3/-0", got)
	}
	if hasKeyword(g, creature, game.Flying) {
		t.Fatal("Short Circuit did not remove flying from the enchanted creature")
	}
}

func TestCanopyDragonGainsFlyingAndLosesTrampleUntilEndOfTurn(t *testing.T) {
	def := keywordChangeCard(t, canopyDragonScenarioOracleID, "Canopy Dragon")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	dragonID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.G, 2)
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(dragonID, nil, 0, nil)) {
		t.Fatal("could not cast Canopy Dragon for {4}{G}{G}")
	}
	if g.Players[game.Player1].ManaPool.Amount(mana.G) != 0 ||
		g.Players[game.Player1].ManaPool.Amount(mana.C) != 0 {
		t.Fatal("Canopy Dragon did not pay {4}{G}{G}")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	dragon, ok := findPermanentByCardID(g, dragonID)
	if !ok {
		t.Fatal("Canopy Dragon did not resolve to the battlefield")
	}
	if !hasKeyword(g, dragon, game.Trample) {
		t.Fatal("Canopy Dragon did not start with trample")
	}
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(dragon.ObjectID, 0, nil, 0)) {
		t.Fatal("Canopy Dragon's {1}{G} ability was not accepted")
	}
	if g.Players[game.Player1].ManaPool.Amount(mana.G) != 0 ||
		g.Players[game.Player1].ManaPool.Amount(mana.C) != 0 {
		t.Fatal("Canopy Dragon did not pay {1}{G}")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !hasKeyword(g, dragon, game.Flying) || hasKeyword(g, dragon, game.Trample) {
		t.Fatalf("after activation, flying=%v trample=%v; want flying and no trample",
			hasKeyword(g, dragon, game.Flying), hasKeyword(g, dragon, game.Trample))
	}

	engine.runEndingPhase(g, simPassPolicies())
	if hasKeyword(g, dragon, game.Flying) || !hasKeyword(g, dragon, game.Trample) {
		t.Fatalf("after end of turn, flying=%v trample=%v; want no flying and restored trample",
			hasKeyword(g, dragon, game.Flying), hasKeyword(g, dragon, game.Trample))
	}
}

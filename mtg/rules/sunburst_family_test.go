package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: pinned Solarion Oracle revision 9070dd3e-7600-493e-84b9-931ac92f44bc,
// Oracle ID 4022e540-8fa0-4eee-bcf1-df5046289f2b. Current Scryfall rulings retrieved
// 2026-11-21 from https://api.scryfall.com/cards/9070dd3e-7600-493e-84b9-931ac92f44bc/rulings:
// the 2004-12-01 ruling says "Double" has its normal English meaning, so three counters
// become six. Official CR effective 2026-11-06, verified at https://magic.wizards.com/en/rules
// and https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt (SHA-256
// 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829); read 601.2h,
// 602.2, 614.1c, 614.12, 122.6, 701.10e, and 702.44a-b.
func TestSolarionSunburstCountersAndActivationDoubleThem(t *testing.T) {
	g, engine, solarion := castSunburstCard(t, "4022e540-8fa0-4eee-bcf1-df5046289f2b", "Solarion",
		[]mana.Color{mana.W, mana.W, mana.U, mana.U, mana.U, mana.R, mana.R})
	if got := solarion.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("Solarion entered with %d +1/+1 counters, want one for each of three colors", got)
	}

	solarion.SummoningSick = false
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(solarion.ObjectID, 0, nil, 0)) {
		t.Fatal("Solarion's {T} ability could not be activated")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := solarion.Counters.Get(counter.PlusOnePlusOne); got != 6 {
		t.Fatalf("Solarion has %d +1/+1 counters after resolving its ability, want 6", got)
	}
	if !solarion.Tapped {
		t.Fatal("Solarion's {T} ability did not tap it")
	}
}

// Source review: pinned Skyreach Manta Oracle revision 3df1e8a0-6a54-4be9-bd0c-ac70f0ae9aa6,
// Oracle ID 58eab745-9afa-4db7-9d8a-f5eb0b9dca6c. Current Scryfall rulings retrieved
// 2026-11-21 from https://api.scryfall.com/cards/3df1e8a0-6a54-4be9-bd0c-ac70f0ae9aa6/rulings:
// none. Official CR 2026-11-06, the verified edition and TXT checksum recorded above;
// read 601.2h, 614.1c, 614.12, 122.6, 702.9a, and 702.44a-b.
func TestSkyreachMantaSunburstAndFlying(t *testing.T) {
	g, _, manta := castSunburstCard(t, "58eab745-9afa-4db7-9d8a-f5eb0b9dca6c", "Skyreach Manta",
		[]mana.Color{mana.W, mana.W, mana.W, mana.U, mana.U})
	if got := manta.Counters.Get(counter.PlusOnePlusOne); got != 2 || effectivePower(g, manta) != 2 {
		t.Fatalf("Skyreach Manta counters=%d power=%d, want 2 counters and 2 power", got, effectivePower(g, manta))
	}
	if !hasKeyword(g, manta, game.Flying) {
		t.Fatal("Skyreach Manta did not have flying")
	}
}

type sunburstManaChoiceAgent struct {
	requests []game.ChoiceRequest
}

func (sunburstManaChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *sunburstManaChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	if len(request.Options) == 5 {
		return []int{request.Options[4].Index}
	}
	return request.DefaultSelection
}

// Source review: pinned Pentad Prism Oracle revision e4beb84e-9eb9-48a2-932a-91fbbf218b5b,
// Oracle ID 72e8d67a-41e6-4074-bcbd-b54ead995237. Current Scryfall rulings retrieved
// 2026-11-21 from https://api.scryfall.com/cards/e4beb84e-9eb9-48a2-932a-91fbbf218b5b/rulings:
// two rulings dated 2020-08-07; Sunburst checks the actual mana spent, even when mana is
// spent as though it had another color, and the Prism remains on the battlefield without
// counters. Official CR 2026-11-06, verified edition/checksum above; read 601.2h, 605.1a,
// 614.1c, 614.12, 122.6, and 702.44a-b.
func TestPentadPrismSunburstAndChosenColorMana(t *testing.T) {
	g, engine, prism := castSunburstCard(t, "72e8d67a-41e6-4074-bcbd-b54ead995237", "Pentad Prism",
		[]mana.Color{mana.W, mana.U})
	if got := prism.Counters.Get(counter.Charge); got != 2 {
		t.Fatalf("Pentad Prism entered with %d charge counters, want 2", got)
	}

	agent := &sunburstManaChoiceAgent{}
	agents := simPassPolicies()
	agents[game.Player1] = agent
	setMainPhasePriority(g, game.Player1)
	if !engine.applyActionWithChoices(g, game.Player1,
		action.ActivateAbility(prism.ObjectID, 0, nil, 0), agents, &TurnLog{}) {
		t.Fatal("Pentad Prism could not remove a charge counter and produce mana")
	}
	if len(agent.requests) != 1 || len(agent.requests[0].Options) != 5 {
		t.Fatalf("Pentad Prism color choice requests = %+v, want one five-color choice", agent.requests)
	}
	if got := prism.Counters.Get(counter.Charge); got != 1 {
		t.Fatalf("Pentad Prism has %d charge counters after activating, want 1", got)
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.G); got != 1 {
		t.Fatalf("Pentad Prism produced %d green mana, want 1", got)
	}
}

// Source review: pinned Suntouched Myr Oracle revision 4c5d4ae3-bc7e-4e60-a5e6-c3ae349fe6fc,
// Oracle ID 75d4f61a-8276-43ad-99e6-0d756a597e4b. Current Scryfall rulings retrieved
// 2026-11-21 from https://api.scryfall.com/cards/4c5d4ae3-bc7e-4e60-a5e6-c3ae349fe6fc/rulings:
// none. Official CR 2026-11-06, verified edition/checksum recorded above; read 601.2h,
// 614.1c, 614.12, 122.1a, 122.6, and 702.44a-b.
func TestSuntouchedMyrSunburstCounters(t *testing.T) {
	g, _, myr := castSunburstCard(t, "75d4f61a-8276-43ad-99e6-0d756a597e4b", "Suntouched Myr",
		[]mana.Color{mana.R, mana.B, mana.G})
	if got := myr.Counters.Get(counter.PlusOnePlusOne); got != 3 || effectivePower(g, myr) != 3 {
		t.Fatalf("Suntouched Myr counters=%d power=%d, want 3 counters and 3 power", got, effectivePower(g, myr))
	}
}

// Source review: pinned Etched Oracle revision 1b4e70c6-8f74-4b0b-9ad7-33fce05c401b,
// Oracle ID 7ecfa47e-1165-46a6-884c-290b1c14d020. Current Scryfall rulings retrieved
// 2026-11-21 from https://api.scryfall.com/cards/1b4e70c6-8f74-4b0b-9ad7-33fce05c401b/rulings:
// none. Official CR 2026-11-06, verified edition/checksum recorded above; read 113.7a,
// 115.1, 118.3a, 121.1, 602.2, 614.1c, 614.12, 122.1a, 122.6, 702.44a-b, and 704.5f.
func TestEtchedOracleSunburstAndTargetedDrawActivation(t *testing.T) {
	g, engine, oracle := castSunburstCard(t, "7ecfa47e-1165-46a6-884c-290b1c14d020", "Etched Oracle",
		[]mana.Color{mana.W, mana.U, mana.B, mana.R})
	if got := oracle.Counters.Get(counter.PlusOnePlusOne); got != 4 {
		t.Fatalf("Etched Oracle entered with %d +1/+1 counters, want 4", got)
	}
	first := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Etched Oracle draw 1"}})
	second := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Etched Oracle draw 2"}})
	third := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Etched Oracle draw 3"}})
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(
		oracle.ObjectID, 0, []game.Target{game.PlayerTarget(game.Player2)}, 0,
	)) {
		t.Fatal("Etched Oracle could not pay {1}, remove four counters, and target a player")
	}
	if oracle.Counters.Get(counter.PlusOnePlusOne) != 0 {
		t.Fatalf("Etched Oracle retained %d +1/+1 counters after paying its cost", oracle.Counters.Get(counter.PlusOnePlusOne))
	}
	engine.applyStateBasedActions(g)
	if _, stillOnBattlefield := permanentByObjectID(g, oracle.ObjectID); stillOnBattlefield {
		t.Fatal("Etched Oracle remained on the battlefield after losing its last counter")
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Hand.Contains(first) ||
		!g.Players[game.Player2].Hand.Contains(second) ||
		!g.Players[game.Player2].Hand.Contains(third) {
		t.Fatal("Etched Oracle's target player did not draw all three cards")
	}
	if g.Players[game.Player1].Hand.Size() != 0 {
		t.Fatal("Etched Oracle drew cards for its controller instead of the target player")
	}
}

func castSunburstCard(t *testing.T, oracleID, name string, manaSpent []mana.Color) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("default registry lookup for %s (%s): %v", name, oracleID, err)
	}
	if def.Name != name {
		t.Fatalf("Oracle ID %s resolved to %q, want %q", oracleID, def.Name, name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for _, color := range manaSpent {
		g.Players[game.Player1].ManaPool.Add(color, 1)
	}
	cardID := addCardToHand(g, game.Player1, def)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatalf("casting %s with %d mana failed", name, len(manaSpent))
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent, ok := findPermanentByCardID(g, cardID)
	if !ok {
		t.Fatalf("%s did not resolve to the battlefield", name)
	}
	return g, engine, permanent
}

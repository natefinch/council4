package agent

// Archetype-derived personality tuning (playbook §1.4). A deck's game plan
// implies how it should be piloted, so DeckPersonality maps the coarse archetype
// from AnalyzeDeck to Personality knobs, letting each seat play in its deck's
// style instead of every agent playing identically. The values are deliberately
// moderate; the zero Personality (a plain generic strategy) remains the neutral
// baseline for a midrange deck.
const (
	// aggroAggression makes an aggressive deck attack and deploy threats harder.
	aggroAggression = 2.0
	// tokensAggression pushes a go-wide deck to develop and swing, a little less
	// recklessly than a dedicated aggro deck.
	tokensAggression = 1.5
	// aristocratsAggression presses a sacrifice deck's board without the
	// all-in bias of aggro, since its payoff is attrition rather than combat.
	aristocratsAggression = 1.0
	// controlPolitics makes a control deck weight the table's biggest threat more
	// heavily when aiming interaction, focusing the real problem.
	controlPolitics = 1.5
	// rampPolitics gives a ramp deck a mild threat focus while it out-values the
	// table, without the aggression of a beatdown deck.
	rampPolitics = 0.5
)

// DeckPersonality returns the Personality that suits a deck's archetype, so an
// agent piloting the deck plays to its game plan: aggro and go-wide decks press
// damage, aristocrats grinds a board, and control and ramp decks focus the
// biggest threat rather than racing. A midrange deck keeps the neutral zero
// Personality.
func DeckPersonality(profile DeckProfile) Personality {
	switch profile.Archetype {
	case ArchetypeAggro:
		return Personality{Aggression: aggroAggression}
	case ArchetypeTokens:
		return Personality{Aggression: tokensAggression}
	case ArchetypeAristocrats:
		return Personality{Aggression: aristocratsAggression}
	case ArchetypeControl:
		return Personality{PoliticsWeight: controlPolitics}
	case ArchetypeRamp:
		return Personality{PoliticsWeight: rampPolitics}
	default:
		return Personality{}
	}
}

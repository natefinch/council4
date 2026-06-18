# Parser Coverage

Parser-only coverage across the eligible Scryfall corpus, measured without running the compiler or lowering. Two distinct metrics are reported:

- **Parser-complete (typed coverage):** every must-cover token of every ability is accounted for by a kind-recognized typed element. This is an upper bound on what the lowerer could consume — it does not require byte-exact reconstruction.
- **Exact round-trip:** the parser reconstructs the original text byte-for-byte (`effect.Exact`). Strictly stronger than typed coverage.

Regenerate with `mage parserCoverage`.

## Headline

- Eligible cards: 31838
- Parser-complete cards (typed coverage): 15430 (48.46%)
- Exact round-trip cards (complete and every effect exact): 7230 (22.71%)
- Resolving effects: 55206
- Exact round-trip effects: 18508 (33.53%)

## Generated ⊆ Parser-complete

- Generated cards: 6690
- Violations: 0

## Uncovered components by blocker

| Blocker | Components |
| --- | --- |
| effect | 16613 |
| condition | 6035 |
| trigger | 2642 |
| cost | 142 |
| modal | 18 |

## Uncovered grammar work queue

Top uncovered span clusters (normalized), ranked by occurrence.

| Rank | Count | Blocker | Cluster | Examples |
| --- | --- | --- | --- | --- |
| 1 | 1028 | condition | if you do | Keldon Raider; Witch's Mark; Duplicity; Minion Reflector; Brawl-Bash Ogre |
| 2 | 170 | condition | if able | Impetuous Devils; The Foretold Soldier; Nacatl Hunt-Pride; Culling Mark; Legion Warboss |
| 3 | 123 | effect | you may choose new targets for the copy. | Verrak, Warped Sengir; Abstruse Archaic; League Guildmage; Psychic Rebuttal; Melek, Izzet Paragon |
| 4 | 103 | effect | this ability triggers only once each turn. | Fang, Fearless l'Cie; Twilight Diviner; Nanoform Sentinel; Cloaked Cadet; G'raha Tia |
| 5 | 99 | condition | if this spell was kicked | Strength of Night; Goblin Barrage; Colossal Growth; Overload; Stall for Time |
| 6 | 98 | effect | crew N (tap any number of creatures you control with total power N or more: this vehicle becomes an artifact creature until end of turn.) | War Balloon; Flywheel Racer; Mukotai Soulripper; Skybox Ferry; Sidequest: Card Collection // Magicked Card |
| 7 | 92 | effect | it's still a land. | Llanowar Loamspeaker; Hall of Storm Giants; Restless Vinestalk; Vastwood Animist; Faerie Conclave |
| 8 | 85 | trigger | when you cast this spell | Empyrial Storm; The Fourteenth Doctor; Ulamog, the Ceaseless Hunger; Temporal Extortion; Malicious Affliction |
| 9 | 82 | effect | it can't be regenerated. | Polymorph; Pongify; Rapid Hybridization; Phage the Untouchable; Pillage |
| 10 | 80 | effect | you may pay{1}. | Ruthless Sniper; Smolder Initiate; Wooden Sphere; Rebellion of the Flamekin; Azorius Aethermage |
| 11 | 68 | effect | level N | Stormchaser's Talent; Builder's Talent; Leader's Talent; Sorcerer Class; Fortune Teller's Talent |
| 12 | 67 | effect | crew N | Mindlink Mech; Mighty Servant of Leuk-o; _____ _____ Rocketship; The Lunar Whale; Rocketeer Boostbuggy |
| 13 | 63 | trigger | whenever this creature enters or attacks | Omnivorous Flytrap; Graveyard Trespasser // Graveyard Glutton; Inferno Titan; Sigarda's Vanguard; Cemetery Illuminator |
| 14 | 61 | effect | changeling (this card is every creature type.) | Barkform Harvester; Guardian Gladewalker; Wings of Velis Vel; Firdoch Core; Mirror Entity |
| 15 | 57 | condition | if you search your library this way | Vraska's Scorn; Ashiok's Forerunner; Niambi, Faithful Healer; Claim Jumper; Fang-Druid Summoner |
| 16 | 56 | effect | prevent the next N damage that would be dealt to any target this turn. | Heal; Master Apothecary; Barrenton Medic; Rakalite; Militant Monk |
| 17 | 55 | effect | partner (you can have two commanders if both have partner.) | Ghost of Ramirez DePietro; Silas Renn, Seeker Adept; Krark, the Thumbless; Vial Smasher the Fierce; Francisco, Fowl Marauder |
| 18 | 52 | effect | enchant creature you control | One with the Kami; Endless Evil; Inferno Fist; Pitiless Fists; Dying Wish |
| 19 | 49 | trigger | when you cycle this card | Deem Worthy; Rampaging War Mammoth; Shefet Monitor; Agonasaur Rex; Windcaller Aven |
| 20 | 46 | trigger | whenever you cast a spell that targets this creature | Akroan Line Breaker; Lagonna-Band Trailblazer; Hero of Iroas; War-Wing Siren; Triton Cavalry |
| 21 | 45 | effect | target creature can't block this turn. | Nacatl Hunt-Pride; Mardu Roughrider; Bola Warrior; Unstoppable Ogre; Goblin Shortcutter |
| 22 | 45 | trigger | whenever you cast your second spell each turn | Sunstar Lightsmith; Kraum, Violent Cacophony; Monk of the Open Hand; Illvoi Operative; Cori Mountain Stalwart |
| 23 | 44 | effect | flip a coin. | Goblin Lyre; Ydwen Efreet; Ral, Monsoon Mage // Ral, Leyline Prodigy; Krark, the Thumbless; Risky Move |
| 24 | 41 | effect | you may play that card this turn. | Party Thrasher; Professional Face-Breaker; Geistflame Reservoir; Dark-Dweller Oracle; Tablet of Discovery |
| 25 | 40 | condition | if it's a land card | Lantern of Revealing; Unexpected Results; Traveling Botanist; Countryside Crusher; Raiders' Karve |
| 26 | 40 | effect | the ring tempts you. | Uruk-hai Berserker; Horses of the Bruinen; Fiery Inscription; Relentless Rohirrim; Shortcut to Mushrooms |
| 27 | 40 | effect | you become the monarch. | Custodi Lich; Court of Ire; Palace Jailer; Grave Venerations; Court of Garenbrig |
| 28 | 39 | effect | they can't be regenerated. | Plague Wind; Wave of Terror; Wrath of God; Obliterate; Tsabo's Decree |
| 29 | 39 | effect | you may pay{2}. | Minion Reflector; Esoteric Duplicator; Kavaron Harrier; Unassuming Sage; Terra, Herald of Hope |
| 30 | 38 | effect | activate only during your turn. | Ruthless Waterbender; Humble Defector; Yes Man, Personal Securitron; Hoofprints of the Stag; June, Bounty Hunter |
| 31 | 38 | condition | if you lose the flip | Goblin Bomb; Goblin Lyre; Ydwen Efreet; Ral, Monsoon Mage // Ral, Leyline Prodigy; Krark, the Thumbless |
| 32 | 38 | effect | swampwalk (this creature can't be blocked as long as defending player controls a swamp.) | Whispering Shade; Dirtwater Wraith; Slithery Stalker; Lost Soul; Quag Vampires |
| 33 | 38 | trigger | whenever this creature or another ally you control enters | Kazandu Blademaster; Murasa Pyromancer; Grovetender Druids; Tuktuk Grunts; Firemantle Mage |
| 34 | 37 | effect | fear (this creature can't be blocked except by artifact creatures and/or black creatures.) | Guiltfeeder; Commander Greven il-Vec; Dread; Lingering Tormentor; Undercity Shade |
| 35 | 37 | condition | if this creature was kicked | Kavu Primarch; Faerie Squadron; Aether Figment; Skyclave Shade; Skyclave Sentinel |
| 36 | 37 | effect | start your engines ! (if you have no speed, it starts at 1. it increases once on each of your turns when an opponent loses life. max speed is 4.) | Momentum Breaker; Point the Way; Lightwheel Enhancements; Burnout Bashtronaut; Goblin Surveyor |
| 37 | 36 | effect | shadow (this creature can block or be blocked by only creatures with shadow.) | Dauthi Horror; Dauthi Mercenary; Thalakos Seer; Augur il-Vec; Soltari Guerrillas |
| 38 | 35 | effect | affinity for artifacts (this spell costs{1}less to cast for each artifact you control.) | Refurbished Familiar; Into Thin Air; Assert Authority; Valkyrie Aerial Unit; Furnace Dragon |
| 39 | 35 | effect | choose one. | Wail of the Forgotten; Jeska's Will; See Double; Inscription of Insight; Will of the Abzan |
| 40 | 35 | condition | if you win the flip | Goblin Bomb; Goblin Lyre; Ral, Monsoon Mage // Ral, Leyline Prodigy; Krark, the Thumbless; Plasma Caster |


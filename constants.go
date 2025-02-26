package main

const (
	maxChecksAtListPage int = 9
	maxCheckBtnInRow    int = 3
)

// commands
const (
	start    string = "start"
	addWhite string = "white"
	addRed   string = "red"
	seeTop   string = "top"
	seeLog   string = "log"
)

// skill identifiers
const (
	intLogic = iota + 1
	intEncyclopedia
	intRhetoric
	intDrama
	intConcept
	intVisual
	psyVolition
	psyInland
	psyEmpathy
	psyAuthority
	psyEsprit
	psySuggestion
	phyEndurance
	phyPain
	phyInstrument
	phyElectrochem
	phyShivers
	phyHalflight
	motCoordintation
	motPerception
	motReaction
	motSavoir
	motInterfacing
	motComposure
)

// skill texts
var skillNames = [25]string{
	"",
	"🟦 Logic",
	"🟦 Encyclopedia",
	"🟦 Rhetoric",
	"🟦 Drama",
	"🟦 Conceptualization",
	"🟦 Visual Calculus",
	"🟪 Volition",
	"🟪 Inland Empire",
	"🟪 Empathy",
	"🟪 Authority",
	"🟪 Esprit De Corps",
	"🟪 Suggestion",
	"🟥 Endurance",
	"🟥 Pain Threshold",
	"🟥 Physical Instrument",
	"🟥 Electrochemistry",
	"🟥 Shivers",
	"🟥 Half Light",
	"🟨 Hand/Eye Coordination",
	"🟨 Perception",
	"🟨 Reaction Speed",
	"🟨 Savoir Faire",
	"🟨 Interfacing",
	"🟨 Composure",
}

var skillDescriptions = [25]string{}

// skill difficulty identifiers
const (
	difTrivial = iota + 1
	difEasy
	difMedium
	difChallenging
	difFormidable
	difLegendary
	difHeroic
	difGodly
	difImpossible
)

// skill difficulty texts
var difficultyNames = [10]string{
	"",
	"Trivial",
	"Easy",
	"Medium",
	"Challenging",
	"Formidable",
	"Legendary",
	"Heroic",
	"Godly",
	"Impossible",
}

// check result identifiers
const (
	resDefault = iota
	resCanceled
	resFailure
	resSuccess
)

// check result texts
var resultNames = [4]string{
	"",
	"Cancel 🚫",
	"Failure 🔴",
	"Success 🟢",
}

const (
	typNonRetriable = iota + 1
	typRetriable
)

// check type texts
var typeNames = [3]string{
	"",
	"Red check",
	"White check",
}

const (
	listCheckDetail = iota
	listCheckForward
	listCheckBackward
	listCheckAction
)

package shared

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

type Language string

const (
	LanguageRU Language = "ru"
	LanguageEN Language = "en"
)

type NsfwLevel string

const (
	NsfwLevelSafe  NsfwLevel = "safe"
	NsfwLevelAdult NsfwLevel = "adult"
)

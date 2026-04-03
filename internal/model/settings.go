package model

type Settings struct {
	WorkMinutes        int
	ShortBreakMinutes  int
	LongBreakMinutes   int
	LongBreakInterval  int
	AutoStartNextPhase bool
	SoundEnabled       bool
	Theme              string
	Language           string
	WindowWidth        float32
	WindowHeight       float32
}

func DefaultSettings() Settings {
	return Settings{
		WorkMinutes:        25,
		ShortBreakMinutes:  5,
		LongBreakMinutes:   15,
		LongBreakInterval:  4,
		AutoStartNextPhase: false,
		SoundEnabled:       true,
		Theme:              "dark",
		Language:           "zh",
		WindowWidth:        420,
		WindowHeight:       760,
	}
}

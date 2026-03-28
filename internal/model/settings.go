package model

type Settings struct {
	WorkMinutes        int
	ShortBreakMinutes  int
	LongBreakMinutes   int
	LongBreakInterval  int
	AutoStartNextPhase bool
}

func DefaultSettings() Settings {
	return Settings{
		WorkMinutes:        25,
		ShortBreakMinutes:  5,
		LongBreakMinutes:   15,
		LongBreakInterval:  4,
		AutoStartNextPhase: false,
	}
}

package logger

type Option func(*Logger)

func FileName(v string) Option {
	return func(l *Logger) {
		l.FileName = v
	}
}

func MaxFileSize(v int) Option {
	return func(l *Logger) {
		l.MaxFileSize = v
	}
}

func MaxFileAge(v int) Option {
	return func(l *Logger) {
		l.MaxFileAge = v
	}
}

func Level(v string) Option {
	return func(l *Logger) {
		l.Level = v
	}
}

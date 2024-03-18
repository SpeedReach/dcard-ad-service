package models

type Platform string

const (
	Android Platform = "android"
	Ios     Platform = "ios"
	Web     Platform = "web"
)

func ValidPlatform(platform Platform) bool {
	switch platform {
	case Android, Ios, Web:
		return true
	}
	return false
}

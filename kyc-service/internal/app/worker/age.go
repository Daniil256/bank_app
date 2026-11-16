package worker

import "time"

func CheckAge(birthDate time.Time) bool {
	now := time.Now()
	age := now.Year() - birthDate.Year()
	if now.YearDay() < birthDate.YearDay() {
		age--
	}
	return age >= 18
}

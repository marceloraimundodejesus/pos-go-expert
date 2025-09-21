package core

import "regexp"

var cepRe = regexp.MustCompile(`^\d{8}$`)

func IsValidCEP(cep string) bool {
	return cepRe.MatchString(cep)
}

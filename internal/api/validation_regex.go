package api

import "regexp"

var recoveryCodeRegex = regexp.MustCompile(`^OVUM-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$`)
var passwordLengthRegex = regexp.MustCompile(`^.{8,}$`)
var passwordUpperRegex = regexp.MustCompile(`\p{Lu}`)
var passwordLowerRegex = regexp.MustCompile(`\p{Ll}`)
var passwordDigitRegex = regexp.MustCompile(`\d`)

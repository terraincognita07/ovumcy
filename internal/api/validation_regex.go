package api

import "regexp"

var hexColorRegex = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
var recoveryCodeRegex = regexp.MustCompile(`^LUME-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$`)
var passwordLengthRegex = regexp.MustCompile(`^.{8,}$`)
var passwordUpperRegex = regexp.MustCompile(`\p{Lu}`)
var passwordLowerRegex = regexp.MustCompile(`\p{Ll}`)
var passwordDigitRegex = regexp.MustCompile(`\d`)

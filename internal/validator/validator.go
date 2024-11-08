package validator

import "regexp"

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// create validator type which maps validation error
type Validator struct {
	Errors map[string]string
}

// New is a helper to create new validator instance with empty error
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// valid return true if the errors map doesn't contain any entry
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// Check add an error message to the map only if a validation check is not 'ok'.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// add error function add errors to validator map  as long as they don't exist in the map
func (v *Validator) AddError(key, message string) {
	if _, exist := v.Errors[key]; !exist {
		v.Errors[key] = message
	}
}

// check if the specific value is in the list of string
func In(value string, list ...string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

// Match true if a string value matches a specific regexp patttern
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// returns true if the all the string in slice are unique
func Unique(values []string) bool {
	uniqueMap := make(map[string]bool)

	for _, val := range values {
		uniqueMap[val] = true
	}
	return len(uniqueMap) == len(values)
}

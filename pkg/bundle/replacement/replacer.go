package replacement

// Replacer replaces the values of fields matched by a selector.
type Replacer interface {
	Replace(source string, selector string, value string) (string, error)
}

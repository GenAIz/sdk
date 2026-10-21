package adapt

// CliArguments implies that the implementation can provide its fields in a CLI compatible slice. Which can be used as arguments to fork a process.
type CliArguments interface {
	// CliArguments returns a list of arguments which represent the Arguments instance as a slice of strings to use with a forked process
	CliArguments() []string
}

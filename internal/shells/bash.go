package shells

// Bash is the bash-specific shell.
type Bash struct{ Generic }

func (b *Bash) FriendlyName() string { return "Bash" }
func (b *Bash) Info() string         { return "Bash" }

func (b *Bash) AppAlias(name string) string {
	return `alias ` + name + `='eval $(TF_ALIAS=` + name + ` gofuck "$(fc -ln -1)")'`
}

package shells

// Zsh is the zsh-specific shell.
type Zsh struct{ Generic }

func (z *Zsh) FriendlyName() string { return "ZSH" }
func (z *Zsh) Info() string         { return "ZSH" }

func (z *Zsh) AppAlias(name string) string {
	return `alias ` + name + `='eval $(TF_ALIAS=` + name + ` gofuck $(fc -ln -1 | tail -n 1))'`
}

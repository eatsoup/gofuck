package shells

// Fish is the fish-specific shell.
type Fish struct{ Generic }

func (f *Fish) FriendlyName() string { return "Fish Shell" }
func (f *Fish) Info() string         { return "Fish Shell" }

func (f *Fish) AppAlias(name string) string {
	return `function ` + name + ` -d "Correct your previous console command"
  set -l fucked_up_command $history[1]
  env TF_ALIAS=` + name + ` gofuck $fucked_up_command | read -l unfucked_command
  if [ "$unfucked_command" != "" ]
    eval $unfucked_command
    builtin history delete --exact --case-sensitive -- $fucked_up_command
    builtin history merge ^ /dev/null
  end
end`
}

func (f *Fish) And(cmds ...string) string {
	return join(cmds, "; and ")
}
func (f *Fish) Or(cmds ...string) string {
	return join(cmds, "; or ")
}

func join(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += sep + p
	}
	return out
}

package termcolor

const (
    Reset = "\x1b[0m"
    Red   = "\x1b[31m"
    Cyan  = "\x1b[36m"
)

// Maybe wraps s with color code if enable is true.
func Maybe(s, colorCode string, enable bool) string {
    if !enable { return s }
    return colorCode + s + Reset
}


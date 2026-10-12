package mtreeapp

type mirrorFile struct {
	source string
	target string
	rel    string // path relative to the source root
}

type mirrorDir struct {
	source string
	target string
	files  []mirrorFile
}

type mirrorPlan struct {
	source string
	target string
	filter string
	dirs   []mirrorDir
}

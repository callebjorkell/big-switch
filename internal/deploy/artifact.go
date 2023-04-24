package deploy

type Artifacts struct {
	Service string
	Prod    Artifact
	Dev     Artifact
}

func (a Artifacts) IsProdBehind() bool {
	if a.Prod.Name == "" || a.Prod.Time == 0 {
		// nothing currently deployed in prod
		return false
	}
	if a.Prod.Name == a.Dev.Name {
		// same artifact in dev and prod
		return false
	}

	if a.Prod.Time >= a.Dev.Time {
		// prod is ahead of dev
		return false
	}

	return true
}

func (a Artifacts) EnvironmentsMatch() bool {
	return a.Prod.Equals(a.Dev)
}

type Artifact struct {
	Time   int64  `json:"date"`
	Name   string `json:"tag"`
	Author string `json:"author"`
}

func (a Artifact) Equals(b Artifact) bool {
	return a.Name == b.Name && a.Time == b.Time
}

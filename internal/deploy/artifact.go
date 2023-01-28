package deploy

type Artifacts struct {
	Service string
	Prod    Artifact
	Dev     Artifact
}

func (a Artifacts) IsProdBehind() bool {
	if a.Prod.Name == "" {
		return false
	}
	if a.Prod.Name == a.Dev.Name {
		return false
	}

	if a.Prod.Time == 0 {
		return false
	}
	if a.Prod.Time >= a.Dev.Time {
		return false
	}

	return true
}

type Artifact struct {
	Time int64  `json:"date"`
	Name string `json:"tag"`
}

func (a Artifact) Equals(b Artifact) bool {
	return a.Name == b.Name && a.Time == b.Time
}

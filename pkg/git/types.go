package git

type GitForgeChange struct {
	FileName string
	Diff     string
}

type GitForgeChanges struct {
	Changes []GitForgeChange
}

type GitForce interface {
	GetCommitChanges(fromSha string, toSha string) (GitForgeChanges, error)
	GetMergeChanges(id string) (GitForgeChanges, error)
}

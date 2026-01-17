package git

type GitForgeChange struct {
	FileName string
	Diff     string
}

type GitForgeChanges struct {
	Changes []GitForgeChange
}

type GitForge interface {
	GetCommitChanges(fromSha string) (GitForgeChanges, error)
	GetMergeChanges(id string) (GitForgeChanges, error)
}

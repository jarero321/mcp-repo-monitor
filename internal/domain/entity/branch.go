package entity

type BranchComparison struct {
	Repository   string
	ProdBranch   string
	DevBranch    string
	AheadBy      int
	BehindBy     int
	TotalCommits int
	Status       DriftStatus
	Commits      []Commit
	Files        []ChangedFile
}

type DriftStatus string

const (
	DriftNone       DriftStatus = "synced"
	DriftProdAhead  DriftStatus = "prod_ahead"
	DriftDevAhead   DriftStatus = "dev_ahead"
	DriftDiverged   DriftStatus = "diverged"
)

type ChangedFile struct {
	Filename  string
	Status    string
	Additions int
	Deletions int
	Changes   int
	Patch     string
}

type SyncPRRequest struct {
	Repository string
	Title      string
	Body       string
	DryRun     bool
}

type SyncPRResult struct {
	Success     bool
	PRURL       string
	PRNumber    int
	Message     string
	FilesChanged int
	Commits     int
}

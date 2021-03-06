package introtator

import "sort"

// BackupFiles is used to satisfy a sort.Sort interface.
type backupFiles struct {
	Files []string
	value []int
}

// Len is part of sort.Interface.
func (b *backupFiles) Len() int {
	return len(b.Files)
}

// Swap is part of sort.Interface. We track two slices, so swap them both!
func (b *backupFiles) Swap(i, j int) {
	b.Files[i], b.Files[j] = b.Files[j], b.Files[i]
	b.value[i], b.value[j] = b.value[j], b.value[i]
}

// Less is part of the sort.Sort interface.
func (b *backupFiles) Less(i, j int) bool {
	return b.value[i] < b.value[j]
}

// Our backupFiles interface must satify a sort.Interface.
var _ sort.Interface = (*backupFiles)(nil)

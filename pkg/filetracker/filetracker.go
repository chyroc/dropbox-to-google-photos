package filetracker

type FileTracker struct {
	workDir string
	levelDB *LevelDBStore
}

func NewFileTracker(workDir string) (*FileTracker, error) {
	store, err := NewStore(workDir + "/tracker.db")
	if err != nil {
		return nil, err
	}
	return &FileTracker{
		workDir: workDir,
		levelDB: store,
	}, nil
}

func (r *FileTracker) Get(key string) string {
	content := r.levelDB.Get(key)
	return string(content)
}

func (r *FileTracker) Set(key, value string) {
	r.levelDB.Set(key, []byte(value))
}

func (r *FileTracker) Close() error {
	return r.levelDB.Close()
}

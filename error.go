package ora

// Error implements oracle error with error code
type Error struct {
	Code    int
	Message string
}

func (err Error) Error() string {
	return err.Message
}

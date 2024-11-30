package utils

// Print error to screen and write to error log file
func HandleErr(err error, logStmt string) bool {
	if err != nil {
		Logger(Error, Error, logStmt)
	}

	return err != nil
}

package postgres

func (s *CoreStorage) Close() {
	if err := s.db.Master.Close(); err != nil {
		s.logger.LogError("postgres — failed to close connection properly", err, "layer", "repository.postgres")
	} else {
		s.logger.LogInfo("postgres — database connection closed", "layer", "repository.postgres")
	}
}

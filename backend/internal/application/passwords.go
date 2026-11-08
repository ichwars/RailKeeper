package application

import "railkeeper/backend/internal/credentials"

func hashPassword(password string) (string, error) { return credentials.Hash(password) }

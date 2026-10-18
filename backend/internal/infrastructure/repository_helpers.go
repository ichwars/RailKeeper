package infrastructure

import "time"

type rowScanner interface{ Scan(...any) error }

func timestamp() string { return time.Now().UTC().Format(time.RFC3339) }

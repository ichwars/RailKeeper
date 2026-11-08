package application

import "railkeeper/backend/internal/inventory/money"

func parseMoneyCents(value string) (int64, bool) { return money.ParseCents(value) }

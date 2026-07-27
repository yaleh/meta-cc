package ftsindex

// truncateBody caps s at limit bytes, returning the (possibly truncated)
// text and whether truncation occurred. This is the sole mechanism that
// bounds indexed text size — it runs once, before either the metadata
// "body" column or the FTS index ever sees the text, so an oversized tool
// output is never fully indexed/searchable in full: only its first `limit`
// bytes are ever written to disk (see itemBody, which calls this for every
// Item).
func truncateBody(s string, limit int) (string, bool) {
	if limit <= 0 || len(s) <= limit {
		return s, false
	}
	return s[:limit], true
}

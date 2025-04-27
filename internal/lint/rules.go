//go:build ignore
// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

// ManualCorrelationID flags any log message with "correlation_id=" in it
func ManualCorrelationID(m dsl.Matcher) {
	m.Match(`$p($*_, $msg, $*_)`).
		Where(m["msg"].Type.Is("string") && m["msg"].Const && m["msg"].Text.Matches(".*correlation_id=.*")).
		Report("Do not manually format correlation_id in log messages. Use logger.WithContext(ctx) instead.")
}

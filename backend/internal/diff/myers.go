package diff

import "strings"

type OpKind string

const (
	OpEqual  OpKind = "equal"
	OpInsert OpKind = "insert"
	OpDelete OpKind = "delete"
)

type Segment struct {
	Kind OpKind `json:"kind"`
	Text string `json:"text"`
}

type Result struct {
	Line []Segment `json:"line"`
	Char []Segment `json:"char"`
}

func Compare(a, b string) Result {
	return Result{
		Line: lcsDiff(splitLines(a), splitLines(b), "\n"),
		Char: lcsDiff(splitRunes(limitRunes(a, 8000)), splitRunes(limitRunes(b, 8000)), ""),
	}
}

func limitRunes(s string, n int) string {
	rs := []rune(s)
	if len(rs) <= n {
		return s
	}
	return string(rs[:n])
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func splitRunes(s string) []string {
	rs := []rune(s)
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = string(r)
	}
	return out
}

func lcsDiff(a, b []string, join string) []Segment {
	n, m := len(a), len(b)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	var raw []Segment
	i, j := 0, 0
	for i < n && j < m {
		if a[i] == b[j] {
			raw = append(raw, Segment{OpEqual, a[i]})
			i++
			j++
			continue
		}
		if dp[i+1][j] >= dp[i][j+1] {
			raw = append(raw, Segment{OpDelete, a[i]})
			i++
		} else {
			raw = append(raw, Segment{OpInsert, b[j]})
			j++
		}
	}
	for i < n {
		raw = append(raw, Segment{OpDelete, a[i]})
		i++
	}
	for j < m {
		raw = append(raw, Segment{OpInsert, b[j]})
		j++
	}
	return coalesce(raw, join)
}

func coalesce(in []Segment, join string) []Segment {
	if len(in) == 0 {
		return nil
	}
	var out []Segment
	cur := in[0]
	for i := 1; i < len(in); i++ {
		if in[i].Kind == cur.Kind {
			if cur.Text == "" {
				cur.Text = in[i].Text
			} else {
				cur.Text = cur.Text + join + in[i].Text
			}
			continue
		}
		out = append(out, cur)
		cur = in[i]
	}
	out = append(out, cur)
	return out
}

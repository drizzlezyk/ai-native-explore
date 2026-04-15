//go:build debug

package main

func extractTestCases(cases []TestCaseWithSource) []TestCase {
	plainCases := make([]TestCase, 0, len(cases))
	for _, tc := range cases {
		plainCases = append(plainCases, tc.TestCase)
	}

	return plainCases
}

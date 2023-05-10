package main

import (
	"strconv"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {

	testCases := []struct {
		input     string
		expect    config
		expectErr string
	}{
		{
			input:  "VERSION_PREFIX=v\n",
			expect: config{versionPrefix: "v"},
		},
		{
			input:  "GIT_SIGN=1\n",
			expect: config{sign: true},
		},
		{
			input:  "VERSION_PREFIX=v\nGIT_SIGN=t\n",
			expect: config{versionPrefix: "v", sign: true},
		},
		{
			input:  "version_prefix=v\ngit_sign=t\n",
			expect: config{versionPrefix: "v", sign: true},
		},
		{
			input:  "# My config\nversion_prefix=v\ngit_sign=t\n",
			expect: config{versionPrefix: "v", sign: true},
		},
		{
			input:  "VERSION_PREFIX=v\nGIT_SIGN=false\n",
			expect: config{versionPrefix: "v", sign: false},
		},
		{
			input:  "VERSION_PREFIX=\"version/\"\nGIT_SIGN=f\n",
			expect: config{versionPrefix: "version/", sign: false},
		},
		{
			input:  "VERSION_PREFIX=\"foo\\\"bar\"\n",
			expect: config{versionPrefix: "foo\"bar"},
		},
		{
			input:     "foo\n",
			expectErr: "error parsing foo: invalid syntax",
		},
		{
			input:     "bad_key=v\n",
			expectErr: "error parsing bad_key=v: unrecognized variable",
		},
		{
			input:     "GIT_SIGN=wibble\n",
			expectErr: "error parsing GIT_SIGN=wibble: invalid boolean value",
		},
		{
			input:     "VERSION_PREFIX=\"foo\\z\"\n",
			expectErr: "error parsing VERSION_PREFIX=\"foo\\z\": invalid quoted string",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			res, err := parseConfig(strings.NewReader(tc.input))
			if tc.expectErr != "" {
				if err == nil {
					t.Error("expected error not returned")
					return
				}
				if err.Error() != tc.expectErr {
					t.Errorf("expected error %q, got %q", tc.expectErr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("got unexpected error: %v", err)
				}
				if res.versionPrefix != tc.expect.versionPrefix {
					t.Errorf("got versionPrefix %s, expected %s", res.versionPrefix, tc.expect.versionPrefix)
				}
				if res.sign != tc.expect.sign {
					t.Errorf("got sign %t, expected %t", res.sign, tc.expect.sign)
				}
			}
		})
	}

}

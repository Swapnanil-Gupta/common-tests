// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package option

import "testing"

func TestSupportsEnvVarPassthrough(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mods   []Modifier
		assert func(*testing.T, *Option)
	}{
		{
			name: "IsEnvVarPassthroughByDefault",
			mods: []Modifier{},
			assert: func(t *testing.T, uut *Option) {
				if !uut.SupportsEnvVarPassthrough() {
					t.Fatal("expected SupportsEnvVarPassthrough to be true")
				}
			},
		},
		{
			name: "IsNotEnvVarPassthrough",
			mods: []Modifier{
				WithNoEnvironmentVariablePassthrough(),
			},
			assert: func(t *testing.T, uut *Option) {
				if uut.SupportsEnvVarPassthrough() {
					t.Fatal("expected SupportsEnvVarPassthrough to be false")
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			uut, err := New([]string{"nerdctl"}, test.mods...)
			if err != nil {
				t.Fatal(err)
			}

			test.assert(t, uut)
		})
	}
}

func TestTranslateWindowsHostPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "BuildContextPath",
			in:   []string{"build", "-t", "foo", `C:\Users\foo\finch-test123`},
			want: []string{"build", "-t", "foo", "/mnt/c/Users/foo/finch-test123"},
		},
		{
			name: "DockerfileAndContext",
			in:   []string{"build", "-f", `D:\a\Dockerfile`, `D:\a`},
			want: []string{"build", "-f", "/mnt/d/a/Dockerfile", "/mnt/d/a"},
		},
		{
			name: "OutputDestKeyValue",
			in:   []string{"build", "--output", `type=tar,dest=C:\Users\foo\out.tar`, `C:\ctx`},
			want: []string{"build", "--output", "type=tar,dest=/mnt/c/Users/foo/out.tar", "/mnt/c/ctx"},
		},
		{
			name: "BindMountHostPathOnly",
			in:   []string{"run", "-v", `C:\host:/container`, "alpine:3.13"},
			want: []string{"run", "-v", "/mnt/c/host:/container", "alpine:3.13"},
		},
		{
			name: "ImageTagAndFlagsUntouched",
			in:   []string{"pull", "alpine:3.13", "--platform", "linux/amd64"},
			want: []string{"pull", "alpine:3.13", "--platform", "linux/amd64"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := translateWindowsHostPaths(test.in)
			if len(got) != len(test.want) {
				t.Fatalf("length mismatch: got %v, want %v", got, test.want)
			}
			for i := range got {
				if got[i] != test.want[i] {
					t.Errorf("arg %d: got %q, want %q", i, got[i], test.want[i])
				}
			}
		})
	}
}

func TestNerdctlVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mods   []Modifier
		assert func(*testing.T, *Option)
	}{
		{
			name: "IsNerdctlV2ByDefault",
			mods: []Modifier{},
			assert: func(t *testing.T, uut *Option) {
				if !uut.IsNerdctlV2() {
					t.Fatal("expected IsNerdctlV2 to be true")
				}
			},
		},
		{
			name: "IsNerdctlV1",
			mods: []Modifier{
				WithNerdctlVersion("1.7.7"),
			},
			assert: func(t *testing.T, uut *Option) {
				if !uut.IsNerdctlV1() {
					t.Fatal("expected IsNerdctlV1 to be true")
				}
			},
		},
		{
			name: "IsNerdctlV2",
			mods: []Modifier{
				WithNerdctlVersion("2.0.2"),
			},
			assert: func(t *testing.T, uut *Option) {
				if !uut.IsNerdctlV2() {
					t.Fatal("expected IsNerdctlV2 to be true")
				}
			},
		},
		{
			name: "IsPatchedNerdctlV2",
			mods: []Modifier{
				WithNerdctlVersion("2.0.2.m"),
			},
			assert: func(t *testing.T, uut *Option) {
				if !uut.IsNerdctlV2() {
					t.Fatal("expected IsNerdctlV2 to be true")
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			uut, err := New([]string{"nerdctl"}, test.mods...)
			if err != nil {
				t.Fatal(err)
			}

			test.assert(t, uut)
		})
	}
}

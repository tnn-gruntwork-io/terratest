package version_checker

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParamValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		param                CheckVersionParams
		containError         bool
		expectedErrorMessage string
	}{
		{
			name:                 "Empty Params",
			param:                CheckVersionParams{},
			containError:         true,
			expectedErrorMessage: "set WorkingDir in params",
		},
		{
			name: "Missing MinimumVersion",
			param: CheckVersionParams{
				VersionCheckerType: MinimumVersion,
				Binary:             Docker,
				VersionConstraint:  "",
				MinimumVersion:     "",
				WorkingDir:         ".",
			},
			containError:         true,
			expectedErrorMessage: "set MinimumVersion in params",
		},
		{
			name: "Missing VersionConstraint",
			param: CheckVersionParams{
				VersionCheckerType: VersionConstraint,
				Binary:             Docker,
				VersionConstraint:  "",
				MinimumVersion:     "",
				WorkingDir:         ".",
			},
			containError:         true,
			expectedErrorMessage: "set VersionConstraint in params",
		},
		{
			name: "Invalid Minimum Version Format",
			param: CheckVersionParams{
				VersionCheckerType: MinimumVersion,
				Binary:             Docker,
				MinimumVersion:     "abc",
				WorkingDir:         ".",
			},
			containError:         true,
			expectedErrorMessage: "invalid minimum version format found {abc}",
		},
		{
			name: "Invalid Version Constraint Format",
			param: CheckVersionParams{
				VersionCheckerType: VersionConstraint,
				Binary:             Docker,
				VersionConstraint:  "abc",
				WorkingDir:         ".",
			},
			containError:         true,
			expectedErrorMessage: "invalid version constraint format found {abc}",
		},
		{
			name: "Success",
			param: CheckVersionParams{
				VersionCheckerType: MinimumVersion,
				Binary:             Docker,
				MinimumVersion:     "1.2.3",
				WorkingDir:         ".",
			},
			containError:         false,
			expectedErrorMessage: "",
		},
	}

	for _, tc := range tests {
		err := validateParams(tc.param)
		if tc.containError {
			require.EqualError(t, err, tc.expectedErrorMessage, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}

func TestExtractVersionFromShellCommandOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		outputStr            string
		expectedVersionStr   string
		containError         bool
		expectedErrorMessage string
	}{
		{
			name:                 "Stand-alone version string",
			outputStr:            "version is 1.2.3",
			expectedVersionStr:   "1.2.3",
			containError:         false,
			expectedErrorMessage: "",
		},
		{
			name:                 "version string with v prefix",
			outputStr:            "version is v1.0.0",
			expectedVersionStr:   "1.0.0",
			containError:         false,
			expectedErrorMessage: "",
		},
		{
			name:                 "2 digit version string",
			outputStr:            "version is v1.0",
			expectedVersionStr:   "1.0",
			containError:         false,
			expectedErrorMessage: "",
		},
		{
			name:                 "invalid output string",
			outputStr:            "version is vabc",
			expectedVersionStr:   "",
			containError:         true,
			expectedErrorMessage: "failed to find version using regex matcher",
		},
		{
			name:                 "empty output string",
			outputStr:            "",
			expectedVersionStr:   "",
			containError:         true,
			expectedErrorMessage: "failed to find version using regex matcher",
		},
	}

	for _, tc := range tests {
		versionStr, err := extractVersionFromShellCommandOutput(tc.outputStr)
		if tc.containError {
			require.EqualError(t, err, tc.expectedErrorMessage, tc.name)
		} else {
			require.NoError(t, err, tc.name)
			require.Equal(t, tc.expectedVersionStr, versionStr, tc.name)
		}
	}
}

func TestCheckMimnimumVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		actualVersionStr     string
		minimumVersionStr    string
		containError         bool
		expectedErrorMessage string
	}{
		{
			name:                 "invalid actualVersionStr",
			actualVersionStr:     "",
			minimumVersionStr:    "1.2.3",
			containError:         true,
			expectedErrorMessage: "invalid version format found for actualVersionStr: ",
		},
		{
			name:                 "invalid minimumVersionStr",
			actualVersionStr:     "1.2.3",
			minimumVersionStr:    "",
			containError:         true,
			expectedErrorMessage: "invalid version format found for minimumVersionStr: ",
		},
		{
			name:                 "same version as minimum version",
			actualVersionStr:     "1.2.3",
			minimumVersionStr:    "1.2.3",
			containError:         false,
			expectedErrorMessage: "",
		},
		{
			name:                 "higher version as minimum version",
			actualVersionStr:     "1.2.4",
			minimumVersionStr:    "1.2.3",
			containError:         false,
			expectedErrorMessage: "",
		},
		{
			name:                 "lower version as minimum version",
			actualVersionStr:     "1.2.2",
			minimumVersionStr:    "1.2.3",
			containError:         true,
			expectedErrorMessage: "actual version {1.2.2} is less than the minimum version required {1.2.3}",
		},
	}

	for _, tc := range tests {
		err := checkMinimumVersion(tc.actualVersionStr, tc.minimumVersionStr)
		if tc.containError {
			require.EqualError(t, err, tc.expectedErrorMessage, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}

func TestCheckVersionConstraint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		actualVersionStr     string
		versionConstraint    string
		containError         bool
		expectedErrorMessage string
	}{
		{
			name:                 "invalid actualVersionStr",
			actualVersionStr:     "",
			versionConstraint:    "1.2.3",
			containError:         true,
			expectedErrorMessage: "invalid version format found for actualVersionStr: ",
		},
		{
			name:                 "invalid versionConstraint",
			actualVersionStr:     "1.2.3",
			versionConstraint:    "",
			containError:         true,
			expectedErrorMessage: "invalid version format found for versionConstraint: ",
		},
		{
			name:                 "pass version constraint",
			actualVersionStr:     "1.2.3",
			versionConstraint:    "1.2.3",
			containError:         false,
			expectedErrorMessage: "",
		},
		{
			name:                 "fail version constraint",
			actualVersionStr:     "1.2.3",
			versionConstraint:    "1.2.4",
			containError:         true,
			expectedErrorMessage: "actual version {1.2.3} failed the version constraint {1.2.4}",
		},
		{
			name:                 "special syntax version constraint",
			actualVersionStr:     "1.0.5",
			versionConstraint:    "~> 1.0.4",
			containError:         false,
			expectedErrorMessage: "",
		},
		{
			name:                 "version constraint w/ operators",
			actualVersionStr:     "1.2.7",
			versionConstraint:    ">= 1.2.0, < 2.0.0",
			containError:         false,
			expectedErrorMessage: ""},
	}

	for _, tc := range tests {
		err := checkVersionConstraint(tc.actualVersionStr, tc.versionConstraint)
		if tc.containError {
			require.EqualError(t, err, tc.expectedErrorMessage, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}

func TestCheckVersionSanityCheck(t *testing.T) {
	t.Parallel()

	// Note: with the current implementation of running shell command, it's not easy to
	// mock the output of running a shell command. So we assume a certain Binary is installed in the working
	// directory and it's greater than 0.
	err := CheckVersionE(t, CheckVersionParams{
		VersionCheckerType: MinimumVersion,
		Binary:             Terraform,
		MinimumVersion:     "0.0.1",
		WorkingDir:         ".",
	})
	require.NoError(t, err)
}

package git_test

import (
	"slices"
	"testing"

	"github.com/julianwyz/git-do/internal/git"
)

func TestExtractFlags_Shorthand(t *testing.T) {
	result, found, pos := git.ExtractFlag(
		[]string{
			"--foo",
			"--bar",
			"-mhello",
		},
		"-m",
	)

	if !found {
		t.Error("should have found value")
	}

	if result != "hello" {
		t.Error("incorrect result", result)
	}

	result, found, pos = git.ExtractFlag(
		[]string{
			"--foo",
			"--bar",
			"-m\"multi word\"",
		},
		"-m",
	)

	if !found {
		t.Error("should have found value")
	}

	if result != "multi word" {
		t.Error("incorrect result", result)
	}

	if !slices.Equal(pos, []int{2}) {
		t.Error("position incorrect", pos)
	}
}

func TestExtractFlags__NotFound(t *testing.T) {
	result, found, pos := git.ExtractFlag(
		[]string{
			"--foo",
			"--bar",
			"-mhello",
		},
		"--whatever",
	)

	if found {
		t.Error("should have not found value")
	}

	if result != "" {
		t.Error("incorrect result", result)
	}

	if !slices.Equal(pos, []int{}) {
		t.Error("position incorrect", pos)
	}
}

func TestExtractFlags__EqualsValue(t *testing.T) {
	result, found, pos := git.ExtractFlag(
		[]string{
			"--foo",
			"--bar",
			"-m=hello",
		},
		"-m",
	)

	if !found {
		t.Error("should have found value")
	}

	if result != "hello" {
		t.Error("incorrect result", result)
	}

	if !slices.Equal(pos, []int{2}) {
		t.Error("position incorrect", pos)
	}

	result, found, pos = git.ExtractFlag(
		[]string{
			"--foo",
			"--bar",
			"-m='hello world'",
		},
		"-m",
	)

	if !found {
		t.Error("should have found value")
	}

	if result != "hello world" {
		t.Error("incorrect result", result)
	}
}

func TestExtractFlags__SpaceValue(t *testing.T) {
	result, found, pos := git.ExtractFlag(
		[]string{
			"--foo",
			"--bar",
			"-m",
			"hello",
		},
		"-m",
	)

	if !found {
		t.Error("should have found value")
	}

	if result != "hello" {
		t.Error("incorrect result", result)
	}

	if !slices.Equal(pos, []int{3, 2}) {
		t.Error("position incorrect", pos)
	}

	result, found, pos = git.ExtractFlag(
		[]string{
			"--foo",
			"--bar",
			"-m",
			"'hello world'",
		},
		"-m",
	)

	if !found {
		t.Error("should have found value")
	}

	if result != "hello world" {
		t.Error("incorrect result", result)
	}
}

func TestExtractFlags__SpaceNoValue(t *testing.T) {
	result, found, pos := git.ExtractFlag(
		[]string{
			"--foo",
			"--bar",
			"-m",
			"hello",
		},
		"-m",
	)

	if !found {
		t.Error("should have found value")
	}

	if result != "hello" {
		t.Error("incorrect result", result)
	}

	if !slices.Equal(pos, []int{3, 2}) {
		t.Error("position incorrect", pos)
	}

	result, found, pos = git.ExtractFlag(
		[]string{
			"--foo",
			"--bar",
			"-m",
		},
		"-m",
	)

	if found {
		t.Error("should have not found value")
	}

	if result != "" {
		t.Error("incorrect result", result)
	}
}

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func home() string {
	file, err := ioutil.TempDir(os.TempDir(), "antibody")
	if err != nil {
		panic(err.Error())
	}
	defer os.RemoveAll(file)
	os.Setenv("ANTIBODY_HOME", file+"/")
	return file + "/"
}

func expectError(t *testing.T) {
	if err := recover(); err != nil {
		t.Log("Recovered from expected error")
	} else {
		t.Error("Expected an error here!")
	}
}

func assertBundledPlugins(t *testing.T, total int, home string) {
	plugins, _ := ioutil.ReadDir(home)
	if len(plugins) != total {
		t.Error("Expected to bundle", total, "plugins, but was", len(plugins))
	}
}

func TestProcessesArgsDoBundle(t *testing.T) {
	home := home()
	ProcessArgs([]string{"bundle", "caarlos0/zsh-pg"}, home)
	assertBundledPlugins(t, 1, home)
}

func TestUpdateWithNoPlugins(t *testing.T) {
	home := home()
	ProcessArgs([]string{"update"}, home)
	assertBundledPlugins(t, 0, home)
}

func TestUpdateWithPlugins(t *testing.T) {
	home := home()
	DoBundle("caarlos0/zsh-pg", home)
	ProcessArgs([]string{"update"}, home)
	assertBundledPlugins(t, 1, home)
}

func TestBundlesSinglePlugin(t *testing.T) {
	home := home()
	DoBundle("caarlos0/zsh-pg", home)
	assertBundledPlugins(t, 1, home)
}

func TestLoadsDefaultHome(t *testing.T) {
	os.Unsetenv("ANTIBODY_HOME")
	if !strings.HasSuffix(Home(), "/.antibody/") {
		t.Error("Expected default ANTIBODY_HOME")
	}
}

func TestLoadsCustomHome(t *testing.T) {
	home := home()
	if home != Home() {
		t.Error("Expected custom ANTIBODY_HOME")
	}
}

func TestAddsTrailingSlashToHome(t *testing.T) {
	home := "/tmp/whatever"
	os.Setenv("ANTIBODY_HOME", home)
	if Home() != home+"/" {
		t.Error("Expected a trailing slash in", Home())
	}
}

func TestFailsToBundleInvalidRepos(t *testing.T) {
	home := home()
	defer expectError(t)
	DoBundle("csadsadp", home)
	assertBundledPlugins(t, 0, home)
}

func TestFailsToProcessInvalidArgs(t *testing.T) {
	home := home()
	defer expectError(t)
	ProcessArgs([]string{"nope", "caarlos0/zsh-pg"}, home)
	assertBundledPlugins(t, 0, home)
}

func TestReadsStdinIsFalse(t *testing.T) {
	if ReadStdin() {
		t.Error("Not reading STDIN")
	}
}

func TestReadsStdinIsTrue(t *testing.T) {
	os.Stdin.Write([]byte("Some STDIN"))
	if ReadStdin() {
		t.Error("Not reading STDIN")
	}
}

func TestProcessStdin(t *testing.T) {
	home := home()
	bundles := bytes.NewBufferString("caarlos0/zsh-pg\ncaarlos0/zsh-add-upstream")
	ProcessStdin(bundles, home)
	assertBundledPlugins(t, 2, home)
}

func TestProcessStdinWithEmptyLines(t *testing.T) {
	home := home()
	bundles := bytes.NewBufferString("\ncaarlos0/zsh-pg\ncaarlos0/zsh-add-upstream\n")
	ProcessStdin(bundles, home)
	assertBundledPlugins(t, 2, home)
}

func TestUpdatesListOfRepos(t *testing.T) {
	home := home()
	bundle1 := "caarlos0/zsh-pg"
	bundle2 := "caarlos0/zsh-add-upstream"
	NewGithubBundle(bundle1, home).Download()
	NewGithubBundle(bundle2, home).Download()
	bundles, err := Update(home)
	if err != nil {
		t.Error("No errors expected")
	}
	if len(bundles) != 2 {
		t.Error(len(bundles), "updated bundles, expected 2")
	}
}

func TestUpdatesBrokenRepo(t *testing.T) {
	home := home()
	bundle := "caarlos0/zsh-mkc"
	folder, _ := NewGithubBundle(bundle, home).Download()
	os.RemoveAll(folder + "/.git")
	bundles, err := Update(home)
	if err == nil {
		t.Error("An error was expected")
	}
	if len(bundles) != 0 {
		t.Error(len(bundles), "updated bundles, expected 0")
	}
}

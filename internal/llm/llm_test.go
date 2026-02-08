package llm_test

import (
	"bytes"
	"encoding/json"
	"io"
	"iter"
	"net/http"
	"testing"

	"github.com/julianwyz/git-do/internal/git"
	"github.com/julianwyz/git-do/internal/llm"
	"golang.org/x/text/language"
)

type (
	ctxLoader struct{}
	roundtrip struct{}
)

// Some of these tests are kinda shallow right now
// as I don't have a great OpenAI API mock spun up.
// But this will cover some baselines

func TestNew(t *testing.T) {
	client, err := llm.New(
		llm.WithOutputLanguage(language.AmericanEnglish),
		llm.WithCommitFormat(git.CommitFormatGithub),
		llm.WithContextLoader(&ctxLoader{}),
		llm.WithAPIBase("http://api.example.com"),
		llm.WithAPIKey("foobar"),
		llm.WithModel("model"),
		llm.WithReasoningLevel(llm.ReasoningLevelMedium),
		llm.WithHTTPClient(http.DefaultClient),
	)
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Fatal("expected client")
	}

	if client.GetModel() != "model" {
		t.Fatal("incorrect model")
	}

	if client.GetAPIDomain() != "example.com" {
		t.Fatal("incorrect domain")
	}
}

func TestExplainCommits(t *testing.T) {
	client, err := llm.New(
		llm.WithOutputLanguage(language.AmericanEnglish),
		llm.WithCommitFormat(git.CommitFormatGithub),
		llm.WithContextLoader(&ctxLoader{}),
		llm.WithReasoningLevel(llm.ReasoningLevelMedium),
		llm.WithHTTPClient(makeClient()),
	)
	if err != nil {
		t.Fatal(err)
	}

	dst := &bytes.Buffer{}
	seq := commitList("hello world")
	if err := client.ExplainCommits(
		t.Context(),
		seq,
		dst,
	); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateCommit(t *testing.T) {
	client, err := llm.New(
		llm.WithOutputLanguage(language.AmericanEnglish),
		llm.WithCommitFormat(git.CommitFormatGithub),
		llm.WithContextLoader(&ctxLoader{}),
		llm.WithReasoningLevel(llm.ReasoningLevelMedium),
		llm.WithHTTPClient(makeClient()),
	)
	if err != nil {
		t.Fatal(err)
	}

	seq := commitList("hello world")
	_, err = client.GenerateCommit(
		t.Context(),
		seq,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExplainStatus(t *testing.T) {
	client, err := llm.New(
		llm.WithOutputLanguage(language.AmericanEnglish),
		llm.WithCommitFormat(git.CommitFormatGithub),
		llm.WithContextLoader(&ctxLoader{}),
		llm.WithReasoningLevel(llm.ReasoningLevelMedium),
		llm.WithHTTPClient(makeClient()),
	)
	if err != nil {
		t.Fatal(err)
	}

	status := `On branch main
Your branch is ahead of 'origin/main' by 3 commits.
  (use "git push" to publish your local commits)

Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
  (use "git restore <file>..." to discard changes in working directory)
	modified:   internal/llm/llm.go
	modified:   internal/llm/options.go

Untracked files:
  (use "git add <file>..." to include in what will be committed)
	internal/llm/llm_test.go

no changes added to commit (use "git add" and/or "git commit -a")`

	dst := &bytes.Buffer{}
	seq := commitList("hello world")
	err = client.ExplainStatus(
		t.Context(),
		status,
		seq,
		dst,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func (*ctxLoader) LoadContextFile() (io.ReadCloser, error) {
	buf := bytes.Buffer{}

	return io.NopCloser(&buf), nil
}

func commitList(items ...string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for _, s := range items {
			if !yield(s, nil) {
				return
			}
		}
	}
}

func makeClient() *http.Client {
	return &http.Client{
		Transport: &roundtrip{},
	}
}

func (recv *roundtrip) RoundTrip(req *http.Request) (res *http.Response, err error) {
	hdr := make(http.Header)
	hdr.Set("content-type", "application/json")

	res = &http.Response{
		Header: hdr,
		Body:   recv.makeBody(map[string]any{}),
	}

	return
}

func (recv *roundtrip) makeBody(obj any) io.ReadCloser {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil
	}

	return io.NopCloser(
		bytes.NewBuffer(data),
	)
}

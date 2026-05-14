package llm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http"
	"strings"
	"text/template"
	"time"

	_ "embed"

	tld "github.com/jpillora/go-tld"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
)

type (
	LLM struct {
		backend provider
		config  *llmConfig
		apiUrl  *tld.URL
	}

	ReasoningLevel string

	commitInstructionsTemplateData struct {
		Language string
		Format   string
	}

	explanationInstructionsTemplateData struct {
		Language string
	}

	statusInstructionsTemplateData struct {
		Color    bool
		Language string
	}

	contextLoader interface {
		LoadContextFile() (io.ReadCloser, error)
	}
)

const (
	defaultModel        = "gpt-5-mini"
	defaultCommitFormat = "github"
)

const (
	ReasoningLevelNone    = ReasoningLevel("none")
	ReasoningLevelMinimal = ReasoningLevel("minimal")
	ReasoningLevelLow     = ReasoningLevel("low")
	ReasoningLevelMedium  = ReasoningLevel("medium")
	ReasoningLevelHigh    = ReasoningLevel("high")
	ReasoningLevelXHigh   = ReasoningLevel("xhigh")
)

var (
	ErrNoPatches = errors.New("no changes to commit")

	defaultLang = language.AmericanEnglish
	//go:embed prompts/gen_commit_instruct.tmpl.md
	genCommitInstSrc      string
	genCommitInstructions = func() *template.Template {
		t, err := template.New("gen_commit_instruct.tmpl.md").Parse(genCommitInstSrc)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse commit instruction template")
		}

		return t
	}()
	//go:embed prompts/explain_instruct.tmpl.md
	explainInstSrc      string
	explainInstructions = func() *template.Template {
		t, err := template.New("explain_instruct.tmpl.md").Parse(explainInstSrc)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse explanation instruction template")
		}

		return t
	}()
	//go:embed prompts/status_instruct.tmpl.md
	statusInstSrc      string
	statusInstructions = func() *template.Template {
		t, err := template.New("status_instruct.tmpl.md").Parse(statusInstSrc)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse status explanation instruction template")
		}

		return t
	}()
)

func New(
	opts ...LLMOpt,
) (*LLM, error) {
	config := &llmConfig{
		model: defaultModel,
		http:  http.DefaultClient,
	}
	for _, o := range opts {
		if err := o(config); err != nil {
			return nil, err
		}
	}

	parsedAPIUrl, err := tld.Parse(config.apiBase)
	if err != nil {
		return nil, err
	}

	var backend provider
	if strings.Contains(config.apiBase, "api.anthropic.com") {
		backend, err = newAnthropicProvider(config)

		log.Debug().
			Msg("using anthropic backend provider")
	} else {
		backend, err = newOpenAIProvider(config)

		log.Debug().
			Str("base", config.apiBase).
			Msg("using openai-compatible backend provider")
	}
	if err != nil {
		return nil, err
	}

	log.Debug().
		Str("base", config.apiBase).
		Msg("configured llm client")

	return &LLM{
		apiUrl:  parsedAPIUrl,
		config:  config,
		backend: backend,
	}, nil
}

func (recv *LLM) ExplainCommits(
	ctx context.Context,
	commits iter.Seq2[string, error],
	dst io.Writer,
) error {
	startTime := time.Now()

	instructions, err := execInstructionTmpl(
		explainInstructions,
		&explanationInstructionsTemplateData{Language: recv.langString()},
	)
	if err != nil {
		return err
	}

	var msgs []llmMessage
	msgs = append(msgs, llmMessage{text: fmt.Sprintf("COMMAND\nThis is being invoked by the `%s` command.", "commit")})
	if ctxText := recv.retrieveContextText(); ctxText != "" {
		msgs = append(msgs, llmMessage{text: ctxText})
	}
	for patch, patchErr := range commits {
		if patchErr != nil {
			return patchErr
		}
		msgs = append(msgs, llmMessage{text: patch})
	}
	msgs = append(msgs, llmMessage{text: "GENERATE"})

	tokensIn, tokensOut, err := recv.backend.streamOutput(ctx, instructions, msgs, dst)
	if err != nil {
		return err
	}

	log.Debug().
		Int64("input_tokens", tokensIn).
		Int64("output_tokens", tokensOut).
		Stringer("latency", time.Since(startTime)).
		Msg("llm response")

	return nil
}

func (recv *LLM) GenerateCommit(
	ctx context.Context,
	commits iter.Seq2[string, error],
	opts ...CommitOpt,
) (string, error) {
	config := &commitConfig{}
	for _, o := range opts {
		if err := o(config); err != nil {
			return "", err
		}
	}

	startTime := time.Now()

	instructionData := &commitInstructionsTemplateData{
		Language: recv.langString(),
		Format:   defaultCommitFormat,
	}
	if len(recv.config.commitFormat) > 0 {
		instructionData.Format = string(recv.config.commitFormat)
	}

	instructions, err := execInstructionTmpl(genCommitInstructions, instructionData)
	if err != nil {
		return "", err
	}

	var (
		msgs       []llmMessage
		patchCount int64
	)

	msgs = append(msgs, llmMessage{text: fmt.Sprintf("COMMAND\nThis is being invoked by the `%s` command.", "explain")})
	if ctxText := recv.retrieveContextText(); ctxText != "" {
		msgs = append(msgs, llmMessage{text: ctxText})
	}

	for patch, patchErr := range commits {
		patchCount++
		if patchErr != nil {
			return "", patchErr
		}
		msgs = append(msgs, llmMessage{text: patch})
	}

	if patchCount == 0 {
		return "", ErrNoPatches
	}

	if len(config.resolutions) > 0 {
		msgs = append(msgs, llmMessage{text: fmt.Sprintf("RESOLUTIONS\n%s",
			strings.Join(config.resolutions, "\n"))})
	}
	if len(config.instructions) > 0 {
		msgs = append(msgs, llmMessage{text: fmt.Sprintf("INSTRUCTIONS\n%s", config.instructions)})
	}
	msgs = append(msgs, llmMessage{text: "GENERATE"})

	output, err := recv.backend.generateCommit(ctx, instructions, msgs)
	if err != nil {
		return "", err
	}

	log.Debug().
		Int64("patch_count", patchCount).
		Stringer("latency", time.Since(startTime)).
		Msg("llm response")

	return output, nil
}

func (recv *LLM) GetModel() string {
	return recv.config.model
}

func (recv *LLM) GetAPIDomain() string {
	return fmt.Sprintf("%s.%s",
		recv.apiUrl.Domain,
		recv.apiUrl.TLD,
	)
}

func (recv *LLM) ExplainStatus(
	ctx context.Context,
	statusOutput string,
	statusChanges iter.Seq2[string, error],
	dst io.Writer,
) error {
	startTime := time.Now()

	instructions, err := execInstructionTmpl(
		statusInstructions,
		&statusInstructionsTemplateData{
			Color:    true,
			Language: recv.langString(),
		},
	)
	if err != nil {
		return err
	}

	var msgs []llmMessage
	msgs = append(msgs, llmMessage{text: fmt.Sprintf("COMMAND\nThis is being invoked by the `%s` command.", "status")})
	if ctxText := recv.retrieveContextText(); ctxText != "" {
		msgs = append(msgs, llmMessage{text: ctxText})
	}
	msgs = append(msgs, llmMessage{text: fmt.Sprintf("STATUS\n%s", statusOutput)})

	for patch, patchErr := range statusChanges {
		if patchErr != nil {
			return patchErr
		}
		msgs = append(msgs, llmMessage{text: patch})
	}
	msgs = append(msgs, llmMessage{text: "GENERATE"})

	tokensIn, tokensOut, err := recv.backend.streamOutput(ctx, instructions, msgs, dst)
	if err != nil {
		return err
	}

	if _, err := dst.Write([]byte("\n")); err != nil {
		return err
	}

	log.Debug().
		Int64("input_tokens", tokensIn).
		Int64("output_tokens", tokensOut).
		Stringer("latency", time.Since(startTime)).
		Msg("llm response")

	return nil
}

func (recv *LLM) langString() string {
	if recv.config.outputLang != nil {
		return recv.config.outputLang.String()
	}
	return defaultLang.String()
}

func (recv *LLM) retrieveContextText() string {
	if recv.config.contextLoader == nil {
		return ""
	}
	rc, err := recv.config.contextLoader.LoadContextFile()
	if err != nil {
		return ""
	}
	defer rc.Close()
	var sb strings.Builder
	sb.WriteString("CONTEXT\n")
	io.Copy(&sb, rc) //nolint:errcheck
	return sb.String()
}

func execInstructionTmpl(t *template.Template, data any) (string, error) {
	dst := &bytes.Buffer{}
	if err := t.Execute(dst, data); err != nil {
		return "", err
	}
	return dst.String(), nil
}

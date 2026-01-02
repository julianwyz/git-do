package llm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"strings"
	"text/template"
	"time"

	_ "embed"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
)

type (
	LLM struct {
		client *openai.Client
		config *llmConfig
	}

	ReasoningLevel string

	instructionsTemplateData struct {
		Language string
		Format   string
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
)

func New(
	opts ...LLMOpt,
) (*LLM, error) {
	config := &llmConfig{
		model: defaultModel,
	}
	for _, o := range opts {
		if err := o(config); err != nil {
			return nil, err
		}
	}

	client := openai.NewClient(
		option.WithBaseURL(config.apiBase),
		option.WithAPIKey(config.apiKey),
	)
	log.Debug().
		Str("base", config.apiBase).
		Msg("configured llm client")

	return &LLM{
		config: config,
		client: &client,
	}, nil
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

	instructionData := &instructionsTemplateData{
		Language: defaultLang.String(),
		Format:   defaultCommitFormat,
	}

	if recv.config.outputLang != nil {
		instructionData.Language = recv.config.outputLang.String()
	}
	if len(recv.config.commitFormat) > 0 {
		instructionData.Format = string(recv.config.commitFormat)
	}

	instructions, err := execInstructionTmpl(
		genCommitInstructions,
		instructionData,
	)

	var tokensIn, tokensOut, patchCount int64
	var commitInput responses.ResponseInputParam

	if recv.config.contextLoader != nil {
		rc, err := recv.config.contextLoader.LoadContextFile()
		if err == nil {
			defer rc.Close()

			msg := bytes.NewBufferString("CONTEXT\n")
			if _, err := io.Copy(msg, rc); err == nil {
				commitInput = append(commitInput, responses.ResponseInputItemUnionParam{
					OfMessage: &responses.EasyInputMessageParam{
						Role: responses.EasyInputMessageRoleUser,
						Content: responses.EasyInputMessageContentUnionParam{
							OfString: param.NewOpt(msg.String()),
						},
					},
				})
			}
		}
	}

	for patch, err := range commits {
		patchCount++
		if err != nil {
			return "", err
		}

		commitInput = append(commitInput, responses.ResponseInputItemUnionParam{
			OfMessage: &responses.EasyInputMessageParam{
				Role: responses.EasyInputMessageRoleUser,
				Content: responses.EasyInputMessageContentUnionParam{
					OfString: param.NewOpt(patch),
				},
			},
		})
	}

	if patchCount == 0 {
		return "", ErrNoPatches
	}

	if len(config.resolutions) > 0 {
		msg := fmt.Sprintf("RESOLUTIONS\n%s",
			strings.Join(config.resolutions, "\n"))

		commitInput = append(commitInput, responses.ResponseInputItemUnionParam{
			OfMessage: &responses.EasyInputMessageParam{
				Role: responses.EasyInputMessageRoleUser,
				Content: responses.EasyInputMessageContentUnionParam{
					OfString: param.NewOpt(msg),
				},
			},
		})
	}

	commitInput = append(commitInput, responses.ResponseInputItemUnionParam{
		OfMessage: &responses.EasyInputMessageParam{
			Role: responses.EasyInputMessageRoleUser,
			Content: responses.EasyInputMessageContentUnionParam{
				OfString: param.NewOpt("GENERATE"),
			},
		},
	})

	respParams := responses.ResponseNewParams{
		Model:        recv.config.model,
		Instructions: param.NewOpt(instructions),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: commitInput,
		},
	}

	if len(recv.config.reasoning) > 0 {
		respParams.Reasoning = shared.ReasoningParam{
			Effort: shared.ReasoningEffort(recv.config.reasoning),
		}
	}

	resp, err := recv.client.Responses.New(
		ctx, respParams,
	)
	if err != nil {
		return "", err
	}

	tokensIn += resp.Usage.InputTokens
	tokensOut += resp.Usage.OutputTokens
	output := resp.OutputText()

	log.Debug().
		Int64("input_tokens", tokensIn).
		Int64("output_tokens", tokensOut).
		Stringer("latency", time.Since(startTime)).
		Msg("llm response")
	log.Debug().Msgf("llm output:\n%s", output)

	return output, nil
}

func execInstructionTmpl(t *template.Template, data any) (string, error) {
	dst := &bytes.Buffer{}
	if err := t.Execute(dst, data); err != nil {
		return "", err
	}

	return dst.String(), nil
}

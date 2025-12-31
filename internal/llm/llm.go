package llm

import (
	"bytes"
	"context"
	"io"
	"iter"
	"text/template"
	"time"

	_ "embed"

	"github.com/julianwyz/git-do/internal/config"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
	"github.com/rs/zerolog/log"
)

type (
	LLM struct {
		config *config.Config
		client *openai.Client
		model  string
	}

	instructionsTemplateData struct {
		Language string
		Format   string
	}
)

const (
	defaultModel        = "gpt-5-mini"
	defaultLang         = "en-US"
	defaultCommitFormat = "github"
)

var (
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
		config: config.config,
		client: &client,
		model:  config.model,
	}, nil
}

func (recv *LLM) GenerateCommit(ctx context.Context, commits iter.Seq2[string, error]) (string, error) {
	startTime := time.Now()

	instructionData := &instructionsTemplateData{
		Language: defaultLang,
		Format:   defaultCommitFormat,
	}
	if recv.config != nil {
		if len(recv.config.Language) > 0 {
			instructionData.Language = recv.config.Language
		}
		if recv.config.Commit != nil {
			if len(recv.config.Commit.Format) > 0 {
				instructionData.Format = string(recv.config.Commit.Format)
			}
		}
	}

	instructions, err := execInstructionTmpl(
		genCommitInstructions,
		instructionData,
	)

	var tokensIn, tokensOut int64
	var commitInput responses.ResponseInputParam

	if recv.config != nil {
		rc, err := recv.config.LoadContextFile()
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

	commitInput = append(commitInput, responses.ResponseInputItemUnionParam{
		OfMessage: &responses.EasyInputMessageParam{
			Role: responses.EasyInputMessageRoleUser,
			Content: responses.EasyInputMessageContentUnionParam{
				OfString: param.NewOpt("GENERATE"),
			},
		},
	})

	resp, err := recv.client.Responses.New(
		ctx, responses.ResponseNewParams{
			Model:        recv.model,
			Instructions: param.NewOpt(instructions),
			Input: responses.ResponseNewParamsInputUnion{
				OfInputItemList: commitInput,
			},
		},
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

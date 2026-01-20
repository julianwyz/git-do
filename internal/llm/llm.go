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

	tld "github.com/jpillora/go-tld"
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
		apiUrl *tld.URL
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

	client := openai.NewClient(
		option.WithBaseURL(config.apiBase),
		option.WithAPIKey(config.apiKey),
	)
	log.Debug().
		Str("base", config.apiBase).
		Msg("configured llm client")

	return &LLM{
		apiUrl: parsedAPIUrl,
		config: config,
		client: &client,
	}, nil
}

func (recv *LLM) ExplainCommits(
	ctx context.Context,
	commits iter.Seq2[string, error],
	dst io.Writer,
) error {
	startTime := time.Now()

	instructionData := &explanationInstructionsTemplateData{
		Language: defaultLang.String(),
	}

	if recv.config.outputLang != nil {
		instructionData.Language = recv.config.outputLang.String()
	}

	instructions, err := execInstructionTmpl(
		explainInstructions,
		instructionData,
	)
	if err != nil {
		return err
	}

	var tokensIn, tokensOut int64
	var explainInput responses.ResponseInputParam

	explainInput = append(explainInput, gitDoContextMsg("commit"))

	if recv.config.contextLoader != nil {
		rc, err := recv.config.contextLoader.LoadContextFile()
		if err == nil {
			defer rc.Close()

			msg := bytes.NewBufferString("CONTEXT\n")
			if _, err := io.Copy(msg, rc); err == nil {
				explainInput = append(explainInput, responses.ResponseInputItemUnionParam{
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
			return err
		}

		explainInput = append(explainInput, responses.ResponseInputItemUnionParam{
			OfMessage: &responses.EasyInputMessageParam{
				Role: responses.EasyInputMessageRoleUser,
				Content: responses.EasyInputMessageContentUnionParam{
					OfString: param.NewOpt(patch),
				},
			},
		})
	}

	explainInput = append(explainInput, responses.ResponseInputItemUnionParam{
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
			OfInputItemList: explainInput,
		},
	}

	if len(recv.config.reasoning) > 0 {
		respParams.Reasoning = shared.ReasoningParam{
			Effort: shared.ReasoningEffort(recv.config.reasoning),
		}
	}

	stream := recv.client.Responses.NewStreaming(
		ctx, respParams,
	)
	for stream.Next() {
		cur := stream.Current()
		if _, err := dst.Write([]byte(cur.Delta)); err != nil {
			return err
		}

		tokensIn += cur.Response.Usage.InputTokens
		tokensOut += cur.Response.Usage.OutputTokens
	}
	if err := stream.Err(); err != nil {
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
	if err != nil {
		return "", err
	}

	var tokensIn, tokensOut, patchCount int64
	var commitInput responses.ResponseInputParam

	commitInput = append(commitInput, gitDoContextMsg("explain"))

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

	if len(config.instructions) > 0 {
		msg := fmt.Sprintf("INSTRUCTIONS\n%s", config.instructions)

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

	instructionData := &statusInstructionsTemplateData{
		Color:    true,
		Language: defaultLang.String(),
	}

	if recv.config.outputLang != nil {
		instructionData.Language = recv.config.outputLang.String()
	}

	instructions, err := execInstructionTmpl(
		statusInstructions,
		instructionData,
	)
	if err != nil {
		return err
	}

	if recv.config.outputLang != nil {
		instructionData.Language = recv.config.outputLang.String()
	}

	var tokensIn, tokensOut int64
	var input responses.ResponseInputParam

	input = append(input, gitDoContextMsg("status"))

	if recv.config.contextLoader != nil {
		rc, err := recv.config.contextLoader.LoadContextFile()
		if err == nil {
			defer rc.Close()

			msg := bytes.NewBufferString("CONTEXT\n")
			if _, err := io.Copy(msg, rc); err == nil {
				input = append(input, responses.ResponseInputItemUnionParam{
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

	input = append(input, responses.ResponseInputItemUnionParam{
		OfMessage: &responses.EasyInputMessageParam{
			Role: responses.EasyInputMessageRoleUser,
			Content: responses.EasyInputMessageContentUnionParam{
				OfString: param.NewOpt(
					fmt.Sprintf("STATUS\n%s", statusOutput),
				),
			},
		},
	})

	for patch, err := range statusChanges {
		if err != nil {
			return err
		}

		input = append(input, responses.ResponseInputItemUnionParam{
			OfMessage: &responses.EasyInputMessageParam{
				Role: responses.EasyInputMessageRoleUser,
				Content: responses.EasyInputMessageContentUnionParam{
					OfString: param.NewOpt(patch),
				},
			},
		})
	}

	input = append(input, responses.ResponseInputItemUnionParam{
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
			OfInputItemList: input,
		},
	}

	if len(recv.config.reasoning) > 0 {
		respParams.Reasoning = shared.ReasoningParam{
			Effort: shared.ReasoningEffort(recv.config.reasoning),
		}
	}

	stream := recv.client.Responses.NewStreaming(
		ctx, respParams,
	)
	for stream.Next() {
		cur := stream.Current()
		if _, err := dst.Write([]byte(cur.Delta)); err != nil {
			return err
		}

		tokensIn += cur.Response.Usage.InputTokens
		tokensOut += cur.Response.Usage.OutputTokens
	}
	if err := stream.Err(); err != nil {
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

func execInstructionTmpl(t *template.Template, data any) (string, error) {
	dst := &bytes.Buffer{}
	if err := t.Execute(dst, data); err != nil {
		return "", err
	}

	return dst.String(), nil
}

func gitDoContextMsg(subcommand string) responses.ResponseInputItemUnionParam {
	return responses.ResponseInputItemUnionParam{
		OfMessage: &responses.EasyInputMessageParam{
			Role: responses.EasyInputMessageRoleUser,
			Content: responses.EasyInputMessageContentUnionParam{
				OfString: param.NewOpt(fmt.Sprintf(
					"COMMAND\nThis is being invoked by the `%s` command.", subcommand,
				),
				),
			},
		},
	}
}

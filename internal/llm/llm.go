package llm

import (
	"bytes"
	"context"
	"fmt"
	"iter"
	"log"
	"text/template"
	"time"

	_ "embed"

	"github.com/julianwyz/git-buddy/internal/config"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
)

type (
	LLM struct {
		config *config.LLM
		client *openai.Client
		model  string
	}

	instructionsTemplateData struct {
		Language string
		Format   string
	}
)

const (
	defaultModel = "gpt-5-mini"
)

var (
	//go:embed prompts/gen_commit_instruct.tmpl.md
	genCommitInstSrc      string
	genCommitInstructions = func() *template.Template {
		t, err := template.New("gen_commit_instruct.tmpl.md").Parse(genCommitInstSrc)
		if err != nil {
			log.Fatal(err.Error())
		}

		return t
	}()
)

func New(c *config.LLM) (*LLM, error) {
	model := defaultModel
	oaiOpts := []option.RequestOption{}
	if c != nil {
		if len(c.APIBase) > 0 {
			oaiOpts = append(oaiOpts, option.WithBaseURL(c.APIBase))
		}
		if len(c.APIKey) > 0 {
			oaiOpts = append(oaiOpts, option.WithAPIKey(c.APIKey))
		}
		if len(c.Model) > 0 {
			model = c.Model
		}
	}

	client := openai.NewClient(oaiOpts...)

	return &LLM{
		config: c,
		client: &client,
		model:  model,
	}, nil
}

func (recv *LLM) GenerateCommit(ctx context.Context, commits iter.Seq2[string, error]) (string, error) {
	startTime := time.Now()
	instructions, err := execInstructionTmpl(
		genCommitInstructions,
		instructionsTemplateData{
			Language: "en-US",
			Format:   "github",
		},
	)

	var tokensIn, tokensOut int64
	var commitInput responses.ResponseInputParam

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

	fmt.Printf("Response:\n%s\n\n↑ %d / ↓ %d | %s\n",
		output,
		tokensIn,
		tokensOut,
		time.Since(startTime),
	)

	return output, nil
}

func execInstructionTmpl(t *template.Template, data any) (string, error) {
	dst := &bytes.Buffer{}
	if err := t.Execute(dst, data); err != nil {
		return "", err
	}

	return dst.String(), nil
}

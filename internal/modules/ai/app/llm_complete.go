package app

import (
	"context"
	"fmt"

	"github.com/kore/kore/internal/modules/ai/app/promptguard"
	"github.com/kore/kore/internal/modules/ai/ports"
)

func (s *Service) llmComplete(
	ctx context.Context,
	capability string,
	taskInstruction string,
	untrusted map[string]string,
) (ports.CompletionResponse, error) {
	userPrompt := promptguard.BuildSandboxedUserPrompt(taskInstruction, untrusted)
	return s.llm.Complete(ctx, ports.CompletionRequest{
		SystemPrompt: fmt.Sprintf("Capability Kore : %s.", capability),
		UserPrompt:   userPrompt,
	})
}

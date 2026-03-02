package onecompiler

import internalonecompiler "github.com/Oleja123/code-vizualization/semantic-analyzer-service/internal/infrastructure/onecompiler"

type Client = internalonecompiler.Client
type FileContent = internalonecompiler.FileContent
type CompileRequest = internalonecompiler.CompileRequest
type CompileResponse = internalonecompiler.CompileResponse

func NewClient(apiURL string, apiKey string, timeoutSeconds int) *Client {
	return internalonecompiler.NewClient(apiURL, apiKey, timeoutSeconds)
}

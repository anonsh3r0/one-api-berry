package azureSpeech

import (
	"bytes"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/types"
)

var outputFormatMap = map[string]string{
	"mp3":  "audio-16khz-128kbitrate-mono-mp3",
	"opus": "audio-16khz-128kbitrate-mono-opus",
	"aac":  "audio-24khz-160kbitrate-mono-mp3",
	"flac": "audio-48khz-192kbitrate-mono-mp3",
}

func CreateSSML(text string, name string, role string) string {
	ssmlTemplate := `<speak version='1.0' xml:lang='en-US'>
        <voice xml:lang='en-US' %s name='%s'>
            %s
        </voice>
    </speak>`

	roleAttribute := ""
	if role != "" {
		roleAttribute = fmt.Sprintf("role='%s'", role)
	}

	return fmt.Sprintf(ssmlTemplate, roleAttribute, name, text)
}

func (p *AzureSpeechProvider) getRequestBody(request *types.SpeechAudioRequest) *bytes.Buffer {
	voiceMap := map[string][]string{
		"alloy":   {"zh-CN-YunxiNeural"},
		"echo":    {"zh-CN-YunyangNeural"},
		"fable":   {"zh-CN-YunxiNeural", "Boy"},
		"onyx":    {"zh-CN-YunyeNeural"},
		"nova":    {"zh-CN-XiaochenNeural"},
		"shimmer": {"zh-CN-XiaohanNeural"},
	}

	voice := ""
	role := ""
	if voiceMap[request.Voice] != nil {
		voice = voiceMap[request.Voice][0]
		if len(voiceMap[request.Voice]) > 1 {
			role = voiceMap[request.Voice][1]
		}
	}

	ssml := CreateSSML(request.Input, voice, role)

	return bytes.NewBufferString(ssml)

}

func (p *AzureSpeechProvider) CreateSpeech(request *types.SpeechAudioRequest) (*http.Response, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(common.RelayModeAudioSpeech)
	if errWithCode != nil {
		return nil, errWithCode
	}
	fullRequestURL := p.GetFullRequestURL(url, request.Model)
	headers := p.GetRequestHeaders()
	responseFormatr := outputFormatMap[request.ResponseFormat]
	if responseFormatr == "" {
		responseFormatr = outputFormatMap["mp3"]
	}
	headers["X-Microsoft-OutputFormat"] = responseFormatr

	requestBody := p.getRequestBody(request)

	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(requestBody), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}
	defer req.Body.Close()

	var resp *http.Response
	resp, errWithCode = p.Requester.SendRequestRaw(req)
	if errWithCode != nil {
		return nil, errWithCode
	}

	p.Usage.TotalTokens = p.Usage.PromptTokens

	return resp, nil
}

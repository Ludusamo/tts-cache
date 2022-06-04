package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tts "cloud.google.com/go/texttospeech/apiv1"
	"github.com/mailgun/groupcache/v2"
	ttsPb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

type TTSParams struct {
	LanguageCode   string  `json:"language_code"`
	VoiceName      string  `json:"voice_name"`
	AudioProfileId string  `json:"audio_profile_id"`
	Text           string  `json:"text"`
	SpeakingRate   float64 `json:"speaking_rate"`
	Pitch          float64 `json:"pitch"`
}

func (params TTSParams) GetKey() string {
	keyBytes, _ := json.Marshal(params)
	return string(keyBytes)
}

func TTSGetterFunc(ctx context.Context, key string, dest groupcache.Sink) error {
	cacheTime, _ := strconv.Atoi(os.Getenv("CACHE_TIME"))
	var params TTSParams
	if err := json.Unmarshal([]byte(key), &params); err != nil {
		return err
	}
	log.Println("fetching tts audio:", params)
	audio, err := getTTSFromGoogle(params)
	if err != nil {
		return err
	}
	return dest.SetBytes(audio, time.Now().Add(time.Minute*time.Duration(cacheTime)))
}

func getTTSFromGoogle(params TTSParams) ([]byte, error) {
	ctx := context.Background()

	client, err := tts.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewClient: %v", err)
	}
	defer client.Close()

	req := &ttsPb.SynthesizeSpeechRequest{
		Input: &ttsPb.SynthesisInput{
			InputSource: &ttsPb.SynthesisInput_Text{Text: params.Text},
		},
		Voice: &ttsPb.VoiceSelectionParams{
			LanguageCode: params.LanguageCode,
			Name:         params.VoiceName,
		},
		AudioConfig: &ttsPb.AudioConfig{
			AudioEncoding:    ttsPb.AudioEncoding_MP3,
			EffectsProfileId: []string{params.AudioProfileId},
			SpeakingRate:     params.SpeakingRate,
			Pitch:            params.Pitch,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, req)
	if err != nil {
		log.Println(fmt.Sprintf("SynthesizeSpeech: %v", err))
		return nil, err
	}

	return resp.GetAudioContent(), nil
}

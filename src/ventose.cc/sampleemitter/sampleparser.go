package sampleemitter

import (
	"errors"

	"ventose.cc/teleemitter"
)

const SampleTypeSize uint8 = 4

type SampleParser struct {
	typeSize uint8
	teleemitter.MessageParser
}

func NewSampleEmitter() (*teleemitter.TeleEmitter, error) {
	sp := &SampleParser{ typeSize:SampleTypeSize }
	em, err := teleemitter.NewEmitter("tcp", ":2000", SampleTypeSize, sp)
	if err != nil {
		panic(err)
	}
	em.Start()
	return em, nil
}

func (sp *SampleParser) ValidateTypeCodeSize(typeCodeSize uint8) error {
	if typeCodeSize != sp.typeSize {
		return errors.New("TypeSize: " + string(typeCodeSize) + " is != " + string(sp.typeSize))
	}
}

func (sp *SampleParser) Parse(b []byte) (teleemitter.IncommingMessageInterface, error) {
	if len(b) < int(sp.typeSize) {
		return nil, errors.New("Message is smaller than the TypeSizeCode")
	}


}
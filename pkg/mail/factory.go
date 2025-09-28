package mail

import (
	"fmt"
	"gin/user-management-api/internal/utils"
)

type ProviderType string

const (
	ProviderMailtrap ProviderType = "mailtrap"
)

type ProviderFactory interface {
	CreateProvider(config *MailConfig) (EmailProviderService, error)
}

type MailTrapProviderFactory struct {
}

func (f *MailTrapProviderFactory) CreateProvider(config *MailConfig) (EmailProviderService, error) {
	return NewMailTrapProvider(config)
}

func NewProviderFactory(providerType ProviderType) (ProviderFactory, error) {
	switch providerType {
	case ProviderMailtrap:
		return &MailTrapProviderFactory{}, nil
	default:
		return nil, utils.NewError(utils.InternalServerError, fmt.Sprintf("Unsuported provider type: %s", utils.ErrorCode(providerType)))
	}
}

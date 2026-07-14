package uploads

import "errors"

var (
	ErrInvalidLogo        = errors.New("invalid logo file")
	ErrLogoTooLarge       = errors.New("logo exceeds size limit")
	ErrUnsupportedExt     = errors.New("unsupported file format")
	ErrInvalidAttachment  = errors.New("invalid attachment file")
	ErrAttachmentTooLarge = errors.New("attachment exceeds size limit")
)

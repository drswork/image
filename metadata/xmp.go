package metadata

import (
	"context"
	"errors"
	"time"

	"github.com/drswork/image"
)

var xmpDecoder func(context.Context, string, ...image.ReadOption) (*XMP, error)

func RegisterXMPDecoder(d func(context.Context, string, ...image.ReadOption) (*XMP, error)) {
	xmpDecoder = d
}

var xmpEncoder func(context.Context, *XMP, ...image.WriteOption) (string, error)

func RegisterXMPEncoder(e func(context.Context, *XMP, ...image.WriteOption) (string, error)) {
	xmpEncoder = e
}

type LanguageAlternative struct {
	Language string
	Text     string
}

// Things in the Dublin Core namespace
type Core struct {
	Contributor []string
	Coverage    string
	Creator     []string
	Date        []time.Time
	Description []LanguageAlternative
	Format      string // this is the mime type
	Identifier  string
	Language    []string // Really locales
	Publisher   []string
	Relation    []string
	Rights      []LanguageAlternative
	Source      string
	Subject     []string
	Title       []LanguageAlternative
	Type        []string
}

// Things in the XMP namespace
type XMPSpecific struct {
	CreateDate   *time.Time
	CreatorTool  string
	Identifier   []string
	Label        string
	MetadataData *time.Time
	ModifiedDate *time.Time
	Rating       float64
}

// Things in the XMP rights management namespace
type XMPRights struct {
	Certificate  string
	Marked       *bool
	Owner        []string
	UsageTerms   []LanguageAlternative
	WebStatement string
}

// These three are placeholders and should be fixed later
type GUID string
type ResourceRef string
type RenditionClass string

// Things in the XMP Media Management namespace
type XMPMediaManagement struct {
	DerivedFrom        ResourceRef
	DocumentID         GUID
	InstanceID         GUID
	OriginalDocumentID GUID
	RenditionClass     RenditionClass
	RenditionParams    string
}

type XMPIDQ struct {
	Scheme string
}

// XMP holds the XMP metadata. It's a collection of sub-types
type XMP struct {
	CoreProperties  Core
	Properties      XMPSpecific
	Rights          XMPRights
	MediaManagement XMPMediaManagement
	IDQ             XMPIDQ
}

func DecodeXMP(ctx context.Context, b string, opt ...image.ReadOption) (*XMP, error) {
	if xmpDecoder == nil {
		return nil, errors.New("No registered XMP decoder")
	}
	return xmpDecoder(ctx, b, opt...)
}

func (x *XMP) Encode(ctx context.Context, opt ...image.WriteOption) (string, error) {
	if xmpEncoder == nil {
		return "", errors.New("No registered XMP encoder")
	}
	return xmpEncoder(ctx, x, opt...)
}

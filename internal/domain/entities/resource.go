// Package entities contains domain entities for the TelemetryFlow GO MCP service
package entities

import (
	"encoding/json"
	"time"

	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

// Resource represents an MCP resource entity
type Resource struct {
	uri         vo.ResourceURI
	name        string
	description string
	mimeType    vo.MimeType
	annotations *ResourceAnnotations
	reader      ResourceReader
	isTemplate  bool
	uriTemplate string
	createdAt   time.Time
	updatedAt   time.Time
	metadata    map[string]interface{}
}

// ResourceReader is the function signature for reading resource content
type ResourceReader func(uri string) (*ResourceContent, error)

// ResourceContent represents the content of a resource
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"` // Base64 encoded binary data
}

// ResourceAnnotations represents annotations for a resource
type ResourceAnnotations struct {
	Audience    []string `json:"audience,omitempty"`    // ["user", "assistant"]
	Priority    float64  `json:"priority,omitempty"`    // 0.0 to 1.0
	Description string   `json:"description,omitempty"` // Extended description
}

// NewResource creates a new Resource entity
func NewResource(uri vo.ResourceURI, name string) (*Resource, error) {
	now := time.Now().UTC()
	return &Resource{
		uri:       uri,
		name:      name,
		mimeType:  vo.MimeType{},
		createdAt: now,
		updatedAt: now,
		metadata:  make(map[string]interface{}),
	}, nil
}

// NewResourceTemplate creates a new resource template
func NewResourceTemplate(uriTemplate, name, description string) (*Resource, error) {
	now := time.Now().UTC()
	return &Resource{
		name:        name,
		description: description,
		isTemplate:  true,
		uriTemplate: uriTemplate,
		createdAt:   now,
		updatedAt:   now,
		metadata:    make(map[string]interface{}),
	}, nil
}

// URI returns the resource URI
func (r *Resource) URI() vo.ResourceURI {
	return r.uri
}

// Name returns the resource name
func (r *Resource) Name() string {
	return r.name
}

// SetName sets the resource name
func (r *Resource) SetName(name string) {
	r.name = name
	r.updatedAt = time.Now().UTC()
}

// Description returns the resource description
func (r *Resource) Description() string {
	return r.description
}

// SetDescription sets the resource description
func (r *Resource) SetDescription(description string) {
	r.description = description
	r.updatedAt = time.Now().UTC()
}

// MimeType returns the resource MIME type
func (r *Resource) MimeType() vo.MimeType {
	return r.mimeType
}

// SetMimeType sets the resource MIME type
func (r *Resource) SetMimeType(mimeType vo.MimeType) {
	r.mimeType = mimeType
	r.updatedAt = time.Now().UTC()
}

// Annotations returns the resource annotations
func (r *Resource) Annotations() *ResourceAnnotations {
	return r.annotations
}

// SetAnnotations sets the resource annotations
func (r *Resource) SetAnnotations(annotations *ResourceAnnotations) {
	r.annotations = annotations
	r.updatedAt = time.Now().UTC()
}

// Reader returns the resource reader
func (r *Resource) Reader() ResourceReader {
	return r.reader
}

// SetReader sets the resource reader
func (r *Resource) SetReader(reader ResourceReader) {
	r.reader = reader
	r.updatedAt = time.Now().UTC()
}

// IsTemplate returns whether the resource is a template
func (r *Resource) IsTemplate() bool {
	return r.isTemplate
}

// URITemplate returns the URI template
func (r *Resource) URITemplate() string {
	return r.uriTemplate
}

// CreatedAt returns the creation timestamp
func (r *Resource) CreatedAt() time.Time {
	return r.createdAt
}

// UpdatedAt returns the last update timestamp
func (r *Resource) UpdatedAt() time.Time {
	return r.updatedAt
}

// Metadata returns the resource metadata
func (r *Resource) Metadata() map[string]interface{} {
	return r.metadata
}

// SetMetadata sets a metadata value
func (r *Resource) SetMetadata(key string, value interface{}) {
	r.metadata[key] = value
	r.updatedAt = time.Now().UTC()
}

// Read reads the resource content
func (r *Resource) Read() (*ResourceContent, error) {
	if r.reader == nil {
		return &ResourceContent{
			URI:      r.uri.String(),
			MimeType: r.mimeType.String(),
			Text:     "",
		}, nil
	}
	return r.reader(r.uri.String())
}

// ToMCPResource converts the resource to MCP format
func (r *Resource) ToMCPResource() map[string]interface{} {
	result := map[string]interface{}{
		"name": r.name,
	}

	if r.isTemplate {
		result["uriTemplate"] = r.uriTemplate
	} else {
		result["uri"] = r.uri.String()
	}

	if r.description != "" {
		result["description"] = r.description
	}
	if !r.mimeType.IsEmpty() {
		result["mimeType"] = r.mimeType.String()
	}
	if r.annotations != nil {
		result["annotations"] = r.annotations
	}

	return result
}

// ToJSON returns the resource as JSON bytes
func (r *Resource) ToJSON() ([]byte, error) {
	return json.Marshal(r.ToMCPResource())
}

// ResourceList represents a list of resources
type ResourceList struct {
	Resources  []*Resource
	NextCursor string
}

// NewResourceList creates a new ResourceList
func NewResourceList() *ResourceList {
	return &ResourceList{
		Resources: make([]*Resource, 0),
	}
}

// Add adds a resource to the list
func (rl *ResourceList) Add(resource *Resource) {
	rl.Resources = append(rl.Resources, resource)
}

// Count returns the number of resources
func (rl *ResourceList) Count() int {
	return len(rl.Resources)
}

// IsEmpty returns whether the list is empty
func (rl *ResourceList) IsEmpty() bool {
	return len(rl.Resources) == 0
}

// ToMCPResourceList converts the list to MCP format
func (rl *ResourceList) ToMCPResourceList() map[string]interface{} {
	resources := make([]map[string]interface{}, len(rl.Resources))
	for i, r := range rl.Resources {
		resources[i] = r.ToMCPResource()
	}

	result := map[string]interface{}{
		"resources": resources,
	}
	if rl.NextCursor != "" {
		result["nextCursor"] = rl.NextCursor
	}
	return result
}

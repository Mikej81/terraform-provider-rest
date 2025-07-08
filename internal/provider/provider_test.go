package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestRestProvider_Metadata(t *testing.T) {
	p := &restProvider{version: "test"}
	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), req, resp)

	if resp.TypeName != "rest" {
		t.Errorf("Expected TypeName to be 'rest', got %s", resp.TypeName)
	}

	if resp.Version != "test" {
		t.Errorf("Expected Version to be 'test', got %s", resp.Version)
	}
}

func TestRestProvider_Schema(t *testing.T) {
	p := &restProvider{}
	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	// Check that required attributes are present
	requiredAttrs := []string{"api_token", "api_url"}
	for _, attr := range requiredAttrs {
		if _, exists := resp.Schema.Attributes[attr]; !exists {
			t.Errorf("Expected required attribute %s to exist", attr)
		}
	}

	// Check that optional attributes are present
	optionalAttrs := []string{"api_header", "timeout", "insecure", "retry_attempts", "max_idle_conns"}
	for _, attr := range optionalAttrs {
		if _, exists := resp.Schema.Attributes[attr]; !exists {
			t.Errorf("Expected optional attribute %s to exist", attr)
		}
	}
}

func TestRestProvider_DataSources(t *testing.T) {
	p := &restProvider{}
	dataSources := p.DataSources(context.Background())

	if len(dataSources) != 1 {
		t.Errorf("Expected 1 data source, got %d", len(dataSources))
	}

	// Test that the data source can be created
	ds := dataSources[0]()
	if ds == nil {
		t.Error("Expected data source to be created")
	}

	if _, ok := ds.(*RestDataSource); !ok {
		t.Errorf("Expected RestDataSource, got %T", ds)
	}
}

func TestRestProvider_Resources(t *testing.T) {
	p := &restProvider{}
	resources := p.Resources(context.Background())

	if len(resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(resources))
	}

	// Test that the resource can be created
	resource := resources[0]()
	if resource == nil {
		t.Error("Expected resource to be created")
	}

	if _, ok := resource.(*RestResource); !ok {
		t.Errorf("Expected RestResource, got %T", resource)
	}
}

func TestNew(t *testing.T) {
	version := "1.0.0"
	providerFunc := New(version)

	if providerFunc == nil {
		t.Error("Expected provider function to be returned")
	}

	provider := providerFunc()
	if provider == nil {
		t.Error("Expected provider to be created")
	}

	if restProvider, ok := provider.(*restProvider); ok {
		if restProvider.version != version {
			t.Errorf("Expected version %s, got %s", version, restProvider.version)
		}
	} else {
		t.Errorf("Expected restProvider, got %T", provider)
	}
}

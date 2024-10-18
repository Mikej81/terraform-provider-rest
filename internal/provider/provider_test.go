// Copyright (c) HashiCorp, Inc.

package provider_test

// import (
// 	"context"
// 	"testing"

// 	"github.com/hashicorp/terraform-plugin-framework/types"
// 	"github.com/stretchr/testify/require"
// )

// func TestGenericRestProvider_Metadata(t *testing.T) {
// 	p := provider.New
// 	resp := &provider.MetadataResponse{}

// 	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)

// 	require.Equal(t, "genericrest", resp.TypeName)
// 	require.Equal(t, "test", resp.Version)
// }

// func TestGenericRestProvider_Schema(t *testing.T) {
// 	p := provider.New("test")()
// 	resp := &provider.SchemaResponse{}

// 	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

// 	require.NotNil(t, resp.Schema)
// 	require.Contains(t, resp.Schema.Attributes, "api_token")
// 	require.Contains(t, resp.Schema.Attributes, "api_url")
// 	require.Contains(t, resp.Schema.Attributes, "api_header")
// }

// func TestGenericRestProvider_Configure(t *testing.T) {
// 	p := provider.New("test")()
// 	resp := &provider.ConfigureResponse{}

// 	config := map[string]interface{}{
// 		"api_token":  "test-token",
// 		"api_url":    "https://example.com",
// 		"api_header": "Authorization",
// 	}

// 	req := provider.ConfigureRequest{
// 		Config: types.ObjectValueFromMap(ctx, config),
// 	}

// 	p.Configure(context.Background(), req, resp)

// 	require.False(t, resp.Diagnostics.HasError())
// 	require.NotNil(t, resp.DataSourceData)
// 	require.NotNil(t, resp.ResourceData)
// }

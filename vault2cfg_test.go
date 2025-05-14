package vault2cfg

import (
	"context"
	"testing"

	"github.com/hashicorp/vault-client-go"
	"github.com/stretchr/testify/assert"
)

type basicConfig struct {
	Username string `vault:"username"`
	Password string `vault:"password"`
}

type nestedConfig struct {
	APIKey string `vault:"api_key"`
	DB     struct {
		Host     string `vault:"db_host"`
		Port     string `vault:"db_port"`
		Username string `vault:"db_user"`
		Password string `vault:"db_pass"`
	}
	Cache struct {
		Redis struct {
			URL      string `vault:"redis_url"`
			Password string `vault:"redis_pass"`
		}
	}
}

type complexConfig struct {
	Basic       basicConfig
	Environment string `vault:"env"`
	APIs        nestedConfig
	unexported  string `vault:"secret"` // Should be skipped
	Empty       string `vault:"empty"`
	NoTag       string
	Pointer     *basicConfig
}

type edgeCasesConfig struct {
	MissingData string `vault:"does_not_exist"`
	NilValue    string `vault:"nil_value"`
	WrongType   string `vault:"wrong_type"`
}

func TestBind_BasicUsage(t *testing.T) {
	cfg := &basicConfig{}
	data := map[string]interface{}{
		"username": "admin",
		"password": "secret123",
	}

	assert.NoError(t, Bind(cfg, data))

	assert.Equal(t, "admin", cfg.Username)
	assert.Equal(t, "secret123", cfg.Password)
}

func TestBind_NestedStructs(t *testing.T) {
	cfg := &nestedConfig{}
	data := map[string]interface{}{
		"api_key":    "abcd1234",
		"db_host":    "localhost",
		"db_port":    "5432",
		"db_user":    "postgres",
		"db_pass":    "postgres",
		"redis_url":  "redis://localhost:6379",
		"redis_pass": "redis123",
	}

	assert.NoError(t, Bind(cfg, data))

	assert.Equal(t, "abcd1234", cfg.APIKey)
	assert.Equal(t, "localhost", cfg.DB.Host)
	assert.Equal(t, "5432", cfg.DB.Port)
	assert.Equal(t, "postgres", cfg.DB.Username)
	assert.Equal(t, "postgres", cfg.DB.Password)
	assert.Equal(t, "redis://localhost:6379", cfg.Cache.Redis.URL)
	assert.Equal(t, "redis123", cfg.Cache.Redis.Password)
}

func TestBind_ComplexStructs(t *testing.T) {
	cfg := &complexConfig{
		Pointer: &basicConfig{},
	}
	data := map[string]interface{}{
		"username": "complex_user",
		"password": "complex_pass",
		"env":      "production",
		"api_key":  "complex_api_key",
		"secret":   "should_not_bind",
		"empty":    "",
	}

	assert.NoError(t, Bind(cfg, data))

	assert.Equal(t, "complex_user", cfg.Basic.Username)
	assert.Equal(t, "complex_pass", cfg.Basic.Password)
	assert.Equal(t, "production", cfg.Environment)
	assert.Equal(t, "complex_api_key", cfg.APIs.APIKey)
	assert.Equal(t, "", cfg.unexported) // Should remain empty
	assert.Equal(t, "", cfg.Empty)      // Should be set to empty string
	assert.Equal(t, "", cfg.NoTag)      // Should remain empty

	// Test pointer to struct
	assert.Equal(t, "complex_user", cfg.Pointer.Username)
	assert.Equal(t, "complex_pass", cfg.Pointer.Password)
}

func TestBind_EdgeCases(t *testing.T) {
	cfg := &edgeCasesConfig{}
	data := map[string]interface{}{
		"nil_value":  nil,
		"wrong_type": 12345, // Not a string
	}

	assert.NoError(t, Bind(cfg, data))

	// Missing data - should remain zero value
	assert.Equal(t, "", cfg.MissingData)

	// Nil value - should remain zero value
	assert.Equal(t, "", cfg.NilValue)

	// Parsed by fmt.Sprintf
	assert.Equal(t, "12345", cfg.WrongType)
}

func TestBind_NonPointer(t *testing.T) {
	// Test with non-pointer value
	cfg := basicConfig{}
	data := map[string]interface{}{
		"username": "direct_user",
		"password": "direct_pass",
	}

	assert.Error(t, Bind(cfg, data))
}

func TestBind_EmptyData(t *testing.T) {
	cfg := &basicConfig{
		Username: "default_user",
		Password: "default_pass",
	}

	// Empty data map should not change anything
	assert.NoError(t, Bind(cfg, map[string]interface{}{}))

	assert.Equal(t, "default_user", cfg.Username)
	assert.Equal(t, "default_pass", cfg.Password)
}

func TestBind_DeepNesting(t *testing.T) {
	type deeplyNested struct {
		Level1 struct {
			Level2 struct {
				Level3 struct {
					Level4 struct {
						Value string `vault:"deep_value"`
					}
				}
			}
		}
	}

	cfg := &deeplyNested{}
	data := map[string]interface{}{
		"deep_value": "found_me",
	}

	assert.NoError(t, Bind(cfg, data))

	assert.Equal(t, "found_me", cfg.Level1.Level2.Level3.Level4.Value)
}

func TestBind_MultipleRuns(t *testing.T) {
	cfg := &basicConfig{}

	// First run
	data1 := map[string]interface{}{
		"username": "first_user",
	}

	assert.NoError(t, Bind(cfg, data1))

	assert.Equal(t, "first_user", cfg.Username)
	assert.Equal(t, "", cfg.Password)

	// Second run with different data
	data2 := map[string]interface{}{
		"password": "second_pass",
	}
	assert.NoError(t, Bind(cfg, data2))

	assert.Equal(t, "first_user", cfg.Username) // Should remain unchanged
	assert.Equal(t, "second_pass", cfg.Password)

	// Third run with overlapping data
	data3 := map[string]interface{}{
		"username": "third_user",
		"password": "third_pass",
	}
	assert.NoError(t, Bind(cfg, data3))

	assert.Equal(t, "third_user", cfg.Username)
	assert.Equal(t, "third_pass", cfg.Password)
}

// ---

type _testClientCfg struct {
	ValueA string `vault:"key_group_1"`
	ValueB string `vault:"key_group_2"`
}

func TestWithClient(t *testing.T) {
	t.Skip("vault client is not available")

	client, err := vault.New(
		vault.WithAddress("http://localhost:8200"),
	)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}

	err = client.SetToken("qwerty")
	if err != nil {
		t.Fatalf("failed to set token: %v", err)
	}

	cfg := new(_testClientCfg)

	dataOne, err := client.Secrets.KvV2Read(
		context.Background(), "some-service/group1", vault.WithMountPath("secret"),
	)
	if err != nil {
		t.Fatalf("failed to read data from vault: %v", err)
	}

	assert.NoError(t, Bind(cfg, dataOne.Data.Data))

	assert.Equal(t, "value_group_1", cfg.ValueA)

	dataTwo, err := client.Secrets.KvV2Read(
		context.Background(), "some-service/group2", vault.WithMountPath("secret"),
	)
	if err != nil {
		t.Fatalf("failed to read data from vault: %v", err)
	}

	assert.NoError(t, Bind(cfg, dataTwo.Data.Data))

	assert.Equal(t, "value_group_2", cfg.ValueB)
}

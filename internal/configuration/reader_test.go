package configuration

import (
	"io/ioutil"
	"os"
	"path"
	"sort"
	"testing"

	"aletheia.icu/broccoli/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/utils"
)

func createTestingTempFile(t *testing.T, dir, name, content string) {
	err := ioutil.WriteFile(path.Join(dir, name), []byte(content), 0600)
	require.NoError(t, err)
}

func resetEnv() {
	_ = os.Unsetenv("AUTHELIA_JWT_SECRET_FILE")
	_ = os.Unsetenv("AUTHELIA_DUO_API_SECRET_KEY_FILE")
	_ = os.Unsetenv("AUTHELIA_SESSION_SECRET_FILE")
	_ = os.Unsetenv("AUTHELIA_SESSION_SECRET_FILE")
	_ = os.Unsetenv("AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE")
	_ = os.Unsetenv("AUTHELIA_NOTIFIER_SMTP_PASSWORD_FILE")
	_ = os.Unsetenv("AUTHELIA_SESSION_REDIS_PASSWORD_FILE")
	_ = os.Unsetenv("AUTHELIA_STORAGE_MYSQL_PASSWORD_FILE")
	_ = os.Unsetenv("AUTHELIA_STORAGE_POSTGRES_PASSWORD_FILE")
}

func setupEnv(t *testing.T) string {
	resetEnv()

	dirEnv := os.Getenv("AUTHELIA_TESTING_DIR")
	if dirEnv != "" {
		return dirEnv
	}

	dir := "/tmp/authelia" + utils.RandomString(10, authentication.HashingPossibleSaltCharacters) + "/"
	err := os.MkdirAll(dir, 0700)
	require.NoError(t, err)

	createTestingTempFile(t, dir, "jwt", "secret_from_env")
	createTestingTempFile(t, dir, "duo", "duo_secret_from_env")
	createTestingTempFile(t, dir, "session", "session_secret_from_env")
	createTestingTempFile(t, dir, "authentication", "ldap_secret_from_env")
	createTestingTempFile(t, dir, "notifier", "smtp_secret_from_env")
	createTestingTempFile(t, dir, "redis", "redis_secret_from_env")
	createTestingTempFile(t, dir, "mysql", "mysql_secret_from_env")
	createTestingTempFile(t, dir, "postgres", "postgres_secret_from_env")

	require.NoError(t, os.Setenv("AUTHELIA_TESTING_DIR", dir))

	return dir
}

func TestShouldErrorNoConfigPath(t *testing.T) {
	_, errors := Read("")

	require.Len(t, errors, 1)

	require.EqualError(t, errors[0], "No config file path provided")
}

func TestShouldErrorNoConfigFileInEmbeddedFS(t *testing.T) {
	oldCfg := cfg
	cfg = fs.New(false, []byte("\x1b~\x00\x80\x8d\x94n\xc2|\x84J\xf7\xbfn\xfd\xf7w;.\x8d m\xb2&\xd1Z\xec\xb2\x05\xb9\xc00\x8a\xf7(\x80^78\t(\f\f\xc3p\xc2\xc1\x06[a\xa2\xb3\xa4P\xe5\xa14\xfb\x19\xb2cp\xf6\x90-Z\xb2\x11\xe0l\xa1\x80\\\x95Vh\t\xc5\x06\x16\xfa\x8c\xc0\"!\xa5\xcf\xf7$\x9a\xb2\a`\xc6\x18\xc8~\xce8\r\x16Z\x9d\xc3\xe3\xff\x00"))
	_, errors := Read("./nonexistent.yml")
	cfg = oldCfg

	require.Len(t, errors, 3)

	require.EqualError(t, errors[0], "Unable to find config file: ./nonexistent.yml")
	require.EqualError(t, errors[1], "Generating config file: ./nonexistent.yml")
	require.EqualError(t, errors[2], "Unable to open config.template.yml: file does not exist")
}

func TestShouldErrorPermissionsOnLocalFS(t *testing.T) {
	_ = os.Mkdir("/tmp/noperms/", 0000)
	_, errors := Read("/tmp/noperms/configuration.yml")

	require.Len(t, errors, 3)

	require.EqualError(t, errors[0], "Unable to find config file: /tmp/noperms/configuration.yml")
	require.EqualError(t, errors[1], "Generating config file: /tmp/noperms/configuration.yml")
	require.EqualError(t, errors[2], "Unable to generate /tmp/noperms/configuration.yml: open /tmp/noperms/configuration.yml: permission denied")
}

func TestShouldErrorAndGenerateConfigFile(t *testing.T) {
	_, errors := Read("./nonexistent.yml")
	_ = os.Remove("./nonexistent.yml")

	require.Len(t, errors, 3)

	require.EqualError(t, errors[0], "Unable to find config file: ./nonexistent.yml")
	require.EqualError(t, errors[1], "Generating config file: ./nonexistent.yml")
	require.EqualError(t, errors[2], "Generated configuration at: ./nonexistent.yml")
}

func TestShouldErrorPermissionsConfigFile(t *testing.T) {
	_ = ioutil.WriteFile("/tmp/authelia/permissions.yml", []byte{}, 0000) // nolint:gosec
	_, errors := Read("/tmp/authelia/permissions.yml")

	require.Len(t, errors, 1)

	require.EqualError(t, errors[0], "Failed to open /tmp/authelia/permissions.yml: permission denied")
}

func TestShouldErrorParseBadConfigFile(t *testing.T) {
	_, errors := Read("./test_resources/config_bad_quoting.yml")

	require.Len(t, errors, 1)

	require.EqualError(t, errors[0], "Error malformed yaml: line 23: did not find expected alphabetic or numeric character")
}

func TestShouldParseConfigFile(t *testing.T) {
	dir := setupEnv(t)

	require.NoError(t, os.Setenv("AUTHELIA_JWT_SECRET_FILE", dir+"jwt"))
	require.NoError(t, os.Setenv("AUTHELIA_DUO_API_SECRET_KEY_FILE", dir+"duo"))
	require.NoError(t, os.Setenv("AUTHELIA_SESSION_SECRET_FILE", dir+"session"))
	require.NoError(t, os.Setenv("AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", dir+"authentication"))
	require.NoError(t, os.Setenv("AUTHELIA_NOTIFIER_SMTP_PASSWORD_FILE", dir+"notifier"))
	require.NoError(t, os.Setenv("AUTHELIA_SESSION_REDIS_PASSWORD_FILE", dir+"redis"))
	require.NoError(t, os.Setenv("AUTHELIA_STORAGE_MYSQL_PASSWORD_FILE", dir+"mysql"))
	require.NoError(t, os.Setenv("AUTHELIA_STORAGE_POSTGRES_PASSWORD_FILE", dir+"postgres"))

	config, errors := Read("./test_resources/config.yml")

	require.Len(t, errors, 0)

	assert.Equal(t, 9091, config.Port)
	assert.Equal(t, "debug", config.LogLevel)
	assert.Equal(t, "https://home.example.com:8080/", config.DefaultRedirectionURL)
	assert.Equal(t, "authelia.com", config.TOTP.Issuer)
	assert.Equal(t, "secret_from_env", config.JWTSecret)

	assert.Equal(t, "api-123456789.example.com", config.DuoAPI.Hostname)
	assert.Equal(t, "ABCDEF", config.DuoAPI.IntegrationKey)
	assert.Equal(t, "duo_secret_from_env", config.DuoAPI.SecretKey)

	assert.Equal(t, "session_secret_from_env", config.Session.Secret)
	assert.Equal(t, "ldap_secret_from_env", config.AuthenticationBackend.Ldap.Password)
	assert.Equal(t, "smtp_secret_from_env", config.Notifier.SMTP.Password)
	assert.Equal(t, "redis_secret_from_env", config.Session.Redis.Password)
	assert.Equal(t, "mysql_secret_from_env", config.Storage.MySQL.Password)

	assert.Equal(t, "deny", config.AccessControl.DefaultPolicy)
	assert.Len(t, config.AccessControl.Rules, 12)
}

func TestShouldParseAltConfigFile(t *testing.T) {
	dir := setupEnv(t)

	require.NoError(t, os.Setenv("AUTHELIA_STORAGE_POSTGRES_PASSWORD_FILE", dir+"postgres"))
	require.NoError(t, os.Setenv("AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", dir+"authentication"))
	require.NoError(t, os.Setenv("AUTHELIA_JWT_SECRET_FILE", dir+"jwt"))
	require.NoError(t, os.Setenv("AUTHELIA_SESSION_SECRET_FILE", dir+"session"))

	config, errors := Read("./test_resources/config_alt.yml")
	require.Len(t, errors, 0)

	assert.Equal(t, 9091, config.Port)
	assert.Equal(t, "debug", config.LogLevel)
	assert.Equal(t, "https://home.example.com:8080/", config.DefaultRedirectionURL)
	assert.Equal(t, "authelia.com", config.TOTP.Issuer)
	assert.Equal(t, "secret_from_env", config.JWTSecret)

	assert.Equal(t, "api-123456789.example.com", config.DuoAPI.Hostname)
	assert.Equal(t, "ABCDEF", config.DuoAPI.IntegrationKey)
	assert.Equal(t, "postgres_secret_from_env", config.Storage.PostgreSQL.Password)

	assert.Equal(t, "deny", config.AccessControl.DefaultPolicy)
	assert.Len(t, config.AccessControl.Rules, 12)
}

func TestShouldNotParseConfigFileWithOldOrUnexpectedKeys(t *testing.T) {
	dir := setupEnv(t)

	require.NoError(t, os.Setenv("AUTHELIA_JWT_SECRET_FILE", dir+"jwt"))
	require.NoError(t, os.Setenv("AUTHELIA_DUO_API_SECRET_KEY_FILE", dir+"duo"))
	require.NoError(t, os.Setenv("AUTHELIA_SESSION_SECRET_FILE", dir+"session"))
	require.NoError(t, os.Setenv("AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", dir+"authentication"))
	require.NoError(t, os.Setenv("AUTHELIA_NOTIFIER_SMTP_PASSWORD_FILE", dir+"notifier"))
	require.NoError(t, os.Setenv("AUTHELIA_SESSION_REDIS_PASSWORD_FILE", dir+"redis"))
	require.NoError(t, os.Setenv("AUTHELIA_STORAGE_MYSQL_PASSWORD_FILE", dir+"mysql"))

	_, errors := Read("./test_resources/config_bad_keys.yml")
	require.Len(t, errors, 2)

	// Sort error slice to prevent shenanigans that somehow occur
	sort.Slice(errors, func(i, j int) bool {
		return errors[i].Error() < errors[j].Error()
	})
	assert.EqualError(t, errors[0], "config key not expected: loggy_file")
	assert.EqualError(t, errors[1], "config key replaced: logs_level is now log_level")
}

func TestShouldValidateConfigurationTemplate(t *testing.T) {
	resetEnv()

	_, errors := Read("../../config.template.yml")
	assert.Len(t, errors, 0)
}

func TestShouldOnlyAllowEnvOrConfig(t *testing.T) {
	dir := setupEnv(t)

	resetEnv()
	require.NoError(t, os.Setenv("AUTHELIA_JWT_SECRET_FILE", dir+"jwt"))
	require.NoError(t, os.Setenv("AUTHELIA_DUO_API_SECRET_KEY_FILE", dir+"duo"))
	require.NoError(t, os.Setenv("AUTHELIA_SESSION_SECRET_FILE", dir+"session"))
	require.NoError(t, os.Setenv("AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", dir+"authentication"))
	require.NoError(t, os.Setenv("AUTHELIA_NOTIFIER_SMTP_PASSWORD_FILE", dir+"notifier"))
	require.NoError(t, os.Setenv("AUTHELIA_SESSION_REDIS_PASSWORD_FILE", dir+"redis"))
	require.NoError(t, os.Setenv("AUTHELIA_STORAGE_MYSQL_PASSWORD_FILE", dir+"mysql"))

	_, errors := Read("./test_resources/config_with_secret.yml")

	require.Len(t, errors, 1)
	require.EqualError(t, errors[0], "error loading secret (jwt_secret): it's already defined in the config file")
}

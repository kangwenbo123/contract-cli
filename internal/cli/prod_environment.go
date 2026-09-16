package cli

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"cn.qfei/contract-cli/internal/config"
	"cn.qfei/contract-cli/internal/credential"
)

const (
	productionEnvironment        = "prod"
	productionOpenPlatformURL    = "https://open.qfei.cn"
	productionAccountOrigin      = "https://myaccount.qfei.cn"
	productionOpenPlatformOrigin = "https://open.qfei.cn"
	testE2EEnvironment           = "test"
	testE2EOpenPlatformURL       = "https://test-open.qtech.cn"
	testE2EAccountOrigin         = "https://test-myaccount.qtech.cn"
	testE2EOpenPlatformOrigin    = "https://test-open.qtech.cn"
)

var blockedProductionBuildHosts = map[string]struct{}{
	"dev-open.qtech.cn":      {},
	"dev-myaccount.qtech.cn": {},
}

var allowedTestE2EHosts = map[string]struct{}{
	"test-open.qtech.cn":      {},
	"test-myaccount.qtech.cn": {},
}

func validateProductionProfile(profile config.Profile) error {
	if testE2EBuild {
		return validateTestE2EProfile(profile)
	}
	if strings.TrimSpace(profile.Environment) != productionEnvironment {
		return productionProfileError(profile.Name)
	}
	if !isExactProductionResource(profile.OpenPlatformBaseURL) || !isExactProductionResource(profile.Resource) {
		return productionProfileError(profile.Name)
	}
	openPlatformURLs := []string{
		profile.AppTokenEndpoint,
		profile.ProtectedResourceMetadataURL,
	}
	for _, rawURL := range openPlatformURLs {
		if !isProductionOriginURL(rawURL, productionOpenPlatformOrigin, false) {
			return productionProfileError(profile.Name)
		}
	}
	accountURLs := []string{
		profile.AuthorizationServerMetadataURL,
		profile.Identities.User.AuthorizationEndpoint,
		profile.Identities.User.DeviceAuthorizationEndpoint,
		profile.Identities.User.TokenEndpoint,
		profile.Identities.User.RevocationEndpoint,
		profile.Identities.User.RegistrationEndpoint,
	}
	for _, rawURL := range accountURLs {
		if !isProductionOriginURL(rawURL, productionAccountOrigin, false) {
			return productionProfileError(profile.Name)
		}
	}
	return nil
}

func validateTestE2EProfile(profile config.Profile) error {
	if strings.TrimSpace(profile.Environment) != testE2EEnvironment {
		return productionProfileError(profile.Name)
	}
	if !isExactEnvironmentResource(profile.OpenPlatformBaseURL, testE2EOpenPlatformURL) ||
		!isExactEnvironmentResource(profile.Resource, testE2EOpenPlatformURL) {
		return productionProfileError(profile.Name)
	}
	for _, rawURL := range []string{profile.AppTokenEndpoint, profile.ProtectedResourceMetadataURL} {
		if !isProductionOriginURL(rawURL, testE2EOpenPlatformOrigin, false) {
			return productionProfileError(profile.Name)
		}
	}
	for _, rawURL := range []string{
		profile.AuthorizationServerMetadataURL,
		profile.Identities.User.AuthorizationEndpoint,
		profile.Identities.User.DeviceAuthorizationEndpoint,
		profile.Identities.User.TokenEndpoint,
		profile.Identities.User.RevocationEndpoint,
		profile.Identities.User.RegistrationEndpoint,
	} {
		if !isProductionOriginURL(rawURL, testE2EAccountOrigin, false) {
			return productionProfileError(profile.Name)
		}
	}
	return nil
}

func validateProductionDeviceCredential(profileName string, stored credential.DeviceCredential) error {
	if stored.DeviceProfile != nil {
		profile, err := restoreDeviceProfile(profileName, stored.DeviceProfile)
		if err != nil || validateProductionProfile(profile) != nil {
			return productionProfileError(profileName)
		}
	}
	return validateProductionPendingTransaction(profileName, stored.Pending)
}

func validateProductionPendingTransaction(profileName string, pending *credential.PendingTransaction) error {
	expectedAccountOrigin := productionAccountOrigin
	if testE2EBuild {
		expectedAccountOrigin = testE2EAccountOrigin
	}
	if pending != nil &&
		(!isProductionOriginURL(pending.TokenEndpoint, expectedAccountOrigin, true) ||
			!isProductionOriginURL(pending.VerificationURIComplete, expectedAccountOrigin, false)) {
		return productionProfileError(profileName)
	}
	return nil
}

func isExactProductionResource(rawURL string) bool {
	return isExactEnvironmentResource(rawURL, productionOpenPlatformURL)
}

func isExactEnvironmentResource(rawURL string, expectedURL string) bool {
	parsed, err := parseProductionURL(rawURL)
	if err != nil {
		return false
	}
	return parsed.Scheme+"://"+parsed.Host == expectedURL &&
		(parsed.EscapedPath() == "" || parsed.EscapedPath() == "/") && parsed.RawQuery == ""
}

func isProductionOriginURL(rawURL, expectedOrigin string, required bool) bool {
	if strings.TrimSpace(rawURL) == "" {
		return !required
	}
	parsed, err := parseProductionURL(rawURL)
	if err != nil {
		return false
	}
	return parsed.Scheme+"://"+parsed.Host == expectedOrigin
}

func parseProductionURL(rawURL string) (*url.URL, error) {
	trimmed := strings.TrimSpace(rawURL)
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
		return nil, errors.New("invalid production URL")
	}
	if parsed.Hostname() != strings.TrimSuffix(parsed.Hostname(), ".") {
		return nil, errors.New("trailing-dot hosts are not allowed")
	}
	return parsed, nil
}

func productionProfileError(profileName string) error {
	if testE2EBuild {
		return fmt.Errorf(
			"profile %q is not a Test profile and is not allowed in this Test E2E build; run `contract-cli config add --env test --name contract-test` and authorize again",
			profileName,
		)
	}
	return fmt.Errorf(
		"profile %q belongs to a non-production environment and is not allowed in this production build; run `contract-cli config add --env prod --name contract` and authorize again",
		profileName,
	)
}

func withProductionNetworkGuard(client *http.Client, logger *slog.Logger) *http.Client {
	guarded := *client
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	guarded.Transport = productionGuardTransport{next: transport, logger: logger}
	return &guarded
}

type productionGuardTransport struct {
	next   http.RoundTripper
	logger *slog.Logger
}

func (transport productionGuardTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	host := strings.ToLower(strings.TrimSuffix(request.URL.Hostname(), "."))
	if testE2EBuild {
		transport.logger.Debug("Test E2E network request", "method", request.Method, "host", host)
		if _, allowed := allowedTestE2EHosts[host]; !allowed {
			transport.logger.Error("Test E2E build blocked non-Test network request", "method", request.Method, "host", host)
			return nil, fmt.Errorf("Test E2E build blocks non-Test host %q", host)
		}
		return transport.next.RoundTrip(request)
	}
	transport.logger.Debug("production network request", "method", request.Method, "host", host)
	if _, blocked := blockedProductionBuildHosts[host]; blocked {
		transport.logger.Error("production build blocked non-production network request", "method", request.Method, "host", host)
		return nil, fmt.Errorf("production build blocks non-production host %q", host)
	}
	return transport.next.RoundTrip(request)
}

func (a *App) loadProductionDeviceCredential(profile config.Profile, store credential.Store) (credential.DeviceCredential, error) {
	stored, err := store.Load(profile.Name)
	if err != nil {
		return credential.DeviceCredential{}, err
	}
	if err := validateProductionDeviceCredential(profile.Name, stored); err != nil {
		a.logger.Error("reject non-production Device credential", "profile", profile.Name, "error", err.Error())
		return credential.DeviceCredential{}, err
	}
	return stored, nil
}

func (a *App) validateProductionDeviceCredentialIfAvailable(profile config.Profile) error {
	store, available, err := a.deviceCredentialStoreIfAvailable()
	if err != nil || !available {
		return err
	}
	stored, err := store.Load(profile.Name)
	if errors.Is(err, credential.ErrCredentialNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := validateProductionDeviceCredential(profile.Name, stored); err != nil {
		a.logger.Error("reject non-production Device credential", "profile", profile.Name, "error", err.Error())
		return err
	}
	return nil
}

func (a *App) deviceCredentialStoreIfAvailable() (credential.Store, bool, error) {
	if a.credentialStore != nil {
		return a.credentialStore, true, nil
	}
	for _, name := range []string{"SKILL_SESSION_WORKSPACE", "CODEBUDDY_SESSION_ID", "SESSION_ID"} {
		if value, ok := a.lookupEnv(name); ok && strings.TrimSpace(value) != "" {
			store, err := a.deviceCredentials()
			return store, err == nil, err
		}
	}
	return nil, false, nil
}

func (a *App) productionProfileRequiresReset(profile config.Profile) (bool, error) {
	if validateProductionProfile(profile) != nil {
		return true, nil
	}
	store, available, err := a.deviceCredentialStoreIfAvailable()
	if err != nil || !available {
		return false, err
	}
	stored, err := store.Load(profile.Name)
	if errors.Is(err, credential.ErrCredentialNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return validateProductionDeviceCredential(profile.Name, stored) != nil, nil
}

func (a *App) clearProfileAuthenticationState(profileName string) error {
	a.logger.Info("clear non-production profile authentication state", "profile", profileName)
	if err := a.secrets.Delete(config.AppSecretKey(profileName)); err != nil {
		a.logger.Error("clear app secret failed", "profile", profileName, "error", err.Error())
		return err
	}
	store, available, err := a.deviceCredentialStoreIfAvailable()
	if err != nil {
		a.logger.Error("resolve Device credential store for reset failed", "profile", profileName, "error", err.Error())
		return err
	}
	if available {
		if err := store.Delete(profileName); err != nil {
			a.logger.Error("clear Device credential failed", "profile", profileName, "error", err.Error())
			return err
		}
	}
	return nil
}

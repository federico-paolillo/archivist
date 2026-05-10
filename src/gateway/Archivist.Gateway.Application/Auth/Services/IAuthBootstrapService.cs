namespace Archivist.Gateway.Application.Auth.Services;

/// <summary>
/// Ensures the personal user's password hash is initialized before the gateway accepts requests.
/// </summary>
public interface IAuthBootstrapService
{
    /// <summary>
    /// Ensures the personal user row exists and that <c>password_hash</c> is populated.
    /// Bootstraps from <c>AUTH_BOOTSTRAP_PASSWORD</c> only when the hash is missing.
    /// </summary>
    Task InitializeAsync(CancellationToken ct = default);
}
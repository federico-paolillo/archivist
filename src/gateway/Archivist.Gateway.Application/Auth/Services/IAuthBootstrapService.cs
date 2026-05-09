namespace Archivist.Gateway.Application.Auth.Services;

/// <summary>
/// Ensures the personal user row exists and that a password hash is stored.
/// Bootstraps the password hash from the configured bootstrap secret only when the row has no stored hash.
/// </summary>
public interface IAuthBootstrapService
{
    /// <summary>
    /// Initializes auth storage.
    /// Inserts the personal user row if absent, then hashes and stores the bootstrap password
    /// if and only if <c>password_hash</c> is currently <c>NULL</c>.
    /// When a hash already exists the bootstrap password is not required and the existing hash is preserved.
    /// </summary>
    /// <param name="ct">Cancellation token.</param>
    Task InitializeAsync(CancellationToken ct = default);
}